# .replicated `lint` Field (Minimal Spec)
This defines only the minimal structure for the new linter. YAML and JSON are both supported; YAML shown here.
## Format
```yaml
repl-lint:
    version: 1                           # lint config schema version
    enabled: true                        # turn linting on/off
    linters:
        helm:
            enabled: true                # run helm lint
            strict: false                # if true, treat warnings as errors
        preflight:
            enabled: true
            strict: true
        support-bundle:
            enabled: true
            strict: false
        embedded-cluster:                # embedded cluster linter validates EC config files
            enabled: false
            strict: false
        kots:                            # kots linter does not exist as of yet
            enabled: false
            strict: false
    tools:                               # tool resolution (optional)
```
Notes:
- Only keys listed above are recognized in this minimal spec. Unknown keys are rejected.
- Omit optional sections to use defaults.
- `version` controls config parsing behavior; defaults to 1 if omitted.
## Examples
1) Pin Helm version (strict mode):
```yaml
  appId: ""
  appSlug: "" 
  promoteToChannelIds: []
  promoteToChannelNames: []
  charts: [
    {
      path: "./chart/something",
      chartVersion: "",
      appVersion: "",
    },
    {
        path: "./chart/new-chart/*",
        chartVersion: "",
        appVersion: "",
    }
  ]
  preflights: [
    {
        path: "./preflights/stuff",
        chartName: "something",
        chartVersion: "1.0.0",
    }
  ]
  releaseLabel: ""  ## some sort of semver pattern?
  manifests: ["replicated/**/*.yaml"]
    repl-lint:
        version: 1
        linters:
            helm:
                disbabled: false                
                strict: false                
            preflight:
                disabled: false
                strict: true
            support-bundle:
                disabled: false
                strict: false
            embedded-cluster:                
                disabled: true
                strict: false
            kots:
                disabled: true
                strict: false
        tools:
            helm: "3.14.4"
            preflight: "0.123.9"
            support-bundle: "0.123.9"
```

## Glob Pattern Support

The `replicated lint` command supports glob patterns for discovering files. This allows you to lint multiple charts, preflights, or manifests with a single pattern.

### Supported Patterns

- `*` - Matches any sequence of characters in a single directory level
- `**` - Matches zero or more directories recursively
- `?` - Matches any single character
- `[abc]` - Matches any character in the brackets
- `[a-z]` - Matches any character in the range
- `{alt1,alt2}` - Matches any of the alternatives (brace expansion)

### Examples

**Helm Charts:**
```yaml
charts:
  - path: "./charts/*"                    # All charts in charts/
  - path: "./charts/{app,api,web}"        # Specific charts only
  - path: "./environments/*/charts/*"     # Charts in any environment
```

**Preflights:**
```yaml
preflights:
  - path: "./preflights/**/*.yaml"        # All YAML files recursively
  - path: "./checks/{basic,advanced}.yaml" # Specific check files
```

**Manifests (Support Bundles):**
```yaml
manifests:
  - "./k8s/**/*.yaml"                     # All YAML in k8s/, recursively
  - "./manifests/{dev,staging,prod}/**/*" # Multiple environments
```

### Important Notes

**Recursive Matching (`**`):**
- `**` matches zero or more directories
- `./manifests/**/*.yaml` matches:
  - `manifests/app.yaml` (no subdirectory)
  - `manifests/base/deployment.yaml` (one level)
  - `manifests/overlays/prod/patch.yaml` (two levels)
  - Any depth recursively

**Brace Expansion (`{}`):**
- `{a,b,c}` expands to multiple separate patterns
- Useful for matching specific directories or files
- Cannot be nested: `{a,{b,c}}` is not supported

**Hidden Files:**
- Unlike shell behavior, glob patterns match hidden files (files starting with `.`)
- `*.yaml` WILL match `.hidden.yaml`
- To exclude hidden files, use explicit patterns that don't start with `.`

Inline directive examples:
- Ignore next line:
```yaml
# repl-lint-ignore-next
image: nginx:latest
```
- Ignore block:
```yaml
# repl-lint-ignore-start
apiVersion: v1
kind: ConfigMap
data:
  KEY: VALUE
# repl-lint-ignore-end
```
- Ignore file (place near top):
```yaml
# repl-lint-ignore-file
```

