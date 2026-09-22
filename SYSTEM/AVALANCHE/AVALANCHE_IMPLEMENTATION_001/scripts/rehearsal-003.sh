#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
go_binary="${GO_BINARY:-go}"
python_binary="${PYTHON_BINARY:-python3}"
avalanchego_binary="${1:-${AVALANCHEGO_BINARY:-}}"

if [[ -z "${avalanchego_binary}" ]]; then
  echo "usage: AVALANCHEGO_BINARY=/absolute/path/to/avalanchego $0" >&2
  exit 2
fi
if [[ ! -x "${avalanchego_binary}" ]]; then
  echo "AvalancheGo binary is not executable: ${avalanchego_binary}" >&2
  exit 2
fi

avalanchego_binary="$(cd "$(dirname "${avalanchego_binary}")" && pwd)/$(basename "${avalanchego_binary}")"
private_root="${PRESENCE_REHEARSAL_ROOT:-${root_dir}/rehearsal/runs}"
public_root="${LOCALITY_PUBLIC_EVIDENCE_ROOT:-${root_dir}/rehearsal/evidence}"
mkdir -p "${private_root}" "${public_root}"
chmod 700 "${private_root}"

timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
run_dir="$(mktemp -d "${private_root}/rehearsal-003-${timestamp}-XXXXXX")"
public_dir="${public_root}/$(basename "${run_dir}")"
mkdir "${public_dir}"
chmod 700 "${run_dir}"

network_summary="${run_dir}/network-start.json"
network_dir=""
success=0

remove_private_run() {
  case "${run_dir}" in
    "${private_root}"/rehearsal-003-*) rm -rf -- "${run_dir}" ;;
    *) echo "refusing to remove unexpected run directory: ${run_dir}" >&2; return 1 ;;
  esac
}

cleanup() {
  local exit_status=$?
  trap - EXIT INT TERM
  if [[ -n "${network_dir}" && -d "${network_dir}" ]]; then
    "${root_dir}/build/locality-rehearsal" stop --network-dir "${network_dir}" >/dev/null 2>&1 || true
  fi
  if [[ "${success}" == "1" && "${LOCALITY_KEEP_PRIVATE_RUN:-0}" != "1" ]]; then
    remove_private_run
  elif [[ "${success}" != "1" ]]; then
    rmdir "${public_dir}" >/dev/null 2>&1 || true
    echo "failed private rehearsal retained for diagnosis: ${run_dir}" >&2
  fi
  exit "${exit_status}"
}
trap cleanup EXIT INT TERM

go_version="$("${go_binary}" version)"
if [[ "$("${go_binary}" env GOVERSION)" != "go1.25.13" ]]; then
  echo "PRESENCE REHEARSAL 001 requires exact Go 1.25.13; observed: ${go_version}" >&2
  exit 1
fi

"${python_binary}" -c '
import json, subprocess, sys
document = json.loads(subprocess.check_output([sys.argv[1], "--version-json"], text=True))
if (
    document.get("application") != "avalanchego/1.15.0"
    or document.get("rpcchainvm") != 46
    or document.get("go") != "1.25.13"
):
    raise SystemExit(f"requires avalanchego/1.15.0 rpcchainvm=46 go=1.25.13; observed {document}")
' "${avalanchego_binary}"

avalanchego_build_info="$("${go_binary}" version -m "${avalanchego_binary}")"
for required_dependency in \
  $'\tdep\tgoogle.golang.org/grpc\tv1.83.2\t' \
  $'\tdep\tgolang.org/x/net\tv0.58.0\t' \
  $'\tdep\tgolang.org/x/text\tv0.41.0\t'; do
  if [[ "${avalanchego_build_info}" != *"${required_dependency}"* ]]; then
    echo "AvalancheGo binary lacks required PRESENCE_SECURITY_OVERLAY_001 dependency: ${required_dependency}" >&2
    exit 1
  fi
done

GO_BINARY="${go_binary}" "${root_dir}/scripts/verify.sh"
(
  cd "${root_dir}/rehearsal"
  GOWORK=off "${go_binary}" test -buildvcs=false ./...
  GOWORK=off "${go_binary}" build -buildvcs=false -o "${root_dir}/build/locality-rehearsal" .
)

vm_binary="${root_dir}/build/presence-avalanche-vm"
vm_id="$("${vm_binary}" --vm-id)"
vm_build_info="$("${go_binary}" version -m "${vm_binary}")"
if [[ "${vm_build_info}" != *$'\tdep\tgoogle.golang.org/grpc\tv1.83.2\t'* ]]; then
  echo "PRESENCE AVALANCHE VM binary lacks the required gRPC security update" >&2
  exit 1
fi
if [[ "${vm_id}" == "pJHx1NU8ghWsiwg1vrqwaqE5uKh1k4EJRBkpB1tKV1QuQFhMi" ]]; then
  echo "successor VM ID must not reuse the formation VM ID" >&2
  exit 1
fi

mkdir -m 700 "${run_dir}/secrets" "${run_dir}/secrets/host" "${run_dir}/secrets/participant"
"${vm_binary}" --generate-authority "${run_dir}/secrets/host" >/dev/null
"${vm_binary}" --generate-authority "${run_dir}/secrets/participant" >/dev/null

