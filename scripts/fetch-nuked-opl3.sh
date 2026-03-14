#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out_dir="$repo_root/third_party/nuked-opl3"
commit="cfedb09efc03f1d7b5fc1f04dd449d77d8c49d50"
base_url="https://raw.githubusercontent.com/nukeykt/Nuked-OPL3/$commit"

mkdir -p "$out_dir"

fetch_file() {
  local name="$1"
  local target="$out_dir/$name"
  local tmp="$target.$$.$RANDOM.tmp"

  if [ -s "$target" ]; then
    return
  fi

  curl -fsSL "$base_url/$name" -o "$tmp"
  mv -f "$tmp" "$target"
}

fetch_file "opl3.h"
fetch_file "opl3.c"

printf '%s\n' "$out_dir"
