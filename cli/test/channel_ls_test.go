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

var _ = Describe("channel ls", func() {

	api := platformclient.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))
	t := GinkgoT()
	var app = &apps.App{Name: mustToken(8)}
	var appChans []channels.AppChannel

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		assert.Nil(t, err)

		appChans, err = api.ListChannels(app.Id)
		assert.Nil(t, err)
		assert.Len(t, appChans, 3)
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("when an app has three channels without releases", func() {
		It("should print all the channels", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"channel", "ls", "--app", app.Slug})
			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)

			assert.Nil(t, err)

			assert.Zero(t, stderr, "Expected no stderr output")
			assert.NotZero(t, stdout, "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Regexp(t, `^ID\s+NAME\s+RELEASE\s+VERSION$`, r.Text())

			for i := 0; i < 3; i++ {
				assert.True(t, r.Scan())
				assert.Regexp(t, `^\w+\s+\w+\s+`, r.Text())
			}

			assert.False(t, r.Scan())
		})
	})
})
