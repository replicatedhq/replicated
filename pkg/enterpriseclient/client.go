package enterpriseclient

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

const apiOrigin = "https://api.replicated.com/enterprise"

// An HTTPClient communicates with the Replicated Enterprise HTTP API.
type HTTPClient struct {
	privateKeyContents []byte
	apiOrigin          string
}

// New returns a new  HTTP client.
func New(privateKeyContents []byte) *HTTPClient {
	return NewHTTPClient(apiOrigin, privateKeyContents)
}

func NewHTTPClient(origin string, privateKeyContents []byte) *HTTPClient {
	c := &HTTPClient{
		privateKeyContents: privateKeyContents,
		apiOrigin:          origin,
	}

	return c
}

func (c *HTTPClient) doJSON(method, path string, successStatus int, reqBody interface{}, respBody interface{}) error {
	endpoint := fmt.Sprintf("%s%s", c.apiOrigin, path)
	var buf bytes.Buffer
	if reqBody != nil {
		if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, endpoint, &buf)
	if err != nil {
		return err
	}

	if c.privateKeyContents != nil {
		// get the private key id as a hint to the server
		var parsedKey interface{}
		decodedPEM, _ := pem.Decode(c.privateKeyContents)
		if parsedKey, err = x509.ParsePKCS1PrivateKey(decodedPEM.Bytes); err != nil {
			return errors.Wrap(err, "failed to parse private key")
		}

		privateKey, ok := parsedKey.(*rsa.PrivateKey)
		if !ok {
			return errors.Wrap(err, "failed to cast key")
		}

		// the key id is the sha256 sum of the public key
		pubDER := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
		pubBlock := pem.Block{
			Type:    "RSA PUBLIC KEY",
			Headers: nil,
			Bytes:   pubDER,
		}
		pubPEM := pem.EncodeToMemory(&pubBlock)

		req.Header.Set("Authorization", fmt.Sprintf("%x", sha256.Sum256(pubPEM)))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
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
