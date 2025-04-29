package util

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
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

	_, _, _, _, err = ssh.ParseAuthorizedKey([]byte(keyContent))
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("invalid SSH public key format in file: %s", path))
	}

	encodedKey := base64.StdEncoding.EncodeToString([]byte(keyContent))
	return encodedKey, nil
}
