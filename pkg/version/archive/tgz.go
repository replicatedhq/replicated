package archive

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// CreateTGZFileFromDir will archive the contents of dir into
// a tgz file, and the caller must delete
func CreateTGZFileFromDir(dir string) (string, error) {
	tmpFile, err := ioutil.TempFile("", "replicated-update")
	if err != nil {
		return "", errors.Wrap(err, "create temp file")
	}

	file, err := os.Create(tmpFile.Name())
	if err != nil {
		return "", errors.Wrap(err, "create file")
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// walk path
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		header.Name = strings.TrimLeft(strings.TrimPrefix(path, dir), string(filepath.Separator))
		header.Size = info.Size()
		header.Mode = int64(info.Mode())
		header.ModTime = info.ModTime()

		singleFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer singleFile.Close()

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if _, err := io.Copy(tw, singleFile); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}
