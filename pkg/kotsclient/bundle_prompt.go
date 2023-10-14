package kotsclient

import (
	"bufio"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type BundlePromptRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

func (c *VendorV3Client) BundlePrompt(bundleID string, model string, prompt string, responseCh chan string) error {
	defer func() {
		close(responseCh)
	}()

	var bundlePromptRequest = BundlePromptRequest{
		Model:  model,
		Prompt: prompt,
	}

	// stream the response back to the client

	req, err := c.BuildRequest("POST", fmt.Sprintf("/v3/ai/bundle/%s/prompt", bundleID), bundlePromptRequest)
	if err != nil {
		return errors.Wrap(err, "bundle prompt api request")
	}

	// add the streaming header
	req.Header.Set("Accept", "text/event-stream")

	// make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "bundle prompt api response")
	}

	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for {
		// Continuously read from the stream
		ok := scanner.Scan()
		if !ok {
			break
		}

		line := scanner.Text()
		if len(line) > 0 {
			if line[:5] == "data:" {
				responseCh <- line[5:]
			}
		}
	}

	return nil
}
