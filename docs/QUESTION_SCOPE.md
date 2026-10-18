# Shared question scope and checked transforms

Implemented: additive planning metadata for existing recipes and capture-derived
follow-ups. This is the first slice of [the consolidation plan](CORE_PRIMITIVES.md),
not a universal query language or automatic OTel convention resolver.

`exploration.CaptureQuestion.Scope()` validates and detaches a descriptor without
a source read. Record scope contains environment, half-open fixed bounds, exact
predicate, source-adapter contract and span/log-record grain. Paging, grouping,
summary and chart settings are presentation/aggregation choices outside population
scope. Query edits do not mutate derived metadata, or vice versa.

Adapters also describe native measurements (exact declared source binding and
complete definition revision; signal is distinct from measurement kind), inventories (underlying span-record population),
and comparison pairs (two separate member scopes/windows). Comparison transforms
require explicit member handling; they do not silently collapse windows.

`DescribeTransform` validates same-population, same-question and exact trace/span
relationship claims. All supported handoffs retain environment/fixed bounds;
unapproved drift fails. Exact-ID relationships require structurally valid nonzero
reported IDs; the capture planner owns selecting those IDs from stored observations.
`DescribeDiscoveryTransform` drops record/metric-specific selectors and describes
service/time metadata discovery. It does not claim the same population or health.

The returned `QuestionTransform` contains `origin`, `target`, `relation`, structured
`changes` (field plus preserved/replaced effect), and derived `preserves`/`replaces`
labels. Grouping, summary/status, chart and page changes come from actual question
fields. No arbitrary subset/widening proof or causal relationship is inferred.
Recipe plans and each supported follow-up action carry `transform`; old follow-up
`preserves`/`changes` fields mirror these checked labels for compatibility.

## OTel boundary

Record selectors preserve resource and record-attribute namespaces, scope name/
library-version predicates, span-kind filters and custom attributes exactly. They
do not borrow selected-row metadata to describe an entire population. Schema URLs
are not exposed by the current record adapter; convention profile and schema
knowledge remain explicit unknowns. Native bindings preserve scope, library version,
exact scope schema URL (including empty), instrument/unit and attribute selectors.
Resource schema URLs remain per-stream evidence; they are not fabricated as one
population-wide schema. A matching schema URL does not itself resolve a convention
profile. Profile expansion/migration remains future work.

Source binding IDs (`clickhouse.otel-records/v1`, `clickhouse.otel-measurements/v1`,
`clickhouse.otel-metric-discovery/v1`) identify the current descriptor's adapter
contracts. They are not historical SQL plan hashes or an authenticated source
identity. Existing capture plans and evidence bases remain authoritative for the
actual stored read; this descriptor is derived using current planning code.

## Compatibility and validation

No persisted query, definition, result or receipt model changed. Descriptors are
returned only in source-free plans, so old capture JSON/hashes/reopen behavior
remain intact. Existing capture integrity/plan-change checks still apply.

Go checks cover record/package/wire roundtrips, OTel/custom namespaces, immutable
metadata, independent action trees, native binding/schema identity, inventory and
comparison members, and rejection of false relationship claims. Existing Hegel
recipe/follow-up state/swarm tests now check the shared descriptors and composed
transformations. The live follow-up Quint ledger compares returned plans with Go
planning and adds negative controls for false grain/relation metadata. Existing
real browser journeys exercise recipe/follow-up rendering, explicit reads,
source-free preview, fixed scope, draft/capture preservation and mobile layout;
Bombadil checks checked scope and unknown convention metadata in the new preview.
See VALIDATION.md for actual run results.
