# Plan: Migrate replicatedhq/replicated to Depot CI

## Goal
Move completely off GitHub Actions and Dagger. All CI runs on Depot. All release automation runs on Depot.

## Current State

### `.github/workflows/main.yaml` — 6 jobs
| Job | What it does |
|-----|-------------|
| `make-unit-tests` | `go test` (excludes `/pact`, `/pkg/integration`) |
| `make-pact-tests` | `go test ./pact/...` + pact-broker publish + can-i-deploy (main only) |
| `make-integration-tests` | `go test ./pkg/integration/...` |
| `make-build` | `go build` the CLI binary |
| `dagger-build` | `make dagger-build` — builds binary via Dagger container |

### Dagger (`dagger/` + `dagger.json`)
| Function | Purpose |
|----------|---------|
| `Build` | Compile binary in container |
| `Release` | Full release: version bump → git tag → Docker image publish → goreleaser → docs PR |
| `GoreleaserDryrun` | Snapshot build verification (used in CI) |
| `GenerateDocs` | Generate CLI docs, open PR against `replicatedhq/replicated-docs` |
| `Validate` | Semgrep security scan + unit tests (no-op compat/perf) |

### Makefile targets to keep
- `test-unit`, `test-pact`, `test-integration`, `test-lint`, `build`
- `publish-pact`, `can-i-deploy`, `unpublish-past-versions`

### Makefile targets to remove
- `dagger-build`, `dagger-goreleaser-dryrun`, `release`, `docs`

---

## Proposed Changes (Single PR)

### 1. Create `.depot/workflows/pr.yml`

**Trigger:** `pull_request` on `main` (opened, synchronize, reopened, ready_for_review)

**Jobs:**

| Job | Purpose | Time | Depot Cost |
|-----|---------|------|------------|
| `lint` | go mod tidy check, go fmt check, `make build`, `make test-lint` | ~2 min | ~$0.01 |
| `unit-tests` | `make test-unit` | ~3 min | ~$0.02 |
| `pact-tests` | `make test-pact` + pact publish + can-i-deploy (main only) | ~3 min | ~$0.02 |
| `integration-tests` | `make build` + `make test-integration` | ~3 min | ~$0.02 |
| `build` | `make build` (verifies binary compiles) | ~1 min | ~$0.01 |
| `goreleaser-dryrun` | `goreleaser release --snapshot --clean` | ~2 min | ~$0.01 |
| `gate` | Aggregates all job results, fails if any failed | <1 min | ~$0.00 |

**Outcome:** All PRs get a single required status check (`gate`). No path filters needed (single Go project). No matrix. Fast feedback — lint runs first, everything else in parallel.

**Total PR wall clock:** ~3 min (parallel), dominated by unit tests + pact tests.
**Total PR cost:** ~$0.10.

---

### 2. Create `.depot/workflows/release.yml`

**Trigger:** `push` of tags matching `v*` (e.g. `v0.130.0`)

**Jobs:**

| Job | Purpose | Time | Depot Cost |
|-----|---------|------|------------|
| `build-and-test` | `make test-unit`, `make test-pact`, `make build` | ~5 min | ~$0.03 |
| `release` | goreleaser proper release (GitHub release + binaries + Homebrew) | ~3 min | ~$0.02 |
| `docker-publish` | Build `replicated/vendor-cli` image, tag latest/major/minor/patch, push to Docker Hub | ~3 min | ~$0.02 |
| `docs` | Generate CLI docs, clone `replicated-docs`, update sidebars, commit to branch, open PR | ~2 min | ~$0.01 |

**Outcome:** Tag push → binaries on GitHub, Docker images on Hub, docs PR opened automatically. No manual `make release`.

**Total release wall clock:** ~5 min (build-and-test blocks release/docker/docs, which can run in parallel after).
**Total release cost:** ~$0.08.

---

### 3. Delete Dagger
- Remove `dagger/` directory entirely
- Remove `dagger.json`
- Remove from `.github/dependabot.yml` if present

---

### 4. Delete `.github/workflows/main.yaml`
- Remove entirely (or disable triggers as transition)

---

### 5. Update Makefile
- Remove: `dagger-build`, `dagger-goreleaser-dryrun`, `release`, `docs`
- No new make target needed — `make build` already does what `dagger-goreleaser-dryrun` verified (binary compiles). The `goreleaser-dryrun` job in PR CI runs `goreleaser release --snapshot --clean` directly, not via make.

---

## Open Questions

1. ~~Release trigger~~: **Resolved** — push a tag to release. No `workflow_dispatch`, no manual `make release`.

2. ~~Docker image publishing~~: **Resolved** — separate `docker-publish` job with manual `docker build`/`docker push`. Simpler and faster than goreleaser Docker support. Build from `Dockerfile`, tag latest/major/minor/patch, push to Docker Hub.

3. ~~Docs generation~~: **Resolved** — port to `scripts/generate-docs.sh`. The release workflow calls this script after goreleaser succeeds. Script handles: clone `replicated-docs`, clean old CLI docs, copy generated, update `sidebars.js`, commit to branch, push, open PR via GitHub API.

4. ~~Pact can-i-deploy~~: **Resolved** — keep in PR validation. Run on every PR (not just main).

5. ~~Semgrep~~: **Resolved** — keep in PR CI. Add a `security` job that runs `semgrep scan --config=p/golang .` directly (not via Dagger).

6. ~~Secrets for Depot~~: **Resolved** — secrets ported from GitHub Actions to Depot with same names and values. No 1Password integration needed.

---

## Why This Fits in One PR

- No path filters needed (single Go project)
- No matrix of integration suites (3 test targets only)
- No frontend, no migrations, no multi-service builds
- Dagger removal is mechanical (delete files, delete Makefile targets)
- Release workflow is straightforward: test → goreleaser → Docker push → docs
