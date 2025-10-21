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
        embedded-cluster:                # embedded cluster and kots linters do not exist as of yet
            enabled: false
            strict: false
        kots:
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
        valuesPath: "./chart/something", # directory to corresponding helm chart
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
