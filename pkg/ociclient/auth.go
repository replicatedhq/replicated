package ociclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/replicatedhq/replicated/pkg/credentials"
)

func getJWTToken(endpoint string) (string, error) {
	creds, err := credentials.GetCurrentCredentials()
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Add("service", endpoint)
	params.Add("scope", "repository:something:push,pull") // Adjust scope as needed

	tokenURL := fmt.Sprintf("https://%s/token?%s", endpoint, params.Encode())
	req, err := http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth("", creds.APIToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get JWT token: %s", resp.Status)
	}

	var tokenResponse struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	return tokenResponse.Token, nil
}
