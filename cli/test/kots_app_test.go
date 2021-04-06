package test

import (
	"bytes"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
)

var _ = Describe("kots apps", func() {
	t := GinkgoT()
	req := assert.New(t) // using assert since it plays nicer with ginkgo
	params, err := GetParams()
	req.NoError(err, fmt.Sprintf("%+v", err))

	httpClient := platformclient.NewHTTPClient(params.APIOrigin, params.APIToken)
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *httpClient}
	kotsGraphqlClient := kotsclient.NewGraphQLClient(params.GraphqlOrigin, params.APIToken, params.KurlOrigin)

	var app *kotsclient.KotsApp
	var tmpdir string

	BeforeEach(func() {
		var err error
		app, err = kotsRestClient.CreateKOTSApp(mustToken(8))
		req.NoError(err)
		tmpdir, err = ioutil.TempDir("", "replicated-cli-test")
		req.NoError(err)

	})

	AfterEach(func() {
		err := kotsGraphqlClient.DeleteKOTSApp(app.ID)
		req.NoError(err)
		err = os.RemoveAll(tmpdir)
		req.NoError(err)
	})

	Context("replicated app --help", func() {
		It("should print usage", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"app", "--help"})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			req.NoError(err)

			req.Empty(stderr.String(), "Expected no stderr output")
			req.NotEmpty(stdout.String(), "Expected stdout output")

			req.Contains(stdout.String(), `list apps and create new apps`)
		})
	})
	Context("replicated app ls", func() {
		It("should list some aps", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"app", "ls"})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			req.NoError(err)

			req.Empty(stderr.String(), "Expected no stderr output")
			req.NotEmpty(stdout.String(), "Expected stdout output")

			req.Contains(stdout.String(), app.ID)
			req.Contains(stdout.String(), app.Name)
			req.Contains(stdout.String(), "kots")
		})
	})
	Context("replicated app ls SLUG", func() {
		It("should list just one app", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"app", "ls", app.Slug})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			req.NoError(err)

			req.Empty(stderr.String(), "Expected no stderr output")
			req.NotEmpty(stdout.String(), "Expected stdout output")

			req.Equal(stdout.String(),
`ID                             NAME           SLUG           SCHEDULER
`+ app.ID +`    ` + app.Name + `    ` + app.Slug + `    kots
`)
		})
	})

	Context("replicated app delete", func() {
		It("should delete an app", func() {
			newName := mustToken(8)
			newName = strings.ReplaceAll(newName, "_", "-")
			newName = strings.ReplaceAll(newName, "=", "-")
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"app", "create", newName})
			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			req.NoError(err)
			req.Empty(stderr.String(), "Expected no stderr output")
			req.NotEmpty(stdout.String(), "Expected stdout output")

			appSlug := strings.ToLower(newName) // maybe?

			req.Contains(stdout.String(), newName)
			req.Contains(stdout.String(), appSlug)
			req.Contains(stdout.String(), "kots")


			stdout.Truncate(0)
			rootCmd = cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"app", "delete", appSlug, "--force"})
			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			req.NoError(err)

			req.Empty(stderr.String(), "Expected no stderr output")
			req.NotEmpty(stdout.String(), "Expected stdout output")

			stdout.Truncate(0)
			rootCmd = cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"app", "ls", appSlug})
			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			req.NoError(err)

			req.NotContains(stdout.String(), appSlug)
			req.Equal(stdout.String(),
`ID    NAME    SLUG    SCHEDULER
`)
		})
	})
})
