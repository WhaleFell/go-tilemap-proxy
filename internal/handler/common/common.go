// common handler
package common

import "github.com/labstack/echo/v4"

func Index(c echo.Context) error {
	return c.String(200, "go-map-proxy server is running!")
}

func HealthCheck(c echo.Context) error {
	return c.String(200, "go-map-proxy server is healthy!")
}
