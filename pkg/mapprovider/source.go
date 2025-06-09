package mapprovider

import "log"

// map source

var MapSourceProviders = []TileMapProvider{
	GmapPureSatellite,
	GmapPureSatellite2,
	OpenStreetMapStandard,
	OpenStreetMapPublicGPS,
	TraceStrackTopoMap,
	ArcgisSatelite,
	BingSateliteMap,
	GoogleHybridOffsetMap,
	OpenStreetMapCyclOSM,
	AmapRoadMap,
	TianDiTuSatellite,
	TianDiTuRoad,
	MapHereSatelite,
	MapTilerContour,
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
