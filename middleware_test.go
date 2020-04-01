package casbinmw_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/labstack/echo/v4"
	"github.com/reedom/echo-middleware-casbin"
)

func setUserMiddleware(user string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user", user)
			return next(c)
		}
	}
}

type testDataSource struct {
}

func (r *testDataSource) GetSubject(c echo.Context) string {
	return c.Get("user").(string)
}

func TestPermissions(t *testing.T) {
	// Given
	ce, err := casbin.NewEnforcer("auth_model.conf", "auth_policy.conf")
	if err != nil {
		t.Fatal(err)
		return
	}

	var tests = []struct {
		user   string
		path   string
		method string
		status int
	}{
		{"admin", "/", http.MethodGet, http.StatusOK},
		{"admin", "/login", http.MethodPost, http.StatusOK},
		{"anonymous", "/", http.MethodGet, http.StatusForbidden},
		{"anonymous", "/login", http.MethodGet, http.StatusOK},
		{"anonymous", "/login", http.MethodPut, http.StatusForbidden},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%v-%v-%v_%v", tt.user, tt.path, tt.method, tt.status)
		t.Run(name, func(t *testing.T) {
			e := echo.New()
			e.Use(setUserMiddleware(tt.user))
			e.Use(casbinmw.Middleware(ce, &testDataSource{}))
			e.Any(tt.path, func(c echo.Context) error {
				return c.JSON(http.StatusOK, "OK")
			})
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)
			if tt.status != rec.Code {
				t.Errorf("got %v, want %v", rec.Code, tt.status)
			}
		})
	}
}
