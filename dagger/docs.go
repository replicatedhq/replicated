package main

import (
	"bufio"
	"bytes"
	"context"
	"dagger/replicated/internal/dagger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func (r *Replicated) GenerateDocs(
	ctx context.Context,

	// +defaultPath="./"
	source *dagger.Directory,

	githubToken *dagger.Secret,
) error {
	err := checkGitTree(ctx, source, githubToken)
	if err != nil {
		return errors.Wrap(err, "failed to check git tree")
	}

	latestVersion, err := getLatestVersion(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get latest version")
	}

	docsContainer := dag.Container().
		From("alpine/git:latest").
		WithWorkdir("/").
		With(CacheBustingExec([]string{"git", "clone", "--depth", "1", "https://github.com/replicatedhq/replicated-docs.git", "/replicated-docs"}))

	rootDocsDirectory := docsContainer.Directory("/replicated-docs")
	docsDirectory := rootDocsDirectory.Directory("/docs/reference/")

	// Remove existing CLI docs
	existingDocs, err := docsDirectory.Entries(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get existing docs")
	}

	for _, existingDoc := range existingDocs {
		// 'replicated-cli-installing.mdx' is a special file that's not a CLI doc
		if existingDoc == "replicated-cli-installing.mdx" {
			continue
		}
		if !strings.HasPrefix(existingDoc, "replicated-cli") {
			continue
		}

		docsDirectory = docsDirectory.WithoutFile(existingDoc)
	}

	// Generate CLI new docs
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
		With(CacheBustingExec([]string{"go", "run", "./docs/gen.go"}))

	generatedDocs := docs.Directory("/go/src/github.com/replicatedhq/replicated/gen/docs")

	entries, err := generatedDocs.Entries(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get generated docs")
	}

	// Add new CLI docs to the docs directory while updating file names and fixing up header level
	newDocFilenames := []string{}
	for _, entry := range entries {
		file := generatedDocs.File(entry)

		content, err := file.Contents(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get generated doc contents")
		}

		content = cleanContent(content, entries)

		destFilename := cobraFileNameToDocsFileName(entry)

		docsDirectory = docsDirectory.WithNewFile(destFilename, content)
		newDocFilenames = append(newDocFilenames, destFilename)
	}

	// Update sidebar config to include new CLI docs
	sidebarFile := rootDocsDirectory.File("sidebars.js")
	sidebarContent, err := sidebarFile.Contents(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get sidebar contents")
	}

	sidebarContent, err = replaceFilenamesInSidebar(sidebarContent, newDocFilenames)
	if err != nil {
		return errors.Wrap(err, "failed to replace filenames in sidebar")
	}

	rootDocsDirectory = rootDocsDirectory.WithNewFile("sidebars.js", sidebarContent)

	docsContainer = docsContainer.
		WithMountedDirectory("/replicated-docs", rootDocsDirectory).
		WithMountedDirectory("/replicated-docs/docs/reference", docsDirectory).
		WithWorkdir("/replicated-docs").
		WithExec([]string{"git", "diff"})
	diffOut, err := docsContainer.Stdout(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get diff")
	}

	if diffOut == "" {
		return errors.New("diff is empty")
	}

	githubTokenPlaintext, err := githubToken.Plaintext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get github token plaintext")
	}

	branchName := fmt.Sprintf("update-cli-docs-%s-%s", latestVersion, time.Now().Format("2006-01-02-150405"))
	docsContainer = docsContainer.
		WithExec([]string{"git", "config", "user.email", "release@replicated.com"}).
		WithExec([]string{"git", "config", "user.name", "Replicated Release Pipeline"}).
		WithExec([]string{"git", "remote", "add", "dagger", fmt.Sprintf("https://%s@github.com/replicatedhq/replicated-docs.git", githubTokenPlaintext)}).
		WithExec([]string{"git", "checkout", "-b", branchName}).
		WithExec([]string{"git", "add", "."}).
		WithExec([]string{"git", "commit", "-m", fmt.Sprintf("Update Replicated CLI docs for %s", latestVersion)}).
		WithExec([]string{"git", "push", "dagger", branchName})

	_, err = docsContainer.Stdout(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get push output")
	}

	err = createPullRequest(ctx, branchName, fmt.Sprintf("Update Replicated CLI docs for %s", latestVersion), "", githubTokenPlaintext)
	if err != nil {
		return errors.Wrap(err, "failed to create pull request")
	}

	return nil
}

