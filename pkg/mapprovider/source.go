package mapprovider

import "log"

// map source

var MapSourceProviders = []TileMapProvider{

	// Google satellite (WGJ84)
	GmapSatellite,
	// 中国路网乱
	GmapSatelliteWithLable,
	// GCJ02 offset version
	GmapSatelliteGCJ02,
	GmapSatelliteGCJ02WithLable,

	// Openstreetmap (WGJ84)
	OpenStreetMapStandard,
	OpenStreetMapPublicGPS,
	OpenStreetMapCyclOSM,
	TraceStrackTopoMap,
	OpenRailwayMap,

	// ArcGIS satellite (WGJ84)
	ArcgisSatelite,

	// BingSatelite (WGJ84)
	BingSateliteMap,

	// Amap Road map 高德地图 (WGJ84, coordinate calibration)
	AmapRoadMap,

	// TianDiTu 天地图 (CGCS2000, approximate to WGS84)
	TianDiTuSatellite,
	TianDiTuRoad,

	// Tencent 腾讯地图 (GCJ02)
	TencentMapRoad,
	TencentMapSatellite,

	// Huawei 华为花瓣地图 (图寻接口)
	TuxunHuaweiStreetMap,

	// lose efficity 失效
	// MapHereSatelite,
	// MapTilerContour,
}

// use struct+slice to prevent the order of map source.
// because map is unordered in golang.
type MapSourceMappingKV struct {
	Key   string
	Value TileMapProvider
}

var (
	MapSourceSlice []MapSourceMappingKV               // 保序 prevent order of map source
	MapSourceIndex = make(map[string]TileMapProvider) // 快速查找 by ID for quick lookup by ID
)

func init() {
	for _, provider := range MapSourceProviders {
		mapMetadata := provider.GetMapMetadata()
		MapSourceSlice = append(MapSourceSlice, MapSourceMappingKV{
			Key:   mapMetadata.ID,
			Value: provider,
		})
		MapSourceIndex[mapMetadata.ID] = provider

		log.Printf("Map ID: %s, Name: %s is registered\n", mapMetadata.ID, mapMetadata.Name)
	}
}
