#!/bin/bash
set -e

RELEASES_URL="https://github.com/replicatedhq/replicated/releases"

last_version() {
	curl --silent --location --fail \
        --output /dev/null --write-out %{url_effective} ${RELEASES_URL}/latest |
		grep -Eo '[0-9]+\.[0-9]+\.[0-9]+$'
}

download_tar() {
	if expr "$(uname -m)" : '.*64$' &>/dev/null; then
		ARCH=amd64
	else
		ARCH=386
	fi
	VERSION="$(last_version)"
	# https://github.com/replicatedhq/replicated/releases/download/v0.5.0/replicated_0.5.0_linux_amd64.tar.gz
	URL="${RELEASES_URL}/download/v${VERSION}/replicated_${VERSION}_$(uname -s)_${ARCH}.tar.gz"

	curl --silent --location --fail "$URL"
}

main() {
	default_dir=/usr/local/bin
	tar_binary=replicated

	if [[ -z "$replicated_bindir" ]]; then
	    replicated_bindir="$default_dir"
	    if [[ "$1" ]]; then
	        replicated_bindir="$1"
	    fi
	fi
	if [[ ! -d "$replicated_bindir" ]]; then
	    cat >&2 <<MSG
Destination directory "$replicated_bindir" is not a directory

Usage: $0 [install-dir]
 If install-dir is not provided, the script will use "$default_dir"
MSG
	    exit 1
	fi
	download_tar | tar -xzf - -C "$replicated_bindir" $tar_binary
}

main "$@"