## Preflight Configuration

Preflight specs require a chart reference for template rendering. Configure preflights by specifying the chart name and version to use:

```yaml
charts:
  - path: "./chart"

preflights:
  - path: "./preflight.yaml"
    chartName: "my-chart"      # Must match chart name in Chart.yaml
    chartVersion: "1.0.0"       # Must match chart version in Chart.yaml
```

**Requirements:**
- Both `chartName` and `chartVersion` are required for each preflight
- The chart name/version must match the values in the chart's `Chart.yaml` file
- The chart must be listed in the `charts` section of your `.replicated` config

### Values File Location

Preflight template rendering requires a chart's values file. The CLI automatically locates this file using these rules:

1. **Checks for `values.yaml` in the chart root directory** (most common)
2. **Falls back to `values.yml`** if `values.yaml` doesn't exist
3. **Returns an error** if neither exists

**Expected Chart Structure:**
```
my-chart/
  ├── Chart.yaml        # Required: defines chart name and version
  ├── values.yaml       # Required: default values for templates
  ├── templates/        # Chart templates
  └── ...
```

**Note:** Custom values file paths are not currently supported. Values files must be named `values.yaml` or `values.yml` and located in the chart root directory.

## HelmChart Manifest Requirements

Every Helm chart configured in your `.replicated` file requires a corresponding `HelmChart` manifest (custom resource with `kind: HelmChart`). This manifest is essential for:

- **Preflight template rendering**: Charts are rendered with builder values before running preflight checks
- **Image extraction**: Images are extracted from chart templates for air gap bundle creation
- **Air gap bundle building**: Charts are packaged with specific values for offline installations

### Configuration

When charts are configured, you must also specify where to find their HelmChart manifests:

```yaml
charts:
  - path: "./charts/my-app"
  - path: "./charts/my-api"

manifests:
  - "./manifests/**/*.yaml"    # HelmChart manifests must be in these paths
```

**Important:** The `manifests` section is required whenever `charts` are configured. If you configure charts but omit manifests, linting will fail with a clear error message.

### HelmChart Manifest Structure

Each HelmChart manifest must specify the chart name and version that match the corresponding chart's `Chart.yaml`:

```yaml
apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: my-app
spec:
  chart:
    name: my-app           # Must match Chart.yaml name
    chartVersion: 1.0.0    # Must match Chart.yaml version
  builder: {}               # Values for rendering (can be empty)
```

The `spec.chart.name` and `spec.chart.chartVersion` fields must exactly match the `name` and `version` in your Helm chart's `Chart.yaml` file.

### Validation Behavior

The linter validates chart-to-HelmChart mapping before running other checks:

- **✅ Success**: All charts have matching HelmChart manifests
- **❌ Error**: One or more charts are missing HelmChart manifests (batch reports all missing)
- **⚠️ Warning**: HelmChart manifest exists but no corresponding chart is configured (orphaned manifest)

#### Error Example

```
Error: chart validation failed: Chart validation failed - 2 charts missing HelmChart manifests:
  - ./charts/frontend (frontend:1.0.0)
  - ./charts/database (database:1.5.0)

Each Helm chart requires a corresponding HelmChart manifest (kind: HelmChart).
Ensure the manifests are in paths specified in the 'manifests' section of .replicated config.
```

#### Warning Example

```
Warning: HelmChart manifest at "./manifests/old-app.yaml" (old-app:1.0.0) has no corresponding chart configured
```

### Auto-Discovery

When no `.replicated` config file exists, the linter automatically discovers all resources including:
- Helm charts (by finding `Chart.yaml` files)
- HelmChart manifests (by finding `kind: HelmChart` in YAML files)
- Preflights and Support Bundles

Auto-discovery validates that all discovered charts have corresponding HelmChart manifests.

### Troubleshooting

**Problem:** "charts are configured but no manifests paths provided"

**Solution:** Add a `manifests` section to your `.replicated` config:
```yaml
manifests:
  - "./manifests/**/*.yaml"
```

**Problem:** "Missing HelmChart manifest for chart"

