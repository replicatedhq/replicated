package print

import (
	"bytes"
	"encoding/json"
	"testing"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CustomerAttrs_JSON_IsHelmInstallEnabled(t *testing.T) {
	tests := []struct {
		name                 string
		customer             *types.Customer
		wantHelmInstall      bool
		wantHelmVMDownload   bool
	}{
		{
			name: "isHelmInstallEnabled true in JSON output",
			customer: &types.Customer{
				ID:                   "cust_abc",
				Name:                 "Test Customer",
				IsHelmInstallEnabled: true,
			},
			wantHelmInstall:    true,
			wantHelmVMDownload: false,
		},
		{
			name: "both helm fields in JSON output",
			customer: &types.Customer{
				ID:                      "cust_abc",
				Name:                    "Test Customer",
				IsHelmInstallEnabled:    true,
				IsHelmVMDownloadEnabled: true,
			},
			wantHelmInstall:    true,
			wantHelmVMDownload: true,
		},
		{
			name: "helm fields false in JSON output",
			customer: &types.Customer{
				ID:   "cust_abc",
				Name: "Test Customer",
			},
			wantHelmInstall:    false,
			wantHelmVMDownload: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)

			err := CustomerAttrs("json", w, "", "", nil, "", tt.customer)
			require.NoError(t, err)

			// Parse the JSON output and verify the fields
			var parsed map[string]interface{}
			err = json.Unmarshal(out.Bytes(), &parsed)
			require.NoError(t, err, "JSON output should be valid JSON: %s", out.String())

			helmInstall, ok := parsed["isHelmInstallEnabled"]
			assert.True(t, ok, "JSON output must contain isHelmInstallEnabled key")
			assert.Equal(t, tt.wantHelmInstall, helmInstall, "isHelmInstallEnabled value mismatch")

			helmVM, ok := parsed["isHelmVmDownloadEnabled"]
			assert.True(t, ok, "JSON output must contain isHelmVmDownloadEnabled key")
			assert.Equal(t, tt.wantHelmVMDownload, helmVM, "isHelmVmDownloadEnabled value mismatch")
		})
	}
}

func Test_CustomerAttrs_TextOutput(t *testing.T) {
	cust := &types.Customer{
		ID:                   "cust_abc",
		Name:                 "Test Customer",
		Email:                "test@example.com",
		InstallationID:       "inst_123",
		IsHelmInstallEnabled: true,
	}

	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)

	err := CustomerAttrs("table", w, "", "", nil, "", cust)
	require.NoError(t, err)

	result := out.String()
	assert.Contains(t, result, "ID: cust_abc")
	assert.Contains(t, result, "NAME: Test Customer")
	assert.Contains(t, result, "EMAIL: test@example.com")
}
