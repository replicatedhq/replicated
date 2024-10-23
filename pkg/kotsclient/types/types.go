package types

import "github.com/replicatedhq/replicated/pkg/types"

type KotsAppResponse struct {
	Apps []types.KotsAppWithChannels `json:"apps"`
}

type CreateKOTSAppResponse struct {
	App *types.KotsAppWithChannels `json:"app"`
}
