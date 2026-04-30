# Task: fix stdout pollution in `replicated` CLI (Shortcut #136974)

You are working in the `replicated` CLI repo at `/Users/jpwhite/code/replicated` (Go, cobra-based). Implement the five fixes below. Each section names exact file:line references and the change to make. All findings have already been verified by independent research — do not re-verify; just implement.

## Background

**Bug**: Commands that emit structured output (kubeconfig YAML, `-o json`) write warning/info text to stdout, corrupting redirected output. Friction-log report:

> `replicated cluster kubeconfig > kubeconfig.yaml` includes warning lines inline in stdout. The resulting file contains both the kubeconfig YAML and warning text, breaking subsequent `kubectl` calls.

**Reproduction (cleanest case)**:
```bash
replicated cluster kubeconfig --name "some cluster" --stdout > kubeconfig.yaml
head -1 kubeconfig.yaml
# Output: "Flag --name has been deprecated, use ID_OR_NAME arguments instead"
```

The deprecation warning is emitted by cobra to stdout because `cli/cmd/root.go:120` explicitly calls `SetOut(stdout)`. There are also several direct `fmt.Print*` info/error calls in `cluster_kubeconfig.go` and a similar bug in `lint.go` for `-o json` output.

## PR structure

All five fixes ship in **one PR** that closes Shortcut #136974. Use a single branch named `justice/sc-136974/cli-stdout-warnings-corrupt-redirected-output`. You may make multiple commits on that branch (one per fix is fine, or group related ones), but the deliverable is one PR.

---

## Fix #1 — Cobra writer should default to stderr

**File**: `cli/cmd/root.go:120`

**Current**:
```go
if stdout != nil {
    runCmds.rootCmd.SetOut(stdout)
}
```

**Change**: Replace `stdout` with `stderr` so cobra's auto-emitted text (deprecation warnings, usage-on-error, any `c.Print*` calls) goes to stderr. This is cobra's default — we are aligning with it.

```go
if stderr != nil {
    runCmds.rootCmd.SetOut(stderr)
}
```

**Why this is safe**:
- The codebase has zero direct calls to `cmd.Print`, `cmd.Println`, `cmd.Printf`. Confirmed via grep.
- All command output is written through `r.w` (a tabwriter wrapping stdout) — independent of cobra's writer.
- `runCmds.rootCmd.SetErr(stderr)` is already called on line 117, so error messages are unaffected.
- No tests assert against cobra's writer for usage/help/deprecation text.

**Behavior change to be aware of**: `replicated --help` will now write help text to stderr instead of stdout. This matches cobra's default. Many users redirect `--help` (e.g. `replicated --help | grep something`) — that will still work because help is shown on the terminal regardless of stream — but a script doing `replicated --help > help.txt` will produce an empty file. If you want to preserve `--help` on stdout while moving usage-on-error to stderr, also add a custom help function:

```go
runCmds.rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
    cmd.SetOut(stdout)
    // call cobra's default help
    ...
})
```

Decide which approach you prefer. The minimal one-line change is acceptable; the help-preserving approach is more polite. Document the choice in the commit message.

## Fix #2 — `cluster_kubeconfig.go` info/error prints

**File**: `cli/cmd/cluster_kubeconfig.go`

Three `fmt.Printf`/`fmt.Println` calls leak diagnostic text to stdout. Move them to stderr.

| Line | Current | Change to |
|------|---------|-----------|
| 126 | `fmt.Printf("kubeconfig written to %s\n", r.args.kubeconfigPath)` | `fmt.Fprintf(os.Stderr, "kubeconfig written to %s\n", r.args.kubeconfigPath)` |
| 183 | `fmt.Printf("failed to remove backup kubeconfig: %s\n", err.Error())` | `fmt.Fprintf(os.Stderr, "failed to remove backup kubeconfig: %s\n", err.Error())` |
| 218 | `fmt.Printf(" ✓  Updated kubernetes context '%s' in '%s'\n", mergedConfig.CurrentContext, kubeconfigPaths[0])` | `fmt.Fprintf(os.Stderr, " ✓  Updated kubernetes context '%s' in '%s'\n", mergedConfig.CurrentContext, kubeconfigPaths[0])` |

