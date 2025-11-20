// common handler
// 公共处理器
package common

import (
	"net/http"

	_ "embed"

	"github.com/labstack/echo/v4"
)

//go:embed index.html
var indexHTML string

// Index serves the embedded homepage to echo context
// Index 函数向 Echo 上下文返回嵌入的首页内容
func Index(c echo.Context) error {
	return c.HTML(http.StatusOK, indexHTML)
}

func HealthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "go-map-proxy server is healthy!")
}
