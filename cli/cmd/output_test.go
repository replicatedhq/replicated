package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestResolveOutputFormat_ExplicitFlagWins(t *testing.T) {
	// Set env var
	t.Setenv("REPLICATED_OUTPUT", "json")

	r := &runners{outputFormat: "table"}
	cmd := &cobra.Command{}
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "")
	cmd.Flags().Set("output", "wide")

	r.resolveOutputFormat(cmd)
	require.Equal(t, "wide", r.outputFormat)
}

func TestResolveOutputFormat_EnvVarWinsOverDefault(t *testing.T) {
	t.Setenv("REPLICATED_OUTPUT", "json")

	r := &runners{outputFormat: "table"}
	cmd := &cobra.Command{}
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "")

	r.resolveOutputFormat(cmd)
	require.Equal(t, "json", r.outputFormat)
}

func TestResolveOutputFormat_DefaultTable(t *testing.T) {
	r := &runners{outputFormat: "table"}
	cmd := &cobra.Command{}
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "")

	r.resolveOutputFormat(cmd)
	require.Equal(t, "table", r.outputFormat)
}
