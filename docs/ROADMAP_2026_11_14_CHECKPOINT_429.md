# Roadmap after measured operator consolidation

The finite 427–429 cycle covers exact definition recall, paired evidence and
measured scope/preservation controls. Executed results belong in
[validation](../VALIDATION.md), not in this proposal's acceptance criteria.
The [architecture checkpoint](reviews/PRODUCT_CHECKPOINT_AFTER_429_ARCH.md) and
[product/UX checkpoint](reviews/PRODUCT_CHECKPOINT_AFTER_429_UX.md) review the
broader product rather than extend another preservation queue.

Status: **429 and 437–444 delivered and validated**. The current cadence is
**443 numeric-condition depth → 444 paired Activity breadth → 445 consolidation**.
Next: **445 single/paired Activity coherence and measured work**. The existing
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

## 445 — consolidate single/paired Activity (planned)

Measure the exercised single/paired scope-edit → Start/Apply → Inspect → fixed
handoff → retained Return workflow at 320/390/1440. Keep one shared control bar and
consistent native focus/refusal behavior without restoring ambiguous global
pair tools. Extract only matching pure scope-admission, viewport/focus and
control-state projections. Retain separate connection, custody and source owners.

Record cold/retained connection counts and per-family SQL work before proposing
caching or polling changes. Include capacity pressure with the existing four-tail
limit, an unavailable family, withdrawn controls and inactive late completion.
Compare the same applied/read state when assessing density or speed. No durable
readiness cache, combined totals, source-health inference or default expansion
of All scope follows. Keep native model/Bombadil/storyboards current; review the
resulting product before adding another narrow language or preservation sequence.

## Retained follow-ons

Presence-safe resource/record string attribute helpers are delivered in 440.
Typed numeric conditions are now delivered in 443 with the existing explicit
presence/validity contract. General arithmetic, coercions and a third-party
lowering registry remain separate proposals.

A merged chronological stream, independently authored lane filters and saved
paired definitions still need explicit semantics beyond the delivered two-lane
shared-scope session. Monitoring delivery, multi-user collaboration,
arbitrary executable kinds and external warehouse/storage integration retain
their separate design and operational gates. Existing dashboards, waterfalls,
templates, annotations and captures are delivered capabilities, not new gaps.
