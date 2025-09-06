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
- [Use Cases](#use-cases)
- [Integration Examples](#integration-examples)
- [Testing](#testing)
- [Performance](#performance)
- [Best Practices](#best-practices)
- [Migration Guide](#migration-guide)
- [Troubleshooting](#troubleshooting)
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
```

## Use Cases

### Web API with Request Tracing

Perfect for REST APIs that need structured logging with request correlation:

```go
package main

import (
	"net/http"
	"os"
	"time"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
)

func main() {
	e := echo.New()
	
	// Create logger for API service
	logger := lecho.New(os.Stdout, 
		lecho.WithTimestamp(),
		lecho.WithFields(map[string]interface{}{
			"service": "user-api",
			"version": "1.2.3",
		}),
	)
	e.Logger = logger
	
	// Essential middleware stack
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
		Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
			// Add API version and client info
			if version := c.Request().Header.Get("API-Version"); version != "" {
				logger = logger.Str("api_version", version)
			}
			if clientID := c.Request().Header.Get("X-Client-ID"); clientID != "" {
				logger = logger.Str("client_id", clientID)
			}
			return logger
		},
	}))
	
	// API routes
	api := e.Group("/api/v1")
	api.GET("/users/:id", getUserHandler)
	api.POST("/users", createUserHandler)
	
	e.Start(":8080")
}

func getUserHandler(c echo.Context) error {
	userID := c.Param("id")
	c.Logger().Infof("Fetching user %s", userID)
	
	// Simulate user lookup
	user := map[string]interface{}{
		"id": userID,
		"name": "John Doe",
		"email": "john@example.com",
	}
	
	return c.JSON(http.StatusOK, user)
}

func createUserHandler(c echo.Context) error {
	c.Logger().Info("Creating new user")
	
	// Simulate user creation
	time.Sleep(50 * time.Millisecond)
	
	user := map[string]interface{}{
		"id": "user123",
		"status": "created",
	}
	
	return c.JSON(http.StatusCreated, user)
}
```

### Microservice with Health Checks

Ideal for microservices that need health monitoring and observability:

```go
package main

import (
	"context"
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
	
	// Production-ready logger configuration
	logger := lecho.New(os.Stdout,
		lecho.WithLevel(log.INFO),
		lecho.WithTimestamp(),
		lecho.WithFields(map[string]interface{}{
			"service":     "payment-service",
			"environment": os.Getenv("ENVIRONMENT"),
			"region":      os.Getenv("AWS_REGION"),
		}),
	)
	e.Logger = logger
	
	// Middleware with health check exclusion
	e.Use(lecho.Middleware(lecho.Config{
		Logger:              logger,
		RequestLatencyLevel: zerolog.WarnLevel,
		RequestLatencyLimit: 200 * time.Millisecond,
		Skipper: func(c echo.Context) bool {
			// Skip logging for health checks to reduce noise
			return c.Path() == "/health" || c.Path() == "/metrics"
		},
		Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
			// Add correlation ID and service mesh headers
			if correlationID := c.Request().Header.Get("X-Correlation-ID"); correlationID != "" {
				logger = logger.Str("correlation_id", correlationID)
			}
			if spanID := c.Request().Header.Get("X-Span-ID"); spanID != "" {
				logger = logger.Str("span_id", spanID)
			}
			return logger
		},
	}))
	
	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
			"service": "payment-service",
			"timestamp": time.Now().ISO8601(),
		})
	})
	
	// Business logic endpoints
	e.POST("/payments", processPaymentHandler)
	e.GET("/payments/:id", getPaymentHandler)
	
	e.Start(":8080")
}

