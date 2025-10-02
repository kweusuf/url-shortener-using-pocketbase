package log

import (
	"flag"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Using a simple library here. Can be changed as we are interfacing this in other classes.
func Info(format string, args ...interface{}) {
	logger(zerolog.InfoLevel, format, args...)
}

func Debug(format string, args ...interface{}) {
	logger(zerolog.DebugLevel, format, args...)
}

func Error(format string, args ...interface{}) {
	logger(zerolog.ErrorLevel, format, args...)
}

func Fatal(format string, args ...interface{}) {
	logger(zerolog.FatalLevel, format, args...)
}

func Panic(format string, args ...interface{}) {
	logger(zerolog.PanicLevel, format, args...)
}

func Warn(format string, args ...interface{}) {
	logger(zerolog.WarnLevel, format, args...)
}

func logger(level zerolog.Level, format string, args ...interface{}) {
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	log.WithLevel(level).Msg(msg)
}

func InitializeLogging() {
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	debug := flag.Bool("debug", false, "sets log level to debug")

	flag.Parse()

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
