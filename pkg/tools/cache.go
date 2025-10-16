package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// GetCacheDir returns the platform-appropriate cache directory for tools
//   macOS: ~/Library/Caches/replicated/tools
//   Linux: ~/.cache/replicated/tools (or $XDG_CACHE_HOME/replicated/tools)
//   Windows: %LOCALAPPDATA%\replicated\tools
func GetCacheDir() (string, error) {
	var cacheBase string

	switch runtime.GOOS {
	case "darwin":
		cacheBase = os.Getenv("HOME")
		if cacheBase == "" {
			return "", fmt.Errorf("HOME environment variable not set")
		}
		return filepath.Join(cacheBase, "Library", "Caches", "replicated", "tools"), nil

	case "linux":
		cacheBase = os.Getenv("XDG_CACHE_HOME")
		if cacheBase == "" {
			home := os.Getenv("HOME")
			if home == "" {
				return "", fmt.Errorf("neither XDG_CACHE_HOME nor HOME environment variables are set")
			}
			cacheBase = filepath.Join(home, ".cache")
		}
		return filepath.Join(cacheBase, "replicated", "tools"), nil

	case "windows":
		cacheBase = os.Getenv("LOCALAPPDATA")
		if cacheBase == "" {
			cacheBase = os.Getenv("USERPROFILE")
		}
		if cacheBase == "" {
			return "", fmt.Errorf("neither LOCALAPPDATA nor USERPROFILE environment variables are set")
		}
		return filepath.Join(cacheBase, "replicated", "tools"), nil

	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// GetToolPath returns the cached path for a specific tool version
// Example: ~/.cache/replicated/tools/helm/3.14.4/darwin-arm64/helm
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
