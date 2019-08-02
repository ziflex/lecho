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
	"github.com/ziflex/lecho"
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
    
    e.Use(lecho.Middleware(lecho.Config{
    	Logger: logger
    }))	
}
```
