// Package imageextract extracts container image references from Kubernetes manifests and Helm charts.
// This implementation is ported from github.com/replicatedhq/airgap/airgap-builder/pkg/builder/images.go
package imageextract

import "context"

// Extractor defines the interface for extracting container image references.
type Extractor interface {
	ExtractFromDirectory(ctx context.Context, dir string, opts Options) (*Result, error)
	ExtractFromChart(ctx context.Context, chartPath string, opts Options) (*Result, error)
	ExtractFromManifests(ctx context.Context, manifests []byte, opts Options) (*Result, error)
}

// Options configures the extraction behavior.
type Options struct {
	HelmValues        map[string]interface{}
	HelmValuesFiles   []string
	Namespace         string
	IncludeDuplicates bool
	NoWarnings        bool
}

// Result contains the extracted image references, warnings, and errors.
type Result struct {
	Images   []ImageRef
	Warnings []Warning
	Errors   []error
}

// ImageRef represents a parsed container image reference.
type ImageRef struct {
	Raw        string   // Original reference string
	Registry   string   // Parsed registry
	Repository string   // Parsed repository
	Tag        string   // Parsed tag
	Digest     string   // Parsed digest (if present)
	Sources    []Source // Where this image was found
}

// Source identifies where an image reference was found.
type Source struct {
	File          string
	Kind          string
	Name          string
	Namespace     string
	Container     string
	ContainerType string // container, initContainer, ephemeralContainer
}

// Warning represents an issue detected with an image reference.
type Warning struct {
	Image   string
	Type    WarningType
	Message string
	Source  *Source
}

// WarningType categorizes different types of warnings.
type WarningType string

const (
	WarningLatestTag     WarningType = "latest-tag"
	WarningNoTag         WarningType = "no-tag"
	WarningInsecure      WarningType = "insecure-registry"
	WarningUnqualified   WarningType = "unqualified-name"
	WarningInvalidSyntax WarningType = "invalid-syntax"
)

// k8s struct definitions ported from airgap (lines 42-77)
// These structs map directly to Kubernetes YAML structure for efficient parsing.

type k8sDoc struct {
	ApiVersion string  `yaml:"apiVersion"`
	Kind       string  `yaml:"kind"`
	Spec       k8sSpec `yaml:"spec"`
}

type k8sPodDoc struct {
	Kind string     `yaml:"kind"`
	Spec k8sPodSpec `yaml:"spec"`
}

type k8sSpec struct {
	Template    k8sTemplate    `yaml:"template"`
	JobTemplate k8sJobTemplate `yaml:"jobTemplate"`
}

type k8sJobTemplate struct {
	Spec k8sJobSpec `yaml:"spec"`
}

type k8sJobSpec struct {
	Template k8sTemplate `yaml:"template"`
}

type k8sTemplate struct {
	Spec k8sPodSpec `yaml:"spec"`
}

type k8sPodSpec struct {
	Containers     []k8sContainer `yaml:"containers"`
	InitContainers []k8sContainer `yaml:"initContainers"`
}

type k8sContainer struct {
	Image string `yaml:"image"`
}
