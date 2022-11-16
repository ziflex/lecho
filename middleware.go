package lecho

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	Config struct {
		Logger          *Logger
		Skipper         middleware.Skipper
		RequestIDHeader string
		RequestIDKey    string
		NestKey         string
	}

	Context struct {
		echo.Context
		logger *Logger
	}
)

func NewContext(ctx echo.Context, logger *Logger) *Context {
	return &Context{ctx, logger}
}

func (c *Context) Logger() echo.Logger {
	return c.logger
}

func Middleware(config Config) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}

	if config.Logger == nil {
		config.Logger = New(os.Stdout, WithTimestamp())
	}

	if config.RequestIDKey == "" {
		config.RequestIDKey = "id"
	}

	if config.RequestIDHeader == "" {
		config.RequestIDHeader = echo.HeaderXRequestID
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			var err error
			req := c.Request()
			res := c.Response()
			start := time.Now()

			id := req.Header.Get(config.RequestIDHeader)

			if id == "" {
				id = res.Header().Get(config.RequestIDHeader)
			}

			logger := config.Logger

			if id != "" {
				logger = From(logger.log, WithField(config.RequestIDKey, id))
			}

			ctx := req.Context()

			if ctx == nil {
				ctx = context.Background()
			}

			// Pass logger down to request context
			c.SetRequest(req.WithContext(logger.WithContext(ctx)))

			c = NewContext(c, logger)

			if err = next(c); err != nil {
				c.Error(err)
			}

			stop := time.Now()

			var mainEvt *zerolog.Event
			if err != nil {
				mainEvt = logger.log.Err(err)
			} else {
				mainEvt = logger.log.WithLevel(logger.log.GetLevel())
			}

			var evt *zerolog.Event
			if config.NestKey != "" { // Start a new event (dict) if there's a nest key.
				evt = zerolog.Dict()
			} else {
				evt = mainEvt
			}
			evt.Str("remote_ip", c.RealIP())
			evt.Str("host", req.Host)
			evt.Str("method", req.Method)
			evt.Str("uri", req.RequestURI)
			evt.Str("user_agent", req.UserAgent())
			evt.Int("status", res.Status)
			evt.Str("referer", req.Referer())
			evt.Dur("latency", stop.Sub(start))
			evt.Str("latency_human", stop.Sub(start).String())

			cl := req.Header.Get(echo.HeaderContentLength)
			if cl == "" {
				cl = "0"
			}

			evt.Str("bytes_in", cl)
			evt.Str("bytes_out", strconv.FormatInt(res.Size, 10))

			if config.NestKey != "" { // Nest the new event (dict) under the nest key.
				mainEvt.Dict(config.NestKey, evt)
			}
			mainEvt.Send()

			return err
		}
	}
}
