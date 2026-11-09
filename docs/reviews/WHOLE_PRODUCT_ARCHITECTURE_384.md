# Whole-product architecture checkpoint — 384

Read-only source review during the 384 layout cut. The recommendations were adopted in the
[November 9 roadmap](../ROADMAP_2026_11_09_CHECKPOINT_384.md) after runtime acceptance.
The source analysis below records the review's rationale; actual geometry and
runtime checks are recorded in [validation](../../VALIDATION.md).

## Keep the existing ownership boundaries

Goo's useful loop is current telemetry, a scoped population, evidence follow-through,
and preserved intent or evidence. The server-side measurement/query adapters and
immutable capture owners already support much of that loop. Adding another query
engine, dashboard manager or navigation framework is not justified by the gaps
below.

`ui/exploration.mjs` owns the mutable fixed question and its session. Its read-only
entry delegates immediately to `ui/saved-records.mjs`; shared layout work must
include both owners. `ui/record-results.mjs` owns row interactions, trace/log
inspectors and their request generations. It should not become a universal page
presenter. `ui/resource-return.mjs` owns retained ancestry presentation while the
origin owner retains authority. Actual parent DOM, selection and action leases
must survive a child visit.

The 384 source candidate introduces a supplied-fact summary shared by live and
stored record views, rather than a new persistent schema or execution layer.
`FixedQuestionEditor` must still force dirty, invalid, failed and canceled editing
states into view. New native disclosures must not authorize captured hidden
callbacks, and requested histogram/time-series visuals must remain available as
primary content rather than disappearing into secondary context.

Source review identified and coordinated fixes for specific consolidation risks:
new disclosures needed visibility guards on source/capture actions; hiding duplicate
status text also hid transient failures and successor feedback; the applied
reported-edge widening action needed to remain available; explicit histogram and
grouped-time-series nodes needed promotion when Distribution is chosen; limited
paging/field discovery and retention needed a visible notice. The current source
candidate preserves these behaviors. Browser geometry and measured no-read/Back
checks remain the acceptance authority.

Current Go 1.27.3, typed request structs, `reflect.TypeFor`, standard slices/maps
helpers and explicit source budgets are appropriate. This UI cut does not justify
a runtime upgrade or public generic API redesign.

## Proposed breadth: outgoing HTTP observations

The source confirms a narrow core-observability gap. `exploration/dashboard.go`
admits HTTP server, current/older RPC server, database-client, messaging
consumer/producer and all-span families. There is no HTTP-client family. In
`measurements/instrument_capabilities.go`, `http.server.request.duration` has a
profile while `http.client.request.duration` remains a generic native binding.
Database-client and messaging functionality already exist; they are not missing
vendor categories to recreate.

A bounded outgoing-HTTP lens can reuse DashboardRequest/Plan, ordinary record
questions, discovered service inventory and existing captures:

- Select exact Client-kind spans with nonempty current `http.request.method`.
  Report client-span record counts, duration and recorded span status, never
  unique request throughput or an inferred universal failure policy.
- Offer qualified remote endpoint dimensions such as `server.address` and
  `server.port`, retaining missing and present-empty values. Do not guess an
  instrumented callee service from a hostname, merge namespaces, or describe
  reported associations as verified topology.
- Keep metadata candidacy separate from conditions checked in the actual read.
  Use a pinned OTel convention profile; historical HTTP aliases require a separate
  explicit policy rather than an implicit OR of differently defined attributes.
- Reopen/save through the existing record-question/capture path. Following a
  reported trace is explicit and retains the outgoing population as its origin.

Native client-duration adoption is a separate later decision: generic execution
may already be eligible, but a named profile requires exact-stream units,
temporality and coverage. It is not a prerequisite for the first span lens.

## Proposed depth: reusable gauge display declarations

Native gauge evaluation and exact-stream timeline presentation already exist.
`ui/measurements.mjs` distinguishes gauge samples from counter/histogram intervals,
retains stale/unknown state and states that unpainted time is not zero. Current
saved native display declarations in `workbench/native_histogram_views.go` are
explicit-histogram export-interval declarations only.

Extend reusable presentation first to **gauges only**, preserving one checked
measurement question and its original exact streams. A bounded timeline/table
choice can reuse the existing supplied-value components and immutable definition
store. Applying display configuration stages the destination question without
reading telemetry or importing the donor's evidence, service identity or window.
Keep unit and stream compatibility explicit; do not silently convert units.

Freshness remains evaluated at the fixed question end. Preserve missing samples,
stale marks, source read basis, original units and display limits. Do not connect
gaps, interpolate, sum independent streams or imply a fleet gauge. Counter display
reuse remains later work because rate intervals, resets and exact differences
need their own declared grain; a gauge timeline is not a counter representation.

## Proposed consolidation: close current capture-family admission

Milestone 383's generated Go/JavaScript corpus exposed search evidence relabeled
as an ordinary capture reaching a UI family shortcut. The targeted fix is now
covered, but `ui/saved-capture-protocol.mjs` still has family-specific early returns.
The existing capture families are explore, inventory, comparison, measurement,
explanation and trace-search.

A finite consolidation can make this boundary easier to audit:

1. Admit the common full-page envelope and exactly one known question/result
   family before dispatch. Explicit null foreign arms must not disappear into
   omitted defaults. Unknown kinds and mixed families fail before family shortcuts.
2. Invoke small pure per-family semantic validators. Keep their exact scope,
   independent read/custody clocks, page/list bounds and historical-plan rules;
   do not reinterpret a search witness list as record paging.
3. Compare Go and JavaScript admission over generated valid cases and rehashed
   semantic mutants across all six current families. Include disguised kinds,
   missing/null/unknown fields, duplicate object semantics, exact numeric values,
   scope mismatches and malformed response combinations.
4. Retain separate APIs for metadata-only document references and full capture
   pages. A summary is not a full receipt/result, and display-only components need
   not acquire capture-store or query authority.

This needs no reflection-driven validator framework, new envelope format, migration
layer, generic registry or public type machinery. Use existing current wire types;
there is no compatibility floor to preserve obsolete aliases. Current Go admission
and immutable-store integrity checks remain authoritative, but the browser should
refuse inconsistent supplied payloads before rendering or navigation.

## Keep deferred work explicit

Log-pattern display grammar remains a useful bounded later facet: patterns,
counts/shares, comparison and saved questions already exist; the missing work is
checked `body_pattern_v1` display admission with its denominator and unclassified
coverage. It is not a new clustering engine.

The contextual review-task chooser followed by a verified kept-occurrence review
target remains a coherent later breadth/depth pair. Captured-record review and
historical export already ship. Do not fabricate an Explore capture to label a
kept item or start a second annotation system.

Only introduce a shared span-family selector descriptor if outgoing HTTP reveals
concrete repeated branches that are simpler as a small pure table/helper. A
hypothetical future family registry is not a prerequisite. Exponential histogram
execution, alert delivery, multi-user authorization/presence, arbitrary executable
kinds and new data stores retain their separate design gates.