func processPaymentHandler(c echo.Context) error {
	c.Logger().Info("Processing payment request")
	
	// Simulate payment processing
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()
	
	// Add structured logging for business events
	zerolog.Ctx(ctx).Info().
		Str("action", "payment_processing").
		Str("payment_method", "credit_card").
		Msg("Payment initiated")
	
	time.Sleep(100 * time.Millisecond) // Simulate processing
	
	zerolog.Ctx(ctx).Info().
		Str("action", "payment_completed").
		Str("transaction_id", "txn_123456").
		Msg("Payment processed successfully")
	
	return c.JSON(http.StatusOK, map[string]interface{}{
		"transaction_id": "txn_123456",
		"status": "completed",
	})
}

func getPaymentHandler(c echo.Context) error {
	paymentID := c.Param("id")
	c.Logger().Infof("Retrieving payment %s", paymentID)
	
	payment := map[string]interface{}{
		"id": paymentID,
		"status": "completed",
		"amount": 99.99,
	}
	
	return c.JSON(http.StatusOK, payment)
}
```

### Development Environment with Debug Logging

Configuration optimized for local development with detailed debugging:

```go
package main

import (
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
	
	// Development logger with debug level and caller info
	logger := lecho.New(os.Stdout,
		lecho.WithLevel(log.DEBUG),
		lecho.WithTimestamp(),
		lecho.WithCaller(), // Show file:line for debugging
		lecho.WithFields(map[string]interface{}{
			"mode": "development",
		}),
	)
	e.Logger = logger
	
	// Development middleware stack
	e.Use(middleware.Logger()) // Echo's default logger for additional info
	e.Use(middleware.RequestID())
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
		// Log all requests in development, including static assets
		Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
			// Add request details useful for debugging
			logger = logger.Str("content_type", c.Request().Header.Get("Content-Type"))
			logger = logger.Str("accept", c.Request().Header.Get("Accept"))
			if c.Request().Header.Get("Authorization") != "" {
				logger = logger.Bool("authenticated", true)
			}
			return logger
		},
	}))
	
	// Debug routes
	e.GET("/debug/headers", func(c echo.Context) error {
		c.Logger().Debug("Dumping request headers")
		
		headers := make(map[string]string)
		for key, values := range c.Request().Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		
		return c.JSON(200, map[string]interface{}{
			"headers": headers,
			"method": c.Request().Method,
			"path": c.Request().URL.Path,
		})
	})
	
	e.GET("/debug/slow", func(c echo.Context) error {
		c.Logger().Debug("Simulating slow endpoint")
		time.Sleep(2 * time.Second)
		return c.String(200, "Slow response completed")
	})
	
	e.Start(":3000")
}
```

## Integration Examples

### Integration with JWT Authentication

Combining lecho with JWT middleware for authenticated API logging:

```go
package main

import (
	"net/http"
	"os"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func main() {
	e := echo.New()
	
	logger := lecho.New(os.Stdout,
		lecho.WithTimestamp(),
		lecho.WithFields(map[string]interface{}{
			"service": "auth-api",
		}),
	)
	e.Logger = logger
	
	// Public routes
	e.POST("/login", loginHandler)
	
	// JWT middleware configuration
	jwtConfig := echojwt.Config{
		SigningKey: []byte("your-secret-key"),
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(JWTClaims)
		},
	}
	
	// Protected routes group
	protected := e.Group("/api")
	protected.Use(echojwt.WithConfig(jwtConfig))
	protected.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
		Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
			// Extract user info from JWT token
			if token, ok := c.Get("user").(*jwt.Token); ok {
				if claims, ok := token.Claims.(*JWTClaims); ok {
					logger = logger.Str("user_id", claims.UserID)
					logger = logger.Str("user_role", claims.Role)
				}
			}
			return logger
		},
	}))
	
	protected.GET("/profile", profileHandler)
	protected.GET("/admin/users", adminUsersHandler)
	
	e.Start(":8080")
}

func loginHandler(c echo.Context) error {
	c.Logger().Info("User login attempt")
	// Login logic here
	return c.JSON(http.StatusOK, map[string]string{"token": "jwt-token-here"})
}

func profileHandler(c echo.Context) error {
	c.Logger().Info("Fetching user profile")
	return c.JSON(http.StatusOK, map[string]string{"profile": "user data"})
}

