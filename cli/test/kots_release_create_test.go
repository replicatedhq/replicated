package test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/types"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("kots release create", func() {
	t := GinkgoT()
	req := assert.New(t) // using assert since it plays nicer with ginkgo
	params, err := GetParams()
	req.NoError(err)

	httpClient := platformclient.NewHTTPClient(params.APIOrigin, params.APIToken)
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *httpClient}

	var app *types.KotsAppWithChannels
	var tmpdir string

	BeforeEach(func() {
		var err error
		app, err = kotsRestClient.CreateKOTSApp(mustToken(8))
		req.NoError(err)
		tmpdir, err = ioutil.TempDir("", "replicated-cli-test")
		req.NoError(err)

	})

	AfterEach(func() {
		err := kotsRestClient.DeleteKOTSApp(app.Id)
		req.NoError(err)
		err = os.RemoveAll(tmpdir)
		req.NoError(err)
	})

	Context("with valid --yaml-dir in an app with no releases", func() {
		It("should create release 1", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			configMap := `apiVersion: v1
kind: ConfigMap
metadata:
  name: fake
data:
  fake: yep it's fake
`
			err := ioutil.WriteFile(filepath.Join(tmpdir, "config.yaml"), []byte(configMap), 0644)
			req.NoError(err)

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "create", "--yaml-dir", tmpdir, "--app", app.Slug})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			req.NoError(err)

			req.Empty(stderr.String(), "Expected no stderr output")
			req.NotEmpty(stdout.String(), "Expected stdout output")
			req.Contains(stdout.String(), "SEQUENCE: 1")
		})
	})
	Context("with the --promote=Unstable flag", func() {
		It("should create release 1 and promote it", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			configMap := `apiVersion: v1
kind: ConfigMap
metadata:
  name: fake
data:
  fake: yep it's fake
`
			err := ioutil.WriteFile(filepath.Join(tmpdir, "config.yaml"), []byte(configMap), 0644)
			req.NoError(err)

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "create", "--yaml-dir", tmpdir, "--app", app.Slug, "--promote", "Unstable"})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			req.NoError(err)

			req.Empty(stderr.String(), "Expected no stderr output")
			req.NotEmpty(stdout.String(), "Expected stdout output")

			req.Contains(stdout.String(), `• Reading manifests from `+tmpdir)
			req.Contains(stdout.String(), `• Creating Release`)
			req.Contains(stdout.String(), `• SEQUENCE: 1`)
			req.Contains(stdout.String(), `• Promoting`)
			req.Contains(stdout.String(), "successfully set to release 1")

			// download it back down and verify content
			stdout.Reset()
			stderr.Reset()
			downloadTmpDir, err := ioutil.TempDir("", "replicated-cli-test")
			req.NoError(err)
			defer os.RemoveAll(downloadTmpDir)

			rootCmd = cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "download", "1", "--dest", downloadTmpDir, "--app", app.Slug})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			req.NoError(err)

			downloadedFile, err := ioutil.ReadFile(filepath.Join(downloadTmpDir, "config.yaml"))
			req.NoError(err)
			req.Equal(configMap, string(downloadedFile))

		})
	})
})
