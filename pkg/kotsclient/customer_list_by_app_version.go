package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type CustomerListWithInstancesResponse struct {
	// Response from /v3/app/{appID}/customers
	Customers      []types.Customer `json:"customers"`
	TotalCustomers int              `json:"totalCustomers"`
}

func (c *VendorV3Client) ListCustomersByAppAndVersion(appID string, appVersion string, appType string) ([]types.Customer, error) {

	matchingCustomers := []types.Customer{}
	page := 0
	for {
		resp := CustomerListWithInstancesResponse{}
		path := fmt.Sprintf("/v3/app/%s/customers?currentPage=%d", appID, page)
		err := c.DoJSON("GET", path, http.StatusOK, nil, &resp)
		if err != nil {
			return nil, errors.Wrapf(err, "List Customers By App Version page %d", page)
		}

		for _, customer := range resp.Customers {
			for _, instance := range customer.Instances {
				for versionIndex, versionHistory := range instance.VersionHistory {
					if versionHistory.VersionLabel == appVersion && versionIndex == 0 {
						// only add the customer if the version matches
						// index[0] is the latest version
						// There has to be a better way to do this without creating a tempCustomer
						tempCustomer := customer
						tempCustomer.Instances = []types.Instance{instance}
						tempCustomer.Instances[0].VersionHistory = []types.VersionHistory{versionHistory}
						matchingCustomers = append(matchingCustomers, tempCustomer)

					}
				}
			}
		}
		if len(matchingCustomers) == resp.TotalCustomers || len(resp.Customers) == 0 {
			// We've reached the end of the customer pages
			break
		}
		page = page + 1
	}

	if len(matchingCustomers) == 0 {
		return matchingCustomers, fmt.Errorf("no matching customers found for app %q and version %q ", appID, appVersion)
	}
	return matchingCustomers, nil
}
