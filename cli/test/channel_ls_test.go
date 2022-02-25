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

var _ = Describe("channel ls", func() {

	var (
		api      *platformclient.HTTPClient
		app      *apps.App
		appChans []channels.AppChannel
		err      error
	)

	BeforeEach(func() {
		api = platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
		app = &apps.App{Name: mustToken(8)}

		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		Expect(err).ToNot(HaveOccurred())

		appChans, err = api.ListChannels(app.Id)
		Expect(err).ToNot(HaveOccurred())
		Expect(appChans).To(HaveLen(3))
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("when an app has three channels without releases", func() {
		It("should print all the channels", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"channel", "ls", "--app", app.Slug})
			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			r := bufio.NewScanner(&stdout)

			Expect(r.Scan()).To(BeTrue())
			Expect(r.Text()).To(MatchRegexp(`^ID\s+NAME\s+RELEASE\s+VERSION$`))

			for i := 0; i < 3; i++ {
				Expect(r.Scan()).To(BeTrue())
				Expect(r.Text()).To(MatchRegexp(`^\w+\s+\w+\s+`))
			}

			Expect(r.Scan()).To(BeFalse())
		})
	})
})
