# lecho :tomato:

Zerolog wrapper for [Echo](https://echo.labstack.com/) web framework.

## Quick start

```go
e := echo.New()
e.Logger = lecho.New(os.Stdout)
```

## Options

```go
e := echo.New()
e.Logger = lecho.New(
	os.Stdout,
	lecho.WithLevel(log.DEBUG),
	lecho.WithFields(map[string]interface{}{ "id": "foobar"}),
	lecho.WithTimestamp(),
	lecho.WithCaller(),
	lecho.WithPrefix("lecho lover"),
	lecho.WithHook(...),
	lecho.WithHookFunc(...),
)
```