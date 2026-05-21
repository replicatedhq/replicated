# Implementation Plan: Global `--output` Persistent Flag

## Goal

Introduce a global `--output` (`-o`) persistent flag on the root command that all subcommands inherit, replacing ~78 local `--output` flag registrations. Support a new `REPLICATED_OUTPUT` environment variable for setting a default output format without providing the flag on every command.

## Resolution Order

1. Explicit `--output` flag on the command line
2. `REPLICATED_OUTPUT` environment variable
3. Default: `"table"`

## Files to Edit

- `cli/cmd/runner.go` — add resolution method, clean up `runnerArgs`
- `cli/cmd/root.go` — register global persistent flag, wire resolution into pre-run hooks
- `cli/cmd/version.go` — migrate from `--json` bool flag to global `--output` flag
- ~70 command files — delete local `--output` flags, refactor local `outputFormat` variables to `r.outputFormat`

---

## Phase 1: Core Infrastructure

### 1.1 Add Output Resolution Method

**File:** `cli/cmd/runner.go`

Add the following method to the `runners` struct:

```go
func (r *runners) resolveOutputFormat(cmd *cobra.Command) {
	if cmd.Flags().Changed("output") {
		return // explicit flag wins
	}
	if env := os.Getenv("REPLICATED_OUTPUT"); env != "" {
		r.outputFormat = env
	}
}
```

Also ensure `"os"` is imported if not already present.

### 1.2 Register Global Persistent Flag

**File:** `cli/cmd/root.go`

Inside `Execute()`, after creating `runCmds`, add the persistent flag:

```go
runCmds.rootCmd.PersistentFlags().StringVarP(
	&runCmds.outputFormat, "output", "o", "table",
	"The output format to use. Supported formats vary by command (json, table, wide). (default 'table', override with REPLICATED_OUTPUT env var)",
)
```

### 1.3 Wire Resolution into Pre-Run Hooks

**File:** `cli/cmd/root.go`

Add `runCmds.resolveOutputFormat(cmd)` at the beginning of both:

- `preRunSetupAPIs` — used by app, registry, cluster, vm, network, api, model, policy, notification, config, default
- `prerunCommand` — used by channel, release, collector, installer, customer, instance, enterprise-portal, cluster-prepare

Example addition in `preRunSetupAPIs`:

```go
preRunSetupAPIs := func(cmd *cobra.Command, args []string) error {
	runCmds.resolveOutputFormat(cmd)
	// ... rest of existing logic
}
```

Same for `prerunCommand`.

---

## Phase 2: Refactor Commands Using Local `outputFormat` Variables

These commands declare a local `var outputFormat string` and pass it to helper methods. Remove the local variable, update the helper method signature to drop the `outputFormat` parameter, and use `r.outputFormat` directly inside the helper.

| File | Local Var | Helper to Update |
|---|---|---|
| `app_ls.go` | `outputFormat` | `listApps(...)` → drop `outputFormat` param |
| `app_create.go` | `outputFormat` | `createApp(...)` → drop `outputFormat` param |
| `app_rm.go` | `outputFormat` | `deleteApp(...)` → drop `outputFormat` param |
| `app_hostname_ls.go` | `outputFormat` | `listAppHostnames(...)` → drop `outputFormat` param |
| `customer_ls.go` | `outputFormat` | `listCustomers(...)` → drop `outputFormat` param |
| `customer_create.go` | `outputFormat` | `createCustomer(...)` → drop `outputFormat` param |
| `customer_inspect.go` | `outputFormat` | `inspectCustomer(...)` → drop `outputFormat` param |
| `policy_ls.go` | `outputFormat` | `policyList(...)` → drop `outputFormat` param |
| `policy_get.go` | `outputFormat` | `policyGet(...)` → drop `outputFormat` param |
| `policy_create.go` | `outputFormat` | `policyCreate(...)` → drop `outputFormat` param |
| `policy_update.go` | `outputFormat` | `policyUpdate(...)` → drop `outputFormat` param |
| `default_show.go` | `outputFormat` | `defaultShow(...)` → drop `outputFormat` param |
| `default_set.go` | `outputFormat` | `defaultSet(...)` → drop `outputFormat` param |
| `enterpriseportal_status_get.go` | `outputFormat` | `enterprisePortalStatusGet(...)` → drop `outputFormat` param |
| `enterpriseportal_status_update.go` | `outputFormat` | `enterprisePortalStatusUpdate(...)` → drop `outputFormat` param |
| `enterprise_portal_invite.go` | `outputFormat` | `enterprisePortalInvite(...)` → drop `outputFormat` param |
| `enterprise_portal_user_ls.go` | `outputFormat` | `enterprisePortalUserLs(...)` → drop `outputFormat` param |