// Change names like "replicated_channel_inspect.md" to "replicated-cli-channel-inspect.mdx"
func cobraFileNameToDocsFileName(filename string) string {
	filename = strings.ReplaceAll(filename, "replicated_", "replicated-cli-")
	filename = strings.ReplaceAll(filename, "_", "-")
	if filepath.Ext(filename) == ".md" {
		filename = strings.TrimSuffix(filename, filepath.Ext(filename)) + ".mdx"
	}
	return filename
}

func cleanContent(content string, filenames []string) string {
	// Header must be level 1 in order for white spaces to be rendered correctly ("replicated api get" vs "replicated_api_get")
	if strings.HasPrefix(content, "## ") {
		content = content[1:]
	}

	// Replace all filenames in the content with the new filenames
	for _, filename := range filenames {
		topicLink := cobraFileNameToDocsFileName(filename)
		topicLink = strings.TrimSuffix(topicLink, filepath.Ext(topicLink))
		content = strings.ReplaceAll(content, filename, topicLink)
	}

	return content
}

func replaceFilenamesInSidebar(sidebarContent string, newDocFilenames []string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(sidebarContent))

	newDocLines := []string{}
	foundCLILabel := false
	wroteNewList := false
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, `reference/replicated-cli-`) { // CLI commands that are generated.
			continue
		}

		if strings.Contains(line, `'reference/replicated'`) { // CLI root command that is also generated.
			continue
		}

		if strings.Contains(line, `label: 'Replicated CLI'`) {
			newDocLines = append(newDocLines, `    label: 'Replicated CLI', // This label is generated. Do not edit.`)
			foundCLILabel = true
			continue
		}

		if foundCLILabel && !wroteNewList && strings.Contains(line, `items: [`) {
			newDocLines = append(newDocLines, `    items: [ // This list is generated. Do not edit.`)
			// 'reference/replicated-cli-installing' is a special file that's not a CLI doc and should at the top of the list
			newDocLines = append(newDocLines, `      'reference/replicated-cli-installing',`)
			for _, newDocFilename := range newDocFilenames {
				newDocLines = append(newDocLines, fmt.Sprintf(`      'reference/%s',`, strings.TrimSuffix(newDocFilename, filepath.Ext(newDocFilename))))
			}
			wroteNewList = true
			continue
		}

		newDocLines = append(newDocLines, line)
	}

	if !wroteNewList {
		return "", fmt.Errorf("no CLI list found in sidebar")
	}

	return strings.Join(newDocLines, "\n"), nil
}

func createPullRequest(ctx context.Context, branchName string, title string, body string, githubTokenPlaintext string) error {
	requestData := map[string]string{
		"title": title,
		"body":  body,
		"head":  branchName,
		"base":  "main",
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return errors.Wrap(err, "failed to marshal request data")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.github.com/repos/replicatedhq/replicated-docs/pulls", bytes.NewReader(requestBody))
	if err != nil {
		return errors.Wrap(err, "failed to create pull request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", githubTokenPlaintext))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to do pull request")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf("failed to create pull request: %s", string(respBody))
	}

	var pullRequest struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.Unmarshal(respBody, &pullRequest); err != nil {
		return errors.Wrap(err, "failed to unmarshal pull request")
	}

	fmt.Printf("created pull request at: %s\n", pullRequest.HTMLURL)

	return nil
}
