# Whole-product journey checkpoint — design 347

Status: proposed priorities after 332–346; no feature below is claimed delivered.
Final 346 visual acceptance remains in its own review and takes precedence.

Goo's strongest product story remains: watch ordinary OTel activity, narrow a
population, inspect traces/logs, compare it, then preserve the useful question or
explanation. Recent work makes that story more trustworthy: correct operand roles,
typed policy continuity, mean duration, retained Back, full-literal facets and
recoverable source refusals. The next round should make that story more useful
across surfaces, rather than further polishing the same facet controls.

## Current source-grounded capabilities

Activity is already the rolling entry, with local Find, loaded-record selection,
time brushing, service context and rolling Activity views. Population is the
fixed-window analysis workspace with ordinary count/mean comparison, typed
filters, paired values, exact supporting records and retained child return.
Trace search, waterfall and log context now share refusal/recovery semantics.
Preserve already distinguishes question, reusable view and independent evidence;
Activity views intentionally differ from fixed exploration views. These are not
missing features to rebuild or rename into a universal manager.

The current prose and settings can still dominate screenshots, but cropped
interaction frames are not a whole-entry density study. The 343 roughly
32,000-element paired tree also remains a real limit; the 346 publication repair
must not be sold as solving all large-data performance.

## 1. Explain a fresh comparison from its natural location

**User question:** “Traffic or mean latency changed. Which populations changed?”

`ui/explanation.mjs` already implements bounded multi-field explanation ranking,
checked requests, side filtering and observed-result custody. It is mounted from
`ui/distributions.mjs` and Activity explanation, while
`ui/population-workspace.mjs` does not mount it for its authored fresh pair. That
leaves a gap between the newer natural comparison workflow and an existing useful
investigation capability.

First propose a contextual Explain differences action on an applied fresh pair.
Reuse both whole original questions and the existing explicit field/read contract;
keep ranked share/volume differences distinct from explaining mean duration or
claiming causal drivers. A user follows a ranked side into records and Back returns
to the held explanation/pair. Preserve observed explanation using existing custody,
without rerunning it. No automatic live anomaly scanner, new ranking engine or
blanket scoring promise is selected.

Acceptance: genuine Activity → fixed comparison → explicit explanation → chosen
side records → Back → save/reopen observed explanation, with exact typed policies,
separate read clocks, partial/unavailable handling and bounded source counts.
Confirm endpoint admission against count versus duration questions before coding.

## 2. Let one service remain inspectable in a large environment

**User question:** “What is this service doing, regardless of how many unrelated
services share its environment?”

Architecture's independent source review finds `service-browser` inventory reads
the environment's service/operation pairs before filtering the chosen service in
the browser; `InventoryQuestion` has no service scope. An unrelated high-cardinality
environment can therefore refuse the inventory for a small selected service. The
natural Activity → observed service journey should not make users reason about
that unrelated cardinality.

Propose exact service/namespace-scoped inventory acquisition and corresponding
count/capture scope before grouping, preserving the typed environment policy and
independent source receipts. Reuse the existing service browser and dashboard;
no new discovery engine, cap increase or dashboard catalogue is needed. Keep
whole-environment discovery available through its existing intentional scope.

Acceptance: ordinary observed service entry in an environment whose unrelated
operations exceed the old bound; chosen small-service inventory succeeds with
exact namespace/service and policy proof. Full-environment refusal remains honest,
service capture cannot quietly store unrelated inventory, and Back retains the
paused Activity. Architecture/planner should close the schema/source contract
before implementation.

Rolling Activity views and fixed exploration templates already have explicit reuse
paths. The earlier possible fixed-to-rolling handoff is therefore deferred until a
real journey demonstrates a missing gesture; do not build another saved abstraction
from terminology alone.

## 3. Author a grammar package from the scopes users actually use

**User question:** “This uninstrumented service has useful telemetry. Can I turn
what I found into a reusable definition?”

Architecture finds the ordinary exploration Create package action disabled whenever
an environment selector is typed, while package definitions remain scalar-scoped.
That blocks the mainstream Activity/Services policy scopes used by recent work
from the user's manifest-free grammar-authoring goal. It is a more concrete
consolidation gap than adding another saved-view abstraction.

Propose typed scope parity through the existing package-authoring path: preserve
checked environment policy and service/namespace meaning in authoring, preview,
save and exported declarations. Reuse current editor/package resources; no full
WYSIWYG replacement, new package catalogue or annotation-generation promise is
selected here. Unsupported bindings need an explicit refusal, never projection to
a weaker scalar environment. The prerelease permits simplifying the active schema
rather than maintaining two divergent meanings.

Acceptance: genuine Activity → scoped service/population → existing Create package
→ preview/save/reopen, with exact typed policies and no hidden telemetry reads.
Use actual entry/preservation frames to confirm that the useful data remains ahead
of settings and that scope/read meaning is concise. Confirm generated annotations
or backend lowering only where current contract actually supports them.

Persistent investigation branches and explicit saved paths remain useful later
candidates, but require a custody/schema-overlap contract. Same-tab draft recovery
and retained local history do not already provide a durable multiplayer whiteboard;
do not imply that they do or start that infrastructure in these three cuts.

## Sequence and stop conditions

Recommend breadth item 1 next, then the service-inventory depth item 2, followed by typed package-authoring scope parity as consolidation, informed by
actual end-to-end entry and preservation frames.
Architect/planner must reconcile these proposals with the existing backlog and
remove overlap. Do not chain implementation automatically from this review or
turn it into Datadog parity, new storage infrastructure or another facet-only loop.
