# 429 architecture and boundary review

Status: final source, focused-test and actual measurement-contract review accepted;
the strict native/model verdict remains separate. This report does not infer
model success from source inspection or browser measurements.

## Keep the consolidation boundary small

427 and 428 complete a useful operator loop without a new artifact family:
immutable saved-view references recover full definitions, while general pair
captures reread and retain ordered evidence. The next improvement should reduce
work the operator repeats, not introduce another query, permission or preservation
framework.

Several shared pure boundaries already exist:

- `Preservation.selectedQuestion`, `comparisonQuestion` and `availability` select
  exact configured intent and distinguish supported evidence profiles from pending
  edits or reads.
- `RecordEnvironment.selection` and `questionFields` retain environment policy
  while preserving the representation used by immutable question identity.
- `nativeControlReady` and `completionControlReady` distinguish native invocation
  readiness from an owned asynchronous completion whose opener is deliberately
  disabled.
- Evidence-scope and comparison-context presenters format supplied facts without
  acquiring telemetry or publishing artifacts.

Reuse these functions before extracting local code. A small pure capture
destination policy can choose which target-specific fields and explanatory text
are relevant. It must not become the authority to publish or navigate.

## Identities and lifetimes that must stay separate

An authored clause path identifies the condition being edited. Semantic query
identity compares complete checked intent. A saved content ID identifies an
immutable definition or receipt. A source profile decides what evaluation means.
Held custody identifies already obtained observations. A private invocation lease
decides whether a completion may affect the current screen. These are different
questions; no universal key should replace them.

Population and fixed-editor recall snapshots contain different generations,
drafts and parent transitions. Their controllers should remain separate. The
capture dialog's authorized publication/attachment may finish after departure;
the navigation lease must still refuse late adoption. Hiding an irrelevant
destination field must also remove any stale semantic selection from confirmation,
not merely make a selected target invisible.

No production Go change or new API is needed for the proposed presentation
consolidation. The current project already targets Go 1.27 and runs on Go 1.27.3;
there is no benefit in adding backend churn to satisfy this UI milestone.

## First implementation review

The new `captureDestination` policy is appropriately small and pure: it projects
current target/question choices into visible controls, a selected key and relevant
destination copy. The capture owner still owns question snapshots, publication,
attachments and its separate navigation lease. Hiding the unused question link or
revision explanation does not introduce another permission model.

One required correction was identified in the new Quick filters disclosure.
Existing shortcut inputs and suggested-breakdown callbacks checked only the
root editor owner. A retained suggestion callback could therefore invoke
`breakdown` and execute a new read while the newly added disclosure was closed.
Folding introduces a new withdrawn state: these callbacks must also check native
control readiness. Closing the disclosure must retain already chosen predicates;
withdrawal of an action is not removal of the question's filters. This finding was
sent to implementation before final acceptance.

The final shortcut handlers and suggested-breakdown controls now reuse
`nativeControlReady`. Independent actual-handler tests cover closed, disabled,
hidden, detached and departed callbacks without an edit or read. The root's
semantic owner checks remain separate; no generalized permission key was added.

Compact counter summaries and timeline scales are supplied-value presentation.
They retain exact raw strings, native units, domain/mark geometry and existing
unknown-state guards. Tiny nonzero rates are not rendered as zero. A review
finding corrected an accuracy claim: a fallback floating `perSecond` value is
labelled **Recorded rate**, while a supplied exact rational retains **Exact rate**.
The storage/evaluation model is unchanged and exact detail remains available.

Independent focused execution passed **21 tests, zero failures/skips, 0.902s**:
429 destination projection/rendering and withdrawn shortcuts, compact rates and
timeline domain retention, existing capture route ownership, actual document
attachment callbacks, full pair admission and detached exact saved-view recall.
The former callback blocker is resolved. No remaining source architecture blocker
is identified; actual before/after measurements and combined browser/model
acceptance remain the parent-owned final gate.

## Acceptance to review

Compare the same applied/read state before and after at 320, 390 and 1440 pixels.
Keep edit state separate. Record position and visible-control counts are layout
measurements, not execution-speed claims. Measure local request and repeated
pause/return work directly before claiming reduced acquisition cost.

