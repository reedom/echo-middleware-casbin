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
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper
		// BeforeFunc defines a function which is executed just before the middleware.
		BeforeFunc middleware.BeforeFunc
		// GetURLPathFunc defines a function which return the requested URL.
		GetURLPathFunc func(ctx echo.Context) string
		// SuccessHandler defines a function which is executed for a granted access.
		SuccessHandler func(echo.Context)
		// ErrorHandler defines a function which is executed for a rejected access.
		// It may be used to define a custom error.
		ErrorHandler func(error, echo.Context) error
		// Enforcer instance.
		Enforcer *casbin.Enforcer
		// DataSource is the interface that extract a subject from echo.Context.
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
	if config.GetURLPathFunc == nil {
		config.GetURLPathFunc = func(c echo.Context) string {
			return c.Request().URL.Path
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			if config.BeforeFunc != nil {
				config.BeforeFunc(c)
			}

			urlPath := config.GetURLPathFunc(c)
			ok, err := config.HasPermission(c, urlPath)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}

			if ok {
				if config.SuccessHandler != nil {
					config.SuccessHandler(c)
				}
				return next(c)
			}
			if config.ErrorHandler != nil {
				return config.ErrorHandler(echo.ErrForbidden, c)
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
func (a *Config) HasPermission(c echo.Context, urlPath string) (bool, error) {
	return a.Enforcer.Enforce(a.GetSubject(c), urlPath, c.Request().Method)
}
