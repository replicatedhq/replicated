package print

import (
	"bytes"
	"regexp"
	"testing"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_ChannelAttrs(t *testing.T) {
	tests := []struct {
		name       string
		appType    string
		appSlug    string
		appChan    *types.Channel
		matchRegex *regexp.Regexp
		wantMatch  bool
	}{
		{
			name:    "kots channel with install commands",
			appType: "kots",
			appSlug: "myapp",
			appChan: &types.Channel{
				ID:              "123",
				Name:            "mychannel",
				Description:     "mychannel description",
				ReleaseSequence: 1,
				ReleaseLabel:    "v1.0.0",
				IsHelmOnly:      false,
			},
			matchRegex: regexp.MustCompile(`(?s).*EXISTING:.*EMBEDDED:.*AIRGAP:.*`), // (?s) mean match line breaks
			wantMatch:  true,
		},
		{
			name:    "kots channel without install commands",
			appType: "kots",
			appSlug: "myapp",
			appChan: &types.Channel{
				ID:              "123",
				Name:            "mychannel",
				Description:     "mychannel description",
				ReleaseSequence: 1,
				ReleaseLabel:    "v1.0.0",
				IsHelmOnly:      true,
			},
			matchRegex: regexp.MustCompile(`(?s).*EXISTING:.*EMBEDDED:.*AIRGAP:.*`), // (?s) mean match line breaks
			wantMatch:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			w := tabwriter.NewWriter(&out, 0, 8, 4, ' ', tabwriter.TabIndent)

			err := ChannelAttrs("text", w, tt.appType, tt.appSlug, tt.appChan)
			assert.NoError(t, err)

			w.Flush()

			if tt.wantMatch {
				assert.Regexp(t, tt.matchRegex, out.String())
			} else {
				assert.NotRegexp(t, tt.matchRegex, out.String())
			}
		})
	}
}
