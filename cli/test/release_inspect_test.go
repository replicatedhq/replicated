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
	releases "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

var _ = Describe("release inspect", func() {
	var (
		api     *platformclient.HTTPClient
		app     *apps.App
		release *releases.AppReleaseInfo
		err     error
	)

	BeforeEach(func() {
		api = platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
		app = &apps.App{Name: mustToken(8)}

		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		Expect(err).ToNot(HaveOccurred())

		release, err = api.CreateRelease(app.Id, yaml)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("with an existing release sequence", func() {
		It("should print full details of the release", func() {
			seq := strconv.Itoa(int(release.Sequence))
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "inspect", seq, "--app", app.Slug})

			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			r := bufio.NewScanner(&stdout)

			Expect(r.Scan()).To(BeTrue())
			Expect(r.Text()).To(MatchRegexp(`^SEQUENCE:\s+` + seq))

			Expect(r.Scan()).To(BeTrue())
			Expect(r.Text()).To(MatchRegexp(`^CREATED:\s+\d+`))

			Expect(r.Scan()).To(BeTrue())
			Expect(r.Text()).To(MatchRegexp(`^EDITED:\s+\d+`))

			Expect(r.Scan()).To(BeTrue())
			Expect(r.Text()).To(Equal("CONFIG:"))

			// remainder of output is the yaml config for the release
		})
	})
})
