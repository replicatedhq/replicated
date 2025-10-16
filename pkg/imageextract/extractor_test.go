package imageextract

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractImagesFromFile_Deployment(t *testing.T) {
	yaml := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.19
      - name: sidecar
        image: gcr.io/project/app:v1
      initContainers:
      - name: init
        image: busybox:latest`

	images, _ := extractImagesFromFile([]byte(yaml))

	if len(images) != 3 {
		t.Fatalf("expected 3 images, got %d", len(images))
	}

	expected := map[string]bool{
		"nginx:1.19":            true,
		"gcr.io/project/app:v1": true,
		"busybox:latest":        true,
	}

	for _, img := range images {
		if !expected[img] {
			t.Errorf("unexpected image: %s", img)
		}
	}
}

func TestExtractImagesFromFile_Pod(t *testing.T) {
	yaml := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: app
    image: myapp:1.0
  initContainers:
  - name: init
    image: alpine:3.14`

	images, _ := extractImagesFromFile([]byte(yaml))

	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(images))
	}
}

func TestExtractImagesFromFile_CronJob(t *testing.T) {
	yaml := `apiVersion: batch/v1
kind: CronJob
metadata:
  name: scheduled
spec:
  schedule: "0 0 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: task
            image: task:v1`

	images, _ := extractImagesFromFile([]byte(yaml))

	if len(images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(images))
	}

	if images[0] != "task:v1" {
		t.Errorf("expected task:v1, got %s", images[0])
	}
}

func TestExtractImagesFromFile_MultiDoc(t *testing.T) {
	yaml := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app1
spec:
  template:
    spec:
      containers:
      - name: web
        image: nginx:1.19
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app2
spec:
  template:
    spec:
      containers:
      - name: api
        image: api:v1.0`

	images, _ := extractImagesFromFile([]byte(yaml))

	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(images))
	}
}

func TestDeduplicateImages(t *testing.T) {
	images := []string{
		"nginx:1.19",
		"nginx:1.19",
		"redis:6",
		"postgres:14",
		"nginx:1.19",
	}

	result := deduplicateImages(images, []string{})

	if len(result) != 3 {
		t.Fatalf("expected 3 unique images, got %d", len(result))
	}
}

func TestDeduplicateImages_WithExclusions(t *testing.T) {
	images := []string{
		"nginx:1.19",
		"redis:6",
		"postgres:14",
	}

	excluded := []string{
		"redis:6",
	}

	result := deduplicateImages(images, excluded)

	if len(result) != 2 {
		t.Fatalf("expected 2 images after exclusion, got %d", len(result))
	}

	for _, img := range result {
		if img == "redis:6" {
			t.Error("redis:6 should have been excluded")
		}
	}
}

func TestParseImageRef(t *testing.T) {
	tests := []struct {
		input    string
		registry string
		repo     string
		tag      string
	}{
		{"nginx:1.19", "docker.io", "library/nginx", "1.19"},
		{"redis", "docker.io", "library/redis", "latest"},
		{"gcr.io/proj/app:v1", "gcr.io", "proj/app", "v1"},
		{"localhost:5000/app:dev", "localhost:5000", "app", "dev"},
		{"user/app:v2", "docker.io", "user/app", "v2"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			img := parseImageRef(tt.input)
			if img.Registry != tt.registry {
				t.Errorf("registry: got %s, want %s", img.Registry, tt.registry)
			}
			if img.Repository != tt.repo {
				t.Errorf("repo: got %s, want %s", img.Repository, tt.repo)
			}
			if img.Tag != tt.tag {
				t.Errorf("tag: got %s, want %s", img.Tag, tt.tag)
			}
		})
	}
}

func TestExtractFromDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "extract-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	yaml1 := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app1
spec:
  template:
    spec:
      containers:
      - name: web
        image: nginx:1.19`

	yaml2 := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app2
spec:
  template:
    spec:
      containers:
      - name: api
        image: api:v1.0`

	os.WriteFile(filepath.Join(tmpDir, "deploy1.yaml"), []byte(yaml1), 0644)
	os.WriteFile(filepath.Join(tmpDir, "deploy2.yml"), []byte(yaml2), 0644)

	e := NewExtractor()
	result, err := e.ExtractFromDirectory(context.Background(), tmpDir, Options{})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(result.Images))
	}
}

func TestExtractFromChart(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "chart-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create minimal chart
	os.WriteFile(filepath.Join(tmpDir, "Chart.yaml"), []byte(`apiVersion: v2
name: test
version: 1.0.0`), 0644)

	os.WriteFile(filepath.Join(tmpDir, "values.yaml"), []byte(`image:
  repository: nginx
  tag: "1.19"`), 0644)

	os.Mkdir(filepath.Join(tmpDir, "templates"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "templates", "deployment.yaml"), []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}
spec:
  template:
    spec:
      containers:
      - name: app
        image: {{ .Values.image.repository }}:{{ .Values.image.tag }}`), 0644)

	e := NewExtractor()
	result, err := e.ExtractFromChart(context.Background(), tmpDir, Options{})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(result.Images))
	}

	if result.Images[0].Repository != "library/nginx" || result.Images[0].Tag != "1.19" {
		t.Errorf("unexpected image: %+v", result.Images[0])
	}
}

func TestGenerateWarnings(t *testing.T) {
	tests := []struct {
		name     string
		image    ImageRef
		wantType WarningType
	}{
		{
			name:     "latest tag",
			image:    ImageRef{Raw: "nginx:latest", Tag: "latest", Sources: []Source{{}}},
			wantType: WarningLatestTag,
		},
		{
			name:     "insecure registry",
			image:    ImageRef{Raw: "http://reg.com/app:v1", Sources: []Source{{}}},
			wantType: WarningInsecure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := generateWarnings(tt.image)
			found := false
			for _, w := range warnings {
				if w.Type == tt.wantType {
					found = true
				}
			}
			if !found {
				t.Errorf("expected warning type %s", tt.wantType)
			}
		})
	}
}

func TestListImagesInDoc_StatefulSet(t *testing.T) {
	doc := &k8sDoc{
		Kind: "StatefulSet",
		Spec: k8sSpec{
			Template: k8sTemplate{
				Spec: k8sPodSpec{
					Containers: []k8sContainer{
						{Image: "redis:6.2"},
					},
					InitContainers: []k8sContainer{
						{Image: "busybox:latest"},
					},
				},
			},
		},
	}

	images := listImagesInDoc(doc)

	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(images))
	}
}

func TestListImagesInPod(t *testing.T) {
	doc := &k8sPodDoc{
		Kind: "Pod",
		Spec: k8sPodSpec{
			Containers: []k8sContainer{
				{Image: "nginx:1.19"},
			},
			InitContainers: []k8sContainer{
				{Image: "alpine:3.14"},
			},
		},
	}

	images := listImagesInPod(doc)

	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(images))
	}
}

// Benchmarks
func BenchmarkExtractFromDirectory(b *testing.B) {
	extractor := NewExtractor()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractor.ExtractFromDirectory(context.Background(), "testdata/complex-app", Options{})
	}
}

func BenchmarkParseImage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseImageRef("gcr.io/project/app:v1.2.0")
	}
}
