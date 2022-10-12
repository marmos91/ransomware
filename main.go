package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

const APP_NAME = "ransomware"

func main() {
	app := &cli.App{
		Name:  APP_NAME,
		Usage: "A simple demonstration tool to simulate a ransomware attack",
		Action: func(*cli.Context) error {
			fmt.Println("Hello world!")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
