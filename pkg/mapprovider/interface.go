package mapprovider

import (
	"net/http"
)

// map coordinate type enumeration
type MapCoordinateType string

const (
	// EPSG:3857 (Web Mercator) / Spherical Mercator
	CoordinateTypeWebMercator MapCoordinateType = "EPSG:3857"
	// EPSG:4326 (WGS 84)
	CoordinateTypeWGS84 MapCoordinateType = "EPSG:4326"
	// GCJ02 (国测局 2002 Coordinate System)
	CoordinateTypeGCJ02 MapCoordinateType = "GCJ02"
	// BD09 (Baidu Coordinate System)
	CoordinateTypeBD09 MapCoordinateType = "BD09"
	// CGCS2000 (China Geodetic Coordinate System 2000, approximate to WGS84)
	// CGCS2000 is the official coordinate system used in China, which is very close to WGS84.
	CoordinateTypeCGCS2000 MapCoordinateType = "CGCS2000"
)

// Map content type enumeration
type MapContentType string

const (
	MapContentTypePNG  MapContentType = "image/png"
	MapContentTypeJPEG MapContentType = "image/jpeg"
	MapContentTypeWebP MapContentType = "image/webp"
)

// Map type: vector or raster (矢量地图或栅格地图)
type MapType string

const (
	MapTypeVector MapType = "vector"
	MapTypeRaster MapType = "raster"
)

// MapSize enumeration
type MapSize int

const (
	MapSize256  MapSize = 256
	MapSize512  MapSize = 512
	MapSize1024 MapSize = 1024
)

type TileMapMetadata struct {
	Name           string            `json:"name"` // Name of the map provider
	ID             string            `json:"id"`   // Unique identifier for the map provider
	MinZoom        int               `json:"min_zoom"`
	MaxZoom        int               `json:"max_zoom"`
	MapType        MapType           `json:"map_type"`
	MapSize        MapSize           `json:"map_size"`
	CoordinateType MapCoordinateType `json:"coordinate_type"`
	ContentType    MapContentType    `json:"content_type"`
}

type TileMapProvider interface {
	// GetMapPic returns a tile map picture io.Reader
	GetMapPic(x, y, z int) (*http.Response, error)

	// GetMapMetadata returns the metadata of the map provider
	GetMapMetadata() *TileMapMetadata
}

func (metadata *TileMapMetadata) GetMetadataWithDefaults() *TileMapMetadata {

	// Name and ID should not be empty
	if metadata.Name == "" && metadata.ID == "" {
		panic("MapMetaData Name and ID cannot be empty")
	}

	if metadata.MapSize == 0 {
		metadata.MapSize = MapSize256
	}
	if metadata.MinZoom == 0 {
		metadata.MinZoom = 0
	}
	if metadata.MaxZoom == 0 {
		metadata.MaxZoom = 18
	}
	if metadata.ContentType == "" {
		metadata.ContentType = MapContentTypePNG
	}
	return metadata
}
