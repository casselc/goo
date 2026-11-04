# Continue from a kept observation — design proposal 327

Status: proposed bounded general-purpose task; no implementation or browser work.

## Missing task

A responder reopens kept evidence and wants to investigate its distributed trace.
`ui/kept-saved.mjs` currently reconstructs original records and provenance, then
offers Inspect kept record only. Preservation is useful, but the investigation
stops at that inspector. Reuse the existing staged records-only Waterfall rather
than adding another trace browser, capture type or catalog.

## Interaction

Offer **Open trace as view** inside the individual kept-record inspector, adjacent
to existing observation actions. This gives the action the member's exact window
and provenance context, avoids adding buttons to every catalog/table row and works
for kept logs carrying a TraceId as well as kept spans.

Stage a span question using the validated member's original read window, checked
typed environment policy and valid nonzero TraceId. For result-page/capture keeps,
use the validated question's fixed window already mapped to that member. Select
existing Waterfall presentation, offset zero and ordinary record limits. Do not
substitute the keep custody time, event timestamp or present clock for the stored
read window. Do not infer an environment from one raw resource key.

The trace question deliberately expands beyond the original record predicates,
service and span restrictions: retaining those could hide ancestors or other
services. Retain environment/window/TraceId only, with a concise **New trace
question · original observation window; kept evidence unchanged** statement and
existing scope details identifying original filters not carried. No automatic
source read or Keep occurs. Existing explicit **Run** acquires the trace.

Invalid/missing TraceId or unsupported checked scope shows one concise unavailable
reason in that inspector. Refuse before side effects; no manual fallback silently
changes the environment/window. A missing source trace after Run means no current
matching source rows, not that the stored observation ceased to exist.

## Evidence and navigation

Keep frozen Kept at, Observed at, stored read basis and original occurrence
identity intact. The new view's result has its own source-read clock, coverage and
readiness. It is a fresh read of an old fixed question, not a reconstruction from
kept records or a complete frozen trace. Explicit Run can legitimately return
changed, expired or unavailable source data. Existing presentation uncertainty
must communicate those states without a second evidence card/wall of prose.

A named **Back to kept observation** returns to the same retained inspector,
member, selection/table scroll and opener with zero telemetry reads. Reopening
from the catalog does not promise a still-retained paused Activity stream; its
return owner is the kept observation, not a fabricated Activity ancestry. Reuse
existing retained child navigation/focus ownership and one deepest Back action.

Action validity requires active generation, current stored member and actual
visibility. Detached/reattached hidden controls cannot stage a new question;
return/disposal invalidates pending child work. Late source completion cannot
replace the restored kept view. Kept bytes/digest and custody fields remain
unchanged through stage, Run, Back and failures.

## Bounded acceptance

Use both a kept log and a kept span with typed environment policy and cross-service
trace ancestry. Reopen through server catalog (not only browser recent links),
inspect one original member, stage zero-IO Waterfall, explicitly Run, then return
without rereading kept/source data. Verify original exact window/policy, omitted
original filters, real cross-service ancestor/sibling visibility and new read clock.
Keep storage bytes/digest unchanged. Exercise invalid TraceId, unavailable source,
stale hidden action and return during read with no fallback or unexpected mutation.

Maintain three actual frames: kept inspector with contextual action/provenance,
staged original-window trace question and freshly read Waterfall with distinct
source clock. Native assertions establish zero-IO staging/return and immutable
custody separately from visual proof. This proposal needs no new backend
preservation format, grammar family, query syntax, registry or compatibility shim.
