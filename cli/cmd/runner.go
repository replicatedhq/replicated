package cmd

import (
	"io"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/client"
)

// Runner holds the I/O dependencies and configurations used by individual
// commands, which are defined as methods on this type.
type runners struct {
	appID string
	api   client.Client
	stdin io.Reader
	w     *tabwriter.Writer
}
