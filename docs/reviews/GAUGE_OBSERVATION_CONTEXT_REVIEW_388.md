# Gauge observation context review — milestone 388

Read-only architecture/usability review, November 9, 2026. Current status: source
blockers resolved; actual runtime/browser acceptance and required model replay
reported passed with zero skips. The initial design snapshot and chronological corrections
follow below. No heavy jobs or production edits by reviewer.

## Minimal contract

Use a distinct pure observation-context projection over a validated ordinary
per-stream service gauge question, one exact stream, one canonical reported
observation timestamp and logs/spans signal. Derive a documented fixed 60-second
radius; clip the half-open interval to the original exact native question bounds.
No current-time fallback, caller-selected unbounded radius or implicit expansion.
Return an ordinary typed record question plus explicit observation/time-association
metadata. Keep existing interval/one-nanosecond and exemplar planners strict.

The enclosing millisecond query window must not weaken integer-nanosecond filters.
Both clip endpoints and the exact half-open record membership require Go/JS parity.
The selected observation must satisfy parent lower-inclusive/upper-exclusive bounds;
use checked integer arithmetic for subtraction near zero and addition near limits.
A boundary-clipped or very short parent is still explicit context, not coverage.

Preserve the reported selected stream's service, instance, exact missing/empty/named
namespace and parent environment policy. Do not infer a callee, trace topology or
record filters from metric attributes, instrumentation/schema identity or unrelated
resource labels. Counter, stock, state, resource and consumer-group authority is
outside this first arm. No native metric evaluation or persistence change is needed.

## Owner and user interaction

Ordinary Measurements is the first surface. The existing native-context follow-up
owner should stage the typed record question without source reads, expose explicit
Run/refusal/retry and restore actual held parent DOM/point with no source/store
reread. A pure planner validates question/stream/timestamp shape, but cannot prove
the point was retained. The held application owner must match the exact stream
and an actual retained sample, along with held question, basis and read clock,
before accepting a continuation. Do not trust a caller-supplied timestamp alone.

Observed, stale and unknown/no-recorded-value samples all have useful reported
timestamps. Unknown magnitude must not remove time context, nor may it fabricate
numeric value or timestamp. The unavailable rail's selectable sample should use
the same action path. Existing duplicate/conflicting observations remain one
unknown evaluated occurrence; no arbitrary record association is introduced.

Actions should read “Logs around observation” and “Spans around observation”. Show
the actual clipped UTC interval and that this is time association, not causality,
exemplar linkage or continuous gauge coverage. Keep the selected observation and
its provenance available on return. No extra wizard, query planner, navigation
framework or global state owner is justified.

Saved/service adoption is a separate explicit admission decision, as the roadmap
allows. The existing saved owner only admits counter/histogram intervals; do not
silently widen it because the ordinary renderer gains gauge links. Retired,
hidden/inert or replaced-parent callbacks must remain unable to stage/read data.

## Acceptance risks

Independent integer-nanosecond tests should cover exact radius edges, parent
clipping, near-zero bounds, submillisecond endpoints and exclusive upper boundary.
Actual browser data should distinguish records within radius from unrelated rows
in the wider parent window and foreign instance/environment/namespace controls.
Stage must read nothing, Run must use the exact staged question, failures/retries
must retain the parent, and Back must preserve selection/evidence without reads.
Keep existing interval/exemplar paths and histogram/counter semantics unchanged.


## Frozen Go source review

The implemented `PlanNativeObservationContext` takes the complete retained
`measurements.Result`, not only a parent question. It replays
`measurements.ValidateResult`, validates the pinned exact stream and requires
exactly one matching retained series/sample at a canonical timestamp. This can
prove retained membership without I/O; the earlier question-only design warning
applies to projections lacking that retained result. Unknown/stale observations
remain eligible by reported timestamp rather than magnitude.

The fixed radius uses distance comparisons before unsigned addition/subtraction,
then clips to parent exact lower-inclusive/upper-exclusive bounds. The private
`nativeRecordContextQuery` helper is shared after each caller has admitted its
own interval/observation semantics. It preserves exact nanosecond predicates within
the enclosing millisecond query, service, instance, namespace and environment,
without copying metric point attributes into record predicates. The existing
interval/one-nanosecond planner retains its original guards.

No actionable Go source blocker was found. The parent reports focused Go checks
passed; this reviewer inspected source and tests without rerunning them. Final UI
source and runtime acceptance remain pending.


