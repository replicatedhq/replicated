package release

import (
	_ "embed"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	releases "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/logger"
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

		release := &releases.AppRelease{
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
