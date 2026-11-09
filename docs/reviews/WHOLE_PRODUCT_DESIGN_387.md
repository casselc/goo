# Whole-product design checkpoint — 387

Read-only review, November 9, 2026. Milestones 385–386 are delivered; 387 is an
implementation candidate with source/generated checks while actual acceptance is
still pending. This review ran no source queries, browser campaign or heavy tests.
It inspected the current capability index, adopted 384 roadmap, recent reviews,
current UI/planner code and actual 385/386 frames. These recommendations are inputs
to the next root-owned roadmap, not another active backlog.

## Preserve the product center

Goo is strongest when a user can see telemetry, isolate a population, follow an
observation, and preserve or reuse the resulting question. Keep ordinary OTel data
useful without a manifest. Keep the live log/span firehose as the primary entry;
this checkpoint does not justify another home dashboard, builder or navigation
framework. A representation change should shorten an investigation, not become a
new source executor or confer new evidence authority.

The recent sequence advances that center:

- **385 breadth:** outgoing HTTP observations now fill a real service-dependency
  gap, without pretending a hostname identifies an instrumented callee or that
  Client-span counts are unique requests.
- **386 depth:** one native-view owner supports gauge and histogram reuse while
  preserving the complete destination question. Missing values now occupy a
  time-aligned rail outside the numeric scale.
- **387 consolidation candidate:** full captures, delivery pages and receipt-only
  references are separate boundaries. The six families keep their own clocks,
  paging, result semantics and historical-plan policies.

Do not recreate existing functionality as a new vendor-inspired category. Live
Activity, filters/facets, selected/rest and time comparisons, ranked explanations,
log patterns, trace search/waterfalls, service/resource views, captures and authored
views already ship. Captured-record review and exports also exist; training or
annotation work does not automatically outrank a basic investigation gap.

## Proposed next finite cadence

### 388 — breadth: find records around a gauge observation

**User journey:** an operator sees a goroutine, queue or memory gauge change,
selects the recorded sample, and asks for logs or spans near that observation.
They inspect a record, optionally follow its reported trace, and return to the
same held measurement and selection.

This is a concrete current gap. The pure
[`PlanNativeRecordContext`](../../exploration/native_record_context.go) can express
only `[atNS, atNS+1)` for a gauge. The UI's
[`recordContextPlans`](../../ui/measurements.mjs) deliberately withholds gauge
navigation and says the observation has no interval. Existing interval context
is implemented for counters/histograms; it should not be presented as missing.
An instantaneous point has no measurement interval, but an explicitly chosen
surrounding-time record question is still useful.

Start with service-bound per-stream gauges, one explicitly selected reported
sample and a small named lookaround radius. Stage the existing typed record
question; only Run reads. Clip the proposed time window to the parent fixed
window, show the anchor/radius/applied bounds and preserve exact namespace,
environment policy, selected service and reported instance. Metric-only dimensions
must not silently become record filters. Missing/empty namespace stays exact.
Any widening beyond the instance or parent bounds is a separate visible edit.

Use a distinct surrounding-point context contract rather than relaxing the existing
one-nanosecond point or interval contract. A stale sample keeps its recorded time;
an unknown sample can offer time context without claiming a numeric spike. The
result is coincident evidence, not causality, an exemplar relation, a metric-to-log
join or proof of complete trace coverage.

**Acceptance:** generated raw-timestamp membership and Go/JavaScript scope parity;
actual gauge → stage → Run → record/trace → held return; zero reads on selection,
radius editing and return; exact clipping near either bound; other-instance and
namespace controls; known-empty, stale and unknown cases; original native/capture
citation and bytes unchanged. Exercise the visible completion with Bombadil and
replay measured read/history counters. Do not infer success from the click.

**Boundaries:** start with the ordinary Measurements owner; Saved and service-card
entry adoption must use their existing retained owners and can be separate cuts
if that is necessary to keep the first acceptance finite. Resource-bound gauges,
stocks and automatic cross-service correlation are not implied.

### 389 — depth: reuse a counter interval display

**User journey:** an operator has a useful counter rate table/timeline and applies
that display to another compatible counter question without inheriting its data,
stream, time range or reset history.

Current counter evaluation and interval rendering already exist. The shared
[`NativeView`](../../workbench/native_views.go) contract only admits histogram and
gauge arms; [`native-view-declaration.mjs`](../../ui/native-view-declaration.mjs)
likewise has no counter declaration. Extend this owner rather than introduce a
third display manager or label a counter as a gauge.

