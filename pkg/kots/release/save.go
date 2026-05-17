package release

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	releaseTypes "github.com/replicatedhq/replicated/pkg/kots/release/types"
	"github.com/replicatedhq/replicated/pkg/logger"
	"github.com/replicatedhq/replicated/pkg/types"
)

func Save(dstDir string, release *types.AppRelease, log *logger.Logger) error {
	var releaseYamls []releaseTypes.KotsSingleSpec
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

func writeReleaseFiles(dstDir string, specs []releaseTypes.KotsSingleSpec, log *logger.Logger) error {
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

func writeReleaseDirectory(dstDir string, spec releaseTypes.KotsSingleSpec, log *logger.Logger) error {
	log.ChildActionWithoutSpinner("%s", spec.Path)

	path, err := resolveReleasePath(dstDir, spec.Path)
	if err != nil {
		return err
	}

	if err := os.Mkdir(path, 0755); err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "create directory")
	}

	err = writeReleaseFiles(dstDir, spec.Children, log)
	if err != nil {
		return errors.Wrap(err, "write children")
	}

	return nil
}

func writeReleaseFile(dstDir string, spec releaseTypes.KotsSingleSpec, log *logger.Logger) error {
	log.ChildActionWithoutSpinner("%s", spec.Path)

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

	path, err := resolveReleasePath(dstDir, spec.Path)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, content, 0644)
	if err != nil {
		return errors.Wrap(err, "write file")
	}

	return nil
}

func resolveReleasePath(dstDir string, releasePath string) (string, error) {
	cleanReleasePath := filepath.Clean(releasePath)
	if filepath.IsAbs(releasePath) ||
		cleanReleasePath == "." ||
		cleanReleasePath == ".." ||
		strings.HasPrefix(cleanReleasePath, ".."+string(filepath.Separator)) {
		return "", errors.Errorf("invalid release path %q", releasePath)
	}

	cleanDstDir := filepath.Clean(dstDir)
	path := filepath.Join(cleanDstDir, cleanReleasePath)

	rel, err := filepath.Rel(cleanDstDir, path)
	if err != nil {
		return "", errors.Wrap(err, "get relative release path")
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", errors.Errorf("release path escapes destination %q", releasePath)
	}

	return path, nil
}
