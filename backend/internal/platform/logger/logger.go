package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func InitLogger(env string) {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	if env == "development" {
		Log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
			With().
			Timestamp().
			Caller().
			Logger()
	} else {
		Log = zerolog.New(os.Stdout).
			With().
			Timestamp().
			Logger()
	}
}
