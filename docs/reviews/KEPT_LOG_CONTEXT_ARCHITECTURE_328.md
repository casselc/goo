# Temporal log context from a kept observation — architecture 328

Status: final source review accepted; generated continuation and trace regression checks pass.

A kept log without a usable TraceId can stage surrounding logs for explicit Run.
The plan is temporal context, not identity, causal explanation or an upgrade of
frozen evidence. The existing 327 staged-view/retained-return owners serve both
trace and log continuations; no endpoint, executor, receipt format or return owner
is added.

## Exact planning boundary

KeptSaved shares detached receipt/member extraction with traceView, then
surroundingView requires a kept log, a reported nonempty ServiceName and a present
plain ResourceAttributes map. Namespace absence, present empty and named value
remain distinct. An own namespace key with a nonstring value refuses; it cannot
fall through ServiceIdentity.namespace(null) into an invented missing scope.
No service, namespace or clock is inferred from another field.

The existing Discovery.surroundingQuestion supplies ±30-second timestamp clipping
inside the original window. Tail origin uses the chosen member's own window;
result-page and saved-capture origin use the original fixed question. The upper
bound retains its existing millisecond +1 convention to include the anchor event,
and the final half-open window stays within the original bounds. Preserve checked
environment fields and page zero. Replace all old predicates with exact service
and namespace through ServiceIdentity.serviceWhere, then existing predicate/view
validation. Original message, severity, trace and other filters are not carried.
The existing general surrounding-log helper's behavior is not globally changed.

## Shared handoff and retained return

The callback is generalized to onOpenView/openKeptView; both actions use the same
private kept member lease, destination Trail lease and named Back. A private
original-member question accompanies the new view only as origin metadata. The
return bar uses that original window, not the clipped child query's window, so
Origin scope remains truthful. The app still stages the child view separately and
requires explicit Run for telemetry.

The shared continuation button handler rejects disabled or unplanned actions as
well as stale, hidden and retired member visits. This prevents direct invocation
of an unavailable action from entering fixed mode before an undefined view fails.
Independent inline reasons describe unavailable trace/service/namespace context.
Compact actions and one Question scopes disclosure preserve the original frozen
inspector and its return focus. The 327 cancellation, retained restore, catalog
suppression and irreversible destination release are reused unchanged.

## Verification

The new generated lane reuses a small data-only kept-receipt fixture from 327;
327's assertions and trace coverage remain unchanged. The fixture constructs and
Go-validates digested tail/page/capture receipts before the browser helper sees
them. The log variant has no trace identity and varies the event position inside
its actual window. Assertions independently derive clipping, check exact typed
policy and service/namespace AST, reject same-named foreign namespaces, refuse
invalid service/maps/clocks/indices/nonlogs and verify detached plans. Resulting
views are admitted by Go query/view compilation.

Actual source-free stage, explicit Run, unchanged kept receipt, original origin
window, namespace isolation and retained Back/cancellation are browser/API checks,
not claims inferred from the pure generated lane. See [VALIDATION](../../VALIDATION.md)
for final execution and the linked maintained journey.


Final generated execution passes TestHegelKeptLogContext328 and the unchanged
TestHegelKeptTraceContinuation327 assertions together in 14.095 seconds: thirty
cases each, sixty total, zero skips. Each case plans both members of all three
Go-validated kept origins, yielding 360 Go-admitted views across the two tests.
The shared data-only fixture preserves the 327 traced variant and constructs the
328 trace-less variant with varied event positions before computing digests and
validating receipts. No new fixture framework or production validation is added.
