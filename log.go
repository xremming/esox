package esox

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	globalLog "github.com/rs/zerolog/log"
)

func setupLogger(isDev bool) zerolog.Logger {
	var w io.Writer = os.Stderr
	if isDev {
		w = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	}

	log := zerolog.New(w).With().Timestamp().Caller().Logger()
	globalLog.Logger = log
	zerolog.DefaultContextLogger = &log

	return log
}
