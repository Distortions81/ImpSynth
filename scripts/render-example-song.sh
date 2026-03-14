#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

input_path="${1:-$repo_root/examples/twinkle.csv}"
output_path="${2:-$repo_root/out/twinkle.wav}"
patch_path="${3:-$repo_root/examples/patches/xylophone.json}"

mkdir -p "$(dirname "$output_path")"

cd "$repo_root"
go run ./cmd/impsynth-wav-example "$input_path" "$output_path" "$patch_path"
