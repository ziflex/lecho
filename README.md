# :tomato: lecho :tomato:

Zerolog wrapper for [Echo](https://echo.labstack.com/) web framework.

## Quick start

```go
e := echo.New()
e.Logger = lecho.New(os.Stdout)
```

## Options

```go

import (
	"os",
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/ziflex/lecho"
)

type Context struct {
	echo.Context
	logger *lecho.Logger
}

func NewContext(c echo.Context, l *lecho.Logger) *Context {
	return &Context{c, l}
}

func (c *Context) Logger() echo.Logger {
	return c.logger
}

func main() {
    e := echo.New()
    e.Logger = lecho.New(
       os.Stdout,
       lecho.WithLevel(log.DEBUG),
       lecho.WithFields(map[string]interface{}{ "name": "lecho factory"}),
       lecho.WithTimestamp(),
       lecho.WithCaller(),
       lecho.WithPrefix("we ❤️ lecho"),
       lecho.WithHook(...),
       lecho.WithHookFunc(...),
    )
}
```

## Logger with Request ID

```go

import (
	"os",
	"github.com/labstack/echo"
	"github.com/ziflex/lecho"
)

type Context struct {
	echo.Context
	logger *lecho.Logger
}

func NewContext(c echo.Context, l *lecho.Logger) *Context {
	return &Context{c, l}
}

func (c *Context) Logger() echo.Logger {
	return c.logger
}

func main() {
    e := echo.New()
    logger := lecho.New(
            os.Stdout,
            lecho.WithLevel(log.DEBUG),
            lecho.WithTimestamp(),
            lecho.WithCaller(),
         )
    e.Logger = logger
    
    e.Use(middleware.RequestID())
    e.Use(func(c echo.Context) error {
        l := logger
        id := c.Response().Header().Get(echo.HeaderXRequestID)

        if id != "" {
            l = logger.Clone(lecho.WithField("id", id))
        }

        return next(NewContext(c, l))
    })	
}
```