package test

import (
	"bytes"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

var _ = Describe("release create", func() {
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
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("with valid --yaml in an app with no releases", func() {
		It("should create release 1", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "create", "--yaml", yaml, "--app", app.Slug})
			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)

			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(ContainSubstring("SEQUENCE: 1"))
		})
	})

	Context(`with "-" argument to --yaml flag in an app with no release where stdin contains valid yaml`, func() {
		It("should create a release from stdin", func() {
			var stdin = bytes.NewBufferString(yaml)
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "create", "--yaml", "-", "--app", app.Slug})
			err := cmd.Execute(rootCmd, stdin, &stdout, &stderr)

			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(ContainSubstring("SEQUENCE: 1"))
		})
	})

	Context("with valid --yaml-file in an app with no releases", func() {
		It("should create release 1", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			file, err := ioutil.TempFile("", app.Slug)
			Expect(err).ToNot(HaveOccurred())

			fileName := file.Name()
			defer os.Remove(fileName)

			_, err = file.WriteString(yaml)
			Expect(err).ToNot(HaveOccurred())

			err = file.Close()
			Expect(err).ToNot(HaveOccurred())

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "create", "--yaml-file", fileName, "--app", app.Slug})
			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)

			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(ContainSubstring("SEQUENCE: 1"))
		})
	})

	Context("error case using --yaml flag with yaml filename", func() {
		It("should return an error telling user to use --yaml-file flag", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			var expectedError = "use the --yaml-file flag when passing a yaml filename"

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "create", "--yaml", "installer.yaml", "--app", app.Slug})

			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedError))

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(ContainSubstring(expectedError))
		})
	})
})
