package common

import (
	"fmt"
	"go-map-proxy/internal/config"
	"go-map-proxy/internal/model"
	"go-map-proxy/pkg/request"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

var (
	HTTPClient     *http.Client
	initClientOnce sync.Once
)

// proxy url
// http://example.com/proxy?url=http://example.com
func URLProxy(c echo.Context) error {
	// init http client
	initClientOnce.Do(func() {
		HTTPClient = request.NewHTTPClient(&request.HTTPClientConfig{
			Timeout:      10 * time.Second,
			Proxy:        config.Cfg.Proxy,
			FollowDirect: true,
		})
	})

	proxyUrl := c.QueryParam("url")
	if proxyUrl == "" {
		return c.JSON(200, model.BaseAPIResponse[any]{
			Code:    400,
			Message: "url is required",
			Data:    nil,
		})
	}

	parsedURL, err := url.Parse(proxyUrl)
	if err != nil {
		return c.JSON(200, model.BaseAPIResponse[any]{
			Code:    400,
			Message: fmt.Sprintf("failed to parse URL: %v", err),
			Data:    nil,
		})
	}

	methods := c.Request().Method

	proxyRequest, err := http.NewRequest(methods, parsedURL.String(), c.Request().Body)
	if err != nil {
		return c.JSON(200, model.BaseAPIResponse[any]{
			Code:    400,
			Message: fmt.Sprintf("failed to create HTTP request: %v", err),
			Data:    nil,
		})
	}

	// set proxy Request header
	for key, values := range c.Request().Header {
		if key == "Host" {
			// skip Host header
			continue
		}
		for _, value := range values {
			proxyRequest.Header.Add(key, value)
		}
	}

	// set host header
	// proxyRequest.Header.Add("Host", parsedURL.Host)
	proxyRequest.Host = parsedURL.Host

	// send proxy request
	proxyResponse, err := HTTPClient.Do(proxyRequest)
	if err != nil {
		return c.JSON(200, model.BaseAPIResponse[any]{
			Code:    400,
			Message: fmt.Sprintf("failed to send proxy HTTP request: %v", err),
			Data:    nil,
		})
	}

	// handle response directly
	responseCode := proxyResponse.StatusCode
	directlyStatusCode := []int{301, 302, 303, 307, 308}
	for _, code := range directlyStatusCode {
		if proxyResponse.StatusCode == code {
			responseCode = 200
		}
	}

	// set response header
	for key, values := range proxyResponse.Header {
		// skip Content-Length header
		// if key == "Content-Length" {
		// 	continue
		// }
		for _, value := range values {
			c.Response().Header().Set(key, value)
		}
	}

	contentType := proxyResponse.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	defer proxyResponse.Body.Close()

	// return c.Stream(responseCode, contentType, proxyResponse.Body)

	// use io.Copy
	c.Response().WriteHeader(responseCode)
	io.Copy(c.Response().Writer, proxyResponse.Body)
	return nil
}
