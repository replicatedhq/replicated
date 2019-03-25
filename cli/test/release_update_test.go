package test

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	releases "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("release update", func() {
	api := platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
	t := GinkgoT()
	app := &apps.App{Name: mustToken(8)}
	var release *releases.AppReleaseInfo

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		assert.Nil(t, err)

		release, err = api.CreateRelease(app.Id, "")
		assert.Nil(t, err)
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
			rootCmd.SetOutput(&stderr)

			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
		})
	})

	Context("with an existing release sequence and valid --yaml-file", func() {
		It("should update the release's config", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			file, err := ioutil.TempFile("", app.Slug)
			assert.Nil(t, err)
			fileName := file.Name()
			defer os.Remove(fileName)
			_, err = file.WriteString(yaml)
			assert.Nil(t, err)
			err = file.Close()
			assert.Nil(t, err)

			sequence := strconv.Itoa(int(release.Sequence))

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "update", sequence, "--yaml-file", fileName, "--app", app.Slug})
			rootCmd.SetOutput(&stderr)

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
		})
	})
})
