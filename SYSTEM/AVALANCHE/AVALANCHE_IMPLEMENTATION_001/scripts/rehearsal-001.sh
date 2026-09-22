#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "usage: $0 /absolute/path/to/avalanchego [PUBLIC_EVIDENCE_DIRECTORY]" >&2
  exit 2
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
avalanchego="$(cd "$(dirname "$1")" && pwd)/$(basename "$1")"
evidence_root="${2:-/tmp/presence-avalanche-rehearsal-evidence}"
go_binary="${GO_BINARY:-go}"

if [[ ! -x "${avalanchego}" ]]; then
  echo "AvalancheGo binary is not executable: ${avalanchego}" >&2
  exit 2
fi
if [[ -e "${evidence_root}" ]]; then
  echo "refusing to overwrite evidence directory: ${evidence_root}" >&2
  exit 2
fi
if [[ "$("${go_binary}" env GOVERSION)" != "go1.25.13" ]]; then
  echo "rehearsal requires exact Go 1.25.13" >&2
  exit 1
fi
"${avalanchego}" --version-json | python3 -c '
import json, sys
document = json.load(sys.stdin)
if (
    document.get("application") != "avalanchego/1.15.0"
    or document.get("rpcchainvm") != 46
    or document.get("go") != "1.25.13"
):
    raise SystemExit(f"incompatible AvalancheGo: {document}")
'

scratch="$(mktemp -d /tmp/presence-avalanche-rehearsal-001.XXXXXX)"
network_dir=""
cleanup() {
  status=$?
  trap - EXIT INT TERM
  if [[ -n "${network_dir}" && -d "${network_dir}" ]]; then
    "${scratch}/bin/rehearsal" stop --network-dir "${network_dir}" >/dev/null 2>&1 || true
  fi
  rm -rf -- "${scratch}"
  exit "${status}"
}
trap cleanup EXIT INT TERM

mkdir -p "${scratch}/bin" "${scratch}/secrets/host" "${scratch}/secrets/participant" \
  "${scratch}/plugins" "${scratch}/network-root" "${scratch}/evidence"
mkdir "${evidence_root}"
chmod 700 "${scratch}" "${scratch}/secrets" "${scratch}/secrets/host" \
  "${scratch}/secrets/participant"

(
  cd "${root}"
  GOWORK=off "${go_binary}" test ./...
  GOWORK=off "${go_binary}" test -race ./...
  GOWORK=off "${go_binary}" vet ./...
  GOWORK=off "${go_binary}" build -buildvcs=false -o "${scratch}/bin/presence-avalanche-vm" .
)
(
  cd "${root}/rehearsal"
  GOWORK=off "${go_binary}" test ./...
  GOWORK=off "${go_binary}" vet ./...
  GOWORK=off "${go_binary}" build -buildvcs=false -o "${scratch}/bin/rehearsal" .
)

vm="${scratch}/bin/presence-avalanche-vm"
vm_id="$("${vm}" --vm-id)"
"${vm}" --generate-authority "${scratch}/secrets/host" >/dev/null
"${vm}" --generate-authority "${scratch}/secrets/participant" >/dev/null

cat >"${scratch}/admin.json" <<'JSON'
{
  "schema": "PRESENCE_AVALANCHE_ADMIN_INPUT_001",
  "genesis_id": "PRESENCE-AVALANCHE-REHEARSAL-GENESIS-001",
  "locality_id": "PRESENCE-AVALANCHE-REHEARSAL-HOST-001",
  "purpose": "Exercise the repaired operation-scope lifecycle on a disposable private three-validator network.",
  "source_reference": "PRESENCE_AVALANCHE_IMPLEMENTATION_001_REHEARSAL_SOURCE",
  "source_sha256": "affeb5738cdfeea7ee4fe985652bf78b15eb6dfbba5871d82f0f2213a81832d0"
}
JSON
"${vm}" --materialize-genesis \
  "${root}/genesis/PRESENCE_AVALANCHE_GENESIS_TEMPLATE_001.json" \
  "${scratch}/admin.json" \
  "${scratch}/secrets/host/presence-avalanche-authority.public.json" \
  "${scratch}/genesis.json" >/dev/null
"${vm}" --check-genesis "${scratch}/genesis.json" >"${scratch}/evidence/genesis-validation.txt"

cp "${vm}" "${scratch}/plugins/${vm_id}"
chmod 700 "${scratch}/plugins/${vm_id}"
"${scratch}/bin/rehearsal" start \
  --avalanchego "${avalanchego}" \
  --plugin-dir "${scratch}/plugins" \
  --genesis "${scratch}/genesis.json" \
  --root "${scratch}/network-root" \
  --vm-id "${vm_id}" >"${scratch}/network-start.json"
network_dir="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["network_dir"])' "${scratch}/network-start.json")"

python3 "${root}/rehearsal/run_lifecycle.py" run \
  --summary "${scratch}/network-start.json" \
  --vm "${vm}" \
  --host-public "${scratch}/secrets/host/presence-avalanche-authority.public.json" \
  --host-private "${scratch}/secrets/host/presence-avalanche-authority.private.json" \
  --participant-public "${scratch}/secrets/participant/presence-avalanche-authority.public.json" \
  --participant-private "${scratch}/secrets/participant/presence-avalanche-authority.private.json" \
  --output "${scratch}/evidence/lifecycle.json"

"${scratch}/bin/rehearsal" restart \
  --network-dir "${network_dir}" >"${scratch}/network-restart.json"
python3 "${root}/rehearsal/run_lifecycle.py" snapshot \
  --summary "${scratch}/network-restart.json" \
  --expected "${scratch}/evidence/lifecycle.json" \
  --label ALL_NODES_RESTARTED \
  --output "${scratch}/evidence/post-restart.json"

"${scratch}/bin/rehearsal" bounce-one \
  --network-dir "${network_dir}" >"${scratch}/network-bounce.json"
python3 "${root}/rehearsal/run_lifecycle.py" snapshot \
  --summary "${scratch}/network-bounce.json" \
  --expected "${scratch}/evidence/lifecycle.json" \
  --label ONE_NODE_RESTARTED_AND_REJOINED \
  --output "${scratch}/evidence/post-rejoin.json"

python3 "${root}/rehearsal/run_lifecycle.py" assemble \
  --lifecycle "${scratch}/evidence/lifecycle.json" \
  --restart "${scratch}/evidence/post-restart.json" \
  --bounce "${scratch}/evidence/post-rejoin.json" \
  --vm "${vm}" \
  --avalanchego "${avalanchego}" \
  --output "${scratch}/evidence/PRESENCE_AVALANCHE_REHEARSAL_001.json"

"${scratch}/bin/rehearsal" stop --network-dir "${network_dir}" >/dev/null
network_dir=""
cp "${scratch}/evidence/PRESENCE_AVALANCHE_REHEARSAL_001.json" "${evidence_root}/"
cp "${scratch}/evidence/genesis-validation.txt" "${evidence_root}/"
(
  cd "${evidence_root}"
  sha256sum PRESENCE_AVALANCHE_REHEARSAL_001.json genesis-validation.txt >SHA256SUMS
)
echo "PRESENCE Avalanche private three-validator rehearsal: PASS"
echo "public evidence: ${evidence_root}"
