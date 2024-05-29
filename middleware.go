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
	// Config is the configuration for the middleware.
	Config struct {
		// Logger is a custom instance of the logger to use.
		Logger *Logger
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper
		// AfterNextSkipper defines a function to skip middleware after the next handler is called.
		AfterNextSkipper middleware.Skipper
		// BeforeNext is a function that is executed before the next handler is called.
		BeforeNext middleware.BeforeFunc
		// Enricher is a function that can be used to enrich the logger with additional information.
		Enricher Enricher
		// RequestIDHeader is the header name to use for the request ID in a log record.
		RequestIDHeader string
		// RequestIDKey is the key name to use for the request ID in a log record.
		RequestIDKey string
		// NestKey is the key name to use for the nested logger in a log record.
		NestKey string
		// HandleError indicates whether to propagate errors up the middleware chain, so the global error handler can decide appropriate status code.
		HandleError bool
		// For long-running requests that take longer than this limit, log at a different level.  Ignored by default
		RequestLatencyLimit time.Duration
		// The level to log at if RequestLatencyLimit is exceeded
		RequestLatencyLevel zerolog.Level
	}

	// Enricher is a function that can be used to enrich the logger with additional information.
	Enricher func(c echo.Context, logger zerolog.Context) zerolog.Context

	// Context is a wrapper around echo.Context that provides a logger.
	Context struct {
		echo.Context
		logger *Logger
	}
)

// NewContext returns a new Context.
func NewContext(ctx echo.Context, logger *Logger) *Context {
	return &Context{ctx, logger}
}

func (c *Context) Logger() echo.Logger {
	return c.logger
}

// Middleware returns a middleware which logs HTTP requests.
func Middleware(config Config) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}

	if config.AfterNextSkipper == nil {
		config.AfterNextSkipper = middleware.DefaultSkipper
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

			cloned := false
			logger := config.Logger

			if id != "" {
				logger = From(logger.log, WithField(config.RequestIDKey, id))
				cloned = true
			}

			if config.Enricher != nil {
				// to avoid mutation of shared instance
				if !cloned {
					logger = From(logger.log)
					cloned = true
				}

				logger.log = config.Enricher(c, logger.log.With()).Logger()
			}

			ctx := req.Context()

			if ctx == nil {
				ctx = context.Background()
			}

			// Pass logger down to request context
			c.SetRequest(req.WithContext(logger.WithContext(ctx)))
			c = NewContext(c, logger)

			if config.BeforeNext != nil {
				config.BeforeNext(c)
			}

			if err = next(c); err != nil {
				if config.HandleError {
					c.Error(err)
				}
			}

			if config.AfterNextSkipper(c) {
				return err
			}

			stop := time.Now()
			latency := stop.Sub(start)
			var mainEvt *zerolog.Event
			if err != nil {
				mainEvt = logger.log.Err(err)
			} else if config.RequestLatencyLimit != 0 && latency > config.RequestLatencyLimit {
				mainEvt = logger.log.WithLevel(config.RequestLatencyLevel)
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
			evt.Dur("latency", latency)
			evt.Str("latency_human", latency.String())

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
