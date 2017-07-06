package cmd

import (
	"text/tabwriter"

	"github.com/replicatedhq/replicated/client"
)

// Runner holds the I/O dependencies and configurations used by individual
// commands, which are defined as methods on this type.
type runners struct {
	api   client.Client
	w     *tabwriter.Writer
	appID string
}
