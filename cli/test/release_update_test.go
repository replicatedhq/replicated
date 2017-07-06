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

var _ = Describe("release update", func() {
	t := GinkgoT()
	app := &apps.App{Name: mustToken(8)}
	var release *releases.AppReleaseInfo

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(app.Name)
		assert.Nil(t, err)

		release, err = api.CreateRelease(app.Id)
		assert.Nil(t, err)
	})

	AfterEach(func() {
		// ignore error, garbage collection
		api.DeleteApp(app.Id)
	})

	Context("with an existing release sequence and valid --yaml", func() {
		It("should update the release's config", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			sequence := strconv.Itoa(int(release.Sequence))

			cmd.RootCmd.SetArgs([]string{"release", "update", sequence, "--yaml", yaml, "--app", app.Slug})
			cmd.RootCmd.SetOutput(&stderr)

			err := cmd.Execute(&stdout)
			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
		})
	})
})
