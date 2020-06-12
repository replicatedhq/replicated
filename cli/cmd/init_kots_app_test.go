package cmd

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateHelmIgnore(t *testing.T) {
	req := require.New(t)

	dir, err := ioutil.TempDir("", "")

	req.NoError(err)

	defer os.RemoveAll(dir)

	err = helmignore(dir)

	req.NoError(err)

	makeFilePath := filepath.Join(dir, ".helmignore")

	bytes, err := ioutil.ReadFile(makeFilePath)

	helmignoreContents := `
kots/`

	req.Equal(helmignoreContents, string(bytes))
}

func TestUpdateGitIgnore(t *testing.T) {
	req := require.New(t)

	dir, err := ioutil.TempDir("", "")

	req.NoError(err)

	defer os.RemoveAll(dir)

	err = gitignore(dir)

	req.NoError(err)

	gitignorePath := filepath.Join(dir, ".gitignore")

	bytes, err := ioutil.ReadFile(gitignorePath)

	gitignoreContents := `
deps/
manifests/*.tgz
`
	req.Equal(gitignoreContents, string(bytes))
}
