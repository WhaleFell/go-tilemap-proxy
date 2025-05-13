package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Cacher interface {
	// Get returns the value for the given key
	GetCache(key string) ([]byte, error)

	// Set sets the value for the given key
	SetCache(key string, value []byte) error

	// Generate cache key for the given string
	GenerateCacheKey(str string) (key string)
}

type MapCache struct {
	CachePath string
}

var (
	Cache         *MapCache
	InitCacheOnce sync.Once
)

func InitMapCache(cachePath string) *MapCache {
	InitCacheOnce.Do(func() {
		Cache = &MapCache{
			CachePath: cachePath,
		}
	})
	fmt.Printf("InitMapCache: cache path: %s\n", Cache.CachePath)
	return Cache
}

// Generate cache key for the given string MD5
func (mapcache *MapCache) GenerateCacheKey(str string) (key string) {

	// create a new MD5 hash
	hash := md5.Sum([]byte(str))

	// convert the hash to a string
	md5String := hex.EncodeToString(hash[:])

	return md5String
}

// get the cache path
// e.g. <CachePath>/6f/1e/d002ab5595859014ebf0951522d9
func (mapcache *MapCache) getCachePath(key string) (string, error) {
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
	cacheFilePath := filepath.Join(mapcache.CachePath, firstDir, secondDir, cacheFileName)

	return cacheFilePath, nil

}

// SetCache: save the data to the cache path, according to the hash value of the key
func (mapcache *MapCache) SetCache(keyStr string, value []byte) error {
	// key e.g. 6f1ed002ab5595859014ebf0951522d9
	key := mapcache.GenerateCacheKey(keyStr)

	// get the cache path
	cacheFilePath, err := mapcache.getCachePath(key)
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
func (mapcache *MapCache) GetCache(keyStr string) ([]byte, error) {
	// key e.g. 6f1ed002ab5595859014ebf0951522d9
	key := mapcache.GenerateCacheKey(keyStr)

	// get the cache path
	cacheFilePath, err := mapcache.getCachePath(key)
	if err != nil {
		return nil, fmt.Errorf("get cache path failed: %w", err)
	}

	// read the value from the cache file
	value, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return nil, fmt.Errorf("read cache file %s failed: %w", cacheFilePath, err)
	}

	return value, nil
}
