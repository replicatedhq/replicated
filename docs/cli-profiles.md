# CLI Authentication Profiles

The Replicated CLI supports multiple authentication profiles, allowing you to manage and switch between different API credentials easily. This is useful when working with multiple Replicated accounts or environments.

## Overview

Authentication profiles store your API token and optionally custom API endpoints. Profiles are stored securely in `~/.replicated/config.yaml` with file permissions 600 (owner read/write only).

### Authentication Priority

The CLI determines which credentials to use in the following order:

1. `REPLICATED_API_TOKEN` environment variable (highest priority)
2. `--profile` flag (per-command override)
3. Default profile from `~/.replicated/config.yaml`
4. Legacy single token (backward compatibility)

## Commands

### `replicated profile add [profile-name]`

Add a new authentication profile with the specified name.

**Flags:**
- `--token` - API token for this profile (optional, will prompt if not provided). Supports environment variable expansion using `$VAR` or `${VAR}` syntax.

**Examples:**

```bash
# Add a production profile (will prompt for token)
replicated profile add prod

# Add a production profile with token flag
replicated profile add prod --token=your-prod-token

# Add a profile using an existing environment variable
replicated profile add prod --token='$REPLICATED_API_TOKEN'
```

If a profile with the same name already exists, it will be updated. If you add the first profile or if no default profile is set, the newly added profile will automatically become the default.

**Note:** When using environment variables, make sure to quote the value (e.g., `'$REPLICATED_API_TOKEN'`) to prevent shell expansion and allow the CLI to expand it instead.

### `replicated profile ls`

List all authentication profiles.

**Examples:**

```bash
replicated profile ls
```

**Output:**

```
  DEFAULT   NAME      API ORIGIN   REGISTRY ORIGIN
  *         prod      <default>    <default>
            dev       <default>    <default>
```

The asterisk (*) in the DEFAULT column indicates which profile is currently set as default.

### `replicated profile use [profile-name]`

Set the default authentication profile.

**Examples:**

```bash
replicated profile use prod
```

This command sets the specified profile as the default for all CLI operations. You can override the default on a per-command basis using the `--profile` flag.

### `replicated profile edit [profile-name]`

Edit an existing authentication profile. Only the flags you provide will be updated; other fields will remain unchanged.

**Flags:**
- `--token` - New API token for this profile (optional). Supports environment variable expansion using `$VAR` or `${VAR}` syntax.

**Examples:**

```bash
# Update the token for a profile
replicated profile edit dev --token=new-dev-token

# Update a profile using an environment variable
replicated profile edit dev --token='$REPLICATED_API_TOKEN'
```

### `replicated profile rm [profile-name]`

Remove an authentication profile.

**Examples:**

```bash
replicated profile rm dev
```

If you remove the default profile and other profiles exist, one of the remaining profiles will be automatically set as the default.

### `replicated profile set-default [profile-name]`

Set the default authentication profile. This is an alias for `replicated profile use`.

**Examples:**

```bash
replicated profile set-default prod
```

## Usage Examples

### Basic Workflow

```bash
# Add a production profile using an existing environment variable
replicated profile add prod --token='$REPLICATED_API_TOKEN'

# Add a development profile with a direct token
replicated profile add dev --token=your-dev-token

# List all profiles
replicated profile ls

# Switch to production profile
replicated profile use prod

# Use development profile for a single command
replicated app ls --profile=dev

# Edit a profile's token
replicated profile edit dev --token=new-dev-token

# Remove a profile
replicated profile rm dev
```

### Working with Multiple Accounts

```bash
# Add profiles for different accounts
replicated profile add company-a --token=token-a
replicated profile add company-b --token=token-b

# Switch between accounts
replicated profile use company-a
replicated app ls  # Lists apps for company-a

replicated profile use company-b
replicated app ls  # Lists apps for company-b
```

## Security

- All credentials are stored in `~/.replicated/config.yaml` with file permissions 600 (owner read/write only)
- Tokens are masked when prompted interactively
- Environment variables take precedence, allowing temporary overrides without modifying stored profiles

## Troubleshooting

### Profile Not Found

If you see "profile not found", use `replicated profile ls` to list available profiles and verify the profile name.

### Permission Denied

If you encounter permission issues with `~/.replicated/config.yaml`, verify the file has the correct permissions:

```bash
chmod 600 ~/.replicated/config.yaml
```

### Multiple Profiles, Wrong One Being Used

Check the default profile with `replicated profile ls` and ensure the correct profile is marked with an asterisk (*). You can change the default with `replicated profile use [profile-name]`.
