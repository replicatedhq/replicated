package test

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

var _ = Describe("channel create", func() {
	var (
		api  *platformclient.HTTPClient
		app  *apps.App
		name string
		desc string
		err  error
	)

	BeforeEach(func() {
		api = platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
		app = &apps.App{Name: mustToken(8)}
		name = mustToken(8)
		desc = mustToken(16)

		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Describe(fmt.Sprintf("--name %s --description %s", name, desc), func() {
		It("should print the created channel", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"channel", "create", "--name", name, "--description", desc, "--app", app.Slug})
			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			r := bufio.NewScanner(&stdout)

			Expect(r.Scan()).To(BeTrue())
			Expect(r.Text()).To(MatchRegexp(`^ID\s+NAME\s+RELEASE\s+VERSION$`))

			// default Stable, Beta, and Unstable channels should be listed too
			for i := 0; i < 3; i++ {
				Expect(r.Scan()).To(BeTrue())
				Expect(r.Text()).To(MatchRegexp(`^\w+\s+\w+`))
			}

			Expect(r.Scan()).To(BeTrue())
			Expect(r.Text()).To(MatchRegexp(`^\w+\s+` + name + `\s+$`))
		})
	})
})
