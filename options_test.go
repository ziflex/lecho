package lecho_test

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/ziflex/lecho/v3"
)

func TestWithCaller(t *testing.T) {
	b := &bytes.Buffer{}

	l := lecho.New(b, lecho.WithCaller())

	l.Print("foobar")

	type Log struct {
		Level   string `json:"level"`
		Caller  string `json:"caller"`
		Message string `json:"message"`
	}

	log := &Log{}

	err := json.Unmarshal(b.Bytes(), log)

	assert.NoError(t, err)

	segments := strings.Split(log.Caller, ":")
	filePath := filepath.Base(segments[0])

	assert.Equal(t, filePath, "logger.go")
}

func TestWithCallerWithSkipFrameCount(t *testing.T) {
	b := &bytes.Buffer{}

	l := lecho.New(b, lecho.WithCallerWithSkipFrameCount(3))

	l.Print("foobar")

	type Log struct {
		Level   string `json:"level"`
		Caller  string `json:"caller"`
		Message string `json:"message"`
	}

	log := &Log{}

	err := json.Unmarshal(b.Bytes(), log)

	assert.NoError(t, err)

	segments := strings.Split(log.Caller, ":")
	filePath := filepath.Base(segments[0])

	assert.Equal(t, filePath, "options_test.go")
}

func TestWithField(t *testing.T) {
	b := &bytes.Buffer{}

	l := lecho.New(b, lecho.WithField("service", "logging"))

	l.Print("foobar")

	type Log struct {
		Level   string `json:"level"`
		Service string `json:"service"`
		Message string `json:"message"`
	}

	log := &Log{}

	err := json.Unmarshal(b.Bytes(), log)

	assert.NoError(t, err)
	assert.Equal(t, log.Service, "logging")
}

func TestWithFields(t *testing.T) {
	b := &bytes.Buffer{}

	l := lecho.New(b, lecho.WithFields(map[string]interface{}{
		"host": "localhost",
		"port": 8080,
	}))

	l.Print("foobar")

	type Log struct {
		Level   string `json:"level"`
		Host    string `json:"host"`
		Port    int    `json:"port"`
		Message string `json:"message"`
	}

	log := &Log{}

	err := json.Unmarshal(b.Bytes(), log)

	assert.NoError(t, err)
	assert.Equal(t, log.Host, "localhost")
	assert.Equal(t, log.Port, 8080)
}

type (
	Hook struct {
		logs []HookLog
	}

	HookLog struct {
		level   zerolog.Level
		message string
	}
)

func (h *Hook) Run(e *zerolog.Event, level zerolog.Level, message string) {
	h.logs = append(h.logs, HookLog{
		level:   level,
		message: message,
	})
}

func TestWithHook(t *testing.T) {
	b := &bytes.Buffer{}
	h := &Hook{}
	l := lecho.New(b, lecho.WithHook(h))

	l.Info("Foo")
	l.Warn("Bar")

	assert.Len(t, h.logs, 2)
	assert.Equal(t, h.logs[0].level, zerolog.InfoLevel)
	assert.Equal(t, h.logs[0].message, "Foo")
	assert.Equal(t, h.logs[1].level, zerolog.WarnLevel)
	assert.Equal(t, h.logs[1].message, "Bar")
}

func TestWithHookFunc(t *testing.T) {
	b := &bytes.Buffer{}
	logs := make([]HookLog, 0, 2)
	l := lecho.New(b, lecho.WithHookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
		logs = append(logs, HookLog{
			level:   level,
			message: message,
		})
	}))

	l.Info("Foo")
	l.Warn("Bar")

	assert.Len(t, logs, 2)
	assert.Equal(t, logs[0].level, zerolog.InfoLevel)
	assert.Equal(t, logs[0].message, "Foo")
	assert.Equal(t, logs[1].level, zerolog.WarnLevel)
	assert.Equal(t, logs[1].message, "Bar")
}

func TestWithLevel(t *testing.T) {
	b := &bytes.Buffer{}
	l := lecho.New(b, lecho.WithLevel(log.WARN))

	l.Debug("Test")

	assert.Equal(t, b.String(), "")

	l.Warn("Foobar")

	assert.Equal(t, b.String(), `{"level":"warn","message":"Foobar"}
`)
}

func TestWithPrefix(t *testing.T) {
	b := &bytes.Buffer{}
	l := lecho.New(b, lecho.WithPrefix("Test"))

	l.Warn("Foobar")

	assert.Equal(t, b.String(), `{"level":"warn","prefix":"Test","message":"Foobar"}
`)
}

func TestWithTimestamp(t *testing.T) {
	b := &bytes.Buffer{}

	l := lecho.New(b, lecho.WithTimestamp())

	l.Print("foobar")

	type Log struct {
		Level   string    `json:"level"`
		Message string    `json:"message"`
		Time    time.Time `json:"time"`
	}

	log := &Log{}

	err := json.Unmarshal(b.Bytes(), log)

	assert.NoError(t, err)
	assert.NotEmpty(t, log.Time)
}

func TestGlobalLevel(t *testing.T) {
	// Save original global level
	originalLevel := zerolog.GlobalLevel()
	defer zerolog.SetGlobalLevel(originalLevel)

	t.Run("should respect GlobalLevel when creating logger from existing zerolog", func(t *testing.T) {
		// Set global level to WARN
		zerolog.SetGlobalLevel(zerolog.WarnLevel)

		b := &bytes.Buffer{}
		// Create a zerolog with default TraceLevel
		zl := zerolog.New(b)

		// Create lecho logger from existing zerolog
		l := lecho.From(zl)

		// The effective level should be WARN (from GlobalLevel), not DEBUG (from TraceLevel)
		assert.Equal(t, log.WARN, l.Level())

		// Debug should not log anything
		l.Debug("debug message")
		assert.Equal(t, "", b.String())

		// Warn should log
		l.Warn("warn message")
		assert.Contains(t, b.String(), "warn message")
	})

	t.Run("should use logger level when GlobalLevel is lower", func(t *testing.T) {
		// Set global level to DEBUG
		zerolog.SetGlobalLevel(zerolog.DebugLevel)

		b := &bytes.Buffer{}
		// Create a zerolog with ErrorLevel (higher than global)
		zl := zerolog.New(b).Level(zerolog.ErrorLevel)

		// Create lecho logger from existing zerolog
		l := lecho.From(zl)

		// The effective level should be ERROR (from logger), not DEBUG (from GlobalLevel)
		assert.Equal(t, log.ERROR, l.Level())

		// Warn should not log anything
		l.Warn("warn message")
		assert.Equal(t, "", b.String())

		// Error should log
		l.Error("error message")
		assert.Contains(t, b.String(), "error message")
	})
}
