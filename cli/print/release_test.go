package print

import (
	"bytes"
	"encoding/json"
	"testing"
	"text/tabwriter"
	"time"

	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelease_Table_RendersChannels(t *testing.T) {
	release := &types.AppRelease{
		Sequence:  42,
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EditedAt:  time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		Channels: []*types.Channel{
			{ID: "c1", Name: "Stable"},
			{ID: "c2", Name: "Beta"},
		},
		Config: "spec: yaml",
	}

	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, Release("table", w, release))

	got := out.String()
	assert.Contains(t, got, "CHANNELS:")
	assert.Contains(t, got, "Stable")
	assert.Contains(t, got, "Beta")
}

func TestRelease_Table_NoChannelsSection_WhenEmpty(t *testing.T) {
	release := &types.AppRelease{Sequence: 1, Config: "x"}
	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, Release("table", w, release))
	assert.NotContains(t, out.String(), "CHANNELS:")
}

func TestRelease_JSON_IncludesChannels(t *testing.T) {
	release := &types.AppRelease{
		Sequence: 7,
		Channels: []*types.Channel{{ID: "c1", Name: "Stable"}},
	}

	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, Release("json", w, release))

	var decoded types.AppRelease
	require.NoError(t, json.Unmarshal(out.Bytes(), &decoded))
	require.Len(t, decoded.Channels, 1)
	assert.Equal(t, "Stable", decoded.Channels[0].Name)
}
