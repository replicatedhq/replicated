package util

import (
	"net/http"
	"os"
)

func UploadFile(fullpath string, url string) error {
	file, err := os.Open(fullpath)
	if err != nil {
		return err
	}
	defer file.Close()

	req, err := http.NewRequest("PUT", url, file)
	if err != nil {
		return err
	}

	fi, err := os.Stat(fullpath)
	if err != nil {
		return err
	}
	req.ContentLength = fi.Size()

	req.Header.Set("Content-Type", "application/x-yaml")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return err
	}

	return nil
}
