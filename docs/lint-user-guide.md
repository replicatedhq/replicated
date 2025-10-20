# Replicated Lint User Guide

This guide covers everything you need to know about using the `replicated lint` command to validate your Helm charts, Preflight specs, and Support Bundle specs.

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Helm Chart Linting](#helm-chart-linting)
- [Preflight Linting](#preflight-linting)
- [Support Bundle Linting](#support-bundle-linting)
- [Tool Version Management](#tool-version-management)
- [CI/CD Integration](#cicd-integration)
- [Troubleshooting](#troubleshooting)

## Introduction

The `replicated lint` command provides local validation for your application resources before deployment. It helps catch configuration errors early in the development process.

### What Gets Validated

- **Helm Charts**: Chart structure, templates, and values
- **Preflight Specs**: Troubleshoot.sh preflight check definitions
- **Support Bundle Specs**: Troubleshoot.sh support bundle collector configurations

### Benefits

- ✅ **Catch errors early**: Find issues before deployment
- ✅ **No API required**: Runs entirely locally
- ✅ **CI/CD friendly**: Exit codes for automated pipelines
- ✅ **Auto-download tools**: Automatically fetches required linting tools
- ✅ **Version pinning**: Ensure consistent linting across environments

## Quick Start

### 1. Create Configuration File

Create a `.replicated` file in your project root:

```yaml
spec-version: 1

charts:
  - path: ./my-chart

preflights:
  - path: ./preflight.yaml

manifests:
  - ./manifests/*.yaml

repl-lint:
  linters:
    helm:
      disabled: false
    preflight:
      disabled: false
    support-bundle:
      disabled: false
```

### 2. Run Linting

```bash
replicated lint
```

### 3. Review Output

```
==> Linting chart: ./my-chart
No issues found

Summary for ./my-chart: 0 error(s), 0 warning(s), 0 info
Status: Passed

==> Overall Summary
charts linted: 1
charts passed: 1
charts failed: 0
Total errors: 0
Total warnings: 0
Total info: 0

Overall Status: Passed
```

## Configuration

### File Format

The `.replicated` file uses YAML (or JSON) format. Here's a comprehensive example:

```yaml
spec-version: 1

# Application metadata (optional)
appId: "your-app-id"
appSlug: "your-app"

# Helm charts to lint
charts:
  - path: ./charts/app-chart
    chartVersion: "1.0.0"
    appVersion: "1.0.0"

  # Glob patterns are supported
  - path: ./charts/*

# Preflight specs to lint
preflights:
  # Single file
  - path: ./preflight.yaml

  # Glob pattern
  - path: ./preflights/*.yaml

  # Optional: provide values context
  - path: ./preflight-with-values.yaml
    valuesPath: ./charts/app-chart

# Kubernetes manifests (for support bundle discovery)
manifests:
  - ./manifests/**/*.yaml
  - ./k8s/*.yaml

# Linter configuration
repl-lint:
  version: 1  # Config schema version

  linters:
    helm:
      disabled: false  # Enable/disable helm linting

    preflight:
      disabled: false  # Enable/disable preflight linting

    support-bundle:
      disabled: false  # Enable/disable support bundle linting

  # Optional: pin tool versions
  tools:
    helm: "3.14.4"
    preflight: "0.123.9"
    support-bundle: "0.123.9"
```

### Glob Patterns

Glob patterns are supported for all path fields:

| Pattern | Matches |
|---------|---------|
| `*` | Any files in the current directory |
| `**` | Any files in current and subdirectories (recursive) |
| `*.yaml` | All YAML files in current directory |
| `**/*.yaml` | All YAML files recursively |
| `charts/*/` | All subdirectories in `charts/` |

**Examples:**
```yaml
charts:
  - path: ./charts/*              # All charts in charts/
  - path: ./charts/app-*          # Charts starting with "app-"

preflights:
  - path: ./preflights/**/*.yaml  # All YAML files recursively

manifests:
  - path: ./**/*.yaml             # All YAML in project recursively
```

## Helm Chart Linting

### How It Works

The Helm linter validates your charts using the official `helm lint` command. It checks:

- Chart.yaml structure and required fields
- Template syntax and Go template functions
- Values.yaml structure
- Chart dependencies
- Best practices and conventions

### Configuration

```yaml
charts:
  - path: ./charts/my-chart
    chartVersion: "1.0.0"
    appVersion: "1.0.0"
```

### Chart Requirements

Each chart directory must contain:
- `Chart.yaml` (or `Chart.yml`)
- Standard Helm chart structure

### Common Errors and Fixes

#### Error: "Chart.yaml not found"

**Cause:** The path doesn't point to a valid Helm chart directory.

**Fix:** Ensure the path contains a `Chart.yaml` file:
```bash
ls ./charts/my-chart/Chart.yaml
```

#### Error: "template: invalid syntax"

**Cause:** Go template syntax error in chart templates.

**Fix:** Review the template file mentioned in the error. Common issues:
- Unclosed brackets: `{{ .Values.foo` → `{{ .Values.foo }}`
- Undefined values: Reference a value that doesn't exist in values.yaml
- Invalid function calls

#### Warning: "icon is recommended"

**Cause:** Chart.yaml is missing optional but recommended fields.

**Fix:** Add the recommended field to Chart.yaml:
```yaml
icon: https://example.com/icon.png
```

### Multiple Charts

Lint multiple charts at once:

```yaml
charts:
  - path: ./charts/frontend
  - path: ./charts/backend
  - path: ./charts/database
```

Or use glob patterns:
```yaml
charts:
  - path: ./charts/*
```

## Preflight Linting

### How It Works

The Preflight linter validates your troubleshoot.sh preflight specifications using the official `preflight lint` command. It checks:

- YAML syntax
- Required fields (apiVersion, kind, metadata, spec)
- Analyzer and collector definitions
- Template syntax (for v1beta3)
- Best practices

### Configuration

```yaml
preflights:
  - path: ./preflight.yaml
  - path: ./preflights/*.yaml
```

### Spec Requirements

- Valid YAML syntax
- `apiVersion`: `troubleshoot.sh/v1beta2` or `troubleshoot.sh/v1beta3`
- `kind`: `Preflight`
- Required sections: `metadata`, `spec`

### Common Errors and Fixes

#### Error: "Missing required field: spec"

**Cause:** Preflight spec is missing the `spec` section.

**Fix:** Add a spec section with at least one collector or analyzer:
```yaml
apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: my-preflight
spec:
  collectors:
    - clusterInfo: {}
  analyzers:
    - clusterVersion:
        outcomes:
          - pass:
              when: ">=1.19.0"
              message: Cluster version is supported
```

#### Error: "YAML syntax error"

**Cause:** Invalid YAML formatting.

**Fix:** Check for common YAML issues:
- Incorrect indentation
- Missing colons
- Unquoted special characters
- Mixing tabs and spaces

#### Warning: "Some collectors are missing docString"

**Cause:** v1beta3 specs should include docString for better documentation.

**Fix:** Add docString to collectors:
```yaml
collectors:
  - clusterInfo:
      docString: "Collects information about the Kubernetes cluster"
```

### With Values Context

If your preflight spec uses templating that references Helm chart values:

```yaml
preflights:
  - path: ./preflight.yaml
    valuesPath: ./charts/my-chart  # Path to chart with values.yaml
```

## Support Bundle Linting

### How It Works

The Support Bundle linter **automatically discovers** support bundle specs from your manifest files. It:

1. Expands glob patterns in the `manifests` array
2. Reads each YAML file
3. Identifies files containing `kind: SupportBundle`
4. Validates each support bundle using `support-bundle lint`

### Configuration

```yaml
# Support bundles are discovered from these globs
manifests:
  - ./manifests/**/*.yaml
  - ./k8s/*.yaml

repl-lint:
  linters:
    support-bundle:
      disabled: false  # Enable auto-discovery and linting
```

### No Explicit Configuration Required

Unlike Helm charts and Preflight specs, you don't need to explicitly list support bundle files. They're automatically discovered from your manifests.

### Spec Requirements

- Valid YAML syntax
- `apiVersion`: `troubleshoot.sh/v1beta2` or `troubleshoot.sh/v1beta3`
- `kind`: `SupportBundle`
- Required sections: `metadata`, `spec`
- At least one collector defined

### Common Errors and Fixes

#### Error: "Support bundle spec must have at least one collector"

**Cause:** The spec section is empty or missing collectors.

**Fix:** Add at least one collector:
```yaml
apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: my-support-bundle
spec:
  collectors:
    - clusterInfo: {}
    - clusterResources: {}
```

#### Error: "Missing 'spec' section"

**Cause:** Support bundle is missing the required spec section.

**Fix:** Add a spec section with collectors:
```yaml
spec:
  collectors:
    - logs:
        selector:
          - app=myapp
```

#### Warning: "Some collectors are missing docString"

**Cause:** v1beta3 recommends docString for documentation.

**Fix:** Add docString to collectors:
```yaml
collectors:
  - logs:
      docString: "Collects application logs from all pods"
      selector:
        - app=myapp
```

### Mixed Manifest Files

Support bundles can be co-located with other Kubernetes resources:

```yaml
# deployment.yaml - Contains both a Deployment and SupportBundle
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
# ... deployment spec ...
---
apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: myapp-bundle
spec:
  collectors:
    - logs:
        selector:
          - app=myapp
```

The linter will ignore the Deployment and only validate the SupportBundle.

### Known Limitations

**v1beta3 Multi-Document Files:**

The upstream `support-bundle lint` tool validates ALL documents in a file. If your v1beta3 examples include both SupportBundles and other Kubernetes resources (Secrets, ConfigMaps) in the same file, you may see errors about those resources.

**Workaround:** Extract SupportBundle specs into separate files for linting.

## Tool Version Management

### Auto-Download

Linting tools are automatically downloaded and cached in `~/.replicated/tools/`:

```
~/.replicated/tools/
├── helm/
│   └── 3.14.4/
│       └── darwin-arm64/
│           └── helm
├── preflight/
│   └── 0.123.9/
│       └── darwin-arm64/
│           └── preflight
└── support-bundle/
    └── 0.123.9/
        └── darwin-arm64/
            └── support-bundle
```

### Version Pinning

Pin specific versions for reproducible builds:

```yaml
repl-lint:
  tools:
    helm: "3.14.4"
    preflight: "0.123.9"
    support-bundle: "0.123.9"
```

**Benefits:**
- Consistent linting across environments
- Avoid surprises from tool updates
- Easy version upgrades (change version, tools auto-download)

### Default Versions

If not specified, default versions are used (see `pkg/tools/types.go`):
- Helm: 3.14.4
- Preflight: 0.123.9
- Support Bundle: 0.123.9

## CI/CD Integration

### Exit Codes

The `replicated lint` command uses standard exit codes:

- **0**: All linting passed (no errors)
- **Non-zero**: Linting failed (errors found)

### GitHub Actions

```yaml
name: Lint

on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install Replicated CLI
        run: |
          curl -o install.sh -sSL https://raw.githubusercontent.com/replicatedhq/replicated/master/install.sh
          sudo bash ./install.sh

      - name: Run Linting
        run: replicated lint
```

### CircleCI

```yaml
version: 2.1

jobs:
  lint:
    docker:
      - image: cimg/base:stable
    steps:
      - checkout

      - run:
          name: Install Replicated CLI
          command: |
            curl -o install.sh -sSL https://raw.githubusercontent.com/replicatedhq/replicated/master/install.sh
            sudo bash ./install.sh

      - run:
          name: Run Linting
          command: replicated lint

workflows:
  version: 2
  build:
    jobs:
      - lint
```

### GitLab CI

```yaml
lint:
  stage: test
  image: ubuntu:latest
  before_script:
    - apt-get update && apt-get install -y curl
    - curl -o install.sh -sSL https://raw.githubusercontent.com/replicatedhq/replicated/master/install.sh
    - bash ./install.sh
  script:
    - replicated lint
```

### Local Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
echo "Running replicated lint..."
replicated lint

if [ $? -ne 0 ]; then
    echo "❌ Linting failed. Fix errors before committing."
    exit 1
fi

echo "✅ Linting passed"
```

Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

## Troubleshooting

### Issue: "no charts found in .replicated config"

**Cause:** The `charts` array is empty or not defined.

**Solution:** Add at least one chart to the configuration:
```yaml
charts:
  - path: ./charts/my-chart
```

### Issue: "no preflights found in .replicated config"

**Cause:** The `preflights` array is empty or not defined, but preflight linting is enabled.

**Solution:** Either add preflight specs or disable the linter:
```yaml
repl-lint:
  linters:
    preflight:
      disabled: true
```

### Issue: "Chart.yaml or Chart.yml not found"

**Cause:** The chart path doesn't point to a valid Helm chart directory.

**Solution:** Verify the path contains a Chart.yaml:
```bash
ls ./charts/my-chart/Chart.yaml
```

### Issue: Tool download fails

**Cause:** Network issues or missing permissions.

**Solutions:**
1. Check network connectivity
2. Verify `~/.replicated/tools/` directory exists and is writable:
   ```bash
   mkdir -p ~/.replicated/tools/
   chmod 755 ~/.replicated/tools/
   ```
3. Try manually downloading the tool

### Issue: "glob pattern matched no files"

**Cause:** The glob pattern doesn't match any files.

**Solution:** Test the glob pattern:
```bash
ls ./charts/*
ls ./preflights/*.yaml
```

Adjust the pattern as needed.

### Issue: Linter is disabled but still getting messages

**Cause:** Configuration syntax error (likely using `enabled` instead of `disabled`).

**Solution:** Use `disabled` field, not `enabled`:
```yaml
# ❌ Wrong
repl-lint:
  linters:
    helm:
      enabled: false

# ✅ Correct
repl-lint:
  linters:
    helm:
      disabled: true
```

### Debug Mode

Enable debug output for troubleshooting:

```bash
replicated lint --debug
```

This shows:
- Configuration file loading
- Tool resolution and downloads
- Detailed error messages

## See Also

- [Configuration Reference](lint-format.md) - Detailed configuration options
- [.replicated.example](../.replicated.example) - Complete example configuration
- [README.md](../README.md) - General CLI documentation
- [Troubleshoot.sh Documentation](https://troubleshoot.sh) - Preflight and Support Bundle specs
