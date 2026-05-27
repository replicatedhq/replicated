package print

import (
	"bytes"
	"encoding/json"
	"testing"
	"text/tabwriter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelImages_Table(t *testing.T) {
	images := []string{
		"nginx:1.27",
		"postgres:14",
	}

	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, ChannelImages("table", w, images))

	got := out.String()
	assert.Contains(t, got, "IMAGE")
	assert.Contains(t, got, "nginx:1.27")
	assert.Contains(t, got, "postgres:14")
}

func TestChannelImages_JSON(t *testing.T) {
	images := []string{
		"nginx:1.27",
		"postgres:14",
	}

	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, ChannelImages("json", w, images))

	var decoded []string
	require.NoError(t, json.Unmarshal(out.Bytes(), &decoded))
	assert.Equal(t, images, decoded)
}

func TestChannelImages_Empty_JSON(t *testing.T) {
	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	require.NoError(t, ChannelImages("json", w, []string{}))

	var decoded []string
	require.NoError(t, json.Unmarshal(out.Bytes(), &decoded))
	assert.Empty(t, decoded)
}

func TestChannelImages_UnknownFormat(t *testing.T) {
	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)
	err := ChannelImages("yaml", w, []string{"nginx:1.27"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown format: yaml")
}
