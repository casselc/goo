# Live-source recovery — 433b acceptance

[Contract](../ACTIVITY_SOURCE_RECOVERY.md) ·
[Storyboard](../journeys/tail-refusal.md) · [Validation](../../VALIDATION.md).

Independent architecture and UX reviews selected a tail-specific closed cause and
retry disposition, retaining the shared five-kind refusal vocabulary. Stable
source/configuration, schema, budget and protocol failures stop for an explicit
new read; proven transport/timeout and expected rollover reconnect separately.
Classification uses typed errors, statuses and exception headers, never source
error prose. The request owner cancels silently.

Source review fixed gaps before the required reset, known DNS name failures that
were being retried automatically, and counts that attributed held observations to
a new connection. The real handler's rollover branch is checked with Go's virtual
clock support rather than a minute-long sleep.

UX follow-up made the specific closed cause the primary label in the existing
shared refusal component. Failed keyboard Retry restores focus only from its own
disappearing control. Inspection retains the same diagnostic DOM and open native
Technical details. Desktop and 390/320 frames were independently accepted; no
additional toolbar or custom disclosure was introduced.

A final interaction review restored local loaded-match selection on genuinely
stopped terminal failures. The two admission gates retain proof, expiry, original
connection/window, current revision, busy state and atomic-selection checks. An
ordinary unavailable/live view still cannot masquerade as a paused held view.
This adds no source-read or store-write authority.

The real browser journey exercises logs/spans against owned missing and malformed
ClickHouse schemas, explicit repair/retry, retained inspection and local selection.
An owned proxy separately injects access denial and transport interruption. The
SSE observer forwards actual bytes unchanged and drains delivery before assertions;
initial observer errors/races were corrected before its evidence was admitted by
native Quint and OpenAPI validation.

Architecture, UX and standards reviewers report no remaining blocker for this
bounded change. Physical assistive-technology speech and cross-engine behavior
were not tested. Proxy access denial is not a newly provisioned denied database
user; expected rollover has a real-handler virtual-time test rather than a
60-second browser wait. Broader source management and permissions UI remain
separate product work.
