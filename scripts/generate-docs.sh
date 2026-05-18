#!/bin/bash
set -euo pipefail

# Generate CLI docs and open a PR against replicatedhq/replicated-docs
# This replaces the Dagger GenerateDocs function.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

VERSION="${GITHUB_REF_NAME:-unknown}"
BRANCH_NAME="update-cli-docs-${VERSION}-$(date +%Y-%m-%d-%H%M%S)"

# Build the binary and generate docs
cd "${REPO_ROOT}"
go run -tags "containers_image_ostree_stub exclude_graphdriver_devicemapper exclude_graphdriver_btrfs containers_image_openpgp" ./docs/gen.go

# Clone replicated-docs
DOCS_DIR="/tmp/replicated-docs"
rm -rf "${DOCS_DIR}"
git clone --depth 1 "https://${GITHUB_TOKEN}@github.com/replicatedhq/replicated-docs.git" "${DOCS_DIR}"

cd "${DOCS_DIR}"
git config user.email "release@replicated.com"
git config user.name "Replicated Release Pipeline"

# Remove existing CLI docs (except installing)
for f in docs/reference/replicated-cli-*.mdx; do
    if [ -f "$f" ] && [ "$(basename "$f")" != "replicated-cli-installing.mdx" ]; then
        rm "$f"
    fi
done

# Remove the root CLI doc too
if [ -f "docs/reference/replicated.mdx" ]; then
    rm "docs/reference/replicated.mdx"
fi

# Copy generated docs
cd "${REPO_ROOT}"
for f in gen/docs/*.md; do
    if [ ! -f "$f" ]; then
        continue
    fi
    # Convert filename: replicated_channel_inspect.md -> replicated-cli-channel-inspect.mdx
    dest_name=$(basename "$f" .md | sed 's/replicated_/replicated-cli-/g; s/_/-/g').mdx
    
    content=$(cat "$f")
    
    # Fix header level
    if [[ "$content" == "## "* ]]; then
        content="${content:1}"
    fi
    
    # Replace internal links
    for ref in gen/docs/*.md; do
        ref_name=$(basename "$ref" .md)
        dest_ref=$(echo "$ref_name" | sed 's/replicated_/replicated-cli-/g; s/_/-/g')
        content="${content//${ref_name}.md/${dest_ref}}"
    done
    
    echo "$content" > "${DOCS_DIR}/docs/reference/${dest_name}"
done

# Update sidebars.js
cd "${DOCS_DIR}"
node -e "
const fs = require('fs');
let sidebar = fs.readFileSync('sidebars.js', 'utf8');

// Find the Replicated CLI section and replace its items
const cliItems = ['reference/replicated-cli-installing'];
const files = fs.readdirSync('docs/reference')
    .filter(f => (f === 'replicated.mdx' || f.startsWith('replicated-cli-')) && f !== 'replicated-cli-installing.mdx')
    .map(f => 'reference/' + f.replace('.mdx', ''))
    .sort();
cliItems.push(...files);

// Simple regex replacement for the items array
const pattern = /(label:\s*['\"]Replicated CLI['\"],?\s*\n\s*items:\s*\[)[^\]]*(\])/;
const replacement = '\$1\n      ' + cliItems.map(i => '\"' + i + '\"').join(',\n      ') + '\n    \$2';
sidebar = sidebar.replace(pattern, replacement);

fs.writeFileSync('sidebars.js', sidebar);
"

# Commit and push
git checkout -b "${BRANCH_NAME}"
git add .
git commit -m "Update Replicated CLI docs for ${VERSION}"
git push origin "${BRANCH_NAME}"

# Open PR
response=$(curl -s -f -X POST \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    -H "Accept: application/vnd.github+json" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    https://api.github.com/repos/replicatedhq/replicated-docs/pulls \
    -d "{
        \"title\": \"Update Replicated CLI docs for ${VERSION}\",
        \"head\": \"${BRANCH_NAME}\",
        \"base\": \"main\"
    }")
echo "${response}" | jq -r '.html_url'
