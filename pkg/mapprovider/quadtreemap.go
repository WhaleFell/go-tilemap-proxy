package mapprovider

import (
	"fmt"
	"go-map-proxy/pkg/request"
	"net/http"
	"strings"
)

type QuadTreeMapProvider struct {
	Name string

	// https://t.ssl.ak.tiles.virtualearth.net/tiles/a{quadkey}.jpeg?g=14482&n=z&prx=1"
	BaseURL string

	CoordinateType string
	ReferenceURL   string
}

func (qmp *QuadTreeMapProvider) GetMapName() string {
	return qmp.Name
}

// cover xyz coordinate to bing Quadtree
func xyzToQuadkey(x, y, zoom int) string {
	quadkey := ""
	for i := zoom; i > 0; i-- {
		digit := 0
		mask := 1 << (i - 1)
		if (x & mask) != 0 {
			digit++
		}
		if (y & mask) != 0 {
			digit += 2
		}
		quadkey += fmt.Sprint(digit)
	}
	return quadkey
}

func (qmp *QuadTreeMapProvider) GetMapPic(x, y, z int) (*http.Response, error) {
	httpClient := request.DefaultHTTPClient
	quadkey := xyzToQuadkey(x, y, z)
	mapUrl := strings.Replace(qmp.BaseURL, "{quadkey}", quadkey, 1)

	request, err := http.NewRequest(http.MethodGet, mapUrl, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	if qmp.ReferenceURL != "" {
		request.Header.Set("Referer", qmp.ReferenceURL)
	} else {
		request.Header.Set("Referer", "https://www.openstreetmap.org/")
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get map tile, status code: %d", response.StatusCode)
	}

	return response, nil
}

// Bing Satelite Map
var BingSateliteMap = &QuadTreeMapProvider{
	Name:           "Bing Satellite Map(必应卫星图)",
	BaseURL:        "https://t.ssl.ak.tiles.virtualearth.net/tiles/a{quadkey}.jpeg?g=14482&n=z&prx=1",
	CoordinateType: "WGJ84",
	ReferenceURL:   "https://www.bing.com/maps",
}
