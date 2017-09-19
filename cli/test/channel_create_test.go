package test

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	"github.com/replicatedhq/replicated/client"
	apps "github.com/replicatedhq/replicated/gen/go/apps"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("channel create", func() {
	api := client.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
	t := GinkgoT()
	var app = &apps.App{Name: mustToken(8)}

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(&client.AppOptions{Name: app.Name})
		assert.Nil(GinkgoT(), err)
	})

	AfterEach(func() {
		// ignore error, garbage collection
		deleteApp(app.Id)
	})

	name := mustToken(8)
	desc := mustToken(16)
	Describe(fmt.Sprintf("--name %s --description %s", name, desc), func() {
		It("should print the created channel", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			cmd.RootCmd.SetArgs([]string{"channel", "create", "--name", name, "--description", desc, "--app", app.Slug})
			cmd.RootCmd.SetOutput(&stderr)
			err := cmd.Execute(nil, &stdout, &stderr)

			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Regexp(t, `^ID\s+NAME\s+RELEASE\s+VERSION$`, r.Text())

			assert.True(t, r.Scan())
			assert.Regexp(t, `^\w+\s+`+name+`\s+$`, r.Text())

			// default Stable, Beta, and Unstable channels should be listed too
			for r.Scan() {
				assert.Regexp(t, `^\w+\s+\w+`, r.Text())
			}
		})
	})
})
