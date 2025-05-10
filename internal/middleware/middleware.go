package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/labstack/echo/v4/middleware"
)

func RegisterMiddleware(e *echo.Echo) {
	// Register all middlewares here
	// e.g.:
	// http.Handle("/path", middleware1(middleware2(http.HandlerFunc(yourHandler))))

	// buildin middlewares
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Custom middlewares
	e.Use(CORSMiddleware())
	e.Use(ElapsedTimeMiddleware())
}
