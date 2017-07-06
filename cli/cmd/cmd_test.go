package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"testing"
)

// generate a random string or call t.Fatal
func token(t *testing.T, n int) string {
	if n == 0 {
		n = 256
	}
	data := make([]byte, int(n))
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}
