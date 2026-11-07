# Saved native interval context — architecture contract 364

Breadth selected by checkpoint 363. Reuse existing native interval planning and
fixed-question execution; no new planner, telemetry join, source engine or generic
origin framework. This source review runs no heavy checks and does not claim
implementation/runtime acceptance.

## Existing paths and the missing connection

`ui/measurements.mjs::renderTimeline` selects an actual rendered point, and
`appendRecordContext` builds logs/spans actions through `recordContextPlans` and
`NativeRecordContext.plan`. The plan retains exact nanosecond interval bounds,
checked environment representation and shared service/namespace/instance (or its
already-supported exact host case). Instrument name, attributes and scope are not
record filters. Live Measurements receives an `onRecordContext` callback.

`ui/captures.mjs::show` currently calls `Measurements.render` with `saved:true` and
no context callback. Add this entry for a verified loaded native capture using the
same renderer. Bind `canAct` and the callback to that exact capture page, receipt,
view generation, loaded visit and visible non-busy route. Do not make all service/
resource/native origins part of this first cut.

The existing `native-record-navigation.mjs` is deliberately live-specific: its
opener must belong to the live native host, its active predicate expects Explore
Measurements, it clears on leaving Explore, and its prose disclaims a durable
citation. Passing a saved-capture callback into it unchanged is incorrect. Use a
small saved-origin adapter or bounded explicit extension with its own captured
origin/return callbacks; keep the original capture owner authoritative.

## Closed data and action contract

- Only existing supported counter/histogram **export intervals** provide this
  saved context entry. Gauges/stocks do not acquire invented intervals; aggregated
  histogram time buckets do not become export intervals. Existing resource/profile
  restrictions stay intact rather than being removed to make links appear.
- Resolve the selected point from the held captured result/stream and its actual
  interval. Validate/recompute the record question using the existing pure planner
  (and existing pure HTTP plan endpoint if the app normally confirms it). A forged
  detail question or a row from another result cannot supply authority.
- Keep the full immutable native citation: capture ID/original receipt and custody
  time, original measurement question, exact stream, interval, native basis and
  original source read clock. Detach metadata before retaining it; do not alter
  capture bytes, definitions, stream labels or evaluated results.
- Preserve exact `[fromNS,toNS)` through the existing canonical timestamp predicate
  and enclosing millisecond bounds. Typed current/effective policy, namespace
  missing/empty and resource instance remain exact. Do not widen to a service
  window, attach metric dimensions to records, or equate time context with metric
  contribution/exemplars/causal relationships.
- Clicking Logs/Spans in interval stages the checked fixed record question and
  shows its saved native origin. It reads zero telemetry. Pure planning and capture
  metadata are separate counted effects, not source acquisition. Explicit Run is
  the new logs/spans read, which may fail while saved native evidence stays valid.

## Saved origin and return lifecycle

Retain the original capture DOM/result and opener while records are explored.
The return action is explicitly **Return to saved measurement**. It cancels or
suspends current record work through the existing owner, restores the original
capture route/hash and retained interval/inspector/focus/scroll, and performs no
metric read or capture refresh. Reopening the same already-held capture should
not force another metadata GET merely to return; if a route owner needs one,
disclose/count it separately rather than claiming no store access.

Current capture identity/generation and action visibility must be checked before
staging or returning. A different capture, new revision, route replacement,
suspension, disposal or late callback cannot revive the old origin. Native saved
scope is not the current live Measurements editor; never overwrite that editor to
simulate a return. Keep only one current displayed context path so an old live
native return bar cannot compete with the saved-origin bar.

If a fresh record read is cancelled/refused or its publication arrives late,
return still restores the valid saved native origin. Failed planning must leave
that capture usable and unchanged. Repeated stage/back/second-stage schedules
must rotate local callbacks without introducing a global reference manager.
Sharing or capturing the record question preserves its ordinary filters; it does
not silently create a durable cross-signal citation artifact or atomic snapshot.
The explicitly displayed saved origin itself has a real immutable capture ID.

## Acceptance

