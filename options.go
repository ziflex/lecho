package lecho

import "github.com/rs/zerolog"

// Option configures a zerolog logger during construction.
type Option func(zerolog.Logger) zerolog.Logger

// WithLevel sets the logger level.
func WithLevel(level zerolog.Level) Option {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.Level(level)
	}
}

// WithField adds a field to every log event.
func WithField(name string, value any) Option {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.With().Interface(name, value).Logger()
	}
}

// WithFields adds fields to every log event.
func WithFields(fields map[string]any) Option {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.With().Fields(fields).Logger()
	}
}

// WithTimestamp adds a timestamp to every log event.
func WithTimestamp() Option {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.With().Timestamp().Logger()
	}
}

// WithCaller adds caller information to every log event.
func WithCaller() Option {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.With().Caller().Logger()
	}
}

// WithCallerWithSkipFrameCount adds caller information with a custom skip frame count.
func WithCallerWithSkipFrameCount(skipFrameCount int) Option {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.With().CallerWithSkipFrameCount(skipFrameCount).Logger()
	}
}

// WithPrefix adds a prefix field to every log event.
func WithPrefix(prefix string) Option {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.With().Str("prefix", prefix).Logger()
	}
}

// WithHook adds a hook to the logger.
func WithHook(hook zerolog.Hook) Option {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.Hook(hook)
	}
}

// WithHookFunc adds a hook function to the logger.
func WithHookFunc(hook zerolog.HookFunc) Option {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.Hook(hook)
	}
}
