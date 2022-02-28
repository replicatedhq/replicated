package test

import (
	"bytes"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/replicatedhq/replicated/cli/cmd"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"io/ioutil"
	"os"
)

var _ = Describe("channel disable semantic versioning", func() {
	var (
		httpClient     *platformclient.HTTPClient
		kotsRestClient kotsclient.VendorV3Client
		app            *types.KotsAppWithChannels
		params         *Params
		err            error
		tmpdir         string
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

	Context("replicated channel disable-semantic-versioning --help ", func() {
		It("should print usage", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"channel", "disable-semantic-versioning", "--help"})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(ContainSubstring("Disable semantic versioning for the CHANNEL_ID."))
		})
	})

	Context("with a valid CHANNEL_ID", func() {
		It("should disable semantic versioning", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			var chanID = app.Channels[0].ID

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"channel", "disable-semantic-versioning", chanID, "--app", app.Id})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(ContainSubstring(fmt.Sprintf("Semantic versioning successfully disabled for channel %s\n", chanID)))
		})
	})
})
