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

var _ = Describe("channel inspect", func() {
	var (
		api     *platformclient.HTTPClient
		app     *apps.App
		appChan *channels.AppChannel
		err     error
	)

	BeforeEach(func() {
		api = platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
		app = &apps.App{Name: mustToken(8)}
		appChan = &channels.AppChannel{}

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
				rootCmd.SetArgs([]string{"channel", "inspect", appChan.Id, "--app", app.Slug})

				err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
				Expect(err).ToNot(HaveOccurred())

				Expect(stderr.String()).To(BeEmpty())
				Expect(stdout.String()).ToNot(BeEmpty())

				r := bufio.NewScanner(&stdout)

				Expect(r.Scan()).To(BeTrue())
				Expect(r.Text()).To(MatchRegexp(`^ID:\s+` + appChan.Id + `$`))

				Expect(r.Scan()).To(BeTrue())
				Expect(r.Text()).To(MatchRegexp(`^NAME:\s+` + appChan.Name + `$`))

				Expect(r.Scan()).To(BeTrue())
				Expect(r.Text()).To(MatchRegexp(`^DESCRIPTION:\s+` + appChan.Description + `$`))

				Expect(r.Scan()).To(BeTrue())
				Expect(r.Text()).To(MatchRegexp(`^RELEASE:\s+`))

				Expect(r.Scan()).To(BeTrue())
				Expect(r.Text()).To(MatchRegexp(`^VERSION:\s+`))
			})
		})
	})
})
