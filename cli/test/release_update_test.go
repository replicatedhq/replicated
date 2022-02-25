package test

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	releases "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

var _ = Describe("release update", func() {
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

		release, err = api.CreateRelease(app.Id, "")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("with an existing release sequence and valid --yaml", func() {
		It("should update the release's config", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			sequence := strconv.Itoa(int(release.Sequence))

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "update", sequence, "--yaml", yaml, "--app", app.Slug})

			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			r := bufio.NewScanner(&stdout)

			Expect(r.Scan()).To(BeTrue())
		})
	})

	Context("with an existing release sequence and valid --yaml-file", func() {
		It("should update the release's config", func() {
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

			sequence := strconv.Itoa(int(release.Sequence))

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "update", sequence, "--yaml-file", fileName, "--app", app.Slug})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			r := bufio.NewScanner(&stdout)

			Expect(r.Scan()).To(BeTrue())
		})
	})

	Context("error case using --yaml flag with yaml filename", func() {
		It("should return an error telling user to use --yaml-file flag", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			sequence := strconv.Itoa(int(release.Sequence))

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "update", sequence, "--yaml", "installer.yaml", "--app", app.Slug})

			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(ContainSubstring(`use the --yaml-file flag when passing a yaml filename`))
		})
	})
})
