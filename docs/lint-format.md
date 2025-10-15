# .replicated `lint` Field (Minimal Spec)
This defines only the minimal structure for the new linter. YAML and JSON are both supported; YAML shown here.
## Format
```yaml
repl-lint:
    version: 1                       # lint config schema version
    enabled: true                    # turn linting on/off
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
      tools:                         # tool resolution (optional)
```
Notes:
- Only keys listed above are recognized in this minimal spec. Unknown keys are rejected.
- Omit optional sections to use defaults.
- `version` controls config parsing behavior; defaults to 1 if omitted.
## Examples
1) Pin Helm version (strict mode):
```yaml
repl-lint:
    version: 1
    enabled: true
    linters:
        helm:
            enabled: true                
            strict: false                
        preflight:
            enabled: true
            strict: true
        support-bundle:
            enabled: true
            strict: false
        embedded-cluster:                
            enabled: false
            strict: false
        kots:
            enabled: false
            strict: false
    tools:
        helm: "3.14.4"
        psql: "15.6"
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
