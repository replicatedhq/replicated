package test

import (
	"bufio"
	"bytes"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	releases "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("release inspect", func() {
	api := platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
	t := GinkgoT()
	app := &apps.App{Name: mustToken(8)}
	var release *releases.AppReleaseInfo

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		assert.Nil(t, err)

		release, err = api.CreateRelease(app.Id, yaml)
		assert.Nil(t, err)
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("with an existing release sequence", func() {
		It("should print full details of the release", func() {
			seq := strconv.Itoa(int(release.Sequence))
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			cmd.RootCmd.SetArgs([]string{"release", "inspect", seq, "--app", app.Slug})
			cmd.RootCmd.SetOutput(&stderr)

			err := cmd.Execute(nil, &stdout, &stderr)
			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Regexp(t, `^SEQUENCE:\s+`+seq, r.Text())

			assert.True(t, r.Scan())
			assert.Regexp(t, `^CREATED:\s+\d+`, r.Text())

			assert.True(t, r.Scan())
			assert.Regexp(t, `^EDITED:\s+\d+`, r.Text())

			assert.True(t, r.Scan())
			assert.Equal(t, "CONFIG:", r.Text())

			// remainder of output is the yaml config for the release
		})
	})
})
