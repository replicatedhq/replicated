package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateKOTSAppRequest struct {
	Name string `json:"name"`
}

type CreateKOTSAppResponse struct {
	App *types.KotsAppWithChannels `json:"app"`
}

func (c *VendorV3Client) CreateKOTSApp(name string) (*types.KotsAppWithChannels, error) {
	reqBody := &CreateKOTSAppRequest{Name: name}
	app := CreateKOTSAppResponse{}
	err := c.DoJSON(context.TODO(), "POST", "/v3/app", http.StatusCreated, reqBody, &app)
	if err != nil {
		return nil, err
	}
	return app.App, nil
}

func (c *VendorV3Client) DeleteKOTSApp(id string) error {
	url := fmt.Sprintf("/v3/app/%s", id)

	err := c.DoJSON(context.TODO(), "DELETE", url, http.StatusOK, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
