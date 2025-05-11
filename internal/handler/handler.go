package handler

import (
	"go-map-proxy/internal/handler/common"

	"github.com/labstack/echo/v4"
)

func RegisterHandlers(echo *echo.Echo) {
	// Register all handlers here
	// e.g. echo.GET("/", common.Index)

	echo.GET("/", common.Index)
	echo.GET("/health", common.HealthCheck)
	echo.GET("/systemInfo", common.SystemInfo)

	echo.Any("/proxy", common.URLProxy)

	echo.HTTPErrorHandler = CustomGlobalHTTPErrorHandler
}
