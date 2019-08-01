package lecho

import (
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type Config struct {
	Logger  *Logger
	Skipper middleware.Skipper
}

func Middleware(config Config) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}

	if config.Logger == nil {
		config.Logger = New(os.Stdout, WithTimestamp())
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

			if err = next(c); err != nil {
				c.Error(err)
			}

			stop := time.Now()

			ctx := config.Logger.log.Log()

			id := req.Header.Get(echo.HeaderXRequestID)

			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}

			if id != "" {
				ctx.Str("id", id)
			}

			ctx.Str("remote_ip", c.RealIP())
			ctx.Str("host", req.Host)
			ctx.Str("method", req.Method)
			ctx.Str("uri", req.RequestURI)
			ctx.Str("user_agent", req.UserAgent())
			ctx.Int("status", res.Status)
			ctx.Str("referer", req.Referer())

			if err != nil {
				ctx.Err(err)
			}

			ctx.Dur("latency", stop.Sub(start))
			ctx.Str("latency_human", stop.Sub(start).String())

			cl := req.Header.Get(echo.HeaderContentLength)
			if cl == "" {
				cl = "0"
			}

			ctx.Str("bytes_in", cl)
			ctx.Str("bytes_out", strconv.FormatInt(res.Size, 10))
			ctx.Msg("")

			return err
		}
	}
}
