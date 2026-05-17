package release

import (
	_ "embed"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	releaseTypes "github.com/replicatedhq/replicated/pkg/kots/release/types"
	"github.com/replicatedhq/replicated/pkg/logger"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/save/input/release.json
var releaseToSave []byte

func Test_Save(t *testing.T) {
	// Just one test case for now

	t.Run("compare Save() output", func(t *testing.T) {
		dstDir, err := os.MkdirTemp("", "kots-release-save-test")
		require.NoError(t, err)

		defer os.RemoveAll(dstDir)

		release := &types.AppRelease{
			Config: string(releaseToSave),
		}
		err = Save(dstDir, release, logger.NewLogger(os.Stdout))
		require.NoError(t, err)

		wantResultsDir := "./testdata/save/want"
		err = filepath.Walk(wantResultsDir,
			func(wantPath string, wantInfo os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				relPath, err := filepath.Rel(wantResultsDir, wantPath)
				if err != nil {
					return errors.Wrap(err, "get rel path")
				}

				if relPath == "." {
					return nil
				}

				gotPath := filepath.Join(dstDir, relPath)
				gotInfo, err := os.Stat(gotPath)
				if err != nil {
					return errors.Wrap(err, "stat want path")
				}

				assert.Equal(t, wantInfo.IsDir(), gotInfo.IsDir(), "is dir", wantPath)

				return nil
			})

		require.NoError(t, err)

		err = filepath.Walk(dstDir,
			func(gotPath string, gotInfo os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				relPath, err := filepath.Rel(dstDir, gotPath)
				if err != nil {
					return errors.Wrap(err, "get rel path")
				}

				if relPath == "." {
					return nil
				}

				wantPath := filepath.Join(wantResultsDir, relPath)
				wantInfo, err := os.Stat(wantPath)
				if err != nil {
					return errors.Wrap(err, "stat want path")
				}

				assert.Equal(t, wantInfo.IsDir(), gotInfo.IsDir(), "is dir", wantPath)

				if gotInfo.IsDir() {
					return nil
				}

				gotContents, err := ioutil.ReadFile(gotPath)
				if err != nil {
					return errors.Wrap(err, "read got file")
				}

				wantContents, err := ioutil.ReadFile(wantPath)
				if err != nil {
					return errors.Wrap(err, "read want file")
				}

				contentsString := string(gotContents)
				wantContentsString := string(wantContents)
				assert.Equal(t, wantContentsString, contentsString, wantPath)

				return nil
			})

		require.NoError(t, err)
	})
}

func Test_SaveRejectsEscapingPaths(t *testing.T) {
	tests := []struct {
		name       string
		path       func(string) string
		escapePath func(string) string
	}{
		{
			name: "parent traversal",
			path: func(string) string {
				return "../escape.yaml"
			},
			escapePath: func(dstDir string) string {
				return filepath.Join(dstDir, "..", "escape.yaml")
			},
		},
		{
			name: "absolute path",
			path: func(dstDir string) string {
				return filepath.Join(filepath.Dir(dstDir), filepath.Base(dstDir)+"-absolute-escape.yaml")
			},
			escapePath: func(dstDir string) string {
				return filepath.Join(filepath.Dir(dstDir), filepath.Base(dstDir)+"-absolute-escape.yaml")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dstDir, err := os.MkdirTemp("", "kots-release-save-test")
			require.NoError(t, err)
			defer os.RemoveAll(dstDir)

			escapePath := tt.escapePath(dstDir)
			_ = os.Remove(escapePath)

			config, err := json.Marshal([]releaseTypes.KotsSingleSpec{
				{
					Path:    tt.path(dstDir),
					Content: "owned",
				},
			})
			require.NoError(t, err)

			release := &types.AppRelease{
				Config: string(config),
			}
			err = Save(dstDir, release, logger.NewLogger(os.Stdout))
			require.Error(t, err)

			_, err = os.Stat(escapePath)
			require.True(t, os.IsNotExist(err), "expected %s not to be created", escapePath)
		})
	}
}
