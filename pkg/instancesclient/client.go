package instancesclient

import (
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

type InstancesClient struct {
	platformclient.HTTPClient
}
