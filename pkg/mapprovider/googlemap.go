package mapprovider

import (
	"fmt"
	"go-map-proxy/pkg/logger"
	"go-map-proxy/pkg/request"
	"io"
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// ref:
// 1. https://wiki.openstreetmap.org/wiki/Raster_tile_providers

type GoogleMapProvider struct {
	*TileMapMetadata

	// https://{serverpart}.example.com/{z}/{x}/{y}.png
	BaseURL string

	// Map coordinate type
	// WGJ84: world geographic coordinate system (World Geodetic System 1984)
	// GCJ02: China geodetic coordinate system (国测局 02 坐标系)
	// BD09: Baidu coordinate system (百度坐标系)
	// CoordinateType string

	ReferenceURL string
}

func (gmp *GoogleMapProvider) GetMapMetadata() *TileMapMetadata {
	return gmp.TileMapMetadata
}

// replace {serverpart:<item1>,<item2>,<item3>} randomly with item1, item2, or item3
func (gmp *GoogleMapProvider) ReplaceServerPart(template string) string {
	re := regexp.MustCompile(`\{serverpart:([^}]+)\}`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		// extract "1,2,3"
		values := re.FindStringSubmatch(match)
		if len(values) != 2 {
			return match // fallback
		}
		options := strings.Split(values[1], ",")
		selected := options[rand.Intn(len(options))]
		return selected
	})
}

