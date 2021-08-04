package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitReleaseDownload(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "download SEQUENCE",
		Short: "Download the YAML config for a release. Same as 'release inspect' for non-KOTS apps",
		Long:  "Download the YAML config for a release. Same as 'release inspect' for non-KOTS apps",
	}
	parent.AddCommand(cmd)
	cmd.RunE = r.releaseDownload
	cmd.Flags().StringVarP(&r.args.releaseDownloadDest, "dest", "d", "", "Directory to which release manifests should be downloaded")
}

func (r *runners) releaseDownload(command *cobra.Command, args []string) error {
	if r.appType != "kots" {
		return r.releaseInspect(command, args)
	}

	if len(args) != 1 {
		return errors.New("release sequence is required")
	}
	seq, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Failed to parse sequence argument %q", args[0])
	}

	if r.args.releaseDownloadDest == "" {
		return errors.New("Downloading a release for a KOTS application requires a --dest directory to unpack the manifests, e.g. \"./manifests\"")
	}

	log := print.NewLogger(os.Stdout)
	log.ActionWithSpinner("Fetching Release %d", seq)
	release, err := r.api.GetRelease(r.appID, r.appType, seq)
	if err != nil {
		log.FinishSpinnerWithError()
		return errors.Wrap(err, "get release")
	}
	log.FinishSpinner()

	log.ActionWithoutSpinner("Writing files to %s", r.args.releaseDownloadDest)
	var releaseYamls []kotsSingleSpec
	err = json.Unmarshal([]byte(release.Config), &releaseYamls)
	if err != nil {
		return errors.Wrap(err, "unmarshal release yamls")
	}

	err = os.MkdirAll(r.args.releaseDownloadDest, 0755)
	if err != nil {
		return errors.Wrapf(err, "create dir %q", r.args.releaseDownloadDest)
	}

	for _, releaseYaml := range releaseYamls {
		path := filepath.Join(r.args.releaseDownloadDest, releaseYaml.Path)
		log.ChildActionWithoutSpinner(releaseYaml.Path)

		var content []byte

		ext := filepath.Ext(releaseYaml.Path)
		switch ext {
		case ".tgz", ".gz":
			decoded, err := base64.StdEncoding.DecodeString(releaseYaml.Content)
			if err == nil {
				content = decoded
			} else {
				content = []byte(releaseYaml.Content)
			}
		default:
			content = []byte(releaseYaml.Content)
		}

		err := ioutil.WriteFile(path, content, 0644)
		if err != nil {
			return errors.Wrapf(err, "write file %q", path)
		}
	}

	return nil

}
