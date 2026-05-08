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

func TestKotsChannelReleases_Table(t *testing.T) {
	demoted := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	releases := []*types.ChannelRelease{
		{
			ChannelSequence: 5,
			Sequence:        12,
			Semver:          "1.2.0",
			Created:         time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			ReleasedAt:      time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			ChannelSequence: 4,
			Sequence:        11,
			Semver:          "1.1.0",
			Created:         time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			ReleasedAt:      time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC),
			IsDemoted:       true,
			DemotedAt:       &demoted,
		},
	}

	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, KotsChannelReleases("table", w, releases))

	got := out.String()
	assert.Contains(t, got, "CHANNEL_SEQUENCE")
	assert.Contains(t, got, "RELEASE_SEQUENCE")
	assert.Contains(t, got, "VERSION")
	assert.Contains(t, got, "STATE")
	assert.Contains(t, got, "1.2.0")
	assert.Contains(t, got, "active")
	assert.Contains(t, got, "demoted")
}

func TestKotsChannelReleases_JSON(t *testing.T) {
	releases := []*types.ChannelRelease{
		{
			ChannelSequence: 5,
			Sequence:        12,
			Semver:          "1.2.0",
		},
	}

	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, KotsChannelReleases("json", w, releases))

	var decoded []*types.ChannelRelease
	require.NoError(t, json.Unmarshal(out.Bytes(), &decoded))
	require.Len(t, decoded, 1)
	assert.Equal(t, "1.2.0", decoded[0].Semver)
	assert.Equal(t, int32(12), decoded[0].Sequence)

	// isDemoted and demotedAt must always appear so agents can do
	// `release.isDemoted === false` instead of seeing undefined.
	assert.Contains(t, out.String(), `"isDemoted": false`)
	assert.Contains(t, out.String(), `"demotedAt": null`)
}

func TestKotsChannelReleases_Empty(t *testing.T) {
	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, KotsChannelReleases("table", w, nil))
	assert.Contains(t, out.String(), "No releases in channel")
}
