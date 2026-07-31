package lecho

import (
	"io"
	"log/slog"

	"github.com/rs/zerolog"
)

// Logger is a thin wrapper around zerolog.Logger that provides bridges for Echo.
type Logger struct {
	log zerolog.Logger
}

// New returns a Logger that writes to out.
func New(out io.Writer, options ...Option) *Logger {
	switch logger := out.(type) {
	case zerolog.Logger:
		return From(logger, options...)
	case *zerolog.Logger:
		if logger == nil {
			return From(zerolog.New(nil), options...)
		}

		return From(*logger, options...)
	default:
		return From(zerolog.New(out), options...)
	}
}

// From returns a Logger using an existing zerolog logger.
func From(logger zerolog.Logger, options ...Option) *Logger {
	for _, option := range options {
		logger = option(logger)
	}

	return &Logger{log: logger}
}

// Level returns the minimum enabled log level.
func (l Logger) Level() zerolog.Level {
	return l.log.GetLevel()
}

// SetLevel sets the minimum enabled log level.
//
// SetLevel is intended for configuration before the logger is used concurrently.
func (l *Logger) SetLevel(level zerolog.Level) {
	l.log = l.log.Level(level)
}

// Output returns the wrapped zerolog logger as an io.Writer.
//
// The returned writer is a snapshot of the current logger configuration.
func (l Logger) Output() io.Writer {
	return l.log
}

// SetOutput sets the output for future log events.
//
// SetOutput is intended for configuration before the logger is used concurrently.
func (l *Logger) SetOutput(out io.Writer) {
	l.log = l.log.Output(out)
}

// Unwrap returns the wrapped zerolog logger.
func (l Logger) Unwrap() zerolog.Logger {
	return l.log
}

// Slog returns a slog logger backed by the wrapped zerolog logger.
//
// The returned logger is a snapshot of the current logger configuration.
func (l Logger) Slog() *slog.Logger {
	return slog.New(zerolog.NewSlogHandler(l.log))
}
