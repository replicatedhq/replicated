package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCacheDir_LegacyPath(t *testing.T) {
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)

	legacyDir := filepath.Join(tempHome, ".replicated", "cache")
	require.NoError(t, os.MkdirAll(legacyDir, 0755))

	cacheDir, err := getCacheDir()
	require.NoError(t, err)
	assert.Equal(t, legacyDir, cacheDir)
}

func TestGetCacheDir_XDGPath(t *testing.T) {
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	originalXDGCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("XDG_CACHE_HOME", originalXDGCacheHome)
		xdg.Reload()
	}()

	os.Setenv("HOME", tempHome)
	os.Setenv("XDG_CACHE_HOME", filepath.Join(tempHome, ".cache"))
	xdg.Reload()

	cacheDir, err := getCacheDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(xdg.CacheHome, "replicated"), cacheDir)
	assert.NotEqual(t, filepath.Join(tempHome, ".replicated", "cache"), cacheDir)
}

func TestGetCacheDir_XDGPathWithoutHome(t *testing.T) {
	tempCache := t.TempDir()
	originalHome := os.Getenv("HOME")
	originalXDGCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("XDG_CACHE_HOME", originalXDGCacheHome)
		xdg.Reload()
	}()

	os.Unsetenv("HOME")
	os.Setenv("XDG_CACHE_HOME", tempCache)
	xdg.Reload()

	cacheDir, err := getCacheDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tempCache, "replicated"), cacheDir)
}

func TestInitCache_XDGPath(t *testing.T) {
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	originalXDGCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("XDG_CACHE_HOME", originalXDGCacheHome)
		xdg.Reload()
	}()

	os.Setenv("HOME", tempHome)
	os.Setenv("XDG_CACHE_HOME", filepath.Join(tempHome, ".cache"))
	xdg.Reload()

	_, err := InitCache()
	require.NoError(t, err)

	expectedPath := filepath.Join(xdg.CacheHome, "replicated", cacheFileName)
	_, err = os.Stat(expectedPath)
	assert.NoError(t, err)

	legacyPath := filepath.Join(tempHome, ".replicated", "cache", cacheFileName)
	_, err = os.Stat(legacyPath)
	assert.True(t, os.IsNotExist(err), "legacy cache file should not exist")
}
