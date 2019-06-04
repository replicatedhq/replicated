package entitlements

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
)

// don't care what it looks like, just gonna json.MarshalIndent it to stdout
type ShipRelease map[string]interface{}

type PremGraphQLClient struct {
	GQLServer      *url.URL
	CustomerID     string
	InstallationID string
	Logger         log.Logger
}

type GraphQLResponseCustomerSpec struct {
	Data   GetCustomerSpecResponse `json:"data,omitempty"`
	Errors []graphql.Error         `json:"errors,omitempty"`
}

type GetCustomerSpecResponse struct {
	ShipRelease ShipRelease `json:"shipRelease,omitempty"` //
}

func (r GraphQLResponseCustomerSpec) GraphQLError() []graphql.Error {
	return r.Errors
}

func (c *PremGraphQLClient) FetchCustomerRelease() (ShipRelease, error) {
	requestObj := graphql.Request{
		Query: `
query {
  shipRelease {
    id
    channelId
    channelName
    channelIcon
    semver
    releaseNotes
    spec
    images {
      url
      source
      appSlug
      imageKey
    }
    githubContents {
      repo
      path
      ref
      files {
        name
        path
        sha
        size
        data
      }
    }
    entitlements {
      values {
        key
        value
      }
      meta {
        customerID
        lastUpdated
      }
      signature
    }
    created
    registrySecret
  }
}`,
		Variables: map[string]interface{}{},
	}
	response := GraphQLResponseCustomerSpec{}
	err := c.ExecuteRequest(requestObj, &response)
	if err != nil {
		return nil, errors.Wrapf(err, "execute request")
	}

	if err := c.checkErrors(response); err != nil {
		return nil, err
	}

	return response.Data.ShipRelease, nil
}

func (c *PremGraphQLClient) ExecuteRequest(
	requestObj graphql.Request,
	deserializeTarget interface{},
) error {
	debug := log.With(level.Debug(c.Logger), "type", "graphQLClient")
	body, err := json.Marshal(requestObj)
	if err != nil {
		return errors.Wrap(err, "marshal body")
	}
	bodyReader := bytes.NewReader(body)
	req, err := http.NewRequest("POST", c.GQLServer.String(), bodyReader)
	if err != nil {
		return errors.Wrap(err, "marshal body")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.CustomerID, c.InstallationID))))
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "marshal body")
	}
	if resp == nil {
		return errors.New("nil response from gql")
	}
	if resp.Body == nil {
		return errors.New("nil response.Body from gql")
	}
	responseBody, err := ioutil.ReadAll(resp.Body)
	debug.Log("body", responseBody)
	if err != nil {
		return errors.Wrap(err, "marshal body")
	}
	if err := json.Unmarshal(responseBody, deserializeTarget); err != nil {
		return errors.Wrap(err, "unmarshal response")
	}

	return nil
}

func (c *PremGraphQLClient) checkErrors(errer graphql.GQLError) error {
	if errer.GraphQLError() != nil && len(errer.GraphQLError()) > 0 {
		var multiErr *multierror.Error
		for _, err := range errer.GraphQLError() {
			multiErr = multierror.Append(multiErr, fmt.Errorf("%s: %s", err.Code, err.Message))

		}
		return multiErr.ErrorOrNil()
	}
	return nil
}
