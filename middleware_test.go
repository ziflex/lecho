package lecho_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/ziflex/lecho/v4"
)

func TestMiddleware(t *testing.T) {
	t.Run("should not trigger error handler when HandleError is false", func(t *testing.T) {
		var called bool
		e := echo.New()
		e.HTTPErrorHandler = func(c *echo.Context, err error) {
			called = true

			c.JSON(http.StatusInternalServerError, err.Error())
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		m := lecho.Middleware(lecho.Config{})

		next := func(c *echo.Context) error {
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
		e.HTTPErrorHandler = func(c *echo.Context, err error) {
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

		next := func(c *echo.Context) error {
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
			Enricher: func(c *echo.Context, logger zerolog.Context) zerolog.Context {
				return logger.Str("test", "test")
			},
		})

		next := func(c *echo.Context) error {
			return nil
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Contains(t, str, `"test":"test"`)
	})

	t.Run("should expose the enriched logger through Echo and request context", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}
		l := lecho.New(b)
		m := lecho.Middleware(lecho.Config{
			Logger: l,
			Enricher: func(c *echo.Context, logger zerolog.Context) zerolog.Context {
				return logger.Str("request_scope", "yes")
			},
		})

		next := func(c *echo.Context) error {
			c.Logger().Info("from Echo", "source", "slog")
			lecho.Ctx(c.Request().Context()).
				Info().
				Str("source", "request_context").
				Msg("from Zerolog")

			return c.String(http.StatusCreated, "hello")
		}

		err := m(next)(c)

		assert.NoError(t, err)
		str := b.String()
		assert.Contains(t, str, `"request_scope":"yes","source":"slog"`)
		assert.Contains(t, str, `"message":"from Echo"`)
		assert.Contains(t, str, `"request_scope":"yes","source":"request_context"`)
		assert.Contains(t, str, `"message":"from Zerolog"`)
		assert.Contains(t, str, `"status":201`)
		assert.Contains(t, str, `"bytes_out":"5"`)
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
			AfterNextEnricher: func(c *echo.Context, logger zerolog.Context) zerolog.Context {
				assert.True(t, nextCalled, "after next enricher should run after next")
				order = append(order, "after")

				return logger.Str("after", "yes")
			},
		})

		next := func(c *echo.Context) error {
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

	t.Run("should use enricher and after next enricher together", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}
		l := lecho.New(b)

		order := make([]string, 0, 3)
		var beforeCalled bool
		var nextCalled bool
		var afterCalled bool

		m := lecho.Middleware(lecho.Config{
			Logger: l,
			Enricher: func(c *echo.Context, logger zerolog.Context) zerolog.Context {
				beforeCalled = true
				order = append(order, "before")

				return logger.Str("before", "yes")
			},
			AfterNextEnricher: func(c *echo.Context, logger zerolog.Context) zerolog.Context {
				afterCalled = true
				// AfterNextEnricher should run after the next handler and after the pre-handler enricher.
				assert.True(t, beforeCalled, "pre-handler enricher should have been called before after next enricher")
				assert.True(t, nextCalled, "next should run before after next enricher")
				order = append(order, "after")

				return logger.Str("after", "yes")
			},
		})

		next := func(c *echo.Context) error {
			nextCalled = true
			order = append(order, "next")

			return nil
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")
		assert.True(t, beforeCalled, "pre-handler enricher should be called")
		assert.True(t, nextCalled, "next handler should be called")
		assert.True(t, afterCalled, "after next enricher should be called")
		assert.Equal(t, []string{"before", "next", "after"}, order, "enrichers and next should be called in the correct order")

		str := b.String()
		assert.Contains(t, str, `"before":"yes"`)
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
			AfterNextEnricher: func(c *echo.Context, logger zerolog.Context) zerolog.Context {
				// read value set by handler
				if v := c.Get("user_id"); v != nil {
					if userID, ok := v.(string); ok {
						return logger.Str("user_id", userID)
					}
				}
				return logger
			},
		})

		next := func(c *echo.Context) error {
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
		l := lecho.New(b, lecho.WithLevel(zerolog.InfoLevel))
		m := lecho.Middleware(lecho.Config{
			Logger:              l,
			RequestLatencyLimit: 5 * time.Millisecond,
			RequestLatencyLevel: zerolog.WarnLevel,
		})

		// Slow request should be logged at the escalated level
		next := func(c *echo.Context) error {
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
		l := lecho.New(b, lecho.WithLevel(zerolog.InfoLevel))
		m := lecho.Middleware(lecho.Config{
			Logger:              l,
			RequestLatencyLimit: 5 * time.Millisecond,
			RequestLatencyLevel: zerolog.WarnLevel,
		})

		// Fast request should be logged at the default level
		next := func(c *echo.Context) error {
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
		l := lecho.New(b, lecho.WithLevel(zerolog.InfoLevel))
		m := lecho.Middleware(lecho.Config{
			Logger: l,
			Skipper: func(c *echo.Context) bool {
				return c.Request().URL.Path == "/skip"
			},
		})

		next := func(c *echo.Context) error {
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
		l := lecho.New(b, lecho.WithLevel(zerolog.InfoLevel))
		m := lecho.Middleware(lecho.Config{
			Logger: l,
			AfterNextSkipper: func(c *echo.Context) bool {
				response, err := echo.UnwrapResponse(c.Response())
				return err == nil && response.Status == http.StatusMovedPermanently
			},
		})

		next := func(c *echo.Context) error {
			return c.Redirect(http.StatusMovedPermanently, "/other")
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Empty(t, str, "should not log anything")
	})

	t.Run("should use default attributes", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}

		l := lecho.New(b)
		m := lecho.Middleware(lecho.Config{
			Logger: l,
		})

		next := func(c *echo.Context) error {
			return nil
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Contains(t, str, `"method":"GET"`)
	})
	t.Run("should use request ID and nest default attributes", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderXRequestID, "request-123")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}
		l := lecho.New(b)
		m := lecho.Middleware(lecho.Config{
			Logger:  l,
			NestKey: "request",
		})

		err := m(func(c *echo.Context) error {
			return c.NoContent(http.StatusAccepted)
		})(c)

		assert.NoError(t, err)
		str := b.String()
		assert.Contains(t, str, `"id":"request-123"`)
		assert.Contains(t, str, `"request":{"remote_ip":`)
		assert.Contains(t, str, `"status":202`)
		assert.Contains(t, str, `"bytes_out":"0"`)
	})
	t.Run("should skip default attributes", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}

		l := lecho.New(b)
		m := lecho.Middleware(lecho.Config{
			Logger:            l,
			SkipDefaultFields: true,
			Enricher: func(c *echo.Context, logger zerolog.Context) zerolog.Context {
				val := map[string]any{
					"http.request.method": c.Request().Method,
				}
				return logger.Fields(val)
			},
		})

		next := func(c *echo.Context) error {
			return nil
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Contains(t, str, `"http.request.method":"GET"`)
		assert.NotContains(t, str, `"method":"GET"`)
	})
	t.Run("should skip default attributes and not consider NestKey", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}

		l := lecho.New(b)
		m := lecho.Middleware(lecho.Config{
			Logger:            l,
			SkipDefaultFields: true,
			NestKey:           "nested",
			Enricher: func(c *echo.Context, logger zerolog.Context) zerolog.Context {
				val := map[string]any{
					"http.request.method": c.Request().Method,
				}
				return logger.Fields(val)
			},
		})

		next := func(c *echo.Context) error {
			return nil
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Contains(t, str, `"http.request.method":"GET"`)
		assert.NotContains(t, str, `"nested.method":"GET"`)
		assert.NotContains(t, str, `"nested"`)
	})
	t.Run("should skip default attributes and not consider RequestID", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("myRequestID", "my.request.id")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}

		l := lecho.New(b)
		m := lecho.Middleware(lecho.Config{
			Logger:            l,
			SkipDefaultFields: true,
			RequestIDHeader:   "myRequestID",
			RequestIDKey:      "my.request.id",
			Enricher: func(c *echo.Context, logger zerolog.Context) zerolog.Context {
				val := map[string]any{
					"http.request.method": c.Request().Method,
				}
				return logger.Fields(val)
			},
		})

		next := func(c *echo.Context) error {
			return nil
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Contains(t, str, `"http.request.method":"GET"`)
		assert.NotContains(t, str, `"my.request.id":"myRequestID"`)
		assert.NotContains(t, str, `"my.request.id"`)
	})
}
