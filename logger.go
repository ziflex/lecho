package lecho

import (
	"fmt"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
)

// Logger is a wrapper around `zerolog.Logger` that provides an implementation of `echo.Logger` interface
type Logger struct {
	echo.Logger
	log     zerolog.Logger
	out     io.Writer
	level   log.Lvl
	prefix  string
	setters []Setter
}

// New returns a new Logger instance
func New(out io.Writer, setters ...Setter) *Logger {
	opts := newOptions(out, setters)

	return &Logger{
		out:     out,
		log:     opts.context.Logger(),
		level:   opts.level,
		prefix:  opts.prefix,
		setters: setters,
	}
}

func (l Logger) Debug(i ...interface{}) {
	l.log.Debug().Msg(fmt.Sprint(i...))
}

func (l Logger) Debugf(format string, i ...interface{}) {
	l.log.Debug().Msgf(format, i...)
}

func (l Logger) Debugj(j log.JSON) {
	l.logJSON(l.log.Debug(), j)
}

func (l Logger) Info(i ...interface{}) {
	l.log.Info().Msg(fmt.Sprint(i...))
}

func (l Logger) Infof(format string, i ...interface{}) {
	l.log.Info().Msgf(format, i...)
}

func (l Logger) Infoj(j log.JSON) {
	l.logJSON(l.log.Info(), j)
}

func (l Logger) Warn(i ...interface{}) {
	l.log.Warn().Msg(fmt.Sprint(i...))
}

func (l Logger) Warnf(format string, i ...interface{}) {
	l.log.Warn().Msgf(format, i...)
}

func (l Logger) Warnj(j log.JSON) {
	l.logJSON(l.log.Warn(), j)
}

func (l Logger) Error(i ...interface{}) {
	l.log.Error().Msg(fmt.Sprint(i...))
}

func (l Logger) Errorf(format string, i ...interface{}) {
	l.log.Error().Msgf(format, i...)
}

func (l Logger) Errorj(j log.JSON) {
	l.logJSON(l.log.Error(), j)
}

func (l Logger) Fatal(i ...interface{}) {
	l.log.Fatal().Msg(fmt.Sprint(i...))
}

func (l Logger) Fatalf(format string, i ...interface{}) {
	l.log.Fatal().Msgf(format, i...)
}

func (l Logger) Fatalj(j log.JSON) {
	l.logJSON(l.log.Fatal(), j)
}

func (l Logger) Panic(i ...interface{}) {
	l.log.Panic().Msg(fmt.Sprint(i...))
}

func (l Logger) Panicf(format string, i ...interface{}) {
	l.log.Panic().Msgf(format, i...)
}

func (l Logger) Panicj(j log.JSON) {
	l.logJSON(l.log.Panic(), j)
}

func (l Logger) Print(i ...interface{}) {
	l.log.WithLevel(zerolog.NoLevel).Str("level", "-").Msg(fmt.Sprint(i...))
}

func (l Logger) Printf(format string, i ...interface{}) {
	l.log.WithLevel(zerolog.NoLevel).Str("level", "-").Msgf(format, i...)
}

func (l Logger) Printj(j log.JSON) {
	l.logJSON(l.log.WithLevel(zerolog.NoLevel).Str("level", "-"), j)
}

func (l Logger) Output() io.Writer {
	return l.out
}

func (l *Logger) SetOutput(newOut io.Writer) {
	l.out = newOut
	l.log = l.log.Output(newOut)
}

func (l Logger) Level() log.Lvl {
	return l.level
}

func (l *Logger) SetLevel(level log.Lvl) {
	zlvl, elvl := fromEchoLevel(level)

	l.setters = append(l.setters, WithLevel(elvl))
	l.level = elvl
	l.log = l.log.Level(zlvl)
}

func (l Logger) Prefix() string {
	return l.prefix
}

func (l Logger) SetHeader(h string) {
	// not implemented
}

func (l *Logger) SetPrefix(newPrefix string) {
	setters := append(l.setters, WithPrefix(newPrefix))
	opts := newOptions(l.out, setters)

	l.setters = setters
	l.prefix = newPrefix
	l.log = opts.context.Logger()
}

func (l Logger) Clone(setters ...Setter) *Logger {
	s := append(l.setters, setters...)

	return New(l.out, s...)
}

func (l Logger) logJSON(event *zerolog.Event, j log.JSON) {
	for k, v := range j {
		event = event.Interface(k, v)
	}

	event.Msg("")
}
