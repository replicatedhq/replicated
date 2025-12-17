package util

import (
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
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

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(keyContent))
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("invalid SSH public key format in file: %s", path))
	}

	parts := strings.Fields(keyContent)
	if len(parts) > 2 {
		comment := strings.Join(parts[2:], " ")

		dangerousPattern := regexp.MustCompile(`[;&|<>$\\\(\)\[\]\{\}\*\?]`)
		if dangerousPattern.MatchString(comment) {
			return "", fmt.Errorf("SSH public key comment contains potentially dangerous characters: %s", path)
		}
	}

	keyType := pubKey.Type()
	if keyType != "ssh-rsa" && keyType != "ssh-ed25519" &&
		!strings.HasPrefix(keyType, "ecdsa-sha2-nistp") && keyType != "ssh-dss" {
		return "", fmt.Errorf("unsupported SSH public key type %s in file: %s", keyType, path)
	}

	encodedKey := base64.StdEncoding.EncodeToString([]byte(keyContent))
	return encodedKey, nil
}
