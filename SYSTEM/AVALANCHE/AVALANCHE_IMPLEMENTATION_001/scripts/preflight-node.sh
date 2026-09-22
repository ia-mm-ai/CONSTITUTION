#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 8 ]]; then
  echo "usage: $0 API_URL VM_ID PLUGIN_DIRECTORY EXPECTED_VM_SHA256 NODE_KEY_DIRECTORY EXPECTED_NODE_ID AVALANCHEGO_BINARY EXPECTED_AVALANCHEGO_SHA256" >&2
  exit 2
fi

api_url="${1%/}"
vm_id="$2"
plugin_directory="$3"
expected_vm_sha256="$4"
node_key_directory="$5"
expected_node_id="$6"
avalanchego_binary="$7"
expected_avalanchego_sha256="$8"
plugin_path="${plugin_directory}/${vm_id}"

for command_name in curl file go sha256sum; do
  if ! command -v "${command_name}" >/dev/null 2>&1; then
    echo "required command is unavailable: ${command_name}" >&2
    exit 1
  fi
done

if [[ ! -x "${plugin_path}" ]]; then
  echo "LOCALITY plugin is missing or not executable: ${plugin_path}" >&2
  exit 1
fi
for digest in "${expected_vm_sha256}" "${expected_avalanchego_sha256}"; do
  if [[ ! "${digest}" =~ ^[0-9a-f]{64}$ ]]; then
    echo "expected binary digests must be 64 lowercase hexadecimal characters" >&2
    exit 1
  fi
done
if [[ ! -x "${avalanchego_binary}" ]]; then
  echo "AvalancheGo binary is missing or not executable: ${avalanchego_binary}" >&2
  exit 1
fi

observed_vm_sha256="$(sha256sum "${plugin_path}" | awk '{print $1}')"
observed_avalanchego_sha256="$(sha256sum "${avalanchego_binary}" | awk '{print $1}')"
if [[ "${observed_vm_sha256}" != "${expected_vm_sha256}" ]]; then
  echo "LOCALITY VM binary digest mismatch" >&2
  exit 1
fi
if [[ "${observed_avalanchego_sha256}" != "${expected_avalanchego_sha256}" ]]; then
  echo "AvalancheGo binary digest mismatch" >&2
  exit 1
fi

plugin_version="$(${plugin_path} --version)"
if [[ "${plugin_version}" != *"avalanchego-profile=v1.15.0+LOCALITY_SECURITY_OVERLAY_001"* || "${plugin_version}" != *"rpcchainvm-protocol=46"* ]]; then
  echo "unexpected LOCALITY plugin compatibility declaration: ${plugin_version}" >&2
  exit 1
fi

for binary in "${plugin_path}" "${avalanchego_binary}"; do
  build_info="$(go version -m "${binary}")"
  if [[ "${build_info}" != *"go1.25.13"* || "${build_info}" != *$'\tdep\tgoogle.golang.org/grpc\tv1.83.2\t'* ]]; then
    echo "binary does not satisfy the required Go/gRPC security profile: ${binary}" >&2
    exit 1
  fi
done
avalanchego_version="$(${avalanchego_binary} --version-json)"
if [[ "${avalanchego_version}" != *'"application":"avalanchego/1.15.0"'* || "${avalanchego_version}" != *'"go":"1.25.13"'* || "${avalanchego_version}" != *'"rpcchainvm":46'* ]]; then
  echo "AvalancheGo binary does not match the qualified version contract" >&2
  exit 1
fi

for key_name in staker.crt staker.key signer.key; do
  key_path="${node_key_directory}/${key_name}"
  if [[ ! -s "${key_path}" ]]; then
    echo "required persisted validator credential is missing: ${key_path}" >&2
    exit 1
  fi
done

for private_name in staker.key signer.key; do
  private_path="${node_key_directory}/${private_name}"
  if stat -f '%Lp' "${private_path}" >/dev/null 2>&1; then
    mode="$(stat -f '%Lp' "${private_path}")"
  else
    mode="$(stat -c '%a' "${private_path}")"
  fi
  if [[ ! "${mode}" =~ ^[0-7]*[0-7]00$ ]]; then
    echo "validator private key permissions are too broad (${mode}): ${private_path}" >&2
    exit 1
  fi
done

version_response="$(curl --fail --silent --show-error -X POST "${api_url}/ext/info" \
  -H 'Content-Type: application/json' \
  --data '{"jsonrpc":"2.0","id":1,"method":"info.getNodeVersion","params":{}}')"
if [[ "${version_response}" != *'"version":"avalanchego/1.15.0"'* || "${version_response}" != *'"rpcProtocolVersion":46'* ]]; then
  echo "node is not the exact tested AvalancheGo v1.15.0 / RPC protocol 46 target" >&2
  echo "${version_response}" >&2
  exit 1
fi

identity_response="$(curl --fail --silent --show-error -X POST "${api_url}/ext/info" \
  -H 'Content-Type: application/json' \
  --data '{"jsonrpc":"2.0","id":1,"method":"info.getNodeID","params":{}}')"
if [[ "${identity_response}" != *'"nodeID":"'"${expected_node_id}"'"'* ]]; then
  echo "running node identity does not match the deliberately registered NodeID ${expected_node_id}" >&2
  echo "${identity_response}" >&2
  exit 1
fi

file "${plugin_path}"
echo "LOCALITY node preflight passed"
echo "plugin: ${plugin_version}"
echo "plugin sha256: ${observed_vm_sha256}"
echo "avalanchego sha256: ${observed_avalanchego_sha256}"
echo "node identity: ${expected_node_id}"
echo "validator credentials: present, non-empty, private permissions bounded"
