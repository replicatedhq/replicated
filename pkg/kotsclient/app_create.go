package kotsclient

import (
	"net/http"

	"github.com/replicatedhq/replicated/pkg/graphql"
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
	err := c.DoJSON("POST", "/v3/app", http.StatusCreated, reqBody, &app)
	if err != nil {
		return nil, err
	}
	return app.App, nil
}

const deleteAppMutation = `
mutation deleteKotsApplication($appId: ID!, $password: String) {
  deleteKotsApplication(appId: $appId, password: $password)
}
`

func (c *GraphQLClient) DeleteKOTSApp(id string) error {
	response := graphql.ResponseErrorOnly{}

	request := graphql.Request{
		Query: deleteAppMutation,
		Variables: map[string]interface{}{
			"appId": id,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return err
	}

	return nil
}
