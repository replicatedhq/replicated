package test

import (
	"bufio"
	"bytes"
	"os"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	"github.com/replicatedhq/replicated/client"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("release create", func() {
	api := client.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
	t := GinkgoT()
	var app = &apps.App{Name: mustToken(8)}

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(&client.AppOptions{Name: app.Name})
		assert.Nil(t, err)
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("with valid --yaml in an app with no releases", func() {
		It("should create release 1", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			cmd.RootCmd.SetArgs([]string{"release", "create", "--yaml", yaml, "--app", app.Slug})
			cmd.RootCmd.SetOutput(&stderr)
			err := cmd.Execute(nil, &stdout, &stderr)

			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Equal(t, "SEQUENCE: 1", r.Text())

			assert.False(t, r.Scan())
		})
	})

	Context(`with "-" argument to --yaml flag in an app with no release where stdin contains valid yaml`, func() {
		It("should create a release from stdin", func() {
			var stdin = bytes.NewBufferString(yaml)
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			cmd.RootCmd.SetArgs([]string{"release", "create", "--yaml", "-", "--app", app.Slug})
			cmd.RootCmd.SetOutput(&stderr)
			err := cmd.Execute(stdin, &stdout, &stderr)

			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Equal(t, "SEQUENCE: 1", r.Text())
		})
	})
})
