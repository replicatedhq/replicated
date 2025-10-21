#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BINARY="${BINARY:-./bin/replicated}"
TESTDATA_DIR="./testdata"

if [ ! -f "$BINARY" ]; then
    echo -e "${RED}Error: Binary not found at $BINARY${NC}"
    echo "Please run 'make build' first"
    exit 1
fi

# Find all directories in testdata that contain expect.json
test_dirs=$(find "$TESTDATA_DIR" -type f -name "expect.json" -exec dirname {} \;)

if [ -z "$test_dirs" ]; then
    echo -e "${YELLOW}No test directories found with expect.json${NC}"
    exit 0
fi

total_tests=0
passed_tests=0
failed_tests=0

echo "Running lint tests..."
echo ""

for test_dir in $test_dirs; do
    total_tests=$((total_tests + 1))
    test_name=$(basename "$test_dir")

    echo -e "Testing: ${YELLOW}$test_name${NC}"

    # Check if .replicated or .replicated.yaml exists
    if [ ! -f "$test_dir/.replicated" ] && [ ! -f "$test_dir/.replicated.yaml" ]; then
        echo -e "  ${RED}✗ FAILED${NC}: No .replicated config found"
        failed_tests=$((failed_tests + 1))
        echo ""
        continue
    fi

    # Run lint command and capture output and exit code
    cd "$test_dir"
    set +e
    output=$("$OLDPWD/$BINARY" lint 2>&1)
    exit_code=$?
    set -e
    cd "$OLDPWD"

    # Read expected results
    expected_file="$test_dir/expect.json"
    expected_lint_messages=$(jq -r '.lintMessages | length' "$expected_file")

    # Determine expected exit code based on lint messages
    # If there are any error-level messages, we expect non-zero exit
    expected_has_errors=$(jq -r '[.lintMessages[] | select(.severity == "ERROR")] | length > 0' "$expected_file")

    if [ "$expected_has_errors" = "true" ]; then
        expected_exit_code=1
    else
        expected_exit_code=0
    fi

    # Check exit code
    if [ "$exit_code" -ne "$expected_exit_code" ]; then
        echo -e "  ${RED}✗ FAILED${NC}: Expected exit code $expected_exit_code but got $exit_code"
        echo "  Output:"
        echo "$output" | sed 's/^/    /'
        failed_tests=$((failed_tests + 1))
        echo ""
        continue
    fi

    # Parse lint messages from output
    # The output format is: [SEVERITY] Path: Message or [SEVERITY] Message
    # We'll count the number of ERROR, WARNING, and INFO messages
    actual_errors=$(echo "$output" | grep -c "\[ERROR\]" || true)
    actual_warnings=$(echo "$output" | grep -c "\[WARNING\]" || true)
    actual_info=$(echo "$output" | grep -c "\[INFO\]" || true)

    # Count expected messages by severity
    expected_errors=$(jq -r '[.lintMessages[] | select(.severity == "ERROR")] | length' "$expected_file")
    expected_warnings=$(jq -r '[.lintMessages[] | select(.severity == "WARNING")] | length' "$expected_file")
    expected_info=$(jq -r '[.lintMessages[] | select(.severity == "INFO")] | length' "$expected_file")

    # Compare counts
    if [ "$actual_errors" -ne "$expected_errors" ] || \
       [ "$actual_warnings" -ne "$expected_warnings" ] || \
       [ "$actual_info" -ne "$expected_info" ]; then
        echo -e "  ${RED}✗ FAILED${NC}: Message count mismatch"
        echo "    Expected: $expected_errors error(s), $expected_warnings warning(s), $expected_info info"
        echo "    Actual:   $actual_errors error(s), $actual_warnings warning(s), $actual_info info"
        echo ""
        echo "  Expected lint messages:"
        jq -r '.lintMessages[] | "    [\(.severity)] \(.path // "")\(.path | if . then ": " else "" end)\(.message)"' "$expected_file"
        echo ""
        echo "  Actual output:"
        echo "$output" | sed 's/^/    /'
        failed_tests=$((failed_tests + 1))
        echo ""
        continue
    fi

    # If we want strict message matching (optional - can be enabled later)
    # For now, we just check counts

    echo -e "  ${GREEN}✓ PASSED${NC}"
    passed_tests=$((passed_tests + 1))
    echo ""
done

# Print summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Total:  $total_tests"
echo -e "Passed: ${GREEN}$passed_tests${NC}"
if [ $failed_tests -gt 0 ]; then
    echo -e "Failed: ${RED}$failed_tests${NC}"
else
    echo -e "Failed: $failed_tests"
fi
echo ""

if [ $failed_tests -gt 0 ]; then
    exit 1
fi

echo -e "${GREEN}All tests passed!${NC}"

