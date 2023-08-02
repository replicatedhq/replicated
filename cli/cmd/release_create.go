package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/client"
	kotstypes "github.com/replicatedhq/replicated/pkg/kots/release/types"
	"github.com/replicatedhq/replicated/pkg/logger"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

const (
	defaultYAMLDir = "manifests"
)

func (r *runners) InitReleaseCreate(parent *cobra.Command) error {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new release",
		Long: `Create a new release by providing YAML configuration for the next release in
  your sequence.`,
		SilenceUsage:  true,
		SilenceErrors: true, // this command uses custom error printing
	}

	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.createReleaseYaml, "yaml", "", "The YAML config for this release. Use '-' to read from stdin. Cannot be used with the --yaml-file flag.")
	cmd.Flags().StringVar(&r.args.createReleaseYamlFile, "yaml-file", "", "The file name with YAML config for this release. Cannot be used with the --yaml flag.")
	cmd.Flags().StringVar(&r.args.createReleaseYamlDir, "yaml-dir", "", "The directory containing multiple yamls for a Kots release. Cannot be used with the --yaml flag.")
	cmd.Flags().StringVar(&r.args.createReleaseChart, "chart", "", "Helm chart to create the release from. Cannot be used with the --yaml, --yaml-file, or --yaml-dir flags.")
	cmd.Flags().StringVar(&r.args.createReleasePromote, "promote", "", "Channel name or id to promote this release to")
	cmd.Flags().StringVar(&r.args.createReleasePromoteNotes, "release-notes", "", "When used with --promote <channel>, sets the **markdown** release notes")
	cmd.Flags().StringVar(&r.args.createReleasePromoteVersion, "version", "", "When used with --promote <channel>, sets the version label for the release in this channel")
	// Fail-on linting flag (from release_lint.go)
	cmd.Flags().StringVar(&r.args.lintReleaseFailOn, "fail-on", "error", "The minimum severity to cause the command to exit with a non-zero exit code. Supported values are [info, warn, error, none].")
	// Replicated release create lint flag
	cmd.Flags().BoolVar(&r.args.createReleaseLint, "lint", false, "Lint a manifests directory prior to creation of the KOTS Release.")
	cmd.Flags().BoolVar(&r.args.createReleasePromoteRequired, "required", false, "When used with --promote <channel>, marks this release as required during upgrades.")
	cmd.Flags().BoolVar(&r.args.createReleasePromoteEnsureChannel, "ensure-channel", false, "When used with --promote <channel>, will create the channel if it doesn't exist")
	cmd.Flags().BoolVar(&r.args.createReleaseAutoDefaults, "auto", false, "generate default values for use in CI")
	cmd.Flags().BoolVarP(&r.args.createReleaseAutoDefaultsAccept, "confirm-auto", "y", false, "auto-accept the configuration generated by the --auto flag")

	// not supported for KOTS
	cmd.Flags().MarkHidden("required")
	cmd.Flags().MarkHidden("yaml-file")
	cmd.Flags().MarkHidden("yaml")

	cmd.RunE = r.releaseCreate
	return nil
}

func (r *runners) gitSHABranch() (sha string, branch string, dirty bool, err error) {
	path := "."
	rev := "HEAD"
	repository, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return "", "", false, errors.Wrapf(err, "open %q", path)
	}
	h, err := repository.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return "", "", false, errors.Wrapf(err, "resolve revision")
	}
	head, err := repository.Head()
	if err != nil {
		return "", "", false, errors.Wrapf(err, "resolve HEAD")
	}

	worktree, err := repository.Worktree()
	if err != nil {
		return "", "", false, errors.Wrap(err, "get git worktree")
	}
	status, err := worktree.Status()
	if err != nil {
		return "", "", false, errors.Wrap(err, "get git status")
	}

	branchName := head.Name().Short()

	// for GH Actions, prefer env branch
	envBranch := os.Getenv("GITHUB_BRANCH_NAME")
	if envBranch != "" {
		branchName = envBranch
	}

	return h.String()[0:7], branchName, !status.IsClean(), nil
}

