# Roadmap after measured operator consolidation

The finite 427–429 cycle covers exact definition recall, paired evidence and
measured scope/preservation controls. Executed results belong in
[validation](../VALIDATION.md), not in this proposal's acceptance criteria.
The [architecture checkpoint](reviews/PRODUCT_CHECKPOINT_AFTER_429_ARCH.md) and
[product/UX checkpoint](reviews/PRODUCT_CHECKPOINT_AFTER_429_UX.md) review the
broader product rather than extend another preservation queue.

Status: **429 and 437–459 delivered and validated**. The current cadence is
**444 paired Activity breadth → 445 consolidation → 446 reusable paired views → 447 native histogram inspection depth → 448 consolidation → 449 cross-signal investigation breadth → 450 independent lane-filter depth → 451 filter consolidation → 452 explicit reported-link target lookup → 453 referring-span discovery depth → 454 selector/reference consolidation → 455 held raw-gauge Points breadth → 456 single-point histogram bucket depth → 457 raw-display consolidation → 458 held log/span chronology breadth → 459 requested UI/UX/accessibility polish and guidance assessment**.
**445 single/paired Activity coherence and measured work is delivered**. The existing
Go runtime, source adapters, typed questions and shared components remain the
foundation; this does not extend a CEL platform queue.

## 437 — explicit All-environments log entry

**Delivered:** [contract](ALL_ENVIRONMENTS_LOG_ENTRY.md),
[storyboard](journeys/all-environments.md) and [validation](../VALIDATION.md).
The design and acceptance basis below records this bounded log-only slice.

An operator who does not know environment labels should be able to explicitly
start a bounded log stream across the configured workspace source. Current named,
empty and unreported scopes are real distinctions; none means All. Blank input or
a failed discovery must never silently broaden the scope.

Design a typed All scope before implementing the entry. Check its meaning through
AST validation, source lowering, links, saved definitions, captures and held return.
Retain fixed bounds, limits and source refusal. Environment selectors are scope,
not a substitute for source authorization. All stays inside the configured
workspace source; this does not introduce multi-user authorization or tenancy.

Start with logs. Show reported environment facts per row and preserve missing
versus empty labels. Narrowing and row-context actions must retain or explicitly
narrow the parent scope; unsupported routes refuse before a read. Do not infer a
replacement query from a display label or treat a named scope as All.

Acceptance: named/empty/unreported data in one source, explicit opt-in, typed
links/recall, source/access refusal, bounded high-cardinality reads, local narrowing
and Return, desktop/320/390 and keyboard. Measure requests and source work. Mixed
log/span streaming needs its own per-family watermarks and failure policy later.

## 438 — inspect raw gauge observations before evaluation

Delivered: [raw observation contract](RAW_GAUGE_OBSERVATIONS.md),
[real OTLP storyboard](journeys/raw-gauge-observations.md) and
[validation](../VALIDATION.md). The design basis below records the bounded slice.

Ad-hoc metric producers can lack service.instance.id. Discovery currently shows
identity labels and counts, while service evaluation correctly refuses known
missing identity. The evaluated data-first view is downstream of that refusal;
it is not a raw-value inspection path.

Add a bounded, explicit, table-first raw gauge read from discovery. Define a
raw-observation profile rather than constructing an evaluated GaugeResult.
Preserve reported attributes, duplicate occurrences, units, timestamps and
diagnostics. Local row keys identify occurrences, not invented producer IDs.
Missing values remain unavailable, not zero. No fleet aggregation, continuity,
rate, reset, health, freshness or completeness claim follows from these rows.

Keep evaluator preflight intact. A timeline is permitted only after its reported
identity and sample-validity contract is established; missing identity does not
authorize connecting points or merging streams. Do not add a new capture carrier
to make this inspector fit. Reuse supplied-data table/detail components.

Acceptance: real OTLP gauges with and without instance identity, repeated rows,
multiple attribute sets/units, malformed or absent values, source limits and
refusal. Opening metadata or changing a local display cannot read values;
Read observations explicitly does; inspecting held rows stays local. Keep
source-read time distinct from observation time.

## 439 — measure first-use acquisition and consolidate matching work

Delivered: [contract](FIRST_USE_ACQUISITION.md),
[measured storyboard](journeys/first-use-acquisition.md) and
[validation](../VALIDATION.md). The bounded change retains settled observations
on temporary navigation; per-request source probes remain independent.

Measure discovery → explicit read → inspect → retained return over the new paths.
Record cold and retained requests, per-family SQL/bytes, cancellation, row work
and high-cardinality behavior before proposing caching or refactoring. Compare
the same question and read state; rolling views need separate clock-aware checks.

