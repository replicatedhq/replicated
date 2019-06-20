package test

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("channel create", func() {
	var graphqlOrigin = "https://g.replicated.com/graphql"
	api := platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), graphqlOrigin, os.Getenv("REPLICATED_API_TOKEN"))
	t := GinkgoT()
	var app = &apps.App{Name: mustToken(8)}

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
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

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"channel", "create", "--name", name, "--description", desc, "--app", app.Slug})
			rootCmd.SetOutput(&stderr)
			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)

			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Regexp(t, `^ID\s+NAME\s+RELEASE\s+VERSION$`, r.Text())

			// default Stable, Beta, and Unstable channels should be listed too
			for i := 0; i < 3; i++ {
				assert.True(t, r.Scan())
				assert.Regexp(t, `^\w+\s+\w+`, r.Text())
			}

			assert.True(t, r.Scan())
			assert.Regexp(t, `^\w+\s+`+name+`\s+$`, r.Text())
		})
	})
})
