package main

import (
	"context"
	"dagger/replicated/internal/dagger"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

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
	gitTreeOK, err := checkGitTree(ctx, source, githubToken)
	if err != nil {
		return err
	}
	if !gitTreeOK {
		return fmt.Errorf("Your git tree is not clean. You cannot release what's not commited.")
	}

	latestVersion, err := getLatestVersion(ctx)
	if err != nil {
		return err
	}

	major, minor, patch, err := parseVersion(ctx, latestVersion, version)
	if err != nil {
		return err
	}

	fmt.Printf("Releasing as version %d.%d.%d\n", major, minor, patch)

	// replace the version in the Makefile
	buildFileContent, err := source.File("./pkg/version/build.go").Contents(ctx)
	if err != nil {
		return err
	}
	buildFileContent = strings.ReplaceAll(buildFileContent, "const version = \"unknown\"", fmt.Sprintf("const version = \"%d.%d.%d\"", major, minor, patch))
	updatedSource := source.WithNewFile("./pkg/version/build.go", buildFileContent)

	// mount that and commit the updated build.go to git (don't push)
	// so that goreleaser won't have a dirty git tree error
	gitCommitContainer := dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", updatedSource).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithExec([]string{"git", "config", "user.email", "release@replicated.com"}).
		WithExec([]string{"git", "config", "user.name", "Replicated Release Pipeline"}).
		WithExec([]string{"git", "add", "pkg/version/build.go"}).
		WithExec([]string{"git", "commit", "-m", fmt.Sprintf("Set version to %d.%d.%d", major, minor, patch)})
	_, err = gitCommitContainer.Stdout(ctx)
	if err != nil {
		return err
	}
	updatedSource = gitCommitContainer.Directory("/go/src/github.com/replicatedhq/replicated")

	githubTokenPlaintext, err := githubToken.Plaintext(ctx)
	if err != nil {
		return err
	}
	tagContainer := dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", updatedSource).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithExec([]string{"git", "remote", "add", "tag", fmt.Sprintf("https://%s@github.com/replicatedhq/replicated.git", githubTokenPlaintext)}).
		With(CacheBustingExec([]string{"git", "tag", fmt.Sprintf("v%d.%d.%d", major, minor, patch)})).
		With(CacheBustingExec([]string{"git", "push", "tag", fmt.Sprintf("v%d.%d.%d", major, minor, patch)}))
	_, err = tagContainer.Stdout(ctx)
	if err != nil {
		return err
	}

	// copy the source that has the tag included in it
	updatedSource = tagContainer.Directory("/go/src/github.com/replicatedhq/replicated")

	goModCache := dag.CacheVolume("replicated-go-mod-122")
	goBuildCache := dag.CacheVolume("replicated-go-build-121")

	replicatedBinary := dag.Container(dagger.ContainerOpts{
		Platform: "linux/amd64",
	}).
		From("golang:1.22").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", updatedSource).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithMountedCache("/go/pkg/mod", goModCache).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithMountedCache("/go/build-cache", goBuildCache).
		WithEnvVariable("GOCACHE", "/go/build-cache").
		With(CacheBustingExec([]string{"make", "build"})).
		File("/go/src/github.com/replicatedhq/replicated/bin/replicated")

	dockerContainer := dag.Container(dagger.ContainerOpts{
		Platform: "linux/amd64",
	}).
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
			WithSource(updatedSource).
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
			WithSource(updatedSource).
			Release(ctx, dagger.GoreleaserReleaseOpts{
				Clean: clean,
			})
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func parseVersion(ctx context.Context, latestVersion string, version string) (int64, int64, int64, error) {
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

var (
	ErrGitTreeNotClean   = errors.New("Your git tree is not clean. You cannot release what's not commited.")
	ErrMainBranch        = errors.New("You must be on the main branch to release")
	ErrCommitNotInGitHub = errors.New("You must merge your changes into the main branch before releasing")
)

// checkGitTree will return true if the local git tree is clean
func checkGitTree(ctx context.Context, source *dagger.Directory, githubToken *dagger.Secret) (bool, error) {
	container := dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithExec([]string{"git", "status", "--porcelain"})

	output, err := container.Stdout(ctx)
	if err != nil {
		return false, err
	}

	if len(output) > 0 {
		return false, ErrGitTreeNotClean
	}

	container = dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithExec([]string{"git", "branch"})

	output, err = container.Stdout(ctx)
	if err != nil {
		return false, err
	}

	if !strings.Contains(output, "* main") {
		return false, ErrMainBranch
	}

	container = dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithExec([]string{"git", "rev-parse", "HEAD"})

	commit, err := container.Stdout(ctx)
	if err != nil {
		return false, err
	}

	commit = strings.TrimSpace(commit)

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/replicatedhq/replicated/commits/%s", commit), nil)
	if err != nil {
		return false, err
	}

	githubTokenPlaintext, err := githubToken.Plaintext(ctx)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", githubTokenPlaintext))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	type GitHubResponse struct {
		SHA    string `json:"sha"`
		NodeID string `json:"node_id"`
		Status string `json:"status"`
	}

	var ghResp GitHubResponse
	if err := json.NewDecoder(resp.Body).Decode(&ghResp); err != nil {
		return false, err
	}

	if ghResp.Status == "422" {
		return false, ErrCommitNotInGitHub
	}

	return false, nil
}