Start with genuine saved counter or histogram data, including a typed-policy
case and exact sub-millisecond interval, namespace/instance and independently
retained source/custody clocks. Verify selection and Stage have zero telemetry
reads, explicit Run requests the exact planned record question, and Back restores
the same captured bytes/result/selected interval/opener with no metrics read.
Exercise a refused/cancelled record source, late response, wrong stream/question,
changed capture and stale old Stage/Back callbacks. Gauges/stocks/aggregated-time
and unsupported resources retain their existing explicit limitations. Test actual
saved-capture route entry, not just the pure native planner or live-origin helper.
Reuse existing independent Go/JS planning oracles and observed lifecycle facts;
do not invent a native-to-record causal join in a model.


## Candidate source acceptance

The frozen candidate adds a saved-native application owner that detaches/freezes
the retained receipt/result, checks measurement question equality, resolves the
actual held stream/export bounds and recomputes the existing context plan before
staging. Root Captures supplies the verified loaded receipt/visit and current
capture ownership; the pure helper is a data-consistency check, not a generic
custody grant. Saved anchor callbacks capture creation-time leases and renew on
resume without rerendering the original measurement DOM. Unsupported kinds and
synthetic buckets do not gain authority.

Captures retains a local native-origin ticket independent of the suspended view's
action lease. New capture loads, refresh, kept/document replacement and unrelated
navigation retire it. Explicit return suspends the record visit, restores the
same native capture/hash/opener/scroll and renews existing anchors with no GET or
metric read; the retained-return flag suppresses unrelated catalogue loading.
Failed refresh can resume the still-valid captured result without adopting a new
source. The app shares its small retained-record return helper with kept
observations, preserving that existing path while clearing competing live native
return presentation. ResourceReturn receives an optional citation/notice, and its
direct summary is correctly the first disclosure child.

Read-only source acceptance is closed with no remaining blocker. Actual owner
focused tests cover foreign stream/question/bounds/basis admission, callback
retirement/lazy renewal, same-DOM return and capture-route replacement/refresh
schedules. The parent reports the main full UI suite passing 1,239 tests and a
final focused disclosure-order check passing. Generated/native/model/actor/frame
gates are recorded separately below; this source review ran no heavy checks.

## Execution acceptance

The parent reports canonical 364 acceptance in 58.811 seconds (runner 60.728),
with all 20 actual facts passing across a typed counter→11 spans and an effective
histogram→12 logs. Each case made three explicit record source attempts (success,
refusal and pending); metric reads stayed zero, and the only capture store read
was the initial open. Return made zero metric/store reads. Pending cancellation
advanced aborts from zero to one in each case, with no late adoption or retired
Stage/Back revival. Full original capture bytes were unchanged.

The canonical native run included an earlier observed check in 5.46 seconds
(package 5.474), which root audit found replayed a fixed script after Go admission.
That check is superseded by the final offline observed `population_drilldown`
projection: **9.67 seconds (package 9.700), zero skips**. It maps actual recorded
fact names and cumulative Source counters into model labels/counts, accepts both
case projections and rejects four cloned actual-fact source/order mutants directly
in the model, bypassing Go admission. Six independent Go admission negatives
remain (three per case). The combined pending macro is logically split using
Stage's Before.Source and Read's After.Source; this is not an observed intermediate
browser snapshot. The model judges source/depth only; custody and IDs remain
independent Go/browser assertions. This repair changed verification only: product,
native UI and the canonical 58.811-second run were not rerun or altered.

The positive actor passed 492 snapshots/seven stages in 35.156 seconds, observing
one explicit record source read, zero metric/store/writes and held DOM/focus return.
Full UI passed 1,239 tests; Hegel passed 40 cases in 3.94 seconds and the existing
Quint wrapper passed in 5.38 seconds. The Go 1.27.3 image's health endpoint returned
200 and six served assets matched the build exactly.

The shared kept-context path also passed its actual 328 regression on the same
binary: stage, Run, Back and pending cancellation, in 18.5 seconds (runner 20.4),
with six genuine 327/328 frames refreshed. Design accepted three new actual saved-
native frames; they establish usable task-positioned content, not an above-the-
fold density claim. Source and execution acceptance are closed.
