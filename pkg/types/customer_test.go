package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomer_UnmarshalJSON_IsHelmInstallEnabled(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		wantHelm bool
	}{
		{
			name:     "isHelmInstallEnabled true",
			payload:  `{"id":"cust_abc","name":"Test","isHelmInstallEnabled":true}`,
			wantHelm: true,
		},
		{
			name:     "isHelmInstallEnabled false",
			payload:  `{"id":"cust_abc","name":"Test","isHelmInstallEnabled":false}`,
			wantHelm: false,
		},
		{
			name:     "isHelmInstallEnabled absent defaults to false",
			payload:  `{"id":"cust_abc","name":"Test"}`,
			wantHelm: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c Customer
			err := json.Unmarshal([]byte(tt.payload), &c)
			require.NoError(t, err)
			assert.Equal(t, tt.wantHelm, c.IsHelmInstallEnabled)
		})
	}
}

func TestCustomer_MarshalJSON_IsHelmInstallEnabled(t *testing.T) {
	c := Customer{
		ID:                   "cust_abc",
		Name:                 "Test",
		IsHelmInstallEnabled: true,
	}

	data, err := json.Marshal(c)
	require.NoError(t, err)

	assert.Contains(t, string(data), `"isHelmInstallEnabled":true`)
}

func TestCustomer_RoundTrip_IsHelmInstallEnabled(t *testing.T) {
	original := Customer{
		ID:                   "cust_abc",
		Name:                 "Test",
		IsHelmInstallEnabled: true,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Customer
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, original.IsHelmInstallEnabled, restored.IsHelmInstallEnabled)
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
	// Verify both fields can coexist in the same payload
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
	// Simulates a realistic Vendor API response payload
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
		"isHelmInstallEnabled": true,
		"isHelmVmDownloadEnabled": false,
		"isIdentityServiceSupported": false,
		"isInstallerSupportEnabled": true,
		"isKotsInstallEnabled": true,
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
	assert.True(t, c.IsHelmInstallEnabled)
	assert.False(t, c.IsHelmVMDownloadEnabled)
	assert.False(t, c.IsIdentityServiceSupported)
	assert.True(t, c.IsInstallerSupportEnabled)
	assert.True(t, c.IsKotsInstallEnabled)
	assert.True(t, c.IsSnapshotSupported)
	assert.True(t, c.IsSupportBundleUploadEnabled)
	assert.False(t, c.IsGitopsSupported)
}
