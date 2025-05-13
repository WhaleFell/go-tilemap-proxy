package tilemap

import (
	"fmt"
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

	type MapSource struct {
		MapType string `json:"map_type"`
		MapName string `json:"map_name"`
	}

	mapSourceList := []MapSource{}

	for key, provider := range mapprovider.MapSourceList {
		mapSourceList = append(mapSourceList, MapSource{
			MapType: key,
			MapName: provider.GetMapName(),
		})
		fmt.Printf("Map source: %s\n", provider.GetMapName())
	}

	return c.JSON(200, model.BaseAPIResponse[[]MapSource]{
		Code:    200,
		Message: "Get tile map source list success",
		Data:    mapSourceList,
	})
}

// TileMapProxy handles tile map requests
// It serves as a proxy for tile map services, allowing users to fetch tiles from various sources.
// It provides unified Google XYZ tile map protocol.
// format: /:mapType/:x/:y/:z?cache=<boolean>
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
			// if cache data is not empty, return it
			if len(cacheData) > 0 {
				logger.Debugf("Tile map cache hit: %s", cacheKey)
				c.Response().Header().Set(echo.HeaderContentType, "image/png")
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
		}
		c.Response().Header().Set("X-cache", "MISS")
		logger.Debugf("Tile map cache miss: %s", cacheKey)
	}

	// find map provider in map source list
	provider, ok := mapprovider.MapSourceList[tileMapParam.MapType]
	if !ok {
		return c.JSON(200, model.BaseAPIResponse[any]{
			Code:    400,
			Message: fmt.Sprintf("Tile map source %s not found", tileMapParam.MapType),
			Data:    nil,
		})
	}

	// get tile map picture response
	tileMapPicResponse, err := provider.GetMapPic(tileMapParam.X, tileMapParam.Y, tileMapParam.Z)
	if err != nil {
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

	// set content length
	c.Response().Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", len(picBytes)))
	// write tile map picture to response
	c.Response().WriteHeader(200)
	_, err = c.Response().Writer.Write(picBytes)
	if err != nil {
		return c.JSON(200, model.BaseAPIResponse[any]{
			Code:    500,
			Message: fmt.Sprintf("Write tile map picture error: %v", err),
			Data:    nil,
		})
	}
	return nil
}
