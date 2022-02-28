package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	if os.Getenv("SKIP_INTEGRATION_TESTING") != "" {
		return
	}
	RunSpecs(t, "CLI Suite")
}

var _ = AfterSuite(func() {
	if len(appsToDelete) == 0 {
		return
	}
	t := GinkgoT()
	origin := os.Getenv("REPLICATED_API_ORIGIN")
	idOrigin := os.Getenv("REPLICATED_ID_ORIGIN")
	if idOrigin == "" {
		idOrigin = "https://id.replicated.com"
	}

	email := os.Getenv("VENDOR_USER_EMAIL")
	password := os.Getenv("VENDOR_USER_PASSWORD")

	if email == "" || password == "" {
		fmt.Println("VENDOR_USER_EMAIL or VENDOR_USER_PASSWORD not set. Skipping app cleanup")
		return
	}

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

	req, err := http.NewRequest("POST", idOrigin+"/v1/login", buf)
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
	sessionToken := respBody.SessionToken

	for _, id := range appsToDelete {
		doDeleteApp(origin, id, t, sessionToken)
	}
})
