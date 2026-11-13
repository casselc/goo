# Recorded Error version comparison review — 419

Status: bounded source/interaction contract and the four actual fraction/zero
frames accepted. Browser and model execution are separate gates in validation.
This reviewer did not implement the production changes and ran no heavy checks.

## Product and semantic contract

The existing service-version entry gains explicit Recorded Error count and
Recorded Error fraction choices. Both stage a fresh, summary-only comparison in
the existing population workspace. There is no new dashboard or discovery owner.
The original exact qualified service, role/operation predicates, operation
grouping, typed environment policy and fixed window remain the cohort context.
Reported resource versions remain explicit distinct choices.

Measurement directives follow the selected measure, preserving the accepted 415
projection: discovery requests counts, mean requests duration, and either Error
choice requests span status. A duration summary on the originating widget need
not be requested again for an Error comparison. This changes requested
measurements rather than silently broadening source predicates or grouping.

Each operand retains Error, explicit OK, unset and unknown status counts whose
sum must equal its captured population count. The fraction uses exact same-row
Error/count support. Both Error count and fraction presentation show the
numerator/denominator and unset/unknown context; unset is not inferred success.
Recorded span Error does not establish HTTP failure, an SLO or deployment
causality. Independent read clocks and quality remain independent.

The signed fraction change uses a distinct `ratio / difference` descriptor and
exact rational value in unit `1`. Percentage points are a display scale of that
supplied difference, not a relative percentage or another bounded fraction.
An empty population has no defined fraction; diagnostic or incompatible changes
remain unavailable rather than acquiring a synthetic zero.

## Ownership and interaction

Application adapters retain query intent, explicit source acquisition and held
return. Passive shared components receive validated supplied data and cannot
subscribe, fetch or publish. Fresh status compatibility requires matching status
profiles and all existing context checks. Captured status declarations and
grouped status comparisons remain refused in this slice.

Native disclosure and visibility checks continue guarding discovery and staging.
Changing the measure is local configuration; staging does not read telemetry.
The planned journey must verify explicit Run, stale controls and Back to the
actual held service context without a reread.

## Finding resolved during review

Initial source inspection found that the Error count choice would omit visible
unset/unknown context because only ratio fields requested fraction support.
The shared value-cell implementation now reuses the recorded Error fraction's
reference facts for the Error count field as well. This resolves the source
finding without inventing another status widget or changing the data model.

## Acceptance limits

No remaining source-contract blocker was identified in the reviewed changes.
This is not a whole-product accessibility certification or vendor parity claim.
Real-source count conservation, exact generated arithmetic, observed read-budget
replay and the final desktop/mobile/zero/diagnostic frames still require their
own executed results before milestone acceptance.

## Actual visual acceptance

Independently inspected the actual fraction frames at 1440, 390 and 320 pixels,
and the 390 pixel empty-baseline control. The fraction view shows 22/132 spans
(approximately 16.67%) against 42/168 (25%), with an approximately 8.33 percentage
point change. Both sides retain visible unset and unknown counts. The common
bars, paired labels and change remain legible at each inspected width; narrow
layouts wrap supporting facts without clipping values or requiring horizontal
page scrolling.

The zero-denominator frame correctly displays `No data` for the 0/0 baseline
fraction and `Unavailable` for the change, while preserving the selected 25%
and its support. It does not present an empty population as a healthy 0% or mark
the valid selected side empty. The comparison frame remains observed, with
per-side value states explicit.

Native calculation and scope/read disclosures remain available. Source review
confirmed existing table/header relationships and decorative geometry hidden
from assistive technology. Text values and numerator/denominator support convey
the result independently of color. These are bounded Chromium visual/source
checks, not a screen-reader or cross-browser certification. No visual blocker
was found in the four inspected frames; Error count rendering has source-level
review but no separate count screenshot was present in this set.

Actual browser, actor and native-model execution results remain independently
recorded in [validation](../../VALIDATION.md).

Final execution is recorded in [validation](../../VALIDATION.md): 1,591 frontend
checks, affected Go packages, OpenAPI, the actual version journey/45-second actor,
and generated/native replay pass. The independent campaign is ordered after the
continuous main ledger; no model budget exception was introduced.
