package cmd

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	kotsrelease "github.com/replicatedhq/replicated/pkg/kots/release"
	"github.com/replicatedhq/replicated/pkg/logger"
	"github.com/replicatedhq/replicated/pkg/tools"
	"github.com/spf13/cobra"
)

func (r *runners) InitReleaseDownload(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "download [RELEASE_SEQUENCE]",
		Short:        "Download application manifests for a release.",
		SilenceUsage: true,
		Long: `Download application manifests for a release to a specified file or directory.

For KOTS applications:
  - Downloads release as a .tgz file if no RELEASE_SEQUENCE specified
  - Can specify --channel to download the current release from that channel
  - Auto-generates filename as app-slug.tgz if --dest not provided

For non-KOTS applications, this is equivalent to the 'release inspect' command.

If no app is specified via --app flag, the app slug will be loaded from the .replicated config file.`,
		Example: `# Download latest release as autoci.tgz
replicated release download

# Download specific sequence
replicated release download 42 --dest my-release.tgz

# Download current release from Unstable channel
replicated release download --channel Unstable

# Download to directory (KOTS only with sequence)
replicated release download 1 --dest ./manifests`,
		Args:    cobra.MaximumNArgs(1),
	}
	parent.AddCommand(cmd)
	
	// Similar to release create, handle config-based flow in PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// Check if --app flag was explicitly provided by the user
		appFlagProvided := cmd.Flags().Changed("app")
		
		// Check if we should use config-based flow (no --app flag was provided)
		// Note: Parent's PersistentPreRunE runs BEFORE our PreRunE, so appID/appSlug 
		// may already be set from cache/env even if user didn't provide --app flag
		usingConfigFlow := false
		if !appFlagProvided {
			parser := tools.NewConfigParser()
			config, err := parser.FindAndParseConfig(".")
			if err == nil && (config.AppSlug != "" || config.AppId != "") {
				usingConfigFlow = true
				// Set app from config
				if config.AppSlug != "" {
					r.appSlug = config.AppSlug
				} else {
					r.appID = config.AppId
				}
			}
		}
		
		if usingConfigFlow {
			// The parent's PersistentPreRunE already ran and may have set wrong app from cache
			// We need to override it with the app from config and re-resolve
			
			// Clear the wrong app state that parent set
			r.appID = ""
			r.appType = ""
			// r.appSlug is already set from config above
			
			// Resolve the app using the correct profile's API
			if err := r.resolveAppTypeForDownload(); err != nil {
				return errors.Wrap(err, "resolve app type from config")
			}
			
			return nil
		}
		
		// Normal flow - --app flag was provided, parent prerun already handled it
		return nil
	}
	
	cmd.RunE = r.releaseDownload
	cmd.Flags().StringVarP(&r.args.releaseDownloadDest, "dest", "d", "", "File or directory to which release should be downloaded. Auto-generated if not specified.")
	cmd.Flags().StringVarP(&r.args.releaseDownloadChannel, "channel", "c", "", "Download the current release from this channel (case sensitive)")
}