Use `fmt.Fprintf(os.Stderr, ...)` directly — that matches the existing convention in `cli/cmd/` (see `root.go:380, 382-384, 395, 400, 528-530`, `enterprise_portal_preview.go:156, 185, 193, 196, 199, 201`). Do **not** add a stderr field to the `runners` struct.

**Leave line 111 as-is** — `fmt.Println(string(kubeconfig))` is the `--stdout` data path and correctly belongs on stdout.

**Bonus fix on line 111**: change `fmt.Println(string(kubeconfig))` to `fmt.Print(string(kubeconfig))` (or `os.Stdout.Write(kubeconfig)`). The kubeconfig payload already ends in `\n`; `Println` adds a second newline. Minor — kubectl tolerates it — but worth fixing while you're in the file.

**Test impact**: There are no tests for `cluster kubeconfig` in this repo. Verify manually:

```bash
# After the fix, this file should contain ONLY YAML, no leading deprecation warning:
go build -o /tmp/replicated ./cli
/tmp/replicated cluster kubeconfig --name "<some cluster>" --stdout > /tmp/kc.yaml
head -1 /tmp/kc.yaml  # should be: apiVersion: v1

# This file should be a valid kubeconfig (currently it contains the success message):
KUBECONFIG=/tmp/test-kc /tmp/replicated cluster kubeconfig <ID> 2>/tmp/stderr
cat /tmp/stderr  # success message lives here now
```

---

## Fix #3 — `lint.go` corrupts `-o json` output in auto-discovery mode

**File**: `cli/cmd/lint.go`

When `replicated release lint -o json` runs in a directory without a fully-populated `.replicated` config, the auto-discovery branch (gated at line 149: `if autoDiscoveryMode { ... }`) writes informational text to `r.w` (stdout) regardless of output format, breaking JSON parsers.

**Why existing tests miss it**: `lint_test.go` fixtures all have a populated `.replicated` config, so `autoDiscoveryMode` is false in tests.

**Offending lines** (all inside the `if autoDiscoveryMode` block at line 149):

- 150: `fmt.Fprintf(r.w, "No .replicated config found. Auto-discovering lintable resources in current directory...\n\n")`
- 193: `fmt.Fprintf(r.w, "Discovered resources:\n")`
- 194-197: the four `fmt.Fprintf(r.w, "  - %d ...")` lines
- 203: `fmt.Fprintf(r.w, "No lintable resources found in current directory.\n")`

**Fix**: Wrap each of those `fmt.Fprintf(r.w, ...)` calls with `if r.outputFormat == "table" { ... }`. Or, more cleanly, define `showAutoDiscoveryMessages := r.outputFormat == "table"` once at the top of the auto-discovery block and gate each print on it. Other prints in the file (e.g., lines 250, 264, 271, 287, 299, 308, 320, 334, 373, 898) already use the `outputFormat == "table"` gate — match that pattern.

The auto-discovery **logic itself must still run** — it populates `config` for the rest of the function. Only the prints get gated.

**Verify after fix**:

```bash
cd /tmp && mkdir lint-test && cd lint-test
go run /Users/jpwhite/code/replicated/cli release lint -o json > out.json 2>err.log
jq -e . out.json   # must succeed (clean JSON)
cat err.log         # nothing on stderr is required either, but spurious is fine
```

---

## Fix #4 — `logger.go` TTY check is hardcoded to `os.Stdout.Fd()`

**File**: `pkg/logger/logger.go`

The logger struct holds `w io.Writer`, but its TTY checks at lines **110, 143, 181, 210, 229** all examine `os.Stdout.Fd()` regardless of which writer the caller passed in. Fix the check so it inspects the configured writer.

**Pattern**: replace each `isatty.IsTerminal(os.Stdout.Fd())` with a helper that checks `l.w`:

