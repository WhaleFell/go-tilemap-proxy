// GoogleEarthEnterprise
// GEE(GoogleEarthEnterprise) Protocol

// Earth Enterprise is the open source release of Google Earth Enterprise, a geospatial application which provides the ability to build and host custom 3D globes and 2D maps:
// ref: https://github.com/google/earthenterprise/
// This package reverse-engineers the authentication method of the Google Earth Pro desktop client
// to enable loading the GEE protocol into third-party mapping applications.
// To enable the retrieval of terrain, satellite imagery, historical satellite imagery,
// and other features from Google Earth in compliance with the GEE protocol, for use within Cesium.

package mapprovider

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"go-map-proxy/pkg/logger"
	"io"
	"net/http"
	"sync"
	"time"
)

type GoogleEarthEngineProvider struct {
	Client *http.Client

	// Default: https://kh.google.com
	BaseURL string

	// Default: /geauth?ct=pro
	AuthURL string

	// Default:
	// 01 00 00 00 03 C9 53 42 0F C9 B7 C1 4A 90 DD CC
	// 62 0D 20 87 F1 DF 4F A5 D6 41 0B 57 E9 DB E4 F6
	// 53 CF 6D 8B 1A 20 0B 23 96 C4 E5 8E C6 E0 46 72
	// 09
	AuthBodyHexString string

	sessionID string
	mu        sync.RWMutex // Protect sessionID from concurrent access
	stopChan  chan struct{}
}

// NewGoogleEarthEngineProvider creates a new GEE provider instance
// 创建一个新的 GEE 提供者实例
func NewGoogleEarthEngineProvider(client *http.Client, baseURL string) *GoogleEarthEngineProvider {
	if client == nil {
		client = http.DefaultClient
	}

	if baseURL == "" {
		baseURL = "https://kh.google.com"
	}

	provider := &GoogleEarthEngineProvider{
		Client:   client,
		BaseURL:  baseURL,
		AuthURL:  "/geauth?ct=pro",
		stopChan: make(chan struct{}),
		// Default auth body hex string for Google Earth Pro
		// Google Earth Pro 的默认认证体十六进制字符串
		AuthBodyHexString: "0100000003C953420FC9B7C14A90DDCC620D2087F1DF4FA5D6410B57E9DBE4F653CF6D8B1A200B2396C4E58EC6E0467209",
	}

	// Perform initial authentication to get SessionID
	// 执行初始认证以获取 SessionID
	if err := provider.authenticateGEE(); err != nil {
		logger.Errorf("Failed to authenticate with GEE: %v", err)
	}

	// Start background goroutine to refresh SessionID every 2 minutes
	// 启动后台 goroutine 每 2 分钟刷新 SessionID
	go provider.startSessionRefreshLoop()

	logger.Infof("GEE Provider initialized successfully with BaseURL: %s", baseURL)
	return provider
}

// authenticateGEE performs authentication with GEE server and extracts SessionID
// 执行 GEE 服务器认证并提取 SessionID
func (gee *GoogleEarthEngineProvider) authenticateGEE() error {
	// Convert hex string to bytes
	// 将十六进制字符串转换为字节
	authBody, err := hex.DecodeString(gee.AuthBodyHexString)
	if err != nil {
		return fmt.Errorf("failed to decode auth body hex string: %w", err)
	}

	// Construct full auth URL
	// 构造完整的认证 URL
	authURL := gee.BaseURL + gee.AuthURL

	// Create POST request
	// 创建 POST 请求
	req, err := http.NewRequest(http.MethodPost, authURL, bytes.NewReader(authBody))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// Execute request
	// 执行请求
	resp, err := gee.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth request failed with status code: %d", resp.StatusCode)
	}

	// Read response body
	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response body: %w", err)
	}

	// Extract SessionID from response
	// 从响应中提取 SessionID
	sessionID, err := gee.extractSessionID(respBody)
	if err != nil {
		return fmt.Errorf("failed to extract SessionID: %w", err)
	}

	// Update sessionID with mutex protection
	// 使用互斥锁保护更新 sessionID
	gee.mu.Lock()
	gee.sessionID = sessionID
	gee.mu.Unlock()

	logger.Infof("Obtain sessionID: %s", sessionID)

	logger.Debugf("GEE authentication successful, SessionID obtained (length: %d)", len(sessionID))
	return nil
}

