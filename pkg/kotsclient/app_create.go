package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	kotsclienttypes "github.com/replicatedhq/replicated/pkg/kotsclient/types"
	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateKOTSAppRequest struct {
	Name string `json:"name"`
}

func (c *VendorV3Client) CreateKOTSApp(ctx context.Context, name string) (*types.KotsAppWithChannels, error) {
	reqBody := &CreateKOTSAppRequest{
		Name: name,
	}
	app := kotsclienttypes.CreateKOTSAppResponse{}
	err := c.DoJSON(ctx, "POST", "/v3/app", http.StatusCreated, reqBody, &app)
	if err != nil {
		return nil, err
	}
	return app.App, nil
}

func (c *VendorV3Client) DeleteKOTSApp(ctx context.Context, id string) error {
	url := fmt.Sprintf("/v3/app/%s", id)

	err := c.DoJSON(ctx, "DELETE", url, http.StatusOK, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
