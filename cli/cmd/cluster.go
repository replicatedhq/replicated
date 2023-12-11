package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

var (
	ErrCompatibilityMatrixTermsNotAccepted = errors.New("You must read and accept the Compatibility Matrix Terms of Service before using this command. To view, please visit https://vendor.replicated.com/compatibility-matrix")
)

func (r *runners) InitClusterCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage test clusters",
		Long:  ``,
	}
	parent.AddCommand(cmd)

	return cmd
}

func parseTags(tags []string) ([]kotsclient.ClusterTag, error) {
	clusterTags := []kotsclient.ClusterTag{}
	for _, tag := range tags {
		tagParts := strings.SplitN(tag, "=", 2)
		if len(tagParts) != 2 {
			return nil, errors.Errorf("invalid tag format: %s", tag)
		}

		clusterTags = append(clusterTags, kotsclient.ClusterTag{
			Key:   tagParts[0],
			Value: tagParts[1],
		})
	}
	return clusterTags, nil
}
