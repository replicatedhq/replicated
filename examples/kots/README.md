# KOTS Config Examples

This directory contains example KOTS Config files for use with the `replicated lint` command.

## Files

### valid-config.yaml

A complete, valid KOTS Config demonstrating:
- Multiple configuration groups (database, application)
- Various field types (text, password, bool, select_one)
- Required and optional fields
- Default values
- Help text and descriptions

**Usage:**
```bash
cd examples/kots
replicated lint
```

This will auto-discover and validate the config file.

### invalid-config.yaml

An intentionally invalid configuration for testing error handling. Contains:
- Invalid field type (`invalid_type_here`)
- Malformed template syntax in `when` expression
- `select_one` field missing required `items`
- Security issue: password field with default value

**Usage:**
```bash
# To see validation errors
replicated lint
```

## Testing with Mixed Configs

You can test KOTS Config alongside Embedded Cluster Config in the same file:

```yaml
---
apiVersion: kots.io/v1beta1
kind: Config
metadata:
  name: kots-config
spec:
  groups:
    - name: app
      items:
        - name: hostname
          type: text
---
apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
metadata:
  name: ec-config
spec:
  version: "1.33.0"
```

The linter uses GVK (Group-Version-Kind) filtering to correctly distinguish between KOTS Config (`kots.io`) and Embedded Cluster Config (`embeddedcluster.replicated.com`), even though both use `kind: Config`.

## Configuration with .replicated.yaml

For explicit configuration, create a `.replicated.yaml` file:

```yaml
appId: ""
appSlug: "my-app"
manifests:
  - "./valid-config.yaml"

replLint:
  linters:
    kots:
      enabled: true
      strict: false
  tools:
    kots: "latest"
```

### Strict Mode

When `strict: true`, warnings are treated as errors:
```yaml
replLint:
  linters:
    kots:
      enabled: true
      strict: true
```

## Important: Single Config Limit

Only **0 or 1** KOTS Config is allowed per project. If multiple KOTS Configs are discovered, all will fail with an error message:

```
Multiple KOTS configs found (2). Only 0 or 1 config per project is supported.
Remove duplicate configs or specify a single config file.
```

This applies to separate files, not to multi-document YAML files containing different resource types.

## Learn More

For complete documentation on KOTS Config:
- [Lint Format Documentation](../../docs/lint-format.md#kots-configuration)
- [KOTS Config Documentation](https://docs.replicated.com/reference/custom-resource-config)
