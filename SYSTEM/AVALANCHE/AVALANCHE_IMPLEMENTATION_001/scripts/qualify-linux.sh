#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
go_binary="${GO_BINARY:-go}"
avalanchego_tag="v1.15.0"
avalanchego_commit="70bd6d063b7343fd2cd8217200aaf77b57f19f68"

if [[ "$(uname -s)" != "Linux" ]]; then
  echo "canonical qualification requires Linux; observed $(uname -s)" >&2
  exit 1
fi
for command_name in git python3 sha256sum; do
  if ! command -v "${command_name}" >/dev/null 2>&1; then
    echo "required command is unavailable: ${command_name}" >&2
    exit 1
  fi
done
if ! go_binary="$(command -v "${go_binary}")"; then
  echo "Go compiler is unavailable" >&2
  exit 1
fi
export PATH="$(dirname "${go_binary}"):${PATH}"
if [[ "$("${go_binary}" env GOVERSION)" != "go1.25.13" ]]; then
  echo "qualification requires exact Go 1.25.13; observed $("${go_binary}" env GOVERSION)" >&2
  exit 1
fi

timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
output_dir="${LOCALITY_QUALIFICATION_OUTPUT:-${root_dir}/artifacts/qualification-${timestamp}}"
if [[ -e "${output_dir}" ]]; then
  echo "refusing to overwrite qualification output: ${output_dir}" >&2
  exit 1
fi
mkdir -p "${output_dir}"

scratch_root="$(mktemp -d "${TMPDIR:-/tmp}/presence-avalanche-vm-qualification-003.XXXXXX")"
cleanup() {
  case "${scratch_root}" in
    "${TMPDIR:-/tmp}"/presence-avalanche-vm-qualification-003.*) rm -rf -- "${scratch_root}" ;;
    *) echo "refusing to remove unexpected qualification directory: ${scratch_root}" >&2 ;;
  esac
}
trap cleanup EXIT INT TERM

"${root_dir}/scripts/verify.sh"
"${root_dir}/scripts/verify_public_origin.py" \
  --report "${output_dir}/PUBLIC_ORIGIN_VERIFICATION.json"
GO_BINARY="${go_binary}" "${root_dir}/scripts/vulnerability-scan.sh" \
  | tee "${output_dir}/VM_VULNERABILITY_SCAN.txt"

avalanchego_source="${scratch_root}/avalanchego"
GO_BINARY="${go_binary}" "${root_dir}/scripts/prepare-avalanchego.sh" "${avalanchego_source}"
source_identity="${avalanchego_source}/AVALANCHEGO_SOURCE_IDENTITY.json"
observed_commit="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1], encoding="utf-8"))["peeled_commit"])' "${source_identity}")"
[[ "${observed_commit}" == "${avalanchego_commit}" ]]
(
  cd "${avalanchego_source}"
  GOTOOLCHAIN=local GOWORK=off "${go_binary}" run \
    golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./main \
    | tee "${output_dir}/AVALANCHEGO_VULNERABILITY_SCAN.txt"
)
(
  cd "${avalanchego_source}"
  AVALANCHEGO_COMMIT="${avalanchego_commit}" bash ./scripts/build.sh
)
avalanchego_binary="${avalanchego_source}/build/avalanchego"
if [[ ! -x "${avalanchego_binary}" ]]; then
  echo "AvalancheGo build did not produce ${avalanchego_binary}" >&2
  exit 1
fi

PRESENCE_REHEARSAL_ROOT="${scratch_root}/private-rehearsal" \
LOCALITY_PUBLIC_EVIDENCE_ROOT="${output_dir}/rehearsal" \
"${root_dir}/scripts/rehearsal-003.sh" "${avalanchego_binary}"