## Frozen UI source review

The new browser mapper consumes the successful ordinary gauge result and verifies
exact question bounds, selected stream and unique retained sample. It deliberately
does not implement a second evaluator or replay/hash protocol. The application
adapter additionally matches the result question to the submitted source question
and current generation. Saved/service callers have no source question and receive
no new gauge actions. The distinct navigation branch retains the existing staged
question/held-return owner and clearly describes clipped time association and
broader reported-trace expansion.

One action-retirement gap requires correction: `appendObservationContext` returns
for modified/nonordinary clicks before checking visibility, inert state and
`canAct`, and has no auxiliary-click handler. Unsafe retired/hidden anchors must
suppress default href navigation before allowing ordinary versus modified gesture
behavior, matching the existing strengthened exemplar guard. Current safe modified
links may retain normal browser behavior. Add a focused retired ctrl/meta/middle
activation regression; no navigation rewrite is necessary.

The parent reports 1,393 frontend checks passed with zero skips and 62 focused
checks passed. These checks were not rerun by this reviewer. Go/UI cross-language
vectors, source inspection and browser execution have different proof boundaries;
actual runtime and the action guard correction remain pending.


## Action guard follow-up

The final observation-anchor handler now checks current ownership, visibility,
inert state and submitted-question identity before gesture classification. Invalid
or drifted actions prevent default and stop propagation immediately. Both click
and auxiliary click use this handler, while safe current modified gestures retain
normal browser navigation. This closes the source-review blocker without changing
query or return ownership. The parent reports 16 focused checks passed; this
reviewer inspected the final handler but did not rerun tests. No remaining source
blocker was identified. Actual runtime/build acceptance remains pending.


## Retained observation staging capability

The actual journey exposed that ordinary navigation intentionally retires the
acquisition generation lease. Requiring that lease to derive another gauge
question after Back forced an unnecessary native reread. The final source adds
only `canStageObservation`: retained result object identity, original plan owner,
matching submitted/current/response question identity, no pending controller and
visible owner are all required. Edits, Apply and new runs clear or replace the
retained result/plan and revoke this capability.

This permits source-free local gauge projection/selection and observation-context
staging after held return. Existing generation-based `canAct` stays unchanged;
existing capture authority is also unchanged, and this review does not claim
capture actions were previously paused. Histogram/counter/exemplar source-action
policies do not acquire the new capability. The visibility/modified-click guard
continues to use the specific staging capability before allowing navigation.

No source blocker was identified in the final closure. The parent reports 18
focused checks passed, including an actual Measurements-owner lifecycle with
repeated staging and zero reads. This reviewer inspected the final source but did
not rerun those checks. Refreshed full UI and actual journey acceptance remain
pending at this checkpoint.


## Final actual-flow report

The parent reports the refreshed actual source journey passed in 58.9 seconds,
with 16 observed facts over 97 raw spans, 97 logs and 16 metric rows. The selected
nearby populations were 9/9/13 records, compared with one-nanosecond populations
1/1/2 and whole-question populations 25/26/25. This distinguishes the new clipped
context from both point equality and the wider parent scope.

The measured main flow used one native read and four record reads, including a
real HTTP 502 refusal and explicit retry, with zero store reads and zero writes.
The actor produced 660 snapshots across eight phases, one native and one log read,
and zero store reads/writes. The parent visually accepted four frames, including
final focus, selected unknown rail and clipped scope bounds. Wire validation
passed for 20 responses and six requests. The final frontend run is reported as
1,395 passing checks.

These are parent-reported runtime/image results; this reviewer inspected source
and updated documentation without rerunning the flow or viewing those final
frames. Source blockers remain resolved. Required model replay is still pending
at this checkpoint and is not implied by the browser or wire pass.


## Required model acceptance

The parent reports final required-Quint replay passed in 3.11 seconds (package
3.120 seconds), with zero skips: 16 facts/steps, five measured source events
(one native and four record reads, including refusal/retry), zero store reads and
zero writes. Independent raw inputs were 97 spans, 97 logs and 16 metric rows,
yielding 12 native records and nearby populations 9/9/13. Ten semantic mutants
and the stage-read negative model control were rejected. These results are
attributed to the separate verification run and were not rerun by this reviewer.
No bounded 388 source/runtime/model acceptance item remains pending in this review.
