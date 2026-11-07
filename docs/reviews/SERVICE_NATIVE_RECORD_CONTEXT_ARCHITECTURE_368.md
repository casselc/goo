# Service native interval context — architecture 368

Status: implemented; final source and runtime acceptance passed. The readiness
contract below records the integration boundaries reviewed before implementation.

## Existing authority and integration seam

`service-dashboard.mjs:renderNative` lazily renders the held result through
`Measurements.render`, but supplies no `onRecordContext`. The exact interval
planner already exists in `native-record-context.mjs`; saved-origin validation in
`saved-native-record-context.mjs:admitHeld` demonstrates how to verify that the
selected stream and export bounds belong to retained evidence. Reuse those
semantics, not a chart bucket or newly inferred stream.

The Measurements-native navigator is not suitable unchanged: its visible-origin
check requires an opener under `measurementWorkbench`, and its Return activates
the measurements workbench. The actual origin here is the retained service card.
Likewise, bare `goo:open-population {question,stage:true}` is not an exact-record
staging API: `investigation-trail.mjs` only propagates stage for its specific
rest/view/template arms, and enters population presentation.

## Bounded closed contract

Expose existing histogram interval Logs/Spans actions inside Native result.
The service dashboard supplies a creation-time action binder and `canAct` using
its current load/data owner. Only an admitted actual held histogram series/export
interval qualifies. Recompute the existing planner result and compare the complete
record question before the service handoff. No gauges, stocks, aggregated chart
time buckets, invented record attributes, metric-label predicate projection or
exemplar/causal attribution is added by this cut.

The service-browser callback checks its captured generation, dashboard revision,
request signal, selected service/result identity and visible connected opener.
Reject while any overview request/retry remains pending: existing workspace
suspend intentionally invalidates pending service work, which would destroy the
promised held origin. This is a source-action admission requirement, not an
implicit cancellation/restart of unrelated widgets.

The app should reuse `openRetainedRecords(question, source, options)` from 364
with a service-origin source token. It stages the exact fixed record question,
retains the actual `liveExploration.visit()` and exposes Back to service. Explicit
Run alone acquires records. The child's source may refuse or be cancelled without
altering the held native result. Return first suspends/releases that child visit,
then activates the original service owner and restores connected opener/focus and
scroll. Existing service `returnPath`/`returnToService` may provide presentation,
but must use this actual child-visit cancellation rather than only clicking a tab.
No new source planner, session manager or durable child-navigation format is needed.

Separate origin validity from action visibility: the service is hidden while the
child is active, so the held token checks generation/data identity rather than
requiring a visible service panel. Conversely every departure action checks its
opener's current visibility. Service replacement, overview rerun, disposal or
another origin invalidates the token; saved callbacks cannot revive after ABA.
Normal unrelated navigation must not expose a stale service Return.

## Evidence language and acceptance

Label the origin “From service native interval”, not saved/captured evidence.
`saved:true` in a passive renderer is presentation, not durable custody. Retain
native basis/read clock, exact stream metadata and selected interval locally;
there is no capture receipt or store read. Logs/spans are coincident service/
namespace/instance records within exact nanosecond bounds; metric dimensions stay
out of record predicates. Unknown coverage is not contribution or a healthy zero.

Focused tests should establish exact staged question and zero telemetry at Stage,
one explicit child read, retained service DOM/disclosures/interval/clocks on Back,
zero native/store reads on Return, old handlers inert and pending/failed child
completion unable to adopt after Return. Include typed policy/namespace and ensure
whole native chart bounds cannot replace the selected export interval. A genuine
service entry is required because isolated Measurements tests miss the host and
workspace-departure mismatch described above.

## Implemented integration

The candidate uses a dedicated internal `goo:service-native-record-context` event
and `openRetainedRecords`, rather than the population event. The service browser
wraps the held source in generation/revision/context guards and refuses both
pending-open and pending-dashboard departures. Back suspends the actual child
visit before restoring the existing service result. Unrelated departure releases
the return token. No new receipt or capture authority is asserted.

`native-held-record-context.mjs:admitHeld` extracts the existing saved interval
admission without changing its semantics; the saved adapter imports it. The
service renderer enables it only for histograms. Each native renderer owns its
anchor token and retains its own frozen result copy. Stage retires old handlers;
Return rebinds current anchors. Service-browser suspension also rotates these
local tokens, so ordinary hide/show cannot revive a saved callback. The bounded
local owner list is cleared on overview replacement/invalidation; it is not a
global reference registry. Syntax and diff checks pass; final integration/runtime gates are recorded below.

## Final source and verification review

Final source is accepted with no material blocker. Six focused mounted/admission
checks exercise the actual renderer, frozen held values, optional typed policy/
namespace context, exact exports and creation-time callback leases, including
ordinary departure ABA and lazy controls. The generated cross-runtime check also
uses independent exact-integer boundary/foreign-instance membership, rather than
using production validation as its only oracle.

The observed adapter projects actual recorded fact names and cumulative source
counts into the existing source/depth model. The combined pending Stage/Run macro
is explicitly split logically at its recorded before/after counters, not presented
as a captured intermediate snapshot. Source/order corruption controls go directly
to the model rather than being rejected only by Go admission. Raw membership,
held evidence identity and cancellation remain separately checked by browser/Go
oracles; model acceptance alone does not prove those. Actual native acceptance passed in approximately 1.3 minutes (runner 1.4
minutes), including successful/refused/pending child reads and stale actions.
Observed Quint replay passed in 5.25 seconds with zero skips; direct source/order
mutants failed as intended. The positive actor completed in 45.145 seconds with
823 snapshots across 11 phases, 12 matching logs and held origin/focus restored.
Root also reports full Go (cmd 8.937 seconds/workbench 1.491 seconds), 1,263 UI
checks in 17.531 seconds and Hegel 40 cases in 9.60 seconds passing. These checks
validate the bounded histogram context flow, not exemplar causality, fleet
completeness or durable browser-reload navigation.
