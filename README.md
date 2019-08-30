# lecho :tomato:

Zerolog wrapper for [Echo](https://echo.labstack.com/) web framework.

## Installation

For Echo v4:

```
go get github.com/ziflex/lecho/v2
```

For Echo v3:

```
go get github.com/ziflex/lecho
```

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
	"github.com/ziflex/lecho/v2"
)

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

## Middleware

### Logging requests and attaching request id to a context logger 

```go

import (
	"os",
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/ziflex/lecho/v2"
)

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
    e.Use(lecho.Middleware(lecho.Config{
    	Logger: logger
    }))	
}
```
