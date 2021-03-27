package lecho

import (
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	Config struct {
		Logger       *Logger
		Skipper      middleware.Skipper
		RequestIDKey string
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

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			var err error
			req := c.Request()
			res := c.Response()
			start := time.Now()

			id := req.Header.Get(echo.HeaderXRequestID)

			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}

			logger := config.Logger

			if id != "" {
				logger = From(logger.log, WithField(config.RequestIDKey, id))
			}

			c = NewContext(c, logger)

			if err = next(c); err != nil {
				c.Error(err)
			}

			stop := time.Now()

			evt := logger.log.Info()
			evt.Str("remote_ip", c.RealIP())
			evt.Str("host", req.Host)
			evt.Str("method", req.Method)
			evt.Str("uri", req.RequestURI)
			evt.Str("user_agent", req.UserAgent())
			evt.Int("status", res.Status)
			evt.Str("referer", req.Referer())

			if err != nil {
				evt.Err(err)
			}

			evt.Dur("latency", stop.Sub(start))
			evt.Str("latency_human", stop.Sub(start).String())

			cl := req.Header.Get(echo.HeaderContentLength)
			if cl == "" {
				cl = "0"
			}

			evt.Str("bytes_in", cl)
			evt.Str("bytes_out", strconv.FormatInt(res.Size, 10))
			evt.Send()

			return err
		}
	}
}
