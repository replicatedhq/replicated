package kotsclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
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
		err := c.DoJSON(context.TODO(), "GET", path, http.StatusOK, nil, &resp)
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
	customer, err := c.GetCustomerByID(nameOrId)
	if err != nil && errors.Cause(err) != platformclient.ErrNotFound {
		return nil, errors.Wrap(err, "get customer by id")
	}

	if customer != nil {
		return customer, nil
	}

	customer, err = c.GetCustomerByName(appID, nameOrId)
	if err != nil {
		return nil, errors.Wrap(err, "get customer by name")
	}

	return customer, nil
}

type CustomerGetResponse struct {
	Customer types.Customer `json:"customer"`
}

func (c *VendorV3Client) GetCustomerByID(customerID string) (*types.Customer, error) {
	resp := CustomerGetResponse{}
	err := c.DoJSON(context.TODO(), "GET", fmt.Sprintf("/v3/customer/%s", url.PathEscape(customerID)), http.StatusOK, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Customer, nil
}

func (c *VendorV3Client) GetCustomerByName(appID string, name string) (*types.Customer, error) {
	// Using the search API, we first to narrow down fuzzy matches to one exact match.
	// Since search API may return stale data, we then also need to use the customer ID to get the exact customer record.
	customers, err := c.listCustomersByName(appID, name)
	if err != nil {
		return nil, err
	}

	if len(customers) == 0 {
		return nil, platformclient.ErrNotFound
	}

	exactMatches := make([]*types.Customer, 0)
	for _, customer := range customers {
		if customer.Name == name {
			exactMatches = append(exactMatches, &customer)
		}
	}

	if len(exactMatches) == 0 {
		return nil, ErrCustomerNotFound{Name: name}
	}

	if len(exactMatches) > 1 {
		return nil, fmt.Errorf("customer %q is ambiguous, please use customer ID", name)
	}

	customer, err := c.GetCustomerByID(exactMatches[0].ID)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// This function will use the search API to find customers by name, which uses fuzzy matching, so it may return multiple customers with similar names.
// In most practical cases, this is still faster than using the /cutomers API to list all customers for the app.
func (c *VendorV3Client) listCustomersByName(appID string, name string) ([]types.Customer, error) {
	if name == "" {
		return nil, errors.New("name is required to search customers")
	}

	allCustomers := []types.Customer{}
	offset := 0
	pageSize := 100
	for {
		payload := struct {
			AppID            string `json:"app_id"`
			Offset           int    `json:"offset"`
			PageSize         int    `json:"page_size"`
			Query            string `json:"query"`
			IncludeActive    bool   `json:"include_active"`
			IncluseArchived  bool   `json:"include_archived"`
			IncludeCommunity bool   `json:"include_community"`
			IncludeDev       bool   `json:"include_dev"`
			IncludeInactive  bool   `json:"include_inactive"`
			IncludePaid      bool   `json:"include_paid"`
			IncludeTrial     bool   `json:"include_trial"`
			InstancePreview  bool   `json:"instance_preview"`
		}{
			AppID:            appID,
			Offset:           offset,
			PageSize:         100,
			Query:            fmt.Sprintf("name:%s", name),
			IncludeActive:    true,
			IncluseArchived:  false,
			IncludeCommunity: true,
			IncludeDev:       true,
			IncludeInactive:  true,
			IncludePaid:      true,
			IncludeTrial:     true,
			InstancePreview:  false,
		}

		resp := struct {
			Customers []types.Customer `json:"customers"`
			TotalHits int              `json:"total_hits"`
		}{}

		err := c.DoJSON(context.TODO(), "POST", "/v3/customers/search", http.StatusOK, payload, &resp)
		if err != nil {
			return nil, errors.Wrapf(err, "search customers offset %d", offset)
		}

		allCustomers = append(allCustomers, resp.Customers...)

		offset = offset + pageSize
		if offset > resp.TotalHits {
			break
		}
	}

	return allCustomers, nil
}
