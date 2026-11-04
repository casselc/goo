# Typed service Counter and explicit Histogram values — design 320

Status: bounded implementation source and all three actual native frames accepted.
Read with [319 native Gauge review](NATIVE_ENVIRONMENT_DESIGN_319.md) and existing
native discovery/measurement contracts.

## User outcome and finite scope

A service selected under current-only/effective named, empty or unreported policy
can read its supported Counter intervals and explicit Histogram distributions
without changing population. Reuse the existing instrument discovery → source-free
Preview → explicit Run flow. This is another admitted native family on the same
checked path, not a new explorer, catalog or evaluator.

Environment precedence and full raw resources remain unchanged. Current and
legacy-key resources are distinct exact streams even when their effective name
matches. Preserve the selected service/namespace and original fixed window through
preview, execution, share/package/capture and return where those existing codecs
admit the family. No string fallback or cross-stream aggregation is implied.

## Counter interpretation

Show actual evaluated interval delta/rate with its unit and exact endpoints, not
a metric-record count or Gauge-like latest scalar. Keep accumulation temporality,
reset, gap, missing boundary and valid/invalid interval support inspectable under
the existing kernel rules. A reset or unavailable interval must not become a
negative/zero rate by presentation, and missing time must not be interpolated.

The default native presentation should distinguish the latest retained interval
from full-window/fleet values. Preserve stream diagnostics and coverage alongside
the value. Do not average per-stream rates or assert a service total simply
because streams share a name/environment. No new counter domain or permissive
integer/decimal coercion is authorized by policy propagation.

## Explicit Histogram interpretation

Show actual bucket counts/distribution with reported measurement unit, exact
bounds, observation count and interval/support. Use the existing bucket table /
heatmap and exact values, retaining open-tail and layout meaning. Counts are
observations per bucket/interval, not a request rate or Gauge unit.

Do not average scalar pXX values, invent a fleet percentile or combine incompatible
bucket layouts/temporality/resources. If existing kernels report a quantile range,
retain that range and native unit rather than turn it into a precise point.
Missing/unavailable intervals and layouts remain distinct from zero-count bins.
Clipped/unsupported distributions must not silently enter an apparently complete
chart. Existing exact tables remain available when geometry is unavailable.

## Existing interval record context

For a valid selected Counter/Histogram interval, the existing record-context
continuation stages its exact typed fixed interval and reported stream identity,
without source IO. Subsequent explicit Run reads records; it is temporal/contextual
support, not proof of causal membership. Retain service, namespace, instance and
supported dimensions through the existing checked constructor. No wider lookaround
or fresh time window is inferred.

Gauge remains a point observation, with its existing UI boundary; interval-family
delivery must not claim a new Gauge context button. Back restores the paused
service/Activity and visible opener with no read/resume. Pending/retired native
requests and old interval controls keep current lease/cancellation checks.

## Unsupported families and exact scope

Expose an executable prototype only after definition, SQL/source membership,
returned stream, evaluation, browser and persisted codecs all admit the typed
family. Unsupported stock/state/lifecycle/exponential Histogram/entity bindings
remain a precise pre-read refusal, not zero or Not observed. Existing current-only
legacy support keeps its established meaning; no retry as a weaker name-only
query after a typed failure.

The UI needs no extra warning wall: family support in current metadata plus local
value/coverage diagnostics is sufficient. Keep checked policy, window and unit
visible/readable in the existing scope/Read facts/identity surfaces. Raw
current/legacy differences must remain inspectable and distinguishable.

## Acceptance

Use actual SDK/native data under effective legacy fallback and accepted empty /
unreported policies, including conflicts/current-empty precedence. Discover,
Preview and Run a supported Counter and explicit Histogram through the native UI.
Verify actual per-stream delta/rate/reset/support and bucket distribution/unit/
bounds/count rather than presence metadata. Keep independent raw resources apart.

Select a valid interval and stage existing record context with no read; explicitly
Run it and return to the same held native/service/Activity state. Share/save/reopen
supported typed definitions without stripping the selector. Exercise reset/gap/
invalid and incompatible-layout controls through meaningful semantic tests, and
report which cases are native versus pure/property coverage accurately.

Actual compact frames must show values/distributions and their interpretation,
not only an editor after Run. A cropped frame cannot prove every policy/raw
identity; membership/effect assertions remain separate. No speculative styling,
new navigation controls or broad native-suite migration is needed for this cut.


## Final source and actual 320 acceptance

The final source widens typed service admission through the existing per-stream
Counter and explicit duration Histogram kernels/codecs. The Histogram profile is
specifically `s`/`ms`, not every arbitrary histogram unit/family. Optional counter
domain rules, raw resource identity, interval/selector bounds and checked
membership remain shared. Unsupported resource/grain/families still refuse before
source work; no new source executor, fallback environment or stream aggregate is
introduced. Architecture reports 60 generated Counter/Histogram cases passed,
covering policy/namespace/raw identity, exact rates/reset unknowns, histogram
subtraction/layout/pinning and context/ownership counterexamples.

All three actual frames are accepted. The
[instrument frame](../journeys/screenshots/320-effective-service-instruments.png)
shows the explicit duration Histogram (`s`, delta temporality) and Counter
(`{request}`, cumulative monotonic) metadata with the existing Preview measurement
actions. It is metadata capability proof, not executed value proof.

The [Counter intervals](../journeys/screenshots/320-effective-counter-intervals.png)
frame shows a valid one-second **25 delta / 25 per second** row followed by
**Not recorded · unknown · counter-decrease**. The selected interval also says
Unavailable/unknown rather than displaying zero or a fabricated negative rate.
Unit and UTC endpoints remain inspectable; two independent streams are shown,
with the other current stream collapsed. Logs in interval / Spans in interval are
existing temporal context actions with an explicit stage/Run and no-causality
explanation. An interval with unknown metric arithmetic can still have valid
bounds for contextual investigation; the links do not assert an observed rate.

The [Histogram interval](../journeys/screenshots/320-effective-histogram-intervals.png)
frame shows **six observations** and **P95 bucket range (0.5, +∞) s · no
interpolation**, retaining the open upper tail instead of a precise/averaged pXX.
The exact bucket arrays and separate current four-observation stream are not both
visible in this crop; those are native/API/property assertions and inspectable
retained details, not facts inferred from the screenshot. The same applies to
policy/raw environment keys and exact membership in these value-focused crops.

The verification owner reports all six native main cases passed: current-only /
effective named for both families, plus empty and unreported Counter cases. The
current Counter sequence's **ten** delta/rate and legacy **25**, reset unknowns,
Histogram counts **four/six** and exact bounds are asserted separately from the
three frames. The storyboard describes preview/share/return zero-I/O and retained
paused context; final execution/campaign timing belongs to validation. No browser
or Docker work was performed by this reviewer during the actor run.

No visual or source UX blocker remains, and no speculative styling, navigation or
additional family changes are requested. This is typed per-stream interval-family
delivery; it does not imply a new Gauge context button, fleet distribution or
entity migration.
