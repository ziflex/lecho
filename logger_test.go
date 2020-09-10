package lecho_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/ziflex/lecho/v2"
)

func TestNew(t *testing.T) {
	b := &bytes.Buffer{}

	l := lecho.New(b)

	l.Print("foo")

	assert.Equal(
		t,
		`{"level":"-","message":"foo"}
`,
		b.String(),
	)
}

func TestLogger_Output(t *testing.T) {
	out1 := &bytes.Buffer{}

	l := lecho.New(out1)

	l.Print("foo")
	l.Print("bar")

	out2 := &bytes.Buffer{}
	l.SetOutput(out2)

	l.Print("baz")

	assert.Equal(
		t,
		`{"level":"-","message":"foo"}
{"level":"-","message":"bar"}
`,
		out1.String(),
	)

	assert.Equal(
		t,
		`{"level":"-","message":"baz"}
`,
		out2.String(),
	)
}

func TestLogger_SetLevel(t *testing.T) {
	b := &bytes.Buffer{}

	l := lecho.New(b)

	l.Debug("foo")

	assert.Equal(
		t,
		`{"level":"debug","message":"foo"}
`,
		b.String(),
	)

	b.Reset()

	l.SetLevel(log.WARN)

	l.Debug("foo")

	assert.Equal(t, "", b.String())
}

func TestLogger_Clone(t *testing.T) {
	b := &bytes.Buffer{}

	l := lecho.New(b)

	l.SetPrefix("L1")

	l.Debug("foo")

	assert.Equal(t, `{"level":"debug","prefix":"L1","message":"foo"}
`, b.String())

	b.Reset()

	l2 := l.Clone()
	l2.SetPrefix("L2")
	l2.Debug("foo")

	assert.Equal(t, `{"level":"debug","prefix":"L2","message":"foo"}
`, b.String())
}

func TestLogger_Debugf(t *testing.T) {
	b := &bytes.Buffer{}
	l := lecho.New(b)

	l.Debugf("Foo%s", "Bar")

	assert.Equal(t, `{"level":"debug","message":"FooBar"}
`,
		b.String())
}

func TestLogger_Debugj(t *testing.T) {
	b := &bytes.Buffer{}
	l := lecho.New(b)

	l.Debugj(log.JSON{
		"message": "foobar",
	})

	assert.Equal(t, `{"level":"debug","message":"foobar"}
`,
		b.String())
}

func TestLogger(t *testing.T) {
	type (
		SimpleLog struct {
			Level zerolog.Level
			Fn    func(i ...interface{})
		}

		FormattedLog struct {
			Level zerolog.Level
			Fn    func(format string, i ...interface{})
		}

		JSONLog struct {
			Level zerolog.Level
			Fn    func(j log.JSON)
		}
	)

	b := &bytes.Buffer{}
	l := lecho.New(b)

	simpleLogs := []SimpleLog{
		{
			Level: zerolog.DebugLevel,
			Fn:    l.Debug,
		},
		{
			Level: zerolog.InfoLevel,
			Fn:    l.Info,
		},
		{
			Level: zerolog.WarnLevel,
			Fn:    l.Warn,
		},
		{
			Level: zerolog.ErrorLevel,
			Fn:    l.Error,
		},
	}

	for _, l := range simpleLogs {
		b.Reset()

		l.Fn("foobar")
		assert.Equal(t, fmt.Sprintf(`{"level":"%s","message":"foobar"}
`, l.Level),
			b.String())
	}

	formattedLogs := []FormattedLog{
		{
			Level: zerolog.DebugLevel,
			Fn:    l.Debugf,
		},
		{
			Level: zerolog.InfoLevel,
			Fn:    l.Infof,
		},
		{
			Level: zerolog.WarnLevel,
			Fn:    l.Warnf,
		},
		{
			Level: zerolog.ErrorLevel,
			Fn:    l.Errorf,
		},
	}

	for _, l := range formattedLogs {
		b.Reset()

		l.Fn("foo%s", "bar")
		assert.Equal(t, fmt.Sprintf(`{"level":"%s","message":"foobar"}
`, l.Level),
			b.String())
	}

	jsonLogs := []JSONLog{
		{
			Level: zerolog.DebugLevel,
			Fn:    l.Debugj,
		},
		{
			Level: zerolog.InfoLevel,
			Fn:    l.Infoj,
		},
		{
			Level: zerolog.WarnLevel,
			Fn:    l.Warnj,
		},
		{
			Level: zerolog.ErrorLevel,
			Fn:    l.Errorj,
		},
	}

	for _, l := range jsonLogs {
		b.Reset()

		l.Fn(log.JSON{
			"message": "foobar",
		})
		assert.Equal(t, fmt.Sprintf(`{"level":"%s","message":"foobar"}
`, l.Level),
			b.String())
	}
}
