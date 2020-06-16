package enterpriseclient

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewNilHTTPClient(t *testing.T) {
	req := require.New(t)
	client := NewHTTPClient("origin", nil)
	req.Equal(&HTTPClient{
		privateKey: nil,
		apiOrigin:  "origin",
	}, client)
}
