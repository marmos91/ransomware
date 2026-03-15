package cli

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/marmos91/ransomware/crypto"
	"github.com/marmos91/ransomware/fs"
	urfavecli "github.com/urfave/cli/v2"
)

const (
	PUBLIC_KEY_NAME  = "pub.pem"
	PRIVATE_KEY_NAME = "priv.pem"
)

var validKeySizes = map[int]bool{2048: true, 3072: true, 4096: true}

func CreateKeys(ctx *urfavecli.Context) error {
	path := ctx.String("path")
	keySize := ctx.Int("keySize")

	if !validKeySizes[keySize] {
		return fmt.Errorf("invalid key size %d: must be 2048, 3072, or 4096", keySize)
	}

	rsaKeypair, err := crypto.NewRandomRsaKeypair(keySize)
	if err != nil {
		return err
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	slog.Info("Generated random keys", "path", absolutePath, "keySize", keySize)
	slog.Warn("Remember to hide your private key!", "file", PRIVATE_KEY_NAME)

	privatePemContent := crypto.ExportRsaPrivateKeyAsPemStr(rsaKeypair.Private)
	publicPemContent, err := crypto.ExportRsaPublicKeyAsPemStr(rsaKeypair.Public)
	if err != nil {
		return err
	}

	privKeyPath := filepath.Join(absolutePath, PRIVATE_KEY_NAME)
	if err := fs.WriteToFileWithMode(privKeyPath, []byte(privatePemContent), 0600); err != nil {
		return err
	}

	pubKeyPath := filepath.Join(absolutePath, PUBLIC_KEY_NAME)
	if err := fs.WriteStringToFile(pubKeyPath, publicPemContent); err != nil {
		if removeErr := os.Remove(privKeyPath); removeErr != nil {
			return fmt.Errorf("failed to write public key: %w (also failed to clean up private key: %v)", err, removeErr)
		}
		return err
	}

	return nil
}
