# Contextual action consistency review — 420

Status: bounded independent source review accepted. Actual browser/model execution and visual
acceptance are separate gates in validation. This reviewer did not implement
production changes or run heavy checks.

## Two concrete owners

Service widget actions now share native invocation checks for ordinary and HTTP
contexts. Open, Add, Capture and Reuse no longer bypass closed disclosure,
disabled fieldset, inert, hidden or retired-owner checks. Customize separates
invocation from asynchronous completion: its deliberately disabled opener may
complete only while the widget lease, DOM containment, source question and native
completion eligibility remain valid. Retiring or withdrawing the owner prevents
a late configuration handoff. Existing source budgets and query semantics remain
with the application owner.

Record inspection and context actions use the same native invocation checks.
Hidden Inspect callbacks cannot switch retained selection. Row/inspector trace
and surrounding-log actions, paging, retries, selection and explicit close retain
owner/generation checks in addition to native readiness. Existing asynchronous
source-response leases are unchanged; making an opener busy is not itself a
reason to discard a current response.

## Retained return focus

The shared focus helper keeps disclosure revelation opt-in. Explicit retained
service/context returns may reopen only necessary native disclosure ancestry and
focus the actual clicked action. First-summary return origins remain visible
without opening their own menu. Default behavior and the record row More menu's
intentional persistent-summary origin remain intact.

The helper rejects detached, hidden, inert, disabled and CSS-withdrawn origins;
it does not remove withdrawal attributes or revive stale owners. Owner checks
must establish the retained return before calling it. An unavailable origin uses
an eligible fallback, and an unavailable fallback receives no focus. Revelation
and focus have no source acquisition or publication authority.

## Resolved review finding

A completion must retain the same HTTP departure gate as invocation. The proposed
Customize completion predicate initially omitted `leaveReady()` while checking
its request/widget lease and native completion readiness. If sibling discovery
withdraws the HTTP departure gate while expansion is pending, that completion
must not stage a destination. The final predicate now rechecks `leaveReady()` at completion. A dedicated HTTP
regression covers a valid disabled-self completion and departure withdrawal before
fulfillment. Source inspection accepts this fix: the gate is retained without
confusing request completion with permission to invoke the disabled opener again.

## Next cadence

After this finite consolidation, the next breadth candidate is a direct Activity
“Facet by…” action on an existing field or row attribute. Counted full-scope
faceting already exists behind the filter editor; this is an entry-point adoption
of that capability, not a new facet engine. Preserve the held applied scope,
explicit aggregate read and source-free field selection, and reuse existing
facet/attribute components. One bounded journey should establish discoverability
and read accounting before choosing another depth milestone. This recommendation
does not replace the whole-product roadmap or reopen deferred platform work.

## Actual visual acceptance

Independently inspected all six generated `contextual-actions-{service,inspector}`
frames at 1440, 390 and 320 pixels. No visual blocker remains in these exercised
states. Service return leaves the relevant Widget actions disclosure open, with
its concrete Compare/Reuse controls reachable and readable. The inspector return
shows a clear focus ring on its actual Expand trace action rather than a generic
row menu. At narrow widths, native action controls wrap or stack within their
owner, the inspector retains its bounded attribute viewport, and the page does
not introduce a second off-screen action surface. The wide diagnosis table keeps
its existing horizontal scroll area; clipping that table viewport is not treated
as loss of the retained full data or as a new responsive-table implementation.

Screenshots establish visible geometry and the inspector focus ring. They do not
independently prove `document.activeElement`, zero rereads, keyboard execution or
stale-callback refusal; actual journey assertions and generated/native checks in
validation establish those separate outcomes. This is a bounded Chromium visual
and source review, not a whole-product screen-reader or cross-engine audit.
