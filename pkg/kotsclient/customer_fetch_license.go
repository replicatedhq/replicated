package kotsclient

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)


func (c *HybridClient) FetchLicense(appSlug string, customerID string) ([]byte, error) {
	c.doRawHTTP("GET", fmt.Sprintf("/kots/license/download/%s/%s"))

}