Consolidate only duplicated pure projections or probe work with identical source,
scope, completeness and lifetime contracts. A schema probe is not evidence of
fresh data. No durable readiness cache without an invalidation contract; no
universal identity or permission layer. Keep source and native evaluator owners
separate. Refresh executable storyboards, responsive frames and native tests.

## 440 — presence-safe CEL attribute conditions

Delivered: [attribute helper guide](CEL_ATTRIBUTE_CONDITIONS.md),
[real OTLP journey](journeys/cel-attributes.md) and [validation](../VALIDATION.md).
Literal-key helpers preserve present-only comparisons; membership and negated
membership express reported/missing. No guarded-index rewriting is introduced.

Extend the existing checked-condition workflow to bounded qualified resource and
record attribute access. Define missing, empty, invalid and typed-value semantics
first; visual attribute gestures and CEL must use the same predicate AST and
source adapter. Test presence guards and short-circuit behavior against reference
and source execution. Preserve the existing revision-owned additional-condition
slot, diagnostic spans and explicit Apply/Run boundaries. No second executor,
implicit coercion, arbitrary extension registry or unbounded dynamic field path.

## 441 — raw sum observations

Delivered: [sum contract](RAW_SUM_OBSERVATIONS.md),
[real OTLP journey](journeys/raw-sum-observations.md) and
[validation](../VALIDATION.md). OTel sums include discovered counters and
cumulative nonmonotonic stocks; both retain exact reported metadata and share
local inspection with gauges. The design basis below records the bounded slice.

Add explicit bounded table inspection of reported sum points when evaluator
readiness blocks identified-stream evaluation. Retain temporality, monotonicity,
reported start/end, scalar flags, units and duplicates. Define its own profile
before acquisition; raw rows do not establish a rate, reset, continuity or fleet
total. Reuse only projections and local display with identical contracts. Keep
native All scope unsupported and evaluator preflight intact. Acceptance must
include real delta/cumulative and monotonic/nonmonotonic OTLP points, no-recorded
values, missing identity, cancellation and local retained inspection.

## 442 — consolidate attribute refinement

Delivered: [shared refinement contract](ATTRIBUTE_REFINEMENT.md). Exclude is the
missing-inclusive complement across Activity, fixed exploration and population
selection. Shared compact presence actions work from observed fields, details,
missing table cells and categorical facets; CEL `.ne()` stays reported-only.
[Storyboard](journeys/cel-attributes.md) and [validation](../VALIDATION.md) record
the executed coverage. The design basis below records the bounded slice.

Exercise expanded attribute use across visual filters, facets, record context
and authored conditions. Consolidate overlapping gestures or helper code only
when scope, slot/revision, presence semantics and acquisition authority match.
Measure clarity and source work before adding controls. Preserve the operator's
parent question and explicit Apply/Run, and refuse partial whole-filter conversion.

The [gotogether shadow trial](GOTOGETHER_ACQUISITION_PILOT.md) now runs a separate
pinned acquisition consumer after the native gate, with Quint 0.33.1 and compiled
replay. Normal verification dependencies and independent source oracles remain
unchanged. A stateful compiler/source differential pilot is still proposed;
measure maintenance benefits before replacing existing bookkeeping.

## 443 — typed numeric attribute conditions

Delivered: [numeric CEL contract](CEL_NUMERIC_ATTRIBUTES.md) and
[real-source storyboard](journeys/cel-numeric-attributes.md). Literal int64,
uint64 and decimal64 declarations expose existing exact numeric predicates and
valid/missing/empty/invalid states. Visual conditions, CEL reference evaluation,
inverse printing and ClickHouse share the existing numeric binding and parser;
no second evaluator, implicit conversion or numeric inference was introduced.

## 444 — paired Logs + Spans Activity

Delivered: [paired Activity contract](PAIRED_ACTIVITY.md),
[real-source storyboard](journeys/paired-activity.md) and
[independent acceptance](reviews/PAIRED_ACTIVITY_444_ACCEPTANCE.md).
One coordinator composes two ordinary Tail owners. Complete common predicates
and narrow environment policy are admitted before either source connection.
Each family retains its own window, buffer, Find, pause/retry/refusal and local
inspection. Fixed drill-in keeps the selected occurrence's full parent/window;
Return reads neither lane. All remains logs-only, and the pair remains session-only.

Live source controls cover refusal, retry, canceled real success, empty family,
visible saved-single adoption and shared-filter refusal. Bombadil and native
Quint trace validation exercise the real mounted workflow with rejecting controls.
Initial-state, native inactive-callback, recall-conflict and focus fixes are part
of the delivered shared path, not separate follow-up promises.

