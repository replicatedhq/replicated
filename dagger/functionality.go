package main

import (
	"context"
	"dagger/replicated/internal/dagger"
)

func validateFunctionality(
	ctx context.Context,

	// +defaultPath="./"
	source *dagger.Directory,
) error {
	goModCache := dag.CacheVolume("replicated-go-mod-122")
	goBuildCache := dag.CacheVolume("replicated-go-build-121")

	// unit tests
	unitTest := dag.Container().
		From("golang:1.24").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithMountedCache("/go/pkg/mod", goModCache).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithMountedCache("/go/build-cache", goBuildCache).
		WithEnvVariable("GOCACHE", "/go/build-cache").
		With(CacheBustingExec([]string{"make", "test-unit"}))

	_, err := unitTest.Stderr(ctx)
	if err != nil {
		return err
	}

	return nil
}
