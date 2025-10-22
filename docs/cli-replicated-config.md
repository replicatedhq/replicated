# .replicated Configuration File

The `.replicated` configuration file is used to define your project structure, resource locations, and linting preferences for the Replicated CLI. This file enables commands like `replicated release create` and `replicated release lint` to automatically discover and process your application resources.

## File Location

The CLI searches for `.replicated` or `.replicated.yaml` starting from the current directory and walking up the directory tree. This supports both:
- Single-repository projects with one `.replicated` file at the root
- Monorepo projects with multiple `.replicated` files at different levels (configs are merged)

## Creating a Configuration File

### Using `replicated config init`

The easiest way to create a `.replicated` file is using the interactive `init` command:

```bash
# Interactive mode with prompts
replicated config init

# Skip auto-detection
replicated config init --skip-detection
```

The `init` command will:
1. Auto-detect Helm charts, preflight specs, and support bundles in your project
2. Prompt for application configuration (app ID/slug)
3. Guide you through linting setup
4. Generate a `.replicated` file with your selections

### Manual Creation

You can also create a `.replicated` file manually using YAML format.

## Configuration Structure

Here's a complete example with all available fields:

```yaml
# Application identification
appId: ""                      # Your application ID (optional)
appSlug: ""                    # Your application slug (optional, more commonly used)

# Automatic promotion channels
promoteToChannelIds: []        # List of channel IDs to promote to
promoteToChannelNames: []      # List of channel names to promote to (e.g., ["beta", "stable"])

# Helm charts
charts:
  - path: "./helm-chart"       # Path or glob pattern to chart directory
    chartVersion: ""           # Override chart version (optional)
    appVersion: ""             # Override app version (optional)

# Preflight checks
preflights:
  - path: "./preflights/**"    # Path or glob pattern to preflight specs
    valuesPath: "./helm-chart" # Path to helm chart for template rendering (required for v1beta3 preflights)

# Kubernetes manifests and support bundles
manifests: ["./support-bundles/**"]  # Glob patterns for manifest files

# Release labeling
releaseLabel: ""               # Label pattern for releases (e.g., "v{{.Semver}}")

# Linting configuration
repl-lint:
  version: 1
  linters:
    helm:
      disabled: false          # Enable/disable Helm linting
    preflight:
      disabled: false          # Enable/disable preflight linting
    support-bundle:
      disabled: false          # Enable/disable support bundle linting
  tools:
    helm: "latest"             # Helm version (semantic version or "latest")
    preflight: "latest"        # Preflight version (semantic version or "latest")
    support-bundle: "latest"   # Support bundle version (semantic version or "latest")
```

## Field Reference

### Application Fields

#### `appId` (string, optional)
Your Replicated application ID. You can find this in the Vendor Portal at vendor.replicated.com.

#### `appSlug` (string, optional)
Your Replicated application slug. This is the human-readable identifier for your app (more commonly used than `appId`).

**Example:**
```yaml
appSlug: "my-application"
```

### Release Configuration

#### `promoteToChannelIds` (array of strings, optional)
List of channel IDs to automatically promote releases to when using `replicated release create`.

#### `promoteToChannelNames` (array of strings, optional)
List of channel names to automatically promote releases to. More convenient than using IDs.

**Example:**
```yaml
promoteToChannelNames: ["beta", "stable"]
```

#### `releaseLabel` (string, optional)
Template string for release labels. Supports Go template syntax.

**Example:**
```yaml
releaseLabel: "v{{.Semver}}"
```

### Resource Configuration

#### `charts` (array of objects, optional)
Helm charts to include in releases and lint operations.

**Fields:**
- `path` (string, required): Path or glob pattern to chart directory (e.g., `./chart` or `./charts/*`)
- `chartVersion` (string, optional): Override the chart version
- `appVersion` (string, optional): Override the app version

**Example:**
```yaml
charts:
  - path: "./helm-chart"
  - path: "./charts/app-*"    # Glob patterns supported
    chartVersion: "1.2.3"
```

#### `preflights` (array of objects, optional)
Preflight check specifications to validate before installation.