## 445 — consolidate single/paired Activity (delivered)

Delivered [shared controls and measured storyboard](ACTIVITY_COHERENCE.md):
responsive Find/field actions, container-aware record columns, UTC inspection,
compact paused facts, paired start admission and native Return focus. Source,
connection and custody owners remain separate. Measured staging/inspection/Return
add no connections or SQL; real four-slot pressure leaves the healthy lane alone.
The existing unavailable/late-read journey, native model gate and Bombadil actor
remain acceptance checks. No readiness cache or polling change was justified.

## 446 — reusable paired Activity definitions (delivered)

Delivered in the existing [Activity view library](ACTIVITY_VIEWS.md), with
[real-source storyboard](journeys/paired-activity-views.md): shared exact scope,
independent lane layouts, applied-only saving, paused recall and explicit Start.
The saved-view selector projects two admitted ordinary tail queries; physical
tail transport still refuses mixed signals. Both grammars admit the full filter;
All remains logs-only. Runtime fields and incompatible branch layouts refuse.

Shared Tail preparation now retains restored layouts across Start for single
and paired views. Independent review, Hegel storage/identity properties,
existing paired-model trace validation, ten negative controls and a real
Bombadil save/recall/Return flow cover the result. No duplicate transport model,
merged chronology or per-lane filter authoring was introduced.

## 447 — raw native histogram point inspection (delivered)

Delivered [explicit histogram points](RAW_HISTOGRAM_OBSERVATIONS.md) through the
shared held viewer: exact count strings, diagnostic bucket vectors, explicit
recording state and unavailable optional-statistic presence. The source adapter
uses the original instrument/temporality/window, retains duplicates and refuses
identity drift or oversized reads. Optional physical statistic columns use the
existing schema read; they do not imply producer presence.

The [actual OTLP storyboard](journeys/raw-histogram-observations.md) covers source
oracle parity, local pages/Find/inspection/Return, declared source refusal,
retired successful response, separate unit/temporality, Bombadil and native
negative controls. No scalar histogram substitute, percentile chart, merged
population or exponential-point profile was introduced.

## 448 — consolidate raw-observation inspection (delivered)

Delivered [compact shared inspection](RAW_OBSERVATION_CONSOLIDATION.md): typed
UTC/range facts, exact native clock attributes, compact Find/paging/inspector
presentation, one scope disclosure and explicit scalar recording uncertainty.
A closed profile map unifies routing/field projection without a gauge fallback;
Go instrument admission now rejects case-variant overrides consistently.

Compared the prior binary with final 448 using all three actual OTLP journeys.
First points appear about 45px sooner on desktop and 122–165px sooner at narrow
widths; first histogram buckets appear 175–474px sooner. Cold point reads retain
one product request/three SQL calls, and local held actions remain zero-IO.
Independent source/a11y/visual review and existing model/actor gates accepted the
result. No source cache, poll change or family-semantic unification was needed.

## 449 — records around a held raw point (delivered breadth)

Delivered: [operator guide](RAW_OBSERVATION_CONTEXT.md),
[real OTLP storyboard](journeys/raw-observation-context.md),
[independent acceptance](reviews/RAW_OBSERVATION_449_ACCEPTANCE.md) and
[validation](../VALIDATION.md). The design basis follows.

Add **Nearby logs** / **Nearby spans** from a selected raw gauge, sum or explicit
histogram point. This is cross-signal investigation, independent of evaluator
readiness or custody. Existing evaluated-gauge navigation is not sufficient:
raw points can have no instance identity or no recorded magnitude.

Plan an ordinary typed record question in a contextual ±60-second event-time
window, clipped to the original discovery window. Preserve exact environment,
service and actual namespace missing/empty/value. Preserve a reported instance;
if absent, make the service-wide time association explicit. Do not copy metric
attributes into record filters or imply causality, contributing transactions,
producer continuity or collection completeness.

Staging performs no IO; Run is explicit. Return restores held read time, page,
Find, detail and keyboard focus without a read. A new read, scope edit or retired
owner invalidates obsolete context actions. Reuse ordinary question/Return
infrastructure through a distinct raw-point capability, never a forged evaluated
gauge result. Gate with real metric→record OTLP journeys, missing-instance and
no-recorded examples, boundary properties, stale controls, model/Bombadil and
narrow-screen checks.

## 450 — independent lane filters (delivered depth)

