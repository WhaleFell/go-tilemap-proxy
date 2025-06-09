package tilemap

import (
	"fmt"
	"go-map-proxy/assets"
	"go-map-proxy/internal/config"
	"go-map-proxy/internal/model"
	"go-map-proxy/internal/utils"
	"go-map-proxy/pkg/logger"
	"go-map-proxy/pkg/mapprovider"
	"io"

	"github.com/labstack/echo/v4"
)

type TileMapPathParam struct {
	MapType string `param:"mapType" json:"mapType"`
	X       int    `param:"x" json:"x"`
	Y       int    `param:"y" json:"y"`
	Z       int    `param:"z" json:"z"`
}

// List tile map sources
func TileMapSourceList(c echo.Context) error {

	mapSourceList := make([]*mapprovider.TileMapMetadata, 0, len(mapprovider.MapSourceSlice))

	// use slice to prevent map source order
	for _, provider := range mapprovider.MapSourceSlice {
		mapSourceList = append(mapSourceList, provider.Value.GetMapMetadata().GetMetadataWithDefaults())
	}

	return c.JSON(200, model.BaseAPIResponse[[]*mapprovider.TileMapMetadata]{
		Code:    200,
		Message: "Get tile map source list success",
		Data:    mapSourceList,
	})
}

// TileMapProxy handles tile map requests
// It serves as a proxy for tile map services, allowing users to fetch tiles from various sources.
// It provides unified Google XYZ tile map protocol.
// format: /:mapID/:x/:y/:z?cache=<boolean>
// mapType: the type of map, e.g. "google", "osm", etc.
// x, y, z: the tile coordinates
// cache: whether to use exist tile cache, default is true
func TileMapHandler(c echo.Context) error {

	tileMapParam := new(TileMapPathParam)

	// Fluent Binding
	err := echo.PathParamsBinder(c).
		MustString("mapType", &tileMapParam.MapType).
		MustInt("x", &tileMapParam.X).
		MustInt("y", &tileMapParam.Y).
		MustInt("z", &tileMapParam.Z).
		BindError()
	if err != nil {
		return c.JSON(200, model.BaseAPIResponse[any]{
			Code:    400,
			Message: fmt.Sprintf("Invalid tile map parameters: %v", err),
			Data:    nil,
		})
	}

	// find map provider in `mapprovider.MapSourceIndex`
	provider, ok := mapprovider.MapSourceIndex[tileMapParam.MapType]
	if !ok {
		return c.JSON(200, model.BaseAPIResponse[any]{
			Code:    400,
			Message: fmt.Sprintf("Tile map source %s not found", tileMapParam.MapType),
			Data:    nil,
		})
	}
	providerMetadata := provider.GetMapMetadata().GetMetadataWithDefaults()

	// handle map cache
	isUseCache := true
	cacheParam := c.QueryParam("cache")
	if cacheParam == "false" {
		isUseCache = false
	}
	cacheKey := fmt.Sprintf("%s/%d/%d/%d", tileMapParam.MapType, tileMapParam.X, tileMapParam.Y, tileMapParam.Z)
	if isUseCache {
		// check if tile map picture is in cache
		if cacheData, err := utils.Cache.GetCache(cacheKey); err == nil {
			logger.Debugf("Tile map cache hit: %s", cacheKey)
			c.Response().Header().Set(echo.HeaderContentType, string(providerMetadata.ContentType))
			c.Response().Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", len(cacheData)))
			// set cache policy
			c.Response().Header().Set(echo.HeaderCacheControl, fmt.Sprintf("max-age=%d", config.Cfg.Cache.MaxAge))
			c.Response().Header().Set("X-cache", "HIT")
			c.Response().WriteHeader(200)
			_, err = c.Response().Writer.Write(cacheData)
			if err != nil {
				return c.JSON(200, model.BaseAPIResponse[any]{
					Code:    500,
					Message: fmt.Sprintf("Write tile map picture error: %v", err),
					Data:    nil,
				})
			}
			return nil
		}
		c.Response().Header().Set("X-cache", "MISS")
		logger.Debugf("Tile map cache miss: %s", cacheKey)
	}

	// get tile map picture response
	tileMapPicResponse, err := provider.GetMapPic(tileMapParam.X, tileMapParam.Y, tileMapParam.Z)
	if err != nil {
		logger.Errorf("Get tile map picture error: %v", err)
		return c.JSON(200, model.BaseAPIResponse[any]{
			Code:    500,
			Message: fmt.Sprintf("Get %s tile map picture error: %v", tileMapParam.MapType, err),
			Data:    nil,
		})
	}
	defer tileMapPicResponse.Body.Close()

	// set response header
	c.Response().Header().Set(echo.HeaderContentType, tileMapPicResponse.Header.Get("Content-Type"))
	c.Response().Header().Set(echo.HeaderContentLength, tileMapPicResponse.Header.Get("Content-Length"))

	// read tile map picture body
	picBytes, err := io.ReadAll(tileMapPicResponse.Body)
	if err != nil {
		logger.Errorf("Read tile map picture error: %v", err)
		return c.JSON(200, model.BaseAPIResponse[any]{
			Code:    500,
			Message: fmt.Sprintf("Read tile map picture error: %v", err),
			Data:    nil,
		})
	}
	// save tile map picture to cache (in new goroutine)
	go func() {
		err := utils.Cache.SetCache(cacheKey, picBytes)
		if err != nil {
			logger.Errorf("Set tile map cache error: %v", err)
		} else {
			logger.Debugf("Set tile map cache success: %s", cacheKey)
		}
	}()

	// if picBytes is empty, return failure picture
	if len(picBytes) == 0 {
		logger.Errorf("Tile map picture is empty")
		return c.Blob(200, "image/png", assets.TileMapFailedPng)
	}

	// set content length
	c.Response().Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", len(picBytes)))
	// write tile map picture to response
	c.Response().WriteHeader(200)
	_, err = c.Response().Writer.Write(picBytes)
	if err != nil {
		logger.Errorf("Write tile map picture error: %v", err)
		return c.JSON(200, model.BaseAPIResponse[any]{
			Code:    500,
			Message: fmt.Sprintf("Write tile map picture error: %v", err),
			Data:    nil,
		})
	}
	return nil
}
