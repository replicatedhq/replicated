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
	"strings"
)

var _ = Describe("kots apps", func() {
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

	Context("replicated app --help", func() {
		It("should print usage", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"app", "--help"})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(ContainSubstring("list apps and create new apps"))
		})
	})
	Context("replicated app ls", func() {
		It("should list some aps", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"app", "ls"})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(ContainSubstring(app.Id))
			Expect(stdout.String()).To(ContainSubstring(app.Name))
			Expect(stdout.String()).To(ContainSubstring("kots"))
		})
	})
	Context("replicated app ls SLUG", func() {
		It("should list just one app", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"app", "ls", app.Slug})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			expectedLsOutput := fmt.Sprintf(`ID                             NAME           SLUG           SCHEDULER
%s    %s    %s    kots
`, app.Id, app.Name, app.Slug)
			Expect(stdout.String()).To(Equal(expectedLsOutput))
		})
	})

	Context("replicated app delete", func() {
		It("should delete an app", func() {
			newName := mustToken(8)
			// this test is fragile - if the first character ends up as - , it assumes the token is a flag and fails
			newName = strings.ReplaceAll(newName, "_", "-")
			newName = strings.ReplaceAll(newName, "=", "-")
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"app", "create", newName})
			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			appSlug := strings.ToLower(newName) // maybe?

			Expect(stdout.String()).To(ContainSubstring(newName))
			Expect(stdout.String()).To(ContainSubstring(appSlug))
			Expect(stdout.String()).To(ContainSubstring("kots"))

			stdout.Truncate(0)
			rootCmd = cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"app", "delete", appSlug, "--force"})
			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			stdout.Truncate(0)
			rootCmd = cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"app", "ls", appSlug})
			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stdout.String()).ToNot(ContainSubstring(appSlug))
			Expect(stdout.String()).To(Equal(`ID    NAME    SLUG    SCHEDULER
`))
		})
	})
})
