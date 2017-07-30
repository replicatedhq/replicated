package test

import (
	"bufio"
	"bytes"
	"strconv"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	"github.com/replicatedhq/replicated/client"
	apps "github.com/replicatedhq/replicated/gen/go/apps"
	channels "github.com/replicatedhq/replicated/gen/go/channels"
	releases "github.com/replicatedhq/replicated/gen/go/releases"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("release promote", func() {
	t := GinkgoT()
	app := &apps.App{Name: mustToken(8)}
	var appChan *channels.AppChannel
	var release *releases.AppReleaseInfo

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(&client.AppOptions{Name: app.Name})
		assert.Nil(t, err)

		release, err = api.CreateRelease(app.Id, nil)
		assert.Nil(t, err)

		appChannels, err := api.ListChannels(app.Id)
		assert.Nil(t, err)
		appChan = &appChannels[0]
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("when a channel with no releases is promoted to release 1", func() {
		It("should succeed", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			sequence := strconv.Itoa(int(release.Sequence))

			cmd.RootCmd.SetArgs([]string{"release", "promote", sequence, appChan.Id, "--app", app.Slug})
			cmd.RootCmd.SetOutput(&stderr)

			err := cmd.Execute(&stdout)
			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Equal(t, "Channel "+appChan.Id+" successfully set to release "+sequence, r.Text())

			assert.False(t, r.Scan())
		})
	})
})
