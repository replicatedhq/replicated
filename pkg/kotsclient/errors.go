package kotsclient

import "github.com/replicatedhq/replicated/pkg/platformclient"

var (
	// ErrNotFound is the error returned for a 404 response from the API.
	ErrNotFound = platformclient.ErrNotFound
)
