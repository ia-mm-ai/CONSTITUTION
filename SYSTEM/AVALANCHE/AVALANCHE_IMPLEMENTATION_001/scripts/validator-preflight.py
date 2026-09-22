#!/usr/bin/env python3
"""Read-only, offline validator binary preflight; never creates a chain or keys."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess


def checked_binary(path, expected):
    path = Path(path).resolve()
    if not re.fullmatch(r"[0-9a-f]{64}", expected):
        raise ValueError("expected binary SHA-256 must be independently supplied")
    if not path.is_file() or not os.access(path, os.X_OK):
        raise ValueError(f"missing executable: {path}")
    if hashlib.sha256(path.read_bytes()).hexdigest() != expected:
        raise ValueError(f"binary digest mismatch: {path}")
    return path


def version(path):
    return json.loads(subprocess.check_output(
        [str(path), "--version-json"], text=True, timeout=30
    ))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--plugin", required=True)
    parser.add_argument("--plugin-sha256", required=True)
    parser.add_argument("--avalanchego", required=True)
    parser.add_argument("--avalanchego-sha256", required=True)
    parser.add_argument("--vm-id", required=True)
    args = parser.parse_args()
    plugin = checked_binary(args.plugin, args.plugin_sha256)
    node = checked_binary(args.avalanchego, args.avalanchego_sha256)
    if plugin.name != args.vm_id:
        raise ValueError("plugin filename must be the official VM ID on every validator")
    vm = version(plugin)
    upstream = version(node)
    if (
        vm.get("vm_id") != args.vm_id
        or vm.get("rpcchainvm") != 46
        or vm.get("version") != "1.0.0"
        or vm.get("implementation") != "AVALANCHE_IMPLEMENTATION_001"
        or vm.get("protocol") != "PRESENCE_AVALANCHE_VM_001"
    ):
        raise ValueError("plugin VM identity or protocol mismatch")
    if (
        upstream.get("application") != "avalanchego/1.15.0"
        or upstream.get("rpcchainvm") != 46
        or upstream.get("go") != "1.25.13"
    ):
        raise ValueError("AvalancheGo compatibility mismatch")
    for binary in (plugin, node):
        info = subprocess.check_output(
            ["go", "version", "-m", str(binary)], text=True, timeout=30
        )
        if "go1.25.13" not in info or "\tdep\tgoogle.golang.org/grpc\tv1.83.2\t" not in info:
            raise ValueError(f"toolchain/dependency profile mismatch: {binary}")
    print(json.dumps({
        "status": "OFFLINE_BINARY_PREFLIGHT_PASSED",
        "vm_id": args.vm_id,
        "unperformed": [
            "VALIDATOR_IDENTITY_AND_CUSTODY",
            "PER_VALIDATOR_PLUGIN_INSTALLATION_AND_RPC_HANDSHAKE",
            "NETWORK_MEMBERSHIP_AND_VALIDATOR_MANAGER",
            "RESTART_REJOIN_AND_PUBLIC_NETWORK_QUALIFICATION",
        ],
        "effect": "NOT_RELEASE_OR_DEPLOYMENT_AUTHORIZATION",
    }, indent=2))


if __name__ == "__main__":
    main()