// extractSessionID extracts SessionID from authentication response
// 从认证响应中提取 SessionID
func (gee *GoogleEarthEngineProvider) extractSessionID(responseBody []byte) (sessionID string, err error) {
	bodyLength := len(responseBody)

	var startIdx, endIdx int

	// Extract SessionID based on response body length
	// 根据响应体长度提取 SessionID
	switch bodyLength {
	case 112:
		startIdx, endIdx = 8, 88
	case 124:
		startIdx, endIdx = 8, 100
	case 136:
		startIdx, endIdx = 8, 112
	case 144:
		startIdx, endIdx = 8, 120
	default:
		return "", fmt.Errorf("unexpected response body length: %d, cannot extract SessionID", bodyLength)
	}

	// Validate indices
	// 验证索引范围
	if endIdx > bodyLength {
		return "", fmt.Errorf("invalid slice range [%d:%d] for body length %d", startIdx, endIdx, bodyLength)
	}

	// Extract and convert bytes to string
	// 提取并将字节转换为字符串
	sessionBytes := responseBody[startIdx:endIdx]
	sessionID = string(sessionBytes)

	if sessionID == "" {
		return "", fmt.Errorf("extracted SessionID is empty")
	}

	return sessionID, nil
}

// startSessionRefreshLoop starts a background goroutine to refresh SessionID every 2 minutes
// 启动后台 goroutine 每 2 分钟刷新 SessionID
func (gee *GoogleEarthEngineProvider) startSessionRefreshLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	logger.Infof("GEE SessionID refresh loop started (interval: 2 minutes)")

	for {
		select {
		case <-ticker.C:
			if err := gee.authenticateGEE(); err != nil {
				logger.Errorf("Failed to refresh GEE SessionID: %v", err)
			} else {
				logger.Debugf("GEE SessionID refreshed successfully")
			}
		case <-gee.stopChan:
			logger.Infof("GEE SessionID refresh loop stopped")
			return
		}
	}
}

// Stop stops the SessionID refresh loop
// 停止 SessionID 刷新循环
func (gee *GoogleEarthEngineProvider) Stop() {
	close(gee.stopChan)
}

// GEERelay forwards GEE requests with SessionID cookie
// 转发 GEE 请求并添加 SessionID cookie
func (gee *GoogleEarthEngineProvider) GEERelay(ctx context.Context, path, method string, body io.Reader) (resp *http.Response, err error) {
	// Construct full URL
	// 构造完整 URL
	fullURL := gee.BaseURL + path

	logger.Infof("fullURL: %s", fullURL)

	// Create HTTP request
	// 创建 HTTP 请求
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create relay request: %w", err)
	}

	req = req.WithContext(ctx)

	// Get current SessionID with read lock
	// 使用读锁获取当前 SessionID
	gee.mu.RLock()
	currentSessionID := gee.sessionID
	gee.mu.RUnlock()

	// Add SessionID cookie
	// 添加 SessionID cookie
	if currentSessionID != "" {
		cookie := &http.Cookie{
			Name:  "SessionId",
			Value: currentSessionID,
			Path:  "/",
		}
		req.AddCookie(cookie)
	} else {
		logger.Warnf("GEE SessionID is empty, request may fail")
	}

	// Set common headers
	// 设置通用请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "*/*")

	// Execute request
	// 执行请求
	resp, err = gee.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute relay request: %w", err)
	}

	logger.Debugf("GEE relay request: %s %s, status: %d", method, fullURL, resp.StatusCode)
	return resp, nil
}

// GetSessionID returns the current SessionID (thread-safe)
// 返回当前 SessionID(线程安全)
func (gee *GoogleEarthEngineProvider) GetSessionID() string {
	gee.mu.RLock()
	defer gee.mu.RUnlock()
	return gee.sessionID
}

// GetAuthResponseBytes performs authentication and returns the raw response bytes
// 执行认证并返回原始响应字节
func (gee *GoogleEarthEngineProvider) GetAuthResponseBytes() (respBody []byte, err error) {
	// Convert hex string to bytes
	// 将十六进制字符串转换为字节
	authBody, err := hex.DecodeString(gee.AuthBodyHexString)
	if err != nil {
		return nil, fmt.Errorf("failed to decode auth body hex string: %w", err)
	}

	// Construct full auth URL
	// 构造完整的认证 URL
	authURL := gee.BaseURL + gee.AuthURL

	// Create POST request
	// 创建 POST 请求
	req, err := http.NewRequest(http.MethodPost, authURL, bytes.NewReader(authBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// Execute request
	// 执行请求
	resp, err := gee.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth request failed with status code: %d", resp.StatusCode)
	}

	// Read and return response body
	// 读取并返回响应体
	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth response body: %w", err)
	}

	logger.Debugf("GEE auth response bytes obtained (length: %d)", len(respBody))
	return respBody, nil
}
