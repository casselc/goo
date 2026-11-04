# Continue from a kept observation — final review 327

Status: final source and all three actual frames accepted. Native and actor passed
as reported by verification. No product/browser edits were made for this review.

## Delivered user task

A responder reopens kept evidence, inspects an original log/span and continues
into its distributed trace. **Open trace as view** lives in the individual kept-
record inspector, beside its provenance. It reuses the staged records-only
Waterfall, existing Run and retained navigation; no new catalog, wizard, evidence
format or capture owner was introduced.

The shared inspector now uses one trace/log action row and the concise note
**New questions · kept evidence unchanged. Run explicitly.** A closed
**Question scopes** disclosure shows the original predicates not carried forward.

## Scope and source contract

Full kept-receipt validation precedes staging. The question uses the validated
member's original read window, checked typed environment policy and valid nonzero
TraceId, with span signal, Waterfall presentation and offset zero. Result-page or
capture keeps use their validated question's fixed window. Custody time, event
timestamp and current clock never substitute for that window.

Original record predicates, service and span restrictions are deliberately omitted
so ancestors and other services can appear. Retain environment/window/TraceId;
do not infer policy from a raw resource key. Missing/invalid TraceId or unsupported
scope produces an unavailable reason before effects, with no silent fallback.

Staging performs no telemetry read or custody/saved-evidence-catalog read/write.
Creating the existing Exploration visit does perform exactly these configuration
metadata requests, counted separately:

- `GET /api/recipes`
- `GET /api/retained-evidence/status`

No recipe POST or broader metadata/source exception is admitted. Staging is
source-free, not all-IO-free. Existing explicit **Run** acquires current source
data through the old fixed question; absent or expired source data does not erase
the kept observation or prove that its trace never existed.

## Provenance and retained return

Frozen Kept at, Observed at, read basis, occurrence identity and stored bytes stay
unchanged. The new view owns its separate read clock, coverage and readiness.
It is not a frozen complete trace reconstructed from selected kept records.

Existing ResourceReturn provides **Back to kept observation**, restoring the same
retained inspector/action focus and scroll instead of reopening the store. Each
Inspect renews a member-visit lease; current generation, actual visibility and
source ownership guard actions. Close returns the existing focus.close result,
preserving the inspector dismissal/focus contract. Detached/reattached old controls cannot stage a
new visit or adopt late results.

Generic Trail navigation is suppressed only while that retained-return lease
exists. Releasing it invalidates the old return and restores ordinary Trail
controls; nested population Back keeps its separate parent meaning. Entering the
fixed trace clears unrelated Activity origin. Catalog reopen does not fabricate
a retained paused Activity ancestry.

Completed-read Back makes zero API requests. Pending-read Back may abort the
browser request and send the existing
`POST /api/queries/<original X-Goo-Query-ID>/cancel`, body `{}`, response 202.
That is cancellation control, not a source/custody/catalog reread. A held response
whose underlying read already completed does not prove physical server
cancellation; late adoption is still refused.

## Actual frames and verification

The three maintained 327 screenshot paths were refreshed by the 328 combined
journey, fixture `45c59ed1`, using a trace-enabled kept log with present-empty
namespace. They show the current shared two-action inspector, its original-window
staged span question and two-service gateway/worker Waterfall. The 50 ms/30 ms
bars share a scale; the new read clock is later than the frozen read window,
and unknown coverage remains visible.

This refresh validates current visual documentation and the kept-log trace path.
It does not claim that the refreshed run re-executed kept-span preservation.
Historical 327 two-signal semantic validation is retained separately below.
The crops do not show Back or expanded omissions; exact predicate/return/custody
claims depend on native assertions, not appearance. No visual blocker found.

Historical 327 verification reports native PASS 64.512s (69.995s runner) and actor PASS 408 states/30.149s. Both
kept signals exercised original-window typed scope, cross-service trace reads and
retained return. Each stage allowed only the two metadata GETs; completed Back
was API-free, and pending Back admitted the exact cancel control without late
adoption. Original inspector identity, focus, scroll and stored bytes survived.
No further UI changes or additional product requirements are requested.
