# Contextual preservation review — 288

2026-10-31. Product and architecture review of the existing fixed-question,
population, investigation and capture owners. This initial proposal is source-based;
implementation and live validation are pending. Executed checks belong in
[VALIDATION](../../VALIDATION.md).

## Decision

Use one native **Preserve** disclosure in the current fixed-question or population
context. Move existing actions into it rather than adding another menu above a
second set of save buttons. Keep it available in fixed Activity analysis without
requiring Edit analysis, and hide the outer fixed-question disclosure when a
population/focused view owns the screen. Live Activity's rolling views and Keep
workflow need not be migrated in this cut.

Keep three objects and effects clear:

| Object | Preserve operation | Reopen/read effect |
| --- | --- | --- |
| Question | Share its link or add an authored question card | Stages fixed intent; Run reads |
| View definition | Save queries and display configuration; optionally add the exact saved view to an investigation | Stages configuration; Run reads |
| Evidence | Open the existing capture confirmation dialog | Confirm explicitly rereads and stores independent evidence; stored reopen does not reread |

Short explanations belong beside the selected object's controls inside the
expanded disclosure. Capture must say it performs a new read and can differ from
the displayed data. A view cannot become evidence by saving it. Question links
contain visible intent, not credentials or execution authority.

## Existing owners remain authoritative

- Fixed Explore reuses Share question, Add question and Capture evidence handlers
  and IDs. Move telemetry-package creation under Advanced; it is a separate family,
  not another ambiguous primary save choice.
- Population Preserve owns its existing view title/save and exact saved-view
  finding attachment. Editors/layout remain View settings. An unsaved changed view
  cannot silently attach the previous saved ID; no implicit save is added.
- Single-query populations can preserve that root question with offset zero.
  Multi-query populations must require an explicit signal choice for question or
  evidence operations. Tabs only choose visibility and must not silently choose
  which query is captured. Save view still preserves the whole configuration.
- Existing capture dialog, destination picker, immutable stores, document card
  kinds and recovery journals remain. Do not add a universal saved-object envelope,
  merged capture, second destination picker or automatic investigation publication.

The old event-only pin path can report refusal solely in the hidden investigation
editor. Prefer a thin authored-record-question controller method, mirroring the
existing stageView guards, that returns or throws to the visible source control.
Do not overload stageQuestion(receipt), which deliberately creates a reusable
question associated with captured evidence. Existing applied draft cards can be
appended to; pending note/link composers, recovery, loading, save operations and
changed destination tokens must be resolved explicitly.

## Acceptance and quality checks

1. One visible Preserve entry per owning context; opening/closing it performs no
   telemetry read, save or capture. Fixed analysis does not require Edit analysis.
2. Question link/card retains exact typed policy, predicates and fixed window.
   Sharing or adding does not capture data or publish the investigation.
3. View save/attachment retains exact current saved ID and declaration. Dirty or
   stale saves cannot attach a previous definition. Opening remains staged.
4. Capture entry opens confirmation only. Confirm performs the independent read;
   cancel, metadata discovery and held evidence reopen add no source reads.
5. Two-query question/capture controls remain unavailable until an explicit signal
   choice. The saved view remains a multi-query definition.
6. Pending note/recovery and stale destination refusal is visible at the source
   without losing the editor's work. Existing card limits remain intact.
7. Detached, hidden or retired source controls cannot pin a question, navigate a
   share link or open a capture for a newer context. Application owners retain
   active/connected/query checks; the disclosure itself receives no source power.
8. Actual screenshots must show compact control grouping and object-specific read
   effects. Test the canonical behavior, not only button relocation or prose.

No production code or heavy validation was performed for this initial review.

## Frozen implementation source review

The implementation uses a small app-only presentation helper, existing capture
confirmation and a thin `stageRecordQuestion` controller. It introduces no durable
format or capture fusion. Source inspection confirms explicit multi-query signal
selection, offset-zero question intent, visible destination refusal, separate View
settings and inactive/disconnected checks before question/capture work.

Review found and corrected two integration defects: the composition source editor
still searched for its old Preserve ancestor after moving to View settings, and a
render-generation check permanently retired the held parent's configuration menu
after returning from a child. Configuration callbacks now revalidate the current
visible owner and current queries; detached controls remain inert. A paused live
configuration alone does not block preserving its fixed query. Capture explicitly
excludes the live buffer and does not combine independent comparison reads.

A final guard review requested equivalent invocation-time ownership checks for
view publication, retaining the existing stale asynchronous response checks.
Browser validation should include applying unchanged valid composition source,
returning from a child, explicit signal selection, pending-note refusal and the
separate capture confirmation boundary. This source review does not claim those
browser checks have run.

## Actual screen review

Reviewed the verifier's [explicit signal](../journeys/screenshots/288-explicit-signal.png),
[independent capture](../journeys/screenshots/288-independent-capture.png) and
[preserved pieces](../journeys/screenshots/288-preserved-pieces.png) screenshots.
The single expanded disclosure distinguishes all three meanings. Its unchosen
signal visibly disables question/capture without blocking whole-view publication.
The capture modal explains the independent reread and shows Capture only; the
investigation retains separate authored question, frozen evidence and saved-view
cards. These screens support the intended object/read boundaries.

One follow-up, outside this bounded cut: the existing saved-view Finding textarea
and attachment form dominate expanded Preserve's height. A contextual attachment
disclosure could reduce that space without hiding the primary Save operation.
No additional controls or durable object types are warranted by this review.

Final source check confirms view publication now refuses inactive, detached,
hidden, child-owned and already-publishing invocation before its metadata POST,
and reports pending draft/selection refusal locally. Existing revision/generation
checks still suppress stale asynchronous acknowledgments. No source-review
blockers remain; executed browser/model/test outcomes are recorded by their owner.
