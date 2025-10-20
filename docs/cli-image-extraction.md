# Image Extraction Command

Extract container image references from Kubernetes manifests and Helm charts locally.

## Overview

The `replicated release extract-images` command reads YAML files and outputs a list of all container image references (like `nginx:1.19`, `postgres:14`) without making network calls or downloading images.

**Use cases:**
- 🔍 Discover what images your application uses
- 📦 Prepare airgap bundles (get list of images to download)
- 🔒 Security audits (find all images for scanning)
- ✅ Pre-flight validation (check before pushing to Replicated)

## Installation

The command is included in the Replicated CLI. Install or update:

```bash
# macOS
brew install replicatedhq/replicated/cli

# Linux
curl -s https://api.github.com/repos/replicatedhq/replicated/releases/latest \
  | grep "browser_download_url.*linux_amd64.tar.gz" \
  | cut -d : -f 2,3 \
  | tr -d \" \
  | xargs curl -sSL | tar xzv
```

## Quick Start

```bash
# Extract from manifest directory
replicated release extract-images --yaml-dir ./manifests

# Extract from Helm chart
replicated release extract-images --chart ./mychart.tgz

# Get JSON output for scripting
replicated release extract-images --yaml-dir ./manifests -o json
```

## Basic Usage

### Extract from Manifest Directory

```bash
replicated release extract-images --yaml-dir ./path/to/manifests
```

**Output:**
```
IMAGE              TAG       REGISTRY     SOURCE
library/nginx      1.19      docker.io    Deployment/web-app
library/postgres   14        docker.io    StatefulSet/database
library/redis      6.2       docker.io    Deployment/cache

Warnings:
⚠  redis:latest - Image uses 'latest' tag which is not recommended for production

Found 3 unique images
```

### Extract from Helm Chart

```bash
# From chart directory
replicated release extract-images --chart ./mychart/

# From packaged chart
replicated release extract-images --chart ./mychart-1.0.0.tgz
```

### Extract with Custom Helm Values

```bash
# Use custom values file
replicated release extract-images \
  --chart ./mychart.tgz \
  --values prod-values.yaml

# Set values on command line
replicated release extract-images \
  --chart ./mychart/ \
  --set image.tag=2.0 \
  --set replicaCount=5
```

## Output Formats

### Table Format (Default)

Human-readable table with image details and warnings:

```bash
replicated release extract-images --yaml-dir ./manifests -o table
```

### JSON Format

Machine-readable output for scripting:

```bash
replicated release extract-images --yaml-dir ./manifests -o json
```

**Example output:**
```json
{
  "images": [
    {
      "raw": "nginx:1.19",
      "registry": "docker.io",
      "repository": "library/nginx",
      "tag": "1.19",
      "digest": "",
      "sources": [
        {
          "file": "deployment.yaml",
          "kind": "Deployment",
          "name": "web-app",
          "container": "nginx",
          "containerType": "container"
        }
      ]
    }
  ],
  "warnings": [],
  "summary": {
    "total": 1,
    "unique": 1
  }
}
```

### List Format

Simple newline-separated list for piping:

```bash
replicated release extract-images --yaml-dir ./manifests -o list
```

**Output:**
```
nginx:1.19
postgres:14
redis:6.2
```

**Use with other tools:**
```bash
# Pull all images
replicated release extract-images --yaml-dir ./manifests -o list | xargs -I {} docker pull {}

# Save to file
replicated release extract-images --yaml-dir ./manifests -o list > images.txt

# Count images
replicated release extract-images --yaml-dir ./manifests -o list | wc -l
```

## Advanced Usage

### Show All Occurrences (No Deduplication)

```bash
replicated release extract-images --yaml-dir ./manifests --show-duplicates
```

Shows every occurrence of each image, useful for finding where duplicates exist.

### Suppress Warnings

```bash
replicated release extract-images --yaml-dir ./manifests --no-warnings
```

Useful when you just want the image list without validation warnings.

### Helm Chart with Multiple Values Files

