package mapprovider

// map source

var MapSourceList = map[string]TileMapProvider{
	"gps":  GmapPureSatellite,
	"gps2": GmapPureSatellite2,
}
