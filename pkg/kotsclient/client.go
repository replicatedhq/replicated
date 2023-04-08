package kotsclient

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

var (
	ErrForbidden = errors.New("the action is not allowed for the current user or team")
)

type VendorV3Client struct {
	platformclient.HTTPClient
}
