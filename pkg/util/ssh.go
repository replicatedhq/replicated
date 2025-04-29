package util

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func ReadAndValidatePublicKey(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", errors.Wrap(err, "read public key file")
	}

	keyContent := strings.TrimSpace(string(content))
	if keyContent == "" {
		return "", fmt.Errorf("public key file is empty: %s", path)
	}

	if !strings.HasPrefix(keyContent, "ssh-rsa") &&
		!strings.HasPrefix(keyContent, "ssh-ed25519") &&
		!strings.HasPrefix(keyContent, "ecdsa-sha2-nistp") &&
		!strings.HasPrefix(keyContent, "ssh-dss") {
		return "", fmt.Errorf("file does not appear to be a public key (must start with ssh-rsa, ssh-ed25519, ecdsa-sha2-nistp*, or ssh-dss): %s", path)
	}

	parts := strings.Fields(keyContent)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid public key format (should have at least type and key data): %s", path)
	}

	encodedKey := base64.StdEncoding.EncodeToString([]byte(keyContent))
	return encodedKey, nil
}
