package lecho_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/ziflex/lecho/v3"
)

func TestMiddleware(t *testing.T) {
	t.Run("should not trigger error handler when HandleError is false", func(t *testing.T) {
		var called bool
		e := echo.New()
		e.HTTPErrorHandler = func(err error, c echo.Context) {
			called = true

			c.JSON(http.StatusInternalServerError, err.Error())
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		m := lecho.Middleware(lecho.Config{})

		next := func(c echo.Context) error {
			return errors.New("error")
		}

		handler := m(next)
		err := handler(c)

		assert.Error(t, err, "should return error")
		assert.False(t, called, "should not call error handler")
	})

	t.Run("should trigger error handler when HandleError is true", func(t *testing.T) {
		var called bool
		e := echo.New()
		e.HTTPErrorHandler = func(err error, c echo.Context) {
			called = true

			c.JSON(http.StatusInternalServerError, err.Error())
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		m := lecho.Middleware(lecho.Config{
			HandleError: true,
		})

		next := func(c echo.Context) error {
			return errors.New("error")
		}

		handler := m(next)
		err := handler(c)

		assert.Error(t, err, "should return error")
		assert.Truef(t, called, "should call error handler")
	})

	t.Run("should use enricher", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}

		l := lecho.New(b)
		m := lecho.Middleware(lecho.Config{
			Logger: l,
			Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
				return logger.Str("test", "test")
			},
		})

		next := func(c echo.Context) error {
			return nil
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Contains(t, str, `"test":"test"`)
	})

	t.Run("should use after next enricher", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}
		l := lecho.New(b)

		order := make([]string, 0, 2)
		var nextCalled bool

		m := lecho.Middleware(lecho.Config{
			Logger: l,
			AfterNextEnricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
				assert.True(t, nextCalled, "after next enricher should run after next")
				order = append(order, "after")

				return logger.Str("after", "yes")
			},
		})

		next := func(c echo.Context) error {
			nextCalled = true
			order = append(order, "next")

			return nil
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")
		assert.Equal(t, []string{"next", "after"}, order, "after next enricher should run after next")

		str := b.String()
		assert.Contains(t, str, `"after":"yes"`)
	})

	t.Run("should use after next enricher with context value", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}
		l := lecho.New(b)

		m := lecho.Middleware(lecho.Config{
			Logger: l,
			AfterNextEnricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
				// read value set by handler
				if v := c.Get("user_id"); v != nil {
					if userID, ok := v.(string); ok {
						return logger.Str("user_id", userID)
					}
				}
				return logger
			},
		})

		next := func(c echo.Context) error {
			// simulate middleware/handler adding context-specific info
			c.Set("user_id", "123")
			return nil
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Contains(t, str, `"user_id":"123"`)
	})
	t.Run("should escalate log level for slow requests", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}
		l := lecho.New(b)
		l.SetLevel(log.INFO)
		m := lecho.Middleware(lecho.Config{
			Logger:              l,
			RequestLatencyLimit: 5 * time.Millisecond,
			RequestLatencyLevel: zerolog.WarnLevel,
		})

		// Slow request should be logged at the escalated level
		next := func(c echo.Context) error {
			time.Sleep(5 * time.Millisecond)
			return nil
		}
		handler := m(next)
		err := handler(c)
		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Contains(t, str, `"level":"warn"`)
		assert.NotContains(t, str, `"level":"info"`)
	})

	t.Run("shouldn't escalate log level for fast requests", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}
		l := lecho.New(b)
		l.SetLevel(log.INFO)
		m := lecho.Middleware(lecho.Config{
			Logger:              l,
			RequestLatencyLimit: 5 * time.Millisecond,
			RequestLatencyLevel: zerolog.WarnLevel,
		})

		// Fast request should be logged at the default level
		next := func(c echo.Context) error {
			time.Sleep(1 * time.Millisecond)
			return nil
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Contains(t, str, `"level":"info"`)
		assert.NotContains(t, str, `"level":"warn"`)
	})

	t.Run("should skip middleware before calling next handler when Skipper func returns true", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/skip", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}
		l := lecho.New(b)
		l.SetLevel(log.INFO)
		m := lecho.Middleware(lecho.Config{
			Logger: l,
			Skipper: func(c echo.Context) bool {
				return c.Request().URL.Path == "/skip"
			},
		})

		next := func(c echo.Context) error {
			return nil
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Empty(t, str, "should not log anything")
	})

	t.Run("should skip middleware after calling next handler when AfterNextSkipper func returns true", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}
		l := lecho.New(b)
		l.SetLevel(log.INFO)
		m := lecho.Middleware(lecho.Config{
			Logger: l,
			AfterNextSkipper: func(c echo.Context) bool {
				return c.Response().Status == http.StatusMovedPermanently
			},
		})

		next := func(c echo.Context) error {
			return c.Redirect(http.StatusMovedPermanently, "/other")
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Empty(t, str, "should not log anything")
	})
}
