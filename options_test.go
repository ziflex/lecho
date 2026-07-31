package lecho_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/ziflex/lecho/v4"
)

func TestOptions(t *testing.T) {
	t.Run("apply in order", func(t *testing.T) {
		var applied []string
		first := lecho.Option(func(logger zerolog.Logger) zerolog.Logger {
			applied = append(applied, "first")
			return logger
		})
		second := lecho.Option(func(logger zerolog.Logger) zerolog.Logger {
			applied = append(applied, "second")
			return logger
		})

		lecho.New(&bytes.Buffer{}, first, second)

		assert.Equal(t, []string{"first", "second"}, applied)
	})

	t.Run("preserve existing context", func(t *testing.T) {
		b := &bytes.Buffer{}
		base := zerolog.New(b).With().Str("existing", "yes").Logger()
		logger := lecho.From(base, lecho.WithField("added", "yes"))

		unwrapped := logger.Unwrap()
		unwrapped.Info().Msg("test")

		assert.Contains(t, b.String(), `"existing":"yes"`)
		assert.Contains(t, b.String(), `"added":"yes"`)
	})

	t.Run("level", func(t *testing.T) {
		b := &bytes.Buffer{}
		logger := lecho.New(b, lecho.WithLevel(zerolog.WarnLevel))
		unwrapped := logger.Unwrap()

		unwrapped.Debug().Msg("debug")
		unwrapped.Warn().Msg("warn")

		assert.Equal(t, zerolog.WarnLevel, unwrapped.GetLevel())
		assert.NotContains(t, b.String(), `"message":"debug"`)
		assert.Contains(t, b.String(), `"message":"warn"`)
	})

	t.Run("field", func(t *testing.T) {
		b := &bytes.Buffer{}
		logger := lecho.New(b, lecho.WithField("service", "logging"))

		unwrapped := logger.Unwrap()
		unwrapped.Info().Msg("test")

		assert.Contains(t, b.String(), `"service":"logging"`)
	})

	t.Run("fields", func(t *testing.T) {
		b := &bytes.Buffer{}
		logger := lecho.New(b, lecho.WithFields(map[string]any{
			"service": "logging",
			"version": 2,
		}))

		unwrapped := logger.Unwrap()
		unwrapped.Info().Msg("test")

		assert.Contains(t, b.String(), `"service":"logging"`)
		assert.Contains(t, b.String(), `"version":2`)
	})

	t.Run("timestamp", func(t *testing.T) {
		b := &bytes.Buffer{}
		logger := lecho.New(b, lecho.WithTimestamp())

		unwrapped := logger.Unwrap()
		unwrapped.Info().Msg("test")

		assert.Contains(t, b.String(), `"`+zerolog.TimestampFieldName+`":`)
	})

	t.Run("caller", func(t *testing.T) {
		b := &bytes.Buffer{}
		logger := lecho.New(b, lecho.WithCaller())

		unwrapped := logger.Unwrap()
		unwrapped.Info().Msg("test")

		assert.Contains(t, b.String(), `"`+zerolog.CallerFieldName+`":`)
		assert.Contains(t, b.String(), "options_test.go")
	})

	t.Run("caller with skip frame count", func(t *testing.T) {
		b := &bytes.Buffer{}
		logger := lecho.New(b, lecho.WithCallerWithSkipFrameCount(3))

		unwrapped := logger.Unwrap()
		unwrapped.Info().Msg("test")

		assert.Contains(t, b.String(), `"`+zerolog.CallerFieldName+`":`)
	})

	t.Run("prefix", func(t *testing.T) {
		b := &bytes.Buffer{}
		logger := lecho.New(b, lecho.WithPrefix("Test"))

		unwrapped := logger.Unwrap()
		unwrapped.Info().Msg("test")

		assert.Contains(t, b.String(), `"prefix":"Test"`)
	})
}

type recordingHook struct {
	levels []zerolog.Level
}

func (h *recordingHook) Run(_ *zerolog.Event, level zerolog.Level, _ string) {
	h.levels = append(h.levels, level)
}

func TestHookOptions(t *testing.T) {
	t.Run("hook", func(t *testing.T) {
		b := &bytes.Buffer{}
		hook := &recordingHook{}
		logger := lecho.New(b, lecho.WithHook(hook))
		unwrapped := logger.Unwrap()

		unwrapped.Info().Msg("info")
		unwrapped.Warn().Msg("warn")

		assert.Equal(t, []zerolog.Level{zerolog.InfoLevel, zerolog.WarnLevel}, hook.levels)
	})

	t.Run("hook function", func(t *testing.T) {
		b := &bytes.Buffer{}
		var messages []string
		logger := lecho.New(b, lecho.WithHookFunc(func(_ *zerolog.Event, _ zerolog.Level, message string) {
			messages = append(messages, message)
		}))
		unwrapped := logger.Unwrap()

		unwrapped.Info().Msg("info")
		unwrapped.Warn().Msg("warn")

		assert.Equal(t, []string{"info", "warn"}, messages)
	})
}

func TestOptionsRespectGlobalLevel(t *testing.T) {
	originalLevel := zerolog.GlobalLevel()
	defer zerolog.SetGlobalLevel(originalLevel)
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	b := &bytes.Buffer{}
	logger := lecho.New(b, lecho.WithLevel(zerolog.DebugLevel))
	unwrapped := logger.Unwrap()

	unwrapped.Debug().Msg("debug")
	unwrapped.Warn().Msg("warn")

	assert.False(t, strings.Contains(b.String(), `"message":"debug"`))
	assert.Contains(t, b.String(), `"message":"warn"`)
}
