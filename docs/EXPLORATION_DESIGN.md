# General exploration and expanded context

Status: first bounded slice implemented. This document defines its contract and
separates the remaining work toward a general observability workbench.

## User journey and design

An engineer starts with captured spans or logs, an explicit environment and a
bounded event-time window. The UI discovers available resource/record attribute
keys, builds typed filters and groups captured records. A result page is an
inspection surface; it never supplies the population denominator. The engineer
can open a trace across services, inspect a span and find logs with the exact
reported trace/span identity. The original query and its basis remain separate
from expanded context.

The code path is UI or `goo plan-exploration` → shared Go query AST → validation →
ClickHouse adapter → bounded source execution → inspectable results and basis.
Physical field/map bindings live in the adapter, not in UI condition branches.
The database plans and optimizes its SQL; the reference `Match` function is a
predicate interpreter for semantics verification, not an alternate query engine
or optimizer. DataFusion is not required for this slice. A later planner could
consume the logical AST without replacing measurement semantics.

The generic API currently shares the packages service and its startup configuration;
it is not yet an independently configured query-server deployment. The physical
adapter targets the pinned exporter schema, not arbitrary ClickHouse schemas.

The raw-record query model and the existing prepared-fact measurement compiler
are currently distinct pilots. This is not a claim that every existing UI query
already uses one universal grammar. The intended next language step is a shared
typed expression layer and source/grain contracts that both lower into, with
YAML resource authoring and an explicitly bounded calculation language (CEL/KCL
still undecided). JSON is today's transport and CLI query input. No arbitrary
SQL or executable calculation strings are accepted.

## Implemented query contract

`exploration.Query` contains:

- Signal `spans` or `logs`, mandatory environment, half-open `fromMs`/`toMs`
  event/start-time window, at most 120 minutes before 2100.
- Optional recursive `all`/`any` predicates: at most 32 nodes and depth four.
  The UI builds up to 16 AND/OR leaves; deeper ASTs use the API/CLI today.
- Fields `{kind, key}`: allowlisted logical columns, arbitrary retained resource
  attributes or signal attributes. Keys are 1–128 UTF-8 bytes, values at most
  1,024 bytes; NUL is rejected. Attribute keys/values are typed SQL parameters. The HTTP adapter additionally
  escapes ClickHouse Escaped-format parameter text after URL decoding, preserving
  literal backslashes/control characters and Unicode.
- `eq`, `ne`, literal case-sensitive `contains`, `exists`, `missing`; UInt64
  `duration_ns` also supports `gte`/`lte` and exact decimal integer comparison.
  Exported map values are strings: numeric-looking attributes are not silently
  converted. Missing attributes do not satisfy `ne` and never equal empty strings.
  Physical columns exist even if their contents are empty. No regex, arbitrary
  order expressions, numeric casts of attributes or percentile calculation yet.
- Zero to two grouping fields (UI offers one), and a 1–100 row page with offsets
  0–1,000. Row ordering includes timestamp, IDs and captured field content to
  reduce equal-key instability; offsets are not stable cursors or frozen snapshots.

`Compile` emits separate row, count, grouping and field-discovery SQL. Counts
are UInt64 decimal strings; group dimensions retain NULL for missing attributes
and empty string for present-empty values. More than 1,000 groups fails the
request, rather than returning partial grouping totals. Up to 128 dynamic fields
are discovered across all matching records; truncation/unsupported oversized keys
are labelled. Missing from a field list does not prove absence. Static columns
are separately offered. No distinct-value/cardinality inventory exists yet.

Counts measure captured records, including duplicate exporter rows and retry
invocations. They are not unique invocation, request or business-subject counts;
those need the existing operation/lifecycle grain contracts. Exploration counts
must not silently replace an SLI denominator.

## Source consistency and budgets

Each response freezes its absolute question and includes a content basis over
question, SQL/parameters and returned evidence. Capture time is a separate server
receipt timestamp, not ingestion custody. The basis covers the returned count,
groups, inventory and page; it is not a hash of every matching raw record. Different
pages therefore have different bases even when their count reads agree.

Queries are sequential source reads, not a transactional snapshot. A detected
count/group/page disagreement returns `changed_between_reads`, retains each actual
result and exposes a warning. This detects some changes, not every source mutation.
An empty consistent count is `no_data`, never healthy; query failure is an HTTP
error with no fabricated zero. UI generations discard superseded results and
clear old rows on source failure. Trace and log requests also have separate
request generations within the active context.

