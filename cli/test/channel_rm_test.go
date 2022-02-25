package test

import (
	"bufio"
	"bytes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	channels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"os"
)

var _ = Describe("channel rm", func() {
	var (
		api     *platformclient.HTTPClient
		app     *apps.App
		appChan *channels.AppChannel
		err     error
	)

	BeforeEach(func() {
		api = platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
		app = &apps.App{Name: mustToken(8)}

		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		Expect(err).ToNot(HaveOccurred())

		appChans, err := api.ListChannels(app.Id)
		Expect(err).ToNot(HaveOccurred())

		// can't archive the default channel
		for _, channel := range appChans {
			if !channel.IsDefault {
				appChan = &channel
				break
			}
		}
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("when the channel ID exists", func() {
		It("should remove the channel", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			// verify length of initial channels
			appChans, err := api.ListChannels(app.Id)
			Expect(err).ToNot(HaveOccurred())
			Expect(appChans).To(HaveLen(3))

			args := []string{"channel", "rm", appChan.Id, "--app", app.Slug}
			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs(args)

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			r := bufio.NewScanner(&stdout)

			Expect(r.Scan()).To(BeTrue())
			Expect(r.Text()).To(Equal("Channel " + appChan.Id + " successfully archived"))

			Expect(r.Scan()).To(BeFalse())

			// verify with the api that it's really gone
			appChans, err = api.ListChannels(app.Id)
			Expect(err).ToNot(HaveOccurred())
			Expect(appChans).To(HaveLen(2))
		})
	})
})