func (r *runners) setKOTSDefaultReleaseParams() error {
	if !r.isFoundationApp && r.args.createReleaseYamlDir == "" {
		r.args.createReleaseYamlDir = "./manifests"
	}

	rev, branch, isDirty, err := r.gitSHABranch()
	if err != nil {
		return errors.Wrapf(err, "get git properties")
	}
	dirtyStatus := ""
	if isDirty {
		dirtyStatus = "-dirty"
	}

	if r.args.createReleasePromoteNotes == "" {
		// set some default release notes
		r.args.createReleasePromoteNotes = fmt.Sprintf(
			`CLI release of %s triggered by %s [SHA: %s%s] [%s]`,
			branch,
			os.Getenv("USER"),
			rev,
			dirtyStatus,
			time.Now().Format(time.RFC822),
		)
		// unless it's GH actions, then we can link to the commit! yay!
		if os.Getenv("GITHUB_ACTIONS") != "" {
			r.args.createReleasePromoteNotes = fmt.Sprintf(
				`GitHub Action release of %s triggered by %s: [%s](https://github.com/%s/commit/%s)`,
				os.Getenv("GITHUB_REF"),
				os.Getenv("GITHUB_ACTOR"),
				os.Getenv("GITHUB_SHA")[0:7],
				os.Getenv("GITHUB_REPOSITORY"),
				os.Getenv("GITHUB_SHA"),
			)
		}
	}

	if r.args.createReleasePromote == "" {
		r.args.createReleasePromote = branch
		if branch == "master" || branch == "main" {
			r.args.createReleasePromote = "Unstable"
		}
	}

	if !r.isFoundationApp && r.args.createReleasePromoteVersion == "" {
		r.args.createReleasePromoteVersion = fmt.Sprintf("%s-%s%s", r.args.createReleasePromote, rev, dirtyStatus)
	}

	r.args.createReleasePromoteEnsureChannel = true
	if !r.isFoundationApp {
		r.args.createReleaseLint = true
	}

	return nil
}

func (r *runners) releaseCreate(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		printIfError(err)
	}()

	log := logger.NewLogger(r.w)

	if r.appType == "kots" && r.args.createReleaseAutoDefaults {
		log.ActionWithSpinner("Reading Environment")
		err = r.setKOTSDefaultReleaseParams()
		if err != nil {
			log.FinishSpinnerWithError()
			return errors.Wrap(err, "resolve kots defaults")
		}
		time.Sleep(500 * time.Millisecond)
		log.FinishSpinner()

		fmt.Fprintf(r.w, `
Prepared to create release with defaults:

    yaml-dir        %q
    promote         %q
    version         %q
    release-notes   %q
    ensure-channel  %t
    lint-release    %t

`, r.args.createReleaseYamlDir, r.args.createReleasePromote, r.args.createReleasePromoteVersion, r.args.createReleasePromoteNotes, r.args.createReleasePromoteEnsureChannel, r.args.createReleaseLint)
		if !r.args.createReleaseAutoDefaultsAccept {
			var confirmed string
			confirmed, err = promptForConfirm()
			if err != nil {
				return errors.Wrap(err, "prompt for confirm")
			}
			if strings.ToLower(confirmed) != "y" {
				return errors.New("configuration declined")
			}
			fmt.Printf("You can use the --confirm-auto or -y flag in the future to skip this prompt.\n")
		}
	}

	err = r.validateReleaseCreateParams()
	if err != nil {
		return errors.Wrap(err, "validate params")
	}

	// Check if --lint argument has been passed in by the enduser
	if r.args.createReleaseLint {
		// Request lint release yaml directory to check
		r.args.lintReleaseYamlDir = r.args.createReleaseYamlDir
		r.args.lintReleaseChart = r.args.createReleaseChart
		// Call release_lint.go releaseLint function
		err = r.releaseLint(cmd, args)
		if err != nil {
			return errors.Wrap(err, "lint yaml")
		}
	}

	if r.args.createReleaseYaml == "-" {
		var bytes []byte
		bytes, err = ioutil.ReadAll(r.stdin)
		if err != nil {
			return errors.Wrap(err, "read from stdin")
		}
		r.args.createReleaseYaml = string(bytes)
	}

	if r.args.createReleaseYamlFile != "" {
		var bytes []byte
		bytes, err = ioutil.ReadFile(r.args.createReleaseYamlFile)
		if err != nil {
			return errors.Wrap(err, "read file yaml")
		}
		r.args.createReleaseYaml = string(bytes)
	}

	if r.args.createReleaseYamlDir != "" {
		fmt.Fprintln(r.w)
		log.ActionWithSpinner("Reading manifests from %s", r.args.createReleaseYamlDir)
		var err error
		r.args.createReleaseYaml, err = makeReleaseFromDir(r.args.createReleaseYamlDir)
		if err != nil {
			log.FinishSpinnerWithError()
			return errors.Wrap(err, "make release from dir")
		}
		log.FinishSpinner()
	}

	if r.args.createReleaseChart != "" {
		if !r.isFoundationApp {
			fmt.Fprint(r.w, "You are creating a release that will only be installable with the helm CLI.\n"+
				"For more information, see \n"+
				"https://docs.replicated.com/vendor/helm-install#about-helm-installations-with-replicated\n")
		}
		fmt.Fprintln(r.w)
		log.ActionWithSpinner("Reading chart from %s", r.args.createReleaseChart)
		r.args.createReleaseYaml, err = makeReleaseFromChart(r.args.createReleaseChart)
		if err != nil {
			log.FinishSpinnerWithError()
			return errors.Wrap(err, "make release from chart")
		}
		log.FinishSpinner()
	}

	// if the --promote param was used make sure it identifies exactly one
	// channel before proceeding
	var promoteChanID string
	if r.args.createReleasePromote != "" {
		var err error
		promoteChanID, err = r.getOrCreateChannelForPromotion(
			r.args.createReleasePromote,
			r.args.createReleasePromoteEnsureChannel,
		)
		if err != nil {
			return errors.Wrapf(err, "get or create channel %q for promotion", promoteChanID)
		}
	}

	log.ActionWithSpinner("Creating Release")
	var release *types.ReleaseInfo
	release, err = r.api.CreateRelease(r.appID, r.appType, r.args.createReleaseYaml)
	if err != nil {
		log.FinishSpinnerWithError()
		return errors.Wrap(err, "create release")
	}
	log.FinishSpinner()

	log.ChildActionWithoutSpinner("SEQUENCE: %d", release.Sequence)

	if promoteChanID != "" {
		log.ActionWithSpinner("Promoting")
		if err = r.api.PromoteRelease(
			r.appID,
			r.appType,
			release.Sequence,
			r.args.createReleasePromoteVersion,
			r.args.createReleasePromoteNotes,
			r.args.createReleasePromoteRequired,
			promoteChanID,
		); err != nil {
			log.FinishSpinnerWithError()
			return errors.Wrap(err, "promote release")
		}
		log.FinishSpinner()

		// ignore error since operation was successful
		log.ChildActionWithoutSpinner("Channel %s successfully set to release %d\n", promoteChanID, release.Sequence)
	}

	return nil
}

