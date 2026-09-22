#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
output="${1:-${root}/build}"
mkdir -p "${output}"
output="$(cd "${output}" && pwd)"
python3 "${root}/scripts/verify-binding.py"
python3 "${root}/scripts/verify-bindings.py"
cd "${root}/vm"
export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.25.13}"
CGO_ENABLED=1 "${GO_BINARY:-go}" build -mod=readonly -buildvcs=false -trimpath -o "${output}/presence-avalanche-vm" .
"${output}/presence-avalanche-vm" --version-json
