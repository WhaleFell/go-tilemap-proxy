// calulate elapsed time for response

package middleware

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ElapsedTimeConfig struct {
	Skipper middleware.Skipper
}

// ref:
// 1. https://stackoverflow.com/questions/73678238/how-to-add-set-header-after-next-in-golang-echo-middleware
// 2. https://stackoverflow.com/a/73678476/22573614
// 3. https://github.com/gin-gonic/gin/issues/2406#issuecomment-1485704921
func ElapsedTimeMiddleware(configs ...ElapsedTimeConfig) echo.MiddlewareFunc {

	config := ElapsedTimeConfig{Skipper: func(c echo.Context) bool {
		return false
	}}

	if len(configs) > 0 {
		config = configs[0]
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			before := time.Now()
			// Before the response is sent, add a header with the elapsed time
			// ref: https://echo.labstack.com/docs/response#hooks
			c.Response().Before(func() {
				elapsed := time.Since(before).Milliseconds()
				c.Response().Header().Add("X-Elapsed-Time", fmt.Sprintf("%dms", elapsed))
			})

			err := next(c)

			return err
		}

	}
}
