#!/bin/bash
set -euo pipefail

VERSION="${1:?Usage: $0 <major|minor|patch|X.Y.Z>}"

# Validate clean git tree on main
if [[ -n $(git status --porcelain) ]]; then
  echo "Error: git tree is not clean" >&2
  exit 1
fi

BRANCH=$(git branch --show-current)
if [[ "$BRANCH" != "main" ]]; then
  echo "Error: must be on main branch (currently on $BRANCH)" >&2
  exit 1
fi

# Verify HEAD is pushed to remote
HEAD=$(git rev-parse HEAD)
git fetch origin main --quiet
REMOTE_HEAD=$(git rev-parse origin/main)
if [[ "$HEAD" != "$REMOTE_HEAD" ]]; then
  echo "Error: local HEAD doesn't match origin/main. Push your changes first." >&2
  exit 1
fi

# Get latest version tag
LATEST_TAG=$(git describe --tags --abbrev=0 --match 'v*' 2>/dev/null || echo "v0.0.0")

# Parse semver from latest tag
if [[ "$LATEST_TAG" =~ ^v?([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
  CUR_MAJOR="${BASH_REMATCH[1]}"
  CUR_MINOR="${BASH_REMATCH[2]}"
  CUR_PATCH="${BASH_REMATCH[3]}"
else
  echo "Error: could not parse latest tag '$LATEST_TAG'" >&2
  exit 1
fi

# Calculate next version
case "$VERSION" in
  major)
    NEW_MAJOR=$((CUR_MAJOR + 1))
    NEW_MINOR=0
    NEW_PATCH=0
    ;;
  minor)
    NEW_MAJOR=$CUR_MAJOR
    NEW_MINOR=$((CUR_MINOR + 1))
    NEW_PATCH=0
    ;;
  patch)
    NEW_MAJOR=$CUR_MAJOR
    NEW_MINOR=$CUR_MINOR
    NEW_PATCH=$((CUR_PATCH + 1))
    ;;
  *)
    if [[ "$VERSION" =~ ^v?([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
      NEW_MAJOR="${BASH_REMATCH[1]}"
      NEW_MINOR="${BASH_REMATCH[2]}"
      NEW_PATCH="${BASH_REMATCH[3]}"
    else
      echo "Error: version must be 'major', 'minor', 'patch', or a semver string (e.g. 1.2.3)" >&2
      exit 1
    fi
    ;;
esac

NEW_VERSION="${NEW_MAJOR}.${NEW_MINOR}.${NEW_PATCH}"
TAG="v${NEW_VERSION}"
RELEASE_BRANCH="release-${NEW_VERSION}"

echo "Releasing as version ${NEW_VERSION} (previous: ${LATEST_TAG})"

# Update build.go with the version
sed -i.bak "s/const version = \"unknown\"/const version = \"${NEW_VERSION}\"/" pkg/version/build.go
rm -f pkg/version/build.go.bak

# Create release branch with version commit
git checkout -b "$RELEASE_BRANCH"
git add pkg/version/build.go
git commit -m "Set version to ${NEW_VERSION}"
git push origin "$RELEASE_BRANCH"

# Create and push tag on the release branch
git tag "$TAG"
git push origin "$TAG"

# Return to main (build.go on main stays as "unknown")
git checkout main

echo ""
echo "✓ Tag ${TAG} created and pushed"
echo "✓ Release branch ${RELEASE_BRANCH} pushed"
echo "GitHub Actions will handle the release build"
