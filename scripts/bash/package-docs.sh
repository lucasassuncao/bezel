#!/usr/bin/env bash
# Writes a README.md into every package matched by the <package> patterns with
# gomarkdoc, linking the source to <repository-url> on <default-branch>.
set -euo pipefail

if [ $# -lt 4 ]; then
	echo "usage: package-docs.sh <gomarkdoc> <repository-url> <default-branch> <package>..." >&2
	exit 2
fi

gomarkdoc=$1
url=$2
branch=$3
shift 3

"$gomarkdoc" -e \
	--repository.url "$url" \
	--repository.default-branch "$branch" \
	--repository.path / \
	-o '{{.Dir}}/README.md' "$@"
