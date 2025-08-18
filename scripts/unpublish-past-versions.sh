#!/bin/bash

set -euo pipefail

PACTICIPANT="${PACTICIPANT:-replicated-cli}"
PACT_BROKER_BASE_URL="${PACT_BROKER_BASE_URL:-https://replicated.pactflow.io}"

if [[ -z "${PACT_BROKER_TOKEN:-}" ]]; then
    echo "Error: PACT_BROKER_TOKEN environment variable is required"
    exit 1
fi

echo "Fetching latest version for pacticipant: $PACTICIPANT"

# Get the latest version
LATEST_VERSION=$(curl -s \
    -H "Authorization: Bearer ${PACT_BROKER_TOKEN}" \
    -H "Accept: application/hal+json" \
    "${PACT_BROKER_BASE_URL}/pacticipants/${PACTICIPANT}/latest-version" | \
    jq -r '.number')

if [[ "$LATEST_VERSION" == "null" || -z "$LATEST_VERSION" ]]; then
    echo "Error: Could not fetch latest version"
    exit 1
fi

echo "Latest version: $LATEST_VERSION"
echo ""

echo "Fetching all versions..."

# Get all versions and extract version numbers (excluding the latest)
ALL_VERSIONS=$(curl -s \
    -H "Authorization: Bearer ${PACT_BROKER_TOKEN}" \
    -H "Accept: application/hal+json" \
    "${PACT_BROKER_BASE_URL}/pacticipants/${PACTICIPANT}/versions?size=400" | \
    jq -r '._embedded.versions[].number' | \
    grep -v "^${LATEST_VERSION}$")

if [[ -z "$ALL_VERSIONS" ]]; then
    echo "No versions to delete (only latest version exists)"
    exit 0
fi

VERSION_COUNT=$(echo "$ALL_VERSIONS" | wc -l | xargs)
echo "Found $VERSION_COUNT versions to delete (excluding latest: $LATEST_VERSION)"
echo ""

# Delete each version
echo "$ALL_VERSIONS" | while IFS= read -r version; do
    echo "Deleting version $version..."
    
    HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null \
        -X DELETE \
        -H "Authorization: Bearer ${PACT_BROKER_TOKEN}" \
        -H "Accept: application/hal+json" \
        "${PACT_BROKER_BASE_URL}/pacticipants/${PACTICIPANT}/versions/${version}")
    
    if [[ "$HTTP_CODE" =~ ^2 ]]; then
        echo "✓ Successfully deleted $version"
    else
        echo "✗ Failed to delete $version (HTTP $HTTP_CODE)"
        exit 1
    fi
    
    echo ""
done

echo "All old versions deleted successfully! Latest version $LATEST_VERSION preserved."