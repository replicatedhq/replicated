package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gorilla/mux"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelReleases(t *testing.T) {
	tests := []struct {
		name        string
		cliArgs     []string
		appType     string // "kots" or "platform" — controls which apps endpoint returns the app
		channels    []map[string]interface{}
		releases    interface{}
		wantFormat  format
		wantLines   int
		wantOutput  string                       // when set, asserted exactly
		wantContain []string                     // substrings the output must contain
		wantQuery   map[string]string            // query params asserted on the releases request
		wantExit    int                          // expected non-zero exit; 0 means must succeed
		assertJSON  func(t *testing.T, raw []byte)
	}{
		{
			name:    "kots channel releases by ID, table, with demoted state",
			cliArgs: []string{"channel", "releases", "chan-1"},
			appType: "kots",
			channels: []map[string]interface{}{
				{"id": "chan-1", "name": "Stable", "channelSlug": "stable"},
			},
			releases: []map[string]interface{}{
				{"channelSequence": 5, "sequence": 12, "semver": "1.2.0", "isDemoted": false},
				{"channelSequence": 4, "sequence": 11, "semver": "1.1.0", "isDemoted": true, "demotedAt": "2026-05-01T00:00:00Z"},
			},
			wantFormat: FormatTable,
			wantContain: []string{
				"CHANNEL_SEQUENCE", "RELEASE_SEQUENCE", "VERSION", "STATE",
				"1.2.0", "active",
				"1.1.0", "demoted",
			},
		},
		{
			// Exercises GetChannelByName's fallthrough: GET /v3/app/.../channel/Stable
			// returns 404 (not an ID), then GET /v3/app/.../channels?channelName=Stable
			// returns the list, and we match by name.
			name:    "kots channel releases by name resolves through ListChannels",
			cliArgs: []string{"channel", "releases", "Stable"},
			appType: "kots",
			channels: []map[string]interface{}{
				{"id": "chan-1", "name": "Stable", "channelSlug": "stable"},
			},
			releases: []map[string]interface{}{
				{"channelSequence": 1, "sequence": 1, "semver": "1.0.0"},
			},
			wantFormat:  FormatTable,
			wantContain: []string{"1.0.0", "active"},
		},
		{
			name:    "--page without --page-size errors",
			cliArgs: []string{"channel", "releases", "chan-1", "--page", "2"},
			appType: "kots",
			channels: []map[string]interface{}{
				{"id": "chan-1", "name": "Stable", "channelSlug": "stable"},
			},
			releases:    []map[string]interface{}{},
			wantExit:    1,
			wantContain: []string{"--page requires --page-size"},
		},
		{
			name:    "kots channel releases JSON includes isDemoted/demotedAt",
			cliArgs: []string{"channel", "releases", "chan-1", "--output", "json"},
			appType: "kots",
			channels: []map[string]interface{}{
				{"id": "chan-1", "name": "Stable", "channelSlug": "stable"},
			},
			releases: []map[string]interface{}{
				{"channelSequence": 4, "sequence": 11, "semver": "1.1.0", "isDemoted": true, "demotedAt": "2026-05-01T00:00:00Z"},
			},
			wantFormat: FormatJSON,
			assertJSON: func(t *testing.T, raw []byte) {
				var got []*types.ChannelRelease
				require.NoError(t, json.Unmarshal(raw, &got))
				require.Len(t, got, 1)
				assert.Equal(t, "1.1.0", got[0].Semver)
				assert.True(t, got[0].IsDemoted)
				require.NotNil(t, got[0].DemotedAt)
			},
		},
		{
			name:    "kots channel releases pagination sends currentPage=0 explicitly",
			cliArgs: []string{"channel", "releases", "chan-1", "--page", "0", "--page-size", "5"},
			appType: "kots",
			channels: []map[string]interface{}{
				{"id": "chan-1", "name": "Stable", "channelSlug": "stable"},
			},
			releases: []map[string]interface{}{
				{"channelSequence": 5, "sequence": 12, "semver": "1.2.0"},
			},
			wantQuery: map[string]string{
				"currentPage": "0",
				"pageSize":    "5",
			},
			wantFormat:  FormatTable,
			wantContain: []string{"1.2.0", "active"},
		},
		{
			name:    "kots channel releases empty channel",
			cliArgs: []string{"channel", "releases", "chan-1"},
			appType: "kots",
			channels: []map[string]interface{}{
				{"id": "chan-1", "name": "Stable", "channelSlug": "stable"},
			},
			releases:    []map[string]interface{}{},
			wantFormat:  FormatTable,
			wantOutput:  "No releases in channel\n",
		},
		{
			name:    "platform channel releases table — legacy app regression guard",
			cliArgs: []string{"channel", "releases", "chan-1"},
			appType: "platform",
			releases: map[string]interface{}{
				"channel": map[string]interface{}{
					"Id":   "chan-1",
					"Name": "Stable",
				},
				"releases": []map[string]interface{}{
					{
						"channel_sequence":    2,
						"release_sequence":    42,
						"created":             "2026-04-01T00:00:00Z",
						"version":             "1.0.0",
						"required":            false,
						"airgap_build_status": "built",
						"release_notes":       "first",
					},
				},
			},
			wantFormat: FormatTable,
			wantContain: []string{
				"CHANNEL_SEQUENCE", "RELEASE_SEQUENCE", "RELEASED", "VERSION", "REQUIRED", "AIRGAP_STATUS",
				"42", "1.0.0", "built",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedQuery sync.Map
			server := setupChannelTestServer(t, tt.appType, tt.channels, tt.releases, &capturedQuery)
			defer server.Close()

			cmd := getCommand(tt.cliArgs, server)
			// Per-test HOME so the app-type cache from one case doesn't bleed into the next
			// (getCommand defaults HOME to os.TempDir(), which is shared across cases).
			cmd.Env = append(cmd.Env, "REPLICATED_APP=test-app", "HOME="+t.TempDir())

			out, err := cmd.CombinedOutput()
			if tt.wantExit != 0 {
				assert.Error(t, err, "expected non-zero exit; output:\n%s", string(out))
				for _, want := range tt.wantContain {
					assert.Contains(t, string(out), want)
				}
				return
			}
			assert.NoError(t, err, "cli failed: %s", string(out))

			if tt.wantOutput != "" {
				require.Equal(t, tt.wantOutput, string(out))
				return
			}

			if tt.wantFormat == FormatJSON && tt.assertJSON != nil {
				tt.assertJSON(t, out)
				return
			}

			for _, want := range tt.wantContain {
				assert.Contains(t, string(out), want, "missing %q in output:\n%s", want, out)
			}

			if tt.wantQuery != nil {
				for k, v := range tt.wantQuery {
					got, ok := capturedQuery.Load(k)
					assert.True(t, ok, "expected query param %q not sent", k)
					assert.Equal(t, v, got, "query param %q mismatch", k)
				}
			}
		})
	}
}

