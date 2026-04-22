package cmxmetadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	mmdsIPv4Addr = "169.254.170.254"
	mmdsPath     = "/latest/vendor-api"
	mmdsTimeout  = 500 * time.Millisecond // fail fast if not in CMX
	tokenLeeway  = 60 * time.Second       // refresh token this early before expiry
)

// ErrNotAvailable is returned when the CMX metadata service is not reachable.
// This is the normal case when the CLI is not running inside a Firecracker VM.
var ErrNotAvailable = errors.New("CMX metadata service not available")

// VMMetadata holds the OIDC client credentials provisioned by vendor-api into
// the Firecracker MMDS.
type VMMetadata struct {
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
	APIURL        string `json:"api_url"`
	TokenEndpoint string `json:"token_endpoint"`
}

// GetVMMetadata attempts to read OIDC credentials from the Firecracker MMDS.
// It returns ErrNotAvailable if the metadata service is not reachable (i.e.
// the CLI is not running inside a CMX VM).
//
// Firecracker MMDS v1 returns a newline-separated key listing when querying a
// nested object path. Sending Accept: application/json causes it to return the
// full JSON subtree instead, which is what we need to parse in one request.
func GetVMMetadata() (*VMMetadata, error) {
	client := &http.Client{
		Timeout: mmdsTimeout,
	}

	mmdsURL := fmt.Sprintf("http://%s%s", mmdsIPv4Addr, mmdsPath)
	req, err := http.NewRequest(http.MethodGet, mmdsURL, nil)
	if err != nil {
		return nil, ErrNotAvailable
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrNotAvailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrNotAvailable
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrNotAvailable
	}

	var meta VMMetadata
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, ErrNotAvailable
	}

	if meta.ClientID == "" || meta.ClientSecret == "" {
		return nil, ErrNotAvailable
	}

	return &meta, nil
}

// tokenCache holds a cached access token along with its expiry time.
type tokenCache struct {
	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

// package-level cache shared across all calls within a process lifetime.
var cache = &tokenCache{}

// GetAccessToken returns a valid access token for the given VMMetadata, using a
// cached token when possible and refreshing it when it is about to expire.
func GetAccessToken(meta *VMMetadata) (string, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.token != "" && time.Until(cache.expiresAt) > tokenLeeway {
		return cache.token, nil
	}

	token, expiresAt, err := exchangeCredentials(meta)
	if err != nil {
		return "", err
	}

	cache.token = token
	if expiresAt != nil {
		cache.expiresAt = *expiresAt
	}

	return token, nil
}

// tokenResponse is the JSON response from the token endpoint.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// exchangeCredentials performs the client_credentials grant against the token
// endpoint and returns the access token along with its absolute expiry time.
func exchangeCredentials(meta *VMMetadata) (string, *time.Time, error) {
	formData := url.Values{}
	formData.Set("grant_type", "client_credentials")
	formData.Set("client_id", meta.ClientID)
	formData.Set("client_secret", meta.ClientSecret)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(
		meta.TokenEndpoint,
		"application/x-www-form-urlencoded",
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		return "", nil, fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("reading token response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("token endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", nil, fmt.Errorf("parsing token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", nil, fmt.Errorf("token endpoint returned empty access_token")
	}

	var expiresAt *time.Time
	if tokenResp.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		expiresAt = &t
	}

	return tokenResp.AccessToken, expiresAt, nil
}