**Solution:** Create a HelmChart manifest with matching name and version:
```yaml
apiVersion: kots.io/v1beta2
kind: HelmChart
metadata:
  name: <your-chart-name>
spec:
  chart:
    name: <your-chart-name>
    chartVersion: <your-chart-version>
  builder: {}
```

**Problem:** Warning about orphaned HelmChart manifest

**Solution:** Either add the corresponding chart to your configuration or remove the unused HelmChart manifest. Warnings are informational and won't cause linting to fail.

## Embedded Cluster Configuration

The embedded-cluster linter validates Embedded Cluster configuration files. Embedded Cluster is Replicated's solution for packaging Kubernetes and your application together as a single appliance-style installer.

### What It Validates

The embedded-cluster linter validates:
- Configuration schema correctness
- Kubernetes version compatibility
- Extension configurations
- Network and storage settings
- High availability settings
- Custom branding configuration

### Platform Requirements

**Important:** The embedded-cluster linter binary is currently only available for **Linux AMD64**.

- ✅ **Linux (x86_64/amd64)**: Full support
- ❌ **macOS**: Not currently available
- ❌ **Windows**: Not currently available
- ❌ **ARM64**: Not currently available

If you're on an unsupported platform:
- The linter will fail gracefully with a clear error message
- Other linters (helm, preflight, support-bundle) will continue running
- Consider running in a Linux container or CI environment

### Configuration

Enable the embedded-cluster linter in your `.replicated` file:

```yaml
repl-lint:
  version: 1
  linters:
    embedded-cluster:
      enabled: true
      strict: false    # Set to true to treat warnings as errors
  tools:
    embedded-cluster: "latest"   # Optional: pin to specific version
```

### Auto-Discovery

When no `.replicated` config exists, the linter automatically discovers embedded-cluster configs by:
- Finding files with `kind: Config` and `apiVersion: embeddedcluster.replicated.com/v1beta1`
- Validating only 0 or 1 config exists (multiple configs are not supported)

### Multiple Config Validation

**Only 0 or 1 embedded cluster config is allowed per project.**

If multiple configs are detected:
- Each config will show as failed with a clear error message
- Other linters (helm, preflight, support-bundle) continue running
- The linting command returns a non-zero exit code

### Example Configuration

**Minimal `.replicated` with embedded-cluster:**
```yaml
appId: ""
appSlug: "my-app"
manifests:
  - "./embedded-cluster/*.yaml"

repl-lint:
  version: 1
  linters:
    embedded-cluster:
      enabled: true
      strict: true
  tools:
    embedded-cluster: "latest"
```

**Example embedded-cluster config file (ec-config.yaml):**
```yaml
apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
metadata:
  name: my-app-config
spec:
  version: "1.33+k8s-1.33"
  roles:
    controller:
      name: "Controller"
      description: "Kubernetes controller node"

  extensions:
    helm:
      repositories:
        - name: my-repo
          url: https://charts.example.com

  network:
    podCIDR: "10.244.0.0/16"
    serviceCIDR: "10.96.0.0/12"

  unsupportedOverrides:
    k0s: |
      apiVersion: k0s.k0sproject.io/v1beta1
      kind: ClusterConfig
      spec:
        network:
          provider: calico
```

### Output Format

The embedded-cluster linter returns JSON output with structured validation results:

```json
{
  "files": [
    {
      "path": "embedded-cluster/config.yaml",
      "valid": true,
      "errors": [],
      "warnings": [],
      "infos": []
    }
  ]
}
```

**Validation Messages:**
- **errors**: Schema violations, invalid values, unsupported configurations
- **warnings**: Deprecated fields, potential issues, best practice violations
- **infos**: Informational messages about configuration choices

### Troubleshooting

**Problem:** "embedded-cluster binaries are only available for linux-amd64"

**Solution:**
- Run linting in a Linux environment or Docker container
- Use CI/CD on Linux runners
- The error won't block other linters from running

**Problem:** "Multiple embedded cluster configs found"

**Solution:** Only one embedded cluster config is allowed per project. Remove duplicate configs or move them to separate projects.

**Problem:** Validation errors about Kubernetes version

**Solution:** Ensure the `spec.version` field uses a supported Kubernetes version. Check the embedded-cluster documentation for supported versions.
