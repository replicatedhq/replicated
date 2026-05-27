package print

import (
	"bytes"
	"encoding/json"
	"testing"
	"text/tabwriter"

	channels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLicenseCounts_Table(t *testing.T) {
	counts := &channels.LicenseCounts{
		Active:   map[string]int64{"paid": 5, "trial": 3},
		Airgap:   map[string]int64{"paid": 2},
		Inactive: map[string]int64{"trial": 1},
		Total:    map[string]int64{"paid": 7, "trial": 4},
	}

	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, LicenseCounts("table", w, counts))

	got := out.String()
	assert.Contains(t, got, "LICENSE_TYPE")
	assert.Contains(t, got, "ACTIVE")
	assert.Contains(t, got, "AIRGAP")
	assert.Contains(t, got, "INACTIVE")
	assert.Contains(t, got, "TOTAL")
	assert.Contains(t, got, "paid")
	assert.Contains(t, got, "trial")
}

func TestLicenseCounts_JSON(t *testing.T) {
	counts := &channels.LicenseCounts{
		Active:   map[string]int64{"paid": 5},
		Airgap:   map[string]int64{"paid": 2},
		Inactive: map[string]int64{},
		Total:    map[string]int64{"paid": 7},
	}

	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, LicenseCounts("json", w, counts))

	var decoded channels.LicenseCounts
	require.NoError(t, json.Unmarshal(out.Bytes(), &decoded))
	assert.Equal(t, int64(5), decoded.Active["paid"])
	assert.Equal(t, int64(2), decoded.Airgap["paid"])
	assert.Equal(t, int64(7), decoded.Total["paid"])
}

func TestLicenseCounts_Empty_Table(t *testing.T) {
	counts := &channels.LicenseCounts{}

	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, LicenseCounts("table", w, counts))

	assert.Contains(t, out.String(), "No active licenses in channel")
}

func TestLicenseCounts_Empty_JSON(t *testing.T) {
	counts := &channels.LicenseCounts{}

	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, LicenseCounts("json", w, counts))

	assert.Equal(t, "{}\n", out.String())
}

func TestLicenseCounts_UnknownFormat(t *testing.T) {
	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	err := LicenseCounts("yaml", w, &channels.LicenseCounts{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown format: yaml")
}
