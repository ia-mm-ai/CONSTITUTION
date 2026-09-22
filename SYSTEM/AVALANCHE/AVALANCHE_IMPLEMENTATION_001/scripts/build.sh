#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
go_binary="${GO_BINARY:-go}"
output_dir="${1:-${root_dir}/build}"

mkdir -p "${output_dir}"

(
  cd "${root_dir}"
  CGO_ENABLED=1 "${go_binary}" build -buildvcs=false -trimpath -ldflags='-s -w' -o "${output_dir}/presence-avalanche-vm" .
)

"${output_dir}/presence-avalanche-vm" --version
echo "built LOCALITY VM artifacts in ${output_dir}"
