package mapprovider

import (
	"net/http"
)

type TileMapProvider interface {
	// GetMapPic returns a tile map picture io.Reader
	GetMapPic(x, y, z int) (*http.Response, error)

	// GetMapName returns the name of the map provider
	GetMapName() string
}
