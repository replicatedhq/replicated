package main

import (
	"context"
	"dagger/replicated/internal/dagger"
)

// Build compiles the replicated CLI binary.
func (r *Replicated) Build(
	ctx context.Context,

	// +defaultPath="./"
	source *dagger.Directory,
) (*dagger.File, error) {
	goModCache := dag.CacheVolume("replicated-go-mod-126")
	goBuildCache := dag.CacheVolume("replicated-go-build-126")

	binary := dag.Container(dagger.ContainerOpts{
		Platform: "linux/amd64",
	}).
		From("golang:1.26").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithoutFile("/go/src/github.com/replicatedhq/replicated/bin/replicated").
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithMountedCache("/go/pkg/mod", goModCache).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithMountedCache("/go/build-cache", goBuildCache).
		WithEnvVariable("GOCACHE", "/go/build-cache").
		With(CacheBustingExec([]string{"make", "build"})).
		File("/go/src/github.com/replicatedhq/replicated/bin/replicated")

	return binary, nil
}
