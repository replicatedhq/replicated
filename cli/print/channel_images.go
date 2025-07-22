package print

import (
	"fmt"
	"sort"
	"text/tabwriter"
)

func ChannelImages(w *tabwriter.Writer, images []string) error {
	// Sort images for consistent output
	sort.Strings(images)
	
	// Print header
	fmt.Fprintln(w, "IMAGE")
	
	// Print each image
	for _, image := range images {
		fmt.Fprintln(w, image)
	}
	
	return w.Flush()
}