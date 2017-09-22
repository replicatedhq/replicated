#!/bin/bash
set -e

TAR_FILE="/tmp/replicated.tar.gz"

RELEASES_URL="https://github.com/replicatedhq/replicated/releases"

test -z "$TMPDIR" && TMPDIR="$(mktemp -d)"

last_version() {
	curl --silent --location --output /dev/null --write-out %{url_effective} ${RELEASES_URL}/latest |
		grep -Eo '[[:digit:]]+\.[[:digit:]]+\.[[:digit:]]$'
}

download() {
	if [[ $(uname -m) =~ '64$' ]]; then
		ARCH=amd64
	else
		ARCH=386
	fi
	VERSION="$(last_version)"
	# https://github.com/replicatedhq/replicated/releases/download/v0.4.0/cli_0.4.0_linux_amd64.tar.gz
	URL="${RELEASES_URL}/download/v${VERSION}/cli_${VERSION}_$(uname -s)_${ARCH}.tar.gz"

	rm -f "$TAR_FILE"
	curl -s -L -o "$TAR_FILE" "$URL"
}

download
tar -xf "$TAR_FILE" -C "$TMPDIR"
mv "${TMPDIR}/replicated" /usr/local/bin
