package test

import (
	"bytes"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	channels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

// This only tests with no active licenses since the vendor API does not provide
// a way to update licenses' last_active field.
var _ = Describe("channel counts", func() {
	var (
		api     *platformclient.HTTPClient
		app     *apps.App
		appChan *channels.AppChannel
		err     error
	)

	BeforeEach(func() {
		api = platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
		appChan = &channels.AppChannel{}
		app = &apps.App{Name: mustToken(8)}

		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		Expect(err).ToNot(HaveOccurred())

		appChans, err := api.ListChannels(app.Id)
		Expect(err).ToNot(HaveOccurred())
		appChan = &appChans[0]
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("with an existing channel ID", func() {
		Context("with no licenses and no releases", func() {
			It("should print the full channel details", func() {
				var stdout bytes.Buffer
				var stderr bytes.Buffer

				rootCmd := cmd.GetRootCmd()
				rootCmd.SetArgs([]string{"channel", "counts", appChan.Id, "--app", app.Slug})

				err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
				Expect(err).ToNot(HaveOccurred())

				Expect(stderr.String()).To(BeEmpty())
				Expect(stdout.String()).ToNot(BeEmpty())

				Expect(stdout.String()).To(Equal("No active licenses in channel\n"))
			})
		})
	})
})
