package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// GetCacheDir returns the cache directory for tools
// Location: ~/.replicated/tools
func GetCacheDir() (string, error) {
	var home string

	// Get home directory
	if runtime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
	} else {
		home = os.Getenv("HOME")
	}

	if home == "" {
		return "", fmt.Errorf("HOME environment variable not set")
	}

	return filepath.Join(home, ".replicated", "tools"), nil
}

// GetToolPath returns the cached path for a specific tool version
// Example: ~/.replicated/tools/helm/3.14.4/darwin-arm64/helm
func GetToolPath(name, version string) (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}

	osArch := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	binaryName := name
	if runtime.GOOS == "windows" {
		binaryName = name + ".exe"
	}

	return filepath.Join(cacheDir, name, version, osArch, binaryName), nil
}

// IsCached checks if a tool version is already cached
func IsCached(name, version string) (bool, error) {
	toolPath, err := GetToolPath(name, version)
	if err != nil {
		return false, err
	}

	info, err := os.Stat(toolPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	// Make sure it's a file, not a directory
	if info.IsDir() {
		return false, nil
	}

	return true, nil
}
