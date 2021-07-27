package lecho

import (
	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
)

type (
	Options struct {
		context zerolog.Context
		level   log.Lvl
		prefix  string
	}

	Setter func(opts *Options)
)

func newOptions(log zerolog.Logger, setters []Setter) *Options {
	elvl, _ := MatchZeroLevel(log.GetLevel())

	opts := &Options{
		context: log.With(),
		level:   elvl,
	}

	for _, set := range setters {
		set(opts)
	}

	return opts
}

func WithLevel(level log.Lvl) Setter {
	return func(opts *Options) {
		zlvl, elvl := MatchEchoLevel(level)

		opts.context = opts.context.Logger().Level(zlvl).With()
		opts.level = elvl
	}
}

func WithField(name string, value interface{}) Setter {
	return func(opts *Options) {
		opts.context = opts.context.Interface(name, value)
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
