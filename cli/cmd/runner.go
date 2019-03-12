package cmd

import (
	"io"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/pkg/shipclient"

	"github.com/replicatedhq/replicated/client"
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

// Runner holds the I/O dependencies and configurations used by individual
// commands, which are defined as methods on this type.
type runners struct {
	appID       string
	appType     string
	api         client.Client
	platformAPI platformclient.Client
	shipAPI     shipclient.Client
	stdin       io.Reader
	w           *tabwriter.Writer
}
