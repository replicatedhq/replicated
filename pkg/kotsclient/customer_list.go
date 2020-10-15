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

func (c *VendorV3Client) ListCustomers(appID string) ([]types.Customer, error) {
	allCustomers := []types.Customer{}
	resp := CustomerListResponse{}
	page := 0
	for len(allCustomers) < resp.TotalCustomers || page == 0 {
		page = page + 1
		path := fmt.Sprintf("/v3/app/%s/customers?currentPage=%d", appID, page)
		err := c.DoJSON("GET", path, http.StatusOK, nil, &resp)
		if err != nil {
			return nil, errors.Wrapf(err, "list customers page %d", page)
		}
		allCustomers = append(allCustomers, resp.Customers...)
	}

	return allCustomers, nil
}

func (c *VendorV3Client) GetCustomerByName(appID string, name string) (*types.Customer, error) {
	allCustomers, err := c.ListCustomers(appID)
	if err != nil {
		return nil, err
	}

	matchingCustomers := make([]types.Customer, 0)
	for _, customer := range allCustomers {
		if customer.ID == name || customer.Name == name {
			matchingCustomers = append(matchingCustomers, customer)
		}
	}

	if len(matchingCustomers) == 0 {
		return nil, ErrCustomerNotFound{Name: name}
	}

	if len(matchingCustomers) > 1 {
		return nil, fmt.Errorf("customer %q is ambiguous, please use customer ID", name)
	}
	return &matchingCustomers[0], nil
}
