package integration

import (
	"fmt"
	"path"
	"time"

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
		Id:           "id",
		Slug:         "slug",
		Name:         "name",
		IsFoundation: true,
	}

	typesKOTSAppRelease = types.KotsAppRelease{
		AppID:                "app-id",
		Sequence:             1,
		CreatedAt:            time.Now(),
		IsArchived:           false,
		Spec:                 "spec",
		ReleaseNotes:         "release-notes",
		IsReleaseNotEditable: false,
		Channels: []*types.Channel{
			{
				ID:   "channel-id",
				Name: "channel-name",
			},
		},
		Charts:               []types.Chart{},
		CompatibilityResults: []types.CompatibilityResult{},
		IsHelmOnly:           false,
	}
)

func Response(key string) interface{} {
	switch key {
	case "app-ls-empty":
		return &kotsclienttypes.KotsAppResponse{
			Apps: []types.KotsAppWithChannels{},
		}
	case "app-ls-single", "app-rm":
		return &kotsclienttypes.KotsAppResponse{
			Apps: []types.KotsAppWithChannels{
				typeKOTSAppWithChannels,
			},
		}
	case "app-create":
		return &kotsclienttypes.CreateKOTSAppResponse{
			App: &typeKOTSAppWithChannels,
		}
	case "release-ls":
		return kotsclienttypes.KotsListReleasesResponse{
			Releases: []*types.KotsAppRelease{
				&typesKOTSAppRelease,
			},
		}
	default:
		panic(fmt.Sprintf("unknown integration test: %s", key))
	}
}

func CLIPath() string {
	return path.Join("..", "..", "bin", "replicated")
}
