package enterpriseclient

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
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
	privateKey, err := decodePrivateKeyPEM(privateKeyContents)
	if err != nil {
		privateKey = nil
	}
	c := &HTTPClient{
		privateKey: privateKey,
		apiOrigin:  origin,
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
		// hash the body and sign the hash
		contentSha := sha512.Sum512(bodyBytes)
		signature, err := c.privateKey.Sign(rand.Reader, contentSha[:], crypto.SHA512)
		if err != nil {
			return errors.Wrap(err, "failed to sign content sha")
		}
		req.Header.Set("Signature", base64.StdEncoding.EncodeToString(signature))

		// include the public key fingerprint as a hint to the server
		fingerprint, err := getFingerprint(&c.privateKey.PublicKey)
		if err != nil {
			return errors.Wrap(err, "failed to get public key fingerprint")
		}
		req.Header.Set("Authorization", fingerprint)
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
