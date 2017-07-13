package test

import (
	"bufio"
	"bytes"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/apps"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("release ls", func() {
	t := GinkgoT()
	app := &apps.App{Name: mustToken(8)}

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(app.Name)
		assert.Nil(t, err)

		_, err = api.CreateRelease(app.Id)
		assert.Nil(t, err)
	})

	AfterEach(func() {
		deleteApp(t, app.Id)
	})

	Context("when an app has one release", func() {
		It("should print a table of releases with one row", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			cmd.RootCmd.SetArgs([]string{"release", "ls", "--app", app.Slug})
			cmd.RootCmd.SetOutput(&stderr)

			err := cmd.Execute(&stdout)
			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Regexp(t, `SEQUENCE\s+VERSION\s+CREATED\s+EDITED\s+ACTIVE_CHANNELS`, r.Text())

			assert.True(t, r.Scan())
			assert.Regexp(t, `\d+\s+`, r.Text())

			assert.False(t, r.Scan())
		})
	})
})
