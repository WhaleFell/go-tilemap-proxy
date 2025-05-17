// GCJ02 coordinate tile map convert to WGS84 tile map with pixel-level correction

package mapprovider

import (
	"bytes"
	"fmt"
	"go-map-proxy/pkg/request"
	"image"
	"image/png"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
)

// Convert tile (x, y, z) to WGS84 top-left lon/lat
// func tileXYToLonLat(x, y, z int) (lon, lat float64) {
// 	n := math.Pow(2, float64(z))
// 	lon = float64(x)/n*360.0 - 180.0
// 	latRad := math.Atan(math.Sinh(math.Pi * (1 - 2*float64(y)/n)))
// 	lat = latRad * 180.0 / math.Pi
// 	return
// }

// Convert lon/lat to pixel position
func lonLatToPixelXY(lon, lat float64, z int) (px, py int) {
	scale := math.Pow(2, float64(z)) * 256
	x := (lon + 180.0) / 360.0
	siny := math.Sin(lat * math.Pi / 180.0)
	y := 0.5 - math.Log((1+siny)/(1-siny))/(4*math.Pi)

	px = int(x * scale)
	py = int(y * scale)
	return
}

// Convert pixel to lon/lat
func pixelXYToLonLat(px, py int, z int) (lon, lat float64) {
	scale := math.Pow(2, float64(z)) * 256
	x := float64(px) / scale
	y := float64(py) / scale

	lon = x*360.0 - 180.0
	n := math.Pi - 2.0*math.Pi*y
	lat = 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	return
}

// Convert WGS84 to GCJ02
func wgs84ToGCJ02(wgsLat, wgsLon float64) (gcjLat, gcjLon float64) {
	dLat := transformLat(wgsLon-105.0, wgsLat-35.0)
	dLon := transformLon(wgsLon-105.0, wgsLat-35.0)
	radLat := wgsLat / 180.0 * math.Pi
	magic := math.Sin(radLat)
	magic = 1 - 0.00669342162296594323*magic*magic
	sqrtMagic := math.Sqrt(magic)
	dLat = (dLat * 180.0) / ((6378245.0 * (1 - 0.00669342162296594323)) / (magic * sqrtMagic) * math.Pi)
	dLon = (dLon * 180.0) / (6378245.0 / sqrtMagic * math.Cos(radLat) * math.Pi)
	gcjLat = wgsLat + dLat
	gcjLon = wgsLon + dLon
	return
}

// latitude offset calculation (GCJ02 encrypted)
func transformLat(x, y float64) float64 {
	ret := -100.0 + 2.0*x + 3.0*y + 0.2*y*y + 0.1*x*y + 0.2*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*math.Pi) + 20.0*math.Sin(2.0*x*math.Pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(y*math.Pi) + 40.0*math.Sin(y/3.0*math.Pi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(y/12.0*math.Pi) + 320*math.Sin(y*math.Pi/30.0)) * 2.0 / 3.0
	return ret
}

// longitude offset calculation (GCJ02 encrypted)
func transformLon(x, y float64) float64 {
	ret := 300.0 + x + 2.0*y + 0.1*x*x + 0.1*x*y + 0.1*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*math.Pi) + 20.0*math.Sin(2.0*x*math.Pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(x*math.Pi) + 40.0*math.Sin(x/3.0*math.Pi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(x/12.0*math.Pi) + 300.0*math.Sin(x/30.0*math.Pi)) * 2.0 / 3.0
	return ret
}

// Check if the coordinate is in mainland China (excluding Hong Kong, Macau, and Taiwan)
func isInMainlandChina(lat, lon float64) bool {
	// Mainland China coordinate range
	if lon < 73.675379 || lon > 135.026311 || lat < 18.197701 || lat > 53.458804 {
		return false
	}

	// exclude Taiwan
	if lon >= 119.0 && lon <= 123.0 && lat >= 21.5 && lat <= 25.5 {
		return false
	}

	// exclude Hong Kong
	// if lon >= 113.8 && lon <= 114.4 && lat >= 22.2 && lat <= 22.6 {
	// 	return false
	// }

	// exclude Macau
	// if lon >= 113.5 && lon <= 113.6 && lat >= 22.1 && lat <= 22.3 {
	// 	return false
	// }

	return true
}