// setupChannelTestServer wires the minimum endpoints needed to drive
// `replicated channel releases` end-to-end for either appType.
func setupChannelTestServer(t *testing.T, appType string, channels []map[string]interface{}, releases interface{}, capturedQuery *sync.Map) *httptest.Server {
	r := mux.NewRouter()

	switch appType {
	case "kots":
		// /v1/apps must 404-equivalent so PlatformClient.GetApp fails and GetAppType falls through to kots.
		r.Methods(http.MethodGet).Path("/v1/apps").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		})

		r.Methods(http.MethodGet).Path("/v3/apps").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"apps":[{"id":"app-123","name":"Test App","slug":"test-app"}]}`))
		})

		// GetChannelByName first tries GetChannel by ID. Returns the channel if the
		// arg matches a known ID, else 404 to trigger the ListChannels fallthrough.
		r.Methods(http.MethodGet).Path("/v3/app/{appID}/channel/{channelID}").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			vars := mux.Vars(req)
			for _, c := range channels {
				if c["id"] == vars["channelID"] {
					body, _ := json.Marshal(map[string]interface{}{"channel": c})
					w.WriteHeader(http.StatusOK)
					w.Write(body)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"not found"}`))
		})

		// ListChannels — fallthrough path when GetChannel-by-ID returned 404 because
		// the arg was actually a name.
		r.Methods(http.MethodGet).Path("/v3/app/{appID}/channels").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			body, _ := json.Marshal(map[string]interface{}{"channels": channels})
			w.WriteHeader(http.StatusOK)
			w.Write(body)
		})

		r.Methods(http.MethodGet).Path("/v3/app/{appID}/channel/{channelID}/releases").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			for k, v := range req.URL.Query() {
				if len(v) > 0 {
					capturedQuery.Store(k, v[0])
				}
			}
			body, _ := json.Marshal(map[string]interface{}{"releases": releases})
			w.WriteHeader(http.StatusOK)
			w.Write(body)
		})

	case "platform":
		r.Methods(http.MethodGet).Path("/v1/apps").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"app":{"Id":"app-123","Name":"Test App","Slug":"test-app","Scheduler":"native"}}]`))
		})

		// Platform GetChannel returns channel + releases in one shot. Hit twice per command
		// (once for GetChannelByName, once by the command itself); same handler serves both.
		r.Methods(http.MethodGet).Path("/v1/app/{appID}/channel/{channelID}/releases").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			body, _ := json.Marshal(releases)
			w.WriteHeader(http.StatusOK)
			w.Write(body)
		})

	default:
		t.Fatalf("unknown appType %q", appType)
	}

	// Catch-all to surface unexpected calls in test output.
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Logf("unexpected %s %s?%s", req.Method, req.URL.Path, req.URL.RawQuery)
		http.NotFound(w, req)
	})

	return httptest.NewServer(r)
}
