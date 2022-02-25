package test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/replicatedhq/replicated/cli/cmd"
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

var _ = Describe("kots release lint", func() {
	var (
		httpClient     *platformclient.HTTPClient
		kotsRestClient kotsclient.VendorV3Client

		app    *types.KotsAppWithChannels
		tmpdir string
		params *Params
		err    error
	)

	BeforeEach(func() {
		params, err = GetParams()
		Expect(err).ToNot(HaveOccurred())

		httpClient = platformclient.NewHTTPClient(params.APIOrigin, params.APIToken)
		kotsRestClient = kotsclient.VendorV3Client{HTTPClient: *httpClient}

		app, err = kotsRestClient.CreateKOTSApp(mustToken(8))
		Expect(err).ToNot(HaveOccurred())
		tmpdir, err = ioutil.TempDir("", "replicated-cli-test")
		Expect(err).ToNot(HaveOccurred())

	})

	AfterEach(func() {
		err := kotsRestClient.DeleteKOTSApp(app.Id)
		Expect(err).ToNot(HaveOccurred())

		err = os.RemoveAll(tmpdir)
		Expect(err).ToNot(HaveOccurred())
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
			Expect(err).ToNot(HaveOccurred())

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "lint", "--yaml-dir", tmpdir, "--app", app.Slug})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(ContainSubstring(`preflight-spec       warn                            Missing preflight spec`))
			Expect(stdout.String()).To(ContainSubstring(`config-spec          warn                            Missing config spec`))
			Expect(stdout.String()).To(ContainSubstring(`troubleshoot-spec    warn                            Missing troubleshoot spec`))
		})
	})
})
