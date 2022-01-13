package test

import (
	"bufio"
	"bytes"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("customer ls", func() {
	t := GinkgoT()
	params, err := GetParams()
	assert.NoError(t, err)

	httpClient := platformclient.NewHTTPClient(params.APIOrigin, params.APIToken)
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *httpClient}

	var app *types.KotsAppWithChannels
	var appCustomers []types.Customer

	BeforeEach(func() {
		var err error
		app, err = kotsRestClient.CreateKOTSApp(mustToken(8))
		assert.NoError(t, err)

		appChannels, err := kotsRestClient.ListChannels(app.Id, app.Slug, "")
		assert.NoError(t, err)

		kotsRestClient.CreateCustomer(mustToken(8), app.Id, appChannels[0].ID, 10*time.Minute)

		appCustomers, err = kotsRestClient.ListCustomers(app.Id)
		assert.NoError(t, err)
		assert.Len(t, appCustomers, 1)
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("when an app has three customers", func() {
		It("should print a table of customers with three row", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"customer", "ls", "--app", app.Slug})
			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)

			assert.NoError(t, err)

			assert.Zero(t, stderr, "Expected no stderr output")
			assert.NotZero(t, stdout, "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Regexp(t, `^ID\s+NAME\s+CHANNELS\s+EXPIRES\s+TYPE$`, r.Text())

			for i := 0; i < 1; i++ {
				assert.True(t, r.Scan())
				assert.Regexp(t, `^\w+\s+\w+\s+\w+\s+`, r.Text())
			}

			assert.False(t, r.Scan())
		})
	})
})
