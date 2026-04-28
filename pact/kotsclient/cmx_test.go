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

const (
	cmxToken     = "team-cmx-token"
	cmxVMID      = "ebc40fe9"
	cmxClusterID = "0fed1234"
	cmxNetworkID = "0fed1234"
)

func Test_CreateGetVM(t *testing.T) {
	var test = func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, cmxToken)
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		vms, ve, err := client.CreateVM(realkotsclient.CreateVMOpts{
			Name:          "vxlan-vm",
			Distribution:  "ubuntu",
			Version:       "22.04",
			Count:         1,
			DiskGiB:       25,
			NetworkPolicy: "airgap",
			TTL:           "2h",
			InstanceType:  "standard",
			Tags:          []types.Tag{{Key: "test", Value: "pact"}},
		})
		require.NoError(t, err)
		require.Nil(t, ve)
		require.Len(t, vms, 1)
		require.Equal(t, cmxVMID, vms[0].ID)

		vm, err := client.GetVM(cmxVMID)
		require.NoError(t, err)
		require.Equal(t, "vxlan-vm", vm.Name)

		return nil
	}

	pact.AddInteraction().
		Given("Create CMX VM").
		UponReceiving("A request to create a CMX VM").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/vm"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String(cmxToken),
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
				"instance_type":  "standard",
				"tags":           []map[string]interface{}{{"key": "test", "value": "pact"}},
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"vms": []map[string]interface{}{
					{
						"id":           dsl.Like(cmxVMID),
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

	pact.AddInteraction().
		Given("Get CMX VM").
		UponReceiving("A request to get a CMX VM").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/vm/" + cmxVMID),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String(cmxToken),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"vm": map[string]interface{}{
					"id":           dsl.Like(cmxVMID),
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

func Test_CreateGetCluster(t *testing.T) {
	var test = func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, cmxToken)
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
		require.Equal(t, cmxClusterID, cluster.ID)

		cluster, err = client.GetCluster(cmxClusterID)
		require.NoError(t, err)
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
				"Authorization": dsl.String(cmxToken),
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
					"id":                      dsl.Like(cmxClusterID),
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

	pact.AddInteraction().
		Given("Get CMX cluster").
		UponReceiving("A request to get a CMX cluster").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/cluster/" + cmxClusterID),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String(cmxToken),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"cluster": map[string]interface{}{
					"id":                      dsl.Like(cmxClusterID),
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

func Test_CreateGetNetwork(t *testing.T) {
	var test = func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, cmxToken)
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		network, ve, err := client.CreateNetwork(realkotsclient.CreateNetworkOpts{
			Name:   "team-cmx-network",
			TTL:    "2h",
			Policy: "airgap",
		})
		require.NoError(t, err)
		require.Nil(t, ve)
		require.Equal(t, cmxNetworkID, network.ID)

		network, err = client.GetNetwork(cmxNetworkID)
		require.NoError(t, err)
		require.Equal(t, "airgap", network.Policy)

		return nil
	}

	pact.AddInteraction().
		Given("Create CMX network").
		UponReceiving("A request to create a CMX network").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/network"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String(cmxToken),
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
					"id":     dsl.Like(cmxNetworkID),
					"name":   dsl.Like("team-cmx-network"),
					"status": dsl.Like("queued"),
					"ttl":    dsl.Like("2h"),
					"policy": dsl.Like("airgap"),
				},
			},
		})

	pact.AddInteraction().
		Given("Get CMX network").
		UponReceiving("A request to get a CMX network").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/network/" + cmxNetworkID),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String(cmxToken),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"network": map[string]interface{}{
					"id":     dsl.Like(cmxNetworkID),
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

func Test_UpdateGetNetwork(t *testing.T) {
	var test = func() error {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, cmxToken)
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		network, err := client.UpdateNetwork(cmxNetworkID, realkotsclient.UpdateNetworkOpts{
			Policy: "airgap",
		})
		require.NoError(t, err)
		require.Equal(t, "airgap", network.Policy)

		network, err = client.GetNetwork(cmxNetworkID)
		require.NoError(t, err)
		require.Equal(t, "airgap", network.Policy)

		return nil
	}

	pact.AddInteraction().
		Given("Update CMX network policy").
		UponReceiving("A request to update a CMX network policy").
		WithRequest(dsl.Request{
			Method: "PUT",
			Path:   dsl.String("/v3/network/" + cmxNetworkID + "/update"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String(cmxToken),
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
					"id":     dsl.Like(cmxNetworkID),
					"name":   dsl.Like("team-cmx-network"),
					"status": dsl.Like("running"),
					"ttl":    dsl.Like("2h"),
					"policy": dsl.Like("airgap"),
				},
			},
		})

	pact.AddInteraction().
		Given("Get updated CMX network").
		UponReceiving("A request to get an updated CMX network").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/network/" + cmxNetworkID),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String(cmxToken),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"network": map[string]interface{}{
					"id":     dsl.Like(cmxNetworkID),
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
