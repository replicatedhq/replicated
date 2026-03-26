package main

import (
	"context"
	"dagger/replicated/internal/dagger"
	"fmt"
	"strings"
	"time"
)

// goVersion reads the go directive from go.mod in the source directory and
// returns the major.minor version (e.g. "1.26" from "go 1.26.1").
func goVersion(ctx context.Context, source *dagger.Directory) (string, error) {
	contents, err := source.File("go.mod").Contents(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod: %w", err)
	}
	for _, line := range strings.Split(contents, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "go ") {
			version := strings.TrimPrefix(line, "go ")
			// strip patch version if present (e.g. "1.26.1" -> "1.26")
			parts := strings.SplitN(version, ".", 3)
			if len(parts) >= 2 {
				return parts[0] + "." + parts[1], nil
			}
			return version, nil
		}
	}
	return "", fmt.Errorf("go directive not found in go.mod")
}

// goImage returns the golang Docker image tag for the version in go.mod.
func goImage(ctx context.Context, source *dagger.Directory) (string, error) {
	v, err := goVersion(ctx, source)
	if err != nil {
		return "", err
	}
	return "golang:" + v, nil
}

// goCacheVolumes returns the mod and build cache volumes keyed by the Go version.
func goCacheVolumes(ctx context.Context, source *dagger.Directory) (*dagger.CacheVolume, *dagger.CacheVolume, error) {
	v, err := goVersion(ctx, source)
	if err != nil {
		return nil, nil, err
	}
	// replace dots for a clean cache key suffix (e.g. "126" from "1.26")
	suffix := strings.ReplaceAll(v, ".", "")
	modCache := dag.CacheVolume("replicated-go-mod-" + suffix)
	buildCache := dag.CacheVolume("replicated-go-build-" + suffix)
	return modCache, buildCache, nil
}

// CacheBustingExec is a helper function that will add a cache busting env var automatically
// to the container. This is useful when Exec target is a dynamic event acting on an entity outside
// of the container that you absolutely want to re-run every time.
//
// Temporary hack until cache controls are a thing: https://docs.dagger.io/cookbook/#invalidate-cache
func CacheBustingExec(args []string, opts ...dagger.ContainerWithExecOpts) dagger.WithContainerFunc {
	return func(c *dagger.Container) *dagger.Container {
		if c == nil {
			panic("CacheBustingExec requires a container, but container was nil")
		}
		return c.WithEnvVariable("DAGGER_CACHEBUSTER_CBE", time.Now().String()).WithExec(args, opts...)
	}
}
