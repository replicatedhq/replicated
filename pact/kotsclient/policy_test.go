package kotsclient

import (
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	realkotsclient "github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

const adminPolicyDefinition = `{"v1":{"name":"Admin","resources":{"allowed":["**/*"],"denied":[]}}}`

func Test_ListPolicies(t *testing.T) {
	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "replicated-cli-list-policies-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		policies, err := client.ListPolicies()
		assert.Nil(t, err)
		assert.NotEmpty(t, policies)
		assert.Equal(t, "Admin", policies[0].Name)
		assert.True(t, policies[0].ReadOnly)

		return nil
	}

	pact.AddInteraction().
		Given("List team RBAC policies").
		UponReceiving("A request to list team RBAC policies").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/policies"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-list-policies-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"policies": dsl.EachLike(map[string]interface{}{
					"id":          dsl.Like("replicated-cli-list-policies-admin"),
					"teamId":      dsl.Like("replicated-cli-list-policies-team"),
					"name":        dsl.Like("Admin"),
					"description": dsl.Like(""),
					"definition":  dsl.Like(adminPolicyDefinition),
					"readOnly":    dsl.Like(true),
				}, 1),
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_GetPolicyByNameOrID(t *testing.T) {
	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "replicated-cli-get-policy-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		policy, err := client.GetPolicyByNameOrID("Custom")
		assert.Nil(t, err)
		assert.Equal(t, "Custom", policy.Name)
		assert.False(t, policy.ReadOnly)

		_, err = client.GetPolicyByNameOrID("Does Not Exist")
		assert.Error(t, err)

		return nil
	}

	pact.AddInteraction().
		Given("Get team RBAC policy by name").
		UponReceiving("A request to list policies in order to resolve one by name").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/policies"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-get-policy-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"policies": dsl.EachLike(map[string]interface{}{
					"id":          dsl.Like("replicated-cli-get-policy-custom"),
					"teamId":      dsl.Like("replicated-cli-get-policy-team"),
					"name":        dsl.Like("Custom"),
					"description": dsl.Like(""),
					"definition":  dsl.Like(adminPolicyDefinition),
					"readOnly":    dsl.Like(false),
				}, 1),
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_CreatePolicy(t *testing.T) {
	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "replicated-cli-create-policy-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		policy, err := client.CreatePolicy("New Policy", "A test policy", adminPolicyDefinition)
		assert.Nil(t, err)
		assert.Equal(t, "New Policy", policy.Name)
		assert.Equal(t, "A test policy", policy.Description)
		assert.False(t, policy.ReadOnly)

		return nil
	}

	pact.AddInteraction().
		Given("Create a team RBAC policy").
		UponReceiving("A request to create a team RBAC policy").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/policy"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-create-policy-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"name":        "New Policy",
				"description": "A test policy",
				"definition":  adminPolicyDefinition,
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"policy": map[string]interface{}{
					"id":          dsl.Like("replicated-cli-create-policy-new"),
					"teamId":      dsl.Like("replicated-cli-create-policy-team"),
					"name":        dsl.String("New Policy"),
					"description": dsl.String("A test policy"),
					"definition":  dsl.Like(adminPolicyDefinition),
					"readOnly":    false,
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_UpdatePolicy(t *testing.T) {
	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "replicated-cli-update-policy-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		policy, err := client.UpdatePolicy(
			"replicated-cli-update-policy-id",
			"Renamed Policy",
			"Updated description",
			adminPolicyDefinition,
		)
		assert.Nil(t, err)
		assert.Equal(t, "Renamed Policy", policy.Name)
		assert.Equal(t, "Updated description", policy.Description)

		return nil
	}

	pact.AddInteraction().
		Given("Update a team RBAC policy").
		UponReceiving("A request to update a team RBAC policy").
		WithRequest(dsl.Request{
			Method: "PUT",
			Path:   dsl.String("/v3/policy/replicated-cli-update-policy-id"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-update-policy-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"name":        "Renamed Policy",
				"description": "Updated description",
				"definition":  adminPolicyDefinition,
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"policy": map[string]interface{}{
					"id":          dsl.Like("replicated-cli-update-policy-id"),
					"teamId":      dsl.Like("replicated-cli-update-policy-team"),
					"name":        dsl.String("Renamed Policy"),
					"description": dsl.String("Updated description"),
					"definition":  dsl.Like(adminPolicyDefinition),
					"readOnly":    false,
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_DeletePolicy(t *testing.T) {
	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "replicated-cli-delete-policy-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		err := client.DeletePolicy("replicated-cli-delete-policy-id")
		assert.Nil(t, err)

		return nil
	}

	pact.AddInteraction().
		Given("Delete a team RBAC policy").
		UponReceiving("A request to delete a team RBAC policy").
		WithRequest(dsl.Request{
			Method: "DELETE",
			Path:   dsl.String("/v3/policy/replicated-cli-delete-policy-id"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-delete-policy-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
