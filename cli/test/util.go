package test

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/onsi/ginkgo"
	"github.com/replicatedhq/replicated/client"
)

var api = client.NewHTTPClient(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))

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
	data := make([]byte, int(n))
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		log.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

var sessionToken string

// for cleanup
func deleteApp(t ginkgo.GinkgoTInterface, id string) {
	origin := os.Getenv("REPLICATED_API_ORIGIN")

	if sessionToken == "" {
		email := os.Getenv("VENDOR_USER_EMAIL")
		password := os.Getenv("VENDOR_USER_PASSWORD")

		creds := map[string]interface{}{
			"email":       email,
			"password":    password,
			"remember_me": false,
		}
		buf := bytes.NewBuffer(nil)
		err := json.NewEncoder(buf).Encode(creds)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest("POST", origin+"/v1/user/login", buf)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			t.Fatalf("Login response status: %d", resp.StatusCode)
		}
		respBody := struct {
			SessionToken string `json:"token"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		if err != nil {
			t.Fatal(err)
		}
		if respBody.SessionToken == "" {
			t.Fatal("Login failed; cannot delete apps")
		}
		sessionToken = respBody.SessionToken
	}

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