The Go boundary accepts 16 KiB JSON and rejects unknown fields. Execution uses a
55-second request context, per-query 10-second source time limit, 256 MiB memory
and read-byte settings, 8 MiB response cap, read-only source queries and strict
row/group budgets. These are local query safeguards, not a shared production
admission-control or tenant authorization system. General authentication/RBAC,
row-level scope enforcement, concurrency budgets and source indexing remain work.

## Trace and log expansion

Trace expansion uses an explicit nonzero lowercase 32-hex trace ID and the
original environment/time bounds. It drops service, outcome, grouping, duration
and other original search predicates. The result can include all services and
libraries within that scope, including parents excluded by the search. It does
not fetch outside the window or claim a complete trace. Pages show 100 spans at
a time, up to the same offset budget; each waterfall is for that page, and missing
parents can live on another page. Parent cycles/missing parents are visible;
OTel links remain references, not invented parent edges. Raw details preserve
resource/span attributes, scope, status message, events and link identity/attributes.

Logs match exact reported trace ID, optionally AND a selected nonzero 16-hex span
ID, in the same environment/time scope across services. They have their own
question, count, page and basis. Matching identity is not durable provenance or
causality proof. Log→trace navigation carries the selected log response basis and
question as its origin, rather than reusing an earlier span-search basis. An async linked consumer trace does not automatically join the
parent trace; following arbitrary links is a later workflow. Business identity
and time-only contextual search must remain distinct relation types.

Operation waterfalls now offer explicit expansion into this surface. An expansion
from a saved capture is a **fresh live read** with its own basis, not a mutation
of the saved bundle. General exploration/context is not yet archived by the
operation-only capture API; saved general query workspaces are a follow-up.

## Executable journeys and next steps

