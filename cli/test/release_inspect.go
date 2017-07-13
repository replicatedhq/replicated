package test

import (
	"bufio"
	"bytes"
	"strconv"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/apps"
	releases "github.com/replicatedhq/replicated/gen/go/releases"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("release inspect", func() {
	t := GinkgoT()
	app := &apps.App{Name: mustToken(8)}
	var release *releases.AppReleaseInfo

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(app.Name)
		assert.Nil(t, err)

		release, err = api.CreateRelease(app.Id)
		assert.Nil(t, err)
		err = api.UpdateRelease(app.Id, release.Sequence, yaml)
		assert.Nil(t, err)
	})

	AfterEach(func() {
		deleteApp(t, app.Id)
	})

	Context("with an existing release sequence", func() {
		It("should print full details of the release", func() {
			seq := strconv.Itoa(int(release.Sequence))
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			cmd.RootCmd.SetArgs([]string{"release", "inspect", seq, "--app", app.Slug})
			cmd.RootCmd.SetOutput(&stderr)

			err := cmd.Execute(&stdout)
			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Equal(t, "SEQUENCE: "+seq, r.Text())

			assert.True(t, r.Scan())
			assert.Regexp(t, `^CREATED: \d+`, r.Text())

			assert.True(t, r.Scan())
			assert.Regexp(t, `^EDITED: \d+`, r.Text())

			assert.True(t, r.Scan())
			assert.Equal(t, "CONFIG:", r.Text())

			// remainder of output is the yaml config for the release
		})
	})
})
