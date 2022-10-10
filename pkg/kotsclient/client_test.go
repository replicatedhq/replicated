package kotsclient

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
)

var (
	pact              dsl.Pact
	testInstallerYAML = `apiVersion: kurl.sh/v1beta1
kind: Installer
metadata:
  name: 'myapp'
help_text: |
  Please check this file exists in root directory: config.yaml
spec:
  kubernetes:
    version: latest
  docker:
    version: latest
  weave:
    version: latest
  rook:
    version: latest
  contour:
    version: latest
  registry:
    version: latest
  prometheus:
    version: latest
  kotsadm:
    version: latest
`
	testYAML = `---
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
)

func TestMain(m *testing.M) {
	if os.Getenv("SKIP_PACT_TESTING") != "" {
		return
	}

	pact = createPact()

	pact.Setup(true)

	code := m.Run()

	if err := pact.WritePact(); err != nil {
		log.Fatalf("Error writing pact file: %v", err)
	}
	pact.Teardown()

	os.Exit(code)
}

func createPact() dsl.Pact {
	dir, _ := os.Getwd()

	pactDir := path.Join(dir, "..", "..", "pacts")
	logDir := path.Join(dir, "..", "..", "logs")

	return dsl.Pact{
		Consumer: "replicated-cli",
		Provider: "vendor-api",
		LogDir:   logDir,
		PactDir:  pactDir,
		LogLevel: "debug",
	}
}
