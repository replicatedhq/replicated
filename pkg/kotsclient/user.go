package kotsclient

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

type GetUserResponse struct {
	GitHubUsername string `json:"gitHubUsername"`
}

func (c *VendorV3Client) GetGitHubUsername() (string, error) {
	var response GetUserResponse

	err := c.DoJSON(context.TODO(), "GET", "/v1/user", http.StatusOK, nil, &response)
	if err != nil {
		return "", errors.Wrap(err, "failed to fetch GitHub username from vendor API")
	}

	return response.GitHubUsername, nil
}
