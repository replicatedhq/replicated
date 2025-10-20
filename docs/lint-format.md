# .replicated `repl-lint` Configuration

This document defines the structure for the `repl-lint` section of the `.replicated` configuration file.

## Format

Both YAML and JSON are supported; YAML is shown here for readability.

```yaml
repl-lint:
  version: 1                           # lint config schema version (currently only 1 is supported)
  linters:
    helm:
      disabled: false                  # run helm lint (default: false = enabled)
    preflight:
      disabled: false                  # run preflight lint (default: false = enabled)
    support-bundle:
      disabled: false                  # run support-bundle lint (default: false = enabled)
    embedded-cluster:                  # embedded cluster and kots linters do not exist yet
      disabled: true
    kots:
      disabled: true
  tools:                               # optional tool version pinning
    helm: "3.14.4"
    preflight: "0.123.9"
    support-bundle: "0.123.9"
```

## Field Reference

### `version`
- **Type:** integer
- **Required:** No (defaults to 1)
- **Description:** Config schema version. Controls config parsing behavior. Currently only version 1 is supported.

### `linters`
- **Type:** object
- **Required:** No
- **Description:** Configuration for each linter type.

#### Linter Configuration
Each linter supports the following field:

- **`disabled`** (boolean, optional): Set to `true` to skip this linter. Defaults to `false` (linter is enabled).

**Available linters:**
- `helm`: Validates Helm charts using `helm lint`
- `preflight`: Validates Preflight specs using `preflight lint` from troubleshoot.sh
- `support-bundle`: Validates Support Bundle specs using `support-bundle lint` from troubleshoot.sh
- `embedded-cluster`: Not yet implemented
- `kots`: Not yet implemented

### `tools`
- **Type:** object (map of tool name to version string)
- **Required:** No
- **Description:** Pin specific versions of linting tools. Tools are automatically downloaded and cached in `~/.replicated/tools/`. If not specified, default versions are used.

**Supported tools:**
- `helm`: Helm CLI version for chart linting
- `preflight`: Preflight CLI version from troubleshoot.sh
- `support-bundle`: Support Bundle CLI version from troubleshoot.sh

## Complete Example

```yaml
spec-version: 1

# Application metadata
appId: "my-app-id"
appSlug: "my-app"

# Helm charts to lint
charts:
  - path: "./charts/my-chart"
    chartVersion: "1.0.0"
    appVersion: "1.0.0"
  - path: "./charts/*"  # Glob patterns supported

# Preflight specs to lint
preflights:
  - path: "./preflight.yaml"
  - path: "./preflights/*.yaml"  # Glob patterns supported

# Manifests for support bundle discovery
# Support bundles are auto-discovered from these globs
# Any YAML file with "kind: SupportBundle" will be linted
manifests:
  - "./manifests/**/*.yaml"
  - "./k8s/*.yaml"

# Release configuration
releaseLabel: "v1.0.0"
promoteToChannelIds: []
promoteToChannelNames: ["Unstable"]

# Linter configuration
repl-lint:
  version: 1
  linters:
    helm:
      disabled: false  # Enable helm linting
    preflight:
      disabled: false  # Enable preflight linting
    support-bundle:
      disabled: false  # Enable support bundle linting
    embedded-cluster:
      disabled: true   # Not yet implemented
    kots:
      disabled: true   # Not yet implemented
  tools:
    helm: "3.14.4"
    preflight: "0.123.9"
    support-bundle: "0.123.9"
```

## Usage

Run linting for all configured resources:
```bash
replicated lint
```

The command will:
1. Lint all Helm charts specified in the `charts` array (if helm linter is enabled)
2. Lint all Preflight specs specified in the `preflights` array (if preflight linter is enabled)
3. Auto-discover and lint Support Bundle specs from `manifests` globs (if support-bundle linter is enabled)

**Exit codes:**
- `0` - All linting passed
- Non-zero - Linting failed (errors found)

## Support Bundle Auto-Discovery

Unlike Helm charts and Preflight specs which are explicitly listed, Support Bundle specs are **automatically discovered** from the `manifests` glob patterns.

The lint command will:
1. Expand all glob patterns in the `manifests` array
2. Read each YAML file
3. Check if any document contains `kind: SupportBundle`
4. Lint all discovered Support Bundle specs

This allows Support Bundles to be co-located with other Kubernetes manifests without requiring explicit configuration.

**Example manifest structure:**
```
manifests/
├── deployment.yaml      # Regular K8s manifest (ignored by support-bundle linter)
├── service.yaml         # Regular K8s manifest (ignored by support-bundle linter)
└── support-bundle.yaml  # Contains "kind: SupportBundle" (will be linted)
```

## Notes

- Omit optional sections to use defaults
- Unknown keys are rejected
- Glob patterns are supported for `charts`, `preflights`, and `manifests`
- Linters default to enabled if not explicitly disabled
- Tool versions default to latest tested versions if not specified

## See Also

- [.replicated.example](../.replicated.example) - Complete example configuration
- [README.md](../README.md) - General CLI documentation
