package middleware

import (
	"go-map-proxy/pkg/logger"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/labstack/echo/v4/middleware"
)

func RegisterMiddleware(e *echo.Echo) {
	// Register all middlewares here
	// e.g.:
	// http.Handle("/path", middleware1(middleware2(http.HandlerFunc(yourHandler))))

	// buildin middlewares
	// e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Logger.Info("Request", zap.String("URI", v.URI), zap.Int("status", v.Status))
			return nil
		},
	}))

	// Custom middlewares
	e.Use(CORSMiddleware())
	e.Use(ElapsedTimeMiddleware())
}