```go
func (l *Logger) isTTY() bool {
    if f, ok := l.w.(*os.File); ok {
        return isatty.IsTerminal(f.Fd())
    }
    return false
}
```

Then replace the five call sites. A `tabwriter.Writer` will fail the type assertion and return false (correct — non-TTY behavior). `os.Stdout` and `os.Stderr` will succeed.

**Also fix the two `release_download.go` callers** that pass `os.Stdout` directly:
- `cli/cmd/release_download.go:108` — change `logger.NewLogger(os.Stdout)` to `logger.NewLogger(os.Stderr)`
- `cli/cmd/release_download.go:283` — same change

`release download` produces no stdout data (it writes to disk), so progress messages belong on stderr by convention. Don't touch the test file `pkg/kots/release/save_test.go:32`.

**Test impact**: no existing tests assert on logger output destinations. Verify manually:

```bash
go run ./cli release download <args> 2>err.log >out.log
# out.log should be empty, err.log contains the action messages
```

---

## Fix #5 — Audit logger spinner call sites in commands with `-o json`

**File**: multiple under `cli/cmd/`

Even with fix #4, the initial `"  • <msg>"` text from `ActionWithSpinner` and `ActionWithoutSpinner` still prints to whatever writer the logger has. For commands that share their writer with structured output (`r.w` = stdout), this leaks into `-o json` payloads.

**The architecturally clean fix** is to make the default logger writer in this codebase be stderr. But that's a larger refactor than the ticket calls for. Instead, apply per-call-site gating using the existing pattern at `cli/cmd/app_hostname_ls.go:80-81`:

```go
showSpinners := outputFormat == "table"
log := logger.NewLogger(r.w)
// ...
if showSpinners {
    log.ActionWithSpinner("Fetching app")
}
```

**Sites to fix** (commands that support `-o json` and have non-gated logger calls):

- `cli/cmd/app_rm.go:59` — `ActionWithSpinner` not gated
- `cli/cmd/app_rm.go:89` — `ActionWithSpinner` not gated
- `cli/cmd/release_create.go:432` — `ChildActionWithoutSpinner` not gated (note: spinners on lines 292, 362, 387, 410, 436 are partially mitigated by `log.Silence()` at line 235, but the `ChildActionWithoutSpinner` calls aren't covered by Silence — verify this and gate them)
- `cli/cmd/release_create.go:452` — same as above

**Do not modify** `installer_create.go` and `cluster_prepare.go` — they don't currently support `-o json`, so leave them alone (this is a "fix only what's needed" instruction; do not add gates for hypothetical future structured-output support).

**Verify**:
```bash
go run ./cli app rm <slug> -o json -f 2>err.log >out.log
jq -e . out.log     # must succeed
```

---

## General implementation guidelines

- Single PR, single branch (`justice/sc-136974/cli-stdout-warnings-corrupt-redirected-output`). Multiple commits on the branch are fine — group related fixes together. Use Conventional Commits format.
- Don't add comments explaining the fix — the commit message and PR description handle that.
- Don't refactor adjacent code. Don't add backwards-compat shims. Don't add fields to the `runners` struct.
- Run `go build ./...` and `go vet ./...` before each commit.
- Run existing tests in affected packages: `go test ./cli/cmd/... ./pkg/logger/...`
- Don't push the branch unless the user asks. Stop when done and report the branch name.

## What NOT to do

- Don't try to "fix everything" — there are ~169 ungated `fmt.Fprintf(r.w, ...)` calls across the CLI for status text. Most are in commands that don't support structured output. Do not try to gate or move them. Stay focused on the five fixes above.
- Don't add a `stderr` field to the `runners` struct. Use `os.Stderr` directly.
- Don't change `cluster_kubeconfig.go:111` to write to stderr — that's the legitimate stdout data path.
- Don't change the version-update banner code (`pkg/version/version.go`) — `PrintIfUpgradeAvailable` already correctly writes to stderr.
