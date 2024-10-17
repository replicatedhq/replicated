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

	onePasswordServiceAccountProduction *dagger.Secret,

	githubToken *dagger.Secret,
) error {
	gitTreeOK, err := checkGitTree(ctx, source)
	if err != nil {
		return err
	}
	if !gitTreeOK {
		return fmt.Errorf("Your git tree is not clean. You cannot release what's not commited.")
	}

	major, minor, patch, err := parseVersion(ctx, version)
	if err != nil {
		return err
	}

	fmt.Printf("Releasing as version %d.%d.%d\n", major, minor, patch)

	githubTokenPlaintext, err := githubToken.Plaintext(ctx)
	if err != nil {
		return err
	}
	tag := dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithExec([]string{"git", "remote", "add", "tag", fmt.Sprintf("https://%s@github.com/replicatedhq/replicated.git", githubTokenPlaintext)}).
		WithExec([]string{"git", "tag", fmt.Sprintf("v%d.%d.%d", major, minor, patch)}).
		WithExec([]string{"git", "push", "tag", fmt.Sprintf("v%d.%d.%d", major, minor, patch)})
	if _, err := tag.Stdout(ctx); err != nil {
		return err
	}

	goModCache := dag.CacheVolume("replicated-go-mod-122")
	goBuildCache := dag.CacheVolume("replicated-go-build-121")

	replicatedBinary := dag.Container().
		From("golang:1.22").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithMountedCache("/go/pkg/mod", goModCache).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithMountedCache("/go/build-cache", goBuildCache).
		WithEnvVariable("GOCACHE", "/go/build-cache").
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
	_, err = dockerContainer.Stdout(ctx)
	if err != nil {
		return err
	}

	username, err := dag.Onepassword().FindSecret(
		onePasswordServiceAccountProduction,
		"Developer Automation Production",
		"Docker Hub Release Account",
		"username",
	).Plaintext(ctx)
	if err != nil {
		panic(err)
	}
	password := dag.Onepassword().FindSecret(
		onePasswordServiceAccountProduction,
		"Developer Automation Production",
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

	goreleaserContainer := dag.Goreleaser(dagger.GoreleaserOpts{
		Version: goreleaserVersion,
	}).Ctr().WithSecretVariable("GITHUB_TOKEN", githubToken)
	if snapshot {
		_, err := dag.
			Goreleaser(dagger.GoreleaserOpts{
				Version: goreleaserVersion,
				Ctr:     goreleaserContainer,
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
				Ctr:     goreleaserContainer,
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
