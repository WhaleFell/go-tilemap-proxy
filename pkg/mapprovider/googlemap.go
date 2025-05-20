package mapprovider

import (
	"fmt"
	"go-map-proxy/pkg/request"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

// ref:
// 1. https://wiki.openstreetmap.org/wiki/Raster_tile_providers

type GoogleMapProvider struct {
	Name string

	// https://{serverpart}.example.com/{z}/{x}/{y}.png
	BaseURL string

	// Map coordinate type
	// WGJ84: world geographic coordinate system (World Geodetic System 1984)
	// GCJ02: China geodetic coordinate system (国测局 02 坐标系)
	// BD09: Baidu coordinate system (百度坐标系)
	CoordinateType string

	ReferenceURL string
}

func (gmp *GoogleMapProvider) GetMapName() string {
	return gmp.Name
}

func (gmp *GoogleMapProvider) GetMapPic(x, y, z int) (*http.Response, error) {
	httpClient := request.DefaultHTTPClient
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

	if gmp.ReferenceURL != "" {
		request.Header.Set("Referer", gmp.ReferenceURL)
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

var OpenStreetMapStandard = &GoogleMapProvider{
	Name:           "OpenStreetMap Standard",
	CoordinateType: "WGJ84",
	BaseURL:        "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
}

var OpenStreetMapPublicGPS = &GoogleMapProvider{
	Name:           "OpenStreetMap Public GPS",
	CoordinateType: "WGJ84",
	BaseURL:        "https://gps.tile.openstreetmap.org/lines/{z}/{x}/{y}.png",
}

// TraceStrack Topo Map
// Note that the tile photo pixel size is 512x512
var TraceStrackTopoMap = &GoogleMapProvider{
	Name:           "TraceStrack Topo Map",
	CoordinateType: "WGJ84",
	BaseURL:        "https://tile.tracestrack.com/topo__/{z}/{x}/{y}.png?key=383118983d4a867dd2d367451720d724",
	ReferenceURL:   "https://www.openstreetmap.org/",
}

// Arcgis Satelite
var ArcgisSatelite = &GoogleMapProvider{
	Name:           "Arcgis Satelite",
	CoordinateType: "WGJ84",
	BaseURL:        "https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}",
	ReferenceURL:   "https://www.arcgis.com/home/webmap/viewer.html?webmap=0c4f2a1b8d5e4f3b8c7e6f3b8c7e6f3b",
}

var GoogleHybridOffsetMap = &GoogleMapProvider{
	Name:           "Google Hybrid Offset Map",
	CoordinateType: "WGJ84",
	BaseURL:        "https://khms${serverpart}.google.com/kh/v=979?x=${x}&y=${y}&z=${z}",
}

// 天地图
// ref: https://www.tianditu.gov.cn/
// Use CGCS2000 coordinate system, it is the same as WGS84, but has a cm level offset.
var TianDiTuSatellite = &GoogleMapProvider{
	Name:           "TianDiTu Satellite 天地图卫星影像",
	CoordinateType: "CGCS2000",
	BaseURL:        "https://t0.tianditu.gov.cn/img_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=img&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=75f0434f240669f4a2df6359275146d2",
	ReferenceURL:   "https://map.tianditu.gov.cn/",
}

// 天地图路网
// TianDiTu Road Map
var TianDiTuRoad = &GoogleMapProvider{
	Name:           "TianDiTu Road Map 天地图路网",
	CoordinateType: "CGCS2000",
	BaseURL:        "https://t0.tianditu.gov.cn/cia_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=cia&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILEMATRIX={z}&TILEROW={y}&TILECOL={x}&tk=75f0434f240669f4a2df6359275146d2",
	ReferenceURL:   "https://map.tianditu.gov.cn/",
}

// Map here satelite maps.here.com
// https://maps.hereapi.com/v3/background/mc/5/6/13/jpeg?xnlp=CL_JSMv3.1.63.1&apikey=xGVgeXEdD-GKS1ABa4dziKYCx94eKQIjqlMWAZOfrz0&style=satellite.day&ppi=200&size=512&lang=zh&lang2=en
var MapHereSatelite = &GoogleMapProvider{
	Name:           "Map Here Satelite",
	CoordinateType: "WGJ84",
	BaseURL:        "https://maps.hereapi.com/v3/background/mc/{z}/{x}/{y}/jpeg?xnlp=CL_JSMv3.1.63.1&apikey=xGVgeXEdD-GKS1ABa4dziKYCx94eKQIjqlMWAZOfrz0&style=satellite.day&ppi=200&size=512&lang=zh&lang2=en",
	ReferenceURL:   "https://maps.here.com/",
}

// data.maptiler.com contour line map (elevation map) 等高线地图
var MapTilerContour = &GoogleMapProvider{
	Name:           "Map Tiler Contour",
	CoordinateType: "WGJ84",
	BaseURL:        "https://api.maptiler.com/tiles/terrain-rgb-v2/{z}/{x}/{y}.webp?key=KjOUJBOUa2Tw2LxazlpQ&mtsid=95a4a50f-1858-47df-a5d6-ecba3a179b55",
	ReferenceURL:   "https://data.maptiler.com/",
}