A [checked design matrix](../verification/gotogether/design/paired-filters.md)
compares shared scope with independent drafts, explicit acquisition and fenced
responses against three interaction alternatives and a retirement control.
The design trial is finite; the implemented feature has its own
[operator guide](PAIRED_ACTIVITY_FILTERS.md), [real OTLP storyboard](journeys/paired-activity-filters.md),
[acceptance review](reviews/PAIRED_FILTERS_450_ACCEPTANCE.md) and
[validation](../VALIDATION.md). The original design basis follows.

Give paired Logs + Spans Activity independent record predicates beneath one
shared environment, service and fixed/live time scope. An operator should be able
to watch error logs beside slow server spans without changing the other lane.
Reuse ordinary typed predicates and the existing visual/CEL editor policy.
Shared scope controls stay singular; each lane exposes its own applied condition.
Keep lane-specific bounds, gaps, refusals and explicit acquisition visible.
Do not imply a merged ordering or atomic snapshot across the two reads.

Gate real OTLP error-log/slow-span populations, changing one lane without
restarting the other, exact missing/empty conditions, Pause/Resume and held
inspection, keyboard/narrow layouts, source work and model/actor controls.

## 451 — single/paired filter consolidation (delivered)

Delivered: [operator contract](FILTER_CONSOLIDATION.md),
[measured storyboard](journeys/filter-consolidation.md), [acceptance](reviews/FILTER_CONSOLIDATION_451_ACCEPTANCE.md)
and [validation](../VALIDATION.md). Shared state and exact condition presentation
are implemented without merging source ownership. The original design basis follows.

After independent filters are measured, unify filter ownership and shared lane
header/condition controls across single and paired Activity. Remove duplicate
staging paths rather than preserve prerelease formats. Check density, accessible
names, progressive enhancement and request counts against the same operator
journeys. Preserve the distinction between shared scope and lane predicates.
Humanize typed duration thresholds (for example, 100 ms rather than 100000000 ns)
in shared applied-condition summaries while retaining exact raw values. Measure
editor/suggestion work and default-screen density before introducing a new cache
or state abstraction.

## Next finite cycle — reported asynchronous references

The [post-451 checkpoint](reviews/PRODUCT_CHECKPOINT_AFTER_451.md) reconciles
independent product/UX and architecture reviews. Choose existing reported span
links as the next general-purpose investigation path. Keep the raw metric plot
proposal for the following cycle rather than start both tracks together.

### 452 — Follow a reported target (breadth)

Delivered: [reported span-link lookup](REPORTED_SPAN_LINKS.md). Its executed
checks are recorded in [validation](../VALIDATION.md); acceptance criteria below
remain the design basis.

Resolve an outside-returned-span link through an explicitly staged target lookup.
Retain the environment policy and fixed event-time window; disclose that service
and namespace restrictions may be broader. Run reads; Back restores the held
origin. Local exact returned targets remain local selections. Admit canonical
trace/span tuples, invalid/missing IDs and duplicates without guessing identity.
Test real cross-trace OTLP links, absent/excluded targets, cancellation and keyboard.
No automatic traversal, expanded-trace/exact-target equivalence or causal claim.

### 453 — Find spans reporting a link to this span (depth)

Delivered: [operator guide](REPORTED_SPAN_LINKS.md),
[storyboard](journeys/span-link-incoming.md) and
[scoped review](reviews/SPAN_LINK_INCOMING_453_ACCEPTANCE.md).

Add bounded inverse discovery for an exact selected trace/span pair. Match IDs at
the same array index with an explicit typed membership selector; independent
array containment is insufficient. Declare matching-record count grain and keep
paging/duplicates visible. Gate crossed tuples, duplicate link entries, malformed
arrays, source absence, window/environment bounds and ClickHouse/reference parity.
A link is a reported reference, not proof of message delivery or processing.

### 454 — Consolidate reference interactions

Delivered: [selector conservation](SELECTOR_CONSERVATION.md),
[scoped acceptance](reviews/SELECTOR_CONSERVATION_454_ACCEPTANCE.md) and
[validation](../VALIDATION.md).

Unify demonstrated reference display, admission, scope disclosure and keyboard
behavior after those workflows are measured. Keep source/navigation owners
explicit and components passive. Do not build a universal investigation owner or
another DSL. Check source work and desktop/narrow operator journeys.

453 exposed manual selector reconstruction across owners. Consolidate two or
three identical pure transformations and add a shared selector-conservation
case bank: exact carry or explicit pre-IO refusal, detached values, and matching
request/response identity. Keep deliberately widened reference/trace questions
explicit and specialized profile restrictions local; no generic policy registry
or readiness cache. See [453 review](reviews/SPAN_LINK_INCOMING_453_ACCEPTANCE.md).

## Retained follow-ons

