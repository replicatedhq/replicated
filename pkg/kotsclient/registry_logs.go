package kotsclient

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/replicatedhq/replicated/pkg/types"
)

type KOTSRegistryLogsResponse struct {
	RegistryLogs []types.RegistryLog `json:"logs"`
}

func (c *VendorV3Client) LogsRegistry(hostname string) ([]types.RegistryLog, error) {
	resp := KOTSRegistryLogsResponse{}
	err := c.DoJSON("GET", fmt.Sprintf(`/v3/external_registry/logs?endpoint=%s`, url.QueryEscape(hostname)), http.StatusOK, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.RegistryLogs, nil
}
