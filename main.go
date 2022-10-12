package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/marmos91/ransomware/crypto"
	"github.com/marmos91/ransomware/ransomfs"
	"github.com/urfave/cli/v2"
)

const APP_NAME = "ransomware"
const APP_DESCRIPTION = "A simple demonstration tool to simulate a ransomware attack"
const APP_VERSION = "v1.0.0"
const AES_KEY_SIZE = 32
const RSA_KEY_SIZE = 2048
const PUBLIC_KEY_NAME = "pub.pem"
const PRIVATE_KEY_NAME = "priv.pem"

func createKeys(ctx *cli.Context) error {
	path := ctx.String("path")

	rsaKeypair, err := crypto.NewRandomRsaKeypair(RSA_KEY_SIZE)

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

func encrypt(ctx *cli.Context) error {
	return nil
}

func decrypt(ctx *cli.Context) error {
	return nil
}

func SplashScreen(silent bool) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		if !silent {
			figure.NewFigure(APP_NAME, "graffiti", true).Print()
		}

		return nil
	}
}

func main() {
	var silent bool

	app := &cli.App{
		Name:     APP_NAME,
		Usage:    APP_DESCRIPTION,
		Version:  APP_VERSION,
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  "Marco Moschettini",
				Email: "marco.moschettini@cubbit.io",
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "silent",
				Usage:       "Runs the tool in silent mode (no logs)",
				Destination: &silent,
				Value:       false,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "create-keys",
				Aliases: []string{"c"},
				Usage:   "Generates a new random keypair and saves it to a file",
				Before:  SplashScreen(silent),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "path",
						Aliases: []string{"p"},
						Usage:   "The path where to save the keys",
						Value:   ".",
					},
				},
				Action: createKeys,
			},
			{
				Name:    "encrypt",
				Usage:   "Encrypts a directory",
				Aliases: []string{"e"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "path",
						Aliases: []string{"p"},
						Usage:   "Runs the tool on a directory",
					},
				},
				Before: SplashScreen(silent),
				Action: encrypt,
			},
			{
				Name:    "decrypt",
				Usage:   "Decrypts a directory",
				Aliases: []string{"d"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "path",
						Aliases: []string{"c"},
						Usage:   "Runs the tool on a directory",
					},
				},
				Before: SplashScreen(silent),
				Action: decrypt,
			},
		},
	}

	// walkErr := ransomfs.WalkFilesWithExtFilter(".", []string{".go", ".sample", ".md"}, true, func(path string, info fs.FileInfo) error {
	// 	fmt.Println(path)

	// 	return nil
	// })

	// if walkErr != nil {
	// 	log.Println(walkErr)
	// }

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