Presence-safe resource/record string attribute helpers are delivered in 440.
Typed numeric conditions are now delivered in 443 with the existing explicit
presence/validity contract. General arithmetic, coercions and a third-party
lowering registry remain separate proposals.

A merged chronological stream still needs semantics beyond the delivered
two-lane session. Independent lane conditions are delivered in 450.
Saved paired definitions are delivered in 446. Monitoring delivery, multi-user collaboration,
arbitrary executable kinds and external warehouse/storage integration retain
their separate design and operational gates. Existing dashboards, waterfalls,
templates, annotations and captures are delivered capabilities, not new gaps.

The visual cycle now has held raw-gauge Points and one explicit histogram-point
bucket display with shared controls, retained local state and prepared projections.
Preserve recording/value validity, exact counts and source grain; use existing
kind/data contracts, no implicit reads, interpolation or fleet-distribution claims.

## Next visual cycle: 455–457

455 — delivered: [held gauge Points](RAW_GAUGE_OBSERVATIONS.md#points-over-the-held-bank),
[component contract](UI_RAW_GAUGE_POINTS.md), [scoped acceptance](reviews/RAW_GAUGE_POINTS_455_ACCEPTANCE.md)
and [validation](../VALIDATION.md). The acceptance basis follows.

Add a local Points display for held raw gauge observations, using the shared
kind/data grammar. Dots represent source occurrences, with UTC time and exact raw
detail. Preserve duplicate multiplicity and explicit recording/value validity;
invalid, missing or uncertain observations remain inspectable and accounted for.
No connected series, invented instance identity, implicit read or fleet mean.

456 — delivered: [single-point histogram bucket bars](UI_RAW_HISTOGRAM_BUCKETS.md),
with exact counts, source boundaries, recording uncertainty and original raw
inspection. Bars/Table and bucket pages are local, preserving all source buckets.
No observations are merged and no fleet percentile is inferred.

457 — delivered: [raw-display consolidation](RAW_OBSERVATION_CONSOLIDATION.md#display-state-and-prepared-banks).
Shared native display controls, retained histogram pages, cached point content,
immediate Table selection and compact statistics preserve exact source semantics.
Prepared immutable banks avoid repeated full-bank admission during inspection;
synthetic projection costs and actual source-free journeys are recorded separately.

## 458 — delivered: held log/span chronology

[Held chronology](HELD_CHRONOLOGY.md) explicitly pauses both existing tails and
orders immutable held observations with exact event times, stable ties and an
unplaced group. Each lane and occurrence retains its source selector, clocks,
windows, gaps and evictions. Find, pages, shared inspection and held return add
no source reads. View adoption and Analyze handoff guard reentrant replacement
across both global and lane epochs. Source capabilities are absent from passive
display data; existing source owners retain the originals.

This is an optional presentation, not the default entry, a merged source,
evaluated aggregate, saved chronology format, or combined completeness/health
verdict. [Executed journey](journeys/held-chronology.md) and
[validation](../VALIDATION.md) distinguish actual OTLP checks from synthetic
ordering/ownership properties and native model negative controls.

## 459 — delivered: shared polish and guidance assessment

The user-requested [polish](UI_POLISH_459.md) takes priority over the next trace
feature. Shared attribute controls/rows, service cards, keyboard bypass, chart-state
markers and chronology announcements are improved. Independent designer and UX
reviews produced a [content/guidance proposal](CONTEXTUAL_GUIDANCE.md). The first
native Activity disclosure is implemented; a broader tutorial system is not.

## Next proposal: 460 contextual guidance

Migrate repeated generic Services/comparison/chart explanations into consistent
contextual help, retaining current scope, values, clocks, uncertainty, failures and
action consequences. Consolidate duplicate applied/Parent predicate prose while
preserving real parent/slot differences. Then add one route-aware, optional,
replayable first-investigation guide using existing native components. Guidance
must never execute source reads, edit filters or create captures on its own.
This proposal is described in the independent inventories; it is not shipped.

## Following depth step: 461 related held trace context

From a selected chronology record, offer explicit **Related held records** using
validated reported TraceId and optional SpanId across the two held lanes. Reuse
the existing record-context/ancestry components. Preserve repeated observations,
ambiguous repeated span IDs, original lane/window provenance and unknown coverage.
Reported parent links and time adjacency must remain distinct.

Keep this as a local derived scope with explicit Return restoring Find, page and
selection; do not silently rewrite text Find, broaden source scope, infer trace
completeness or start a query. Saving a chronology-specific format, merged
selection/capture and paired baseline attachments remain separate design work.

462 then consolidates the demonstrated record-context projections, interaction
state, accessible controls and performance before the next breadth cycle.
