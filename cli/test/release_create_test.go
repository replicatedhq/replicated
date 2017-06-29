package test

import (
	"bufio"
	"bytes"

	. "github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/cli/cmd"
	apps "github.com/replicatedhq/replicated/gen/go/apps"
	"github.com/stretchr/testify/assert"
)

var yaml = `---
replicated_api_version: 2.9.2
name: "Test"

#
# https://www.replicated.com/docs/packaging-an-application/application-properties
#
properties:
  app_url: http://{{repl ConfigOption "hostname" }}
  console_title: "Test"

#
# https://www.replicated.com/docs/kb/supporting-your-customers/install-known-versions
#
host_requirements:
  replicated_version: ">=2.9.2"

#
# Settings screen
# https://www.replicated.com/docs/packaging-an-application/config-screen
#
config:
- name: hostname
  title: Hostname
  description: Ensure this domain name is routable on your network.
  items:
  - name: hostname
    title: Hostname
    value: '{{repl ConsoleSetting "tls.hostname"}}'
    type: text
    test_proc:
    display_name: Check DNS
    command: resolve_host

#
# Define how the application containers will be created and started
# https://www.replicated.com/docs/packaging-an-application/components-and-containers
#
components: []
`

var _ = Describe("release create", func() {
	t := GinkgoT()
	var app = &apps.App{Name: mustToken(8)}

	BeforeEach(func() {
		var err error
		app, err = api.CreateApp(app.Name)
		assert.Nil(t, err)
	})

	AfterEach(func() {
		// ignore error, garbage collection
		api.DeleteApp(app.Id)
	})

	Context("with valid --yaml in an app with no releases", func() {
		It("should create release 1", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			cmd.RootCmd.SetArgs([]string{"release", "create", "--yaml", yaml, "--app", app.Slug})
			cmd.RootCmd.SetOutput(&stderr)
			err := cmd.Execute(&stdout)

			assert.Nil(t, err)

			assert.Empty(t, stderr.String(), "Expected no stderr output")
			assert.NotEmpty(t, stdout.String(), "Expected stdout output")

			r := bufio.NewScanner(&stdout)

			assert.True(t, r.Scan())
			assert.Equal(t, "SEQUENCE: 1", r.Text())

			assert.False(t, r.Scan())
		})
	})
})
