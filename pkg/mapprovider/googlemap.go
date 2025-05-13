package mapprovider

import (
	"fmt"
	"go-map-proxy/pkg/request"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

type GoogleMapProvider struct {
	Name string

	// https://{serverpart}.example.com/{z}/{x}/{y}.png
	BaseURL string

	// Map coordinate type
	// WGJ84: world geographic coordinate system (World Geodetic System 1984)
	// GCJ02: China geodetic coordinate system (国测局 02 坐标系)
	// BD09: Baidu coordinate system (百度坐标系)
	CoordinateType string
}

func (gmp *GoogleMapProvider) GetMapName() string {
	return gmp.Name
}

func (gmp *GoogleMapProvider) GetMapPic(x, y, z int) (*http.Response, error) {
	httpClient := request.GetDefaultHTTPClient()
	mapUrl := gmp.BaseURL
	mapUrl = strings.Replace(mapUrl, "{x}", strconv.Itoa(x), 1) // Replace {x} with the actual x value
	mapUrl = strings.Replace(mapUrl, "{y}", strconv.Itoa(y), 1) // Replace {y} with the actual y value
	mapUrl = strings.Replace(mapUrl, "{z}", strconv.Itoa(z), 1) // Replace {z} with the actual z value
	// replace {serverpart} with 1-3 numbers randomly
	mapUrl = strings.Replace(mapUrl, "{serverpart}", strconv.Itoa(rand.Intn(3)+1), 1)

	// Make a GET request to the map URL
	request, err := http.NewRequest(http.MethodGet, mapUrl, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get map tile, status code: %d", response.StatusCode)
	}

	return response, nil
}

var GmapPureSatellite = &GoogleMapProvider{
	Name:           "Google Pure Satellite",
	CoordinateType: "WGJ84",
	BaseURL:        "https://www.google.com/maps/vt?lyrs=s@189&x={x}&y={y}&z={z}",
}

var GmapPureSatellite2 = &GoogleMapProvider{
	Name:           "Google Pure Satellite 2",
	CoordinateType: "WGJ84",
	BaseURL:        "https://khms{serverpart}.google.com/kh/v=979?x={x}&y={y}&z={z}",
}
