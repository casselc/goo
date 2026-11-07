# Whole-product architecture checkpoint — 372

Status: source-only checkpoint after accepted 370/371. Final 372 integration
acceptance is recorded in VALIDATION.md. Proposed next cuts are not delivered;
root adoption is in ROADMAP_RECONCILIATION_372.md.

## Delivered direction and remaining product seam

370 makes typed error populations navigable with current-result action leases
and retained Back. 371 adds a genuinely separate native histogram display arm:
measurement calculation identity stays independent of presentation identity,
ordinary record templates retain their guards, and actual destination exports
supply bins only after explicit Run and interval selection. 372 addresses a
sibling typed continuation seam; its selected contract intentionally retires
selection proof on Back until parent Run rather than reviving old range intents.

The product now has substantial general record exploration, native/resource
measurements, OTel role views, reusable record compositions, captures and bounded
review/export. Do not schedule those foundations again or describe the native
histogram toggle alone as reusable data acquisition. The next gap is making the
new reusable configuration naturally discoverable, not another executor/store.

## 373 breadth: discover saved native displays without copying IDs

`ui/native-histogram-reuse.mjs` currently requires an exact view ID. That was an
honest finite pilot, but users cannot naturally find a saved display again.
`DefinitionLibrary` already serves measurement definitions, Activity views and
investigation revisions. `workbench.Store.Documents` provides bounded cursor
pagination; the new native-view artifact family already has verified load.

Add an explicit bounded metadata list for that family and reuse the existing
library presenter in Measurements. A list result can show title, kind, unit and
ID; exact Inspect/Apply still loads/verifies the immutable artifact. Apply binds
the current checked question unchanged and retains explicit Run/interval selection.
Do not select donor measurement scope merely by selecting a display row.

Use deterministic ID order and a small page limit (the existing 20-item document
pattern is sufficient), not a new index or misleading “recent” order: native views
have no creation clock. Bound files loaded and response bytes per page; do not
load 10,000 maximum-size artifacts to build a convenient client list. Search must
say loaded page unless a bounded server search is deliberately implemented.
Unknown/corrupt artifact handling must be explicit rather than silently implying
a complete catalogue. No task registry, backend migration or background scan.

Acceptance: save → discover → Inspect → Apply to another current destination →
Run → actual interval, with metadata/telemetry counters separate, zero Apply
telemetry, exact destination identity and stale list/load callbacks rejected.

## 374 depth: exponential histogram support readiness, not a guessed alias

Current `metrics/histogram.go` represents explicit bounds and exact bucket counts;
`measurements.HistogramView` and the shared dataset require explicit-bin closure.
Current discovery/source/known limitations do not establish exponential histogram
support. An OTel application exporting that representation is a real coverage gap,
but a metadata label or UI kind cannot solve it.

First pin an official OTLP/exponential histogram contract and confirm actual
collector/ClickHouse table fields. Close scale, offsets, positive/negative/zero
buckets, zero threshold, count overflow, temporality/reset and stream identity.
An exact conversion into current explicit binary64-bound displays must not be
assumed: exponential boundaries at positive scales may require a new bound
representation or honest conservative intervals. Do not lose exact counts or
claim producer conformance merely because a name/unit matches.

The deliverable is a finite implement/defer decision with a genuine exporter-shaped
fixture and independent numeric oracle. If a small supported nonnegative duration
arm closes, reuse existing measurement ownership and acquisition paths; do not
mix explicit and exponential streams or silently rebin. If a new broad histogram
representation is required, defer with the actual blocker named and choose the
next user outcome. No new query engine or native chart platform is implied.

## 375 consolidation: close the native-view admission truth table

371 found concrete drift across Go, JS and OpenAPI: explicit `resource:null` and
empty native bound strings were accepted on one boundary and refused on another.
These are wire-presence semantics, not a reason for a universal validator.

Close a finite native-view question contract matrix spanning create, immutable
load and reuse, with whole question preservation and omission/null/empty/Unicode/
paired-nanosecond cases. Reuse existing strict Go codec, JS measurement-scope
protocol and schema generator; fix actual differences rather than layering
compatibility aliases. Keep independently specified expected outcomes and raw
boundary/membership oracles; generating every expectation from one validator
would hide drift instead of detecting it.

This should accompany the chosen native feature where possible, not become a
permanent test-only maintenance queue. If 373/374 already closes the matrix, record
375 as covered and do not manufacture a refactor. Plain typed Go APIs, maps.Clone,
slices.Clone and immutable publication already fit this code; no advanced generic
user machinery or Go upgrade task is justified (1.27.3 is pinned).

## Reconcile and defer honestly

Retire stale claims that saved/service interval context, current RPC, trace tails
or captured-record review foundations are absent. Keep differences between held
return, shared question, durable capture and view configuration explicit.

Broker backlog still needs one real pinned exporter and duplicate-observer
contract; demo queue depth is not generic broker health. Exemplars need actual
source fields, not time-coincident trace guesses. Tenancy/authenticated review,
current-head consensus, broad WYSIWYG and Databricks lowering remain separate
contracts. No annotation platform or global navigation manager follows from them.

369's bounded paging must retain its full-model counts, scale and unknown states;
repeat actual styled tasks if future row controls materially change cost, rather
than automatically opening another facet loop. Continue genuine owner-entry
checks and independent integer/identity oracles, with tracing disabled for latency
claims. Review again after this finite sequence or an explicit depth deferral.
