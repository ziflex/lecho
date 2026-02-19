# lecho :tomato:

A high-performance [Zerolog](https://github.com/rs/zerolog) wrapper for [Echo](https://echo.labstack.com/) web framework that provides structured logging with minimal overhead.

## Features

- 🚀 **High Performance** - Built on top of zerolog, one of the fastest structured loggers for Go
- 📊 **Structured Logging** - JSON formatted logs with rich contextual information
- 🔗 **Request Correlation** - Automatic request ID tracking and context propagation
- ⚡ **Low Latency** - Minimal overhead request/response logging
- 🛠 **Highly Configurable** - Extensive middleware configuration options
- 🔍 **Request Enrichment** - Add custom fields based on request context
- 📈 **Performance Monitoring** - Built-in slow request detection and alerting

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration Options](#options)
- [Middleware](#middleware)
- [Helpers](#helpers)
- [Advanced Configuration](#advanced-configuration)

## Installation

Install lecho based on your Echo version:

**For Echo v4 (recommended):**
```bash
go get github.com/ziflex/lecho/v3
```

**For Echo v3 (legacy):**
```bash
go get github.com/ziflex/lecho
```

## Quick Start

### Basic Usage

Replace Echo's default logger with lecho for structured logging:

```go
package main

import (
	"os"
	"github.com/labstack/echo/v4"
	"github.com/ziflex/lecho/v3"
)

func main() {
	e := echo.New()
	e.Logger = lecho.New(os.Stdout)
	
	// Your routes and middleware here
	e.Start(":8080")
}
```

### Using Existing Zerolog Instance

If you already have a zerolog logger configured, you can wrap it with lecho:

```go
package main

import (
	"os"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
)

func main() {
	// Configure your zerolog instance
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	
	e := echo.New()
	e.Logger = lecho.From(log)
	
	e.Start(":8080")
}
```

## Options

Lecho provides several configuration options to customize logging behavior:

```go
package main

import (
	"os"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/ziflex/lecho/v3"
)

func main() {
	e := echo.New()
	e.Logger = lecho.New(
		os.Stdout,
		lecho.WithLevel(log.DEBUG),                                    // Set log level
		lecho.WithFields(map[string]interface{}{"service": "api"}),    // Add default fields
		lecho.WithTimestamp(),                                         // Add timestamp to logs
		lecho.WithCaller(),                                           // Add caller information
		lecho.WithPrefix("MyApp"),                                    // Add a prefix to logs
		// lecho.WithHook(myHook),                                    // Add custom hooks
		// lecho.WithHookFunc(myHookFunc),                            // Add hook functions
	)
	
	e.Start(":8080")
}
```

### Available Options

- **`WithLevel(level log.Lvl)`** - Set the minimum log level (DEBUG, INFO, WARN, ERROR, OFF)
- **`WithFields(fields map[string]interface{})`** - Add default fields to all log entries
- **`WithField(key string, value interface{})`** - Add a single default field
- **`WithTimestamp()`** - Include timestamp in log entries
- **`WithCaller()`** - Include caller file and line information
- **`WithCallerWithSkipFrameCount(count int)`** - Include caller info with custom skip frame count
- **`WithPrefix(prefix string)`** - Add a prefix field to all log entries
- **`WithHook(hook zerolog.Hook)`** - Add a custom zerolog hook
- **`WithHookFunc(hookFunc zerolog.HookFunc)`** - Add a custom hook function

## Middleware

The lecho middleware provides automatic request logging with rich contextual information. It integrates seamlessly with Echo's request lifecycle and supports various customization options.

### Basic Request Logging

```go
package main

import (
	"net/http"
	"os"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
)

func main() {
	e := echo.New()
	
	// Create and configure logger
	logger := lecho.New(
		os.Stdout,
		lecho.WithLevel(log.DEBUG),
		lecho.WithTimestamp(),
		lecho.WithCaller(),
	)
	e.Logger = logger
	
	// Add request ID middleware (optional but recommended)
	e.Use(middleware.RequestID())
	
	// Add lecho middleware for request logging
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
	}))
	
	// Example route
	e.GET("/", func(c echo.Context) error {
		// Log using Echo's logger interface
		c.Logger().Info("Processing request")
		
		// Or use zerolog directly from context
		zerolog.Ctx(c.Request().Context()).Info().Msg("Using zerolog interface")
		
		return c.String(http.StatusOK, "Hello, World!")
	})
	
	e.Start(":8080")
}
```

**Sample output:**
```json
{"level":"info","id":"123e4567-e89b-12d3-a456-426614174000","remote_ip":"127.0.0.1","host":"localhost:8080","method":"GET","uri":"/","user_agent":"curl/7.68.0","status":200,"referer":"","latency":1.234,"latency_human":"1.234ms","bytes_in":"0","bytes_out":"13","time":"2023-10-15T10:30:00Z"}
```

### Escalate Log Level for Slow Requests

Monitor and highlight slow requests by logging them at a higher level when they exceed a specified duration:

```go
package main

import (
	"os"
	"time"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
)

func main() {
	e := echo.New()
	logger := lecho.New(os.Stdout, lecho.WithTimestamp())
	
	e.Use(lecho.Middleware(lecho.Config{
		Logger:              logger,
		RequestLatencyLevel: zerolog.WarnLevel,        // Log level for slow requests
		RequestLatencyLimit: 500 * time.Millisecond,  // Threshold for slow requests
	}))
	
	// Requests taking longer than 500ms will be logged at WARN level
	// instead of the default INFO level
}
```

**Output for slow request:**
```json
{"level":"warn","remote_ip":"127.0.0.1","method":"GET","uri":"/slow","status":200,"latency":750.123,"latency_human":"750.123ms","time":"2023-10-15T10:30:00Z"}
```


### Nesting Under a Sub Dictionary

Organize request information under a nested key for better log structure:

```go
package main

import (
	"os"
	"github.com/labstack/echo/v4"
	"github.com/ziflex/lecho/v3"
)

func main() {
	e := echo.New()
	logger := lecho.New(os.Stdout, lecho.WithTimestamp())
	
	e.Use(lecho.Middleware(lecho.Config{
		Logger:  logger,
		NestKey: "request", // Nest all request info under this key
	}))
	
	// All request-related fields will be nested under "request"
}
```

**Sample output:**
```json
{"level":"info","request":{"remote_ip":"127.0.0.1","method":"GET","uri":"/api/users","status":200,"latency":15.234,"latency_human":"15.234ms"},"time":"2023-10-15T10:30:00Z"}
```

### Enricher

The Enricher function allows you to add custom fields to log entries based on request context. This is useful for adding user IDs, trace IDs, or other contextual information:

```go
package main

import (
	"os"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
)

func main() {
	e := echo.New()
	logger := lecho.New(os.Stdout, lecho.WithTimestamp())
	
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
		Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
			// Add user ID if available in context
			if userID := c.Get("user_id"); userID != nil {
				logger = logger.Str("user_id", userID.(string))
			}
			
			// Add trace ID from header
			if traceID := c.Request().Header.Get("X-Trace-ID"); traceID != "" {
				logger = logger.Str("trace_id", traceID)
			}
			
			return logger
		},
	}))
	
	// Set up routes that use user context
	e.GET("/api/profile", func(c echo.Context) error {
		c.Set("user_id", "user123") // This will be logged
		return c.JSON(200, map[string]string{"status": "ok"})
	})
}
```

#### AfterNextEnricher

The `AfterNextEnricher` function allows you to add custom fields to log entries after the next handler has executed. This is useful for adding fields that depend on the outcome of the request processing, such as response status or values set during request handling.

```go
package main

import (
    "os"
    "github.com/labstack/echo/v4"
    "github.com/rs/zerolog"
    "github.com/ziflex/lecho/v3"
)

func main() {
    e := echo.New()
    logger := lecho.New(os.Stdout, lecho.WithTimestamp())
    
    e.Use(lecho.Middleware(lecho.Config{
        Logger: logger,
        AfterNextEnricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
            // Add response status code after the handler has executed
            logger = logger.Int("status", c.Response().Status)
            return logger
        },
    }))
    
    e.GET("/api/profile", func(c echo.Context) error {
        c.Set("user_id", "user123") // This will be logged
        return c.JSON(200, map[string]string{"status": "ok"})
    })
}
```

**Sample output:**
```json
{"level":"info","user_id":"user123","trace_id":"abc-def-123","remote_ip":"127.0.0.1","method":"GET","uri":"/api/profile","status":200,"time":"2023-10-15T10:30:00Z"}
```

### Error Handling

Control how errors are handled in the middleware chain. By default, lecho logs errors but doesn't propagate them to Echo's error handler:

```go
package main

import (
	"errors"
	"net/http"
	"os"
	"github.com/labstack/echo/v4"
	"github.com/ziflex/lecho/v3"
)

func main() {
	e := echo.New()
	logger := lecho.New(os.Stdout, lecho.WithTimestamp())
	
	// Configure error handling
	e.Use(lecho.Middleware(lecho.Config{
		Logger:      logger,
		HandleError: true, // Propagate errors to Echo's error handler
	}))
	
	// Custom error handler
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		c.JSON(code, map[string]string{"error": err.Error()})
	}
	
	// Route that may return an error
	e.GET("/error", func(c echo.Context) error {
		return errors.New("something went wrong")
	})
}
```

**With `HandleError: false` (default):** Errors are logged but not propagated to Echo's error handler.  
**With `HandleError: true`:** Errors are both logged and passed to Echo's error handler for proper HTTP response handling.

### Middleware Configuration Options

The `lecho.Config` struct provides extensive customization options:

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `Logger` | `*lecho.Logger` | Custom logger instance | `lecho.New(os.Stdout, WithTimestamp())` |
| `Skipper` | `middleware.Skipper` | Function to skip middleware | `middleware.DefaultSkipper` |
| `AfterNextSkipper` | `middleware.Skipper` | Skip logging after handler execution | `middleware.DefaultSkipper` |
| `BeforeNext` | `middleware.BeforeFunc` | Function executed before next handler | `nil` |
| `Enricher` | `lecho.Enricher` | Function to add custom fields | `nil` |
| `AfterNextEnricher` | `lecho.Enricher` | Function to add custom fields after the next handler runs; invoked only if `AfterNextSkipper` returns false | `nil` |
| `RequestIDHeader` | `string` | Header name for request ID | `"X-Request-ID"` |
| `RequestIDKey` | `string` | JSON key for request ID in logs | `"id"` |
| `NestKey` | `string` | Key for nesting request fields | `""` (no nesting) |
| `HandleError` | `bool` | Propagate errors to error handler | `false` |
| `RequestLatencyLimit` | `time.Duration` | Threshold for slow request detection | `0` (disabled) |
| `RequestLatencyLevel` | `zerolog.Level` | Log level for slow requests | `zerolog.InfoLevel` |

## Helpers

### Level Converters

Lecho provides utilities to convert between Echo and Zerolog log levels:

```go
package main

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
)

func main() {
	// Convert Echo log level to Zerolog level
	zeroLevel, echoLevel := lecho.MatchEchoLevel(log.WARN)
	fmt.Printf("Echo WARN -> Zerolog: %s, Echo: %s\n", zeroLevel, echoLevel)
	
	// Convert Zerolog level to Echo level
	echoLevel2, zeroLevel2 := lecho.MatchZeroLevel(zerolog.InfoLevel)
	fmt.Printf("Zerolog INFO -> Echo: %s, Zerolog: %s\n", echoLevel2, zeroLevel2)
}
```

### Context Logger Access

Access the logger from Echo context in your handlers:

```go
e.GET("/api/users", func(c echo.Context) error {
	// Method 1: Using Echo's logger interface
	c.Logger().Info("Fetching users")
	
	// Method 2: Using zerolog directly from request context
	zerolog.Ctx(c.Request().Context()).Info().Str("action", "fetch_users").Msg("Processing request")
	
	// Method 3: If using lecho.Context wrapper
	if lechoCtx, ok := c.(*lecho.Context); ok {
		lechoCtx.Logger().Info("Using lecho context")
	}
	
	return c.JSON(200, []string{"user1", "user2"})
})
```

## Advanced Configuration

### Complete Configuration Example

```go
package main

import (
	"net/http"
	"os"
	"time"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
)

func main() {
	e := echo.New()
	
	// Configure logger with all options
	logger := lecho.New(
		os.Stdout,
		lecho.WithLevel(log.INFO),
		lecho.WithTimestamp(),
		lecho.WithCaller(),
		lecho.WithFields(map[string]interface{}{
			"service": "api",
			"version": "1.0.0",
		}),
	)
	e.Logger = logger
	
	// Add middleware stack
	e.Use(middleware.RequestID())
	e.Use(lecho.Middleware(lecho.Config{
		Logger:              logger,
		HandleError:         true,
		RequestLatencyLevel: zerolog.WarnLevel,
		RequestLatencyLimit: 200 * time.Millisecond,
		NestKey:            "http",
		Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
			if userID := c.Get("user_id"); userID != nil {
				logger = logger.Str("user_id", userID.(string))
			}
			return logger
		},
		AfterNextEnricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
			// Example of adding a field after the next handler runs, based on context value
			logger = logger.Interface("some_key", c.Get("some_key"))
			return logger
		},
		Skipper: func(c echo.Context) bool {
			// Skip logging for health check endpoints
			return c.Request().URL.Path == "/health"
		},
	}))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("some_key", "some_value") // Example of setting context value for after next enricher
			return next(c)
		}
	})
	
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	
	e.GET("/api/slow", func(c echo.Context) error {
		time.Sleep(300 * time.Millisecond) // Simulates slow operation
		return c.JSON(200, map[string]string{"result": "completed"})
	})
	
	e.Start(":8080")
}
```
