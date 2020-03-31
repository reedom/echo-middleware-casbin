echo-middleware-casbin
======================

An [Echo][] middleware to authenticate a user through [casbin][].

[Echo]: https://echo.labstack.com/
[casbin]: https://casbin.org/

### Install

```sh
go get -u github.com/reedom/echo-middleware-casbin
```

### Usage

```go
package main

import (
	"net/http"

	"github.com/casbin/casbin"
	"github.com/labstack/echo"
	casbinmw "github.com/reedom/echo-middleware-casbin"
)

// casbinDataSource is a datasource for the middleware.
type casbinDataSource struct {
}

// GetSubject gets a subject from echo.Context.
// In this sample, it expects other middleware has set a user name at "user".
func (r *casbinDataSource) GetSubject(c echo.Context) string {
	return c.Get("user").(string)
}

// Introduce another middleware which extracts the accessor's user name and set
// it to "user" in echo.Context.
func setUserMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := // ... extract the user name somehow.
			c.Set("user", user)
			return next(c)
		}
	}
}

func main() {
	e := echo.New()

	ce, err := casbin.NewEnforcer("auth_model.conf", "auth_policy.conf")
	if err != nil {
		e.Logger.Fatal(err)
	}

	e.Use(setUserMiddleware())
	e.Use(casbinmw.Middleware(ce, &casbinDataSource{}))

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "OK")
	})

	e.GET("/login", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "OK")
	})

	e.Logger.Fatal(e.Start(":8080"))
}
```

### License

MIT
