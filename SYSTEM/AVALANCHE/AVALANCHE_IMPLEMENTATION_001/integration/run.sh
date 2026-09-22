#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
work="$root/integration/.coupled-work-$$"
mkdir -m 700 "$work"
trap 'rm -rf "$work"' EXIT HUP INT TERM
export TMPDIR="$work" GOTMPDIR="$work"
cd "$root/vm"
go test -tags=integration -run '^TestCoupledFIELDLifecycle$' -count=1 -v .
