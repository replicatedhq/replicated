package test

import (
	"bufio"
	"bytes"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	channels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("channel releases", func() {
	api := platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
	t := GinkgoT()
	app := &apps.App{Name: mustToken(8)}
	var appChan channels.AppChannel

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		assert.Nil(t, err)

		appChans, err := api.ListChannels(app.Id)
		assert.Nil(t, err)
		appChan = appChans[0]

		release, err := api.CreateRelease(app.Id, "")
		assert.Nil(t, err)
		err = api.PromoteRelease(app.Id, release.Sequence, "v1", "Big", true, appChan.Id)
		assert.Nil(t, err)

		release, err = api.CreateRelease(app.Id, "")
		assert.Nil(t, err)
		err = api.PromoteRelease(app.Id, release.Sequence, "v2", "Small", false, appChan.Id)
		assert.Nil(t, err)
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("when a channel has two releases", func() {
		It("should print a table of releases with one row", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			cmd.RootCmd.SetArgs([]string{"channel", "releases", appChan.Id, "--app", app.Slug})
			cmd.RootCmd.SetOutput(&stderr)

			err := cmd.Execute(nil, &stdout, &stderr)
			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Regexp(t, `CHANNEL_SEQUENCE\s+RELEASE_SEQUENCE\s+RELEASED\s+VERSION\s+REQUIRED\s+AIRGAP_STATUS\s+RELEASE_NOTES$`, r.Text())

			assert.True(t, r.Scan())
			words := strings.Fields(r.Text())

			// reverse chronological order
			assert.Equal(t, "1", words[0], "CHANNEL_SEQUENCE")
			assert.Equal(t, "2", words[1], "RELEASE_SEQUENCE")
			assert.Equal(t, "v2", words[3], "VERSION")
			assert.Equal(t, "false", words[4], "REQURED")
			assert.Equal(t, "Small", words[5], "RELEASE_NOTES")

			assert.True(t, r.Scan())
			words = strings.Fields(r.Text())

			assert.Equal(t, "0", words[0], "CHANNEL_SEQUENCE")
			assert.Equal(t, "1", words[1], "RELEASE_SEQUENCE")
			assert.Equal(t, "v1", words[3], "VERSION")
			assert.Equal(t, "true", words[4], "REQUIRED")
			assert.Equal(t, "Big", words[5], "RELEASE_NOTES")

			assert.False(t, r.Scan())
		})
	})
})
