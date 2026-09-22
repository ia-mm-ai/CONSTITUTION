#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
output="${1:-${root}/build}"
source_dir="${2:-.}"
case "${source_dir}" in
  .|vm) ;;
  *) echo "source directory must be . or vm" >&2; exit 1 ;;
esac
mkdir -p "${output}"
output="$(cd "${output}" && pwd)"
python3 "${root}/scripts/verify-binding.py"
python3 "${root}/scripts/verify-bindings.py"
cd "${root}/${source_dir}"
export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.25.13}"
CGO_ENABLED=1 "${GO_BINARY:-go}" build -mod=readonly -buildvcs=false -trimpath -o "${output}/presence-avalanche-vm" .
"${output}/presence-avalanche-vm" --version-json
