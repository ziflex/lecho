package lecho

import (
	"strconv"
	"time"

	"github.com/labstack/echo"
)

func Middleware(logger *Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var err error
			req := c.Request()
			res := c.Response()
			start := time.Now()

			if err = next(c); err != nil {
				c.Error(err)
			}

			params := make(map[string]interface{})
			stop := time.Now()

			id := req.Header.Get(echo.HeaderXRequestID)

			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}

			if id != "" {
				params["id"] = id
			}

			params["remote_ip"] = c.RealIP()
			params["host"] = req.Host
			params["method"] = req.Method
			params["uri"] = req.RequestURI
			params["user_agent"] = req.UserAgent()
			params["status"] = res.Status
			params["referer"] = req.Referer()

			if err != nil {
				params["error"] = err
			}

			params["latency"] = stop.Sub(start)
			params["latency_human"] = stop.Sub(start).String()

			cl := req.Header.Get(echo.HeaderContentLength)
			if cl == "" {
				cl = "0"
			}

			params["bytes_in"] = cl
			params["bytes_out"] = strconv.FormatInt(res.Size, 10)

			logger.Printj(params)

			return err
		}
	}
}