Retain the existing source/store and row-copy budgets. Re-run pending-save,
retired callback, pause reentrancy, scope ABA, exact definition reopen and actual
opener return checks where changed presentation affects those paths. Local
disclosure, checking, staging and Return must not acquire telemetry.

## Actual measurement-contract acceptance

Independently inspected the matched `before-d5e1cbe007` and final
`after-d5e1cbe007-d10518` browser reports. Both are passed and cleaned up with no
browser errors. The fixed fixture objects are identical. Activity deliberately
uses each runtime's actual rolling clock and a separate phase bank; its intervals
are not claimed identical to the fixed before/after fixture.

The report matches the documented claims:

- Applied first-record positions remain 368/459/481 px at 1440/390/320 widths,
  with the same three records and 15 visible actions. Applied population layout
  and action counts are also unchanged. There is no claimed read-state speedup.
- Editing moves the first row from 815/1242/1322 to 688/960/1040 px and reduces
  visible actions from 52 to 48. Capture-only reduces visible actions from nine
  to eight and modal-relative Save offsets from 715/995/1130 to 522/745/843 px.
  Desktop Recent action distance falls from 913 to 10 px; mobile rows use their
  existing stacked layout, so a horizontal gap comparison is not meaningful.
- Main source SQL request count stays **22 before and 22 after**. The final
  exploratory actor adds six owned source requests and reaches its observed
  completion in 452 snapshots. These are acquisition-budget observations, not
  production latency or throughput improvements.
- Held Activity Return preserves row keys, Find, DOM identity and opener focus;
  the report records zero row mutations. Return and repeated local Preserve
  operations add no source, store or metadata requests. First opening of the
  capture form still performs its existing store-only `GET /api/documents`;
  it does not query telemetry. The distinction is recorded rather than hidden.
- The actual closed Quick-filter callback is refused. The actor performs one
  explicit comparison Run, cancels the form, returns focus and confirms that a
  retained closed confirmation callback adds no requests.

This closes the architecture acceptance of the measured source/UI contract.
The strict native/model gate and the separately rerun CEL preservation journey
remain the parent's final execution checks; neither is inferred from these values.

## Follow-up: construction-time events

The existing real CEL preservation regression then exposed a concrete Trail
construction bug despite its semantic steps passing: `Exploration.mount` can emit
question/saved-scope events synchronously before `create` receives and assigns its
API. Unconditional `entry.api.currentSavedView` access raised page errors.

The bounded fix treats those early API reads as not yet available. A checked
question event can supply the existing typed-question fallback; a saved-scope
event alone cannot acknowledge an exact saved definition. Once the owned API is
available, its current saved reference takes precedence, and a later ordinary
publish does not flatten a composition to its selected query. Read-only focus and
recall-state access similarly tolerate the brief construction boundary. No new
lifecycle, event registry or saved-view carrier is introduced.

The production Trail mount is tested with a dependency that emits both events
inside its synchronous mount call before returning the API. It covers fallback,
refusal of event-only acknowledgment, later exact reference, non-flattening publish,
defensive copies and edited/unacknowledged fallback. Independent execution of
Trail, detached recall, route ownership and same-ID recall passes **23 tests,
zero failures/skips, 0.826s**. Source acceptance is closed; clean real-browser
regression reruns remain the parent's final gate.

## Questions for the next whole-product checkpoint

After this finite cycle is accepted, review the complete operator loop rather
than automatically add another narrow queue:

1. Can a new operator reach useful unfamiliar OTLP data, identify a population,
   compare it and return without reconstructing scope or learning artifact names?
2. Which controls or narratives remain duplicated across live, fixed and saved
   views, and which differences are necessary because acquisition or custody
   differs? Use measured repeated work and user-visible friction as evidence.
3. Are source refusal, partial signal availability, unknown completeness and
   nontransactional comparisons consistently visible without overwhelming the
   primary investigation?
4. Which operational gaps prevent a real team from adopting Goo, versus features
   that already exist but need discoverability or coherence? Reassess source
   authorization/tenancy and monitoring delivery separately from UI polish.
5. Do packages, checked predicates, saved views, capture cards and user-defined
   kinds still form one comprehensible product? Avoid adding a carrier until an
   actual reuse gap requires it.

These are checkpoint questions, not a new implementation roadmap. Broader CEL
vocabulary, executable kinds, classifier training, warehouse lowering and native
metric expansion retain their existing design/operational gates.
