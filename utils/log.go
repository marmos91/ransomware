package utils

import (
	"io"
	"log"
)

func DisableLogs() {
	log.SetFlags(0)
	log.SetOutput(io.Discard)
}
