package enterpriseclient

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func Test_sigAndFingerprint(t *testing.T) {
	req := require.New(t)

	var testData string // just needs to be of sufficient length
	for i := 0; i < 100; i++ {
		testData = testData + "abcdefghijklmnopqurstuvwxyz123456789"
	}

	// make a new private key
	privateKey, err := generatePrivateKey()
	req.NoError(err)

	// encode that private key to bytes
	privateKeyBytes := encodePrivateKeyToPEM(privateKey)
	// decode private key bytes
	privateKey, err = decodePrivateKeyPEM(privateKeyBytes)
	req.NoError(err)

	_, sig, fingerprint, err := sigAndFingerprint(privateKey, []byte(testData))
	req.NoError(err)

	pubKey := encodePublicKey(&privateKey.PublicKey)

	// everything past this depends only on testData, sig, fingerprint, pubKey and req
	// NOT privateKey

	parsedPubKey, err := ssh.ParsePublicKey(pubKey)
	req.NoError(err)

	validated, _, err := ValidatePayload(parsedPubKey, sig, "", []byte(testData))
	req.NoError(err)
	req.True(validated)

	req.Equal(fingerprint, ssh.FingerprintSHA256(parsedPubKey))

	// test bad signatures/data

	fakeSig, _, err := ValidatePayload(parsedPubKey, base64.StdEncoding.EncodeToString([]byte("fakesig")), "", []byte(testData))
	req.Error(err)
	req.False(fakeSig)

	editData, _, err := ValidatePayload(parsedPubKey, sig, "", append([]byte(testData), []byte("malicious")...))
	req.Error(err)
	req.False(editData)
}

func Test_sigblockAndFingerprint(t *testing.T) {
	req := require.New(t)

	var testData string // just needs to be of sufficient length
	for i := 0; i < 100; i++ {
		testData = testData + "abcdefghijklmnopqurstuvwxyz123456789"
	}

	// make a new private key
	privateKey, err := generatePrivateKey()
	req.NoError(err)

	// encode that private key to bytes
	privateKeyBytes := encodePrivateKeyToPEM(privateKey)
	// decode private key bytes
	privateKey, err = decodePrivateKeyPEM(privateKeyBytes)
	req.NoError(err)

	sigBlock, _, fingerprint, err := sigAndFingerprint(privateKey, []byte(testData))
	req.NoError(err)

	pubKey := encodePublicKey(&privateKey.PublicKey)

	// everything past this depends only on testData, sig, fingerprint, pubKey and req
	// NOT privateKey

	parsedPubKey, err := ssh.ParsePublicKey(pubKey)
	req.NoError(err)

	validated, nonce, err := ValidatePayload(parsedPubKey, "", sigBlock, []byte(testData))
	req.NoError(err)
	req.True(validated)
	req.NotNil(nonce)

	req.Equal(fingerprint, ssh.FingerprintSHA256(parsedPubKey))

	// test bad signatures/data

	badSig := SigBlock{
		Timestamp: []byte(time.Now().Format(time.RFC3339Nano)),
		Nonce:     nonce,
		Signature: []byte("abcdefg"),
	}
	badSigBytes, err := json.Marshal(badSig)
	req.NoError(err)

	fakeSig, nonce, err := ValidatePayload(parsedPubKey, "", base64.StdEncoding.EncodeToString(badSigBytes), []byte(testData))
	req.Error(err)
	req.False(fakeSig)
	req.Nil(nonce)

	editData, nonce, err := ValidatePayload(parsedPubKey, "", sigBlock, append([]byte(testData), []byte("malicious")...))
	req.Error(err)
	req.False(editData)
	req.Nil(nonce)
}
