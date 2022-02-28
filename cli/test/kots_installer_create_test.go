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

var _ = Describe("kots installer create", func() {
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

	Context("with a standard kubernetes kurl installer", func() {
		It("should create and promote the installer", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			installer := `apiVersion: kurl.sh/v1beta1
kind: Installer
metadata:
  name: 'myapp'
help_text: |
  Please check this file exists in root directory: config.yaml
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
			Expect(err).ToNot(HaveOccurred())

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"installer", "create", "--yaml-file", installerPath, "--app", app.Slug, "--promote", "Unstable"})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(ContainSubstring(`SEQUENCE: 2`))
			Expect(stdout.String()).To(ContainSubstring(`successfully set to installer 2`))

			rootCmd.SetArgs([]string{"installer", "ls", "--app", app.Slug})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
		})
	})

	Context("error case using --yaml flag with yaml filename", func() {
		It("should return an error telling user to use --yaml-file flag", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			var errorMsg = "use the --yaml-file flag when passing a yaml filename"

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"installer", "create", "--yaml", "installer.yaml", "--app", app.Slug, "--promote", "Unstable"})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(errorMsg))

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(ContainSubstring(errorMsg))
		})
	})
})
