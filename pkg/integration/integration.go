package integration

import (
	"fmt"
	"path"

	kotsclienttypes "github.com/replicatedhq/replicated/pkg/kotsclient/types"
	"github.com/replicatedhq/replicated/pkg/types"
)

type contextKey string

const IntegrationTestContextKey contextKey = "integration-test"
const APICallLogContextKey contextKey = "api-call-log"

type format string

const (
	FormatJSON  format = "json"
	FormatTable format = "table"
)

var (
	typeKOTSAppWithChannels = types.KotsAppWithChannels{
		Channels: []types.Channel{
			{
				ID:   "channel-id",
				Name: "channel-name",
			},
		},
		Slug:         "slug",
		Name:         "name",
		IsFoundation: true,
	}
)

func Response(key string) interface{} {
	switch key {
	case "app-ls-empty":
		return kotsclienttypes.KotsAppResponse{
			Apps: []types.KotsAppWithChannels{},
		}
	case "app-ls-single":
		return kotsclienttypes.KotsAppResponse{
			Apps: []types.KotsAppWithChannels{
				typeKOTSAppWithChannels,
			},
		}
	default:
		panic(fmt.Sprintf("unknown integration test: %s", key))
	}

}

func CLIPath() string {
	return path.Join("..", "..", "bin", "replicated")
}
