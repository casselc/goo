# Project direction

The user supplied the Grammar of Observability chat handoff to start this repository.
The final implementation should be **Rust or Go**. Goo now implements the product runtime and semantic libraries in Go. The imported
Python server is retired; Python remains only in independent sample workloads and
offline tooling/reference tests. See [runtime decision](docs/SINGLE_GO_RUNTIME.md).
Clojure/EDN suggestions in the historical handoff are background ideas, not a
requirement to migrate to Clojure. The attached continuation prompt is a proposed
backlog; importing it does not mean all its milestones have been implemented.

## Related work to inspect before designing the core

- `../observability/otel-chdb/query`: Go query service, ClickHouse SQL parsing,
  scoped access, bounded queries, completeness/late-data labels, and repeatable
  query bases. Its README explains the choice of Go and the ClickHouse parser.
- `../observability/otel-chdb/alerts`: evaluation and correction semantics over
  query results; useful when moving beyond on-demand alarm previews.
- `../observability/otel-chdb/otap-rs`: Rust telemetry ingestion, storage, and
  deterministic/property tests. It has a pinned external upstream dependency;
  it is not a small standalone grammar library.
- `../observability/otel-chdb/lakeui`: query planning, evidence completeness, and
  browser presentation over Parquet; distinguish its storage protocol from
  the measurement language proposed here.
- `../oscope`: OTLP collection, embedded chDB, query/export/viewer boundaries,
  explicit lifecycle/configuration, and integration tests. Reuse contracts and
  lessons where appropriate; this project's implementation language remains Go
  or Rust rather than inheriting oscope's Jolt runtime.

Keep measurement semantics separate from physical source adapters and UI code.
First establish deterministic examples and reference/ClickHouse parity, then
choose integration boundaries with the existing repositories. Existing work
should inform that decision before duplicating query or telemetry infrastructure.

## Current development environment

Python baseline: `/workspace/.venvs/goo/bin/python`.
Run from `/workspace/goo`:

```sh
/workspace/.venvs/goo/bin/python -m unittest discover -s tests -v
/workspace/.venvs/goo/bin/python tests/local_flow.py
```

Cloud startup and image-build workarounds are saved in the environment setup
configuration. The imported SHA256SUMS describes the original archive only;
this direction note and subsequent validation evidence are additions.

## Declarative authoring direction

The user prefers an IaC grammar along the lines of **YAML with CEL expressions**,
**KCL**, or a comparable declarative language. This is a design preference, not
an already selected parser or an implemented API. Keep Go/Rust runtime choice
separate from authoring syntax. Normalize the selected format into a versioned,
typed measurement model shared by managed resources and interactive queries.
Evaluate validation/type checking, bounded expression execution, deterministic
time and units, tooling, and Go/Rust embedding before selecting a format. Do not
assume historical EDN examples or the prototype's JSON transport are the final
IaC grammar.

As development proceeds, the user requests executable storyboards of real sample
app and investigation-UI journeys, using Playwright or equivalent tests for
documentation. Semantic changes should use hegel-go properties and state/swarm
tests where appropriate, with Quint models and the sibling quint-go tooling for
model-based replay and observed-trace validation. Current implementations and
commands are linked from docs/journeys/README.md and verification/README.md.

## Long-term platform compatibility

The user wants definitions to compile/translate cleanly into Databricks Unity
Catalog metric views so ClickStack data can move into a longer-term platform.
Treat Databricks as an explicit compiler target. Preserve a portable typed model,
emit dimensions/measures as metric-view YAML, and lower deduplication, deadline
classification, evidence quality, and corrections into declared preparation
artifacts where required. Data-ingestion compatibility is a separate contract.
See [the research and proposed compatibility contract](docs/DATABRICKS_METRIC_VIEWS.md)
and [illustrative target SQL](examples/databricks/reservation_metrics.sql).

A [Go aggregate compiler pilot](docs/PORTABLE_COMPILER.md) now implements the
prepared-fact count/ratio subset, with a reference evaluator, generated ClickHouse
SQL, and generated Databricks YAML/DDL. Go was chosen for this bounded slice to
reuse the installed toolchain and sibling query-service ecosystem. It leaves the
Python lab intact and does not settle CEL versus KCL, implement raw lifecycle
preparation, or certify execution on Databricks.

## Reusable semantic packages

The user emphasizes goo's overall purpose beyond the demo: OTel conventions
should drive reusable measures, relations, views, recipes and export plans. The
Go prebuilt registry is a contract-recognition foundation. An initial HTTP
server-span binding/expansion now emits shared aggregate/view/recipe artifacts
and prepares normalized requests; general package resolution and raw-source
integration remain future work. The sample queue telemetry is one explicitly custom adapter.
See docs/SEMANTIC_PACKAGES.md for implementation boundaries and the next layer.

The HTTP package now has a bounded real ClickHouse source adapter and a
descriptor-driven aggregate UI. Instrumentation provenance remains explicit
(schema URLs are dropped by the pinned exporter). Orchestrion-style integration
can supply library coverage plus custom Quint-observable semantic seams; keep
its tooling isolated from the runtime and distinguish build declarations from
observed conformance. See docs/LIVE_HTTP_PACKAGE.md and docs/INSTRUMENTATION.md.

The user approved operation/event annotations as observation inputs, with aggregate
policy retained in packages. A first bounded dialect, source-contract manifest,
OTel runtime hooks and test-only weaving command now exist; source declarations
must not be treated as proof of linked-build coverage or runtime completeness.
See docs/GOO_ANNOTATIONS.md.

A separate woven Go operation sample now exports real spans through the collector
to ClickHouse. The Go source adapter and investigation overview compare completed
invocations to the source contract while retaining explicit evidence gaps.
See docs/LIVE_OPERATIONS.md. Counts remain invocations, not business subjects.

Scoped subject navigation now connects declared annotation events across captured
operation traces using environment/service/namespace/tenant/ID. It is an
inspectable relation, not a durable lifecycle or retry classifier.
See docs/SUBJECT_NAVIGATION.md.

A strict YAML operation-package slice now shares bindings, source-contract
validation, identity, enabled views and evidence limits between the CLI and UI.
A separate report application exercises reuse without application-specific query
or UI branches. This does not settle the full IaC language or implement CEL/KCL.
See docs/OPERATION_PACKAGES.md.

Native counter interval rates now have bounded Go semantics, YAML declarations
and a live ClickHouse-backed operation view. Unknown reset/gap/boundary intervals
remain explicit; native metrics and span context retain separate evidence bases.
See docs/NATIVE_COUNTERS.md and the requested docs/ARCHITECTURE_DIAGRAMS.md.
