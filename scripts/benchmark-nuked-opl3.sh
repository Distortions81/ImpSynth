#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out_bin="${TMPDIR:-/tmp}/nuked-opl3-bench"
iterations="${1:-2048}"
sample_rate="${2:-49716}"
nuked_dir="$("$repo_root/scripts/fetch-nuked-opl3.sh")"

cc -O3 -std=c99 \
  -I"$nuked_dir" \
  "$repo_root/bench/nuked_opl3_bench.c" \
  "$nuked_dir/opl3.c" \
  -o "$out_bin"

"$out_bin" "$iterations" "$sample_rate"
