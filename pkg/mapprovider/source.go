package mapprovider

// map source

var MapSourceList = map[string]TileMapProvider{
	"gps":                GmapPureSatellite,
	"gps2":               GmapPureSatellite2,
	"osms":               OpenStreetMapStandard,
	"osmg":               OpenStreetMapPublicGPS,
	"topomap":            TraceStrackTopoMap,
	"arcgisSatellite":    ArcgisSatelite,
	"bingSatellite":      BingSateliteMap,
	"googleHybridOffset": GoogleHybridOffsetMap,
	"cyclosm":            OpenStreetMapCyclOSM,
	"amapRoad":           AmapRoadMap,
	"tiandituSatellite":  TianDiTuSatellite,
}