[Journey 20](journeys/README.md#20-general-exploration-and-population-safe-paging)
uses a labelled 123-span OTLP fixture across two synthetic resource names to prove
100/23 paging with independent full counts, null-versus-empty grouping, arbitrary
key escaping, AND/OR predicates, discovered fields, no-data and stale response
rejection. This fixture is a controlled collector integration, not 123 real app
requests. [Journey 21](journeys/README.md#21-real-distributed-trace-and-exact-linked-logs)
uses the real web→gRPC sample request to expand a service-filtered result into a
cross-service waterfall and exact trace/span log scopes. Journey 22 demonstrates
that fresh expansion from a saved operation receipt keeps that receipt unchanged. Neither journey mocks
telemetry query responses; the race test delays a real source response.

Go reference/compiler/source checks complement Hegel AST state/swarm tests and
Quint observed search/page/trace/log validation, including negative controls for
invented population/scope/identity agreement. See VALIDATION.md for actual runs.

Next proposed slices:

1. Type-aware numeric bindings and multi-attribute distributions; time/latency
   aggregate charts sharing the same source/grain contract.
2. Cursor strategy or materialized captures for stable paging; stored expanded
   evidence with explicit successor refresh, and trace-window widening.
3. Persisted general cohort receipts and broader attribute ranking; explicit
   sampling/coverage contracts. One-field descriptive comparison is implemented.
4. Named query/board resources and plan/apply; capability diagnostics for warehouse
   lowering. Raw row exploration is not a Databricks metric-view definition. The
   existing portable aggregate compiler remains artifact-only for Databricks;
   this milestone does not add a warehouse runtime or validate an instance.

Live tail, arbitrary metrics/gauges, whole-trace accumulation across pages, text
search across every field, RUM/session replay, dependency maps, production SLOs,
alerts and team collaboration remain outside this slice. Log `body contains`
provides literal text search, not an indexed/full-text search language.

## Workspace and cohort distributions

The [product workspace contract](PRODUCT_WORKSPACE.md) defines implemented result
tabs, fixed-window baselines, comparison arithmetic and responsive trace inspection.
General cohorts remain browser-held, separate from durable operation receipts.

## Reusable definitions from captured questions

[Telemetry packages](TELEMETRY_PACKAGES.md) now reuse the same query AST without
source manifests. They store selectors/grain/views, separate from preview bounds,
pages and data bases. Applying one preserves imported nested predicates/two-field
groups and allows local refinements without mutating the saved definition.

## Log discovery, time context and span-duration buckets

Implemented: message/service/severity search adds visible AND leaves to the same
query AST. Body matching is literal and case-sensitive; severity is an exact
captured string. Colour is only a display aid, not a replacement severity mapping.
Messages preserve whitespace and render as text; long messages expand on demand.
The attribute composer remains available behind a disclosure after quick search.

**Surrounding logs** is a separate, cancellable read around a selected log. It
retains environment, restricts to the same captured service name, and intersects
±30 seconds / ±1 minute / ±5 minutes with the original half-open question window.
The anchor's millisecond is included in full; its exact nanosecond value remains
in evidence. Original trace/span IDs, package, body, severity and attribute filters
are omitted explicitly. This is time context, not exact correlation or causality.
A fresh question/basis/count is retained separately; paging does not redefine its
denominator. Identical captured records cannot be uniquely distinguished, so the
anchor highlight says its fields match. It does not invent an event identity.
The Go reference `SurroundingQuestion` validates the live browser's composed AST;
Hegel state/swarm checks scope, clipping and count/page separation.

Trace discovery groups the full matching captured span population by `trace_id`.
Valid reported IDs can open the existing independent cross-service trace read in
original fixed bounds. Counts describe matching spans; roots, end-to-end latency
and complete trace capture are not inferred. Invalid/zero IDs remain visible but
cannot open a trace. The existing 1,000-group budget fails rather than truncates.

The spans-only logical field `column:duration_bucket` is a fixed string-valued
partition over UInt64 `duration_ns`. The Go reference and ClickHouse adapter use:

| Bucket (ms) | Lower bound inclusive (ns) | Upper bound exclusive (ns) |
| --- | ---: | ---: |
| `[0,1) ms` | 0 | 1000000 |
| `[1,5) ms` | 1000000 | 5000000 |
| `[5,10) ms` | 5000000 | 10000000 |
| `[10,50) ms` | 10000000 | 50000000 |
| `[50,100) ms` | 50000000 | 100000000 |
| `[100,500) ms` | 100000000 | 500000000 |
| `[500,1000) ms` | 500000000 | 1000000000 |
| `[1000,+inf) ms` | 1000000000 | unbounded within UInt64 |

Grouping and equality filters use this same field. Bars use complete captured
groups, never the displayed row page. The UI orders buckets by their intervals.
Minimum duration input converts decimal milliseconds to exact nanoseconds with
at most six fractional places, without floating-point multiplication. These are
span-record distributions, not native histogram metrics or trace latency.

Suggested breakdowns use available logical fields (services, status/severity,
trace IDs, duration buckets and observed exception type). They are user-selected
investigation leads, not inferred business semantics. Create package retains the
captured filters/grouping; preview and fresh reuse share the Go compiler. No
new wizard language, automatic annotations or Databricks lowering is implied.
See [journey 30](journeys/README.md#30-from-log-search-to-reusable-span-distributions).

The duration bucket bounds are part of this logical field's fixed contract.
Changing them requires a new/versioned contract rather than silently reinterpreting
saved definitions. Captured evidence bases include the actual compiled plan.

Re-anchoring on a nearby record retains the initial search window/basis while
recording the anchor's own selection question and evidence basis separately.
A neighbor need not satisfy the original message/severity filters, so its provenance
must not be attributed to that initial filtered capture.

## Service inventory entry point

The Services destination uses the shared physical adapter and duration partition
for a bounded single-read aggregate, then opens a separate exact service/name
Explore question in the same environment and fixed bounds. Capture totals remain
independent of local search and the Explore record page. See
[SERVICE_INVENTORY.md](SERVICE_INVENTORY.md) for semantics, budgets and limitations.

## Durable general captures and stable stored paging

Explore span/log questions and service inventories now have their own immutable
capture path, separate from operation receipts. Full counts/groups and at most
1,000 ordered inspection records are retained; cursor pages read only the stored
bank. Reopen is source-free, while fixed-question/plan refresh creates a successor.
Ordinary live offsets remain mutable. Expanded context and browser baseline pairs
are not automatically stored. Explicit [comparison capture](ERROR_COMPARISONS.md) now rereads and archives two fixed questions plus the comparison. See [the complete contract](GENERAL_CAPTURES.md).
