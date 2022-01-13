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

var _ = Describe("channel inspect", func() {
	t := GinkgoT()
	params, err := GetParams()
	assert.NoError(t, err)
	api := platformclient.NewHTTPClient(params.APIOrigin, params.APIToken)
	var app = &apps.App{Name: mustToken(8)}
	var appChan = &channels.AppChannel{}

	BeforeEach(func() {
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
				rootCmd.SetArgs([]string{"channel", "inspect", appChan.Id, "--app", app.Slug})

				err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
				assert.Nil(t, err)

				assert.Zero(t, stderr, "Expected no stderr output")
				assert.NotZero(t, stdout, "Expected stdout output")

				r := bufio.NewScanner(&stdout)

				assert.True(t, r.Scan())
				assert.Regexp(t, `^ID:\s+`+appChan.Id+`$`, r.Text())

				assert.True(t, r.Scan())
				assert.Regexp(t, `^NAME:\s+`+appChan.Name+`$`, r.Text())

				assert.True(t, r.Scan())
				assert.Regexp(t, `^DESCRIPTION:\s+`+appChan.Description+`$`, r.Text())

				assert.True(t, r.Scan())
				assert.Regexp(t, `^RELEASE:\s+`, r.Text())
				/*
					assert.True(t, r.Scan())
					assert.Equal(t, "LICENSE_COUNTS", r.Text())

					assert.True(t, r.Scan())
					assert.Equal(t, "No licenses in channel", r.Text())

					assert.True(t, r.Scan())
					assert.Equal(t, "", r.Text())

					assert.True(t, r.Scan())
					assert.Equal(t, "RELEASES", r.Text())

					assert.True(t, r.Scan())
					assert.Equal(t, "No releases in channel", r.Text())
				*/
			})
		})
	})
})
