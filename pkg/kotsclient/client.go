package kotsclient

import (
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

type VendorV3Client struct {
	platformclient.HTTPClient
}
