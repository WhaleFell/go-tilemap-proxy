package handler

import (
	"go-map-proxy/internal/handler/common"
	"go-map-proxy/internal/handler/tilemap"

	"github.com/labstack/echo/v4"
)

func RegisterHandlers(echo *echo.Echo) {
	// Register all handlers here
	// e.g. echo.GET("/", common.Index)

	echo.GET("/", common.Index)
	echo.GET("/health/", common.HealthCheck)
	echo.GET("/systemInfo/", common.SystemInfo)

	echo.Any("/proxy/", common.URLProxy)

	// tile map server

	tilemapGroup := echo.Group("/map/")
	tilemapGroup.GET("list/", tilemap.TileMapSourceList)
	tilemapGroup.Any(":mapType/:x/:y/:z/", tilemap.TileMapHandler)
	tilemapGroup.GET("testpage/", tilemap.TileMapTestPageHandler)

	echo.HTTPErrorHandler = CustomGlobalHTTPErrorHandler
}
