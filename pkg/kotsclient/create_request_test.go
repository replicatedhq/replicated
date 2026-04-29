package kotsclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/require"
)

func TestCreateNetworkIncludesPolicy(t *testing.T) {
	var requestBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v3/network", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&requestBody))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"network":{"id":"network-id"}}`))
	}))
	defer server.Close()

	httpClient := platformclient.NewHTTPClient(server.URL, "fake-api-key")
	client := &VendorV3Client{HTTPClient: *httpClient}

	_, ve, err := client.CreateNetwork(CreateNetworkOpts{
		Name:   "test-network",
		TTL:    "1h",
		Policy: "airgap",
	})
	require.NoError(t, err)
	require.Nil(t, ve)
	require.Equal(t, "test-network", requestBody["name"])
	require.Equal(t, "1h", requestBody["ttl"])
	require.Equal(t, "airgap", requestBody["policy"])
}

func TestCreateVMIncludesNetworkPolicy(t *testing.T) {
	var requestBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v3/vm", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&requestBody))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"vms":[]}`))
	}))
	defer server.Close()

	httpClient := platformclient.NewHTTPClient(server.URL, "fake-api-key")
	client := &VendorV3Client{HTTPClient: *httpClient}

	_, ve, err := client.CreateVM(CreateVMOpts{
		Name:          "test-vm",
		Distribution:  "ubuntu",
		Version:       "22.04",
		Count:         1,
		Network:       "network-id",
		NetworkPolicy: "airgap",
	})
	require.NoError(t, err)
	require.Nil(t, ve)
	require.Equal(t, "test-vm", requestBody["name"])
	require.Equal(t, "network-id", requestBody["network_id"])
	require.Equal(t, "airgap", requestBody["network_policy"])
}

func TestCreateClusterIncludesNetworkPolicy(t *testing.T) {
	var requestBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v3/cluster", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&requestBody))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"cluster":{"id":"cluster-id"}}`))
	}))
	defer server.Close()

	httpClient := platformclient.NewHTTPClient(server.URL, "fake-api-key")
	client := &VendorV3Client{HTTPClient: *httpClient}

	_, ve, err := client.CreateCluster(CreateClusterOpts{
		Name:                   "test-cluster",
		KubernetesDistribution: "k3s",
		KubernetesVersion:      "1.31",
		NodeCount:              1,
		NetworkPolicy:          "airgap",
	})
	require.NoError(t, err)
	require.Nil(t, ve)
	require.Equal(t, "test-cluster", requestBody["name"])
	require.Equal(t, "airgap", requestBody["network_policy"])
}
