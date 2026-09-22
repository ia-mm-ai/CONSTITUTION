#!/usr/bin/env bash
set -euo pipefail
if [[ $# -ne 4 ]]; then
  echo "usage: bash network.sh AVALANCHEGO_BINARY VM_BINARY REHEARSAL_BINARY PUBLIC_RESULT_JSON" >&2
  exit 2
fi
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
node_binary="$(realpath "$1")"
vm_binary="$(realpath "$2")"
runner="$(realpath "$3")"
result="$(realpath -m "$4")"
work="$(mktemp -d /tmp/presence-validator-qualification.XXXXXX)"
chmod 700 "$work"
network_dir=""
cleanup() {
  if [[ -n "$network_dir" ]]; then
    "$runner" stop --network-dir "$network_dir" >/dev/null 2>&1 || true
  fi
  rm -rf "$work"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
node_build_info="$(go version -m "$node_binary")"
if [[ "$node_build_info" != *"go1.25.13"* || "$node_build_info" != *$'\tdep\tgoogle.golang.org/grpc\tv1.83.2\t'* ]]; then
  echo "AvalancheGo binary does not satisfy PRESENCE_AVALANCHE_SECURITY_OVERLAY_001" >&2
  exit 1
fi
mkdir -p "$work/plugins" "$work/networks"
vm_id="$("$vm_binary" --vm-id)"
cp "$vm_binary" "$work/plugins/$vm_id"
"$vm_binary" --generate-authority "$work/authority" >/dev/null
"$vm_binary" --generate-authority "$work/participant" >/dev/null
python3 - "$work/administrative.json" <<'PY'
import hashlib, json, sys
with open(sys.argv[1], "w") as target:
    json.dump({
        "schema": "PRESENCE_AVALANCHE_DEPLOYMENT_DESCRIPTOR_001",
        "genesis_id": "PRESENCE-LOCAL-QUALIFICATION-GENESIS",
        "locality_id": "PRESENCE-LOCAL-QUALIFICATION",
        "purpose": "Disposable local implementation qualification only",
        "source_reference": "AVALANCHE_IMPLEMENTATION_001",
        "source_sha256": hashlib.sha256(b"disposable local source test").hexdigest(),
    }, target)
PY
"$vm_binary" --materialize-genesis \
  "$root/vm/genesis/PRESENCE_AVALANCHE_RUNTIME_GENESIS_TEMPLATE_001.json" \
  "$work/administrative.json" "$work/authority/locality-authority.public.json" "$work/genesis.json" >/dev/null
"$runner" start --avalanchego "$node_binary" --plugin-dir "$work/plugins" \
  --genesis "$work/genesis.json" --root "$work/networks" --vm-id "$vm_id" >"$work/network.json"
network_dir="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["network_dir"])' "$work/network.json")"
node "$root/integration/network-driver.mjs" enact "$work/network.json" "$work"
"$runner" bounce-one --network-dir "$network_dir" >"$work/network.json"
node "$root/integration/network-driver.mjs" verify "$work/network.json" "$work"
"$runner" restart --network-dir "$network_dir" >"$work/network.json"
node "$root/integration/network-driver.mjs" verify "$work/network.json" "$work"
python3 - "$work/result.json" "$result" <<'PY'
import json, sys
with open(sys.argv[1]) as source:
    result = json.load(source)
result["validator_bounce_and_rejoin"] = "PASSED_AT_DECLARED_LOCAL_SCOPE"
result["full_network_process_restart"] = "PASSED_AT_DECLARED_LOCAL_SCOPE"
result["avalanchego_profile"] = "v1.15.0+PRESENCE_AVALANCHE_SECURITY_OVERLAY_001"
result["disposable_credentials"] = "PRIVATE_TEMPORARY_DIRECTORY_REMOVED_BY_EXIT_TRAP"
with open(sys.argv[2], "w") as target:
    json.dump(result, target, indent=2)
    target.write("\n")
PY
