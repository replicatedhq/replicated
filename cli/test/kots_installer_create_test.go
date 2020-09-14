package test

import (
	"bytes"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("kots installer create", func() {
	t := GinkgoT()
	req := assert.New(t) // using assert since it plays nicer with ginkgo
	params, err := GetParams()
	req.NoError(err)

	httpClient := platformclient.NewHTTPClient(params.APIOrigin, params.APIToken)
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *httpClient}
	kotsGraphqlClient := kotsclient.NewGraphQLClient(params.GraphqlOrigin, params.APIToken, params.KurlOrigin)

	var app *kotsclient.KotsApp
	var tmpdir string

	BeforeEach(func() {
		var err error
		app, err = kotsRestClient.CreateKOTSApp(mustToken(8))
		req.NoError(err)
		tmpdir, err = ioutil.TempDir("", "replicated-cli-test")
		req.NoError(err)

	})

	AfterEach(func() {
		err := kotsGraphqlClient.DeleteKOTSApp(app.ID)
		req.NoError(err)
		err = os.RemoveAll(tmpdir)
		req.NoError(err)
	})

	Context("with a standard kubernetes kurl installer", func() {
		It("should create and promote the installer", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			installer := `apiVersion: kurl.sh/v1beta1
kind: Installer
metadata:
  name: 'myapp'
spec:
  kubernetes:
    version: latest
  docker:
    version: latest
  weave:
    version: latest
  rook:
    version: latest
  contour:
    version: latest
  registry:
    version: latest
  prometheus:
    version: latest
  kotsadm:
    version: latest
`
			installerPath := filepath.Join(tmpdir, "installer.yaml")
			err := ioutil.WriteFile(installerPath, []byte(installer), 0644)
			req.NoError(err)

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"installer", "create", "--yaml-file", installerPath, "--app", app.Slug, "--promote", "Unstable"})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			req.NoError(err)

			req.Empty(stderr.String(), "Expected no stderr output")
			req.NotEmpty(stdout.String(), "Expected stdout output")

			req.Contains(stdout.String(), `SEQUENCE: 2`)
			req.Contains(stdout.String(), `successfully set to release 2`)
		})
	})
})
