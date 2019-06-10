package shipclient

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

type GraphQLResponseListCollectors struct {
	Data   *SupportBundleSpecsData `json:"data,omitempty"`
	Errors []graphql.GQLError      `json:"errors,omitempty"`
}

type SupportBundleSpecsData struct {
	SupportBundleSpecs []SupportBundleSpec `json:"supportBundleSpecs"`
}

type GraphQLResponseUpdateCollector struct {
	Data *SupportBundleUpdateSpecData `json:"data,omitempty"`
	// Errors []graphql.GQLError           `json:"errors,omitempty"`
}

type SupportBundleUpdateSpecData struct {
	UpdateSupportBundleSpec *UpdateSupportBundleSpec `json:"updateSupportBundleSpec"`
}

type UpdateSupportBundleSpec struct {
	ID     string `json:"id"`
	Config string `json:"spec,omitempty"`
}

type SupportBundleSpec struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	CreatedAt string          `json:"createdAt"`
	Channels  []types.Channel `json:"channels"`
	Config    string          `json:"spec,omitempty"`
}

type GraphQLResponseCreateCollector struct {
	Data   *SupportBundleSpecFinalizeCreateSpecData `json:"data,omitempty"`
	Errors []graphql.GQLError                       `json:"errors,omitempty"`
}

type SupportBundleSpecFinalizeCreateSpecData struct {
	SupportBundleSpec *SupportBundleSpec `json:"finalizeUploadedCollector"`
}

type GraphQLResponseUploadCollector struct {
	Data   SupportBundleSpecUploadData `json:"data,omitempty"`
	Errors []graphql.GQLError          `json:"errors,omitempty"`
}

type SupportBundleSpecUploadData struct {
	SupportBundleSpecPendingSpecData *SupportBundleSpecPendingSpecData `json:"uploadCollector"`
}

type SupportBundleSpecPendingSpecData struct {
	UploadURI string `json:"uploadUri"`
	UploadID  string `json:"id"`
}

func (c *GraphQLClient) CreateCollector(appID string, name string, yaml string) (*types.CollectorInfo, error) {
	response := GraphQLResponseCreateCollector{}

	request := graphql.Request{
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

	if err := c.ExecuteRequest(request, &response); err != nil {
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

	// if err := util.UploadFile(tmpFile.Name(), response.Data.SupportBundleSpec); err != nil {
	// 	return nil, err
	// }

	request = graphql.Request{
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
			"uploadId": response.Data.SupportBundleSpec,
		},
	}

	// call finalize release
	finalizeCollectorResponse := GraphQLResponseListApps{}

	if err := c.ExecuteRequest(request, &finalizeCollectorResponse); err != nil {
		return nil, err
	}

	collectorInfo := types.CollectorInfo{
		AppID: appID,
		Name:  name,
	}

	return &collectorInfo, nil
}

func (c *GraphQLClient) UpdateCollector(appID string, specID, yaml string) error {
	// response := GraphQLResponseUpdateCollector{}
	response := graphql.ResponseErrorOnly{}

	request := graphql.Request{
		Query: `
		mutation updateSupportBundleSpec($id: ID!, $spec: String!, $githubRef: GitHubRefInput, $isArchived: Boolean) {
			updateSupportBundleSpec(id: $id, spec: $spec, githubRef: $githubRef, isArchived: $isArchived) {
				id
				spec
				createdAt
				updatedAt
				isArchived
				githubRef {
					owner
					repoFullName
					branch
					path
				}
			}
		}
	`,

		Variables: map[string]interface{}{
			// "githubRef":  null,
			"id": specID,
			// "isArchived": null,
			"spec": yaml,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return err
	}

	return nil
}

func (c *GraphQLClient) ListCollectors(appID string) ([]types.CollectorInfo, error) {
	response := GraphQLResponseListCollectors{}

	request := graphql.Request{
		Query: `
		query supportBundleSpecs($appId: String) {
			supportBundleSpecs(appId: $appId) {
			  id
			  name
			  spec
			  createdAt
			  updatedAt
			  isArchived
			  isImmutable
			  githubRef {
				owner
				repoFullName
				branch
				path
			  }
			  channels {
				id
				name
			  }
			  platformChannels {
				id
				name
			  }
			}
		  }`,
		Variables: map[string]interface{}{
			"appId": appID,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	// location, err := time.LoadLocation("Local")
	// if err != nil {
	// 	return nil, err
	// }

	collectorInfos := make([]types.CollectorInfo, 0, 0)
	for _, shipCollector := range response.Data.SupportBundleSpecs {
		// createdAt, err := util.ParseTime(shipCollector.CreatedAt)
		// if err != nil {
		// 	return nil, err
		// }
		collectorInfo := types.CollectorInfo{
			AppID: appID,
			// CreatedAt:      createdAt.In(location),
			Name:           shipCollector.Name,
			SpecID:         shipCollector.ID,
			ActiveChannels: shipCollector.Channels,
		}

		collectorInfos = append(collectorInfos, collectorInfo)
	}

	return collectorInfos, nil
}

// GetCollector returns a collector's properties.
func (c *GraphQLClient) GetCollector(appID string, id string) (*types.CollectorInfo, error) {
	allcollectors, err := c.ListCollectors(appID)
	if err != nil {
		return nil, err
	}

	for _, collector := range allcollectors {
		if collector.SpecID == id {
			return &collector, nil
		}
	}

	return nil, errors.New("Not found")
}

// PromoteCollector assigns collector to a specified channel.
func (c *GraphQLClient) PromoteCollector(appID string, specID string, channelIDs ...string) error {
	response := graphql.ResponseErrorOnly{}

	request := graphql.Request{
		Query: `
mutation  promoteTroubleshootSpec($channelIds: [String], $specId: ID!) {
	promoteTroubleshootSpec(channelIds: $channelIds, specId: $specId) {
		id
	}
}`,
		Variables: map[string]interface{}{
			"channelIds": channelIDs,
			"specId":     specID,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return err
	}

	if len(response.Errors) != 0 {
		return errors.New(response.Errors[0].Message)
	}

	return nil
}
