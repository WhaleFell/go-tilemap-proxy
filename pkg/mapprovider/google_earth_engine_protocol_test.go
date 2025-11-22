package mapprovider

import (
	"fmt"
	"go-map-proxy/pkg/request"
	"testing"
	"time"
)

func TestGoogleEarthEngineProtocol(t *testing.T) {
	// Test implementation here

	httpClient := request.NewHTTPClient(&request.HTTPClientConfig{
		Timeout: 10 * time.Second,
	})

	geeProvider := NewGoogleEarthEngineProvider(httpClient, "")

	authResp, err := geeProvider.GetAuthResponseBytes()
	if err != nil {
		t.Fatalf("Failed to get GEE auth response: %v", err)
	}
	fmt.Printf("authResp len: %d", len(authResp))

	time.Sleep(10 * time.Second)

}
