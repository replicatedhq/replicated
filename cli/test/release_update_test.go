package test

import (
	"bufio"
	"bytes"
	"strconv"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	"github.com/replicatedhq/replicated/client"
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
		app, err = api.CreateApp(&client.AppOptions{Name: app.Name})
		assert.Nil(t, err)

		release, err = api.CreateRelease(app.Id, nil)
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
