package lecho

import (
	"context"
	"github.com/rs/zerolog"
)

func (l *Logger) WithContext(ctx context.Context) context.Context {
	zerologger := l.Unwrap()
	return zerologger.WithContext(ctx)
}

func Ctx(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}