Use an explicit counter interval/rate arm, beginning with a bounded supported
per-stream service domain. The declaration must identify the representation it can
consume, and applicability must retain exact units and supported calculation-domain
semantics. The entire validated destination question wins, including its gap,
source and exact-stream policies. Do not copy donor intervals, reset states,
predecessors or a selected export. No browser rate calculation, fleet summation,
interpolation, implicit conversion or rounded large-integer claim is warranted.

**Acceptance:** pure declaration/store round trips, generated compatible/refused
reuse, exact destination request and compile identity, source-free Save/Browse/
Inspect/Apply, explicit destination Run, reset/gap/unknown intervals and real zero
rates retained distinctly. Use the existing supplied-data timeline/table and
unavailable rail. Existing interval-to-record actions must keep their scope and
held return after configuration reuse. A real cumulative/delta fixture and
independent raw counter oracle should complement the model and actor checks.

**Boundaries:** this is reusable presentation of already evaluated rates, not a new
counter evaluator, alerting system or aggregation policy. Resolve the exact
supported domain in architecture/design review before selecting public wire names.

### 390 — consolidate: make native evidence lead the native screen

**User journey:** after an explicit native Run or opening a native capture, the user
sees the chosen chart/table and a compact source/scope/state summary immediately.
Detailed identity, diagnostics, authored definition and save forms remain available
when needed. A selected missing observation remains visibly missing.

384 established this data-first contract for fixed record surfaces, not every
native page. Current
[`measurements.render`](../../ui/measurements.mjs) places stream identity,
diagnostics and Save controls before the visual. The actual 386 narrow frame is
useful after scrolling to the chart; it is not evidence of natural-top native
entry. Reorder shared native composition rather than shrink text or build another
page owner. Keep active failures, unavailable/limited states and dirty editing
visible; source/custody clocks remain accessible without repeating prose before
all primary data. Show the save-title form only when preserving configuration;
a stale default title such as "Gauge table display" should not compete with a
Timeline selection in the primary inspection flow.

The same cut should scope the owning layout styles so primitive geometry is not
changed by application-wide element rules. Current
[`product.css`](../../ui/product.css) still has broad `section,.card` and nested
section rules. In 386 those rules physically shifted a newly composed time rail.
The local fix is delivered; global style ownership still merits bounded cleanup.
This is a layout concern, not a reason to replace framework-neutral light-DOM
components or add another UI framework.

**Acceptance:** matched natural-top before/after geometry in the actual scroll
container at desktop and 390px; a complete first meaningful chart region or table
row plus scope/actions visible, no page pan, no scripted scroll standing in for
entry. Cover live native, Saved native and service-card native composition. Retain
standalone/SSR and nested component geometry, numeric/unavailable time alignment,
keyboard focus, current-owner guards, warnings and exact selected marks. Definition
editing and Save disclosures read nothing; retained return preserves DOM/citation
without rereading. Preserve source questions, counts, clocks and artifact bytes.

Keep service ordinary-widget versus contextual Back unification as a later bounded
usability task if it does not fit this shared native renderer. It should not
silently turn this cut into a whole-navigation rewrite.

## Reconcile the previous proposals

| Previous candidate | Checkpoint disposition |
| --- | --- |
| Outgoing HTTP observation lens | Delivered in 385; not an open breadth item. Native client-duration binding remains separate. |
| Native gauge display reuse | Delivered in 386; no second gauge catalog is needed. |
| Full capture-family closure | Current 387 acceptance; do not repeat as a new platform or schema project. |
| Contextual review-task chooser and kept-occurrence review | Coherent later pair, but existing annotation/review already works. Lower immediate priority than native observation follow-through. |
| Log-pattern display grammar admission | Still a valid bounded alternative depth cut. Patterns/counts/coverage/execution are delivered; the missing declaration should not become another clustering engine. |
| Counter display reuse | Promote as proposed 389 because the shared native owner and uncertainty rail now exist. |
| Native density and style isolation | Promote as proposed 390 based on actual 386 composition/geometry evidence. |
| Service exemplar adoption / native HTTP-client metric profile | Independent later facet work, with separate citation or source-semantics admission. |
| General exponential evaluation | Still gated by source/mathematical fidelity, not a convenient replacement for an available task. |
| WYSIWYG/executable kinds, alerts, broad DB/broker health, authenticated multiplayer, warehouse/new stores | Wider initiatives; no prerequisite for this finite sequence. |

The root planner should adopt one queue after independent architecture review and
387 actual acceptance. If runtime work finds a more serious core gap, revise that
queue rather than layering these proposals alongside another active backlog.
