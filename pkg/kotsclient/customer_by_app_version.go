package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

// var _ error = ErrAppVersionNotFound{}
//
// type ErrAppVersionNotFound struct {
//Name string
//}

// func (e ErrAppVersionNotFound) Error() string {
//return fmt.Sprintf("version %q not found", e.Name)
//}

//type CustomerListResponse struct {
//Customers      []types.Customer `json:"customers"`
//TotalCustomers int              `json:"totalCustomers"`
//}

func (c *VendorV3Client) ListCustomersByAppVersion(appID string, appVersion string, appType string) ([]types.Customer, error) {
	println("--------------DEBUG-------------------------")
	println("DEBUG: kotsclient:ListCustomersByAppVersion")
	println("DEBUG: appID:      " + appID)
	println("DEBUG: appVersion: " + appVersion)
	println("DEBUG: appType:    " + appType)
	println("\n")

	matchingCustomers := []types.Customer{}
	page := 0
	for {
		resp := CustomerListResponse{}
		path := fmt.Sprintf("/v3/app/%s/customers?currentPage=%d", appID, page)
		err := c.DoJSON("GET", path, http.StatusOK, nil, &resp)
		if err != nil {
			return nil, errors.Wrapf(err, "list customers page %d", page)
		}

		for index, customer := range resp.Customers {
			fmt.Printf("DEBUG: customer.Name: %v\n", customer.Name)
			// for each customer, print the instances
			for _, instance := range customer.Instances {
				fmt.Printf("         instance.InstanceId: %v\n", instance.InstanceId)
				// for each instance print the VersionHistory VersionLabel
				for _, versionHistory := range instance.VersionHistory {
					if versionHistory.VersionLabel == appVersion {
						fmt.Printf("           versionHistory.VersionLabel: %v MATCH\n", versionHistory.VersionLabel)
						matchingCustomers = append(matchingCustomers, resp.Customers[index])
					} else {
						fmt.Printf("           versionHistory.VersionLabel: %v\n", versionHistory.VersionLabel)
					}
				}
			}

		}

		if len(matchingCustomers) == resp.TotalCustomers || len(resp.Customers) == 0 {
			break
		}

		page = page + 1
	}

	println("--------------DEBUG-------------------------\n")
	return matchingCustomers, nil
}
