package main

import (
	"context"
	"dagger/replicated/internal/dagger"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Masterminds/semver"
)

var goreleaserVersion = "v2.3.2"

func (r *Replicated) Release(
	ctx context.Context,

	// +defaultPath="./"
	source *dagger.Directory,

	version string,

	// +default=false
	snapshot bool,

	// +default=false
	clean bool,

	onePasswordServiceAccount *dagger.Secret,
) error {
	gitTreeOK, err := checkGitTree(ctx, source)
	if err != nil {
		return err
	}
	if !gitTreeOK {
		return fmt.Errorf("git tree is not clean")
	}

	major, minor, patch, err := parseVersion(ctx, version)
	if err != nil {
		return err
	}

	_ = dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithExec([]string{"git", "tag", fmt.Sprintf("v%d.%d.%d", major, minor, patch)}).
		WithExec([]string{"git", "push", "origin", fmt.Sprintf("v%d.%d.%d", major, minor, patch)})

	replicatedBinary := dag.Container().
		From("golang:1.22").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithExec([]string{"make", "build"}).
		File("/go/src/github.com/replicatedhq/replicated/bin/replicated")

	dockerContainer := dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "--no-cache", "ca-certificates", "curl", "git", "nodejs", "npm"}).
		WithExec([]string{"update-ca-certificates"}).
		WithExec([]string{"npm", "install", "-g", "replicated-lint"}).
		WithEnvVariable("IN_CONTAINER", "1").
		WithLabel("com.replicated.vendor_cli", "true").
		WithWorkdir("/out").
		WithEntrypoint([]string{"/replicated"}).
		WithFile("/replicated", replicatedBinary)

	username, err := dag.Onepassword().FindSecret(
		onePasswordServiceAccount,
		"Developer Automation",
		"Docker Hub Release Account",
		"username",
	).Plaintext(ctx)
	if err != nil {
		panic(err)
	}
	password := dag.Onepassword().FindSecret(
		onePasswordServiceAccount,
		"Developer Automation",
		"Docker Hub Release Account",
		"password",
	)

	dockerContainer = dockerContainer.WithRegistryAuth("", username, password)
	if _, err := dockerContainer.Publish(ctx, "replicated/vendor-cli:latest"); err != nil {
		panic(err)
	}
	if _, err := dockerContainer.Publish(ctx, fmt.Sprintf("replicated/vendor-cli:%d", major)); err != nil {
		panic(err)
	}
	if _, err := dockerContainer.Publish(ctx, fmt.Sprintf("replicated/vendor-cli:%d.%d", major, minor)); err != nil {
		panic(err)
	}
	if _, err := dockerContainer.Publish(ctx, fmt.Sprintf("replicated/vendor-cli:%d.%d.%d", major, minor, patch)); err != nil {
		panic(err)
	}

	if snapshot {
		_, err := dag.
			Goreleaser(dagger.GoreleaserOpts{
				Version: goreleaserVersion,
			}).
			WithSource(source).
			Snapshot(ctx, dagger.GoreleaserSnapshotOpts{
				Clean: clean,
			})
		if err != nil {
			panic(err)
		}
	} else {
		_, err := dag.
			Goreleaser(dagger.GoreleaserOpts{
				Version: goreleaserVersion,
			}).
			WithSource(source).
			Release(ctx, dagger.GoreleaserReleaseOpts{
				Clean: clean,
			})
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func parseVersion(ctx context.Context, version string) (int64, int64, int64, error) {
	latestVersion, err := getLatestVersion(ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	parsedLatestVersion, err := semver.NewVersion(latestVersion)
	if err != nil {
		return 0, 0, 0, err
	}

	switch version {
	case "major":
		return parsedLatestVersion.Major() + 1, 0, 0, nil
	case "minor":
		return parsedLatestVersion.Major(), parsedLatestVersion.Minor() + 1, 0, nil
	case "patch":
		return parsedLatestVersion.Major(), parsedLatestVersion.Minor(), parsedLatestVersion.Patch() + 1, nil
	default:
		v, err := semver.NewVersion(version)
		if err != nil {
			return 0, 0, 0, err
		}
		return v.Major(), v.Minor(), v.Patch(), nil
	}
}

func getLatestVersion(ctx context.Context) (string, error) {
	resp, err := http.DefaultClient.Get("https://api.github.com/repos/replicatedhq/replicated/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

// checkGitTree will return true if the local git tree is clean
func checkGitTree(ctx context.Context, source *dagger.Directory) (bool, error) {
	container := dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithExec([]string{"git", "status", "--porcelain"})

	output, err := container.Stdout(ctx)
	if err != nil {
		return false, err
	}

	if len(output) == 0 {
		return true, nil
	}

	return false, nil
}