### 2.1 Refactor `notification.go`

All 10 notification subcommands use local `var outputFormat string` inside closures that already capture `r *runners`. For each subcommand:

1. Remove `var outputFormat string` declaration
2. Remove `cmd.Flags().StringVarP(&outputFormat, "output", ...)` line
3. Inside the closure, replace `outputFormat` with `r.outputFormat`

Affected functions in `notification.go`:
- `InitNotificationSubscriptionList`
- `InitNotificationSubscriptionGet`
- `InitNotificationSubscriptionCreate`
- `InitNotificationSubscriptionUpdate`
- `InitNotificationSubscriptionEvents`
- `InitNotificationEventList`
- `InitNotificationEventTypeList`
- `InitNotificationEmailResendVerification`
- `InitNotificationEmailVerify`
- `InitNotificationWebhookTest`

### 2.2 Refactor `cluster_addon_ls.go`

1. Remove `outputFormat string` from `clusterAddonLsArgs` struct
2. Remove `cmd.Flags().StringVarP(&args.outputFormat, ...)` line from `clusterAddonLsFlags`
3. Update `addonClusterLsRun` signature: `func (r *runners) addonClusterLsRun(clusterID string) error`
4. In `InitClusterAddonLs`, capture `args.clusterID` directly instead of passing the whole struct
5. Use `r.outputFormat` inside `addonClusterLsRun` instead of `args.outputFormat`

### 2.3 Refactor `cluster_addon_create_objectstore.go`

1. Remove `clusterAddonCreateObjectStoreOutput string` from `runnerArgs` struct in `runner.go`
2. Remove `cmd.Flags().StringVarP(&r.args.clusterAddonCreateObjectStoreOutput, ...)` from `clusterAddonCreateObjectStoreFlags`
3. In `clusterAddonCreateObjectStoreCreateRun`, replace `r.args.clusterAddonCreateObjectStoreOutput` with `r.outputFormat`

---

## Phase 3: Delete Local `--output` Flags (Already Using `r.outputFormat`)

These commands already bind their `--output` flag directly to `r.outputFormat`. Simply delete the `cmd.Flags().StringVarP(&r.outputFormat, "output", ...)` line in each file.

**Files to edit:**

- `release_ls.go`
- `release_create.go`
- `release_inspect.go`
- `release_lint.go`
- `installer_ls.go`
- `channel_ls.go`
- `channel_create.go`
- `channel_inspect.go`
- `channel_releases.go`
- `instance_ls.go`
- `instance_inspect.go`
- `instance_tag.go`
- `registry_ls.go`
- `registry_add_dockerhub.go`
- `registry_add_ecr.go`
- `registry_add_gar.go`
- `registry_add_gcr.go`
- `registry_add_ghcr.go`
- `registry_add_quay.go`
- `registry_add_other.go`
- `model_ls.go`
- `model_rm.go`
- `collection_ls.go`
- `collection_create.go`
- `collection_rm.go`
- `collection_addmodel.go`
- `collection_removemodel.go`
- `cluster_ls.go`
- `cluster_create.go`
- `cluster_upgrade.go`
- `cluster_versions.go`
- `cluster_update_ttl.go`
- `cluster_update_nodegroup.go`
- `cluster_nodegroup_ls.go`
- `cluster_port_ls.go`
- `cluster_port_expose.go`
- `cluster_port_rm.go`
- `vm_ls.go`
- `vm_create.go`
- `vm_versions.go`
- `vm_update_ttl.go`
- `vm_port_ls.go`
- `vm_port_expose.go`
- `vm_port_rm.go`
- `network_ls.go`
- `network_create.go`
- `customer_update.go`

---

## Phase 4: Clean Up `runnerArgs`

**File:** `cli/cmd/runner.go`

Remove from `runnerArgs` struct:

```go
clusterAddonCreateObjectStoreOutput string
```

---

## Phase 5: Migrate `version` Command to Global `--output`

**File:** `cli/cmd/version.go`

