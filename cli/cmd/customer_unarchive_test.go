package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestCustomerUnarchiveMissingArgs(t *testing.T) {
	r := &runners{appID: "app-id", appType: "kots"}
	parent := r.InitCustomersCommand(&cobra.Command{Use: "replicated"})
	cmd := r.InitCustomersUnarchiveCommand(parent)
	require.EqualError(t, cmd.Args(cmd, nil), "requires at least 1 arg(s), only received 0")
}
