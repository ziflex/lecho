# :tomato: lecho :tomato:

Zerolog wrapper for [Echo](https://echo.labstack.com/) web framework.

## Quick start

```go
e := echo.New()
e.Logger = lecho.New(os.Stdout)
```

## Options

```go
e := echo.New()
l := lecho.New(
     	os.Stdout,
     	lecho.WithLevel(log.DEBUG),
     	lecho.WithFields(map[string]interface{}{ "id": "foobar"}),
     	lecho.WithTimestamp(),
     	lecho.WithCaller(),
     	lecho.WithPrefix("lecho lover"),
     	lecho.WithHook(...),
     	lecho.WithHookFunc(...),
     )
e.Logger = l

e.Use(func(c echo.Context) error {
    l := logger
    id := c.Request().Header.Get(echo.HeaderXRequestID)

    if id != "" {
        l = logger.Clone(lecho.WithField("id", id))
    }

    return next(NewContext(c, l))
})
```