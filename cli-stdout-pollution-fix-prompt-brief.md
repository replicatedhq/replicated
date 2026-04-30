# Task: fix stdout pollution in `replicated` CLI (Shortcut #136974)

Repo: `/Users/jpwhite/code/replicated` (Go, cobra). Ship all five fixes in **one PR** on branch `justice/sc-136974/cli-stdout-warnings-corrupt-redirected-output`. Multiple commits on the branch are fine. Findings below are already verified — implement, don't re-research.

## Bug

Commands that emit structured output (kubeconfig YAML, `-o json`) write warning/info text to stdout, corrupting redirected output. Cleanest reproduction: `replicated cluster kubeconfig --name X --stdout > kc.yaml` puts a cobra deprecation warning as the first line of the file, which then breaks `kubectl`. Root cause is `cli/cmd/root.go:120` calling `SetOut(stdout)`, plus several direct `fmt.Print*` info/error calls scattered across the CLI.

## Fixes

**1. `cli/cmd/root.go:120`** — change `runCmds.rootCmd.SetOut(stdout)` to use `stderr`. Aligns with cobra's default. Behavior change: `--help` output moves to stderr. If you want to keep `--help` on stdout, add a custom `SetHelpFunc` that writes to stdout while leaving usage-on-error on stderr. Either approach is acceptable; document choice in the commit. Safe because nothing in the codebase calls `cmd.Print*` directly and command output flows through `r.w`, not cobra's writer.

**2. `cli/cmd/cluster_kubeconfig.go`** — three info/error prints leak to stdout. Move them to stderr using `fmt.Fprintf(os.Stderr, ...)` (matches existing convention in `cli/cmd/`):
- Line 126 — "kubeconfig written to %s" message
- Line 183 — "failed to remove backup kubeconfig" defer
- Line 218 — " ✓  Updated kubernetes context" success message

**Leave line 111 alone** — `fmt.Println(string(kubeconfig))` is the legitimate `--stdout` data path. Bonus: line 111 should be `fmt.Print` (not `Println`) since the kubeconfig already ends in `\n`.

Do not add a `stderr` field to the `runners` struct. Use `os.Stderr` directly.

**3. `cli/cmd/lint.go`** — auto-discovery branch (gated at line 149: `if autoDiscoveryMode`) writes status text to `r.w` regardless of output format, breaking `-o json`. Existing tests miss it because their fixtures all have a populated `.replicated` config. Gate every `fmt.Fprintf(r.w, ...)` inside that block on `r.outputFormat == "table"` — specifically lines 150, 193, 194, 195, 196, 197, and 203. The auto-discovery logic itself must still run; only the prints get gated. Match the gating pattern already used elsewhere in the file (e.g., lines 250, 264, 271).

**4. `pkg/logger/logger.go`** — TTY detection at lines 110, 143, 181, 210, 229 hardcodes `isatty.IsTerminal(os.Stdout.Fd())` regardless of the writer the logger was given. Replace with a helper that inspects `l.w` (type-asserts to `*os.File`; returns false otherwise — `tabwriter.Writer` will correctly be treated as non-TTY). Also change the two `release_download.go` callers passing `os.Stdout` to pass `os.Stderr` (lines 108 and 283). Don't touch `pkg/kots/release/save_test.go:32`.

**5. Spinner gating in -o json commands** — apply the existing pattern from `cli/cmd/app_hostname_ls.go:80-81` (`showSpinners := outputFormat == "table"`, gate each spinner call). Sites to fix:
- `cli/cmd/app_rm.go:59, 89` — `ActionWithSpinner`, not gated
- `cli/cmd/release_create.go:432, 452` — `ChildActionWithoutSpinner`, not gated. Note the spinners on lines 292/362/387/410/436 are mitigated by `log.Silence()` at line 235, but verify Silence covers `ChildActionWithoutSpinner` — if not, gate those too.

Do **not** modify `installer_create.go` or `cluster_prepare.go`; they don't support `-o json` today.

## Guardrails

- Don't try to fix the broader pattern of ~169 ungated `fmt.Fprintf(r.w, ...)` calls across the CLI. Stay scoped to the five fixes above.
- Don't refactor adjacent code, add comments explaining the fix, or add backwards-compat shims.
- Run `go build ./...`, `go vet ./...`, and `go test ./cli/cmd/... ./pkg/logger/...` before each commit.
- Don't push the branch. Stop when done and report the branch name.

## Verification

After implementing, run:

```bash
go build -o /tmp/replicated ./cli
# Fix #1 + #2: kubeconfig must be clean YAML
/tmp/replicated cluster kubeconfig --name "<some cluster>" --stdout > /tmp/kc.yaml
head -1 /tmp/kc.yaml   # must be: apiVersion: v1

# Fix #3: lint JSON must parse in an empty directory
mkdir -p /tmp/lint-empty && cd /tmp/lint-empty
/tmp/replicated release lint -o json > out.json 2>err.log
jq -e . out.json       # must succeed

# Fix #5: app rm JSON must parse
/tmp/replicated app rm <slug> -o json -f > out.json 2>err.log
jq -e . out.json       # must succeed
```

A more detailed companion doc with full code snippets and rationale lives at `cli-stdout-pollution-fix-prompt.md` in this repo. This brief is the source of truth for *what* to change; the long doc has *why*.
