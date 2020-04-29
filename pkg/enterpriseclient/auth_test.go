package enterpriseclient

import (
	"encoding/base64"
	"strings"
	"testing"

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

	sig, fingerprint, err := sigAndFingerprint(privateKey, []byte(testData))
	req.NoError(err)

	pubKey := encodePublicKey(&privateKey.PublicKey)

	// everything past this depends only on testData, sig, fingerprint, pubKey and req
	// NOT privateKey

	sigBytes, err := base64.StdEncoding.DecodeString(sig)
	req.NoError(err)

	parsedPubKey, err := ssh.ParsePublicKey(pubKey)
	req.NoError(err)

	req.True(strings.HasPrefix(parsedPubKey.Type(), "ecdsa-sha2-"), parsedPubKey.Type())

	err = parsedPubKey.Verify([]byte(testData), &ssh.Signature{
		Format: parsedPubKey.Type(),
		Blob:   sigBytes,
	})
	req.NoError(err)

	req.Equal(fingerprint, ssh.FingerprintSHA256(parsedPubKey))
}
