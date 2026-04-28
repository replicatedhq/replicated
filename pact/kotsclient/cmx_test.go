package kotsclient

import (
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	realkotsclient "github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/require"
)

// Each interaction below uses a dedicated team + token fixture on the provider
// side so that the tests do not chain state with one another. This mirrors the
// pattern used by the policy pact tests and matches what
// pact-provider-verifier actually does (every interaction is replayed
// independently — there is no Create→Get plumbing).

func Test_CreateVM(t *testing.T) {
	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "cli-create-vm-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		vms, ve, err := client.CreateVM(realkotsclient.CreateVMOpts{
			Name:          "vxlan-vm",
			Distribution:  "ubuntu",
			Version:       "22.04",
			Count:         1,
			DiskGiB:       25,
			NetworkPolicy: "airgap",
			TTL:           "2h",
			InstanceType:  "r1.small",
			PublicKeys:    []string{"c3NoLWVkMjU1MTkgQUFBQUMzTnphQzFsWkRJMU5URTVBQUFBSVBNUWJqZUdIcWNiaWRjTmc1T3NSRWZZbDExOG9OT1F3Rml5V2cvZzZ3ZmkgdmFuZG9vci1wYWN0QGV4YW1wbGUuY29t"},
			Tags:          []types.Tag{{Key: "test", Value: "pact"}},
		})
		require.NoError(t, err)
		require.Nil(t, ve)
		require.Len(t, vms, 1)
		require.Equal(t, "vxlan-vm", vms[0].Name)

		return nil
	}

	pact.AddInteraction().
		Given("Create CMX VM").
		UponReceiving("A request to create a CMX VM").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/vm"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("cli-create-vm-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"name":           "vxlan-vm",
				"distribution":   "ubuntu",
				"version":        "22.04",
				"count":          1,
				"disk_gib":       25,
				"network_id":     "",
				"network_policy": "airgap",
				"ttl":            "2h",
				"instance_type":  "r1.small",
				"public_keys":    []string{"c3NoLWVkMjU1MTkgQUFBQUMzTnphQzFsWkRJMU5URTVBQUFBSVBNUWJqZUdIcWNiaWRjTmc1T3NSRWZZbDExOG9OT1F3Rml5V2cvZzZ3ZmkgdmFuZG9vci1wYWN0QGV4YW1wbGUuY29t"},
				"tags":           []map[string]interface{}{{"key": "test", "value": "pact"}},
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"vms": []map[string]interface{}{
					{
						"id":           dsl.Like("ebc40fe9"),
						"name":         dsl.Like("vxlan-vm"),
						"distribution": dsl.Like("ubuntu"),
						"version":      dsl.Like("22.04"),
						"network_id":   dsl.Like("befee8ca"),
						"status":       dsl.Like("queued"),
						"ttl":          dsl.Like("2h"),
						"tags":         []map[string]interface{}{{"key": "test", "value": "pact"}},
					},
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_GetVM(t *testing.T) {
	const vmID = "cli-get-vm-id"

	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "cli-get-vm-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		vm, err := client.GetVM(vmID)
		require.NoError(t, err)
		require.Equal(t, "vxlan-vm", vm.Name)

		return nil
	}

	pact.AddInteraction().
		Given("Get CMX VM").
		UponReceiving("A request to get a CMX VM").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/vm/" + vmID),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("cli-get-vm-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"vm": map[string]interface{}{
					"id":           dsl.Like(vmID),
					"name":         dsl.Like("vxlan-vm"),
					"distribution": dsl.Like("ubuntu"),
					"version":      dsl.Like("22.04"),
					"network_id":   dsl.Like("befee8ca"),
					"status":       dsl.Like("running"),
					"ttl":          dsl.Like("2h"),
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_CreateCluster(t *testing.T) {
	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "cli-create-cluster-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		cluster, ve, err := client.CreateCluster(realkotsclient.CreateClusterOpts{
			Name:                   "vxlan-cluster",
			KubernetesDistribution: "fake",
			KubernetesVersion:      "1.25",
			NodeCount:              1,
			DiskGiB:                50,
			TTL:                    "2h",
			InstanceType:           "r1.small",
			Tags:                   []types.Tag{{Key: "test", Value: "pact"}},
		})
		require.NoError(t, err)
		require.Nil(t, ve)
		require.Equal(t, "vxlan-cluster", cluster.Name)

		return nil
	}

	pact.AddInteraction().
		Given("Create CMX cluster").
		UponReceiving("A request to create a CMX cluster").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/cluster"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("cli-create-cluster-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"name":                    "vxlan-cluster",
				"kubernetes_distribution": "fake",
				"kubernetes_version":      "1.25",
				"ip_family":               "",
				"license_id":              "",
				"node_count":              1,
				"min_node_count":          nil,
				"max_node_count":          nil,
				"disk_gib":                50,
				"ttl":                     "2h",
				"node_groups":             nil,
				"instance_type":           "r1.small",
				"tags":                    []map[string]interface{}{{"key": "test", "value": "pact"}},
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"cluster": map[string]interface{}{
					"id":                      dsl.Like("0fed1234"),
					"name":                    dsl.Like("vxlan-cluster"),
					"kubernetes_distribution": dsl.Like("fake"),
					"kubernetes_version":      dsl.Like("1.25"),
					"network_id":              dsl.Like("befee8ca"),
					"status":                  dsl.Like("queued"),
					"ttl":                     dsl.Like("2h"),
					"tags":                    []map[string]interface{}{{"key": "test", "value": "pact"}},
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_GetCluster(t *testing.T) {
	const clusterID = "cli-get-cluster-id"

	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "cli-get-cluster-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		cluster, err := client.GetCluster(clusterID)
		require.NoError(t, err)
		require.Equal(t, "vxlan-cluster", cluster.Name)

		return nil
	}

	pact.AddInteraction().
		Given("Get CMX cluster").
		UponReceiving("A request to get a CMX cluster").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/cluster/" + clusterID),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("cli-get-cluster-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"cluster": map[string]interface{}{
					"id":                      dsl.Like(clusterID),
					"name":                    dsl.Like("vxlan-cluster"),
					"kubernetes_distribution": dsl.Like("fake"),
					"kubernetes_version":      dsl.Like("1.25"),
					"network_id":              dsl.Like("befee8ca"),
					"status":                  dsl.Like("running"),
					"ttl":                     dsl.Like("2h"),
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_CreateNetwork(t *testing.T) {
	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "cli-create-network-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		network, ve, err := client.CreateNetwork(realkotsclient.CreateNetworkOpts{
			Name:   "team-cmx-network",
			TTL:    "2h",
			Policy: "airgap",
		})
		require.NoError(t, err)
		require.Nil(t, ve)
		require.Equal(t, "team-cmx-network", network.Name)

		return nil
	}

	pact.AddInteraction().
		Given("Create CMX network").
		UponReceiving("A request to create a CMX network").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/network"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("cli-create-network-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"name":   "team-cmx-network",
				"ttl":    "2h",
				"policy": "airgap",
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"network": map[string]interface{}{
					"id":     dsl.Like("0fed1234"),
					"name":   dsl.Like("team-cmx-network"),
					"status": dsl.Like("queued"),
					"ttl":    dsl.Like("2h"),
					"policy": dsl.Like("airgap"),
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_GetNetwork(t *testing.T) {
	const networkID = "cli-get-network-id"

	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "cli-get-network-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		network, err := client.GetNetwork(networkID)
		require.NoError(t, err)
		require.Equal(t, "airgap", network.Policy)

		return nil
	}

	pact.AddInteraction().
		Given("Get CMX network").
		UponReceiving("A request to get a CMX network").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/network/" + networkID),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("cli-get-network-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"network": map[string]interface{}{
					"id":     dsl.Like(networkID),
					"name":   dsl.Like("team-cmx-network"),
					"status": dsl.Like("running"),
					"ttl":    dsl.Like("2h"),
					"policy": dsl.Like("airgap"),
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_UpdateNetwork(t *testing.T) {
	const networkID = "cli-update-network-id"

	test := func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		api := platformclient.NewHTTPClient(u, "cli-update-network-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		network, err := client.UpdateNetwork(networkID, realkotsclient.UpdateNetworkOpts{
			Policy: "airgap",
		})
		require.NoError(t, err)
		require.Equal(t, "airgap", network.Policy)

		return nil
	}

	pact.AddInteraction().
		Given("Update CMX network policy").
		UponReceiving("A request to update a CMX network policy").
		WithRequest(dsl.Request{
			Method: "PUT",
			Path:   dsl.String("/v3/network/" + networkID + "/update"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("cli-update-network-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"policy": "airgap",
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"network": map[string]interface{}{
					"id":     dsl.Like(networkID),
					"name":   dsl.Like("team-cmx-network"),
					"status": dsl.Like("running"),
					"ttl":    dsl.Like("2h"),
					"policy": dsl.Like("airgap"),
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
