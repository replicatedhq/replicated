package test

import (
	"bufio"
	"bytes"
	"os"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("release ls", func() {
	api := platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_SHIP_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
	t := GinkgoT()
	app := &apps.App{Name: mustToken(8)}

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		assert.Nil(t, err)

		_, err = api.CreateRelease(app.Id, "")
		assert.Nil(t, err)
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("when an app has one release", func() {
		It("should print a table of releases with one row", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "ls", "--app", app.Slug})
			rootCmd.SetOutput(&stderr)

			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Regexp(t, `SEQUENCE\s+CREATED\s+EDITED\s+ACTIVE_CHANNELS`, r.Text())

			assert.True(t, r.Scan())
			assert.Regexp(t, `\d+\s+`, r.Text())

			assert.False(t, r.Scan())
		})
	})
})