vm_binary="${root_dir}/build/presence-avalanche-vm"
vm_id="$("${vm_binary}" --vm-id)"
vm_sha256="$(sha256sum "${vm_binary}" | awk '{print $1}')"
avalanchego_sha256="$(sha256sum "${avalanchego_binary}" | awk '{print $1}')"
go_version="$("${go_binary}" version)"
"${go_binary}" version -m "${vm_binary}" >"${output_dir}/VM_BUILD_INFO.txt"
"${go_binary}" version -m "${avalanchego_binary}" >"${output_dir}/AVALANCHEGO_BUILD_INFO.txt"
cp "${source_identity}" "${output_dir}/AVALANCHEGO_SOURCE_IDENTITY.json"
export LOCALITY_REPORT_PATH="${output_dir}/QUALIFICATION_REPORT.json"
export LOCALITY_OBSERVED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
export LOCALITY_VM_ID="${vm_id}"
export LOCALITY_VM_SHA256="${vm_sha256}"
export LOCALITY_AVALANCHEGO_SHA256="${avalanchego_sha256}"
export LOCALITY_GO_VERSION="${go_version}"
export LOCALITY_OVERLAY_SHA256="$(sha256sum "${root_dir}/compat/AVALANCHEGO_SECURITY_OVERLAY_001.json" | awk '{print $1}')"
python3 - <<'PY'
import json
import os
import pathlib
import platform

report = {
    "schema": "LOCALITY_VM_QUALIFICATION_REPORT_003",
    "result": "PASS",
    "observed_at_utc": os.environ["LOCALITY_OBSERVED_AT"],
    "host": {"system": platform.system(), "machine": platform.machine()},
    "vm": {
        "protocol": "PRESENCE_AVALANCHE_VM_001",
        "version": "3.0.0",
        "vm_id": os.environ["LOCALITY_VM_ID"],
        "binary_sha256": os.environ["LOCALITY_VM_SHA256"],
    },
    "avalanchego": {
        "tag": "v1.15.0",
        "peeled_commit": "70bd6d063b7343fd2cd8217200aaf77b57f19f68",
        "profile": "v1.15.0+PRESENCE_SECURITY_OVERLAY_001",
        "security_overlay": {
            "id": "PRESENCE_SECURITY_OVERLAY_001",
            "sha256": os.environ["LOCALITY_OVERLAY_SHA256"],
        },
        "binary_sha256": os.environ["LOCALITY_AVALANCHEGO_SHA256"],
        "rpcchainvm_protocol": 46,
    },
    "go_version": os.environ["LOCALITY_GO_VERSION"],
    "gates": [
        "REPOSITORY_INTEGRITY",
        "PUBLIC_ORIGIN_EXACT_BINDING",
        "MACHINE_TO_HUMAN_SOURCE_QUOTATION_BINDING",
        "GO_TEST",
        "GO_RACE",
        "GO_VET",
        "NATIVE_BUILD",
        "EXACT_AVALANCHEGO_TAG_AND_COMMIT",
        "AUTHENTICATED_AVALANCHEGO_MODULE_SOURCE",
        "REPRODUCIBLE_AVALANCHEGO_SECURITY_OVERLAY",
        "ZERO_REACHABLE_VULNERABILITIES_VM",
        "ZERO_REACHABLE_VULNERABILITIES_AVALANCHEGO",
        "THREE_VALIDATOR_LIFECYCLE",
        "DEPARTURE_STATE_CHECKPOINT",
        "SINGLE_USE_CARRIED_STATE_REENTRY",
        "LIVING_CAPACITY_AND_P0_DORMANCY",
        "FORMATION_AUTHORITY_EXHAUSTION",
        "ATTESTED_SUCCESSOR_FREEZE",
        "FULL_NETWORK_RESTART_REPLAY",
        "SINGLE_VALIDATOR_REJOIN",
    ],
    "private_material_exported": False,
}
path = pathlib.Path(os.environ["LOCALITY_REPORT_PATH"])
path.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
PY
cp "${root_dir}/RELEASE_MANIFEST.json" "${output_dir}/"
cp "${root_dir}/MANIFEST.sha256" "${output_dir}/SOURCE_MANIFEST.sha256"
(
  cd "${output_dir}"
  find . -type f ! -name SHA256SUMS -print0 | sort -z | xargs -0 sha256sum >SHA256SUMS
)

echo "PRESENCE AVALANCHE VM 003 canonical qualification: PASS"
echo "public evidence: ${output_dir}"
