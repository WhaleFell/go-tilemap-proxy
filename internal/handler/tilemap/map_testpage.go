package tilemap

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Embed the HTML test page using Go 1.16+ embed directive
// 使用 Go 1.16+ 的 embed 指令嵌入 HTML 测试页面
//go:embed testpage.html
var testPageHTML []byte

// TileMapTestPageHandler serves the embedded HTML test page for map tiles
// TileMapTestPageHandler 提供嵌入的地图瓦片HTML测试页面
func TileMapTestPageHandler(c echo.Context) error {
	// Set the appropriate content type for HTML
	// 为HTML设置适当的内容类型
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	
	// Set cache control headers to ensure fresh content during development
	// 设置缓存控制头以确保开发期间内容的新鲜性
	c.Response().Header().Set(echo.HeaderCacheControl, "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")
	
	// Return the embedded HTML content
	// 返回嵌入的HTML内容
	return c.Blob(http.StatusOK, echo.MIMETextHTMLCharsetUTF8, testPageHTML)
}