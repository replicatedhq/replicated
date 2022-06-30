package release

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	releases "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/kots/release/types"
	"github.com/replicatedhq/replicated/pkg/logger"
)

func Save(dstDir string, release *releases.AppRelease, log *logger.Logger) error {
	var releaseYamls []types.KotsSingleSpec
	if err := json.Unmarshal([]byte(release.Config), &releaseYamls); err != nil {
		return errors.Wrap(err, "unmarshal release yamls")
	}

	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return errors.Wrapf(err, "create dir %q", dstDir)
	}

	if err := writeReleaseFiles(dstDir, releaseYamls, log); err != nil {
		return errors.Wrap(err, "write release files")
	}

	return nil

}

func writeReleaseFiles(dstDir string, specs []types.KotsSingleSpec, log *logger.Logger) error {
	for _, spec := range specs {
		if len(spec.Children) > 0 {
			err := writeReleaseDirectory(dstDir, spec, log)
			if err != nil {
				return errors.Wrapf(err, "write direcotry %s", filepath.Join(dstDir, spec.Path))
			}
		} else {
			err := writeReleaseFile(dstDir, spec, log)
			if err != nil {
				return errors.Wrapf(err, "write file %s", filepath.Join(dstDir, spec.Path))
			}
		}
	}

	return nil
}

func writeReleaseDirectory(dstDir string, spec types.KotsSingleSpec, log *logger.Logger) error {
	log.ChildActionWithoutSpinner(spec.Path)

	if err := os.Mkdir(filepath.Join(dstDir, spec.Path), 0755); err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "create directory")
	}

	err := writeReleaseFiles(dstDir, spec.Children, log)
	if err != nil {
		return errors.Wrap(err, "write children")
	}

	return nil
}

func writeReleaseFile(dstDir string, spec types.KotsSingleSpec, log *logger.Logger) error {
	log.ChildActionWithoutSpinner(spec.Path)

	var content []byte

	ext := filepath.Ext(spec.Path)
	switch ext {
	case ".tgz", ".gz":
		decoded, err := base64.StdEncoding.DecodeString(spec.Content)
		if err != nil {
			content = []byte(spec.Content)
		} else {
			content = decoded
		}
	default:
		content = []byte(spec.Content)
	}

	err := ioutil.WriteFile(filepath.Join(dstDir, spec.Path), content, 0644)
	if err != nil {
		return errors.Wrap(err, "write file")
	}

	return nil
}
