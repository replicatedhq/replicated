package test

import (
	"bufio"
	"bytes"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	"github.com/replicatedhq/replicated/client"
	apps "github.com/replicatedhq/replicated/gen/go/apps"
	channels "github.com/replicatedhq/replicated/gen/go/channels"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("channel rm", func() {
	t := GinkgoT()
	var app = &apps.App{Name: mustToken(8)}
	var appChan *channels.AppChannel

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

	Context("when the channel ID exists", func() {
		It("should remove the channel", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			cmd.RootCmd.SetArgs([]string{"channel", "rm", appChan.Id, "--app", app.Slug})
			cmd.RootCmd.SetOutput(&stderr)

			err := cmd.Execute(&stdout)
			assert.Nil(t, err)

			assert.Zero(t, stderr, "Expected no stderr output")
			assert.NotZero(t, stdout, "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Equal(t, "Channel "+appChan.Id+" successfully archived", r.Text())

			assert.False(t, r.Scan())

			// verify with the api that it's really gone
			appChans, err := api.ListChannels(app.Id)
			assert.Nil(t, err)
			assert.Len(t, appChans, 2)
		})
	})
})
