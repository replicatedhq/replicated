package main

import (
	"context"
	"dagger/replicated/internal/dagger"
	"fmt"
	"strings"
)

func (r *Replicated) Docs(
	ctx context.Context,

	// +defaultPath="./"
	source *dagger.Directory,

	githubToken *dagger.Secret,
) (string, error) {
	docsClone := dag.Container().
		From("alpine/git:latest").
		WithWorkdir("/").
		With(CacheBustingExec([]string{"git", "clone", "--depth", "1", "https://github.com/replicatedhq/replicated-docs.git", "/replicated-docs"}))
	_, err := docsClone.Stdout(ctx)
	if err != nil {
		return "", err
	}

	docsDirectory := docsClone.Directory("/replicated-docs")

	goModCache := dag.CacheVolume("replicated-go-mod-122")
	goBuildCache := dag.CacheVolume("replicated-go-build-121")

	// generate the docs from this current commit
	docs := dag.Container().
		From("golang:1.22").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithMountedCache("/go/pkg/mod", goModCache).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithMountedCache("/go/build-cache", goBuildCache).
		WithEnvVariable("GOCACHE", "/go/build-cache").
		WithEnvVariable("CGO_ENABLED", "0").
		With(CacheBustingExec([]string{"go", "run", "./docs/"}))
	_, err = docs.Stdout(ctx)
	if err != nil {
		return "", err
	}

	generatedDocs := docs.Directory("/go/src/github.com/replicatedhq/replicated/gen/docs")

	// remove all previous cli docs
	entries, err := docsDirectory.Directory("/docs/reference").Entries(ctx)
	if err != nil {
		return "", err
	}
	entriesToDelete := []string{}
	for _, entry := range entries {
		if strings.HasPrefix(entry, "replicated-cli") {
			entriesToDelete = append(entriesToDelete, fmt.Sprintf("/docs/reference/%s", entry))
		}
	}
	docsDirectory = docsDirectory.WithoutFiles(entriesToDelete)

	// copy the generated docs to the replicatedhq/replicated-docs repo
	generatedEntries, err := generatedDocs.Entries(ctx)
	if err != nil {
		return "", err
	}

	for _, generatedDoc := range generatedEntries {
		content, err := generatedDocs.File(generatedDoc).Contents(ctx)
		if err != nil {
			return "", err
		}

		// the cobra generate docs creates files with names like "replicated_app_ls.md"
		// our docs want files with names like "replicated-cli-app-ls.mdx"
		updatedFilename := strings.ReplaceAll(generatedDoc, ".md", ".mdx")
		updatedFilename = strings.ReplaceAll(updatedFilename, "replicated_", "replicated-cli-")
		updatedFilename = strings.ReplaceAll(updatedFilename, "_", "-")

		docsDirectory = docsDirectory.WithNewFile(fmt.Sprintf("/docs/reference/%s", updatedFilename), content)
	}

	// create a debug container to look
	debug := dag.Container().
		From("ubuntu:latest").
		WithMountedDirectory("/replicated-docs", docsDirectory).Terminal()
	_, err = debug.Stdout(ctx)
	if err != nil {
		return "", err
	}

	return "", nil

}
