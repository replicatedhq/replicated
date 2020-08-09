package kotsclient

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

const kotsListCustomers = `
	query customers($appId: String!, $appType: String!) {
		customers(appId: $appId, appType: $appType) {
            customers {
		        id
		        name
		        channels {
		            id
		            name
		            currentVersion
		        }
            }
        }
	}
	`

type GraphQLResponseListCustomers struct {
	Data   *CustomerDataWrapper `json:"data,omitempty"`
	Errors []graphql.GQLError   `json:"errors,omitempty"`
}

type CustomerDataWrapper struct {
	Customers CustomerData `json:"customers"`
}

type CustomerData struct {
	Customers []*Customer `json:"customers"`
}

type Customer struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Channels  []*KotsChannel `json:"channels"`
	Type      string         `json:"type"`
	ExpiresAt string         `json:"expiresAt"`
}

type ErrCustomerNotFound struct {
	Name string
}
func (e ErrCustomerNotFound) Error() string {
	return fmt.Sprintf("customer %q not found", e.Name)
}



func (c *GraphQLClient) ListCustomers(appID string) ([]types.Customer, error) {
	response := GraphQLResponseListCustomers{}

	request := graphql.Request{
		Query: kotsListCustomers,

		Variables: map[string]interface{}{
			"appId":   appID,
			"appType": "kots",
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, errors.Wrap(err, "execute gql request")
	}

	customers := make([]types.Customer, 0, 0)
	for _, kotsCustomer := range response.Data.Customers.Customers {

		kotsChannels := make([]types.Channel, 0, 0)
		for _, kotsChannel := range kotsCustomer.Channels {
			channel := types.Channel{
				ID:              kotsChannel.ID,
				Name:            kotsChannel.Name,
				ReleaseLabel:    kotsChannel.CurrentVersion,
				ReleaseSequence: kotsChannel.ReleaseSequence,
			}
			kotsChannels = append(kotsChannels, channel)
		}
		customer, err := types.Customer{
			ID:       kotsCustomer.ID,
			Name:     kotsCustomer.Name,
			Type:     kotsCustomer.Type,
			Channels: kotsChannels,
		}.WithExpiryTime(kotsCustomer.ExpiresAt)

		if err != nil {
			return nil, errors.Wrapf(err, "set expiry time for customer %q", kotsCustomer.ID)
		}

		customers = append(customers, customer)
	}

	return customers, nil
}

func (c *GraphQLClient) GetCustomerByName(appID string, name string) (*types.Customer, error) {
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
