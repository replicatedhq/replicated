package initkotsapp

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

	err = Helmignore(dir)

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

	err = Gitignore(dir)

	req.NoError(err)

	gitignorePath := filepath.Join(dir, ".gitignore")

	bytes, err := ioutil.ReadFile(gitignorePath)

	gitignoreContents := `
deps/
manifests/*.tgz
`
	req.Equal(gitignoreContents, string(bytes))
}

func TestPreflights(t *testing.T) {
	req := require.New(t)

	dir, err := ioutil.TempDir("", "")

	defer os.RemoveAll(dir)

	chartYaml := ChartYaml{
		Name:    "mychart",
		Version: "0.0.1",
	}

	err = PreflightFile(chartYaml, dir)

	req.NoError(err)

	preflightPath := filepath.Join(dir, "preflight.yaml")

	bytes, err := ioutil.ReadFile(preflightPath)

	preflightContents := `kind: Preflight
apiVersion: troubleshoot.replicated.com/v1beta1
metadata:
    name: mychart
spec:
    analyzers:
      - clusterVersion:
            checkName: Kubernetes Version
            outcomes:
              - fail:
                    when: < 1.15.0
                    message: This app requires at least Kubernetes 1.15.0
                    uri: https://www.kubernetes.io
              - pass:
                    when: '>= 1.15.0'
                    message: This app has at least Kubernetes 1.15.0
                    uri: https://www.kubernetes.io
      - nodeResources:
            checkName: Total CPU Capacity
            outcomes:
              - fail:
                    when: sum(cpuCapacity) < 4
                    message: This app requires a cluster with at least 4 cores.
                    uri: https://kurl.sh/docs/install-with-kurl/system-requirements
              - pass:
                    message: This cluster has at least 4 cores.
`

	req.Equal(preflightContents, string(bytes))
}
