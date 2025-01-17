package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/pkg/errors"
	kotsrelease "github.com/replicatedhq/replicated/pkg/kots/release"
	"github.com/replicatedhq/replicated/pkg/logger"
	"github.com/spf13/cobra"
)

func (r *runners) InitReleaseDownload(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "download RELEASE_SEQUENCE",
		Short: "Download application manifests for a release.",
		Long: `Download application manifests for a release to a specified directory.

For non-KOTS applications, this is equivalent to the 'release inspect' command.

Example:
replicated release download 1 --dest ./manifests`,
		Args: cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)
	cmd.RunE = r.releaseDownload
	cmd.Flags().StringVarP(&r.args.releaseDownloadDest, "dest", "d", "", "Directory to which release manifests should be downloaded")
}

func (r *runners) releaseDownload(command *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if r.appType != "kots" {
		return r.releaseInspect(command, args)
	}

	seq, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Failed to parse sequence argument %q", args[0])
	}

	if r.args.releaseDownloadDest == "" {
		return errors.New("Downloading a release for a KOTS application requires a --dest directory to unpack the manifests, e.g. \"./manifests\"")
	}

	log := logger.NewLogger(os.Stdout)
	log.ActionWithSpinner("Fetching Release %d", seq)
	release, err := r.api.GetRelease(r.appID, r.appType, seq)
	if err != nil {
		log.FinishSpinnerWithError()
		return errors.Wrap(err, "get release")
	}
	log.FinishSpinner()

	log.ActionWithoutSpinner("Writing files to %s", r.args.releaseDownloadDest)

	err = kotsrelease.Save(r.args.releaseDownloadDest, release, log)
	if err != nil {
		return errors.Wrap(err, "save release")
	}

	return nil

}
