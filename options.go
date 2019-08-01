package lecho

import (
	"io"

	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
)

type (
	Options struct {
		context zerolog.Context
		level log.Lvl
		prefix string
	}

	Setter func(opts *Options)
)

func newOptions(out io.Writer, setters []Setter) *Options {
	opts := &Options{
		context: zerolog.New(out).With(),
		level: log.INFO,
	}

	for _, set := range setters {
		set(opts)
	}

	return opts
}

func WithLevel(level log.Lvl) Setter {
	return func(opts *Options) {
		zlvl := levels[level]

		opts.context = opts.context.Logger().Level(zlvl).With()
		opts.level = level
	}
}

func WithFields(fields map[string]interface{}) Setter {
	return func(opts *Options) {
		opts.context = opts.context.Fields(fields)
	}
}

func WithTimestamp() Setter {
	return func(opts *Options) {
		opts.context = opts.context.Timestamp()
	}
}

func WithCaller() Setter {
	return func(opts *Options) {
		opts.context = opts.context.Caller()
	}
}

func WithCallerWithSkipFrameCount(skipFrameCount int) Setter {
	return func(opts *Options) {
		opts.context = opts.context.CallerWithSkipFrameCount(skipFrameCount)
	}
}

func WithPrefix(prefix string) Setter {
	return func(opts *Options) {
		opts.context = opts.context.Str("prefix", prefix)
		opts.prefix = prefix
	}
}

func WithHook(hook zerolog.Hook) Setter {
	return func(opts *Options) {
		opts.context = opts.context.Logger().Hook(hook).With()
	}
}

func WithHookFunc(hook zerolog.HookFunc) Setter {
	return func(opts *Options) {
		opts.context = opts.context.Logger().Hook(hook).With()
	}
}