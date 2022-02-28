package test

import (
	"bufio"
	"bytes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"os"
)

var _ = Describe("release ls", func() {
	var (
		api *platformclient.HTTPClient
		app *apps.App
		err error
	)

	BeforeEach(func() {
		api = platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
		app = &apps.App{Name: mustToken(8)}

		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		Expect(err).ToNot(HaveOccurred())

		_, err = api.CreateRelease(app.Id, "")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("when an app has one release", func() {
		It("should print a table of releases with one row", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "ls", "--app", app.Slug})

			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			r := bufio.NewScanner(&stdout)

			Expect(r.Scan()).To(BeTrue())
			Expect(r.Text()).To(MatchRegexp(`SEQUENCE\s+CREATED\s+EDITED\s+ACTIVE_CHANNELS`))

			Expect(r.Scan()).To(BeTrue())
			Expect(r.Text()).To(MatchRegexp(`\d+\s+`))

			Expect(r.Scan()).To(BeFalse())
		})
	})
})
