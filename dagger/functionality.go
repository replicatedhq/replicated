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
	image, err := goImage(ctx, source)
	if err != nil {
		return err
	}
	goModCache, goBuildCache, err := goCacheVolumes(ctx, source)
	if err != nil {
		return err
	}

	// unit tests
	unitTest := dag.Container().
		From(image).
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithMountedCache("/go/pkg/mod", goModCache).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithMountedCache("/go/build-cache", goBuildCache).
		WithEnvVariable("GOCACHE", "/go/build-cache").
		With(CacheBustingExec([]string{"make", "test-unit"}))

	_, err = unitTest.Stderr(ctx)
	if err != nil {
		return err
	}

	return nil
}
