package kotsclient

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type LoadBundleRequest struct {
	Model   string `json:"model"`
	Content string `json:"content"`
}

type LoadBundleResponse struct {
	Bundle *types.Bundle `json:"bundle"`
}

func (c *VendorV3Client) LoadBundle(model string, content *bytes.Buffer) (*types.Bundle, error) {
	// base64 encoded the content
	encoded := base64.StdEncoding.EncodeToString(content.Bytes())
	req := LoadBundleRequest{
		Model:   model,
		Content: encoded,
	}

	resp := LoadBundleResponse{}

	err := c.DoJSON("POST", "/v3/ai/bundle", http.StatusCreated, req, &resp)
	if err != nil {
		if strings.HasSuffix(strings.TrimSpace(err.Error()), "403:") {
			return nil, ErrAINotEntitled
		}
		return nil, errors.Wrap(err, "load bundle")
	}

	fmt.Printf("resp.Bundle: %#v\n", resp.Bundle)
	return resp.Bundle, nil
}
