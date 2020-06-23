package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/manifoldco/promptui"
)

type kotsSingleSpec struct {
	Name     string   `json:"name"`
	Path     string   `json:"path"`
	Content  string   `json:"content"`
	Children []string `json:"children"`
}

func (r *runners) InitReleaseCreate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new release",
		Long: `Create a new release by providing YAML configuration for the next release in
  your sequence.`,
		SilenceUsage: true,
	}

	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.createReleaseYaml, "yaml", "", "The YAML config for this release. Use '-' to read from stdin.  Cannot be used with the `yaml-file` flag.")
	cmd.Flags().StringVar(&r.args.createReleaseYamlFile, "yaml-file", "", "The file name with YAML config for this release.  Cannot be used with the `yaml` flag.")
	cmd.Flags().StringVar(&r.args.createReleaseYamlDir, "yaml-dir", "", "The directory containing multiple yamls for a Kots release.  Cannot be used with the `yaml` flag.")
	cmd.Flags().StringVar(&r.args.createReleasePromote, "promote", "", "Channel name or id to promote this release to")
	cmd.Flags().StringVar(&r.args.createReleasePromoteNotes, "release-notes", "", "When used with --promote <channel>, sets the **markdown** release notes")
	cmd.Flags().StringVar(&r.args.createReleasePromoteVersion, "version", "", "When used with --promote <channel>, sets the version label for the release in this channel")
	cmd.Flags().BoolVar(&r.args.createReleasePromoteRequired, "required", false, "When used with --promote <channel>, marks this release as required during upgrades.")
	cmd.Flags().BoolVar(&r.args.createReleasePromoteEnsureChannel, "ensure-channel", false, "When used with --promote <channel>, will create the channel if it doesn't exist")

	cmd.RunE = r.releaseCreate
}

func (r *runners) releaseCreate(_ *cobra.Command, _ []string) error {
	if r.args.createReleaseYamlDir == "" {
		kotsManifestsDir, err := promptForAppYAMLDir("manifests")
		if err != nil {
			return errors.Wrap(err, "prompt for app name")
		}

		r.args.createReleaseYamlDir = kotsManifestsDir
	}

	if r.args.createReleaseYaml == "" &&
		r.args.createReleaseYamlFile == "" &&
		r.args.createReleaseYamlDir == "" {
		return errors.New("one of --yaml, --yaml-file, --yaml-dir is required")
	}

	// can't ensure a channel if you didn't pass one
	if r.args.createReleasePromoteEnsureChannel && r.args.createReleasePromote == "" {
		return errors.New("cannot use the flag --ensure-channel without also using --promote <channel> ")
	}

	// we check this again below, but lets be explicit and fail fast
	if r.args.createReleasePromoteEnsureChannel && !(r.appType == "ship" || r.appType == "kots") {
		return errors.Errorf("the flag --ensure-channel is only supported for KOTS and Ship applications, app %q is of type %q", r.appID, r.appType)
	}

	if r.args.createReleaseYaml != "" && r.args.createReleaseYamlFile != "" {
		return errors.New("only one of --yaml or --yaml-file may be specified")
	}

	if r.args.createReleaseYaml == "-" {
		bytes, err := ioutil.ReadAll(r.stdin)
		if err != nil {
			return errors.Wrap(err, "read from stdin")
		}
		r.args.createReleaseYaml = string(bytes)
	}

	if r.args.createReleaseYamlFile != "" {
		bytes, err := ioutil.ReadFile(r.args.createReleaseYamlFile)
		if err != nil {
			return errors.Wrap(err, "read file yaml")
		}
		r.args.createReleaseYaml = string(bytes)
	}

	if r.args.createReleaseYamlDir != "" {
		var err error
		r.args.createReleaseYaml, err = readYAMLDir(r.args.createReleaseYamlDir)
		if err != nil {
			return errors.Wrap(err, "read yaml dir")
		}
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

	release, err := r.api.CreateRelease(r.appID, r.appType, r.args.createReleaseYaml)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(r.w, "SEQUENCE: %d\n", release.Sequence); err != nil {
		return err
	}
	r.w.Flush()

	if promoteChanID != "" {
		if err := r.api.PromoteRelease(
			r.appID,
			r.appType,
			release.Sequence,
			r.args.createReleasePromoteVersion,
			r.args.createReleasePromoteNotes,
			r.args.createReleasePromoteRequired,
			promoteChanID,
		); err != nil {
			return err
		}

		// ignore error since operation was successful
		fmt.Fprintf(r.w, "Channel %s successfully set to release %d\n", promoteChanID, release.Sequence)
		r.w.Flush()
	}

	return nil
}

func (r *runners) getOrCreateChannelForPromotion(channelName string, createIfAbsent bool) (string, error) {

	description := "" // todo: do we want a flag for the desired channel description

	channel, err := r.api.GetChannelByName(
		r.appID,
		r.appType,
		channelName,
		description,
		createIfAbsent,
	)
	if err != nil {
		return "", errors.Wrapf(err, "get-or-create channel %q", channelName)
	}

	return channel.ID, nil
}

func encodeKotsFile(prefix, path string, info os.FileInfo, err error) (*kotsSingleSpec, error) {
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
	switch ext {
	case ".tgz", ".gz", ".yaml", ".yml":
		// continue
	default:
		return nil, nil
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "read file %s", path)
	}

	var str string
	switch ext {
	case ".tgz", ".gz":
		str = base64.StdEncoding.EncodeToString(bytes)
	default:
		str = string(bytes)
	}

	return &kotsSingleSpec{
		Name:     info.Name(),
		Path:     singlefile,
		Content:  str,
		Children: []string{},
	}, nil
}

func readYAMLDir(yamlDir string) (string, error) {

	var allKotsReleaseSpecs []kotsSingleSpec
	err := filepath.Walk(yamlDir, func(path string, info os.FileInfo, err error) error {
		spec, err := encodeKotsFile(yamlDir, path, info, err)
		if err != nil {
			return err
		} else if spec == nil {
			return nil
		}
		allKotsReleaseSpecs = append(allKotsReleaseSpecs, *spec)
		return nil
	})
	if err != nil {
		return "", errors.Wrapf(err, "walk %s", yamlDir)
	}

	jsonAllYamls, err := json.Marshal(allKotsReleaseSpecs)
	if err != nil {
		return "", errors.Wrap(err, "marshal spec")
	}
	return string(jsonAllYamls), nil
}

func promptForAppYAMLDir(chartName string) (string, error) {

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . | bold }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     "Enter the app YAML dir to use:",
		Templates: templates,
		Default:   chartName,
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("invalid app name")
			}

			return nil
		},
	}

	for {
		result, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				os.Exit(-1)
			}
			continue
		}

		return result, nil
	}
}
