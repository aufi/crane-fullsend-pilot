#!/usr/bin/env bash

set -euo pipefail

tracked_go_files=$(git ls-files -- '*.go')
if [[ -z "$tracked_go_files" ]]; then
	exit 0
fi

mapfile -t go_files <<< "$tracked_go_files"
offending_files=$(gofmt -l "${go_files[@]}")
if [[ -n "$offending_files" ]]; then
	echo "gofmt would change the following tracked Go files:"
	printf '%s\n' "$offending_files"
	exit 1
fi
