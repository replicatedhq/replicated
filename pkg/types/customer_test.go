package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