func (gmp *GoogleMapProvider) GetMapPic(x, y, z int) (*http.Response, error) {

	// check zoom level
	if z < gmp.MinZoom || z > gmp.MaxZoom {
		return nil, fmt.Errorf("map: %s zoom level %d is out of range [%d, %d]", gmp.Name, z, gmp.MinZoom, gmp.MaxZoom)
	}

	httpClient := request.DefaultHTTPClient
	mapUrl := gmp.BaseURL
	mapUrl = strings.Replace(mapUrl, "{x}", strconv.Itoa(x), 1) // Replace {x} with the actual x value
	mapUrl = strings.Replace(mapUrl, "{y}", strconv.Itoa(y), 1) // Replace {y} with the actual y value
	mapUrl = strings.Replace(mapUrl, "{z}", strconv.Itoa(z), 1) // Replace {z} with the actual z value
	// replace `{serverpart:<item1>,<item2>,<item3>}` randomly with item1, item2, or item3
	mapUrl = gmp.ReplaceServerPart(mapUrl)

	logger.Debugf("[GoogleMapProvider: %s] tile URL: %s", gmp.Name, mapUrl)

	// Make a GET request to the map URL
	request, err := http.NewRequest(http.MethodGet, mapUrl, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	if gmp.ReferenceURL != "" {
		request.Header.Set("Referer", gmp.ReferenceURL)
		request.Header.Set("Origin", gmp.ReferenceURL)
	} else {
		request.Header.Set("Referer", "https://www.openstreetmap.org/")
		request.Header.Set("Origin", "https://www.openstreetmap.org/")
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		// read response body for debug
		defer response.Body.Close()
		body, _ := io.ReadAll(response.Body)
		logger.Warnf("[GoogleMapProvider: %s] tile error response: %s", gmp.Name, body)

		return nil, fmt.Errorf("failed to get map tile, status code: %d", response.StatusCode)
	}

	return response, nil
}

var GmapPureSatellite = &GoogleMapProvider{
	// Name:           "Google Pure Satellite",
	// CoordinateType: "WGJ84",
	TileMapMetadata: &TileMapMetadata{
		Name:           "Google Pure Satellite",
		ID:             "google_pure_satellite",
		MinZoom:        0,
		MaxZoom:        20,
		MapSize:        MapSize256,
		MapType:        MapTypeRaster,
		ContentType:    MapContentTypePNG,
		CoordinateType: CoordinateTypeWGS84,
	},

	BaseURL: "https://www.google.com/maps/vt?lyrs=s@189&x={x}&y={y}&z={z}",
}

var GmapPureSatellite2 = &GoogleMapProvider{
	TileMapMetadata: &TileMapMetadata{
		Name:           "Google Pure Satellite 2",
		ID:             "google_pure_satellite_2",
		MinZoom:        0,
		MaxZoom:        20,
		MapSize:        MapSize256,
		MapType:        MapTypeRaster,
		ContentType:    MapContentTypePNG,
		CoordinateType: CoordinateTypeWGS84,
	},
	BaseURL: "https://khms{serverpart:1,2,3}.google.com/kh/v=979?x={x}&y={y}&z={z}",
}

var OpenStreetMapStandard = &GoogleMapProvider{
	TileMapMetadata: &TileMapMetadata{
		Name:           "OpenStreetMap Standard",
		ID:             "open_street_map_standard",
		MinZoom:        0,
		MaxZoom:        18,
		MapSize:        MapSize256,
		MapType:        MapTypeRaster,
		ContentType:    MapContentTypePNG,
		CoordinateType: CoordinateTypeWGS84,
	},
	BaseURL: "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
}

var OpenStreetMapPublicGPS = &GoogleMapProvider{
	TileMapMetadata: &TileMapMetadata{
		Name:           "OpenStreetMap Public GPS",
		ID:             "open_street_map_public_gps",
		MinZoom:        0,
		MaxZoom:        18,
		MapSize:        MapSize256,
		MapType:        MapTypeRaster,
		ContentType:    MapContentTypePNG,
		CoordinateType: CoordinateTypeWGS84,
	},
	BaseURL: "https://gps.tile.openstreetmap.org/lines/{z}/{x}/{y}.png",
}

// TraceStrack Topo Map
// Note that the tile photo pixel size is 512x512
// https://tile.tracestrack.com/topo__/13/6676/3544.webp?key=383118983d4a867dd2d367451720d724
var TraceStrackTopoMap = &GoogleMapProvider{
	TileMapMetadata: &TileMapMetadata{
		Name:           "TraceStrack Topo Map",
		ID:             "trace_strack_topo_map",
		MinZoom:        0,
		MaxZoom:        18,
		MapSize:        MapSize512,
		MapType:        MapTypeRaster,
		ContentType:    MapContentTypeWebP,
		CoordinateType: CoordinateTypeWGS84,
	},
	BaseURL:      "https://tile.tracestrack.com/topo__/{z}/{x}/{y}.webp?key=383118983d4a867dd2d367451720d724",
	ReferenceURL: "https://www.openstreetmap.org/",
}

// Arcgis Satelite
var ArcgisSatelite = &GoogleMapProvider{
	TileMapMetadata: &TileMapMetadata{
		Name:           "(ESRI) Arcgis Satelite",
		ID:             "arcgis_satelite",
		MinZoom:        0,
		MaxZoom:        18,
		MapSize:        MapSize256,
		MapType:        MapTypeRaster,
		ContentType:    MapContentTypePNG,
		CoordinateType: CoordinateTypeWGS84,
	},
	BaseURL:      "https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}",
	ReferenceURL: "https://www.arcgis.com/home/webmap/viewer.html?webmap=0c4f2a1b8d5e4f3b8c7e6f3b8c7e6f3b",
}

var GoogleHybridOffsetMap = &GoogleMapProvider{
	TileMapMetadata: &TileMapMetadata{
		Name:           "Google Hybrid Offset Map",
		ID:             "google_hybrid_offset_map",
		MinZoom:        0,
		MaxZoom:        18,
		MapSize:        MapSize256,
		MapType:        MapTypeRaster,
		ContentType:    MapContentTypePNG,
		CoordinateType: CoordinateTypeWGS84,
	},
	BaseURL: "https://khms${serverpart:1,2,3}.google.com/kh/v=979?x=${x}&y=${y}&z=${z}",
}

// 天地图
// ref: https://www.tianditu.gov.cn/
// Use CGCS2000 coordinate system, it is the same as WGS84, but has a cm level offset.
var TianDiTuSatellite = &GoogleMapProvider{
	TileMapMetadata: &TileMapMetadata{
		Name:           "TianDiTu Satellite 天地图卫星影像",
		ID:             "tianditu_satellite",
		MinZoom:        0,
		MaxZoom:        18,
		MapSize:        MapSize256,
		MapType:        MapTypeRaster,
		ContentType:    MapContentTypePNG,
		CoordinateType: CoordinateTypeCGCS2000,
	},
	BaseURL:      "https://t0.tianditu.gov.cn/img_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=img&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILECOL={x}&TILEROW={y}&TILEMATRIX={z}&tk=75f0434f240669f4a2df6359275146d2",
	ReferenceURL: "https://map.tianditu.gov.cn/",
}

// 天地图路网
// TianDiTu Road Map
var TianDiTuRoad = &GoogleMapProvider{
	TileMapMetadata: &TileMapMetadata{
		Name:           "TianDiTu Road Map 天地图路网",
		ID:             "tianditu_road",
		MinZoom:        0,
		MaxZoom:        18,
		MapSize:        MapSize256,
		MapType:        MapTypeRaster,
		ContentType:    MapContentTypePNG,
		CoordinateType: CoordinateTypeCGCS2000,
	},
	BaseURL:      "https://t0.tianditu.gov.cn/cia_w/wmts?SERVICE=WMTS&REQUEST=GetTile&VERSION=1.0.0&LAYER=cia&STYLE=default&TILEMATRIXSET=w&FORMAT=tiles&TILEMATRIX={z}&TILEROW={y}&TILECOL={x}&tk=75f0434f240669f4a2df6359275146d2",
	ReferenceURL: "https://map.tianditu.gov.cn/",
}

// Map here satelite maps.here.com
// https://maps.hereapi.com/v3/background/mc/5/6/13/jpeg?xnlp=CL_JSMv3.1.63.1&apikey=xGVgeXEdD-GKS1ABa4dziKYCx94eKQIjqlMWAZOfrz0&style=satellite.day&ppi=200&size=512&lang=zh&lang2=en
var MapHereSatelite = &GoogleMapProvider{
	TileMapMetadata: &TileMapMetadata{
		Name:           "Map Here Satelite",
		ID:             "map_here_satelite",
		MinZoom:        0,
		MaxZoom:        18,
		MapSize:        MapSize512, // Map size is 512x512
		MapType:        MapTypeRaster,
		ContentType:    MapContentTypePNG,
		CoordinateType: CoordinateTypeWGS84,
	},
	BaseURL:      "https://maps.hereapi.com/v3/background/mc/{z}/{x}/{y}/jpeg?xnlp=CL_JSMv3.1.63.1&apikey=xGVgeXEdD-GKS1ABa4dziKYCx94eKQIjqlMWAZOfrz0&style=satellite.day&ppi=200&size=512&lang=zh&lang2=en",
	ReferenceURL: "https://maps.here.com/",
}

// data.maptiler.com contour line map (elevation map) 等高线地图
var MapTilerContour = &GoogleMapProvider{
	TileMapMetadata: &TileMapMetadata{
		Name:           "Map Tiler Contour",
		ID:             "map_tiler_contour",
		MinZoom:        0,
		MaxZoom:        18,
		MapSize:        MapSize256,
		MapType:        MapTypeRaster,
		ContentType:    MapContentTypeWebP,
		CoordinateType: CoordinateTypeWGS84,
	},
	BaseURL:      "https://api.maptiler.com/tiles/terrain-rgb-v2/{z}/{x}/{y}.webp?key=KjOUJBOUa2Tw2LxazlpQ&mtsid=95a4a50f-1858-47df-a5d6-ecba3a179b55",
	ReferenceURL: "https://data.maptiler.com/",
}

var OpenStreetMapCyclOSM = &GoogleMapProvider{
	TileMapMetadata: &TileMapMetadata{
		Name:           "OpenStreetMap CyclOSM",
		ID:             "openstreetmap_cyclosm",
		MinZoom:        0,
		MaxZoom:        18,
		MapSize:        MapSize256,
		MapType:        MapTypeRaster,
		ContentType:    MapContentTypePNG,
		CoordinateType: CoordinateTypeWGS84,
	},

	BaseURL:      "https://{serverpart:a,b,c}.tile-cyclosm.openstreetmap.fr/cyclosm/{z}/{x}/{y}.png",
	ReferenceURL: "https://www.openstreetmap.org/",
}

var TencentMapRoad = &GoogleMapProvider{
	TileMapMetadata: &TileMapMetadata{
		Name:           "Tencent Map Road 腾讯路网",
		ID:             "tencent_map_road",
		MinZoom:        0,
		MaxZoom:        18,
		MapSize:        MapSize256,
		MapType:        MapTypeRaster,
		ContentType:    MapContentTypePNG,
		CoordinateType: CoordinateTypeGCJ02,
	},
	// https://rt2.map.gtimg.com/tile?z=6&x=50&y=36&type=vector&styleid=3
	// https://rt0.map.gtimg.com/tile?z=11&x=1692&y=1207&type=vector&styleid=3
	BaseURL:      "https://rt{serverpart:0,1,2,3}.map.gtimg.com/tile?z={z}&x={x}&y={y}&styleid=3",
	ReferenceURL: "https://map.qq.com/",
}

// Tencent Map Satellite algorithm example:
// https://p1.map.gtimg.com/sateTiles/6/3/2/51_35.jpg
//
//	var satelliteTileLayer = new qq.maps.TileLayer({
//	  getTileUrl: function(coord, zoom) {
//	    return "http://p1.map.gtimg.com/sateTiles/"+zoom+"/"+Math.floor(coord.x/16)+"/"+Math.floor(coord.y/16)+"/"+coord.x+"_"+coord.y+".jpg";
//	  },
//	  tileSize: new qq.maps.Size(256, 256),
//	  name: "卫星图"
//	});
type TencentMapSatelliteProvider struct {
	*GoogleMapProvider
}

var TencentMapSatellite = &TencentMapSatelliteProvider{
	GoogleMapProvider: &GoogleMapProvider{
		TileMapMetadata: &TileMapMetadata{
			Name:           "Tencent Map Satellite 腾讯卫星影像",
			ID:             "tencent_map_satellite",
			MinZoom:        0,
			MaxZoom:        18,
			MapSize:        MapSize256,
			MapType:        MapTypeRaster,
			ContentType:    MapContentTypeJPEG,
			CoordinateType: CoordinateTypeGCJ02,
		},
		BaseURL:      "https://p{serverpart:0,1,2,3}.map.gtimg.com/sateTiles/{z}/{x/16}/{y/16}/{x}_{y}.jpg",
		ReferenceURL: "https://map.qq.com/",
	},
}

// override the TencentMapSatelite GetMapPic to match the tile url pattern
func (tmp *TencentMapSatelliteProvider) GetMapPic(x, y, z int) (*http.Response, error) {
	// check zoom level
	if z < tmp.MinZoom || z > tmp.MaxZoom {
		return nil, fmt.Errorf("map: %s zoom level %d is out of range [%d, %d]", tmp.Name, z, tmp.MinZoom, tmp.MaxZoom)
	}

	httpClient := request.DefaultHTTPClient
	mapUrl := tmp.BaseURL

	// y = int.Parse( Math.Pow(2, z).ToString()) - 1 - y;
	y = int(math.Pow(2, float64(z))) - 1 - y
	mapUrl = strings.Replace(mapUrl, "{y}", strconv.Itoa(y), 1) // Replace {y} with the actual y value

	mapUrl = strings.Replace(mapUrl, "{x}", strconv.Itoa(x), 1) // Replace {x} with the actual x value
	mapUrl = strings.Replace(mapUrl, "{y}", strconv.Itoa(y), 1) // Replace {y} with the actual y value
	mapUrl = strings.Replace(mapUrl, "{z}", strconv.Itoa(z), 1) // Replace {z} with the actual z value
	// replace {x/16} and {y/16}
	mapUrl = strings.Replace(mapUrl, "{x/16}", strconv.Itoa(x/16), 1)
	mapUrl = strings.Replace(mapUrl, "{y/16}", strconv.Itoa(y/16), 1)

	// replace `{serverpart:<item1>,<item2>,<item3>}` randomly with item1, item2, or item3
	mapUrl = tmp.ReplaceServerPart(mapUrl)

	logger.Debugf("Tencent Map Satellite tile URL: %s", mapUrl)

	// Make a GET request to the map URL
	request, err := http.NewRequest(http.MethodGet, mapUrl, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	if tmp.ReferenceURL != "" {
		request.Header.Set("Referer", tmp.ReferenceURL)
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

// tuxun.cn huawei street map (petal maps 华为花瓣地图)
// https://maprastertile-drcn.dbankcdn.cn/display-service/v1/online-render/getTile/25.06.13.20/8/209/110/?language=zh&p=46&scale=2&mapType=ROADMAP&presetStyleId=standard&pattern=JPG&key=DAEDANitav6P7Q0lWzCzKkLErbrJG4kS1u/CpEe5ZyxW5u0nSkb40bJ+YAugRN03fhf0BszLS1rCrzAogRHDZkxaMrloaHPQGO6LNg==
// https://maprastertile-drcn.dbankcdn.cn/display-service/v1/online-render/getTile/25.06.13.20/10/834/434/?language=zh&p=46&scale=2&mapType=ROADMAP&presetStyleId=standard&pattern=JPG&key=DAEDANitav6P7Q0lWzCzKkLErbrJG4kS1u/CpEe5ZyxW5u0nSkb40bJ+YAugRN03fhf0BszLS1rCrzAogRHDZkxaMrloaHPQGO6LNg==
// DAEDANitav6P7Q0lWzCzKkLErbrJG4kS1u%2FCpEe5ZyxW5u0nSkb40bJ%2BYAugRN03fhf0BszLS1rCrzAogRHDZkxaMrloaHPQGO6LNg==
var TuxunHuaweiStreetMap = &GoogleMapProvider{
	TileMapMetadata: &TileMapMetadata{
		Name:           "Tuxun Huawei Street Map 华为花瓣地图",
		ID:             "tuxun_huawei_street_map",
		MinZoom:        0,
		MaxZoom:        18,
		MapSize:        MapSize512,
		MapType:        MapTypeRaster,
		ContentType:    MapContentTypeJPEG,
		CoordinateType: CoordinateTypeGCJ02,
	},
	BaseURL:      "https://maprastertile-drcn.dbankcdn.cn/display-service/v1/online-render/getTile/25.06.13.20/{z}/{x}/{y}/?language=zh&p=46&scale=2&mapType=ROADMAP&presetStyleId=standard&pattern=JPG&key=DAEDANitav6P7Q0lWzCzKkLErbrJG4kS1u%2FCpEe5ZyxW5u0nSkb40bJ%2BYAugRN03fhf0BszLS1rCrzAogRHDZkxaMrloaHPQGO6LNg==",
	ReferenceURL: "https://tuxun.fun/",
}
