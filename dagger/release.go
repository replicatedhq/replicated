package main

import (
	"context"
	"dagger/replicated/internal/dagger"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
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
	err := checkGitTree(ctx, source, githubToken)
	if err != nil {
		return errors.Wrap(err, "failed to check git tree")
	}

	previousVersionTag, err := getLatestVersion(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get latest version")
	}

	previousReleaseBranchName, err := getReleaseBranchName(ctx, previousVersionTag)
	if err != nil {
		return errors.Wrap(err, "failed to get release branch name")
	}

	major, minor, patch, err := getNextVersion(ctx, previousVersionTag, version)
	if err != nil {
		return errors.Wrap(err, "failed to get next version")
	}

	fmt.Printf("Releasing as version %d.%d.%d\n", major, minor, patch)

	// replace the version in the Makefile
	buildFileContent, err := source.File("./pkg/version/build.go").Contents(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get build file contents")
	}
	buildFileContent = strings.ReplaceAll(buildFileContent, "const version = \"unknown\"", fmt.Sprintf("const version = \"%d.%d.%d\"", major, minor, patch))
	updatedSource := source.WithNewFile("./pkg/version/build.go", buildFileContent)

	releaseBranchName := fmt.Sprintf("release-%d.%d.%d", major, minor, patch)
	githubTokenPlaintext, err := githubToken.Plaintext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get github token plaintext")
	}

	// mount that and commit the updated build.go to git (don't push)
	// so that goreleaser won't have a dirty git tree error
	gitCommitContainer := dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", updatedSource).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		WithExec([]string{"git", "config", "user.email", "release@replicated.com"}).
		WithExec([]string{"git", "config", "user.name", "Replicated Release Pipeline"}).
		WithExec([]string{"git", "remote", "add", "dagger", fmt.Sprintf("https://%s@github.com/replicatedhq/replicated.git", githubTokenPlaintext)}).
		WithExec([]string{"git", "checkout", "-b", releaseBranchName}).
		WithExec([]string{"git", "add", "pkg/version/build.go"}).
		WithExec([]string{"git", "commit", "-m", fmt.Sprintf("Set version to %d.%d.%d", major, minor, patch)}).
		WithExec([]string{"git", "push", "dagger", releaseBranchName})
	_, err = gitCommitContainer.Stdout(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get git commit stdout")
	}
	updatedSource = gitCommitContainer.Directory("/go/src/github.com/replicatedhq/replicated")

	nextVersionTag := fmt.Sprintf("v%d.%d.%d", major, minor, patch)

	tagContainer := dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", updatedSource).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		With(CacheBustingExec([]string{"git", "tag", nextVersionTag})).
		With(CacheBustingExec([]string{"git", "push", "dagger", nextVersionTag})).
		With(CacheBustingExec([]string{"git", "fetch", "dagger", previousReleaseBranchName})).
		With(CacheBustingExec([]string{"git", "fetch", "dagger", "--tags"}))
	_, err = tagContainer.Stdout(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get tag stdout")
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
		WithoutFile("/go/src/github.com/replicatedhq/replicated/bin/replicated").
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
		return errors.Wrap(err, "failed to get docker container stdout")
	}

	username, err := dag.Onepassword().FindSecret(
		onePasswordServiceAccountProduction,
		"Developer Automation Production",
		"Docker Hub Release Account",
		"username",
	).Plaintext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get docker hub username")
	}
	password := dag.Onepassword().FindSecret(
		onePasswordServiceAccountProduction,
		"Developer Automation Production",
		"Docker Hub Release Account",
		"password",
	)

	dockerContainer = dockerContainer.WithRegistryAuth("", username, password)
	if _, err := dockerContainer.Publish(ctx, "replicated/vendor-cli:latest"); err != nil {
		return errors.Wrap(err, "failed to publish latest docker container")
	}
	if _, err := dockerContainer.Publish(ctx, fmt.Sprintf("replicated/vendor-cli:%d", major)); err != nil {
		return errors.Wrap(err, "failed to publish major docker container")
	}
	if _, err := dockerContainer.Publish(ctx, fmt.Sprintf("replicated/vendor-cli:%d.%d", major, minor)); err != nil {
		return errors.Wrap(err, "failed to publish minor docker container")
	}
	if _, err := dockerContainer.Publish(ctx, fmt.Sprintf("replicated/vendor-cli:%d.%d.%d", major, minor, patch)); err != nil {
		return errors.Wrap(err, "failed to publish patch docker container")
	}

	goreleaserContainer := dag.Goreleaser(dagger.GoreleaserOpts{
		Version: goreleaserVersion,
	}).Ctr().
		WithSecretVariable("GITHUB_TOKEN", githubToken).
		WithEnvVariable("GORELEASER_CURRENT_TAG", nextVersionTag).
		WithEnvVariable("GORELEASER_PREVIOUS_TAG", previousVersionTag)

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
			return errors.Wrap(err, "failed to snapshot goreleaser")
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
			return errors.Wrap(err, "failed to release goreleaser")
		}
	}

	return nil
}

