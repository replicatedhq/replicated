package version

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

// debugMode controls whether debug messages are printed
var debugMode bool

// SetDebugMode enables or disables debug mode
func SetDebugMode(debug bool) {
	debugMode = debug
}

// debugLog logs messages when debug mode is enabled
func debugLog(format string, args ...interface{}) {
	if debugMode {
		fmt.Fprintf(os.Stderr, "[DEBUG] Version: "+format+"\n", args...)
	}
}

// UpdateCache represents the cached update information
type UpdateCache struct {
	Version       string     `json:"cachedVersion"`
	UpdateInfo    UpdateInfo `json:"updateCheckerInfo"`
	LastCheckedAt time.Time  `json:"lastCheckedAt"`
}

// getCacheDir returns the directory where the cache file is stored
func getCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get user home directory")
	}

	return filepath.Join(homeDir, ".replicated"), nil
}

// getCachePath returns the path to the update cache file
func getCachePath() (string, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get cache directory")
	}

	return filepath.Join(cacheDir, "cache.json"), nil
}

// LoadUpdateCache loads the update information from the cache file
// Returns nil without error if the cache doesn't exist or is invalid
func LoadUpdateCache() (*UpdateCache, error) {
	cachePath, err := getCachePath()
	if err != nil {
		debugLog("Failed to get cache path: %v", err)
		return nil, nil // Silently return nil
	}

	debugLog("Loading update cache from %s", cachePath)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			debugLog("Cache file doesn't exist")
			return nil, nil // Cache doesn't exist yet, not an error
		}
		debugLog("Error reading cache file: %v", err)
		return nil, nil // Silently return nil on any error
	}

	var cache UpdateCache
	err = json.Unmarshal(data, &cache)
	if err != nil {
		debugLog("Invalid cache format: %v", err)
		return nil, nil // Invalid cache, silently return nil
	}

	debugLog("Successfully loaded cache: version=%s, latest=%s, checked=%s",
		cache.Version, cache.UpdateInfo.LatestVersion, cache.LastCheckedAt.Format(time.RFC3339))
	return &cache, nil
}

// ClearUpdateCache removes the update cache file
// Silently fails - no errors are returned
func ClearUpdateCache() {
	cachePath, err := getCachePath()
	if err != nil {
		debugLog("Failed to get cache path: %v", err)
		return
	}

	// Check if file exists first
	_, err = os.Stat(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			debugLog("Cache file doesn't exist, nothing to clear")
			return
		}
		debugLog("Error checking cache file: %v", err)
		return
	}

	debugLog("Clearing update cache at %s", cachePath)
	err = os.Remove(cachePath)
	if err != nil {
		debugLog("Failed to remove cache file: %v", err)
	} else {
		debugLog("Successfully cleared update cache")
	}
}

// SaveUpdateCache saves the update information to the cache file
// Silently fails - no errors are returned
func SaveUpdateCache(currentVersion string, updateInfo *UpdateInfo) {
	if updateInfo == nil {
		debugLog("Not saving cache: updateInfo is nil")
		return
	}

	debugLog("Saving update cache for version %s, latest version %s",
		currentVersion, updateInfo.LatestVersion)

	cacheDir, err := getCacheDir()
	if err != nil {
		debugLog("Failed to get cache directory: %v", err)
		return // Silently fail
	}

	// Ensure directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		debugLog("Failed to create cache directory: %v", err)
		return // Silently fail
	}

	cachePath, err := getCachePath()
	if err != nil {
		debugLog("Failed to get cache path: %v", err)
		return // Silently fail
	}

	cache := UpdateCache{
		Version:       currentVersion,
		UpdateInfo:    *updateInfo,
		LastCheckedAt: time.Now(),
	}

	data, err := json.Marshal(cache)
	if err != nil {
		debugLog("Failed to marshal cache data: %v", err)
		return // Silently fail
	}

	// Write to temp file first to avoid corruption
	tmpFile := cachePath + ".tmp"
	debugLog("Writing to temporary file %s", tmpFile)
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		debugLog("Failed to write temporary file: %v", err)
		return // Silently fail
	}

	// Rename to the actual file (atomic on POSIX systems)
	debugLog("Renaming %s to %s", tmpFile, cachePath)
	if err := os.Rename(tmpFile, cachePath); err != nil {
		debugLog("Failed to rename temporary file: %v", err)
		os.Remove(tmpFile) // Clean up temp file if rename fails
		return             // Silently fail
	}

	debugLog("Successfully saved update cache")
}

// CheckForUpdatesInBackground starts a goroutine to check for updates
// and save the results to the cache file. It never blocks and silently
// handles all errors.
func CheckForUpdatesInBackground(currentVersion string, homebrewFormula string) {
	debugLog("Starting background update check for version %s", currentVersion)

	go func() {
		// Create update checker with shorter timeout for background checks
		debugLog("Creating update checker with formula %s", homebrewFormula)
		updateChecker, err := NewUpdateChecker(currentVersion, homebrewFormula)
		if err != nil {
			debugLog("Failed to create update checker: %v", err)
			return // Silently fail
		}

		// Use a shorter timeout for background checks
		updateChecker.httpTimeout = 1 * time.Second

		// Get update info with a longer timeout for background check
		debugLog("Checking for updates in background")
		updateInfo, err := updateChecker.GetUpdateInfo()
		if err != nil {
			debugLog("Failed to get update info: %v", err)
			return // Silently fail
		}

		if updateInfo == nil {
			debugLog("No update info returned (nil)")
			return
		}

		debugLog("Update check complete, latest version: %s", updateInfo.LatestVersion)

		// Save to cache
		SaveUpdateCache(currentVersion, updateInfo)
	}()
}
