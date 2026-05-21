package cmd

import (
	"testing"

	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCustomerToPreviewLicense_HelmInstallEnabled(t *testing.T) {
	app := &types.App{
		ID:   "app_abc",
		Name: "Test App",
	}

	tests := []struct {
		name     string
		customer types.Customer
		want     bool
	}{
		{
			name: "IsHelmInstallEnabled true is mapped",
			customer: types.Customer{
				ID:                   "cust_abc",
				Name:                 "Helm Customer",
				IsHelmInstallEnabled: true,
			},
			want: true,
		},
		{
			name: "IsHelmInstallEnabled false is mapped",
			customer: types.Customer{
				ID:                   "cust_def",
				Name:                 "No Helm Customer",
				IsHelmInstallEnabled: false,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := customerToPreviewLicense(app, tt.customer)
			assert.Equal(t, tt.want, pl.IsHelmInstallEnabled)
		})
	}
}

func TestCustomerToPreviewLicense_KurlInstallEnabled(t *testing.T) {
	app := &types.App{
		ID:   "app_abc",
		Name: "Test App",
	}

	tests := []struct {
		name     string
		customer types.Customer
		want     bool
	}{
		{
			name: "IsKurlInstallEnabled true is mapped",
			customer: types.Customer{
				ID:                   "cust_abc",
				Name:                 "Kurl Customer",
				IsKurlInstallEnabled: true,
			},
			want: true,
		},
		{
			name: "IsKurlInstallEnabled false is mapped",
			customer: types.Customer{
				ID:                   "cust_def",
				Name:                 "No Kurl Customer",
				IsKurlInstallEnabled: false,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := customerToPreviewLicense(app, tt.customer)
			assert.Equal(t, tt.want, pl.IsKurlInstallEnabled)
		})
	}
}

func TestCustomerToPreviewLicense_HelmAirgapEnabled(t *testing.T) {
	app := &types.App{
		ID:   "app_abc",
		Name: "Test App",
	}

	tests := []struct {
		name     string
		customer types.Customer
		want     bool
	}{
		{
			name: "IsHelmAirgapEnabled true is mapped",
			customer: types.Customer{
				ID:                  "cust_abc",
				Name:                "Airgap Customer",
				IsHelmAirgapEnabled: true,
			},
			want: true,
		},
		{
			name: "IsHelmAirgapEnabled false is mapped",
			customer: types.Customer{
				ID:   "cust_def",
				Name: "No Airgap Customer",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := customerToPreviewLicense(app, tt.customer)
			assert.Equal(t, tt.want, pl.IsHelmAirgapEnabled)
		})
	}
}

func TestCustomerToPreviewLicense_AllCapabilities(t *testing.T) {
	app := &types.App{
		ID:   "app_abc",
		Name: "Test App",
	}

	cust := types.Customer{
		ID:                                "cust_full",
		Name:                              "Full Customer",
		Email:                             "full@example.com",
		InstallationID:                    "inst_123",
		IsAirgapEnabled:                   true,
		IsGitopsSupported:                 true,
		IsIdentityServiceSupported:        true,
		IsGeoaxisSupported:                true,
		IsSnapshotSupported:               true,
		IsSupportBundleUploadEnabled:      true,
		IsEmbeddedClusterDownloadEnabled:  true,
		IsEmbeddedClusterMultinodeEnabled: true,
		IsKotsInstallEnabled:              true,
		IsHelmInstallEnabled:              true,
		IsKurlInstallEnabled:              true,
		IsHelmAirgapEnabled:               true,
	}

	pl := customerToPreviewLicense(app, cust)

	assert.Equal(t, "inst_123", pl.ID)
	assert.Equal(t, app.ID, pl.AppID)
	assert.Equal(t, app.Name, pl.AppName)
	assert.Equal(t, "cust_full", pl.CustomerID)
	assert.Equal(t, "Full Customer", pl.CustomerName)
	assert.Equal(t, "full@example.com", pl.CustomerEmail)

	assert.True(t, pl.IsAirgapSupported)
	assert.True(t, pl.IsGitopsSupported)
	assert.True(t, pl.IsIdentityServiceSupported)
	assert.True(t, pl.IsGeoaxisSupported)
	assert.True(t, pl.IsSnapshotSupported)
	assert.True(t, pl.IsSupportBundleUploadSupported)
	assert.True(t, pl.IsEmbeddedClusterDownloadEnabled)
	assert.True(t, pl.IsEmbeddedClusterMultiNodeEnabled)
	assert.True(t, pl.IsKotsInstallEnabled)
	assert.True(t, pl.IsHelmInstallEnabled)
	assert.True(t, pl.IsKurlInstallEnabled)
	assert.True(t, pl.IsHelmAirgapEnabled)
}
