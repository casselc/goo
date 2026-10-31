# Responder flow architecture review — 293

2026-10-31. Source review of the Activity → compare → chosen group side → inspect → Preserve → reopen flow, with the roadmap and 289–292 reviews. This document records architecture analysis, not a newly executed browser, model or performance check. Executed validation belongs in [VALIDATION](../../VALIDATION.md).

## Decision

The connected flow has the necessary ownership boundaries. Continue with one bounded correction: make saved-view investigation attachment optional and bind each attachment command to the view publication it actually represents. No new query API, storage type, general action registry or plotting framework is justified by this review.

The disclosure recommendation agrees with the [design review](RESPONDER_FLOW_DESIGN_293.md). A metadata attachment is secondary to saving the reusable view; opening its form must neither read telemetry nor publish a document.

## Concrete correctness gap

Before this cut, `mountViewAttachment` closes over its textarea while reading mutable workspace `savedID` and `title` at invocation. Its refusal check considers current pending state and activity, but does not prove that the button still owns the current attachment form or publication.

A retained button from view A can therefore be invoked after view B is loaded or rendered. If B is otherwise clean, the old handler may attach B's ID with A's finding and clear the newer workspace finding draft. The visible current form can then disagree with the draft actually retained for the next redraw. This is a cross-owner publication bug, not a reason to change query authority.

## Ownership contract

Keep the existing application owner responsible for attachment. A current invocation must prove that its mount and controls are connected and visible, the parent panel is not hidden behind a child, the workspace is active, and the command's saved ID, title and saved/draft revisions still match. Check these conditions before consulting the document publisher or mutating a finding.

A successful Save can change publication identity without replacing the surrounding DOM. Renew the attachment button for that legitimate new identity and permanently retire the old button. Preserve the textarea and its draft, focus/caret where possible, and the disclosure's open state. A finding edit alone does not change the saved definition's identity. Refusal must not clear either the old or the current draft.

Do not use the global source generation as the attachment's sole lease. This operation attaches saved configuration metadata. Source actions have a different lifetime: entering a child retires existing source actions, and Back selectively renews group/focus controls. Restoring a retained parent must not revive older plot or record source handles. The attachment may become usable again on the visible parent when its publication and mount remain current.

The existing `Documents.stageView` path retains responsibility for destination and pending-document policy. This cut must not bypass its refusal behavior, introduce implicit Save, or turn a saved definition into captured evidence.

## Acceptance and review limits

- Retain an attachment button, then load/render another view or save a new publication: invoking the old button publishes nothing and clears no finding.
- Save while the current Preserve DOM remains mounted: a fresh attachment command can attach the exact saved reference and current finding.
- A collapsed disclosure, detached form or parent hidden behind a child cannot publish through a retained handle. Returning to the legitimate parent preserves its draft and permits its current command.
- Expanding, collapsing and renewing the action performs no telemetry read or document mutation. Pending or dirty state remains a visible refusal.
- Preserve the existing distinction between question intent, reusable view configuration and captured evidence, including explicit signal choice for a two-signal question.

The broader source/read contract remains stable. Chosen group-side navigation stages an exact independent question and requires Run; it does not promise stored rows or a shared snapshot. Back's selective action renewal is intentional. Repeated warning prose and row-action density are smaller follow-ups when those owners are next touched, not reasons to widen this milestone.

The selected implementation has passed this source review: a fresh exact-publication button replaces the retired action while the textarea and native disclosure remain stable; mount/visibility/child checks precede publication; post-publication clearing verifies current ownership and unchanged text. Existing document-destination and pending checks remain in force. Browser acceptance remains the executing agent’s responsibility and is recorded separately. This review does not claim a whole-product security audit or new live journey coverage.
