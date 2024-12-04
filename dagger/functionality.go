package main

import (
	"context"
	"dagger/replicated/internal/dagger"
)

func validateFunctionality(
	ctx context.Context,

	// +defaultPath="./"
	source *dagger.Directory,
) (bool, map[string]Logs, error) {
	goModCache := dag.CacheVolume("replicated-go-mod-122")
	goBuildCache := dag.CacheVolume("replicated-go-build-121")

	checkLogs := map[string]Logs{}

	// unit tests
	unitTest := dag.Container().
		From("golang:1.22").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithMountedCache("/go/pkg/mod", goModCache).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithMountedCache("/go/build-cache", goBuildCache).
		WithEnvVariable("GOCACHE", "/go/build-cache").
		With(CacheBustingExec([]string{"make", "test-unit"}))

	unitTestStdout, err := unitTest.Stdout(ctx)
	if err != nil {
		return false, nil, err
	}
	unitTestStderr, err := unitTest.Stderr(ctx)
	if err != nil {
		return false, nil, err
	}
	checkLogs["unit-tests"] = Logs{
		Stdout: unitTestStdout,
		Stderr: unitTestStderr,
	}

	return true, checkLogs, nil
}
