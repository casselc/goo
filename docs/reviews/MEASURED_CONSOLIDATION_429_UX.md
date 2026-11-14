# 429 scope and preservation consolidation: UX review

Status: source, focused checks and final measured runtime presentation accepted.
The parent owns the complete browser/Bombadil/native verdict separately.
This is a bounded refinement of the delivered
427/428 flow, not a new product or artifact family.

## Priority and rationale

1. **Show capture destination controls only when applicable.** With Capture only,
   the current dialog shows a disabled question selector, a separate-card note
   and a saved-revision explanation before the operand preview. Keep title,
   destination, both fixed scopes, a short reread explanation and Save/Cancel.
   A current draft may expose its actual question candidates; a saved revision
   needs a concise child-revision explanation. Hidden destinations must reset
   obsolete selections and retain native disabled/ownership guards.
2. **Keep Recent actions next to their reference.** Existing desktop rows put
   Stage/Pin/Forget nearly a full viewport away from short titles. Use a bounded
   content column and adjacent intrinsic action group. Preserve complete titles,
   exact IDs and UTC detail; narrow rows may wrap actions below their summary.
3. **Fold fixed quick-filter chrome into its existing editor.** Service, message,
   literal severity and span shortcuts duplicate the filter editor's entry space.
   Keep these accessible through a native disclosure, without changing staged
   predicates or implying that literal severity text is numeric OTel severity.
   Opening the disclosure must not read data. Applied and edit states need
   separate measurements rather than crediting hidden editing controls as a
   first-record improvement.

The 428 desktop capture frame makes the first priority concrete: roughly 150px
of unrelated destination controls separate the destination choice from the
operand preview. This estimate motivates measurement; it is not a measured
performance or final-layout claim.

## Interaction constraints

- Keep one obvious execution action in each context. Stage, disclosure, local
  search, pinning, preserved-data reopen and held Return do not execute telemetry.
- Hide irrelevant controls with native `hidden` state so they leave tab order.
  Retain a visible label and native state for every remaining input.
- Keep current destination semantics explicit: Capture only stores evidence;
  adding a capture reference to a draft still requires saving its revision.
- Preserve the 428 confirmation and attachment leases. Cancel/Escape return to a
  connected current opener; departed completions cannot navigate or steal focus.
  An authorized publication may finish after withdrawal without implying rollback.
- Keep the visible stage/applied/frozen distinction and exact scope available.
  Compactness must not turn unknown coverage into complete data or replace full
  source totals with the number of displayed rows.
- Preserve selected-query versus whole-comparison capture and the distinction
  between a saved definition, general pair capture and saved comparison receipt.

## Acceptance to perform

The journey should compare the same question, result, viewport, read state and
scroll origin before/after at 320, 390 and 1440px. Record first-row position,
visible action counts, destination modal geometry and request/work counts across
Activity, fixed investigation, exact recall, pair capture and retained Return.
Do not claim an application speedup from a smaller screenshot.

Inspect final desktop/narrow frames for wrapping, aligned action groups and
visible scope. Check keyboard reachability, disclosure operation, current
opener/destination focus, native hidden/disabled behavior and late callbacks.
Physical assistive-technology, cross-engine and touch ergonomics remain separate
from source/frame/browser-keyboard acceptance.

## First production pass

The initial implementation follows the bounded proposal: a pure destination
presentation projection controls native hidden/disabled question fields and a
conditional note; fixed editing uses a native Quick filters disclosure; Recent
uses an intrinsic summary/action grid rather than a viewport-wide action gap.
Current full predicate chips remain visible, so closing quick controls does not
conceal the meaning of active shortcut predicates.

The existing recent-mode reparenting keeps its service shortcut available while
its remaining controls retain their own Filters and measures disclosure. Returning
to fixed editing restores those controls under Quick filters. No universal owner,
request cache or evidence carrier is introduced by this presentation change.
Relevant older executable authoring steps must now explicitly open the native
Quick filters disclosure before filling those inputs; keeping them programmatically
visible just for old tests would defeat the interaction change.

