# Comparison template design review — 300

2026-11-01. Scope: reuse the existing captured-count comparison composition with
an explicitly chosen destination baseline. This begins as a source/design review;
actual browser and screenshot outcomes are recorded separately below and in
VALIDATION.md. Proposed acceptance is not a claim that the feature is validated.

## Product outcome

A team can preserve its useful “what changed?” presentation and apply it in another
service or grouping without importing the specimen's evidence. The destination
keeps its current query and explicitly supplies the baseline capture it wants to
compare. The existing baseline picker, count comparison, preparation owner and
saved-view catalog remain the product surfaces. Do not introduce a second capture
catalog, parameter editor or comparison card beside the existing one.

The source capture reference is a validation specimen, not a default, fallback,
latest capture or authorization to inspect that evidence in a different scope.
An explicit `baseline: destination` template binding makes this requirement clear;
every expansion must provide its chosen destination capture. A selected identifier
is configuration, not proof that the capture is compatible.

## Authoring and discovery

Keep the existing Reusable template and Data profile controls under Preserve.
For a comparison composition, a short explanation such as “Destination baseline
required when reused” is sufficient. A second optional input or a copied source ID
would make the default unsafe and confuse stored evidence with reusable display.

In another scope, Inspect template identifies the destination grouping, selected
destination capture and preserved comparison presentation. If no destination
reference is configured, Apply remains unavailable. Offer one recovery action,
“Choose destination baseline,” which reveals and focuses the existing Compare with
captured baseline disclosure/picker. It changes focus/disclosure only; it must not
load captures, change the query, choose the source reference or prepare evidence.
For a noncomposed destination, recovery must accurately lead through the existing
composition action rather than promise a picker that is not mounted.

Keep reference display compact: a short identifier may have the full value in its
tooltip. Metadata already loaded through the picker may identify the chosen
capture, but missing metadata must not trigger automatic detail fetching.

## Verification language and effects

A syntactically admitted destination reference remains **unverified**. The preview
should say “Unverified · Prepare checks this baseline against the current scope,”
or equivalent concise wording. Do not use an unqualified “Compatible template”
message that implies the stored capture's grouping, filters or environment have
already been checked. Distinguish template-shape applicability from baseline
compatibility explicitly if both statuses are shown.

Apply changes only presentation and the explicitly bound reference. It performs
no telemetry read or stored capture-detail fetch. The existing Prepare action
loads stored evidence, verifies its root against the current destination and then
uses the existing comparison path. Wrong service/grouping/filter/environment,
unsupported count profile, missing capture and read failures remain unavailable
with actionable reasons; none become zeros or apparently successful comparison.
Run keeps its established documented effects and independent read clocks.

Choosing a different reference retires the previous preview and prepared pair.
Late responses, retained controls and reattached DOM must not apply the old
reference. Pending record selection, edits and reads continue to block replacement.
The specimen capture may be explicitly chosen by the user only through the same
destination-input path and verification as any other capture; equality never
implies an automatic choice.

## Acceptance journey

1. In service A, use the native count comparison/picker and save a reusable view.
   The template retains order, visibility, title, Summary/Groups and group order.
2. In service B with a different grouping, inspect it without a chosen baseline.
   Confirm disabled Apply, explicit prerequisite and keyboard-visible recovery.
3. Choose a B capture through the existing picker. Preview shows the exact
   destination reference and unverified state. Source A's capture is not loaded.
4. Apply locally with the target query unchanged. Prepare explicitly; only this
   step verifies the chosen stored capture and renders the count comparison.
5. Change baseline while a preview/preparation response is pending; the old
   response cannot restore A or the prior B reference. Save/reopen preserves the
   explicit B reference and stages the existing workflow.
6. Exercise a valid-looking wrong-root ID and an unavailable ID. Both must fail
   on explicit Prepare without showing stale prepared values. Missing or malformed
   parameters refuse before expansion/application, including hidden comparisons.

Keep current timeline compatibility and Table/List action tests as regressions.
A count timeline may coexist only within existing profile rules. Status, numeric,
native, reported-edge and general comparison parameters remain outside this cut.
Unequal window durations remain raw-count comparisons with visible durations;
no automatic rate conversion is implied.

## Review boundaries

No new component renderer, query planner, expression language, evidence store,
multiplayer state or baseline completeness witness is needed. The public display
components remain passive. Review actual native authoring, missing/selected
baseline preview and prepared destination frames after executing the journey.

## Implementation source review

