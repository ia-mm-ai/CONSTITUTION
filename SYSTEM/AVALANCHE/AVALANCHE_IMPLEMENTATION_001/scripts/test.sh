#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
repository="$(cd "${root}/../../.." && pwd)"

python3 "${root}/scripts/verify-binding.py"
python3 "${root}/scripts/verify-coherence.py"
python3 -m unittest discover -s "${root}/scripts" -p 'test_*.py'
python3 "${repository}/SYSTEM/verify.py" --revision 5e09d1fbe3c1008937d95497fcc162a6ebd4190d
python3 -m unittest discover -s "${repository}/SYSTEM" -p 'test_*.py'
python3 -m json.tool "${root}/vm/protocol/operations.json" >/dev/null
python3 -m json.tool "${root}/vm/genesis/PRESENCE_AVALANCHE_RUNTIME_GENESIS_TEMPLATE_001.json" >/dev/null
python3 -m json.tool "${root}/vm/genesis/genesis.schema.json" >/dev/null
python3 -m json.tool "${root}/vm/genesis/administrative-input.schema.json" >/dev/null

(
  cd "${root}/vm"
  go test ./...
  go test -race ./...
  go vet ./...
  go mod verify
)
(
  cd "${root}/vm/rehearsal"
  go test ./...
  go vet ./...
  go mod verify
)
(
  cd "${root}/field"
  npm test
)
"${root}/integration/run.sh"
