package test

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"os"

	"github.com/replicatedhq/replicated/client"
)

var api = client.New(os.Getenv("REPLICATED_API_ORIGIN"), os.Getenv("REPLICATED_API_TOKEN"))

func mustToken(n int) string {
	if n == 0 {
		n = 256
	}
	data := make([]byte, int(n))
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		log.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}
