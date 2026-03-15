package utils

import (
	"io"
	"log/slog"
	"os"
)

// SetupLogging configures the default slog logger. When verbose is true,
// messages at Debug level and above are written to stderr. When verbose is
// false, all output is discarded. If jsonFormat is true the handler emits
// JSON; otherwise it emits plain text.
func SetupLogging(verbose, jsonFormat bool) {
	if !verbose {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
		return
	}

	opts := &slog.HandlerOptions{Level: slog.LevelDebug}

	var handler slog.Handler
	if jsonFormat {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	slog.SetDefault(slog.New(handler))
}
