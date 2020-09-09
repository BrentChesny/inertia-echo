package inertia

import (
	"elipzis.com/inertia-echo/util"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
)

//
type MiddlewareConfig struct {
	Inertia *Inertia
	Skipper middleware.Skipper
}

// Create a default Inertia Middleware for the given echo reference
// if the Inertia instance itself is not required in other instances
func Middleware(echo *echo.Echo) echo.MiddlewareFunc {
	return MiddlewareWithConfig(MiddlewareConfig{
		Inertia: NewInertia(echo),
	})
}

// The Inertia Middleware to check every request for what it needs
func MiddlewareWithConfig(config MiddlewareConfig) echo.MiddlewareFunc {
	if config.Inertia == nil {
		log.Fatal("[Inertia] Please provide an Inertia reference with your config!")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip, if configured and true
			if config.Skipper != nil && config.Skipper(c) {
				return next(c)
			}

			// Run Inertia post
			if err := next(c); err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			if req.Header.Get(HeaderPrefix) == "" {
				return nil
			}

			if req.Method == "GET" && req.Header.Get(HeaderVersion) != config.Inertia.GetVersion() {
				// Reflash?
				if s, err := session.Get("session", c); err == nil {
					flashes := s.Flashes()
					_ = s.Save(c.Request(), c.Response())
					// config.Inertia.Share("flash", flashes)
					for _, flash := range flashes {
						s.AddFlash(flash)
					}
				}

				res.Header().Set(HeaderLocation, req.URL.String())
				return c.String(http.StatusConflict, "")
			}

			if exists, _ := util.InArray(req.Method, []string{"PUT", "PATCH", "DELETE"}); exists && res.Status == 302 {
				res.Status = http.StatusSeeOther
			}

			return nil
		}
	}
}