```bash
replicated release extract-images \
  --chart ./mychart.tgz \
  --values base-values.yaml \
  --values prod-values.yaml \
  --set image.tag=override
```

Values are merged in order, with `--set` taking highest precedence.

### Custom Namespace for Helm Rendering

```bash
replicated release extract-images \
  --chart ./mychart/ \
  --namespace production
```

## Common Scenarios

### Scenario 1: Pre-Push Validation

Before pushing a release, check what images it contains:

```bash
cd my-kots-app/manifests
replicated release extract-images --yaml-dir .

# Review output for unexpected images or warnings
# Fix any issues
# Then push to Replicated
```

### Scenario 2: Airgap Bundle Preparation

Get a list of all images to download for offline installation:

```bash
# Get list
replicated release extract-images --yaml-dir ./manifests -o list > images.txt

# Download all images
cat images.txt | while read img; do
    docker pull "$img"
    docker save "$img" -o "$(echo $img | tr '/:' '_').tar"
done

# Create bundle
tar czf airgap-images.tar.gz *.tar
```

### Scenario 3: Security Audit

Extract images with warnings for security review:

```bash
replicated release extract-images --yaml-dir ./manifests -o json > audit.json

# Send to security team
# Or scan for vulnerabilities
cat audit.json | jq -r '.images[].raw' | while read img; do
    trivy image "$img"
done
```

### Scenario 4: Helm Values Testing

Test different Helm configurations:

```bash
# Development environment
replicated release extract-images \
  --chart ./mychart/ \
  --values dev-values.yaml \
  -o list

# Production environment
replicated release extract-images \
  --chart ./mychart/ \
  --values prod-values.yaml \
  -o list

# Compare outputs
diff <(cmd1) <(cmd2)
```

### Scenario 5: CI/CD Integration

Fail builds on image issues:

```bash
#!/bin/bash
# In CI pipeline

# Extract images
OUTPUT=$(replicated release extract-images --yaml-dir ./manifests -o json)

# Check for warnings
WARNINGS=$(echo "$OUTPUT" | jq '.warnings | length')

if [ "$WARNINGS" -gt 0 ]; then
    echo "Image warnings detected:"
    echo "$OUTPUT" | jq -r '.warnings[] | "⚠ \(.image): \(.message)"'
    exit 1
fi

echo "✓ All images validated successfully"
```

## Warnings Explained

### ⚠ latest-tag

**Issue:** Image uses the `:latest` tag

**Why it matters:** The `:latest` tag is mutable and can change, causing unexpected updates or broken deployments.

**Fix:**
```yaml
# Bad
image: nginx:latest

# Good
image: nginx:1.21.6
```

### ⚠ no-tag

**Issue:** Image has no tag specified (defaults to `:latest`)

**Why it matters:** Same as `latest-tag` - unpredictable behavior.

**Fix:**
```yaml
# Bad
image: nginx

# Good  
image: nginx:1.21.6
```

### ⚠ insecure-registry

**Issue:** Image uses HTTP registry (not HTTPS)

**Why it matters:** Security risk - images could be tampered with in transit.

**Fix:**
```yaml
# Bad
image: http://my-registry.com/app:v1

# Good
image: https://my-registry.com/app:v1
# Or use a secure registry
```

### ⚠ unqualified-name

**Issue:** No registry specified (assumes Docker Hub)

**Why it matters:** May not work in airgap or private environments.

**Fix:**
```yaml
# Less clear
image: nginx:1.19

# More explicit
image: docker.io/library/nginx:1.19
```

## Supported Resources

### Kubernetes Resources

- ✅ Pod
- ✅ Deployment
- ✅ StatefulSet
- ✅ DaemonSet
- ✅ ReplicaSet
- ✅ Job
- ✅ CronJob

### Container Types

- ✅ `containers` - Main application containers
- ✅ `initContainers` - Initialization containers
- ✅ `ephemeralContainers` - Debug containers

### KOTS Resources