The current app integration uses the existing baseline picker and its disclosure.
Missing-input recovery opens/focuses that owner, or accurately reveals composition
controls when the target has no comparison editor. It does not load capture detail
or choose a reference. Preview shows the selected destination ID with full-value
tooltip, distinguishes presentation applicability from unverified evidence, and
keeps explicit Prepare as the actual baseline verification step. Apply uses the
same local presentation owner.

The Reusable template control derives the destination-binding requirement from
its composition. Native addition/removal of comparison synchronizes this binding,
so removing a comparison cannot strand an ordinary template with an invalid
leftover baseline parameter. Existing draft changes retire preview state; reference
identity is checked again when applying a returned expansion. No source UX blocker
was identified. Review suggested removing duplicate unverified prose between the
reference line and status; this is a copy consolidation, not a new state or flow.
Actual keyboard, source-effect and screenshot checks remain separate validation.

## Proposed next outcome — inspect supplied span timing

After temporal and explicitly bound comparison reuse, another template profile is
not the most useful default next step. A bounded span-timing presentation can help
users see where observed time went, find an outsized or overlapping span, inspect
its attributes, and Keep the exact observation without losing their current scope.
It also brings an existing specialized visualization into the reusable display
contract. That user task is stronger than adding a third kind solely for variety.

The entry must remain honest about what it has: a supplied page of spans, not a
complete distributed trace. Generic service pages often contain one span per trace
across several minutes. The existing waterfall's page-global time extent would
turn millisecond spans into nearly invisible slivers in that case. A useful
multi-trace shape needs explicit trace grouping with a per-trace relative origin
and one shared duration scale, retaining UTC information, or a deliberately bounded
entry into a checked single-trace/multi-span scope. Use the largest supplied
trace-fragment extent for the common scale: removing idle gaps must not make a
5 ms and a 500 ms fragment look equal through independent autoscaling. Label the
origin as the earliest supplied span per trace; it is not proof of the true root,
request start or complete trace latency. Do not label an unreadable page-global projection a
trace investigation simply because it has bars.

A bounded acceptance story should start with a real multi-span population, switch
Table/List/span timing without a read, visually identify a long/overlapping span,
Inspect and Keep it, explicitly Expand trace to acquire cross-service context,
then save/reuse/reopen the presentation. Group multiple trace IDs visibly; reported
parent relationships must distinguish missing, ambiguous duplicate and cyclic
parents. Parent-first display reordering must preserve the original page-position
selection identity and duplicate occurrences. Invalid geometry must retain an
accessible exact-table alternative rather than silently dropping observations.

Keep source acquisition, grouping/refinement, selection and publication in the
existing application owners; keep public presentation passive. Explicit trace
expansion remains a distinct independent read. No new topology planner, trace
store, auto-expansion, causal inference, complete-trace claim or second Keep flow
is needed. This is a proposed next outcome, not delivered 300 functionality.

## Actual authoring frame

Reviewed the first actual 300 authoring frame from the native journey. Preserve
retains the existing Question/Evidence/View definition structure, one Reusable
template toggle and Data profile selector. The short destination-baseline
requirement is visible before Save, and the resulting saved-view link is shown.
No separate parameter editor or copied capture field was added. No visual UX
blocker was found in this frame. Destination preview/preparation review follows
below.

## Actual destination-frame review

Reviewed the actual destination-baseline preview and prepared comparison frames.
No visual UX blocker was found. Preview names the destination service/namespace,
`attribute:db.namespace` grouping, current fixed bounds and profile. The chosen
short baseline ID is explicitly Unverified; the explanation distinguishes local
Apply from Prepare's scope check. This avoids suggesting that successful template
expansion has already validated stored evidence.

The prepared frame names the Ledger baseline and shows the actual destination
population: ledger changes from 100 to 101 records (+1), while archive remains
2 to 2. Baseline/current bars use a disclosed common count scale, with exact values
and change beside them. Independent held-evidence context remains visible; detailed
scope/evidence provenance stays in its existing collapsed disclosure. No source
specimen or automatic rate/completeness claim is presented.

The native journey additionally exposed a Capture-to-return control-retirement
boundary, which source review alone did not establish. Acceptance requires fresh
current controls over the held aggregate, source-free return and refusal by
reattached old baseline/trace handlers. Pending selected/submitted visits remain
explicitly paused rather than reauthorizing their old publication handles. Actual
source/effect accounting and late-response/capture-return outcomes belong in
VALIDATION.md; screenshot review is not a screen-reader audit or a substitute for
those executable checks.
