package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitReleaseUpdate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "update SEQUENCE",
		Short: "Updated a release's yaml config",
		Long:  "Updated a release's yaml config",
	}

	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.updateReleaseYaml, "yaml", "", "The new YAML config for this release. Use '-' to read from stdin.  Cannot be used with the `yaml-file` flag.")
	cmd.Flags().StringVar(&r.args.updateReleaseYamlFile, "yaml-file", "", "The file name with YAML config for this release.  Cannot be used with the `yaml` flag.")
	cmd.Flags().StringVar(&r.args.updateReleaseYamlDir, "yaml-dir", "", "The directory containing multiple yamls for a Kots release.  Cannot be used with the `yaml` flag.")

	cmd.RunE = r.releaseUpdate
}

func (r *runners) releaseUpdate(cmd *cobra.Command, args []string) error {
	if r.args.updateReleaseYaml == "" && r.args.updateReleaseYamlFile == "" && r.args.updateReleaseYamlDir == "" {
		return fmt.Errorf("yaml is required")
	}

	if r.args.updateReleaseYaml != "" && r.args.updateReleaseYamlFile != "" {
		return fmt.Errorf("only yaml or yaml-file has to be specified")
	}

	if r.args.updateReleaseYaml == "-" {
		bytes, err := ioutil.ReadAll(r.stdin)
		if err != nil {
			return err
		}
		r.args.updateReleaseYaml = string(bytes)
	}

	if r.args.updateReleaseYamlFile != "" {
		bytes, err := ioutil.ReadFile(r.args.updateReleaseYamlFile)
		if err != nil {
			return err
		}
		r.args.updateReleaseYaml = string(bytes)
	}

	if len(args) < 1 {
		return errors.New("release sequence is required")
	}
	seq, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid release sequence: %s", args[0])
	}

	if r.args.updateReleaseYamlDir != "" {
		var allKotsReleaseSpecs []kotsSingleSpec
		err := filepath.Walk(r.args.updateReleaseYamlDir, func(path string, info os.FileInfo, err error) error {

			singlefile := strings.TrimPrefix(path, r.args.updateReleaseYamlDir)

			if err != nil {
				return errors.Wrapf(err, "walk %s", info.Name())
			}

			if info.IsDir() {
				return nil
			}
			if strings.HasPrefix(info.Name(), ".") {
				return nil
			}
			ext := filepath.Ext(info.Name())
			if ext != ".yaml" && ext != ".yml" {
				return nil
			}

			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				return errors.Wrapf(err, "read file %s", path)
			}

			spec := kotsSingleSpec{
				Name:     info.Name(),
				Path:     singlefile,
				Content:  string(bytes),
				Children: []string{},
			}
			allKotsReleaseSpecs = append(allKotsReleaseSpecs, spec)
			return nil
		})
		if err != nil {
			return errors.Wrapf(err, "walk %s", r.args.updateReleaseYamlDir)
		}

		jsonAllYamls, err := json.Marshal(allKotsReleaseSpecs)
		if err != nil {
			return errors.Wrap(err, "marshal spec")
		}
		r.args.updateReleaseYaml = string(jsonAllYamls)
	}
	if err := r.api.UpdateRelease(r.appID, r.appType, seq, r.args.updateReleaseYaml); err != nil {
		return fmt.Errorf("Failure setting new yaml config for release: %v", err)
	}

	// ignore the error since operation was successful
	fmt.Fprintf(r.w, "Release %d updated\n", seq)
	r.w.Flush()

	return nil
}
