package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReleaseList(t *testing.T) {
	tests := []struct {
		name          string
		cliArgs       []string
		appType       string // "kots" or "platform"
		totalReleases int
		failPage      int // kots page that returns 500; -1 for none
		wantExit      bool
		wantSequences []int64  // with --output json, the sequences printed, in order
		wantLines     int      // table output line count, including the header
		wantRequests  []string // kots release list queries, in order
		wantStdout    []string // substrings stdout must contain
		wantStderr    []string // substrings stderr must contain
	}{
		{
			name:          "kots pages at 20 until an empty page",
			cliArgs:       []string{"release", "ls"},
			appType:       "kots",
			totalReleases: 45,
			failPage:      -1,
			wantLines:     46,
			wantRequests: []string{
				"currentPage=0&pageSize=20",
				"currentPage=1&pageSize=20",
				"currentPage=2&pageSize=20",
				"currentPage=3&pageSize=20",
			},
			wantStdout: []string{"SEQUENCE", "CREATED", "EDITED", "ACTIVE_CHANNELS"},
		},
		{
			name:          "kots --page-size fetches only the first page",
			cliArgs:       []string{"release", "ls", "--page-size", "5", "--output", "json"},
			appType:       "kots",
			totalReleases: 45,
			failPage:      -1,
			wantSequences: []int64{45, 44, 43, 42, 41},
			wantRequests:  []string{"currentPage=0&pageSize=5"},
		},
		{
			name:          "kots --page and --page-size fetch one page",
			cliArgs:       []string{"release", "ls", "--page", "1", "--page-size", "50", "--output", "json"},
			appType:       "kots",
			totalReleases: 55,
			failPage:      -1,
			wantSequences: []int64{5, 4, 3, 2, 1},
			wantRequests:  []string{"currentPage=1&pageSize=50"},
		},
		{
			name:          "kots --limit below 20 is used as the page size",
			cliArgs:       []string{"release", "ls", "--limit", "5", "--output", "json"},
			appType:       "kots",
			totalReleases: 45,
			failPage:      -1,
			wantSequences: []int64{45, 44, 43, 42, 41},
			wantRequests:  []string{"currentPage=0&pageSize=5"},
		},
		{
			name:          "kots --limit above 20 stops paging once enough releases are fetched",
			cliArgs:       []string{"release", "ls", "--limit", "25"},
			appType:       "kots",
			totalReleases: 45,
			failPage:      -1,
			wantLines:     26,
			wantRequests: []string{
				"currentPage=0&pageSize=20",
				"currentPage=1&pageSize=20",
			},
		},
		{
			name:          "kots --limit larger than the total lists every release",
			cliArgs:       []string{"release", "ls", "--limit", "50"},
			appType:       "kots",
			totalReleases: 30,
			failPage:      -1,
			wantLines:     31,
			wantRequests: []string{
				"currentPage=0&pageSize=20",
				"currentPage=1&pageSize=20",
				"currentPage=2&pageSize=20",
			},
		},
		{
			name:          "kots --limit with --page-size pages at that size",
			cliArgs:       []string{"release", "ls", "--limit", "120", "--page-size", "50"},
			appType:       "kots",
			totalReleases: 200,
			failPage:      -1,
			wantLines:     121,
			wantRequests: []string{
				"currentPage=0&pageSize=50",
				"currentPage=1&pageSize=50",
				"currentPage=2&pageSize=50",
			},
		},
		{
			name:          "kots --page-size above the API maximum surfaces the API error",
			cliArgs:       []string{"release", "ls", "--page-size", "500"},
			appType:       "kots",
			totalReleases: 45,
			failPage:      -1,
			wantExit:      true,
			wantRequests:  []string{"currentPage=0&pageSize=500"},
			wantStderr:    []string{"400: pageSize must be at most 100"},
		},
		{
			name:          "--page with --limit is rejected before listing",
			cliArgs:       []string{"release", "ls", "--page", "0", "--page-size", "20", "--limit", "5"},
			appType:       "kots",
			totalReleases: 45,
			failPage:      -1,
			wantExit:      true,
			wantRequests:  []string{},
			wantStderr:    []string{"--page cannot be used with --limit"},
		},
		{
			name:          "negative --limit is rejected before listing",
			cliArgs:       []string{"release", "ls", "--limit", "-1"},
			appType:       "kots",
			totalReleases: 45,
			failPage:      -1,
			wantExit:      true,
			wantRequests:  []string{},
			wantStderr:    []string{"--limit must be 0 or greater"},
		},
		{
			name:          "platform --limit truncates the list",
			cliArgs:       []string{"release", "ls", "--limit", "2", "--output", "json"},
			appType:       "platform",
			totalReleases: 3,
			failPage:      -1,
			wantSequences: []int64{3, 2},
		},
		{
			name:          "kots app with no releases",
			cliArgs:       []string{"release", "ls", "--output", "json"},
			appType:       "kots",
			totalReleases: 0,
			failPage:      -1,
			wantSequences: []int64{},
			wantRequests:  []string{"currentPage=0&pageSize=20"},
		},
		{
			name:          "kots failed page exits non-zero without a partial list",
			cliArgs:       []string{"release", "ls"},
			appType:       "kots",
			totalReleases: 45,
			failPage:      1,
			wantExit:      true,
			wantRequests: []string{
				"currentPage=0&pageSize=20",
				"currentPage=1&pageSize=20",
			},
			wantStderr: []string{"page failed"},
		},
		{
			name:          "kots failed single page exits non-zero",
			cliArgs:       []string{"release", "ls", "--page", "1", "--page-size", "20"},
			appType:       "kots",
			totalReleases: 45,
			failPage:      1,
			wantExit:      true,
			wantRequests:  []string{"currentPage=1&pageSize=20"},
			wantStderr:    []string{"page failed"},
		},
		{
			name:          "--page without --page-size is left to the API",
			cliArgs:       []string{"release", "ls", "--page", "1", "--output", "json"},
			appType:       "kots",
			totalReleases: 45,
			failPage:      -1,
			wantSequences: []int64{},
			wantRequests:  []string{"currentPage=1&pageSize=0"},
		},
		{
			name:          "--debug logs each page fetched",
			cliArgs:       []string{"release", "ls", "--debug"},
			appType:       "kots",
			totalReleases: 25,
			failPage:      -1,
			wantLines:     26,
			wantRequests: []string{
				"currentPage=0&pageSize=20",
				"currentPage=1&pageSize=20",
				"currentPage=2&pageSize=20",
			},
			wantStderr: []string{
				"[DEBUG] Fetching releases page 0 (0 releases so far)\n",
				"[DEBUG] Fetching releases page 1 (20 releases so far)\n",
				"[DEBUG] Fetching releases page 2 (25 releases so far)\n",
			},
		},
		{
			name:          "platform lists every release",
			cliArgs:       []string{"release", "ls", "--output", "json"},
			appType:       "platform",
			totalReleases: 3,
			failPage:      -1,
			wantSequences: []int64{3, 2, 1},
		},
		{
			name:          "platform rejects --page-size",
			cliArgs:       []string{"release", "ls", "--page-size", "2"},
			appType:       "platform",
			totalReleases: 3,
			failPage:      -1,
			wantExit:      true,
			wantStderr:    []string{"not currently supported for Platform applications"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mu sync.Mutex
			gotRequests := []string{}
			server := setupReleaseListTestServer(t, tt.appType, tt.totalReleases, tt.failPage, func(query string) {
				mu.Lock()
				defer mu.Unlock()
				gotRequests = append(gotRequests, query)
			})
			defer server.Close()

			cmd := getCommand(tt.cliArgs, server)
			// Per-test HOME so the app-type cache from one case doesn't bleed into the next.
			cmd.Env = append(cmd.Env, "REPLICATED_APP=test-app", "HOME="+t.TempDir())
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if tt.wantExit {
				require.Error(t, err, "expected non-zero exit; stdout:\n%s", stdout.String())
				assert.NotContains(t, stdout.String(), "SEQUENCE", "partial list printed")
			} else {
				require.NoError(t, err, "cli failed; stderr:\n%s", stderr.String())
			}

			if tt.appType == "kots" {
				assert.Equal(t, tt.wantRequests, gotRequests)
			}
			for _, want := range tt.wantStdout {
				assert.Contains(t, stdout.String(), want)
			}
			for _, want := range tt.wantStderr {
				assert.Contains(t, stderr.String(), want)
			}

			if tt.wantSequences != nil {
				releases := []struct {
					Sequence int64 `json:"sequence"`
				}{}
				require.NoError(t, json.Unmarshal(stdout.Bytes(), &releases), stdout.String())
				gotSequences := []int64{}
				for _, r := range releases {
					gotSequences = append(gotSequences, r.Sequence)
				}
				assert.Equal(t, tt.wantSequences, gotSequences)
			}
			if tt.wantLines > 0 {
				AssertCLIOutput(t, stdout.String(), FormatTable, tt.wantLines)
			}
		})
	}
}

