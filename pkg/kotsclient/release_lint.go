package kotsclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/replicatedhq/replicated/pkg/version"
)

var linterOrigin = "https://lint.replicated.com"

func init() {
	if v := os.Getenv("LINTER_API_ORIGIN"); v != "" {
		linterOrigin = v
	}
}

// this is part of the gql client with plans to rename gql client to kotsclient
// and have endpoints for multiple release services included
func (c *VendorV3Client) LintRelease(data []byte, isBuildersRelease bool, contentType string) ([]types.LintMessage, error) {
	url := fmt.Sprintf("%s/v1/lint", linterOrigin)
	if isBuildersRelease {
		url = fmt.Sprintf("%s/v1/builders-lint", linterOrigin)
	}

	reader := bytes.NewReader(data)
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("User-Agent", fmt.Sprintf("Replicated/%s", version.Version()))
	req.Header.Set("Content-Type", contentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Wrap(err, "non OK response from linter")
	}

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response")
	}

	type LintResponse struct {
		LintExpressions []types.LintMessage `json:"lintExpressions"`
	}
	br := LintResponse{}

	if err := json.Unmarshal(msg, &br); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response")
	}

	return br.LintExpressions, nil
}
