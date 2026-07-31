package lecho_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/ziflex/lecho/v4"
)

func TestNew(t *testing.T) {
	b := &bytes.Buffer{}
	logger := lecho.New(b)

	unwrapped := logger.Unwrap()
	unwrapped.Info().Msg("foo")

	assert.Equal(t, `{"level":"info","message":"foo"}`+"\n", b.String())
}

func TestNewWithZerolog(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		b := &bytes.Buffer{}
		zerologger := zerolog.New(b).With().Str("key", "value").Logger()
		logger := lecho.New(zerologger)

		unwrapped := logger.Unwrap()
		unwrapped.Info().Msg("foo")

		assert.Contains(t, b.String(), `"key":"value"`)
	})

	t.Run("pointer", func(t *testing.T) {
		b := &bytes.Buffer{}
		zerologger := zerolog.New(b).With().Str("key", "pointer").Logger()
		logger := lecho.New(&zerologger)

		unwrapped := logger.Unwrap()
		unwrapped.Info().Msg("foo")

		assert.Contains(t, b.String(), `"key":"pointer"`)
	})

	t.Run("nil pointer", func(t *testing.T) {
		var zerologger *zerolog.Logger

		assert.NotPanics(t, func() {
			lecho.New(zerologger)
		})
	})
}

func TestFrom(t *testing.T) {
	b := &bytes.Buffer{}
	zerologger := zerolog.New(b).With().Str("key", "test").Logger()
	logger := lecho.From(zerologger, lecho.WithField("source", "from"))

	unwrapped := logger.Unwrap()
	unwrapped.Info().Msg("foo")

	assert.Contains(t, b.String(), `"key":"test"`)
	assert.Contains(t, b.String(), `"source":"from"`)
}

func TestSlog(t *testing.T) {
	b := &bytes.Buffer{}
	logger := lecho.New(b, lecho.WithField("logger", "lecho"))

	logger.Slog().Info("from slog", "answer", 42)

	assert.Contains(t, b.String(), `"logger":"lecho"`)
	assert.Contains(t, b.String(), `"answer":42`)
	assert.Contains(t, b.String(), `"message":"from slog"`)
}

func TestLoggerLevel(t *testing.T) {
	b := &bytes.Buffer{}
	logger := lecho.New(b, lecho.WithLevel(zerolog.DebugLevel))

	assert.Equal(t, zerolog.DebugLevel, logger.Level())

	logger.SetLevel(zerolog.WarnLevel)

	assert.Equal(t, zerolog.WarnLevel, logger.Level())

	unwrapped := logger.Unwrap()
	unwrapped.Debug().Msg("debug")
	unwrapped.Warn().Msg("warn")

	assert.NotContains(t, b.String(), `"message":"debug"`)
	assert.Contains(t, b.String(), `"message":"warn"`)
}

func TestLoggerSetOutput(t *testing.T) {
	first := &bytes.Buffer{}
	second := &bytes.Buffer{}
	logger := lecho.New(
		first,
		lecho.WithLevel(zerolog.WarnLevel),
		lecho.WithField("logger", "lecho"),
	)

	before := logger.Unwrap()
	before.Warn().Msg("before")

	logger.SetOutput(second)

	assert.Equal(t, zerolog.WarnLevel, logger.Level())

	after := logger.Unwrap()
	after.Info().Msg("filtered")
	after.Warn().Msg("after")

	assert.Contains(t, first.String(), `"logger":"lecho"`)
	assert.Contains(t, first.String(), `"message":"before"`)
	assert.NotContains(t, first.String(), `"message":"after"`)
	assert.Contains(t, second.String(), `"logger":"lecho"`)
	assert.NotContains(t, second.String(), `"message":"filtered"`)
	assert.Contains(t, second.String(), `"message":"after"`)
}

func TestLoggerOutput(t *testing.T) {
	b := &bytes.Buffer{}
	logger := lecho.New(b, lecho.WithField("logger", "lecho"))

	written, err := io.WriteString(logger.Output(), "through output")

	assert.NoError(t, err)
	assert.Equal(t, len("through output"), written)
	assert.Contains(t, b.String(), `"logger":"lecho"`)
	assert.Contains(t, b.String(), `"message":"through output"`)
}

func TestLoggerSlogSnapshot(t *testing.T) {
	first := &bytes.Buffer{}
	second := &bytes.Buffer{}
	logger := lecho.New(first, lecho.WithLevel(zerolog.InfoLevel))
	snapshot := logger.Slog()

	logger.SetLevel(zerolog.ErrorLevel)
	logger.SetOutput(second)

	snapshot.Info("snapshot")
	logger.Slog().Info("filtered")
	logger.Slog().Error("current")

	assert.Contains(t, first.String(), `"message":"snapshot"`)
	assert.NotContains(t, first.String(), `"message":"current"`)
	assert.NotContains(t, second.String(), `"message":"filtered"`)
	assert.Contains(t, second.String(), `"message":"current"`)
}