1. Remove `var versionJson bool`
2. Remove `cmd.Flags().BoolVar(&versionJson, "json", false, "output version info in json")`
3. Change function signature from `func Version() *cobra.Command` to `func (r *runners) Version() *cobra.Command`
4. Replace `if !versionJson` with `if r.outputFormat != "json"`
5. Replace `else` branch with JSON output

**File:** `cli/cmd/root.go`

Update the call from:
```go
runCmds.rootCmd.AddCommand(Version())
```
to:
```go
runCmds.rootCmd.AddCommand(runCmds.Version())
```

**Note:** `version_upgrade.go` does not need changes — it has no output formatting logic.

---

## Phase 6: Test Update Pass

### 6.1 Existing Tests That Directly Set `runners.outputFormat`

Tests in these files directly instantiate `&runners{outputFormat: "json"}` — they continue working unchanged:

- `lint_test.go`
- `release_create_test.go`
- `vm_create_test.go`
- `cluster_ls_test.go`
- `notification_test.go`
- `release_lint_test.go`
- `release_image_ls_test.go`
- `image_extraction_test.go`
- `completion_test.go`
- `cluster_prepare_test.go`
- `cluster_create_test.go`

No changes needed for these.

### 6.2 New Tests to Add

Create a new test file (e.g., `cli/cmd/output_test.go`) with tests for `resolveOutputFormat`:

```go
func TestResolveOutputFormat_ExplicitFlagWins(t *testing.T) {
	// Set env var
	t.Setenv("REPLICATED_OUTPUT", "json")

	r := &runners{outputFormat: "table"}
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "table", "")
	cmd.Flags().Set("output", "wide")

	r.resolveOutputFormat(cmd)
	require.Equal(t, "wide", r.outputFormat)
}

func TestResolveOutputFormat_EnvVarWinsOverDefault(t *testing.T) {
	t.Setenv("REPLICATED_OUTPUT", "json")

	r := &runners{outputFormat: "table"}
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "table", "")

	r.resolveOutputFormat(cmd)
	require.Equal(t, "json", r.outputFormat)
}

func TestResolveOutputFormat_DefaultTable(t *testing.T) {
	r := &runners{outputFormat: "table"}
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "table", "")

	r.resolveOutputFormat(cmd)
	require.Equal(t, "table", r.outputFormat)
}
```

### 6.3 Compile Check

After all edits, run:

```bash
go build ./...
go test ./cli/cmd/...
```

---

## Phase 7: Verification & Exclusions

### 7.1 Commands Intentionally Excluded

These commands have an `--output` flag but it is **NOT** a format flag — do **NOT** modify them:

| File | Flag Purpose |
|---|---|
| `customer_download_license.go` | `--output` is a **file path** (e.g., `--output license.yaml`) |

These commands have no output formatting and need no changes:

| File | Reason |
|---|---|
| `login.go` | No output formatting |
| `logout.go` | No output formatting |
| `completion.go` | Shell script generation, not tabular/json data |
| `config.go` | No output formatting |

### 7.2 Per-Command Validation Stays

Different commands support different format sets (`json|table` vs `json|table|wide`). Existing validation blocks like:

```go
if r.outputFormat != "table" && r.outputFormat != "json" {
	return errors.Errorf("invalid output: %s", r.outputFormat)
}
```

continue to work unchanged because they validate `r.outputFormat`, which is now set globally.

---

## Checklist

- [ ] `runner.go`: Add `resolveOutputFormat` method, remove `clusterAddonCreateObjectStoreOutput` from `runnerArgs`
- [ ] `root.go`: Add global `--output` persistent flag, wire `resolveOutputFormat` into `preRunSetupAPIs` and `prerunCommand`
- [ ] `version.go`: Migrate from `--json` bool to global `--output`, change to `(r *runners) Version()`
- [ ] `root.go`: Update `runCmds.Version()` call
- [ ] Refactor all commands with local `outputFormat` vars (Phase 2)
- [ ] Delete local `--output` flags from ~50 commands already using `r.outputFormat` (Phase 3)
- [ ] Add `output_test.go` with resolution tests (Phase 6)
- [ ] Run `go build ./...` — zero errors
- [ ] Run `go test ./cli/cmd/...` — all pass
- [ ] Verify `REPLICATED_OUTPUT=json` works without explicit `--output json`
- [ ] Verify explicit `--output wide` still overrides env var
- [ ] Verify `--output` appears in `replicated --help` Global Flags section
- [ ] Verify `customer download-license --output` still works as a file path flag