type GCJ02MapProvider struct {
	Name           string
	BaseURL        string
	ReferenceURL   string
	CoordinateType string
}

func (gcjmap *GCJ02MapProvider) GetMapName() string {
	return gcjmap.Name
}

func (gcjmap *GCJ02MapProvider) GetMapPic(x, y, z int) (*http.Response, error) {
	httpClient := request.DefaultHTTPClient

	wgsLonTopLeft, wgsLatTopLeft := pixelXYToLonLat(x*256, y*256, z)
	wgsLonBottomRight, wgsLatBottomRight := pixelXYToLonLat((x+1)*256-1, (y+1)*256-1, z)

	// Check if the tile is outside China
	if !isInMainlandChina(wgsLatTopLeft, wgsLonTopLeft) && !isInMainlandChina(wgsLatBottomRight, wgsLonBottomRight) {
		// if the tile is outside China, return the original tile directly
		url := strings.Replace(gcjmap.BaseURL, "{x}", strconv.Itoa(x), 1)
		url = strings.Replace(url, "{y}", strconv.Itoa(y), 1)
		url = strings.Replace(url, "{z}", strconv.Itoa(z), 1)
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("User-Agent", request.DefaultUserAgent)
		if gcjmap.ReferenceURL != "" {
			req.Header.Set("Referer", gcjmap.ReferenceURL)
		} else {
			req.Header.Set("Referer", "https://www.amap.com/")
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	// Create destination tile image
	tile := image.NewRGBA(image.Rect(0, 0, 256, 256))

	// Create destination tile buffer
	// Create a cache for the source tile images
	sourceTileCache := make(map[string]image.Image)

	// retrieve the destination tile pixel
	// For each pixel in the tile...
	for py := range 256 {
		for px := range 256 {
			// Convert target pixel to WGS84 coordinate
			wgsLon, wgsLat := pixelXYToLonLat(x*256+px, y*256+py, z)
			// Convert WGS84 coordinate to GCJ02
			gcjLat, gcjLon := wgs84ToGCJ02(wgsLat, wgsLon)
			// Get GCJ02 pixel location
			gx, gy := lonLatToPixelXY(gcjLon, gcjLat, z)

			tx := gx / 256
			ty := gy / 256
			sx := gx % 256
			sy := gy % 256

			tileKey := fmt.Sprintf("%d_%d_%d", tx, ty, z)
			srcTile, ok := sourceTileCache[tileKey]

			// Fetch source tile if not cached
			if !ok {
				url := strings.Replace(gcjmap.BaseURL, "{x}", strconv.Itoa(tx), 1)
				url = strings.Replace(url, "{y}", strconv.Itoa(ty), 1)
				url = strings.Replace(url, "{z}", strconv.Itoa(z), 1)

				req, _ := http.NewRequest(http.MethodGet, url, nil)
				req.Header.Set("User-Agent", request.DefaultUserAgent)
				if gcjmap.ReferenceURL != "" {
					req.Header.Set("Referer", gcjmap.ReferenceURL)
				}

				resp, err := httpClient.Do(req)
				// TODO: handle empty response

				if err != nil || resp.StatusCode != http.StatusOK {
					continue
				}
				img, err := png.Decode(resp.Body)
				resp.Body.Close()
				if err != nil {
					continue
				}
				sourceTileCache[tileKey] = img
				srcTile = img
			}

			// Copy source pixel to target pixel
			if rgbaImg, ok := srcTile.(*image.RGBA); ok {
				if sx < 256 && sy < 256 {
					color := rgbaImg.RGBAAt(sx, sy)
					tile.SetRGBA(px, py, color)
				}
			} else {
				c := srcTile.At(sx, sy)
				tile.Set(px, py, c)
			}
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, tile)
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(&buf),
		Header:     http.Header{"Content-Type": []string{"image/png"}},
	}, nil
}

var AmapRoadMap = &GCJ02MapProvider{
	Name:           "Amap Road Map 高德路网",
	BaseURL:        "https://webst01.is.autonavi.com/appmaptile?style=8&x={x}&y={y}&z={z}",
	ReferenceURL:   "https://www.amap.com/",
	CoordinateType: "GCJ02",
}