**Fields:**
- `path` (string, required): Path or glob pattern to preflight spec files
- `valuesPath` (string, optional but recommended): Path to Helm chart directory for template rendering. Required for v1beta3 preflights that use templating.

**Example:**
```yaml
preflights:
  - path: "./preflights/preflight.yaml"
    valuesPath: "./helm-chart"  # Chart directory for rendering templates
  - path: "./preflights/**/*.yaml"  # Glob pattern
    valuesPath: "./helm-chart"
```

#### `manifests` (array of strings, optional)
Glob patterns for Kubernetes manifest files, including support bundle specs. These are searched for support bundle specifications during linting.

**Example:**
```yaml
manifests:
  - "./manifests/**/*.yaml"
  - "./support-bundles/**"
```

### Linting Configuration

#### `repl-lint` (object, optional)
Configuration for the linting subsystem.

**Fields:**
- `version` (integer): Configuration version (currently `1`)
- `linters` (object): Enable/disable specific linters
- `tools` (map): Tool versions to use

**Example:**
```yaml
repl-lint:
  version: 1
  linters:
    helm:
      disabled: false          # Enable Helm linting
    preflight:
      disabled: false          # Enable preflight linting
    support-bundle:
      disabled: false          # Enable support bundle linting
    embedded-cluster:
      disabled: true           # Disable embedded cluster linting
    kots:
      disabled: true           # Disable KOTS linting
  tools:
    helm: "3.14.4"             # Specific version
    preflight: "latest"        # Use latest version
    support-bundle: "0.123.9"
```

### Linter Configuration

Each linter under `repl-lint.linters` supports:
- `disabled` (boolean): Set to `true` to disable the linter, `false` or omit to enable

**Available linters:**
- `helm`: Validates Helm chart syntax and best practices
- `preflight`: Validates preflight specification syntax
- `support-bundle`: Validates support bundle specification syntax
- `embedded-cluster`: Validates embedded cluster configurations (disabled by default)
- `kots`: Validates KOTS manifests (disabled by default)

### Tool Versions

The `tools` map specifies which versions of linting tools to use:

- `helm`: Helm CLI version for chart validation
- `preflight`: Preflight CLI version for preflight spec validation
- `support-bundle`: Support Bundle CLI version for support bundle validation

**Version formats:**
- `"latest"`: Automatically fetch the latest stable version from GitHub
- Semantic version: Specific version (e.g., `"3.14.4"`, `"v0.123.9"`)

**Example:**
```yaml
tools:
  helm: "latest"              # Always use latest Helm
  preflight: "0.123.9"        # Pin preflight to specific version
  support-bundle: "latest"
```

## Path Resolution

### Relative Paths
All paths in the configuration file are resolved relative to the directory containing the `.replicated` file. This ensures commands work correctly regardless of where they're invoked.

### Glob Patterns
Paths support glob patterns for flexible resource discovery:
- `*`: Matches any characters except `/`
- `**`: Matches any characters including `/` (recursive)
- `?`: Matches any single character
- `[abc]`: Matches any character in brackets
- `{a,b}`: Matches any of the comma-separated patterns

**Examples:**
```yaml
charts:
  - path: "./charts/*"           # All immediate subdirectories
  - path: "./services/**/chart"  # Any chart directory under services

preflights:
  - path: "./checks/**/*.yaml"   # All YAML files recursively

manifests:
  - "./*/manifests/**"           # Manifests in any top-level directory
```

## Monorepo Support

For monorepo projects, you can have multiple `.replicated` files at different directory levels. The CLI will:
1. Find all `.replicated` files from the current directory up to the root
2. Merge them with child configurations taking precedence
3. Accumulate resources (charts, preflights, manifests) from all levels
4. Override scalar fields (appId, appSlug) with child values

**Example structure:**
```
monorepo/
├── .replicated           # Root config with shared settings
│   └── appSlug: "company-suite"
├── service-a/
│   ├── .replicated       # Service A specific config
│   │   └── charts: ["./chart"]
│   └── chart/
└── service-b/
    ├── .replicated       # Service B specific config
    │   └── charts: ["./helm"]
    └── helm/
```