func adminUsersHandler(c echo.Context) error {
	c.Logger().Info("Admin accessing user list")
	return c.JSON(http.StatusOK, []string{"user1", "user2"})
}
```

### Integration with Database and CORS

Complete setup with database logging and CORS for web applications:

```go
package main

import (
	"database/sql"
	"net/http"
	"os"
	"time"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
	_ "github.com/lib/pq" // PostgreSQL driver
)

type App struct {
	DB     *sql.DB
	Logger *lecho.Logger
}

func main() {
	// Database connection
	db, err := sql.Open("postgres", "postgres://user:pass@localhost/db?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	
	// Logger setup
	logger := lecho.New(os.Stdout,
		lecho.WithTimestamp(),
		lecho.WithFields(map[string]interface{}{
			"service": "web-api",
		}),
	)
	
	app := &App{
		DB:     db,
		Logger: logger,
	}
	
	e := echo.New()
	e.Logger = logger
	
	// CORS configuration
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "https://myapp.com"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))
	
	// Request logging with database query tracking
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
		Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
			// Add request context
			if origin := c.Request().Header.Get("Origin"); origin != "" {
				logger = logger.Str("origin", origin)
			}
			return logger
		},
	}))
	
	// Routes
	e.GET("/api/users", app.getUsersHandler)
	e.POST("/api/users", app.createUserHandler)
	
	e.Start(":8080")
}

func (app *App) getUsersHandler(c echo.Context) error {
	start := time.Now()
	
	// Log database query
	zerolog.Ctx(c.Request().Context()).Debug().
		Str("query", "SELECT * FROM users").
		Msg("Executing database query")
	
	rows, err := app.DB.Query("SELECT id, name, email FROM users")
	if err != nil {
		zerolog.Ctx(c.Request().Context()).Error().
			Err(err).
			Msg("Database query failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer rows.Close()
	
	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var name, email string
		rows.Scan(&id, &name, &email)
		users = append(users, map[string]interface{}{
			"id": id, "name": name, "email": email,
		})
	}
	
	// Log successful query
	zerolog.Ctx(c.Request().Context()).Info().
		Int("user_count", len(users)).
		Dur("query_duration", time.Since(start)).
		Msg("Users fetched successfully")
	
	return c.JSON(http.StatusOK, users)
}

func (app *App) createUserHandler(c echo.Context) error {
	c.Logger().Info("Creating new user")
	
	// User creation logic with database insert
	// ...
	
	return c.JSON(http.StatusCreated, map[string]string{"status": "created"})
}
```

## Testing

### Testing Handlers with Lecho

How to test your Echo handlers that use lecho middleware:

```go
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/ziflex/lecho/v3"
)

func TestUserHandler(t *testing.T) {
	// Create buffer to capture logs
	logBuffer := &bytes.Buffer{}
	
	// Setup Echo with lecho
	e := echo.New()
	logger := lecho.New(logBuffer, lecho.WithTimestamp())
	e.Logger = logger
	
	// Add middleware
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
	}))
	
	// Setup route
	e.GET("/users/:id", getUserHandler)
	
	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	rec := httptest.NewRecorder()
	
	// Execute request
	e.ServeHTTP(rec, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)
	
	// Parse and validate logs
	logEntries := strings.Split(strings.TrimSpace(logBuffer.String()), "\n")
	assert.Greater(t, len(logEntries), 0)
	
	// Parse the last log entry (middleware log)
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(logEntries[len(logEntries)-1]), &logEntry)
	assert.NoError(t, err)
	
	// Validate log structure
	assert.Equal(t, "info", logEntry["level"])
	assert.Equal(t, "GET", logEntry["method"])
	assert.Equal(t, "/users/123", logEntry["uri"])
	assert.Equal(t, float64(200), logEntry["status"])
	assert.Contains(t, logEntry, "latency")
}

