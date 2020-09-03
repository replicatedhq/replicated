package test

import (
	"bytes"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("release create", func() {
	api := platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
	t := GinkgoT()
	var app = &apps.App{Name: mustToken(8)}

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		assert.Nil(t, err)
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

			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")
			assert.Contains(t, stdout.String(), "SEQUENCE: 1")
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

			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			assert.Contains(t, stdout.String(), "SEQUENCE: 1")
		})
	})

	Context("with valid --yaml-file in an app with no releases", func() {
		It("should create release 1", func() {
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

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "create", "--yaml-file", fileName, "--app", app.Slug})
			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)

			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			assert.Contains(t, stdout.String(), "SEQUENCE: 1")
		})
	})
})
