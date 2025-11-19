package mapprovider

import (
	"go-map-proxy/pkg/request"
	"testing"
	"time"
)

func TestGoogleEarthEngineProtocol(t *testing.T) {
	// Test implementation here

	httpClient := request.NewHTTPClient(&request.HTTPClientConfig{
		Timeout: 10,
	})

	_ = NewGoogleEarthEngineProvider(httpClient, "")

	time.Sleep(10 * time.Second)

}
