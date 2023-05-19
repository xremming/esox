package esox

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func setupLogger(isDev bool) zerolog.Logger {
	var w io.Writer = os.Stderr
	if isDev {
		w = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	}

	log := zerolog.New(w).With().Timestamp().Caller().Logger()
	zerolog.DefaultContextLogger = &log

	return log
}
