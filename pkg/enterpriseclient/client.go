package enterpriseclient

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

const apiOrigin = "https://api.replicated.com/enterprise"

// An HTTPClient communicates with the Replicated Enterprise HTTP API.
type HTTPClient struct {
	privateKey *ecdsa.PrivateKey
	apiOrigin  string
}

// New returns a new  HTTP client.
func New(privateKeyContents []byte) *HTTPClient {
	return NewHTTPClient(apiOrigin, privateKeyContents)
}

func NewHTTPClient(origin string, privateKeyContents []byte) *HTTPClient {
	c := &HTTPClient{
		apiOrigin: origin,
	}
	if privateKeyContents != nil {
		privateKey, err := decodePrivateKeyPEM(privateKeyContents)
		if err != nil {
			privateKey = nil
		}
		c.privateKey = privateKey
	}

	return c
}

func (c *HTTPClient) doJSON(method, path string, successStatus int, reqBody interface{}, respBody interface{}) error {
	endpoint := fmt.Sprintf("%s%s", c.apiOrigin, path)
	var bodyBytes []byte
	if reqBody != nil {
		var err error
		bodyBytes, err = json.Marshal(reqBody)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	if c.privateKey != nil {
		sigWithNonce, sig, fingerprint, err := sigAndFingerprint(c.privateKey, bodyBytes)
		if err != nil {
			return err
		}
		req.Header.Set("Signature", sig)
		req.Header.Set("Authorization", fingerprint)
		req.Header.Set("SignatureNonce", sigWithNonce)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to do request")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return errors.New("not found")
	}
	if resp.StatusCode != successStatus {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%s %s %d: %s", method, endpoint, resp.StatusCode, body)
	}
	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("%s %s response decoding: %v", method, endpoint, err)
		}
	}

	return nil
}