- ✅ Application (`spec.additionalImages`, `spec.excludedImages`)
- ✅ Preflight (`spec.collectors[].run.image`)
- ✅ SupportBundle (`spec.collectors[].run.image`)
- ✅ Collector (`spec.collectors[].run.image`)

### Helm Charts

- ✅ Chart directories
- ✅ Packaged charts (.tgz)
- ✅ Custom values files
- ✅ Command-line value overrides

## Troubleshooting

### No images found

**Problem:** Command returns 0 images

**Possible causes:**
1. Wrong directory path
2. No YAML files in directory
3. YAML files don't contain supported resources

**Solutions:**
```bash
# Check directory exists
ls -la ./manifests/

# Check for YAML files
find ./manifests/ -name "*.yaml" -o -name "*.yml"

# Try with absolute path
replicated release extract-images --yaml-dir /full/path/to/manifests
```

### Helm chart rendering fails

**Problem:** Error rendering Helm chart

**Solutions:**
```bash
# Test with helm directly
helm template ./mychart/

# Check Chart.yaml is valid
cat ./mychart/Chart.yaml

# Try without custom values first
replicated release extract-images --chart ./mychart/
```

### Malformed YAML errors

**Problem:** Parse errors on some files

**Solution:** The command continues processing other files. Check the specific file:

```bash
# Validate YAML
yamllint problematic-file.yaml

# Or use yq
yq eval . problematic-file.yaml
```

### Wrong images extracted

**Problem:** Missing or incorrect images

**Check:**
1. Verify YAML structure matches Kubernetes format
2. Check that images are in `spec.containers[].image` or `spec.template.spec.containers[].image`
3. Use `--show-duplicates` to see all occurrences

## FAQ

**Q: Does this command download images?**

No. It only reads YAML files and extracts text strings. No network calls are made.

**Q: Can I use this offline?**

Yes! It's pure local file parsing with no network dependency.

**Q: Does it work with KOTS applications?**

Yes! It extracts images from KOTS Application, Preflight, and SupportBundle resources.

**Q: What about images in HelmChart CRs?**

Not yet supported. Coming in V2. For now, render the chart manually and extract from that.

**Q: Can I exclude certain images?**

Yes, use KOTS Application `spec.excludedImages`:

```yaml
apiVersion: kots.io/v1beta1
kind: Application
spec:
  excludedImages:
  - internal-debug-tool:latest
```

**Q: Does it validate that images exist in registries?**

No. It only extracts references, it doesn't check if they're pullable.

**Q: Can I use this in CI/CD?**

Yes! Use `-o json` or `-o list` for machine-readable output. The command exits with code 0 on success.

**Q: What's the difference from `replicated release image ls`?**

| Command | When | What |
|---------|------|------|
| `extract-images` | Before push | Extracts from **local** files |
| `image ls` | After promote | Shows images from **promoted** release |

Use `extract-images` during development, `image ls` for released versions.

**Q: How fast is it?**

Very fast - pure YAML parsing with no network. Typical performance:
- 10 files: < 100ms
- 100 files: < 1s
- 1000 files: < 5s

## Examples

All examples use the test fixtures included with the CLI:

```bash
# Simple deployment
replicated release extract-images \
  --yaml-dir pkg/imageextract/testdata/simple-deployment

# Complex multi-resource app  
replicated release extract-images \
  --yaml-dir pkg/imageextract/testdata/complex-app

# Helm chart
replicated release extract-images \
  --chart pkg/imageextract/testdata/helm-chart

# Multi-document YAML
replicated release extract-images \
  --yaml-dir pkg/imageextract/testdata/multi-doc
```

## Related Commands

- `replicated release create` - Create a new release
- `replicated release lint` - Lint manifests (includes image validation)
- `replicated release image ls` - List images from promoted release

## Getting Help

```bash
# Command help
replicated release extract-images --help

# General help
replicated release --help
```

## Feedback

Found a bug or have a feature request? Please open an issue in the Replicated CLI repository.

