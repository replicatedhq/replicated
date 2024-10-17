package main

import (
	"context"
	"dagger/replicated/internal/dagger"
)

func validateSecurity(
	ctx context.Context,

	// +defaultPath="./"
	source *dagger.Directory,
) error {
	goModCache := dag.CacheVolume("replicated-go-mod-122")
	goBuildCache := dag.CacheVolume("replicated-go-build-121")

	// run semgrep
	semgrep := dag.Container().
		From("returntocorp/semgrep").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithMountedCache("/go/pkg/mod", goModCache).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithMountedCache("/go/build-cache", goBuildCache).
		WithEnvVariable("GOCACHE", "/go/build-cache").
		With(CacheBustingExec([]string{"semgrep", "scan", "--config=p/golang", "."}))

	_, err := semgrep.Stderr(ctx)
	if err != nil {
		return err
	}

	return nil
}