func (r *runners) releaseDownload(command *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if r.appType != "kots" {
		return r.releaseInspect(command, args)
	}

	log := logger.NewLogger(os.Stdout)

	// Determine sequence to download
	var seq int64
	var err error
	
	if r.args.releaseDownloadChannel != "" {
		// Download from channel
		if len(args) > 0 {
			return errors.New("cannot specify both sequence and --channel flag")
		}
		
		log.ActionWithSpinner("Finding channel %q", r.args.releaseDownloadChannel)
		channel, err := r.api.GetChannelByName(r.appID, r.appType, r.args.releaseDownloadChannel)
		if err != nil {
			log.FinishSpinnerWithError()
			return errors.Wrapf(err, "get channel %q", r.args.releaseDownloadChannel)
		}
		
		if channel.ReleaseSequence == 0 {
			log.FinishSpinnerWithError()
			return errors.Errorf("channel %q has no releases", r.args.releaseDownloadChannel)
		}
		
		seq = channel.ReleaseSequence
		log.FinishSpinner()
		log.ActionWithoutSpinner("Channel %q is at sequence %d", r.args.releaseDownloadChannel, seq)
	} else if len(args) > 0 {
		// Use provided sequence
		seq, err = strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse sequence argument %q", args[0])
		}
	} else {
		// Download latest release
		log.ActionWithSpinner("Finding latest release")
		channels, err := r.api.ListChannels(r.appID, r.appType, "")
		if err != nil {
			log.FinishSpinnerWithError()
			return errors.Wrap(err, "list channels to find latest release")
		}
		
		var latestSeq int64
		for _, channel := range channels {
			if channel.ReleaseSequence > latestSeq {
				latestSeq = channel.ReleaseSequence
			}
		}
		
		if latestSeq == 0 {
			log.FinishSpinnerWithError()
			return errors.New("no releases found")
		}
		
		seq = latestSeq
		log.FinishSpinner()
		log.ActionWithoutSpinner("Latest release is sequence %d", seq)
	}

	// Determine destination and whether to save as file or directory
	dest := r.args.releaseDownloadDest
	saveAsFile := false
	
	if dest == "" {
		// Auto-generate filename for .tgz
		dest = r.generateDownloadFilename()
		saveAsFile = true
	} else {
		// Check if dest is an existing directory or should be treated as a file
		destInfo, statErr := os.Stat(dest)
		if statErr == nil {
			// Path exists - check if it's a directory
			saveAsFile = !destInfo.IsDir()
		} else {
			// Path doesn't exist - determine intent based on file extension
			// If it has .tgz or .tar.gz extension, treat as file; otherwise treat as directory
			if filepath.Ext(dest) == ".tgz" || filepath.Ext(dest) == ".gz" {
				saveAsFile = true
			} else {
				saveAsFile = false
			}
		}
	}

	if saveAsFile {
		// Download as .tgz file
		log.ActionWithSpinner("Downloading Release %d as %s", seq, dest)
		if err := r.downloadReleaseArchive(seq, dest); err != nil {
			log.FinishSpinnerWithError()
			return errors.Wrap(err, "download release archive")
		}
		log.FinishSpinner()
		log.ActionWithoutSpinner("Release %d downloaded to %s", seq, dest)
	} else {
		// Unpack to directory
		log.ActionWithSpinner("Fetching Release %d", seq)
		release, err := r.api.GetRelease(r.appID, r.appType, seq)
		if err != nil {
			log.FinishSpinnerWithError()
			return errors.Wrap(err, "get release")
		}
		log.FinishSpinner()

		log.ActionWithoutSpinner("Writing files to %s", dest)
		err = kotsrelease.Save(dest, release, log)
		if err != nil {
			return errors.Wrap(err, "save release")
		}
	}

	return nil
}

// resolveAppTypeForDownload resolves the app type for download command
func (r *runners) resolveAppTypeForDownload() error {
	if r.appID == "" && r.appSlug == "" {
		return nil
	}

	appSlugOrID := r.appSlug
	if appSlugOrID == "" {
		appSlugOrID = r.appID
	}

	app, appType, err := r.api.GetAppType(context.Background(), appSlugOrID, true)
	if err != nil {
		return errors.Wrapf(err, "get app type for %q", appSlugOrID)
	}

	r.appType = appType
	r.appID = app.ID
	r.appSlug = app.Slug

	return nil
}

// generateDownloadFilename generates a filename like app-slug.tgz or app-slug-2.tgz if it exists
func (r *runners) generateDownloadFilename() string {
	base := r.appSlug
	if base == "" {
		base = r.appID
	}
	
	filename := fmt.Sprintf("%s.tgz", base)
	if _, err := os.Stat(filename); err == nil {
		// File exists, try with incrementing number
		for i := 2; i < 1000; i++ {
			filename = fmt.Sprintf("%s-%d.tgz", base, i)
			if _, err := os.Stat(filename); err != nil {
				break
			}
		}
	}
	
	return filename
}

// downloadReleaseArchive downloads the release archive (.tgz) from the API
func (r *runners) downloadReleaseArchive(seq int64, dest string) error {
	// Get release to find the download URL
	release, err := r.api.GetRelease(r.appID, r.appType, seq)
	if err != nil {
		return errors.Wrap(err, "get release")
	}

	// The release config is base64 encoded JSON, we need to get the raw archive
	// For now, we'll use the kotsrelease.Save to a temp dir then tar it up
	// TODO: Look for a direct archive download endpoint
	
	tempDir, err := os.MkdirTemp("", "replicated-download-*")
	if err != nil {
		return errors.Wrap(err, "create temp directory")
	}
	defer os.RemoveAll(tempDir)

	log := logger.NewLogger(os.Stdout)
	if err := kotsrelease.Save(tempDir, release, log); err != nil {
		return errors.Wrap(err, "save release to temp dir")
	}

	// Create tar.gz from the temp directory
	return tarDirectory(tempDir, dest)
}

// tarDirectory creates a .tgz archive from a directory
func tarDirectory(srcDir, destFile string) error {
	// Create the destination file
	outFile, err := os.Create(destFile)
	if err != nil {
		return errors.Wrapf(err, "create output file %s", destFile)
	}
	defer outFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Walk the source directory
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return errors.Wrapf(err, "get relative path for %s", path)
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return errors.Wrapf(err, "create tar header for %s", path)
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return errors.Wrapf(err, "write tar header for %s", path)
		}

		// If it's a file, write its contents
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return errors.Wrapf(err, "open file %s", path)
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return errors.Wrapf(err, "write file %s to tar", path)
			}
		}

		return nil
	})
}
