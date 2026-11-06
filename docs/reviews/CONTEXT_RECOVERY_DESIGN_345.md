# Contextual investigation recovery — design review 345

Status: implemented source and all five actual desktop/narrow frame reviews
accepted. Execution results are recorded separately in the journey/validation.

After finding a record, users follow its trace, related logs or surrounding time
context. A source refusal at that point should leave them oriented and able to
retry the affected read. Reuse 344's concise reason, safe Technical details and
current application-owned recovery; add no second search engine or error console.

## Source findings and interaction footguns

`ui/trace-search.mjs` already has Run trace search, Cancel and Close. Its raw
`Error(data.error)` loses the closed refusal information. Keep the original
conditions and fixed parent scope available when showing the shared refusal.
Run trace search remains the natural explicit recovery action.

`ui/record-results.mjs` owns distinct trace, linked-log and surrounding controllers.
Its catches currently concatenate `e.message`. Linked logs already have an explicit
read gesture; surrounding logs have their Run action and radius control. A trace
span failure disables More, so it must not leave the user with no usable recovery
except closing and reopening. Reuse a meaningful existing action if available;
otherwise a small owner-scoped Retry trace read action is justified. Do not call
a disabled paging control Retry or imply a linked-log retry also rereads spans.

A successful `traceRows` currently proceeds to `logs`. If that continues, a trace
retry is a span read followed by a separate log read, not a trace-only request.
Its label/context and observed source budget must remain truthful. A failed linked
log read must retain the known waterfall and independent span receipt.

## Scope and recovery contract

- Trace search preserves all staged span conditions, exact typed environment
  policy and original fixed UTC bounds. Its match witnesses remain reported
  telemetry, not complete traces or causal claims.
- Exact trace continuation preserves trace/span IDs and originating policy/bounds.
  Supporting linked logs remain a separate read with their own receipt and selected
  trace-versus-span scope. Failed logs are unavailable, not “no related logs”.
- Surrounding logs retain the anchor, service/namespace identity, chosen radius,
  clipped original window and established contextual scope. Retry cannot silently
  restore original message filters or claim temporal proximity is causality.
- Recovery uses the failed request's intended page/offset, or explicitly names a
  reset to the first page. Never silently retry a different page because the failed
  response cleared the last successful page object.
- Known independent observations remain available with their original clocks.
  Failed acquisition does not turn them into a complete joined snapshot; no-data,
  unknown and source failure retain their distinct meanings.
- Primary copy is short and contains no transport URL. Technical details is
  keyboard reachable, bounded safe text and source-free. Existing recovery controls
  sit in the owning context with unambiguous accessible names and pending state.

## Ownership and navigation

Cancel/Close, changing the selected trace/span/log scope/radius, Back and departing
retire old requests and recovery handlers. Late refusal or success cannot overwrite
a newer context, steal focus or reactivate a departed retry. An active canonical
source cancellation may report cancellation; local owner retirement stays quiet.
No auto-retry, automatic source reads from disclosure, budget increase, implicit
capture or broader time window is introduced.

On explicit Close return to the current visible opener; retain 342's working Back
controls and 340's unobstructed focus behavior. Do not add explanatory prose about
internal controllers to the ordinary investigation surface.

## Native acceptance

Use genuine OTel entry through ordinary exploration into at least trace-wide
search/continuation and a log context. Trigger controlled source refusal, inspect
shared details by keyboard with zero IO, then explicitly recover and independently
verify exact question/policy/IDs/bounds and advertised read count. Check a failed
linked-log read with a still-known waterfall, as well as surrounding anchor/radius
and failed-page ownership where covered. Close/changed-context/late-response
negative controls prove old handlers cannot publish or read.

Actual desktop/narrow frames must show concise reason, appropriate recovery,
retained known context and readable wrapped detail without page overflow. Frames
provide appearance only; native ledgers establish read boundaries and ownership.

## Frozen implementation source acceptance

The frozen implementation reuses the shared refusal parser/display for trace
search and record context adapters, preserving the typed envelope through saved
and live record owners. Trace refusal now appears in the primary inspector status,
outside collapsed evidence. The new failure-only Retry trace context action has
compact “Reads spans and linked logs” help, accurately describing the existing
success path's two independent reads. It is bound to the failed context and offset.

Linked-log and surrounding recovery retain failed offsets; changing scope or
radius intentionally resets to page zero. Pending read actions are disabled,
origin scope is deeply copied and current generation/context handlers own reads.
This addresses the identified retry ambiguity without another mode or automatic
behavior. Source acceptance passes. Actual frames still need to establish primary
reason/retry readability and retained known waterfall versus unavailable logs on
desktop and narrow layouts; execution checks separately establish IO and custody.

## Actual frame acceptance

Reviewed all five genuine 345 frames: trace-search refusal, trace-context refusal,
linked-log refusal at 1280 and 390 widths, and linked-log recovery. Trace search
keeps the conditions and original scoped context visible above a concise reason;
keyboard-focused Technical details shows the safe explanation without a URL.
The trace-context failure places its reason before collapsed evidence and exposes
Retry trace context with the short spans-and-linked-logs boundary help. No empty
waterfall or zero-span result is fabricated.

Both partial linked-log frames retain the known two-span 50 ms waterfall, including
the worker error, while logs explicitly remain unavailable. Find linked logs is
the nearby current recovery gesture. Narrow controls wrap within the inspector,
side content remains readable and no horizontal overflow or clipping is visible.
The recovery frame keeps the same waterfall and shows one captured exact-trace log
with its separate-evidence explanation. Actual visual acceptance passes.

These are task-positioned frames, not claims about fresh-entry density or all
specialized contexts. Native source counts, unchanged policy/IDs/windows, pending
button behavior and old-owner negatives independently establish the operational
contract; screenshots alone cannot establish those properties.
