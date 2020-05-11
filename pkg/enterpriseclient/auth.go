package enterpriseclient

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

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

		fmt.Printf("\nYour authentication request has been submitted. Please contact your organization or Replicated at support@replicated.com to complete this request with the following code: %s\n\n", authInitResponse.Code)
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

type SigBlock struct {
	Nonce     []byte
	Timestamp []byte
	Signature []byte
}

// gets the (base64 encoded) signature and fingerprint for a given key and data
// returns the signature with a nonce and timestamp, the signature of the data alone, the fingerprint, and error
func sigAndFingerprint(privateKey *ecdsa.PrivateKey, data []byte) (string, string, string, error) {
	// generate a timestamp and nonce
	nonce := make([]byte, 512/8)        // 512 bits
	_, _ = rand.Read(nonce)             // returns len, nil - neither of which we need
	ts, err := time.Now().MarshalText() // marshals to RFC3339Nano
	if err != nil {
		return "", "", "", errors.Wrap(err, "failed to get timestamp")
	}

	// hash the body and sign the hash
	// store the signature in an ecSig struct and marshal it with the ssh wire format
	contentSha := sha512.Sum512(data)
	var ecSig struct {
		R *big.Int
		S *big.Int
	}
	ecSig.R, ecSig.S, err = ecdsa.Sign(rand.Reader, privateKey, contentSha[:])
	if err != nil {
		return "", "", "", errors.Wrap(err, "failed to sign content sha")
	}
	signatureString := base64.StdEncoding.EncodeToString(ssh.Marshal(ecSig))

	// do the same for the data combined with the timestamp and nonce
	contentSha = sha512.Sum512(combineTsNonceData(ts, nonce, data))
	ecSig.R, ecSig.S, err = ecdsa.Sign(rand.Reader, privateKey, contentSha[:])
	if err != nil {
		return "", "", "", errors.Wrap(err, "failed to sign content sha")
	}
	tsSig := ssh.Marshal(ecSig)

	// include the public key fingerprint as a hint to the server
	fingerprint, err := getFingerprint(&privateKey.PublicKey)
	if err != nil {
		return "", "", "", errors.Wrap(err, "failed to get public key fingerprint")
	}

	sigBlock := SigBlock{
		Timestamp: ts,
		Nonce:     nonce,
		Signature: tsSig,
	}
	sigBlockBytes, err := json.Marshal(sigBlock)
	if err != nil {
		return "", "", "", errors.Wrap(err, "failed to marshal signature block")
	}
	sigBlockString := base64.StdEncoding.EncodeToString(sigBlockBytes)

	return sigBlockString, signatureString, fingerprint, nil
}

func combineTsNonceData(ts, nonce, data []byte) []byte {
	tsBytes := []byte{}
	tsBytes = append(tsBytes, nonce...)
	tsBytes = append(tsBytes, ts...)
	tsBytes = append(tsBytes, data...)
	return tsBytes
}

// ValidatePayload checks that the payload was signed by the provided public key
// if fingerprintSigString is not empty, sigString is ignored and it is used instead, and the nonce will be returned if valid
func ValidatePayload(pubkey ssh.PublicKey, sigString, fingerprintSigString string, data []byte) (bool, []byte, error) {
	if !strings.HasPrefix(pubkey.Type(), "ecdsa-sha2-") {
		return false, nil, fmt.Errorf("%q is not an accepted public key type", pubkey.Type())
	}

	if fingerprintSigString != "" {
		sigBlockBytes, err := base64.StdEncoding.DecodeString(fingerprintSigString)
		if err != nil {
			return false, nil, errors.Wrap(err, "invalid signature string")
		}

		decodedSig := SigBlock{}
		err = json.Unmarshal(sigBlockBytes, &decodedSig)
		if err != nil {
			return false, nil, errors.Wrap(err, "invalid signature object")
		}

		err = pubkey.Verify(combineTsNonceData(decodedSig.Timestamp, decodedSig.Nonce, data), &ssh.Signature{
			Format: pubkey.Type(),
			Blob:   decodedSig.Signature,
		})
		if err != nil {
			return false, nil, errors.Wrap(err, "invalid signature")
		}

		sentTime, err := time.Parse(time.RFC3339Nano, string(decodedSig.Timestamp))
		if err != nil {
			return false, nil, errors.Wrap(err, "invalid timestamp")
		}
		if !sentTime.Before(time.Now().Add(time.Hour)) {
			return false, nil, fmt.Errorf("date %s is more than an hour after the current time %s", string(decodedSig.Timestamp), time.Now().Format(time.RFC3339Nano))
		}
		if !sentTime.After(time.Now().Add(-time.Hour)) {
			return false, nil, fmt.Errorf("date %s is more than an hour before the current time %s", string(decodedSig.Timestamp), time.Now().Format(time.RFC3339Nano))
		}
		return true, decodedSig.Nonce, nil
	}

	sigBytes, err := base64.StdEncoding.DecodeString(sigString)
	if err != nil {
		return false, nil, errors.Wrap(err, "invalid signature string")
	}

	err = pubkey.Verify(data, &ssh.Signature{
		Format: pubkey.Type(),
		Blob:   sigBytes,
	})
	if err != nil {
		return false, nil, errors.Wrap(err, "invalid signature")
	}

	return true, nil, nil
}
