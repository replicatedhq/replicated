# Lint Integration Tests

This directory contains integration/e2e tests for the `replicated lint` command.

## Structure

Each subdirectory in `testdata/` represents a test case. A valid test case must contain:

1. **`.replicated` or `.replicated.yaml`**: Configuration file that defines what to lint (charts, preflights, support bundles)
2. **`expect.json`**: Expected lint results in JSON format

## Test Case Format

### `.replicated` Configuration

The configuration file defines the resources to lint. Example:

```yaml
appSlug: "test-case-name"
charts: [
  {
    path: "./charts/my-chart",
    chartVersion: "",
    appVersion: "",
  },
]
preflights: []
repl-lint:
  version: 1
  linters:
    helm:
      disabled: false
    preflight:
      disabled: false
    support-bundle:
      disabled: true
```

### `expect.json` Format

Defines the expected lint messages (if any):

```json
{
  "lintMessages": [
    {
      "severity": "ERROR",
      "path": "templates/deployment.yaml",
      "message": "Expected error message"
    },
    {
      "severity": "WARNING",
      "path": "values.yaml",
      "message": "Expected warning message"
    }
  ]
}
```

For tests that expect **no lint errors** (clean pass):

```json
{
  "lintMessages": []
}
```

#### Severity Levels

- `ERROR`: Critical issues that fail the lint check
- `WARNING`: Non-critical issues that don't fail the lint check
- `INFO`: Informational messages

## Running Tests

### Run all lint tests:

```bash
make test-lint
```

This will:
1. Build the `replicated` binary (if needed)
2. Find all test cases in `testdata/`
3. Run `replicated lint` in each test directory
4. Compare actual results with `expect.json`
5. Report pass/fail for each test

### Run manually:

```bash
# Build the binary first
make build

# Run the test script
./scripts/test-lint.sh
```

### Run lint in a specific test directory:

```bash
cd testdata/chart-with-required-values
../../bin/replicated lint
```

## Test Validation

The test script validates:

1. **Exit code**: 
   - Exit code 0 if no ERROR-level messages expected
   - Exit code 1 if ERROR-level messages expected

2. **Message counts**: 
   - Number of ERROR messages matches
   - Number of WARNING messages matches
   - Number of INFO messages matches

3. **Output**: Test script shows diff if validation fails

## Adding New Tests

1. Create a new directory in `testdata/`:
   ```bash
   mkdir testdata/my-new-test
   ```

2. Add your test resources (Helm chart, preflight spec, etc.)

3. Create `.replicated` config:
   ```bash
   cat > testdata/my-new-test/.replicated << EOF
   appSlug: "my-new-test"
   charts: [
     {
       path: "./my-chart",
       chartVersion: "",
       appVersion: "",
     },
   ]
   repl-lint:
     version: 1
     linters:
       helm:
         disabled: false
   EOF
   ```

4. Run lint manually to see what messages are produced:
   ```bash
   cd testdata/my-new-test
   ../../bin/replicated lint
   ```

5. Create `expect.json` based on the output:
   ```bash
   cat > expect.json << EOF
   {
     "lintMessages": []
   }
   EOF
   ```

6. Run `make test-lint` to verify

## Example Test Cases

### `chart-with-required-values/`

Tests a valid Helm chart with required values that should pass linting with no errors.

- **Purpose**: Verify that well-formed charts produce no lint errors
- **Expected**: Exit code 0, no lint messages

## Troubleshooting

### Test fails with message count mismatch

The actual lint output doesn't match the expected messages in `expect.json`. The test output will show:
- Expected message counts by severity
- Actual message counts
- The full lint output for debugging

**Fix**: Update `expect.json` to match the actual output, or fix the test case to produce the expected output.

### Binary not found

```
Error: Binary not found at ./bin/replicated
Please run 'make build' first
```

**Fix**: Run `make build` to build the binary.

### No .replicated config found

```
✗ FAILED: No .replicated config found
```

**Fix**: Add a `.replicated` or `.replicated.yaml` file to the test directory.

## CI Integration

Add to your CI pipeline:

```yaml
- name: Run lint tests
  run: make test-lint
```

This ensures lint behavior remains consistent across changes.

