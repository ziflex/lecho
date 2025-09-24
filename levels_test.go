package lecho_test

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/ziflex/lecho/v3"
)

func TestGetEffectiveZerologLevel(t *testing.T) {
	// Save original global level
	originalLevel := zerolog.GlobalLevel()
	defer zerolog.SetGlobalLevel(originalLevel)

	t.Run("should return global level when it's higher than logger level", func(t *testing.T) {
		// Set global level to WARN
		zerolog.SetGlobalLevel(zerolog.WarnLevel)

		// Create logger with TraceLevel (lower than global)
		buf := &bytes.Buffer{}
		logger := zerolog.New(buf)

		effective := lecho.GetEffectiveZerologLevel(logger)
		assert.Equal(t, zerolog.WarnLevel, effective)
	})

	t.Run("should return logger level when it's higher than global level", func(t *testing.T) {
		// Set global level to DEBUG
		zerolog.SetGlobalLevel(zerolog.DebugLevel)

		// Create logger with ErrorLevel (higher than global)
		buf := &bytes.Buffer{}
		logger := zerolog.New(buf).Level(zerolog.ErrorLevel)

		effective := lecho.GetEffectiveZerologLevel(logger)
		assert.Equal(t, zerolog.ErrorLevel, effective)
	})

	t.Run("should return same level when logger and global levels are equal", func(t *testing.T) {
		// Set global level to INFO
		zerolog.SetGlobalLevel(zerolog.InfoLevel)

		// Create logger with InfoLevel (same as global)
		buf := &bytes.Buffer{}
		logger := zerolog.New(buf).Level(zerolog.InfoLevel)

		effective := lecho.GetEffectiveZerologLevel(logger)
		assert.Equal(t, zerolog.InfoLevel, effective)
	})
}