When running from `monorepo/service-a/`, both configs are merged:
- `appSlug` from root is used (unless overridden in child)
- Charts from both configs are included
- Lint settings are merged with child taking precedence

## Auto-Discovery Mode

If no `.replicated` file is found, the CLI operates in auto-discovery mode:
- Automatically searches for Helm charts in the current directory
- Auto-detects preflight specs (files with `kind: Preflight`)
- Auto-detects support bundle specs (files with `kind: SupportBundle`)
- Uses default linting configuration

This allows quick testing without configuration, but creating a `.replicated` file is recommended for consistent builds.

## Examples

### Simple Single-Chart Project

```yaml
appSlug: "my-application"

charts:
  - path: "./chart"

manifests:
  - "./manifests/**/*.yaml"

repl-lint:
  version: 1
  linters:
    helm:
      disabled: false
  tools:
    helm: "latest"
```

### Multi-Chart with Preflights

```yaml
appSlug: "complex-app"
promoteToChannelNames: ["beta"]

charts:
  - path: "./charts/frontend"
  - path: "./charts/backend"
  - path: "./charts/database"

preflights:
  - path: "./preflights/infrastructure.yaml"
    valuesPath: "./charts/backend"
  - path: "./preflights/networking.yaml"
    valuesPath: "./charts/frontend"

manifests:
  - "./support-bundles/**"

repl-lint:
  version: 1
  linters:
    helm:
      disabled: false
    preflight:
      disabled: false
    support-bundle:
      disabled: false
  tools:
    helm: "3.14.4"
    preflight: "latest"
    support-bundle: "latest"
```

### Monorepo Service Configuration

```yaml
# Parent .replicated at monorepo root
appSlug: "enterprise-platform"
promoteToChannelNames: ["stable"]

repl-lint:
  version: 1
  linters:
    helm:
      disabled: false
    preflight:
      disabled: false
  tools:
    helm: "latest"
    preflight: "latest"
```

```yaml
# Child .replicated in services/auth/
charts:
  - path: "./chart"

preflights:
  - path: "./preflights/*.yaml"
    valuesPath: "./chart"
```

### Minimal Configuration

```yaml
# Minimal config - relies on auto-detection
appSlug: "simple-app"
```

## Usage with CLI Commands

### Linting
```bash
# Lint all resources defined in .replicated
replicated release lint

# With verbose output (shows discovered images)
replicated release lint --verbose

# JSON output
replicated release lint --output json
```

### Creating Releases
```bash
# Create release with resources from .replicated
replicated release create --auto

# Create and automatically promote to channels from config
replicated release create --auto --promote
```

### Initialization
```bash
# Interactive setup
replicated config init
```

## Best Practices

1. **Version Control**: Always commit `.replicated` to version control
2. **Use Glob Patterns**: Leverage globs for flexible resource discovery
3. **Pin Tool Versions**: Use specific versions in CI/CD for reproducible builds
4. **Document Custom Paths**: Add comments for non-standard path structures
5. **Start Simple**: Begin with minimal config and expand as needed
6. **Test Locally**: Run `replicated release lint` before committing
7. **Monorepo Organization**: Use parent configs for shared settings, child configs for service-specific resources

## Troubleshooting

### Config Not Found
If the CLI can't find your `.replicated` file:
- Ensure it's named `.replicated` or `.replicated.yaml`
- Check that it's in the current directory or a parent directory
- Verify file permissions (should be readable)

### Invalid Glob Patterns
If you see glob pattern errors:
- Ensure patterns use valid glob syntax
- Test patterns with `ls` or `find` commands first
- Quote patterns in shell commands to prevent shell expansion

### Tool Version Errors
If tool version resolution fails:
- Use `"latest"` for automatic version detection
- Specify semantic versions without `v` prefix (e.g., `"3.14.4"`)
- Check internet connectivity for latest version fetching

### Merge Conflicts in Monorepos
If resource merging isn't working as expected:
- Check file paths are relative to each config file
- Verify both configs are valid YAML
- Use `--debug` flag to see merge details

## See Also

- [CLI Linting Documentation](./lint-format.md)
- [CLI Image Extraction](./cli-image-extraction.md)
- [CLI Profiles](./cli-profiles.md)