## Source and focused checks

Independent native-control/owner execution passes **20 tests**, zero failures or
skips, across the new destination policy, capture route, preservation owner and
Recent picker. A separate independent run passes **four compact-rate tests**,
zero failures/skips. These are focused source/callback/component checks, not the
full browser campaign.

Compact numeric presentation preserves exact scalar/rational detail and units,
marks approximations, and never formats a nonzero rate as zero. Shared scalar
axes keep the original data, geometry and UTC bounds. Their existing zero-anchored
domain prevents two distinct nonzero endpoints collapsing to the same compact
label; the regression includes nearby positive values to check that contract.
No new gauge/counter interpretation or coverage claim is introduced.

The cross-reviewed next-cycle proposal is separately recorded in
[product checkpoint](PRODUCT_CHECKPOINT_AFTER_429_UX.md). It does not expand the
429 implementation scope.


## Measured runtime presentation

All nine final after-frames were inspected independently: fixed editing and
Recent at 1440/390/320, Capture only at 1440/320, and the native counter at 1440.
The two retained before-capture frames provide direct visual comparison.

- The desktop Capture-only dialog shrinks from 782 to 589px. Its title,
  destination, both ordered scopes, short independent-reread explanation and
  Save/Cancel remain readable. The 320px modal retains contained native scrolling
  and reachable actions without horizontal clipping.
- The measured desktop edit path moves its first data row from 815 to 688px.
  Quick filters stays reachable through native disclosure; the complete active
  selector remains visible outside it. This is an editing-state gain. Applied
  record position and action counts are unchanged and are not claimed faster.
- The short desktop Recent reference/action gap shrinks from 913 to 9.75px.
  Long fixed-query summaries wrap while actions remain associated with their
  own row. Narrow lists retain internal scrolling instead of page-width overflow.
- The native counter now shows a small nonzero compact rate rather than zero;
  exact values and UTC bounds remain available in selected-interval detail.
  The plot's supplied geometry and zero baseline are unchanged.

The actual journey reports keyboard disclosure/Save/Cancel, focus return,
hidden/stale controls and late completion checks passing. The parent owns the
complete Bombadil/native verdict separately; source/frame review does not infer
that verdict.

One inherited visual issue was returned before closing: fixed editing showed
both Show recent activity and Rerun as filled primary actions. The accepted
color-only correction scopes the former to existing secondary surface/ink tokens
in a fixed question. It retains font, dimensions, handlers and native hover/focus,
leaving Rerun as the primary fixed-question execution action. The selector excludes
explicit Recent mode, preserving its separate primary action. Refreshed 1440/390/320
frames confirm that Rerun is filled accent while Show recent activity uses the
secondary surface. The actual passed report independently records distinct computed
background colors (`rgb(138, 185, 255)` versus `rgb(21, 26, 35)`), cleanup complete
and no browser errors. This closes the finding without changing dimensions or
execution behavior. No source or presentation blocker remains.

Physical assistive-technology, cross-engine and touch-ergonomics testing are not
claimed. The separate product checkpoint remains a proposal for later work.

## Existing preservation-path follow-up

The final regenerated [320px preserved view](../journeys/screenshots/4365-preserved-view-320.png)
and [1440px restored source](../journeys/screenshots/4365-restored-source-1440.png)
were inspected independently. Native Preserve controls, saved definition link,
checked-source disclosure and the optional CEL editor remain readable without
horizontal clipping. The protected OR parent and separate numeric-severity
condition remain visible; the fixed Run action retains primary emphasis while
Show recent activity is secondary. No additional polish scope is required to
accept this regression path.

The parent reports the complete existing preservation regression passing in
52.7s (54.4s runner), with 21 steps, 392 actor snapshots, no browser errors and
owned cleanup. The focused Chromium style probe also verifies secondary fixed
versus primary Recent-mode tone, unchanged geometry and retained native hover/
keyboard focus. Final 429 native/campaign acceptance remains separately owned by
the parent; these checks close the UX review rather than replace that verdict.
