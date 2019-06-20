package test

import (
	"bufio"
	"bytes"
	"os"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	channels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

// This only tests with no active licenses since the vendor API does not provide
// a way to update licenses' last_active field.
var _ = Describe("channel adoption", func() {
	var graphqlOrigin = "https://g.replicated.com/graphql"
	api := platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), graphqlOrigin, os.Getenv("REPLICATED_API_TOKEN"))
	t := GinkgoT()
	var app = &apps.App{Name: mustToken(8)}
	var appChan = &channels.AppChannel{}

	BeforeEach(func() {
		t.Logf("%+v\n", api)
		var err error
		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		assert.Nil(t, err)

		appChans, err := api.ListChannels(app.Id)
		assert.Nil(t, err)
		appChan = &appChans[0]
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("with an existing channel ID", func() {
		Context("with no licenses and no releases", func() {
			It("should print the full channel details", func() {
				var stdout bytes.Buffer
				var stderr bytes.Buffer

				rootCmd := cmd.GetRootCmd()
				rootCmd.SetArgs([]string{"channel", "adoption", appChan.Id, "--app", app.Slug})
				err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
				assert.Nil(t, err)

				assert.Zero(t, stderr, "Expected no stderr output")
				assert.NotZero(t, stdout, "Expected stdout output")

				r := bufio.NewScanner(&stdout)

				assert.True(t, r.Scan())
				assert.Equal(t, "No active licenses in channel", r.Text())
			})
		})
	})
})
