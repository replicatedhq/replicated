package test

import (
	"bufio"
	"bytes"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	channels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("channel rm", func() {
	t := GinkgoT()
	params, err := GetParams()
	assert.NoError(t, err)
	api := platformclient.NewHTTPClient(params.APIOrigin, params.APIToken)
	var app = &apps.App{Name: mustToken(8)}
	var appChan *channels.AppChannel

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		assert.Nil(t, err)

		appChans, err := api.ListChannels(app.Id)
		assert.Nil(t, err)

		// can't archive the default channel
		for _, channel := range appChans {
			if !channel.IsDefault {
				appChan = &channel
				break
			}
		}
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("when the channel ID exists", func() {
		It("should remove the channel", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			args := []string{"channel", "rm", appChan.Id, "--app", app.Slug}
			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs(args)

			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
			assert.Nil(t, err, "execute channel rm -- args: %v", args)

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
