# Whole-product usability review — 2026-10-25

Baseline: **`9e07b88`**. This is an expert review of committed application source,
existing executable storyboards and their retained screenshots. **No browser was
launched, no application tests were run, and no runtime or application source was
changed for this review.** This document is the only deliverable edited here.
Recommendations are inputs to the revised roadmap, not another implementation
queue or claims of delivered work.

## Judgment

Goo now has enough investigative primitives to make its central promise useful:
start with ordinary telemetry, narrow a population, inspect it, compare evidence
and preserve or reuse an explanation. The next usability gain comes from making
those primitives feel like one product, rather than adding another persistence
format or another specialist workbench.

The strongest recent example is trace search: a contextual operation action,
compact actual predicates, explicit Run, bounded qualified witnesses and a return
to existing trace inspection/comparison. Its first-result distance improved to
464 px through composition, without changing its query semantics. Apply that
approach across the product. Do not reopen the now-delivered underlying features
as if they are missing.

## Evidence and limits

- **Frame observation:** a screenshot actually shows the described arrangement
  in its exercised scenario. Latest 184–186 frames document the delivered trace
  features. Older frames are marked historical below; they are not fresh captures
  of every surface at this baseline.
- **Source observation:** current code constructs the stated controls, ordering
  or continuation. A source-backed structural issue does not establish a newly
  reproduced browser defect or measured task failure.
- **Hypothesis:** likely user effort or a proposed improvement. No customer study,
  task-time comparison, exhaustive keyboard/accessibility audit or production
  adoption research was performed.

Precision-stress fixtures such as the int64-extrema decimal histogram are not a
representative operational workload. Their enormous bounds are useful correctness
evidence, not a reason to conclude that all ordinary distributions are unreadable.
Similarly, the native service frame deliberately exercises unavailable coverage;
it is not evidence that native instruments generally fail.

Reviewed material includes [product essence](PRODUCT_ESSENCE.md), the
[Oct 22 usability review](WHOLE_PRODUCT_USABILITY_REVIEW.md),
[Oct 24 product review](WHOLE_PRODUCT_PRODUCT_REVIEW_2026_10_24.md),
[stream-first design](STREAM_FIRST_WHOLE_PRODUCT_DESIGN.md), current UI owners,
[component contracts](UI_COMPONENTS.md) and storyboards through
[186](../journeys/186-trace-search-storyboard.md).

## Reconcile the older plans first

These are delivered foundations, not proposed next milestones:

| Area | Delivered foundation to retain | Remaining usability question |
| --- | --- | --- |
| Activity | Stream-first entry; compact mobile records; filters, configurable fields, paused inspection, Activity views, completed-window analysis/explanation and error/pattern entrances | How does a person discover the right investigative action without knowing which fixed tool exposes it? |
| Fixed Explore | Full-population counts; grouping/facets; selected/rest populations; typed int64/uint64/decimal values; numeric distributions and comparisons; explanations | Can the result and its next action be understood before learning all the authoring controls? |
| Logs and traces | Patterns across time, passive exception detail, waterfalls, page interval accounting, local trace comparisons and cross-span trace search | Can these tools share an obvious entry/return vocabulary without confusing different read effects? |
| Services | Observed inventory, namespaced identity, default overviews, diagnosis, supported native values and HTTP/database/messaging suggestions | Can the overview foreground the useful answer while moving supporting identity/provenance detail behind disclosure? |
| Operations | Declared behavior, invocation/error/coverage evidence, native views and linked retained spans | Is its primary presentation as readable and consistently named as the generic record flow? |
| Native measurements | Instrument discovery, direct exact-stream questions and typed reset/temporality/coverage diagnostics | Does its landing page lead with observing an instrument, rather than package authoring? |
| Saved and Definitions | Task-specific Saved composition, compact revision-aware libraries, selected detail and explicit continuation | Can the selected artifact's result come before repeated wrappers and competing authoring choices? |
| Authored meaning | Exact annotation targets, contextual tags/notes, revisions, applicability and ordinary same-tab draft recovery | Keep contextual annotation entry small while sharing the action vocabulary; do not make its catalog a prerequisite. |

The Oct 24 recommendations for trace-anchor continuity, ordinary investigation
draft recovery and a concise README have since been delivered. Numeric binding
guidance, publication-winner integrity, service/native coverage presentation and
shared record/component work also exist. Do not reschedule those implementations.
The old Saved/live-Operations reparenting problem has been addressed by separate
owners; current [saved-workspace.js](../ui/saved-workspace.mjs) is task composition,
not the earlier combined live editor.

## Top five issues

### 1. Investigation actions accumulate, while their availability depends on the entrance

