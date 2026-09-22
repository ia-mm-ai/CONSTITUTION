#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

"${root}/scripts/verify-bindings.py"
python3 "${root}/scripts/verify-binding.py"
python3 -m unittest discover -s "${root}/scripts" -p 'test_*.py'
python3 -m json.tool "${root}/vm/protocol/operations.json" >/dev/null
python3 -m json.tool "${root}/vm/protocol/binding.json" >/dev/null
python3 -m json.tool "${root}/vm/genesis/PRESENCE_AVALANCHE_RUNTIME_GENESIS_TEMPLATE_001.json" >/dev/null
python3 -m json.tool "${root}/vm/genesis/genesis.schema.json" >/dev/null
python3 -m json.tool "${root}/vm/genesis/administrative-input.schema.json" >/dev/null

(
  cd "${root}/vm"
  go test ./...
  go vet ./...
)
(
  cd "${root}/vm/rehearsal"
  go test ./...
  go vet ./...
)
(
  cd "${root}/field"
  npm test
)
