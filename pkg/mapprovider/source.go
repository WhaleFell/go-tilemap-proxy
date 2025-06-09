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

var MapSourceMapping = map[string]TileMapProvider{}

func init() {
	for _, provider := range MapSourceProviders {
		mapMetadata := provider.GetMapMetadata()
		MapSourceMapping[mapMetadata.ID] = provider
		log.Printf("Map ID: %s, Name:%s is registered \n", mapMetadata.ID, mapMetadata.Name)
	}
}
