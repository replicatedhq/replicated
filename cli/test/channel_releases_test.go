package test

import (
	"bufio"
	"bytes"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	channels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

var _ = Describe("channel releases", func() {
	var (
		api     *platformclient.HTTPClient
		app     *apps.App
		appChan channels.AppChannel
		err     error
	)

	BeforeEach(func() {
		api = platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
		app = &apps.App{Name: mustToken(8)}
		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		Expect(err).ToNot(HaveOccurred())

		appChans, err := api.ListChannels(app.Id)
		Expect(err).ToNot(HaveOccurred())
		appChan = appChans[0]

		release, err := api.CreateRelease(app.Id, "")
		Expect(err).ToNot(HaveOccurred())
		err = api.PromoteRelease(app.Id, release.Sequence, "v1", "Big", true, appChan.Id)
		Expect(err).ToNot(HaveOccurred())

		release, err = api.CreateRelease(app.Id, "")
		Expect(err).ToNot(HaveOccurred())
		err = api.PromoteRelease(app.Id, release.Sequence, "v2", "Small", false, appChan.Id)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("when a channel has two releases", func() {
		It("should print a table of releases with one row", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"channel", "releases", appChan.Id, "--app", app.Slug})

			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			r := bufio.NewScanner(&stdout)

			Expect(r.Scan()).To(BeTrue())
			Expect(r.Text()).To(MatchRegexp(`CHANNEL_SEQUENCE\s+RELEASE_SEQUENCE\s+RELEASED\s+VERSION\s+REQUIRED\s+AIRGAP_STATUS\s+RELEASE_NOTES$`))

			Expect(r.Scan()).To(BeTrue())
			words := strings.Fields(r.Text())

			// reverse chronological order
			Expect(words[0]).To(Equal("1"))     // CHANNEL_SEQUENCE
			Expect(words[1]).To(Equal("2"))     // RELEASE_SEQUENCE
			Expect(words[3]).To(Equal("v2"))    // VERSION
			Expect(words[4]).To(Equal("false")) // REQUIRED
			Expect(words[5]).To(Equal("Small")) // RELEASE_NOTES

			Expect(r.Scan()).To(BeTrue())
			words = strings.Fields(r.Text())
			Expect(words[0]).To(Equal("0"))    // CHANNEL_SEQUENCE
			Expect(words[1]).To(Equal("1"))    // RELEASE_SEQUENCE
			Expect(words[3]).To(Equal("v1"))   // VERSION
			Expect(words[4]).To(Equal("true")) // REQUIRED
			Expect(words[5]).To(Equal("Big"))  // RELEASE_NOTES

			Expect(r.Scan()).To(BeFalse())
		})
	})
})
