package enterpriseclient

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

func (c HTTPClient) AuthInit(organizationName string) error {
	// by default, we store the key in ~/.replicated/enterprise
	_, err := os.Stat(filepath.Join(homeDir(), ".replicated", "enterprise"))
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to check for directory")
	}
	if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Join(homeDir(), ".replicated", "enterprise"), 0755); err != nil {
			return errors.Wrap(err, "failed to mkdir")
		}
	}
	pubKeyPath := filepath.Join(homeDir(), ".replicated", "enterprise", "ecdsa.pub")
	privKeyPath := filepath.Join(homeDir(), ".replicated", "enterprise", "ecdsa")

	_, pubKeyErr := os.Stat(pubKeyPath)
	_, privKeyErr := os.Stat(privKeyPath)

	if pubKeyErr != nil && !os.IsNotExist(pubKeyErr) {
		return errors.Wrap(pubKeyErr, "failed to read public key")
	}
	if privKeyErr != nil && !os.IsNotExist(privKeyErr) {
		return errors.Wrap(privKeyErr, "failed to read private key")
	}

	missingPublicKey := os.IsNotExist(pubKeyErr)
	missingPrivateKey := os.IsNotExist(privKeyErr)

	if !missingPrivateKey && !missingPublicKey {
		return errors.New("already authenticated")
	}

	privateKey, err := generatePrivateKey()
	if err != nil {
		return errors.Wrap(err, "failed to generate private key")
	}
	if err := ioutil.WriteFile(privKeyPath, encodePrivateKeyToPEM(privateKey), 0600); err != nil {
		return errors.Wrap(err, "failed to write private key to file")
	}

	if err := ioutil.WriteFile(pubKeyPath, encodePublicKey(&privateKey.PublicKey), 0600); err != nil {
		return errors.Wrap(err, "failed to write public key to file")
	}

	if organizationName != "" {
		// --create-org flag is provided, create the organization
		// send the PUBLIC key and the organization name to the replicated server and return the organization id
		type CreateOrgRequest struct {
			PublicKeyBytes   []byte `json:"publicKey"`
			OrganizationName string `json:"organizationName"`
		}
		createOrgRequest := CreateOrgRequest{
			PublicKeyBytes:   encodePublicKey(&privateKey.PublicKey),
			OrganizationName: organizationName,
		}

		type CreateOrgResponse struct {
			OrganizationID string `json:"organizationId"`
		}
		createOrgResponse := CreateOrgResponse{}

		err = c.doJSON("POST", "/v1/organization", 201, createOrgRequest, &createOrgResponse)
		if err != nil {
			return errors.Wrap(err, "failed to create organization")
		}

		fmt.Printf("\nOrganization has been created successfully with the following ID: %s\n\n", createOrgResponse.OrganizationID)
	} else {
		// --create-org flag is NOT provided, begin authentication process
		// send the PUBLIC key to the replicated server and return the key id
		type AuthRequest struct {
			PublicKeyBytes []byte `json:"publicKey"`
		}
		authRequest := AuthRequest{
			PublicKeyBytes: encodePublicKey(&privateKey.PublicKey),
		}

		type AuthInitResponse struct {
			Code string `json:"code"`
		}
		authInitResponse := AuthInitResponse{}

		err = c.doJSON("POST", "/v1/auth", 201, authRequest, &authInitResponse)
		if err != nil {
			return errors.Wrap(err, "failed to init auth with server")
		}

		fmt.Printf("\nYour authentication request has been submitted. Please contact Replicated at support@replicated.com to complete this request with the following code: %s\n\n", authInitResponse.Code)
	}

	return nil
}

func (c HTTPClient) AuthApprove(fingerprint string) error {
	type AuthApproveRequest struct {
		Fingerprint string `json:"fingerprint"`
	}
	authApproveRequest := AuthApproveRequest{
		Fingerprint: fingerprint,
	}

	err := c.doJSON("PUT", "/v1/auth/approve", 204, authApproveRequest, nil)
	if err != nil {
		return errors.Wrap(err, "failed to approve auth request")
	}

	fmt.Print("\nAuthentication request approved successfully\n\n")
	return nil
}

func generatePrivateKey() (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate key")
	}

	return privateKey, nil
}

func encodePrivateKeyToPEM(privateKey *ecdsa.PrivateKey) []byte {
	privDER, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		panic(err) // this should never happen - if it does, that means things are rather broken
	}

	privBlock := pem.Block{
		Type:    "EC PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

func decodePrivateKeyPEM(privateBytes []byte) (*ecdsa.PrivateKey, error) {
	privBlock, _ := pem.Decode(privateBytes)

	if privBlock.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("private key type is %s, not 'EC PRIVATE KEY'", privBlock.Type)
	}

	key, err := x509.ParseECPrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "decode ec private key")
	}
	return key, nil
}

func encodePublicKey(publicKey *ecdsa.PublicKey) []byte {
	pubKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		panic(errors.Wrap(err, "create ssh pubkey")) // this should never happen - if it does, that means things are rather broken
	}

	return pubKey.Marshal()
}

func getFingerprint(publicKey *ecdsa.PublicKey) (string, error) {
	pubKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return "", errors.Wrap(err, "create ssh pubkey")
	}
	return ssh.FingerprintSHA256(pubKey), nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
