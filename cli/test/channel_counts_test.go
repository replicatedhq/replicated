package test

import (
	"bufio"
	"bytes"
	"os"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	"github.com/replicatedhq/replicated/client"
	apps "github.com/replicatedhq/replicated/gen/go/apps"
	channels "github.com/replicatedhq/replicated/gen/go/channels"
	"github.com/stretchr/testify/assert"
)

// This only tests with no active licenses since the vendor API does not provide
// a way to update licenses' last_active field.
var _ = Describe("channel counts", func() {
	api := client.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
	t := GinkgoT()
	var app = &apps.App{Name: mustToken(8)}
	var appChan = &channels.AppChannel{}

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(&client.AppOptions{Name: app.Name})
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

				cmd.RootCmd.SetArgs([]string{"channel", "counts", appChan.Id, "--app", app.Slug})
				cmd.RootCmd.SetOutput(&stderr)

				err := cmd.Execute(nil, &stdout, &stderr)
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
