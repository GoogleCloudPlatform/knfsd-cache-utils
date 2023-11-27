#!/usr/bin/env bash

shopt -s extglob globstar

function usage() {
	printf 'Syntax: ./update-tags.sh <TAG>\n'
	printf '    Example: ./update-tags.sh v1.0\n'
}

if [[ $1 == '' ]]; then
	usage 1>&2
	exit 2
fi

match='(source\s*=\s*"github.com/GoogleCloudPlatform/knfsd-cache-utils//.+\?ref=).*"'
replace='\1'$1'"'

# Exclude ./docs/changes/ as it contains historical information about old releases.
find . -type f -and \
	\( -name '*.tf' -or -name '*.md' \) \
	-and \! -path './docs/changes/*' \
	-print0 |
while read -rd $'\0' f; do
	sed -i -r "s#$match#$replace#g" "$f"
done