func TestErrorHandling(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	
	e := echo.New()
	logger := lecho.New(logBuffer, lecho.WithTimestamp())
	e.Logger = logger
	
	// Configure error handling
	e.Use(lecho.Middleware(lecho.Config{
		Logger:      logger,
		HandleError: true,
	}))
	
	// Handler that returns an error
	e.GET("/error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	})
	
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	
	e.ServeHTTP(rec, req)
	
	// Verify error was logged
	logs := logBuffer.String()
	assert.Contains(t, logs, "error")
	assert.Contains(t, logs, "Bad request")
}

func TestLogEnrichment(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	
	e := echo.New()
	logger := lecho.New(logBuffer, lecho.WithTimestamp())
	e.Logger = logger
	
	// Middleware with enricher
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
		Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
			return logger.Str("test_field", "test_value")
		},
	}))
	
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	
	e.ServeHTTP(rec, req)
	
	// Verify enrichment
	logs := logBuffer.String()
	assert.Contains(t, logs, "test_field")
	assert.Contains(t, logs, "test_value")
}

// Helper function for your actual handlers
func getUserHandler(c echo.Context) error {
	userID := c.Param("id")
	c.Logger().Infof("Fetching user %s", userID)
	
	return c.JSON(http.StatusOK, map[string]string{
		"id": userID,
		"name": "Test User",
	})
}
```

### Benchmark Testing

Performance testing to ensure lecho doesn't impact application performance:

```go
package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/labstack/echo/v4"
	"github.com/ziflex/lecho/v3"
)

func BenchmarkWithoutMiddleware(b *testing.B) {
	e := echo.New()
	e.GET("/benchmark", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	
	req := httptest.NewRequest(http.MethodGet, "/benchmark", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
	}
}

func BenchmarkWithLechoMiddleware(b *testing.B) {
	logBuffer := &bytes.Buffer{}
	
	e := echo.New()
	logger := lecho.New(logBuffer)
	e.Logger = logger
	
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
	}))
	
	e.GET("/benchmark", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	
	req := httptest.NewRequest(http.MethodGet, "/benchmark", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		logBuffer.Reset() // Clear buffer to prevent memory growth
	}
}

func BenchmarkWithEnricher(b *testing.B) {
	logBuffer := &bytes.Buffer{}
	
	e := echo.New()
	logger := lecho.New(logBuffer)
	e.Logger = logger
	
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
		Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
			return logger.Str("request_id", "12345")
		},
	}))
	
	e.GET("/benchmark", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	
	req := httptest.NewRequest(http.MethodGet, "/benchmark", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		logBuffer.Reset()
	}
}
```

## Performance

### Performance Characteristics

Lecho is built on zerolog, one of the fastest JSON loggers for Go. Here are some performance considerations:

```go
package main

import (
	"fmt"
	"os"
	"time"
	"github.com/labstack/echo/v4"
	"github.com/ziflex/lecho/v3"
)

func main() {
	// High-performance configuration for production
	logger := lecho.New(os.Stdout,
		lecho.WithTimestamp(),
		// Avoid expensive options in production:
		// - WithCaller() adds file:line lookup overhead
		// - Complex enrichers can slow down requests
	)
	
	e := echo.New()
	e.Logger = logger
	
	// Minimal middleware configuration for best performance
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
		// Skip non-essential requests
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/health" || c.Path() == "/metrics"
		},
		// Minimal enricher
		Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
			// Only add essential fields to minimize allocations
			if requestID := c.Response().Header().Get("X-Request-ID"); requestID != "" {
				return logger.Str("request_id", requestID)
			}
			return logger
		},
	}))
	
	e.GET("/performance", performanceTestHandler)
	e.Start(":8080")
}

func performanceTestHandler(c echo.Context) error {
	start := time.Now()
	
	// Simulate work
	time.Sleep(10 * time.Millisecond)
	
	// Structured logging with minimal overhead
	c.Logger().Infof("Request processed in %v", time.Since(start))
	
	return c.JSON(200, map[string]interface{}{
		"processed_in": time.Since(start).String(),
		"timestamp": time.Now().Unix(),
	})
}
```

### Memory Management

Best practices for memory-efficient logging:

```go
package main

