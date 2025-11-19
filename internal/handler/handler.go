package handler

import (
	"go-map-proxy/internal/handler/common"
	"go-map-proxy/internal/handler/geeprotocol"
	"go-map-proxy/internal/handler/tilemap"
	"go-map-proxy/pkg/mapprovider"
	"go-map-proxy/pkg/request"

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
	tilemapGroup.Any(":mapType/:z/:x/:y/", tilemap.TileMapHandler)
	tilemapGroup.GET("testpage/", tilemap.TileMapTestPageHandler)

	// init GEE provider and handler
	// 初始化 GEE 提供者和处理器
	geeProvider := mapprovider.NewGoogleEarthEngineProvider(request.DefaultHTTPClient, "")
	geeHandler := geeprotocol.NewGEEHandler(geeProvider)

	// GEE protocol proxy server
	// GEE 协议代理服务器
	geeGroup := echo.Group("/gee")
	geeGroup.Any("/:path", geeHandler.GEEHandle)

	echo.Any("/*", geeHandler.GEEHandle)

	echo.HTTPErrorHandler = CustomGlobalHTTPErrorHandler
}
