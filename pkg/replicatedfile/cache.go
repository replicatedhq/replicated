package replicatedfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/usrbinapp/usrbin-go/pkg/updatechecker"
)

const (
	cacheTTLInMinutes = 15

	replicatedDir = ".replicated"
	cacheFile     = "cache.json"
)

type cache struct {
	CachedVersion     string                    `json:"cachedVersion,omitempty"`
	UpdateCheckerInfo *updatechecker.UpdateInfo `json:"updateCheckerInfo,omitempty"`
}

var (
	Cache *cache
)

func init() {
	var err error
	Cache, err = loadCache()
	if err != nil {
		fmt.Printf("Failed to load cache file: %v\n", err)
	}
}

func DeleteCacheFile() error {
	return os.Remove(cacheFilePath())
}

func (c *cache) SaveUpdateCheckerInfo(currentVersion string, updateCheckerInfo *updatechecker.UpdateInfo) error {
	if c == nil {
		c = &cache{}
	}
	c.CachedVersion = currentVersion
	c.UpdateCheckerInfo = updateCheckerInfo
	return c.save()
}

func (c *cache) IsUpdateCheckerInfoExpired(currentVersion string) bool {
	if c == nil || c.UpdateCheckerInfo == nil {
		return true
	}

	if c.UpdateCheckerInfo.CheckedAt.Add(time.Duration(cacheTTLInMinutes) * time.Minute).Before(time.Now()) {
		return true
	}

	if currentVersion != c.CachedVersion {
		return true
	}

	return false
}

func (u *cache) GetUpdateCheckerInfo() *updatechecker.UpdateInfo {
	if u == nil {
		return nil
	}

	return u.UpdateCheckerInfo
}

func (c *cache) save() error {
	cacheFile := cacheFilePath()
	if err := os.MkdirAll(path.Dir(cacheFile), 0755); err != nil {
		return err
	}

	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, b, 0600)
}

func loadCache() (*cache, error) {
	updateCacheFile := cacheFilePath()
	if _, err := os.Stat(updateCacheFile); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	b, err := os.ReadFile(updateCacheFile)
	if err != nil {
		return nil, err
	}

	updateCache := cache{}
	if err := json.Unmarshal(b, &updateCache); err != nil {
		return nil, err
	}

	return &updateCache, nil
}

func cacheFilePath() string {
	return filepath.Join(homeDir(), replicatedDir, cacheFile)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
