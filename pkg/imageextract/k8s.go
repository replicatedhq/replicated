package imageextract

import (
	"bytes"

	kotsv1beta1 "github.com/replicatedhq/kotskinds/apis/kots/v1beta1"
	kotsscheme "github.com/replicatedhq/kotskinds/client/kotsclientset/scheme"
	troubleshootv1beta2 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	troubleshootscheme "github.com/replicatedhq/troubleshoot/pkg/client/troubleshootclientset/scheme"
	"github.com/replicatedhq/troubleshoot/pkg/docrewrite"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes/scheme"
)

// init registers KOTS and Troubleshoot types with the Kubernetes scheme
// so that the Universal Deserializer can decode these custom resources.
func init() {
	kotsscheme.AddToScheme(scheme.Scheme)
	troubleshootscheme.AddToScheme(scheme.Scheme)
}

// extractImagesFromFile extracts all image references from a YAML file.
// Ported from airgap-builder/pkg/builder/images.go lines 212-239
func extractImagesFromFile(fileData []byte) ([]string, []string) {
	allImages := []string{}
	excludedImages := []string{}

	// Split multi-document YAML - CRITICAL: use \n---\n as airgap does
	yamlDocs := bytes.Split(fileData, []byte("\n---\n"))

	for _, yamlDoc := range yamlDocs {
		parsed := &k8sDoc{}
		if err := yaml.Unmarshal(yamlDoc, parsed); err != nil {
			continue // Skip unparseable docs gracefully
		}

		// Handle Pod separately (different structure)
		if parsed.Kind != "Pod" {
			allImages = append(allImages, listImagesInDoc(parsed)...)
		} else {
			parsedPod := &k8sPodDoc{}
			if err := yaml.Unmarshal(yamlDoc, parsedPod); err != nil {
				continue
			}
			allImages = append(allImages, listImagesInPod(parsedPod)...)
		}

		// Extract from KOTS kinds (Application, Preflight, SupportBundle, Collector)
		kotsImages, excluded := listKotsKindsImages(yamlDoc)
		allImages = append(allImages, kotsImages...)
		if len(excluded) > 0 {
			excludedImages = append(excludedImages, excluded...)
		}
	}

	return allImages, excludedImages
}

// listImagesInDoc extracts images from Deployment, StatefulSet, DaemonSet, ReplicaSet, Job, CronJob.
// Ported from airgap-builder/pkg/builder/images.go lines 352-370
func listImagesInDoc(doc *k8sDoc) []string {
	images := make([]string, 0)

	// Deployment, StatefulSet, DaemonSet, ReplicaSet, Job
	for _, container := range doc.Spec.Template.Spec.Containers {
		if container.Image != "" {
			images = append(images, container.Image)
		}
	}
	for _, container := range doc.Spec.Template.Spec.InitContainers {
		if container.Image != "" {
			images = append(images, container.Image)
		}
	}

	// CronJob (has extra JobTemplate layer)
	for _, container := range doc.Spec.JobTemplate.Spec.Template.Spec.Containers {
		if container.Image != "" {
			images = append(images, container.Image)
		}
	}
	for _, container := range doc.Spec.JobTemplate.Spec.Template.Spec.InitContainers {
		if container.Image != "" {
			images = append(images, container.Image)
		}
	}

	return images
}

// listImagesInPod extracts images from Pod resources.
// Ported from airgap-builder/pkg/builder/images.go lines 372-383
func listImagesInPod(doc *k8sPodDoc) []string {
	images := make([]string, 0)

	for _, container := range doc.Spec.Containers {
		if container.Image != "" {
			images = append(images, container.Image)
		}
	}
	for _, container := range doc.Spec.InitContainers {
		if container.Image != "" {
			images = append(images, container.Image)
		}
	}
	for _, container := range doc.Spec.EphemeralContainers {
		if container.Image != "" {
			images = append(images, container.Image)
		}
	}

	return images
}

// listKotsKindsImages extracts images from KOTS Application and Troubleshoot resources.
// Ported from airgap-builder/pkg/builder/images.go lines 385-433
func listKotsKindsImages(yamlDoc []byte) ([]string, []string) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, gvk, err := decode(yamlDoc, nil, nil)
	if err != nil {
		return make([]string, 0), make([]string, 0)
	}

	// KOTS Application - AdditionalImages and ExcludedImages
	if gvk.Group == "kots.io" && gvk.Version == "v1beta1" && gvk.Kind == "Application" {
		app := obj.(*kotsv1beta1.Application)
		return app.Spec.AdditionalImages, app.Spec.ExcludedImages
	}

	// Troubleshoot specs - convert to v1beta2
	newDoc, err := docrewrite.ConvertToV1Beta2(yamlDoc)
	if err != nil {
		return make([]string, 0), make([]string, 0)
	}

	obj, gvk, err = decode(newDoc, nil, nil)
	if err != nil {
		return make([]string, 0), make([]string, 0)
	}

	if gvk.Group != "troubleshoot.sh" || gvk.Version != "v1beta2" {
		return make([]string, 0), make([]string, 0)
	}

	var collectors []*troubleshootv1beta2.Collect
	switch gvk.Kind {
	case "Collector":
		o := obj.(*troubleshootv1beta2.Collector)
		collectors = o.Spec.Collectors
	case "SupportBundle":
		o := obj.(*troubleshootv1beta2.SupportBundle)
		collectors = o.Spec.Collectors
	case "Preflight":
		o := obj.(*troubleshootv1beta2.Preflight)
		collectors = o.Spec.Collectors
	}

	images := make([]string, 0)
	for _, collect := range collectors {
		if collect.Run != nil && collect.Run.Image != "" {
			images = append(images, collect.Run.Image)
		}
	}

	return images, make([]string, 0)
}
