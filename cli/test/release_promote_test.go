package test

import (
	"bufio"
	"bytes"
	"strconv"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	channels "github.com/replicatedhq/replicated/gen/go/v1"
	releases "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("release promote", func() {
	t := GinkgoT()
	params, err := GetParams()
	assert.NoError(t, err)
	api := platformclient.NewHTTPClient(params.APIOrigin, params.APIToken)
	app := &apps.App{Name: mustToken(8)}
	var appChan *channels.AppChannel
	var release *releases.AppReleaseInfo

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		assert.NoError(t, err)

		release, err = api.CreateRelease(app.Id, "")
		assert.NoError(t, err)

		appChannels, err := api.ListChannels(app.Id)
		assert.NoError(t, err)
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

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "promote", sequence, appChan.Id, "--app", app.Slug})

			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
			assert.NoError(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Equal(t, "Channel "+appChan.Id+" successfully set to release "+sequence, r.Text())

			assert.False(t, r.Scan())
		})
	})
})
