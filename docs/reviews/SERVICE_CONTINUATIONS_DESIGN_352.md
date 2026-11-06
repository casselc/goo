# Typed service continuation and reuse — design review 352

Status: implemented source accepted; actual-frame review pending. The initial
contract below records the source gap before this change. Exact typed reported-edge execution remains the separate 353 task.

From discovered distributed services, users should be able to inspect a wider
child-service population and find saved rolling configurations under the same
checked policy. Unsupported exact-edge analysis must not prevent those useful
paths from rendering or silently weaken the environment selection.

## Source findings

`ui/distributed-services.mjs` currently builds both exact-edge and wider-child
questions with scalar `environment: scope.environment`. The exact edge question
cannot currently admit a typed environment selector, so eagerly constructing its
link may fail before a useful wider continuation is rendered. This cut carries
typed fields to the supported wider question and explains the exact-edge limit
before constructing an unsupported link; it does not claim backend support.

`ui/service-browser.mjs` gates the saved Activity library on scalar native support,
while `ui/activity-views.mjs` already accepts a typed scope object. Reuse that
existing configuration library with its checked environment policy. Native/entity/
operation scalar restrictions retain their own contracts; do not remove every
nearby guard.

## Concise action meaning

Keep distinct visible actions:

- **Investigate reported edge** for currently supported exact-edge questions.
- **Open child service spans (wider population)** for the wider service scope.
- For unsupported typed exact edges, a non-link explanation such as **Exact
  reported edge unavailable for this policy**. No broken link, silent omission or
  disabled mystery control is needed.

The wider action stages the exact child service/namespace and original fixed
window/policy, without a reported-edge predicate. It does not promise that every
child span belongs to the displayed parent relationship. Existing explicit Run
reads supporting records. Record counts remain captured child spans, not requests,
causal flow or complete traces.

Keep a short scoped library heading/note: saved Activity views use this environment
policy, while their own filters may target other services. Opening stages a
configuration; explicit Start reads telemetry. Do not silently inject the selected
service into a saved selector, call every result a service match or carry the fixed
discovery window into a rolling definition. Existing library search/revision
inspection is metadata-only and local search covers its loaded page.

## Identity, state and return

Preserve named/effective/current-only/empty/unreported policy exactly. Child service
and namespace missing/empty/literal distinctions remain explicit. A wider question
must not fabricate a scalar name from a display label or drop empty identity.
Original discovery diagnostics and read clocks remain held and visible.

Back returns to the current discovered-service/distributed context with usable
controls, local Find and focus on the current visible semantic opener. Saved-view
opening respects existing dirty-destination acceptance; a late metadata response or
old handler cannot replace a newer scope or start telemetry. Disclosure, local
Find, link staging and Back have no replacement source reads. Unsupported exact
edge explanation is local admission, not a source outage or no-data observation.

## Acceptance

Genuine typed discovery renders both graph/table and useful wider action despite
unsupported exact edges. Follow wider child → staged fixed question → explicit Run
→ Back, verifying exact policy/child namespace/window and source budgets. Include
current/legacy conflicts, empty/unreported policies and a wider-population contrast.
Check scalar exact-edge regression without claiming it establishes typed support.

From a typed service visit, inspect a saved Activity definition from the same
policy whose saved filters differ from the selected service. Open it and prove
those filters stay unchanged, then explicit Start owns its rolling telemetry read.
Metadata catalog absence/partial failure must not invent telemetry absence.
Actual desktop/narrow frames should show clear exact-versus-wider affordances,
readable scoped library and current return focus without more verbose controls.

## Library source acceptance

The service library now receives only shared checked environment fields from the
active scope. It injects no selected service, namespace or fixed bounds into saved
Activity definitions, and leaves unrelated native/entity guards intact. Mounted
cases exercise all four typed policies, preserve a saved view's own different
service predicate, distinguish metadata from telemetry and retire late responses.
This one-line scope adoption is accepted.

For service-context clarity, prefer the existing scoped library note to say
“Exact environment policy; saved filters may target other services. Opening stages;
Start reads telemetry” and the scoped heading to name the environment policy. This
is reuse of existing copy, not a new paragraph, API or service-matching feature.
Distributed continuation source and actual-frame acceptance remain pending.

## Frozen complete source acceptance

Wider child questions now use shared typed environment fields and ordinary query
validation while retaining child namespace/service and fixed bounds. Exact typed
edges are refused before a step link is constructed and rendered as a local
non-link explanation; scalar exact-edge behavior remains. The saved library now
explicitly names environment policy and warns that saved filters may target other
services. Embedded Activity library wording is unchanged.

Complete source acceptance passes. In actual frames, check spacing between the
inline exact-edge explanation and wider link, plus long table behavior at narrow
width; the source alone does not establish their visible separation. No new
source execution or typed exact-edge capability is claimed by this cut.

## First actual distributed frames

Reviewed `352-distributed-1280.png` and `352-distributed-390.png`. Desktop clearly
separates unsupported exact-edge text from the wider child action and preserves
namespace-qualified service identities. Narrow currently shows a horizontally
scrolled table's action side, with parent/child identities clipped off the left.
That is a task-readability limitation: the action's population cannot be checked
alongside its identity in this viewport. It is not by itself proof of the reported
document-overflow cause; native element bounds must identify that independently.

Prefer a bounded owner containment/wrapping correction where the measured overflow
origin lies. If a mobile table adaptation is needed, reuse the existing shared
stacked table and per-cell labels instead of a graph redesign or new table system.
Final narrow acceptance remains pending corrected actual frames. No exact typed
edge support is implied by making the wider action readable.

## Regenerated distributed and library frame acceptance

Reviewed the regenerated `352-distributed-1280.png`,
`352-distributed-390.png` and `352-service-activity-library.png` after the shared
stacked-table/receipt fix. Narrow now places reported parent, qualified child,
record count, unsupported exact-edge explanation and wider action together in a
readable card. Empty, missing and literal namespaces remain distinguishable; the
previous clipped-identity layout is resolved. Desktop retains aligned columns.
Actual appearance passes for both graph views.

The service library frame explicitly names environment policy and warns that
saved filters may target another service. The frontend service context and saved
worker predicate are both visible, so the result is not misrepresented as a service
match. Open Activity view remains distinct from Start telemetry. This actual frame
passes. Native document bounds check overflow independently; the source receipt is
below these cropped frames, so its full wrapping is not inferred here. Final
unreported-policy Activity narrow acceptance remains pending.

## Final actual-frame acceptance

Reviewed `352-unreported-activity-390.png`. The explicit Unreported policy remains
visible alongside the saved worker view name and its own service/namespace/release
predicate. The actual worker log is readable; the paused stream and Resume-with-
recent-logs meaning are clear. No selected frontend predicate was injected and
no fixed discovery window is portrayed as the rolling Activity window. Narrow
controls and row content remain within the viewport.

All four current frames pass visual acceptance. Native four-policy source budgets,
explicit Start, current-owner checks and no-overflow assertions establish execution
facts separately; this review does not infer complete telemetry or exact typed-edge
support from the views. 352 retains the local exact-edge limitation for 353.