import (
	"io"
	"os"
	"sync"
	"time"
	"bytes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
)

// Pool for reusing log buffers
var loggerPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

func createProductionLogger() *lecho.Logger {
	// Use a writer that supports efficient buffering
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    true, // Disable colors in production for performance
	}
	
	return lecho.New(output,
		lecho.WithTimestamp(),
		lecho.WithLevel(log.INFO), // Avoid DEBUG in production
	)
}

func efficientEnricher(c echo.Context, logger zerolog.Context) zerolog.Context {
	// Pre-allocate map for multiple fields
	fields := make(map[string]interface{}, 3)
	
	if userID := c.Get("user_id"); userID != nil {
		fields["user_id"] = userID
	}
	
	if traceID := c.Request().Header.Get("X-Trace-ID"); traceID != "" {
		fields["trace_id"] = traceID
	}
	
	if sessionID := c.Request().Header.Get("X-Session-ID"); sessionID != "" {
		fields["session_id"] = sessionID
	}
	
	return logger.Fields(fields)
}
```

## Best Practices

### Production Configuration

Recommended setup for production environments:

```go
package main

import (
	"os"
	"strings"
	"time"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
)

func getLoggerOptions() []lecho.Setter {
	options := []lecho.Setter{
		lecho.WithLevel(getLogLevel()),
		lecho.WithTimestamp(),
		lecho.WithFields(map[string]interface{}{
			"service":     getServiceName(),
			"version":     getVersion(),
			"environment": getEnvironment(),
			"region":      getRegion(),
		}),
	}
	
	// Only add caller info in development
	if getEnvironment() == "development" {
		options = append(options, lecho.WithCaller())
	}
	
	return options
}

func main() {
	e := echo.New()
	
	// Production logger configuration
	logger := lecho.New(os.Stdout, getLoggerOptions()...)
	e.Logger = logger
	
	// Production middleware stack
	e.Use(middleware.Secure())
	e.Use(middleware.RequestID())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(100)))
	
	// Optimized logging middleware
	e.Use(lecho.Middleware(lecho.Config{
		Logger:              logger,
		HandleError:         true,
		RequestLatencyLevel: zerolog.WarnLevel,
		RequestLatencyLimit: 500 * time.Millisecond,
		Skipper:            productionSkipper,
		Enricher:           productionEnricher,
	}))
	
	e.Start(":8080")
}

func getLogLevel() log.Lvl {
	level := os.Getenv("LOG_LEVEL")
	switch strings.ToUpper(level) {
	case "DEBUG":
		return log.DEBUG
	case "WARN":
		return log.WARN
	case "ERROR":
		return log.ERROR
	default:
		return log.INFO
	}
}


func productionSkipper(c echo.Context) bool {
	path := c.Path()
	// Skip health checks, metrics, and static assets
	return path == "/health" || 
		   path == "/metrics" || 
		   path == "/ready" ||
		   strings.HasPrefix(path, "/static/")
}

func productionEnricher(c echo.Context, logger zerolog.Context) zerolog.Context {
	// Only add essential fields
	if requestID := c.Response().Header().Get("X-Request-ID"); requestID != "" {
		logger = logger.Str("request_id", requestID)
	}
	
	// Add user context if available
	if userID := c.Get("user_id"); userID != nil {
		logger = logger.Str("user_id", userID.(string))
	}
	
	// Add tracing information
	if traceID := c.Request().Header.Get("X-Trace-ID"); traceID != "" {
		logger = logger.Str("trace_id", traceID)
	}
	
	return logger
}

func getServiceName() string {
	if name := os.Getenv("SERVICE_NAME"); name != "" {
		return name
	}
	return "unknown-service"
}

