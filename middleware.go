package casbinmw

import (
	"net/http"

	"github.com/casbin/casbin/v2"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	// Config defines the config for this middleware.
	Config struct {
		Skipper    middleware.Skipper
		Enforcer   *casbin.Enforcer
		DataSource DataSource
	}
)

var (
	DefaultConfig = Config{
		Skipper: middleware.DefaultSkipper,
	}
)

// DataSource is the interface that extract a subject from echo.Context.
type DataSource interface {
	GetSubject(c echo.Context) string
}

// Middleware returns a Echo middleware.
func Middleware(ce *casbin.Enforcer, ds DataSource) echo.MiddlewareFunc {
	c := DefaultConfig
	c.Enforcer = ce
	c.DataSource = ds
	return MiddlewareWithConfig(c)
}

// MiddlewareWithConfig returns an Echo middleware with config.
func MiddlewareWithConfig(config Config) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultConfig.Skipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			if ok, err := config.HasPermission(c); err == nil && ok {
				return next(c)
			} else if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			return echo.ErrForbidden
		}
	}
}

// GetSubject extract a subject from the request.
func (a *Config) GetSubject(c echo.Context) string {
	return a.DataSource.GetSubject(c)
}

// HasPermission checks a resource access permission against casbin with the subject/method/path combination from the request.
// Returns true (permission granted) or false (permission forbidden).
func (a *Config) HasPermission(c echo.Context) (bool, error) {
	return a.Enforcer.Enforce(a.GetSubject(c), c.Request().URL.Path, c.Request().Method)
}