**Rank: first. Source observation, supported by current feature frames.**

A fixed attribute can expose Copy, Include, Exclude, Find traces, Plot values,
Filter and run, and Break down and run, depending on supplied capabilities
([attributes.js:101–106](../ui/attributes.mjs#L101)). These are explicit about their
individual effects, but the collection makes the user select among implementation
paths on every field. Numeric selection adds three parallel population/comparison
buttons plus an explanation disclosure
([numeric-distribution.js:31–32](../ui/numeric-distribution.mjs#L31),
[frame 180a](journeys/screenshots/180a-spans-decimal-range.png)).

Meanwhile the newest trace-search owner is wired into fixed Explore
([exploration.js:29](../ui/exploration.mjs#L29)). Activity's inspector has Plot,
facet and grouping continuations, but no trace-search callback in
[stream-entry.js](../ui/stream-entry.mjs). Held service span inspection passes an
empty action set to record detail
([service-browser.js:23](../ui/service-browser.mjs#L23)). This is a code-backed
availability difference, not a fresh browser failure. Plot values already works
across several origins; it should not be incorrectly classified as missing.

**User impact hypothesis:** users learn “go to the fixed page to find the other
operation,” rather than investigating the span they are already looking at.
More field buttons will deepen that problem.

**Bounded change:** establish one supplied contextual action menu with a small
primary set: inspect/narrow, then **Investigate…** for plotting, grouping,
trace-wide conditions and explanation where compatible. Reuse current owners
and validators. Start with fixed and paused Activity record inspection; adopt
held service observations only with their actual supporting-read scope.

**Remove/consolidate:** stop adding a permanent button per new field capability.
Move immediate Filter-and-run and Break-down-and-run alternatives into an
explicit secondary action group; do not silently change their effects. A single
Compare choice can reveal count versus declared numeric measure, instead of
parallel planner-oriented buttons.

**Gate:** the existing qualified field, original observation window and owner
epoch survive each handoff. Staging stays local; named Run/Start/fresh-read actions
retain their real effects. Unsupported native/field cases have a concise reason.
Keyboard/touch access and opener focus remain available. This is a small action
composition, not a new global command engine or shared mutable scope.

### 2. Evidence scaffolding still competes with the answer

**Rank: second. Historical frame observations plus current construction.**

[Saved frame 117a](journeys/screenshots/117a-selected-saved-evidence.png) shows a
single retained record below several action, frozen-scope, receipt, count and
paging blocks. That is an older frame, but current
[index.html:14](../ui/index.html#L14) and [captures.js](../ui/captures.mjs) still
compose an outer capture header/receipt with a full inner result owner. Multiple
places expose exact scope/evidence details; they are not all needed above the
first useful record.

[Native-service frame 75a](journeys/screenshots/75a-native-service-overview.png)
shows repeated coverage explanations around per-stream values. Current
[service-dashboard.js:13–24](../ui/service-dashboard.mjs#L13) retains summary,
stream table, coverage text, retained-native detail and widget-level context.
Its semantic refusal to invent a fleet value is correct; the repeated framing
can be compressed without removing the unavailable state.

[Operations frame 163](journeys/screenshots/163-readable-operation-events.png)
shows severe word fragmentation on mobile. Current
[operations.js:105–113](../ui/operations.mjs#L105) still emits a plain seven-column
overview. This is a source-supported presentation concern requiring a fresh
bounded layout check before calling it a current browser regression. The primary
button still says **Save investigation** while its title and handler concern an
operation evidence capture ([index.html:12](../ui/index.html#L12)).

**User impact hypothesis:** the product makes users repeatedly prove they
understand the evidence model before showing what they came to inspect. A label
also conflates a capture with a narrative investigation.

**Bounded change:** adopt an answer-first result composition in two consumers
first: selected Saved evidence and Operations. One compact mode/scope/state
header, one primary result, relevant continuation actions, then one exact-proof
area. Use the existing typed tables, evidence summaries and inspector shell.
For Operations, choose a readable row/card representation with selected detail
rather than squeezing all seven fields into mobile columns.

**Remove/consolidate:** merge repeated wrapper counts/provenance into the single
owning summary; move long contract prose and raw hashes into exact details. Rename
the operation capture action by its real object/effect. Do not remove decisive
partial/unavailable/coverage facts from the affected result.

**Gate:** different source/read/custody clocks remain distinguishable; unknown
never becomes zero or healthy. Native exact-stream identity, units and reset
semantics survive. Saved reopening stays source-free. A small representative
record/capture should show its primary answer in the initial viewport; measure
that in an actual refreshed journey, not by assuming a universal pixel target.

### 3. Scope and return are correct locally, but not one obvious product interaction

**Rank: third. Source observation and navigation-frame evidence.**

Activity maintains its paused return anchor
([stream-entry.js:75–79](../ui/stream-entry.mjs#L75)); fixed questions use an
investigation trail; service continuation adds its own Back-to-service bar
([service-browser.js:28–33](../ui/service-browser.mjs#L28)); saved work can add an
active-investigation return banner. [Frame 117a](journeys/screenshots/117a-selected-saved-evidence.png)
shows that additional banner above the selected evidence. These owners have
legitimate different lifetimes, but their presentation can compete.

Time has improved: Saved hides the new-query lookback, and Activity hides the
outer top bar. Do not re-raise those old issues as unfixed. However,
[workspace.js:9](../ui/workspace.mjs#L9) still shows **HTTP preview lookback** across
Definitions, which contains record, measurement and application definitions,
not only the HTTP card. Native discovery and measurement bounds are explicitly
separate in [measurements.js:94–101](../ui/measurements.mjs#L94).

**User impact hypothesis:** users need to infer which scope or Back control owns
the current view, and whether a time input is relevant to the selected object.

**Bounded change:** define a compact parent/return presentation and local
mode/time header. Show the active context's one primary return destination,
while secondary ancestors remain a breadcrumb/disclosure. Put time authoring
inside the owner that consumes it; move the HTTP lookback into its HTTP preview.

**Remove/consolidate:** competing visible return bars and irrelevant global time
inputs, not retained visits or independent scope owners. Keep Close for the
frontmost inspector distinct from returning to its parent investigation.

**Gate:** Activity returns paused with the original selection; service return
preserves service/namespace; Back does not run telemetry. Existing direct links,
explicit Run and focus restoration keep their contracts. Do not replace these
owners with one global scope store or erase pending-save/recovery state.

### 4. Native investigation still introduces authoring before observation

**Rank: fourth. Current source observation; older frames are supporting context.**

The Measurements landing says **1 Define → 2 Preview → 3 Investigate**, starts
with a package picker and exposes time/preview/run controls.
**Discover metrics from an application** is an initially closed disclosure
([measurements.js:93–95](../ui/measurements.mjs#L93)). The capability already exists;
the issue is its ordering and mental model, not a missing metric backend.

Definitions also presents separate record, measurement, application and bound
HTTP sections ([index.html:17](../ui/index.html#L17)).
[Frame 110a](journeys/screenshots/110a-selected-definition-revisions.png) is a
historical example of the lengthy combined landing. Its compact revision rows
and selected details are progress, not something to rebuild. Annotations are
already a local disclosure with exact targets
([meaning.js:17–31](../ui/meaning.mjs#L17)); they should not become a prerequisite
for opening an instrument.

**User impact hypothesis:** an engineer with emitting OTel metrics may believe
that they must design and save a package before examining a counter or histogram.

**Bounded change:** let Measurements start with **observe an instrument** or
**use a saved definition**, carrying service/environment context when available.
Reuse the existing discovery and exact-stream question path from service views.
Create/revise a reusable definition becomes a subsequent explicit task, with the
same required type/unit/temporality checks and plan preview.

**Remove/consolidate:** the mandatory-looking three-step authoring banner on
ordinary instrument entry. Keep one family-aware definition picker where useful,
not a second copy of the full Definitions editor. Do not invent a global catalog
search claim over merely loaded revisions.

**Gate:** this works without a manifest or an initial save. Unsupported metadata
stays explicit; a metric name alone cannot become an arbitrary fleet aggregate.
Import/definition/YAML expert workflows remain reachable, and selected immutable
revisions remain exact rather than title-based “latest” guesses.

### 5. Shared charts exist; their selection and comparison composition is not yet equally shared

**Rank: fifth. Source observation and latest frame contrasts.**

The current visual foundation is substantially better than the old review:
[184a](journeys/screenshots/184a-observed-page-latency.png),
[185a](journeys/screenshots/185a-two-observed-trace-pages.png) and
[186c](journeys/screenshots/186c-matching-trace-page-comparison.png) show physically
shared timing scales and readable exact differences. Counts/rates/shares and
numeric signed comparisons already exist. Do not schedule another generic chart
library to replace them.

What differs is the surrounding interaction. Numeric selection offers separate
Run/Stage-count/Stage-numeric actions
([numeric-distribution.js:31](../ui/numeric-distribution.mjs#L31)); trace comparison
has its own Pin/Compare/Clear tray plus long reference descriptions
([record-results.js:23–29](../ui/record-results.mjs#L23)); service time selections
can reread the overview through their adapter, while a dashboard interval link
stages a fixed question ([service-browser.js:78](../ui/service-browser.mjs#L78),
[service-dashboard.js:59](../ui/service-dashboard.mjs#L59)). Native interval
columns deliberately need not represent equal elapsed time. These are different
valid semantics, but visual similarity alone does not teach their effects.

**User impact hypothesis:** a person can understand a bar but still hesitate over
what selecting it will read, compare or replace.

**Bounded change:** a shared selection/comparison composition with visible named
sides, selected population/window, measure and next action. Keep execution in
its typed adapter. Start with the two count/numeric selection trays; then reuse
presentation in another real consumer rather than designing a universal option
matrix. Offer one Compare entry with measure choice when compatible.

**Remove/consolidate:** parallel baseline/measure action rows and repeated
interaction instructions. Keep a short discoverable gesture hint, keyboard and
exact-table fallback. Use compact human values with exact strings in details;
no arbitrary rescaling or denominator changes.

**Gate:** local preview gestures and toggles stay source-free; a read-producing
activation states its effect. Existing one-click explicit reads do not acquire
unnecessary confirmation steps. Invalid,
empty, missing and unknown remain distinct. Stored, fixed and live sides keep
their own clocks. Native bucket compatibility, unequal-width bins and interval
semantics are not flattened into a generic percentage chart.

## Ranked next outcomes for the roadmap

These are proposed outcomes, not authorized implementation or a competing queue.
Each needs its own bounded acceptance and review before the next proceeds.

1. **An observation offers one understandable next investigation.** Consolidate
   fixed/Activity field actions and expose compatible existing tools through a
   contextual chooser. Start with one real paused span journey; preserve scope,
   effects and focus. This is the most direct adoption improvement.
2. **The answer appears before its scaffolding.** Adopt the compact result
   composition in Saved evidence and Operations, including the capture label and
   mobile operation table. Reuse existing components, do not build new stores.
3. **A person always knows which scope and parent they are in.** Standardize local
   mode/window/return presentation across one Activity→fixed→trace path and one
   Services→analysis path; contextualize unrelated time inputs.
4. **Already emitting native metrics can be investigated before authoring.**
   Promote existing discovery/direct-stream entry and a shared selected-definition
   picker while preserving semantic validation and YAML workflows.
5. **Selection and comparison feel consistent across typed results.** Adopt one
   display/action composition in two real consumers, retaining separate source
   owners and the current passive Web Components.

The first two can incorporate small naming/layout fixes without a large redesign.
Do not bundle all five into one sweeping rewrite. The maintainable unit is a
user task with two actual consumers, not another framework or a universal result
schema. Native/source semantics, cancellation, evidence custody and exact numeric
values are invariants throughout.

## Why saved trace search should not automatically be first

Trace search is currently ephemeral, and a durable reusable request may eventually
be useful. That is a real boundary, not proof that another store is the next
highest-value feature. There is no new user study or measured abandonment evidence
here showing that lost search configuration outweighs finding the action,
understanding the result and returning to the original observation.

First make the existing observation→conditions→witness→trace route discoverable
and compact across compatible entrances. Keep a request in an existing retained
visit where that has a concrete user benefit; do not confuse that with durable
publication. Only then select a specific persistence task: **reuse a versioned
query**, **preserve a historical observed result**, and **recover an unfinished
draft** have different contracts and should not become one “Save search” button.
A future proposal must state which object, read/custody clock, version and
failure/retry behavior it actually preserves.

Startup/adoption work should similarly resolve a concrete first task: configured
source plus correct environment, useful available signal, inspectable activity
and a visible next action. An empty stream is not proof of a broken collector;
provide an explicit next diagnostic/discovery path without a fabricated health
indicator. This review does not authorize a connection wizard, ingestion-health
backend or another onboarding product.

## Whole-product acceptance lens

A later implementation should refresh actual desktop/mobile journeys rather
than relying on the historical frames cited here. Measure first useful result,
visible competing actions, return behavior and read counts. Include ordinary
non-Go OTel, an empty/unsupported case, a native stream needing review and one
saved exact revision. Preserve keyboard alternatives, focus, no-overflow and
source-free inspection/staging. Use existing property/state/Quint/Bombadil
coverage for owner/scope effects; do not replace it with screenshot snapshots.

Keep the framework-neutral passive components. `GooUI.inspector`, typed values,
evidence summaries, tables, time bars, waterfalls and comparison geometry already
provide the building blocks. Share compact compositions and capabilities between
concrete adapters. Source fetching, stream subscription, navigation, query
planning and publication must remain outside display elements.
