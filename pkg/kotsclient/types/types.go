package types

import "github.com/replicatedhq/replicated/pkg/types"

type KotsAppResponse struct {
	Apps []types.KotsAppWithChannels `json:"apps"`
}
