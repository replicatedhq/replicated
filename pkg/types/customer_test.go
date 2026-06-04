package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomer_UnmarshalJSON_IsKurlInstallEnabled(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		wantKurl bool
	}{
		{
			name:     "isKurlInstallEnabled true",
			payload:  `{"id":"cust_abc","isKurlInstallEnabled":true}`,
			wantKurl: true,
		},
		{
			name:     "isKurlInstallEnabled false",
			payload:  `{"id":"cust_abc","isKurlInstallEnabled":false}`,
			wantKurl: false,
		},
		{
			name:     "isKurlInstallEnabled absent defaults to false",
			payload:  `{"id":"cust_abc"}`,
			wantKurl: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c Customer
			err := json.Unmarshal([]byte(tt.payload), &c)
			require.NoError(t, err)
			assert.Equal(t, tt.wantKurl, c.IsKurlInstallEnabled)
		})
	}
}

func TestCustomer_UnmarshalJSON_IsHelmAirgapEnabled(t *testing.T) {
	tests := []struct {
		name       string
		payload    string
		wantAirgap bool
	}{
		{
			name:       "isHelmAirgapEnabled true",
			payload:    `{"id":"cust_abc","isHelmAirgapEnabled":true}`,
			wantAirgap: true,
		},
		{
			name:       "isHelmAirgapEnabled false",
			payload:    `{"id":"cust_abc","isHelmAirgapEnabled":false}`,
			wantAirgap: false,
		},
		{
			name:       "isHelmAirgapEnabled absent defaults to false",
			payload:    `{"id":"cust_abc"}`,
			wantAirgap: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c Customer
			err := json.Unmarshal([]byte(tt.payload), &c)
			require.NoError(t, err)
			assert.Equal(t, tt.wantAirgap, c.IsHelmAirgapEnabled)
		})
	}
}

func TestCustomer_MarshalJSON_NewFields(t *testing.T) {
	c := Customer{
		ID:                   "cust_abc",
		Name:                 "Test",
		IsHelmInstallEnabled: true,
		IsKurlInstallEnabled: true,
		IsHelmAirgapEnabled:  true,
	}

	data, err := json.Marshal(c)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"isHelmInstallEnabled":true`)
	assert.Contains(t, jsonStr, `"isKurlInstallEnabled":true`)
	assert.Contains(t, jsonStr, `"isHelmAirgapEnabled":true`)
}

func TestCustomer_RoundTrip_AllInstallFields(t *testing.T) {
	original := Customer{
		ID:                   "cust_abc",
		Name:                 "Test",
		IsHelmInstallEnabled: true,
		IsKurlInstallEnabled: true,
		IsHelmAirgapEnabled:  true,
		IsKotsInstallEnabled: true,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Customer
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, original.IsHelmInstallEnabled, restored.IsHelmInstallEnabled)
	assert.Equal(t, original.IsKurlInstallEnabled, restored.IsKurlInstallEnabled)
	assert.Equal(t, original.IsHelmAirgapEnabled, restored.IsHelmAirgapEnabled)
	assert.Equal(t, original.IsKotsInstallEnabled, restored.IsKotsInstallEnabled)
}

func TestCustomer_UnmarshalJSON_IsHelmVMDownloadEnabled_BackwardsCompat(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantVM  bool
	}{
		{
			name:    "isHelmVmDownloadEnabled true",
			payload: `{"id":"cust_abc","isHelmVmDownloadEnabled":true}`,
			wantVM:  true,
		},
		{
			name:    "isHelmVmDownloadEnabled false",
			payload: `{"id":"cust_abc","isHelmVmDownloadEnabled":false}`,
			wantVM:  false,
		},
		{
			name:    "isHelmVmDownloadEnabled absent defaults to false",
			payload: `{"id":"cust_abc"}`,
			wantVM:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c Customer
			err := json.Unmarshal([]byte(tt.payload), &c)
			require.NoError(t, err)
			assert.Equal(t, tt.wantVM, c.IsHelmVMDownloadEnabled)
		})
	}
}

func TestCustomer_MarshalJSON_IsHelmVMDownloadEnabled_BackwardsCompat(t *testing.T) {
	c := Customer{
		ID:                      "cust_abc",
		IsHelmVMDownloadEnabled: true,
	}

	data, err := json.Marshal(c)
	require.NoError(t, err)

	assert.Contains(t, string(data), `"isHelmVmDownloadEnabled":true`)
}

func TestCustomer_UnmarshalJSON_BothHelmFields(t *testing.T) {
	payload := `{
		"id": "cust_abc",
		"name": "Test Customer",
		"isHelmInstallEnabled": true,
		"isHelmVmDownloadEnabled": true
	}`

	var c Customer
	err := json.Unmarshal([]byte(payload), &c)
	require.NoError(t, err)

	assert.True(t, c.IsHelmInstallEnabled, "IsHelmInstallEnabled should be true")
	assert.True(t, c.IsHelmVMDownloadEnabled, "IsHelmVMDownloadEnabled should be true")
}

func TestCustomer_MarshalJSON_BothHelmFields(t *testing.T) {
	c := Customer{
		ID:                      "cust_abc",
		IsHelmInstallEnabled:    true,
		IsHelmVMDownloadEnabled: true,
	}

	data, err := json.Marshal(c)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"isHelmInstallEnabled":true`)
	assert.Contains(t, jsonStr, `"isHelmVmDownloadEnabled":true`)
}

func TestCustomer_UnmarshalJSON_FullAPIPayload(t *testing.T) {
	payload := `{
		"id": "cust_abc123",
		"customId": "custom-id-456",
		"name": "Acme Corp",
		"email": "admin@acme.com",
		"type": "prod",
		"airgap": true,
		"isEmbeddedClusterDownloadEnabled": true,
		"isEmbeddedClusterMultinodeEnabled": false,
		"isGeoaxisSupported": false,
		"isHelmAirgapEnabled": true,
		"isHelmInstallEnabled": true,
		"isHelmVmDownloadEnabled": false,
		"isIdentityServiceSupported": false,
		"isInstallerSupportEnabled": true,
		"isKotsInstallEnabled": true,
		"isKurlInstallEnabled": true,
		"isSnapshotSupported": true,
		"isSupportBundleUploadEnabled": true,
		"isGitopsSupported": false
	}`

	var c Customer
	err := json.Unmarshal([]byte(payload), &c)
	require.NoError(t, err)

	assert.Equal(t, "cust_abc123", c.ID)
	assert.Equal(t, "custom-id-456", c.CustomID)
	assert.Equal(t, "Acme Corp", c.Name)
	assert.Equal(t, "admin@acme.com", c.Email)
	assert.Equal(t, "prod", c.Type)
	assert.True(t, c.IsAirgapEnabled)
	assert.True(t, c.IsEmbeddedClusterDownloadEnabled)
	assert.False(t, c.IsEmbeddedClusterMultinodeEnabled)
	assert.False(t, c.IsGeoaxisSupported)
	assert.True(t, c.IsHelmAirgapEnabled)
	assert.True(t, c.IsHelmInstallEnabled)
	assert.False(t, c.IsHelmVMDownloadEnabled)
	assert.False(t, c.IsIdentityServiceSupported)
	assert.True(t, c.IsInstallerSupportEnabled)
	assert.True(t, c.IsKotsInstallEnabled)
	assert.True(t, c.IsKurlInstallEnabled)
	assert.True(t, c.IsSnapshotSupported)
	assert.True(t, c.IsSupportBundleUploadEnabled)
	assert.False(t, c.IsGitopsSupported)
}
