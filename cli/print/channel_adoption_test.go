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

func TestChannelAdoption_Table(t *testing.T) {
	adoption := &channels.ChannelAdoption{
		CurrentVersionCountActive:  map[string]int64{"paid": 5, "trial": 3},
		CurrentVersionCountAll:     map[string]int64{"paid": 10, "trial": 6},
		PreviousVersionCountActive: map[string]int64{"paid": 2},
		PreviousVersionCountAll:    map[string]int64{"paid": 4},
		OtherVersionCountActive:    map[string]int64{"trial": 1},
		OtherVersionCountAll:       map[string]int64{"trial": 2},
	}

	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, ChannelAdoption("table", w, adoption))

	got := out.String()
	assert.Contains(t, got, "LICENSE_TYPE")
	assert.Contains(t, got, "CURRENT")
	assert.Contains(t, got, "PREVIOUS")
	assert.Contains(t, got, "OTHER")
	assert.Contains(t, got, "paid")
	assert.Contains(t, got, "trial")
}

func TestChannelAdoption_JSON(t *testing.T) {
	adoption := &channels.ChannelAdoption{
		CurrentVersionCountActive:  map[string]int64{"paid": 5},
		CurrentVersionCountAll:     map[string]int64{"paid": 10},
		PreviousVersionCountActive: map[string]int64{},
		PreviousVersionCountAll:    map[string]int64{},
		OtherVersionCountActive:    map[string]int64{},
		OtherVersionCountAll:       map[string]int64{},
	}

	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, ChannelAdoption("json", w, adoption))

	var decoded channels.ChannelAdoption
	require.NoError(t, json.Unmarshal(out.Bytes(), &decoded))
	assert.Equal(t, int64(5), decoded.CurrentVersionCountActive["paid"])
	assert.Equal(t, int64(10), decoded.CurrentVersionCountAll["paid"])
}

func TestChannelAdoption_Empty_Table(t *testing.T) {
	adoption := &channels.ChannelAdoption{}

	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, ChannelAdoption("table", w, adoption))

	assert.Contains(t, out.String(), "No active licenses in channel")
}

func TestChannelAdoption_Empty_JSON(t *testing.T) {
	adoption := &channels.ChannelAdoption{}

	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, ChannelAdoption("json", w, adoption))

	assert.Equal(t, "{}\n", out.String())
}

func TestChannelAdoption_UnknownFormat(t *testing.T) {
	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	err := ChannelAdoption("yaml", w, &channels.ChannelAdoption{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown format: yaml")
}
