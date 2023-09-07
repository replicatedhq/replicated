package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

var _ error = ErrCustomerNotFound{}

type ErrCustomerNotFound struct {
	Name string
}

func (e ErrCustomerNotFound) Error() string {
	return fmt.Sprintf("customer %q not found", e.Name)
}

type CustomerListResponse struct {
	Customers      []types.Customer `json:"customers"`
	TotalCustomers int              `json:"totalCustomers"`
}

func (c *VendorV3Client) ListCustomers(appID string, includeTest bool) ([]types.Customer, error) {
	allCustomers := []types.Customer{}
	page := 0
	for {
		resp := CustomerListResponse{}
		path := fmt.Sprintf("/v3/app/%s/customers?currentPage=%d", appID, page)
		if includeTest {
			path = fmt.Sprintf("%s&includeTest=true", path)
		}
		err := c.DoJSON("GET", path, http.StatusOK, nil, &resp)
		if err != nil {
			return nil, errors.Wrapf(err, "list customers page %d", page)
		}

		allCustomers = append(allCustomers, resp.Customers...)

		if len(allCustomers) == resp.TotalCustomers || len(resp.Customers) == 0 {
			break
		}

		page = page + 1
	}

	return allCustomers, nil
}

func (c *VendorV3Client) GetCustomerByNameOrId(appID string, nameOrId string) (*types.Customer, error) {
	allCustomers, err := c.ListCustomers(appID, false)
	if err != nil {
		return nil, err
	}

	matchingCustomers := make([]types.Customer, 0)
	for _, customer := range allCustomers {
		if customer.ID == nameOrId || customer.Name == nameOrId {
			matchingCustomers = append(matchingCustomers, customer)
		}
	}

	if len(matchingCustomers) == 0 {
		return nil, ErrCustomerNotFound{Name: nameOrId}
	}

	if len(matchingCustomers) > 1 {
		return nil, fmt.Errorf("customer %q is ambiguous, please use customer ID", nameOrId)
	}
	return &matchingCustomers[0], nil
}