func getNextVersion(ctx context.Context, latestVersion string, version string) (int64, int64, int64, error) {
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

func getReleaseBranchName(ctx context.Context, latestVersion string) (string, error) {
	parsedLatestVersion, err := semver.NewVersion(latestVersion)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("release-%d.%d.%d", parsedLatestVersion.Major(), parsedLatestVersion.Minor(), parsedLatestVersion.Patch()), nil
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
	// ErrGitTreeNotClean   = errors.New("Your git tree is not clean. You cannot release what's not commited.")
	ErrMainBranch        = errors.New("You must be on the main branch to release")
	ErrCommitNotInGitHub = errors.New("You must merge your changes into the main branch before releasing")
)

// checkGitTree will return nil if the local git tree is clean or an error if it's not
func checkGitTree(ctx context.Context, source *dagger.Directory, githubToken *dagger.Secret) error {
	container := dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		With(CacheBustingExec([]string{"git", "status", "--porcelain"}))

	gitStatusOutput, err := container.Stdout(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get git status")
	}

	gitStatusOutput = strings.TrimSpace(gitStatusOutput)

	if len(gitStatusOutput) > 0 {
		fmt.Printf("output: %s\n", gitStatusOutput)
		return fmt.Errorf("error: dirty tree: %q", gitStatusOutput)
	}

	container = dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		With(CacheBustingExec([]string{"git", "branch"}))

	gitBranchOutput, err := container.Stdout(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get git branch")
	}

	gitBranchOutput = strings.TrimSpace(gitBranchOutput)

	if !strings.Contains(gitBranchOutput, "* main") {
		return ErrMainBranch
	}

	container = dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/go/src/github.com/replicatedhq/replicated", source).
		WithWorkdir("/go/src/github.com/replicatedhq/replicated").
		With(CacheBustingExec([]string{"git", "rev-parse", "HEAD"}))

	commit, err := container.Stdout(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get git commit")
	}

	commit = strings.TrimSpace(commit)

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/replicatedhq/replicated/commits/%s", commit), nil)
	if err != nil {
		return errors.Wrap(err, "failed to create github request")
	}

	githubTokenPlaintext, err := githubToken.Plaintext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get github token plaintext")
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", githubTokenPlaintext))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to do github request")
	}
	defer resp.Body.Close()

	type GitHubResponse struct {
		SHA    string `json:"sha"`
		NodeID string `json:"node_id"`
		Status string `json:"status"`
	}

	var ghResp GitHubResponse
	if err := json.NewDecoder(resp.Body).Decode(&ghResp); err != nil {
		return errors.Wrap(err, "failed to decode github response")
	}

	if ghResp.Status == "422" {
		return ErrCommitNotInGitHub
	}

	return nil
}
