#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
go_binary="${GO_BINARY:-go}"

if ! go_binary="$(command -v "${go_binary}")"; then
  echo "Go compiler is unavailable" >&2
  exit 1
fi
gofmt_binary="$(dirname "${go_binary}")/gofmt"
if [[ ! -x "${gofmt_binary}" ]]; then
  echo "gofmt is unavailable beside ${go_binary}" >&2
  exit 1
fi

cd "${root_dir}"
if [[ "$("${go_binary}" env GOVERSION)" != "go1.25.13" ]]; then
  echo "verification requires exact Go 1.25.13; observed $("${go_binary}" env GOVERSION)" >&2
  exit 1
fi
unformatted="$("${gofmt_binary}" -l ./*.go rehearsal/*.go)"
if [[ -n "${unformatted}" ]]; then
  echo "Go source is not gofmt-clean:" >&2
  echo "${unformatted}" >&2
  exit 1
fi
"${go_binary}" test -buildvcs=false ./...
"${go_binary}" test -buildvcs=false -race ./...
"${go_binary}" vet -buildvcs=false ./...
(
  cd "${root_dir}/rehearsal"
  GOTOOLCHAIN=local GOWORK=off "${go_binary}" test -buildvcs=false ./...
  GOTOOLCHAIN=local GOWORK=off "${go_binary}" vet -buildvcs=false ./...
)
bash "${root_dir}/scripts/build.sh"
