package geeprotocol

import (
	"fmt"
	"go-map-proxy/internal/model"
	"go-map-proxy/pkg/logger"
	"go-map-proxy/pkg/mapprovider"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type GEEHandler struct {
	GEEClient *mapprovider.GoogleEarthEngineProvider
}

// NewGEEHandler creates a new GEE handler instance
// 创建一个新的 GEE 处理器实例
func NewGEEHandler(geeClient *mapprovider.GoogleEarthEngineProvider) *GEEHandler {
	return &GEEHandler{
		GEEClient: geeClient,
	}
}

// GEEHandle handles requests for Google Earth Engine protocol
// GEE 协议请求处理器
// Route: `/gee/*` - forwards all requests to GEE server with SessionID cookie
// 路由: `/gee/*` - 将所有请求转发到 GEE 服务器并添加 SessionID cookie
func (h *GEEHandler) GEEHandle(c echo.Context) error {
	// Get the full path after /gee/
	// 获取 /gee/ 后的完整路径
	path := c.Request().URL.Path
	if len(path) >= 4 && path[:4] == "/gee" {
		path = path[4:] // Remove "/gee" prefix
	}

	logger.Infof("GEE proxy request: %s %s", c.Request().Method, c.Request().URL.Path)

	if path == "" {
		return c.JSON(http.StatusBadRequest, model.BaseAPIResponse[any]{
			Code:    http.StatusBadRequest,
			Message: "Path parameter is required",
			Data:    nil,
		})
	}

	// Preserve query parameters
	// 保留查询参数
	if c.Request().URL.RawQuery != "" {
		path = path + "?" + c.Request().URL.RawQuery
	}

	logger.Infof("GEE proxy request: %s", path)

	// Special handling for /geauth?ct=pro requests
	// 特殊处理 /geauth?ct=pro 请求
	if strings.Contains(path, "geauth") {
		logger.Infof("Handling GEE authentication request: %s", path)

		// Get authentication response bytes directly
		// 直接获取认证响应字节
		authBytes, err := h.GEEClient.GetAuthResponseBytes()
		if err != nil {
			logger.Errorf("Failed to get GEE auth response: %v", err)
			return c.JSON(http.StatusInternalServerError, model.BaseAPIResponse[any]{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Failed to get GEE auth response: %v", err),
				Data:    nil,
			})
		}

		// Return raw authentication response bytes
		// 返回原始认证响应字节
		c.Response().Header().Set("Content-Type", "application/octet-stream")
		c.Response().WriteHeader(http.StatusOK)
		_, err = c.Response().Write(authBytes)
		if err != nil {
			logger.Errorf("Failed to write auth response: %v", err)
			return err
		}

		logger.Debugf("GEE auth response sent (length: %d bytes)", len(authBytes))
		return nil
	}

	// For other requests, forward to GEE server
	// 对于其他请求，转发到 GEE 服务器
	method := c.Request().Method
	body := c.Request().Body

	logger.Debugf("GEE proxy request: %s %s", method, path)

	// Forward request to GEE server using GEERelay
	// 使用 GEERelay 转发请求到 GEE 服务器
	resp, err := h.GEEClient.GEERelay(c.Request().Context(), path, method, body)
	if err != nil {
		logger.Errorf("GEE relay failed: %v", err)
		return c.JSON(http.StatusBadGateway, model.BaseAPIResponse[any]{
			Code:    http.StatusBadGateway,
			Message: fmt.Sprintf("GEE relay failed: %v", err),
			Data:    nil,
		})
	}
	defer resp.Body.Close()

	// Copy response headers from GEE server
	// 复制来自 GEE 服务器的响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Response().Header().Set(key, value)
		}
	}

	// Set content type if not present
	// 如果没有 Content-Type 则设置默认值
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Write response status code
	// 写入响应状态码
	c.Response().WriteHeader(resp.StatusCode)

	// Stream response body directly to client
	// 直接将响应体流式传输给客户端
	_, err = io.Copy(c.Response().Writer, resp.Body)
	if err != nil {
		logger.Errorf("Failed to stream GEE response: %v", err)
		return err
	}

	logger.Debugf("GEE proxy response: status=%d, content-type=%s", resp.StatusCode, contentType)
	return nil
}
