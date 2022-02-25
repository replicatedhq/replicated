package test

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo/v2"
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

func mustToken(n int) string {
	if n == 0 {
		n = 256
	}
	data := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		log.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

var appsToDelete = make([]string, 0)

// Mark app for deletion. The actual deletion will happen in the AfterSuite when
// all tests are finished since this requires logging in as a user, which is
// rate-limited by the Vendor API.
func deleteApp(id string) {
	appsToDelete = append(appsToDelete, id)
}

func doDeleteApp(origin string, id string, t GinkgoTInterface, sessionToken string) {
	req, err := http.NewRequest("DELETE", origin+"/v1/app/"+id, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", sessionToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		t.Fatalf("Delete app response status: %d", resp.StatusCode)
	}
}

type Params struct {
	APIOrigin          string
	IDOrigin           string
	VendorUserEmail    string
	VendorUserPassword string
	GraphqlOrigin      string
	KurlOrigin         string
	APIToken           string
}

func GetParams() (*Params, error) {
	p := &Params{
		APIOrigin:          os.Getenv("REPLICATED_API_ORIGIN"),
		IDOrigin:           os.Getenv("REPLICATED_ID_ORIGIN"),
		VendorUserEmail:    os.Getenv("VENDOR_USER_EMAIL"),
		VendorUserPassword: os.Getenv("VENDOR_USER_PASSWORD"),
		GraphqlOrigin:      os.Getenv("REPLICATED_GRAPHQL_ORIGIN"),
		KurlOrigin:         os.Getenv("REPLICATED_KURL_ORIGIN"),
		APIToken:           os.Getenv("REPLICATED_API_TOKEN"),
	}
	if p.VendorUserPassword == "" || p.VendorUserEmail == "" || p.APIToken == "" {
		return nil, errors.New("Must provide each of VENDOR_USER_EMAIL, VENDOR_USER_PASSWORD, REPLICATED_API_TOKEN")
	}

	if p.APIOrigin == "" {
		p.APIOrigin = "https://api.replicated.com/vendor"
	}

	if p.IDOrigin == "" {
		p.IDOrigin = "https://id.replicated.com"
	}

	if p.GraphqlOrigin == "" {
		p.GraphqlOrigin = "https://g.replicated.com/graphql"
	}

	if p.KurlOrigin == "" {
		p.KurlOrigin = "https://kurl.sh"
	}
	return p, nil
}