func getVersion() string {
	if version := os.Getenv("SERVICE_VERSION"); version != "" {
		return version
	}
	return "unknown"
}

func getEnvironment() string {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return env
	}
	return "development"
}

func getRegion() string {
	return os.Getenv("AWS_REGION")
}
```

### Security Considerations

Protecting sensitive information in logs:

```go
package main

import (
	"os"
	"regexp"
	"strings"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
)

var (
	// Patterns for sensitive data
	emailPattern    = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	phonePattern    = regexp.MustCompile(`\b\d{3}-\d{3}-\d{4}\b`)
	creditCardPattern = regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`)
)

func main() {
	e := echo.New()
	
	// Logger with sanitization
	logger := lecho.New(os.Stdout,
		lecho.WithTimestamp(),
		lecho.WithHookFunc(sanitizationHook),
	)
	e.Logger = logger
	
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
		Enricher: secureEnricher,
	}))
	
	e.POST("/users", createUserHandler)
	e.Start(":8080")
}

func sanitizationHook(e *zerolog.Event, level zerolog.Level, msg string) {
	// Sanitize message content
	sanitized := sanitizeString(msg)
	if sanitized != msg {
		e.Str("original_message_sanitized", "true")
	}
}

func sanitizeString(s string) string {
	// Replace sensitive patterns with placeholders
	s = emailPattern.ReplaceAllString(s, "[EMAIL]")
	s = phonePattern.ReplaceAllString(s, "[PHONE]")
	s = creditCardPattern.ReplaceAllString(s, "[CREDIT_CARD]")
	return s
}

func secureEnricher(c echo.Context, logger zerolog.Context) zerolog.Context {
	// Never log authorization headers
	if auth := c.Request().Header.Get("Authorization"); auth != "" {
		logger = logger.Bool("has_auth", true)
		// Log only the auth type, not the token
		if strings.HasPrefix(auth, "Bearer ") {
			logger = logger.Str("auth_type", "bearer")
		} else if strings.HasPrefix(auth, "Basic ") {
			logger = logger.Str("auth_type", "basic")
		}
	}
	
	// Sanitize user agent
	userAgent := c.Request().UserAgent()
	if len(userAgent) > 100 {
		userAgent = userAgent[:100] + "..."
	}
	logger = logger.Str("user_agent_truncated", userAgent)
	
	return logger
}

func createUserHandler(c echo.Context) error {
	// Never log raw request body for sensitive endpoints
	c.Logger().Info("User creation request received")
	
	// Process user creation...
	c.Logger().Info("User created successfully")
	
	return c.JSON(200, map[string]string{"status": "created"})
}
```

## Migration Guide

### From Standard Echo Logger

Migrating from Echo's default logger to lecho:

```go
// Before: Using Echo's default logger
package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	
	// Old way with Echo's default logger
	e.Use(middleware.Logger())
	
	e.GET("/", func(c echo.Context) error {
		c.Logger().Info("Processing request")
		return c.String(200, "Hello")
	})
	
	e.Start(":8080")
}

// After: Using lecho
package main

import (
	"os"
	"github.com/labstack/echo/v4"
	"github.com/ziflex/lecho/v3"
)

func main() {
	e := echo.New()
	
	// New way with lecho
	logger := lecho.New(os.Stdout, lecho.WithTimestamp())
	e.Logger = logger
	
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
	}))
	
	e.GET("/", func(c echo.Context) error {
		// Same interface, structured output
		c.Logger().Info("Processing request")
		return c.String(200, "Hello")
	})
	
	e.Start(":8080")
}
```

### From Logrus

Migrating from Logrus to lecho:

```go
// Before: Using Logrus
package main

import (
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func main() {
	// Logrus configuration
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	
	e := echo.New()
	e.Logger = /* custom logrus wrapper */
	
	e.GET("/", func(c echo.Context) error {
		logrus.WithFields(logrus.Fields{
			"method": c.Request().Method,
			"path":   c.Request().URL.Path,
		}).Info("Request received")
		
		return c.String(200, "Hello")
	})
	
	e.Start(":8080")
}

// After: Using lecho
package main

import (
	"os"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
)

func main() {
	// Lecho configuration (similar to logrus)
	logger := lecho.New(os.Stdout,
		lecho.WithTimestamp(),
		lecho.WithLevel(log.INFO),
	)
	
	e := echo.New()
	e.Logger = logger
	
	// Automatic request logging with middleware
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
	}))
	
	e.GET("/", func(c echo.Context) error {
		// Structured logging with zerolog
		zerolog.Ctx(c.Request().Context()).Info().
			Str("method", c.Request().Method).
			Str("path", c.Request().URL.Path).
			Msg("Request received")
		
		return c.String(200, "Hello")
	})
	
	e.Start(":8080")
}
```

## Troubleshooting

### Common Issues and Solutions

#### Issue: No logs appearing

```go
// Problem: Logger not properly configured
e := echo.New()
// Missing: e.Logger = lecho.New(...)

// Solution: Always set the logger
logger := lecho.New(os.Stdout, lecho.WithTimestamp())
e.Logger = logger
```

#### Issue: Missing request logs

```go
// Problem: Middleware not added
e.Use(someOtherMiddleware())
// Missing: e.Use(lecho.Middleware(...))

// Solution: Add lecho middleware
e.Use(lecho.Middleware(lecho.Config{
	Logger: logger,
}))
```

#### Issue: Performance impact

```go
// Problem: Expensive configuration
logger := lecho.New(os.Stdout,
	lecho.WithCaller(),     // Expensive: file/line lookup
	lecho.WithLevel(log.DEBUG), // Verbose: too many logs
)

// Solution: Optimize for production
logger := lecho.New(os.Stdout,
	lecho.WithTimestamp(),
	lecho.WithLevel(log.INFO), // Appropriate level
	// Remove WithCaller() in production
)
```

#### Issue: Sensitive data in logs

```go
// Problem: Logging sensitive information
c.Logger().Infof("Processing payment for user %s with card %s", userID, creditCard)

// Solution: Sanitize or avoid logging sensitive data
c.Logger().Infof("Processing payment for user %s", userID)
// Log transaction ID instead of card details
zerolog.Ctx(c.Request().Context()).Info().
	Str("user_id", userID).
	Str("transaction_id", transactionID).
	Msg("Payment processed")
```

#### Issue: Memory leaks with large logs

```go
// Problem: Unbounded log growth
func handler(c echo.Context) error {
	for i := 0; i < 1000; i++ {
		c.Logger().Debugf("Processing item %d", i) // Too many logs
	}
	return c.String(200, "OK")
}

// Solution: Batch logging or use appropriate levels
func handler(c echo.Context) error {
	items := make([]int, 1000)
	// Process items...
	
	c.Logger().Infof("Processed %d items", len(items)) // Single summary log
	return c.String(200, "OK")
}
```

### Debug Mode

Enable debug logging for troubleshooting:

```go
package main

import (
	"os"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/ziflex/lecho/v3"
)

func main() {
	// Debug configuration
	debug := os.Getenv("DEBUG") == "true"
	
	var logger *lecho.Logger
	if debug {
		logger = lecho.New(os.Stdout,
			lecho.WithLevel(log.DEBUG),
			lecho.WithTimestamp(),
			lecho.WithCaller(), // Show file:line in debug mode
		)
	} else {
		logger = lecho.New(os.Stdout,
			lecho.WithLevel(log.INFO),
			lecho.WithTimestamp(),
		)
	}
	
	e := echo.New()
	e.Debug = debug
	e.Logger = logger
	
	if debug {
		logger.Debug("Debug mode enabled")
	}
	
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
	}))
	
	e.Start(":8080")
}

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
		Skipper: func(c echo.Context) bool {
			// Skip logging for health check endpoints
			return c.Request().URL.Path == "/health"
		},
	}))
	
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