func (r *runners) validateReleaseCreateParams() error {
	specSources := []string{
		r.args.createReleaseYaml,
		r.args.createReleaseYamlFile,
		r.args.createReleaseYamlDir,
		r.args.createReleaseChart,
	}

	numSources := 0
	for _, specSource := range specSources {
		if specSource != "" {
			numSources++
		}
	}

	if numSources == 0 {
		return errors.New("one of --yaml, --yaml-file, --yaml-dir, or --chart is required")
	}

	if numSources > 1 {
		return errors.New("only one of --yaml, --yaml-file, --yaml-dir, or --chart may be specified")
	}

	if r.appType != "kots" {
		if r.args.createReleaseYaml == "" && r.args.createReleaseYamlFile == "" {
			return errors.New("one of --yaml or --yaml-file must be provided")
		}

		if (strings.HasSuffix(r.args.createReleaseYaml, ".yaml") || strings.HasSuffix(r.args.createReleaseYaml, ".yml")) &&
			len(strings.Split(r.args.createReleaseYaml, " ")) == 1 {
			return errors.New("use the --yaml-file flag when passing a yaml filename")
		}
	} else {
		if r.args.createReleaseYaml != "" {
			return errors.Errorf("the --yaml flag is not supported for KOTS applications, use --yaml-dir instead")
		}

		if r.args.createReleaseYamlFile != "" {
			return errors.Errorf("the --yaml-file flag is not supported for KOTS applications, use --yaml-dir instead")
		}
	}

	// can't ensure a channel if you didn't pass one
	if r.args.createReleasePromoteEnsureChannel && r.args.createReleasePromote == "" {
		return errors.New("cannot use the flag --ensure-channel without also using --promote <channel> ")
	}

	// we check this again below, but lets be explicit and fail fast
	if r.args.createReleasePromoteEnsureChannel && r.appType != "kots" {
		return errors.Errorf("the flag --ensure-channel is only supported for KOTS applications, app %q is of type %q", r.appID, r.appType)
	}

	return nil
}

