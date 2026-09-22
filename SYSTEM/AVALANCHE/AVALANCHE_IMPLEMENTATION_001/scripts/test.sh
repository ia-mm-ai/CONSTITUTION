#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

"${root}/scripts/verify-bindings.py"
python3 -m json.tool "${root}/protocol/operations.json" >/dev/null
python3 -m json.tool "${root}/genesis/PRESENCE_AVALANCHE_GENESIS_TEMPLATE_001.json" >/dev/null
python3 -m json.tool "${root}/schemas/PRESENCE_AVALANCHE_GENESIS_001.schema.json" >/dev/null
python3 -m json.tool "${root}/schemas/PRESENCE_AVALANCHE_ADMIN_INPUT_001.schema.json" >/dev/null
python3 -c 'import pathlib,sys; path=pathlib.Path(sys.argv[1]); compile(path.read_text(encoding="utf-8"), str(path), "exec")' \
  "${root}/rehearsal/run_lifecycle.py"

(
  cd "${root}"
  go test ./...
  go vet ./...
)
(
  cd "${root}/rehearsal"
  go test ./...
  go vet ./...
)
(
  cd "${root}/field"
  npm test
)
