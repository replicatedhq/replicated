package test

import (
	"bufio"
	"bytes"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	channels "github.com/replicatedhq/replicated/gen/go/v1"
	releases "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

var _ = Describe("release promote", func() {
	var (
		api     *platformclient.HTTPClient
		app     *apps.App
		appChan *channels.AppChannel
		release *releases.AppReleaseInfo
		err     error
	)

	BeforeEach(func() {
		api = platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
		app = &apps.App{Name: mustToken(8)}
		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		Expect(err).ToNot(HaveOccurred())

		release, err = api.CreateRelease(app.Id, "")
		Expect(err).ToNot(HaveOccurred())

		appChannels, err := api.ListChannels(app.Id)
		Expect(err).ToNot(HaveOccurred())
		appChan = &appChannels[0]
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("when a channel with no releases is promoted to release 1", func() {
		It("should succeed", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			sequence := strconv.Itoa(int(release.Sequence))

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "promote", sequence, appChan.Id, "--app", app.Slug})

			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			r := bufio.NewScanner(&stdout)

			Expect(r.Scan()).To(BeTrue())
			Expect(r.Text()).To(Equal("Channel " + appChan.Id + " successfully set to release " + sequence))

			Expect(r.Scan()).To(BeFalse())
		})
	})
})
