package print

import (
	"encoding/json"
	"fmt"
	"sort"
	"text/tabwriter"
)

func ChannelImages(format string, w *tabwriter.Writer, images []string) error {
	// Sort images for consistent output
	sort.Strings(images)

	switch format {
	case "json":
		out, err := json.MarshalIndent(images, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(out)); err != nil {
			return err
		}
		return w.Flush()
	case "table":
		// Print header
		fmt.Fprintln(w, "IMAGE")

		// Print each image
		for _, image := range images {
			fmt.Fprintln(w, image)
		}

		return w.Flush()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}
