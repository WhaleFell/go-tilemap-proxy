package request_test

import (
	"fmt"
	"go-map-proxy/pkg/request"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestHTTPClient(t *testing.T) {
	config := &request.HTTPClientConfig{
		Timeout:      10 * time.Second,
		FollowDirect: true,
		Proxy:        "",
	}
	client := request.NewHTTPClient(config)

	request, err := http.NewRequest("GET", "http://baidu.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("Failed to do request: %v", err)
	}
	defer response.Body.Close()

	// get response

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// t.Logf("Response body: %s", body)
	fmt.Printf("Response body: %s\n", body)
	fmt.Printf("Response status code: %d\n", response.StatusCode)

}
