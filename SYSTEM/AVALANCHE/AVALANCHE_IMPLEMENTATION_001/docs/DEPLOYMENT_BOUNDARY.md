# Deployment boundary

This tree is reviewable implementation source only. `scripts/build.sh` creates
a local native VM binary; it does not install it, create a chain or L1, register
a Validator Manager, create an Authority, deploy, release, tag, or mutate a
public network.

`PRESENCE_AVALANCHE_SECURITY_OVERLAY_001` remains required for an AvalancheGo
v1.15.0 node. A run against plain v1.15.0 is base-version evidence only and
does not qualify the overlay. Network scripts accept only loopback disposable
networks and remove private credentials and network state on exit.

No private key, mnemonic, credential, generated binary, live network directory,
deployment-specific genesis, or production validator configuration belongs in
this repository.
