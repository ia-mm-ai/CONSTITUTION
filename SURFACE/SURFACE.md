# SURFACE

## Derivative exposure

This is the first editable exposure contract, not a new constitutional layer,
an adopted Body, or a claim that a POINT is already deployed. The public
location is explicitly **UNCONFIGURED**. The repository's historical coordinate
in [ORIGIN](../STATE/LINEAGE/ORIGIN.json) is provenance, not an invented service
URL or a currently published revision.

[surface.json](surface.json) has three sections:

- **point** names the entrance, its bounded exposure scope, unestablished public
  location, and direct provenance references into STATE/LINEAGE.
- **resources** explicitly selects the Source human/machine pair; STATE and
  its three modules; the existing schemas and vectors; and the origin and two
  extraction records. Paths are relative to this declaration. Existing core
  IDs are reused exactly. Resources without an ID receive derivative
  `urn:presence:surface:…:1` locators, not new core definitions.
- **operations** names the three actual handlers in `serve.py`, their GET
  routes, closed bounded query contracts, outputs, and effect ceilings.
  The registry is fixed: changing a string cannot import or execute a new handler.

Schemas and core definitions are not copied into an interface-specific model.
Their selected exact files remain available for offline reference resolution.
Stable constitutional refs such as `reading.7` and `section.19.2` resolve to a
JSON Pointer in the bound machine Constitution. References locate material;
they do not import its truth, support, permission, or constitutional effect.
The origin's inspected snapshot is distinct from an edition's selected commit.

## One edition for humans and machines

`index.html` is an entrance template, not a second hand-maintained resource
list. [SYSTEM](../SYSTEM/SYSTEM.md) exports it with generated `index.json` and
`manifest.json` alongside the selected files. The HTML reads that inventory,
shows its exact revision and verification limits, and links only its listed
material. Opening this source template without a generated inventory reports
that no usable edition is available; it never supplies sample release claims.
The index and HTML bytes are bound by the same manifest as the core material.

No generated inventory or manifest is checked in as a placeholder. Their
revision and hashes can only be fixed when the selected commit exists.
Changes to this declaration take effect only in a newly exported revision,
not through a live union of a publication and the working tree.

## Discovery, reading, resolution, checks

Install the pinned requirements in `SYSTEM/requirements.txt`, then run:

```sh
python /home/runner/work/PRESENCE/PRESENCE/SURFACE/serve.py --edition /tmp/presence-edition --manifest-sha256 "$MANIFEST_SHA256"
```

The interface listens on `127.0.0.1:8765` by default. It verifies the complete
export against an independently supplied manifest digest before listening,
then serves those same checked bytes from memory. Disk changes do not silently
change the edition. This is a bounded local interface, not a hardened public
hosting stack; a public location needs a separate deployment decision.

- `GET /` is the human entrance. `GET /index.json` is machine discovery;
  `GET /manifest.json` supplies the byte bindings.
- `GET /api/read?resource=<encoded-resource-ref>` returns the exact listed
  bytes with the declared media type. Listed paths are also directly readable.
- `GET /api/resolve?ref=<encoded-ref>` returns the edition revision, requested
  reference, file path, and JSON Pointer (empty for a whole resource).
- `GET /api/check?resource=<encoded-resource-ref>` compares those served bytes
  with their edition binding and returns SHA-256, length, and `MATCH` or
  `MISMATCH`. Its `conformance: NOT_EVALUATED` is intentional: this bounded
  operation does not run semantic rules on arbitrary submitted records.

Query values must be URL-encoded. Each call accepts exactly one nonempty
parameter of at most 512 characters. Unknown parameters and malformed requests
return 400; unlisted resources and references return 404. Unsupported methods
return 501. There are no upload, mutation, export, recovery, arbitrary path,
remote fetch, dynamic import, or command-execution endpoints.

The initial repository lacks the core checker described by STATE. An explicitly
requested incomplete preview may still be read and byte-checked; its inventory
and entrance retain `INCOMPLETE` / `UNAVAILABLE`. Serving it cannot convert
those statuses into conformance. Static hosting can expose the same files,
but it does not by itself implement the declared API handlers.

## First externalisation, not established reality

The placeholders are the deliberately open deployment location and operational
qualification, not lorem ipsum resources or fabricated provenance. Maintenance
checks do not prove the authenticity of the two historical evidence packs,
refresh DCR observations, preserve a living State through CSC, or authorize a
constitutional act. The current scope ends at reading, locating, and checking
this edition's material.
