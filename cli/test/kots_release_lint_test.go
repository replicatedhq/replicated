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

var _ = Describe("kots release lint", func() {
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

	Context("with just a single config map", func() {
		It("should have errors about missing files", func() {
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
			rootCmd.SetArgs([]string{"release", "lint", "--yaml-dir", tmpdir, "--app", app.Slug})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			req.NoError(err)

			req.Empty(stderr.String(), "Expected no stderr output")
			req.NotEmpty(stdout.String(), "Expected stdout output")

			req.Contains(stdout.String(), `preflight-spec       warn                            Missing preflight spec`)
			req.Contains(stdout.String(), `config-spec          warn                            Missing config spec`)
			req.Contains(stdout.String(), `troubleshoot-spec    warn                            Missing troubleshoot spec`)
		})
	})
})