// setupReleaseListTestServer serves totalReleases releases, newest first, for
// `replicated release ls` on either appType. Kots releases are paginated by the
// currentPage and pageSize query params, and each kots query is passed to onList.
func setupReleaseListTestServer(t *testing.T, appType string, totalReleases int, failPage int, onList func(query string)) *httptest.Server {
	r := mux.NewRouter()

	switch appType {
	case "kots":
		// /v1/apps returns nothing so GetAppType falls through to kots.
		r.Methods(http.MethodGet).Path("/v1/apps").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		})

		r.Methods(http.MethodGet).Path("/v3/apps").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"apps":[{"id":"app-123","name":"Test App","slug":"test-app"}]}`))
		})

		r.Methods(http.MethodGet).Path("/v3/app/app-123/releases").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			onList(req.URL.RawQuery)
			page, err := strconv.Atoi(req.URL.Query().Get("currentPage"))
			assert.NoError(t, err)
			pageSize, err := strconv.Atoi(req.URL.Query().Get("pageSize"))
			assert.NoError(t, err)

			if pageSize > 100 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"pageSize must be at most 100"}`))
				return
			}
			if page == failPage {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"page failed"}`))
				return
			}

			releases := []map[string]any{}
			for i := page * pageSize; i < (page+1)*pageSize && i < totalReleases; i++ {
				releases = append(releases, map[string]any{
					"appId":     "app-123",
					"sequence":  totalReleases - i,
					"createdAt": "2026-11-01T00:00:00Z",
				})
			}
			body, _ := json.Marshal(map[string]any{"releases": releases})
			w.WriteHeader(http.StatusOK)
			w.Write(body)
		})

	case "platform":
		r.Methods(http.MethodGet).Path("/v1/apps").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"app":{"Id":"app-123","Name":"Test App","Slug":"test-app","Scheduler":"native"}}]`))
		})

		r.Methods(http.MethodGet).Path("/v1/app/app-123/releases").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			releases := []string{}
			for i := totalReleases; i > 0; i-- {
				releases = append(releases, fmt.Sprintf(`{"AppId":"app-123","Sequence":%d}`, i))
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[" + strings.Join(releases, ",") + "]"))
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