func (r *runners) getOrCreateChannelForPromotion(channelName string, createIfAbsent bool) (string, error) {
	description := "" // todo: do we want a flag for the desired channel description

	opts := client.GetOrCreateChannelOptions{
		AppID:          r.appID,
		AppType:        r.appType,
		NameOrID:       channelName,
		Description:    description,
		CreateIfAbsent: createIfAbsent,
	}
	channel, err := r.api.GetOrCreateChannelByName(opts)
	if err != nil {
		return "", errors.Wrapf(err, "get-or-create channel %q", channelName)
	}

	return channel.ID, nil
}

func encodeKotsFile(prefix, path string, info os.FileInfo, err error) (*kotstypes.KotsSingleSpec, error) {
	if err != nil {
		return nil, err
	}

	singlefile := strings.TrimPrefix(filepath.Clean(path), filepath.Clean(prefix)+"/")

	if info.IsDir() {
		return nil, nil
	}
	if strings.HasPrefix(info.Name(), ".") {
		return nil, nil
	}
	ext := filepath.Ext(info.Name())
	if !isSupportedExt(ext) {
		return nil, nil
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "read file %s", path)
	}

	var str string
	switch ext {
	case ".tgz", ".gz", ".woff", ".woff2", ".ttf", ".otf", ".eot", ".svg":
		str = base64.StdEncoding.EncodeToString(bytes)
	default:
		str = string(bytes)
	}

	return &kotstypes.KotsSingleSpec{
		Name:     info.Name(),
		Path:     singlefile,
		Content:  str,
		Children: []kotstypes.KotsSingleSpec{},
	}, nil
}

func makeReleaseFromDir(fileDir string) (string, error) {
	fileInfo, err := os.Stat(fileDir)
	if err != nil {
		return "", errors.Wrapf(err, "stat %s", fileDir)
	}

	if !fileInfo.IsDir() {
		return "", errors.Errorf("path %s is not a directory", fileDir)
	}

	var allKotsReleaseSpecs []kotstypes.KotsSingleSpec
	err = filepath.Walk(fileDir, func(path string, info os.FileInfo, err error) error {
		spec, err := encodeKotsFile(fileDir, path, info, err)
		if err != nil {
			return err
		} else if spec == nil {
			return nil
		}
		allKotsReleaseSpecs = append(allKotsReleaseSpecs, *spec)
		return nil
	})
	if err != nil {
		return "", errors.Wrapf(err, "walk %s", fileDir)
	}

	jsonAllFiles, err := json.Marshal(allKotsReleaseSpecs)
	if err != nil {
		return "", errors.Wrap(err, "marshal spec")
	}
	return string(jsonAllFiles), nil
}

func makeReleaseFromChart(chartFile string) (string, error) {
	fileInfo, err := os.Stat(chartFile)
	if err != nil {
		return "", errors.Wrapf(err, "stat %s", chartFile)
	}

	if fileInfo.IsDir() {
		return "", errors.Errorf("chart path %s is a directory", chartFile)
	}

	dirName, _ := filepath.Split(chartFile)
	spec, err := encodeKotsFile(dirName, chartFile, fileInfo, nil)
	if err != nil {
		return "", errors.Wrap(err, "encode chart file")
	}

	if spec == nil {
		return "", errors.Errorf("chart file %s is not supported", chartFile)
	}

	allKotsReleaseSpecs := []kotstypes.KotsSingleSpec{
		*spec,
	}

	jsonAllFiles, err := json.Marshal(allKotsReleaseSpecs)
	if err != nil {
		return "", errors.Wrap(err, "marshal spec")
	}

	return string(jsonAllFiles), nil
}

func promptForConfirm() (string, error) {
	prompt := promptui.Prompt{
		Label:     "Create with these properties? (default Yes) [Y/n]",
		Templates: templates,
		Default:   "y",
		Validate: func(input string) error {
			switch strings.ToLower(input) {
			case "y", "n":
				return nil
			default:
				return errors.New(`please choose "y" or "n"`)
			}
		},
	}

	for {
		result, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return "", errors.New("interrupted")
			}
			continue
		}

		return result, nil
	}
}

func isSupportedExt(ext string) bool {
	switch ext {
	case ".tgz", ".gz", ".yaml", ".yml", ".css", ".woff", ".woff2", ".ttf", ".otf", ".eot", ".svg":
		return true
	default:
		return false
	}
}
