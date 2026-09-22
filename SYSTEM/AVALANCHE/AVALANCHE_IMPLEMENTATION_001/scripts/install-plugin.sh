#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 3 ]]; then
  echo "usage: $0 VM_ID PLUGIN_DIRECTORY LOCALITY_VM_BINARY" >&2
  exit 2
fi

vm_id="$1"
plugin_directory="$2"
source_binary="$3"

if [[ ! "${vm_id}" =~ ^[1-9A-HJ-NP-Za-km-z]+$ ]]; then
  echo "invalid Avalanche VM ID: ${vm_id}" >&2
  exit 1
fi
if [[ ! -f "${source_binary}" || ! -x "${source_binary}" ]]; then
  echo "VM binary is missing or not executable: ${source_binary}" >&2
  exit 1
fi
if [[ ! -d "${plugin_directory}" ]]; then
  echo "plugin directory does not exist: ${plugin_directory}" >&2
  exit 1
fi

target="${plugin_directory}/${vm_id}"
if [[ -e "${target}" ]]; then
  echo "refusing to overwrite existing plugin: ${target}" >&2
  exit 1
fi

install -m 0755 "${source_binary}" "${target}"
"${target}" --version
echo "installed PRESENCE AVALANCHE VM plugin at ${target}"
