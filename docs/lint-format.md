# .replicated `lint` Field (Minimal Spec)
This defines only the minimal structure for the new linter. YAML and JSON are both supported; YAML shown here.
## Format
```yaml
lint:
    version: 1                      # lint config schema version
    enabled: true                    # turn linting on/off
  stages:
    helm:
        lint:
            enabled: true           # run helm lint
            strict: false           # if true, treat warnings as errors
    preflight:
        lint:
            enabled: true
            strict: true
      tools:                            # tool resolution (optional)
        toolchainDir: ""              # set for air gap, e.g. /tools
```
Notes:
- Only keys listed above are recognized in this minimal spec. Unknown keys are rejected.
- Omit optional sections to use defaults.
- `version` controls config parsing behavior; defaults to 1 if omitted.
## Examples
1) Pin Helm version (strict mode):
```yaml
lint:
    version: 1
    enabled: true
    stages:
        helm:
        lint:
            enabled: true
            strict: true
            - "deprecated-api.*"
    tools:
        versions:
        helm: "3.14.4"
```
2) Air gap with pinned Helm and Postgres client (psql) versions:
```yaml
lint:
    version: 1
    enabled: true
    stages:
        helm:
        lint:
            enabled: true
    tools:
        toolchainDir: "/tools"   # use pre-fetched bundle only
        versions:
        helm: "3.14.4"
        psql: "15.6"            # example: pin postgres client if used by checks
```

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
