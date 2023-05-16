package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type CustomerListWithInstancesResponse struct {
	Customers      []types.Customer `json:"customers"`
	TotalCustomers int              `json:"totalCustomers"`
}

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
		resp := CustomerListWithInstancesResponse{}
		path := fmt.Sprintf("/v3/app/%s/customers?currentPage=%d", appID, page)
		err := c.DoJSON("GET", path, http.StatusOK, nil, &resp)
		if err != nil {
			return nil, errors.Wrapf(err, "List Customers By App Version page %d", page)
		}

		for customersIndex, customer := range resp.Customers {
			fmt.Println("-----------------------------")
			fmt.Printf("DEBUG: customer.Name:    %v\n", customer.Name)
			fmt.Printf("DEBUG: customer.ID:      %v\n", customer.ID)
			fmt.Printf("DEBUG: customersIndex:   %v\n", customersIndex)
			for _, instance := range customer.Instances {
				fmt.Printf("      DEBUG: instance.InstanceId: %v\n", instance.InstanceId)
				for versionIndex, versionHistory := range instance.VersionHistory {
					fmt.Println("      DEBUG: versionHistory: ", versionHistory)
					if versionHistory.VersionLabel == appVersion && versionIndex == 0 {
						fmt.Printf("      DEBUG: versionHistory.VersionLabel: %v MATCH\n", versionHistory.VersionLabel)
						// Append the customer to the matchingCustomers slice
						// matchingCustomers = append(matchingCustomers, customer)
						// Apppend the customer and
						// the instance to the matchingCustomers slice
						// matchingCustomers = append(matchingCustomers, customer, instance)
						// Append the customer and
						// the instance and
						// the versionHistory to the matchingCustomers slice
						tempCustomer := customer
						tempCustomer.Instances = []types.Instance{instance}
						tempCustomer.Instances[0].VersionHistory = []types.VersionHistory{versionHistory}
						matchingCustomers = append(matchingCustomers, tempCustomer)
						//matchingCustomers = append(matchingCustomers, customer, instance, versionHistory)

					} else {
						fmt.Printf("      DEBUG: versionHistory.VersionLabel: %v\n", versionHistory.VersionLabel)
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
	if len(matchingCustomers) == 0 {
		return matchingCustomers, fmt.Errorf("no matching customers found for app %q and version %q ", appID, appVersion)
	}
	return matchingCustomers, nil
}
