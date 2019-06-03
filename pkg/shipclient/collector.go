package shipclient

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/replicatedhq/replicated/pkg/util"
)

type GraphQLResponseListCollectors struct {
	Data   *ShipCollectorsData `json:"data,omitempty"`
	Errors []GraphQLError      `json:"errors,omitempty"`
}

type ShipCollectorsData struct {
	ShipCollectors []*ShipCollector `json:"allCollectors"`
}

type ShipCollector struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created"`
}

type GraphQLResponseFinalizeCollector struct {
	Data   *ShipFinalizeCreateCollectorData `json:"data,omitempty"`
	Errors []GraphQLError                   `json:"errors,omitempty"`
}

type ShipFinalizeCreateCollectorData struct {
	ShipCollector *ShipCollector `json:"finalizeUploadedCollector"`
}

type GraphQLResponseUploadCollector struct {
	Data   ShipCollectorUploadData `json:"data,omitempty"`
	Errors []GraphQLError          `json:"errors,omitempty"`
}

type ShipCollectorUploadData struct {
	ShipPendingCollectorData *ShipPendingCollectorData `json:"uploadCollector"`
}

type ShipPendingCollectorData struct {
	UploadURI string `json:"uploadUri"`
	UploadID  string `json:"id"`
}

func (c *GraphQLClient) CreateCollector(appID string, name string, yaml string) (*types.CollectorInfo, error) {
	response := GraphQLResponseUploadCollector{}

	request := GraphQLRequest{
		Query: `
mutation uploadCollector($appId: ID!) {
  uploadCollector(appId: $appId) {
    id
    uploadUri
  }
}`,
		Variables: map[string]interface{}{
			"appId": appID,
		},
	}

	if err := c.executeRequest(request, &response); err != nil {
		return nil, err
	}

	tmpFile, err := ioutil.TempFile(os.TempDir(), "replicated-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yaml)); err != nil {
		return nil, err
	}
	tmpFile.Close()

	if err := util.UploadFile(tmpFile.Name(), response.Data.ShipPendingCollectorData.UploadURI); err != nil {
		return nil, err
	}

	request = GraphQLRequest{
		Query: `
mutation finalizeUploadedCollector($appId: ID! $uploadId: String) {
  finalizeUploadedRelease(appId: $appId, uploadId: $uploadId) {
    id
    name
    spec
    created
  }
}`,
		Variables: map[string]interface{}{
			"appId":    appID,
			"uploadId": response.Data.ShipPendingCollectorData.UploadID,
		},
	}

	// call finalize release
	finalizeCollectorResponse := GraphQLResponseFinalizeCollector{}

	if err := c.executeRequest(request, &finalizeCollectorResponse); err != nil {
		return nil, err
	}

	location, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}

	createdAt, err := util.ParseTime(finalizeCollectorResponse.Data.ShipCollector.CreatedAt)
	if err != nil {
		return nil, err
	}

	collectorInfo := types.CollectorInfo{
		AppID:     appID,
		CreatedAt: createdAt.In(location),
		Name:      name,
	}

	return &collectorInfo, nil
}

func (c *GraphQLClient) UpdateCollector(appID string, name string, yaml string) error {
	return nil
}

func (c *GraphQLClient) PromoteCollector(appID string, name string, channelIDs ...string) error {
	response := GraphQLResponseErrorOnly{}

	request := GraphQLRequest{
		Query: `
mutation promoteShipCollector($appId: ID!, $name: String!, $channelIds: [String]) {
  promoteShipCollector(appId: $appId, name: $name, channelIds: $channelIds) {
    id
  }
}`,
		Variables: map[string]interface{}{
			"appId":      appID,
			"name":       name,
			"channelIds": channelIDs,
		},
	}

	if err := c.executeRequest(request, &response); err != nil {
		return err
	}

	if len(response.Errors) != 0 {
		return errors.New(response.Errors[0].Message)
	}

	return nil
}

func (c *GraphQLClient) ListCollectors(appID string) ([]types.CollectorInfo, error) {
	response := GraphQLResponseListCollectors{}

	request := GraphQLRequest{
		Query: `
query allCollectors($appId: ID!) {
  allReleases(appId: $appId) {
    id
    name
    spec
    created
    channels {
      id
      name
      currentVersion
      numReleases
    }
  }
}`,
		Variables: map[string]interface{}{
			"appId": appID,
		},
	}

	if err := c.executeRequest(request, &response); err != nil {
		return nil, err
	}

	location, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}

	collectorInfos := make([]types.CollectorInfo, 0, 0)
	for _, shipCollector := range response.Data.ShipCollectors {
		createdAt, err := util.ParseTime(shipCollector.CreatedAt)
		if err != nil {
			return nil, err
		}
		collectorInfo := types.CollectorInfo{
			AppID:     appID,
			CreatedAt: createdAt.In(location),
			Name:      shipCollector.Name,
		}

		collectorInfos = append(collectorInfos, collectorInfo)
	}

	return collectorInfos, nil
}
