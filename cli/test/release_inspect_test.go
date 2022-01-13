package test

import (
	"bufio"
	"bytes"
	"strconv"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/v1"
	releases "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("release inspect", func() {
	t := GinkgoT()
	params, err := GetParams()
	assert.NoError(t, err)
	api := platformclient.NewHTTPClient(params.APIOrigin, params.APIToken)
	app := &apps.App{Name: mustToken(8)}
	var release *releases.AppReleaseInfo

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(&platformclient.AppOptions{Name: app.Name})
		assert.Nil(t, err)

		release, err = api.CreateRelease(app.Id, yaml)
		assert.Nil(t, err)
	})

	AfterEach(func() {
		deleteApp(app.Id)
	})

	Context("with an existing release sequence", func() {
		It("should print full details of the release", func() {
			seq := strconv.Itoa(int(release.Sequence))
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "inspect", seq, "--app", app.Slug})

			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Regexp(t, `^SEQUENCE:\s+`+seq, r.Text())

			assert.True(t, r.Scan())
			assert.Regexp(t, `^CREATED:\s+\d+`, r.Text())

			assert.True(t, r.Scan())
			assert.Regexp(t, `^EDITED:\s+\d+`, r.Text())

			assert.True(t, r.Scan())
			assert.Equal(t, "CONFIG:", r.Text())

			// remainder of output is the yaml config for the release
		})
	})
})
