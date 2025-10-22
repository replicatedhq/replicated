# Image Extraction Test Fixtures

This directory contains test fixtures for image extraction with builder values.

## Test Scenarios

### chart-with-required-values-test/
**Purpose**: Tests successful image extraction when builder values are provided via HelmChart manifest.

- **Chart**: `test-required-app:1.0.0`
- **Required Values**: `database.image.repository`, `database.image.tag`
- **HelmChart Manifest**: Provides matching name:version with builder values
- **Expected Result**: Successfully extracts `postgres:15-alpine`
- **Tests**: Builder values enable rendering of charts with required values

### simple-chart-test/
**Purpose**: Tests backward compatibility - charts without required values work without HelmChart manifests.

- **Chart**: `simple-app:1.0.0`
- **Required Values**: None (has defaults)
- **HelmChart Manifest**: Not needed
- **Expected Result**: Successfully extracts `nginx:1.21`
- **Tests**: Charts with default values don't need builder values

### multi-image-chart-test/
**Purpose**: Tests extraction of multiple images from a single chart.

- **Chart**: `multi-image-app:2.0.0`
- **Required Values**: `frontend.image.*`, `backend.image.*`, `cache.image.*`
- **HelmChart Manifest**: Provides builder values for all three services
- **Expected Result**: Extracts 3 images:
  - `nginx:1.21-alpine` (frontend)
  - `node:18-alpine` (backend)
  - `redis:7-alpine` (cache)
- **Tests**: Multiple image extraction and deduplication

### non-matching-helmchart-test/
**Purpose**: Tests that non-matching HelmChart manifests don't apply builder values.

- **Chart**: `app-requiring-values:1.0.0`
- **HelmChart Manifest**: Has name `different-app` (doesn't match)
- **Expected Result**: Fails to render (0 images), warning about missing values
- **Tests**: Matching logic requires exact name:version match

## Usage in Tests

Tests should reference these fixtures using relative paths:
```go
chartPath := filepath.Join("testdata", "image-extraction", "chart-with-required-values-test", "chart")
manifestPath := filepath.Join("testdata", "image-extraction", "chart-with-required-values-test", "manifests", "*.yaml")
```

## Structure

Each test directory follows this structure:
```
test-name/
├── chart/
│   ├── Chart.yaml          # Chart metadata (name, version)
│   ├── values.yaml         # Default values (may have required fields)
│   └── templates/
│       └── deployment.yaml # Templates with image references
└── manifests/
    └── helmchart.yaml      # HelmChart CR with builder values (if needed)
```

