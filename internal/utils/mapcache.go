package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"go-map-proxy/pkg/logger"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Cacher interface {
	// Get returns the value for the given key
	GetCache(key string) ([]byte, error)

	// Set sets the value for the given key
	SetCache(key string, value []byte) error

	// Generate cache key for the given string
	// GenerateCacheKey(str string) (key string)
}

// map cache base on key hash path (like nginx cache strategy)
// e.g.: <CachePath>/6f/1e/d002ab5595859014ebf0951522d9
type HashMapCache struct {
	CachePath string
}

var (
	Cache         Cacher
	InitCacheOnce sync.Once
)

func NewHashMapCache(cachePath string) *HashMapCache {
	// check if the global Cache is already initialized
	if Cache != nil {
		panic("Cacher is already initialized")
	}

	InitCacheOnce.Do(func() {
		Cache = &HashMapCache{
			CachePath: cachePath,
		}
	})

	// type assert
	fmt.Printf("InitHashMapCache: cache path: %s\n", Cache.(*HashMapCache).CachePath)

	return Cache.(*HashMapCache)
}

// Generate cache key for the given string MD5
func (hashmapcache *HashMapCache) GenerateCacheKey(str string) (key string) {

	// create a new MD5 hash
	hash := md5.Sum([]byte(str))

	// convert the hash to a string
	md5String := hex.EncodeToString(hash[:])

	return md5String
}

// get the cache path
// e.g. <CachePath>/6f/1e/d002ab5595859014ebf0951522d9
func (hashmapcache *HashMapCache) getCachePath(key string) (string, error) {
	// key is hash md5 value
	// e.g. 6f1ed002ab5595859014ebf0951522d9

	// check if the key is valid
	if len(key) != 32 {
		return "", fmt.Errorf("invalid key length: %d, expected 32", len(key))
	}

	// get the cache path
	// e.g <CachePath>/6f/1e/d002ab5595859014ebf0951522d9
	firstDir := key[0:2]
	secondDir := key[2:4]
	cacheFileName := key[4:]
	cacheFilePath := filepath.Join(hashmapcache.CachePath, firstDir, secondDir, cacheFileName)

	return cacheFilePath, nil

}

// SetCache: save the data to the cache path, according to the hash value of the key
func (hashmapcache *HashMapCache) SetCache(keyStr string, value []byte) error {
	// check if the value is empty
	if len(value) == 0 {
		return fmt.Errorf("value is empty, nothing to cache")
	}

	// key e.g. 6f1ed002ab5595859014ebf0951522d9
	key := hashmapcache.GenerateCacheKey(keyStr)

	// get the cache path
	cacheFilePath, err := hashmapcache.getCachePath(key)
	if err != nil {
		return fmt.Errorf("get cache path failed: %w", err)
	}

	// create the cache path
	if err := os.MkdirAll(filepath.Dir(cacheFilePath), os.ModePerm); err != nil {
		return fmt.Errorf("create cache path %s failed: %w", cacheFilePath, err)
	}

	// write the value to the cache file
	if err := os.WriteFile(cacheFilePath, value, os.ModePerm); err != nil {
		return fmt.Errorf("write cache file %s failed: %w", cacheFilePath, err)
	}

	return nil

}

// GetCache: get the data from the cache path, according to the hash value of the key
func (hashmapcache *HashMapCache) GetCache(keyStr string) ([]byte, error) {
	// key e.g. 6f1ed002ab5595859014ebf0951522d9
	key := hashmapcache.GenerateCacheKey(keyStr)

	// get the cache path
	cacheFilePath, err := hashmapcache.getCachePath(key)
	if err != nil {
		return nil, fmt.Errorf("get cache path failed: %w", err)
	}

	// read the value from the cache file
	value, err := os.ReadFile(cacheFilePath)
	if err != nil {
		logger.Debugf("read cache file %s failed: %v", cacheFilePath, err)
		return nil, fmt.Errorf("read cache file %s failed: %w", cacheFilePath, err)
	}

	// check if the value is empty
	if len(value) == 0 {
		return nil, fmt.Errorf("cache file %s is empty", cacheFilePath)
	}

	return value, nil
}

// map cache base on clean path
// <mapType>/<z>/<x>/<y>/<mapCacheFileExtension|.png|.webp|.jpeg>
// e.g. <CachePath>/googlemap/6/10/20.png
type PathMapCache struct {
	CachePath string
}

func NewPathMapCache(cachePath string) *PathMapCache {
	// check if the global Cache is already initialized
	if Cache != nil {
		panic("Cacher is already initialized")
	}

	InitCacheOnce.Do(func() {
		Cache = &PathMapCache{
			CachePath: cachePath,
		}
	})

	// type assert
	fmt.Printf("InitPathMapCache: cache path: %s\n", Cache.(*PathMapCache).CachePath)

	return Cache.(*PathMapCache)
}

// parse the key string to the map type, z, x, y and extension and combine them to a cache path
func (pathmapcache *PathMapCache) getCachePath(keyStr string) (string, error) {
	// keyStr is like "googlemap/6/10/20.png"
	parts := strings.Split(keyStr, "/")
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid key string: %s, expected format: <mapType>/<z>/<x>/<y>.<extension>", keyStr)
	}

	// join the parts to a cache path
	cacheFilePath := filepath.Join(pathmapcache.CachePath, parts[0], parts[1], parts[2], parts[3])

	return cacheFilePath, nil
}

func (pathmapcache *PathMapCache) SetCache(keyStr string, value []byte) error {
	// check if the value is empty
	if len(value) == 0 {
		return fmt.Errorf("value is empty, nothing to cache")
	}

	// get the cache path
	cacheFilePath, err := pathmapcache.getCachePath(keyStr)
	if err != nil {
		return fmt.Errorf("get cache path failed: %w", err)
	}

	// create the cache path
	if err := os.MkdirAll(filepath.Dir(cacheFilePath), os.ModePerm); err != nil {
		return fmt.Errorf("create cache path %s failed: %w", cacheFilePath, err)
	}

	// write the value to the cache file
	if err := os.WriteFile(cacheFilePath, value, os.ModePerm); err != nil {
		return fmt.Errorf("write cache file %s failed: %w", cacheFilePath, err)
	}

	return nil
}

func (pathmapcache *PathMapCache) GetCache(keyStr string) ([]byte, error) {
	// get the cache path
	cacheFilePath, err := pathmapcache.getCachePath(keyStr)
	if err != nil {
		return nil, fmt.Errorf("get cache path failed: %w", err)
	}

	// read the value from the cache file
	value, err := os.ReadFile(cacheFilePath)
	if err != nil {
		logger.Debugf("read cache file %s failed: %v", cacheFilePath, err)
		return nil, fmt.Errorf("read cache file %s failed: %w", cacheFilePath, err)
	}

	// check if the value is empty
	if len(value) == 0 {
		return nil, fmt.Errorf("cache file %s is empty", cacheFilePath)
	}

	return value, nil
}
