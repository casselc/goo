# Compiler-assisted instrumentation for goo

Orchestrion-style instrumentation is useful at two distinct layers of goo.

## Library coverage and package discovery

Build-time integration with HTTP, gRPC, database and messaging libraries can emit
standard signals without application authors repeating instrumentation code. A
build manifest can propose the corresponding semantic package bindings:

- Application revision, Go module/dependency versions and instrumentation version.
- Enabled integrations, excluded packages/endpoints and configured sampling.
- Scope names/versions, semantic schema mapping and expected signal contracts.
- Required application hooks that library instrumentation cannot supply.

The manifest is a declaration. Runtime source inventory and conformance journeys
must check the actual data before goo treats a capability as present. New library
versions, disabled integrations, sampling and exporter transformations can all
invalidate a proposed binding. This is particularly relevant because the current
ClickHouse exporter drops scope schema URLs.

Datadog Orchestrion's standard integrations use Datadog tracing. Their output
needs an explicit source/semantic mapping; it must not be relabelled as OTel
convention-compatible by assumption. **Custom Orchestrion aspects** can instead
call goo/OTel hooks directly. These are opt-in weaving choices, not requirements
that goo adopt Datadog's tracing or backend.

## Model-observable semantic seams

The sibling `observability/quintgo/qobs/orchestrion.yml` already implements custom
`//quint:action` and `//quint:thread` aspects. They record observed steps on function
return, with adapters for memory, JSONL and OTel sinks. Importing those aspects
alone does not enable dd-trace library integrations.

Goo now annotates the actual `prebuilt.Expansion.PrepareHTTP` function. A successful
return records selected-request count and prepared-fact count. This is the real
preparer, not a stand-in implementation. An isolated `verification/weaving`
module enables only the quintgo aspects and checks the emitted trace against a
Quint invariant: successful preparation retains every selected request, including
unknown outcomes. A deliberate dropped-fact trace is the negative control.

Plain `go build` ignores the comment and imports no recorder. The optional woven
build adds recorder calls through the aspect. The pilot tests model observation;
it does not yet send woven records to the running collector, weave HTTP library
instrumentation into the sample, or implement build-manifest generation.

Run from `verification/weaving` with the repository's Node/Quint and Go toolchains:

```sh
QUINTGO_BACKEND=typescript \
  go run github.com/DataDog/orchestrion go test ./scenario -v
```

Validated here with Orchestrion 1.13.1 and an isolated `golang.org/x/tools`
0.52.0 override. Orchestrion's selected 0.49.0 importer could not decode Go 1.27's
export-data version 5; 0.52.0 explicitly supports that format and the woven test
passes. No runtime dependency was upgraded for this compatibility fix. An initial
large module download was denied at Go's storage backend; fetching that exact
module from its official GitHub source with Go checksum verification succeeded.

Keep the repository `node_modules/.bin` on PATH so quintgo can find Quint.
`QUINTGO_KEEP` retains native step logs, model-level ITF and generated validation
artifacts. The test intentionally fails on an unwoven build with no steps; it
does not turn missing instrumentation into a passing skipped check.

The weaving module contains the Orchestrion tool dependency; the root runtime
and existing native Hegel module remain separate. Tests do not run concurrently
because qtest uses a global recorder. Weaving/model checks are optional developer
checks, not part of every production HTTP request.

## Goo operation/event slice

[GOO_ANNOTATIONS.md](GOO_ANNOTATIONS.md) now documents the implemented operation
and event aspects, source-contract manifests and `goo weave` test wrapper. These
emit custom OTel observations and diagnose event drops. They complement the
Quint observation pilot; they do not implement automatic library instrumentation
or a full linked-binary manifest.

## Useful next applications

1. Generate and validate the build manifest, then use it as a proposed binding
   input alongside observed inventory. Preserve the distinction between enabled
   instrumentation and confirmed runtime evidence.
2. Record acceptance, terminal decision and correction-publication seams through
   the same aspect mechanism, using the existing lifecycle/Quint contracts.
3. Emit model observations through the qobs OTel sink when wanted for investigation,
   with explicit sampling and retention independent of request-span sampling.
4. Add generated domain-hook scaffolding for business identity and outcomes.
   Automatically weaving a function return cannot establish that a business
   transaction committed or that an external side effect completed.

Compiler injection supplies observation, not causality, transactional guarantees,
or a business SLO. Those remain explicit goo semantics and evidence contracts.
