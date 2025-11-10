# Embedded Cluster Configuration Examples

This directory contains example Embedded Cluster configuration files for use with the `replicated lint` command.

## Files

### valid-config.yaml

A complete, valid Embedded Cluster configuration demonstrating:
- Kubernetes version specification
- Controller role configuration
- Helm repository extensions
- Network CIDR configuration
- K0s overrides for custom networking

**Usage:**
```bash
cd examples/embedded-cluster
replicated lint
```

This will auto-discover and validate the config file.

### invalid-config.yaml

An intentionally invalid configuration for testing error handling. Contains:
- Missing required `version` field
- Missing required `description` field in role
- Invalid CIDR format
- Missing required `url` field in Helm repository

**Usage:**
```bash
# To see validation errors
replicated lint
```

## Configuration with .replicated file

For explicit configuration, create a `.replicated` file:

```yaml
appId: ""
appSlug: "my-app"
manifests:
  - "./valid-config.yaml"

repl-lint:
  version: 1
  linters:
    embedded-cluster:
      enabled: true
      strict: true
  tools:
    embedded-cluster: "latest"
```

## Platform Requirements

**Important:** The embedded-cluster linter binary is currently only available for **Linux AMD64**.

- ✅ Linux (x86_64/amd64): Full support
- ❌ macOS: Not available
- ❌ Windows: Not available
- ❌ ARM64: Not available

If you're on an unsupported platform, consider running the linter in:
- Docker container: `docker run --rm -v $(pwd):/app -w /app replicated/replicated lint`
- CI/CD environment (GitHub Actions, GitLab CI, etc.)
- Linux VM or WSL2

## Learn More

For complete documentation on Embedded Cluster configuration:
- [Lint Format Documentation](../../docs/lint-format.md#embedded-cluster-configuration)
- [Embedded Cluster Documentation](https://docs.replicated.com/vendor/embedded-overview)
