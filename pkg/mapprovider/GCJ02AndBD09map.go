// GCJ02 coordinate tile map convert to WGS84 tile map with pixel-level correction
// 同时支持 GCJ02（高德）和 BD09（百度）坐标系的图像获取

package mapprovider

import (
	"bytes"
	"fmt"
	"go-map-proxy/pkg/logger"
	"go-map-proxy/pkg/request"
	"image"
	"image/jpeg"
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

// Convert GCJ02 to BD09 坐标转换：从 GCJ02 转换到 百度 BD09
func gcj02ToBd09(gcjLat, gcjLon float64) (bdLat, bdLon float64) {
	x := gcjLon
	y := gcjLat
	z := math.Sqrt(x*x+y*y) + 0.00002*math.Sin(y*math.Pi*3000.0/180.0)
	theta := math.Atan2(y, x) + 0.000003*math.Cos(x*math.Pi*3000.0/180.0)
	bdLon = z*math.Cos(theta) + 0.0065
	bdLat = z*math.Sin(theta) + 0.006
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
	if lon < 73.675379 || lon > 135.026311 || lat < 18.197701 || lat > 53.458804 {
		return false
	}
	if lon >= 119.0 && lon <= 123.0 && lat >= 21.5 && lat <= 25.5 {
		return false // exclude Taiwan
	}
	return true
}

// tms xyz to google xyz
func tmsToGoogleXY(x, y, z int) (gx, gy, gz int) {
	gz = z
	// TMS y coordinate is inverted
	gy = (1 << z) - 1 - y
	gx = x
	return
}

// GCJ02MapProvider 支持 GCJ02 和 BD09，可通过 CoordinateType 字段区分
// GCJ02MapProvider support GCJ02 and BD09, which can be distinguished by the CoordinateType field
type GCJ02MapProvider struct {
	Name           string
	BaseURL        string
	ReferenceURL   string
	CoordinateType string // 可为 "GCJ02" 或 "BD09" Can be "GCJ02" or "BD09"
	IsTMS          bool   // 是否为 TMS 坐标系 Whether it is a TMS coordinate system
}

func (gcjmap *GCJ02MapProvider) GetMapName() string {
	return gcjmap.Name
}

func (gcjmap *GCJ02MapProvider) GetMapPic(x, y, z int) (*http.Response, error) {
	logger.Debugf("GetMapPic: %s, %d, %d, %d", gcjmap.Name, x, y, z)

	httpClient := request.DefaultHTTPClient

	wgsLonTopLeft, wgsLatTopLeft := pixelXYToLonLat(x*256, y*256, z)
	wgsLonBottomRight, wgsLatBottomRight := pixelXYToLonLat((x+1)*256-1, (y+1)*256-1, z)

	if !isInMainlandChina(wgsLatTopLeft, wgsLonTopLeft) && !isInMainlandChina(wgsLatBottomRight, wgsLonBottomRight) {
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

	tile := image.NewRGBA(image.Rect(0, 0, 256, 256))

	// temporary cache for source tiles
	sourceTileCache := make(map[string]image.Image)

	for py := 0; py < 256; py++ {
		for px := 0; px < 256; px++ {
			wgsLon, wgsLat := pixelXYToLonLat(x*256+px, y*256+py, z)
			gcjLat, gcjLon := wgs84ToGCJ02(wgsLat, wgsLon)

			if gcjmap.CoordinateType == "BD09" {
				gcjLat, gcjLon = gcj02ToBd09(gcjLat, gcjLon) // 转 BD09
			}

			gx, gy := lonLatToPixelXY(gcjLon, gcjLat, z)
			tx := gx / 256
			ty := gy / 256
			sx := gx % 256
			sy := gy % 256

			tileKey := fmt.Sprintf("%d_%d_%d", tx, ty, z)
			// get cached tile
			srcTile, ok := sourceTileCache[tileKey]

			if !ok {
				logger.Debugf("Tile %s not found in cache, fetching from %s", tileKey, gcjmap.BaseURL)
				// if isTMS, convert to Google XYZ
				if gcjmap.IsTMS {
					tx, ty, z = tmsToGoogleXY(tx, ty, z)
				}

				url := strings.Replace(gcjmap.BaseURL, "{x}", strconv.Itoa(tx), 1)
				url = strings.Replace(url, "{y}", strconv.Itoa(ty), 1)
				url = strings.Replace(url, "{z}", strconv.Itoa(z), 1)
				req, _ := http.NewRequest(http.MethodGet, url, nil)
				req.Header.Set("User-Agent", request.DefaultUserAgent)
				if gcjmap.ReferenceURL != "" {
					req.Header.Set("Referer", gcjmap.ReferenceURL)
				}
				resp, err := httpClient.Do(req)

				// handle error
				if err != nil || resp.StatusCode != http.StatusOK {
					logger.Errorf("Failed to fetch tile %s: %v", tileKey, err)
					continue
				}

				// check content type
				contentType := resp.Header.Get("Content-Type")
				var img image.Image
				if strings.Contains(contentType, "image/png") {
					img, err = png.Decode(resp.Body)
				} else if strings.Contains(contentType, "image/jpeg") {
					img, err = jpeg.Decode(resp.Body)
				} else {
					logger.Errorf("Unsupported content type %s for tile %s", contentType, tileKey)
					resp.Body.Close()
					return nil, fmt.Errorf("unsupported response content type %s", contentType)
				}

				resp.Body.Close()
				if err != nil {
					logger.Errorf("Failed to decode tile %s: %v", tileKey, err)
					continue
				}
				sourceTileCache[tileKey] = img
				srcTile = img
			}

			if rgbaImg, ok := srcTile.(*image.RGBA); ok {
				if sx < 256 && sy < 256 {
					tile.SetRGBA(px, py, rgbaImg.RGBAAt(sx, sy))
				}
			} else {
				tile.Set(px, py, srcTile.At(sx, sy))
			}
		}
	}

	var buf bytes.Buffer
	// prelocalize buffer
	buf.Grow(256 * 256 * 4) // 256x256 RGBA

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

// ref: http://www.maps5.com/s/list1/50.html
// var BaiduSatelliteMap = &GCJ02MapProvider{
// 	Name: "Baidu Map 百度地图影像图",
// 	// https://maponline0.bdimg.com/starpic/?qt=satepc&u=x=768;y=160;z=12;v=009;type=sate&fm=46&app=webearth2&v=009&udt=20250515
// 	BaseURL:        "https://maponline0.bdimg.com/starpic/?qt=satepc&u=x={x};y={y};z={z};v=009;type=sate&fm=46&app=webearth2&v=009&udt=20250515",
// 	ReferenceURL:   "https://map.baidu.com/",
// 	CoordinateType: "BD09",
// 	IsTMS:          true,
// }
