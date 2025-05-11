package request

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

type HTTPClientConfig struct {
	// if timeout is nil, use default 10s
	Timeout time.Duration

	// if proxy is empty, use default system proxy
	// if proxy is "direct", use direct connection
	// if proxy string parse failed, use default system proxy
	Proxy string

	// follow 302 redirect
	FollowDirect bool
}

var defaultUserAgent = "Mozilla/5.0 (Linux; Android 10; Pixel 3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.181 Mobile Safari/537.36"

func NewHTTPClient(config *HTTPClientConfig) *http.Client {

	timeout := 10 * time.Second
	if config.Timeout != 0 {
		timeout = config.Timeout
	}

	proxyFunc := http.ProxyFromEnvironment
	proxyInfo := "default system proxy"

	if config.Proxy != "" {
		if config.Proxy == "direct" {
			proxyFunc = nil
			proxyInfo = "direct"
		} else {
			proxyURL, err := url.Parse(config.Proxy)
			if err != nil {
				proxyFunc = http.ProxyFromEnvironment
				proxyInfo = "default system proxy"
			} else {
				proxyFunc = http.ProxyURL(proxyURL)
				proxyInfo = config.Proxy
			}
		}
	}

	log.Printf("HTTP Client use proxy: %s", proxyInfo)

	// checkRedirect is used to handle 302 redirect
	// if FollowDirect is true, it will return http.ErrUseLastResponse and return the follow response.
	var checkRedirect func(req *http.Request, via []*http.Request) error
	if config.FollowDirect {
		checkRedirect = nil
	}

	var HTTPClient = &http.Client{
		Timeout: timeout,

		// handle direct
		CheckRedirect: checkRedirect,

		// transport config
		Transport: &http.Transport{
			Proxy:               proxyFunc,
			MaxIdleConns:        20000,
			MaxIdleConnsPerHost: 10000,

			// TLS config
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			TLSHandshakeTimeout: timeout,

			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 3 * timeout,
			}).DialContext,

			// IdleConnTimeout: 2 * timeout,

			// ExpectContinueTimeout: 1 * time.Second,

			DisableKeepAlives: false,
		},
	}

	return HTTPClient

}

func NewHTTPRequest(method, url string, body io.Reader, useragent ...string) (*http.Request, error) {

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// set user agent
	if len(useragent) > 0 {
		request.Header.Set("User-Agent", useragent[0])
	} else {
		request.Header.Set("User-Agent", defaultUserAgent)
	}

	request.Header.Set("Accept", "*/*")
	request.Header.Set("Accept-Language", "en-US,en;q=0.5")
	// if set Accept-Encoding, the response body will not be decompressed gzip automatically.
	// ref: https://stackoverflow.com/a/38954490/22573614
	// request.Header.Set("Accept-Encoding", "gzip, deflate, br")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Upgrade-Insecure-Requests", "1")
	request.Header.Set("Cache-Control", "max-age=0")

	return request, nil
}

func GetRequest(client *http.Client, url string) ([]byte, error) {
	request, err := NewHTTPRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
