# Product checkpoint after 429: UX priorities

Status: independent proposal, not implemented work or a new execution queue.
Based on current feature contracts, recent actual 427/428 frames, the 429 source
review, [known issues](../KNOWN_ISSUES.md), and the earlier
[vendor gap analysis](PRODUCT_GAP_ANALYSIS_2026_11_12.md). No new vendor account,
hands-on comparison or production workload benchmark was performed.

Goo's core loop is now substantial: watch current records, narrow a population,
inspect traces/attributes, compare, and preserve either exact intent or evidence.
The next cycle should improve the first useful observation for unfamiliar OTel
data, rather than add another variation of capture or another top-level workspace.

## Proposed next finite cycle

Cross-review with architecture favors **437 breadth → 438 depth → 439
consolidation**. There is no UX blocker to that order; these are proposed outcomes,
not implementation acceptance or authority to weaken existing source contracts.

| Priority | Operator outcome | Bounded proposal and acceptance |
| --- | --- | --- |
| 437 — breadth | **“Show recent logs; I do not know the environment names yet.”** | Add an explicit bounded All environments logs entry within the configured source. Typed all must remain distinct from named, empty and unreported environment scope. Never choose all as a fallback for an invalid/refused selector or silently auto-start it. Preserve the configured source's authorization boundary, source/read budgets, visible applied scope and explicit Start/Run. Named and empty/unreported fixtures must prove distinct populations; a high-volume source must show its actual bounded/refused state rather than infer completeness. |
| 438 — depth | **“My instrument exists; let me inspect its values.”** | Discovery and service preflight correctly refuse a missing `service.instance.id`, but that is not an OTLP requirement and leaves ordinary service metrics as metadata only. Add a bounded gauge-only raw observation continuation from the instrument row, initially table-first. Show recorded values, times, resource/scope identity and diagnostic ambiguity without invoking a service measurement kernel, merging unidentified producers or inventing IDs. Explicit Run owns the read. A fixture with only `service.name` must reveal real observations while derived rates/aggregate health remain unavailable. Do not draw a connected series across unidentified producers, erase duplicate occurrences or substitute zero for missing values. Existing exact-instance and host/pod paths keep their semantics; a new capture family is not part of the first slice. |
| 439 — consolidate | **“Finding my data remains predictable as the catalog grows.”** | Measure the exercised first-use paths before optimizing: family capability probes, environment and instrument discovery, repeated scope changes and retained return. Use large catalogs and source refusals, and record actual requests, work and visible first-data position. Consolidate only matching pure projections or request-local probe work; do not merge acquisition/coverage/custody owners or add an unbounded persistent readiness cache. Source absence, denial, incompatible schema and no-data remain distinct. Recheck explicit all-scope and raw-gauge paths on keyboard/narrow layouts and under withdrawn/late completion. |

The raw-native gap is verified in source: `DiscoveryObservation` retains identity
labels rather than values/interval validity; the instrument inspector shows
resource/attribute presence and counts. Existing value tables/timelines follow a
successful evaluated Run, which service preflight blocks for known missing
instance identity. Existing data-first native views therefore do not close this
first-observation gap. Architecture must accept the raw acquisition/identity
boundary before implementation.

## Retained alternatives, not parallel queues

Presence-safe CEL attributes remain useful: visual filters can already express
resource/log/span attributes, while optional CEL has a fixed nonnullable field
vocabulary. A future bounded slice should start with discovered string fields and
missing/present-empty distinctions through the existing typed lowering path,
not numeric coercion or an arbitrary extension registry. This is deferred behind
users who currently cannot see their first metric values; visual filtering is
already a working alternative.

An optional combined log/span Activity view is also a real extension: current
Activity supports one active family. It requires per-family readers, watermarks,
limits and partial-failure/hold state, and must not imply a total arrival order,
unique request count or common snapshot. Explicit All environments logs is the
smaller first-use change; signal union should not be bundled into it.

## Keep the product coherent

- Reuse Activity, Measurements, current inspectors and shared supplied-data
  components. Do not create another dashboard builder, Saved area or universal
  owner/cache to deliver these outcomes.
- Keep names and conditional controls task-oriented. Source schemas, compiler
  versions and exact receipts belong in technical disclosures; the first screen
  should show useful current data and the scope needed to understand it.
- Preserve the 429 lesson: measure the same read state, retain exact detail, and
  remove irrelevant controls instead of hiding active meaning. Source-free local
  interaction, source reads and publication remain separate actions.
- Existing service/resource dashboards, waterfalls, exception groups, population
  explanations, templates, annotation tools and captures are delivered. Their
  older proposals do not make them missing primitives.

Durable monitoring/notifications, authenticated team sharing, pipeline ingestion
health and external warehouse execution remain important separate gates. They
are not quick additions to this cycle and should not be disguised as green health
badges, ground-truth labels or outbound delivery from a local prototype.

## Source anchors and limits

[Live Activity](../LIVE_ACTIVITY.md) specifies one active signal and bounded event-
time monitoring. [Signal availability](../SOURCE_CAPABILITIES.md) supplies current
partial-family readiness. [Instrument discovery](../NATIVE_INSTRUMENT_DISCOVERY.md)
and [preflight](../INSTRUMENT_PREFLIGHT.md) explicitly distinguish metadata from
values and service identity policy from OTLP requirements. [CEL product design](../CEL_PRODUCT_INTEGRATION.md)
and [current editor](../CEL_CONTEXTUAL_EDITOR.md) establish the delivered optional
language path and its protected parent.

The capability index's older intro still called CEL integration unimplemented
when this review began; the parent was notified to reconcile it against current
contracts. Historical gap tables were treated as research, not authoritative
current feature inventories. This proposal itself ran no new product integration
tests; 429 source/frame/browser acceptance is recorded separately.
