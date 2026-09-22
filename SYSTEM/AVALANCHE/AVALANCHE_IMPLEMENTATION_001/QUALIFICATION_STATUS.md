# Qualification status

Observed on 2026-09-22 UTC.

## Status set

- `SOURCE_IMPLEMENTED`
- `FUNCTIONALLY_TESTED`
- `QUALIFICATION_INCOMPLETE`
- `NOT_RELEASED`
- `NOT_DEPLOYED`
- `NOT_READY_FOR_MAINNET`

`READY_FOR_ADMIN_INSTANTIATION` is not asserted.

## Passed observations

- Source, STATE, and both predecessor archive hashes verified against
  `bindings/SOURCE_STATE_BINDING_001.json`.
- Exact Go `1.25.13` VM tests, race tests, vet, native build, and version output
  passed.
- FIELD loaded and all 25 Node tests passed; its package declares no third-party
  runtime or development dependencies.
- Genesis authority generation, administrative schema validation,
  materialization, genesis schema validation, and executable genesis checking
  passed using disposable `/tmp` credentials, which were removed.
- Pinned AvalancheGo `v1.15.0`, commit
  `70bd6d063b7343fd2cd8217200aaf77b57f19f68`, Go `1.25.13`, and RPCChainVM
  protocol `46` built with the declared security overlay.
- A disposable private three-validator network accepted 18 transitions. Its
  first crossing was `BOUND → DECLARE_CAPACITY` with an empty capacity
  `locus_id`, followed by an equal renewed read from all three nodes.
- The lifecycle continued through pulse, presentation, admission, entry,
  observation, matter disposition, emergence, correction, checkpoint,
  exit/close, a second locus, reentry, and addressed-residue incorporation.
- Full-network restart replay and single-validator stop/restart/rejoin both
  preserved the exact final revision, height, state bytes, and commitment.
- No private authority material was exported or persisted in the repository.
- No public-network mutation occurred.
- A GitHub advisory database query for direct dependency AvalancheGo `v1.15.0`
  returned no known advisory. A temporary-lockfile `npm audit --omit=dev` of
  the dependency-free FIELD package reported zero vulnerabilities.
- CodeQL analyzed the Go, Python, and JavaScript changes and reported zero
  alerts.

The compact observed rehearsal result is
`qualification/REHEARSAL_OBSERVATION_001.json`.

## Unresolved external observations

Vulnerability status is `UNRESOLVED_EXTERNAL_OBSERVATION`.

Both of these exact `govulncheck v1.7.0` queries failed while fetching
`https://vuln.go.dev/index/modules.json.gz` because DNS resolution through
`127.0.0.53:53` returned `server misbehaving`:

```text
go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./main
```

The first query targeted this VM module; the second targeted the exact
security-overlaid AvalancheGo source. Therefore no claim of zero Go
vulnerabilities is made.

The automated code-review backend was also unavailable. Its exact failure was
that model `capi-prod-claude-sonnet-4.6` was not found in the configured model
registry. No review findings were returned, and successful CodeQL analysis is
not represented as a substitute for that review.

Independent cryptographic/Avalanche consensus review, public test-network
rehearsal, production credential restoration, adversarial partition and
resource testing, deployment operations, release qualification, and Mainnet
readiness remain unperformed. None is implied by the local source crossing.