deployment_descriptor="${run_dir}/PRESENCE_REHEARSAL_DEPLOYMENT_003.json"
"${python_binary}" -c '
import json, pathlib, sys
document = {
    "schema": "PRESENCE_DEPLOYMENT_DESCRIPTOR_001",
    "genesis_id": "PRESENCE-REHEARSAL-GENESIS-003",
    "locality_id": "PRESENCE-REHEARSAL-HOST-003",
    "purpose": "Exercise PRESENCE AVALANCHE VM 003 origin-bound living-capacity lifecycle boundaries on a disposable three-validator network.",
    "source_reference": "PRESENCE_AVALANCHE_VM_001_REHEARSAL_SOURCE",
    "source_sha256": "c02f9a3927f7e2a48c49a4d2267b0cc7c8a932694817374366d8ec5709c419bd",
}
pathlib.Path(sys.argv[1]).write_text(json.dumps(document, indent=2) + "\n", encoding="utf-8")
' "${deployment_descriptor}"

genesis_path="${run_dir}/PRESENCE_RUNTIME_GENESIS_REHEARSAL_003.json"
"${vm_binary}" --materialize-genesis \
  "${root_dir}/genesis/PRESENCE_RUNTIME_GENESIS_TEMPLATE_001.json" \
  "${deployment_descriptor}" \
  "${run_dir}/secrets/host/locality-authority.public.json" \
  "${genesis_path}" >/dev/null
"${vm_binary}" --check-genesis "${genesis_path}" >"${run_dir}/genesis-validation.txt"

mkdir -m 700 "${run_dir}/plugins" "${run_dir}/network-root" "${run_dir}/evidence"
cp "${vm_binary}" "${run_dir}/plugins/${vm_id}"
chmod 700 "${run_dir}/plugins/${vm_id}"

"${root_dir}/build/locality-rehearsal" start \
  --avalanchego "${avalanchego_binary}" \
  --plugin-dir "${run_dir}/plugins" \
  --genesis "${genesis_path}" \
  --root "${run_dir}/network-root" \
  --vm-id "${vm_id}" >"${network_summary}"

network_dir="$("${python_binary}" -c 'import json,sys; print(json.load(open(sys.argv[1], encoding="utf-8"))["network_dir"])' "${network_summary}")"

"${python_binary}" "${root_dir}/rehearsal/run_lifecycle.py" run \
  --summary "${network_summary}" \
  --vm "${vm_binary}" \
  --host-public "${run_dir}/secrets/host/locality-authority.public.json" \
  --host-private "${run_dir}/secrets/host/locality-authority.private.json" \
  --participant-public "${run_dir}/secrets/participant/locality-authority.public.json" \
  --participant-private "${run_dir}/secrets/participant/locality-authority.private.json" \
  --output "${run_dir}/evidence/lifecycle.json"

"${root_dir}/build/locality-rehearsal" restart \
  --network-dir "${network_dir}" >"${run_dir}/network-restart.json"
"${python_binary}" "${root_dir}/rehearsal/run_lifecycle.py" snapshot \
  --summary "${run_dir}/network-restart.json" \
  --expected "${run_dir}/evidence/lifecycle.json" \
  --label "ALL_NODES_RESTARTED" \
  --output "${run_dir}/evidence/post-restart.json"

"${root_dir}/build/locality-rehearsal" bounce-one \
  --network-dir "${network_dir}" >"${run_dir}/network-bounce.json"
"${python_binary}" "${root_dir}/rehearsal/run_lifecycle.py" snapshot \
  --summary "${run_dir}/network-bounce.json" \
  --expected "${run_dir}/evidence/lifecycle.json" \
  --label "ONE_NODE_RESTARTED_AND_REJOINED" \
  --output "${run_dir}/evidence/post-bounce.json"

"${python_binary}" "${root_dir}/rehearsal/run_lifecycle.py" assemble \
  --lifecycle "${run_dir}/evidence/lifecycle.json" \
  --restart "${run_dir}/evidence/post-restart.json" \
  --bounce "${run_dir}/evidence/post-bounce.json" \
  --vm "${vm_binary}" \
  --avalanchego "${avalanchego_binary}" \
  --output "${run_dir}/evidence/PRESENCE_REHEARSAL_003_EVIDENCE.json"

"${root_dir}/build/locality-rehearsal" stop --network-dir "${network_dir}" >"${run_dir}/network-stop.json"
network_dir=""

cp "${run_dir}/evidence/PRESENCE_REHEARSAL_003_EVIDENCE.json" "${public_dir}/"
cp "${run_dir}/evidence/lifecycle.json" "${public_dir}/PRESENCE_REHEARSAL_003_LIFECYCLE.json"
cp "${run_dir}/evidence/post-restart.json" "${public_dir}/PRESENCE_REHEARSAL_003_POST_RESTART.json"
cp "${run_dir}/evidence/post-bounce.json" "${public_dir}/PRESENCE_REHEARSAL_003_POST_REJOIN.json"
cp "${run_dir}/genesis-validation.txt" "${public_dir}/"
(
  cd "${public_dir}"
  shasum -a 256 ./* >SHA256SUMS
)

success=1
if [[ "${LOCALITY_KEEP_PRIVATE_RUN:-0}" != "1" ]]; then
  remove_private_run
fi
trap - EXIT INT TERM
echo "PRESENCE REHEARSAL 001 PASS"
echo "public evidence: ${public_dir}"
