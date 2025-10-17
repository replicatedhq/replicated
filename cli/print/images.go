package print

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/replicatedhq/replicated/pkg/imageextract"
)

// Images prints extracted image references in the specified format
func Images(format string, w *tabwriter.Writer, result *imageextract.Result) error {
	switch format {
	case "table":
		return printImagesTable(w, result)
	case "json":
		return printImagesJSON(w, result)
	case "list":
		return printImagesList(w, result)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func printImagesTable(w *tabwriter.Writer, result *imageextract.Result) error {
	if len(result.Images) == 0 {
		fmt.Fprintln(w, "No images found")
		w.Flush()
		return nil
	}

	// Print header
	fmt.Fprintln(w, "IMAGE\tTAG\tREGISTRY\tSOURCE")

	// Print each image
	for _, img := range result.Images {
		source := ""
		if len(img.Sources) > 0 {
			s := img.Sources[0]
			if s.Kind != "" && s.Name != "" {
				source = fmt.Sprintf("%s/%s", s.Kind, s.Name)
			} else if s.File != "" {
				source = s.File
			}
		}

		repository := img.Repository
		if repository == "" {
			repository = img.Raw
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", repository, img.Tag, img.Registry, source)
	}

	w.Flush()

	// Print warnings
	if len(result.Warnings) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Warnings:")
		for _, warning := range result.Warnings {
			fmt.Fprintf(w, "⚠  %s - %s\n", warning.Image, warning.Message)
		}
		w.Flush()
	}

	// Print summary
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Found %d unique images\n", len(result.Images))
	w.Flush()

	return nil
}

func printImagesJSON(w *tabwriter.Writer, result *imageextract.Result) error {
	type JSONOutput struct {
		Images   []imageextract.ImageRef `json:"images"`
		Warnings []imageextract.Warning  `json:"warnings"`
		Summary  map[string]int          `json:"summary"`
	}

	output := JSONOutput{
		Images:   result.Images,
		Warnings: result.Warnings,
		Summary: map[string]int{
			"total":  len(result.Images),
			"unique": len(result.Images),
		},
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return err
	}

	w.Flush()
	return nil
}

func printImagesList(w *tabwriter.Writer, result *imageextract.Result) error {
	for _, img := range result.Images {
		fmt.Fprintln(w, img.Raw)
	}
	w.Flush()
	return nil
}
