#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 OUTPUT_DIRECTORY" >&2
  exit 2
fi

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
output_dir="$1"
go_binary="${GO_BINARY:-go}"
if [[ -e "${output_dir}" ]]; then
  echo "refusing to overwrite output directory: ${output_dir}" >&2
  exit 1
fi
if ! go_binary="$(command -v "${go_binary}")"; then
  echo "Go compiler is unavailable" >&2
  exit 1
fi
if [[ "$("${go_binary}" env GOVERSION)" != "go1.25.13" ]]; then
  echo "AvalancheGo build requires exact Go 1.25.13" >&2
  exit 1
fi
go_binary_dir="$(dirname "${go_binary}")"
export PATH="${go_binary_dir}:${PATH}"

scratch_root="$(mktemp -d "${TMPDIR:-/tmp}/presence-avalanchego-build-001.XXXXXX")"
cleanup() {
  case "${scratch_root}" in
    "${TMPDIR:-/tmp}"/presence-avalanchego-build-001.*) rm -rf -- "${scratch_root}" ;;
    *) echo "refusing to remove unexpected build directory: ${scratch_root}" >&2 ;;
  esac
}
trap cleanup EXIT INT TERM

GO_BINARY="${go_binary}" "${root_dir}/scripts/prepare-avalanchego.sh" "${scratch_root}/avalanchego"
(
  cd "${scratch_root}/avalanchego"
  AVALANCHEGO_COMMIT="70bd6d063b7343fd2cd8217200aaf77b57f19f68" bash ./scripts/build.sh
)

mkdir -p "${output_dir}"
output_dir="$(cd "${output_dir}" && pwd)"
install -m 0755 "${scratch_root}/avalanchego/build/avalanchego" "${output_dir}/avalanchego"
cp "${root_dir}/compat/AVALANCHEGO_SECURITY_OVERLAY_001.json" "${output_dir}/"
"${go_binary}" version -m "${output_dir}/avalanchego" >"${output_dir}/BUILD_INFO.txt"
(
  cd "${output_dir}"
  sha256sum avalanchego AVALANCHEGO_SECURITY_OVERLAY_001.json BUILD_INFO.txt >SHA256SUMS
)

echo "security-overlaid AvalancheGo build: ${output_dir}"
