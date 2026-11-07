# Reuse a native histogram view — design 371

Status: final source and actual-frame acceptance complete for the bounded exact-export
Histogram/Table pilot. Canonical runtime/actor/model results are recorded separately
in VALIDATION.md; this review does not infer their completion from screenshots.
The original timeline/heatmap proposal below is superseded by the closed contract.

## One meaningful existing display choice

Measurements already renders native interval observations, selected interval bucket
distributions and compatible-layout bucket heatmaps. It also saves immutable
measurement packages and previews them without reading source. A small useful
first reuse contract is a default **Interval observations** versus **Bucket heatmap**
presentation for one checked histogram question. Reuse actual native renderers,
not an ordinary-record template alias, new chart kind, WYSIWYG or universal registry.

The choice is presentation configuration, not a saved stream observation or selected
interval. Do not persist a bucket-layout index, selected stream/interval position,
result value, read clock or specimen destination scope as portable presentation.
Those belong to a held result/capture. The existing fixed/native question remains
source authority, with explicit preview then Run.

## Current destination and compatibility

Apply only to an admitted current histogram destination question with a closed
compatibility rule for exact metric/unit/source kind and supported native shape.
Root/architecture must specify that rule, not infer it from a display label. Keep
the destination's service, environment policy, resource identity, exact stream if
pinned, temporality, fixed UTC/nanosecond bounds and all definition constraints.
No donor query/result is silently substituted. Wrong kind/unit/unsupported input
must refuse without mutating destination draft or held evidence.

Saving/opening/configuration inspection and local Apply read no telemetry. Only
explicit Run evaluates the destination; its result has its own read clock. A capture
continues to preserve evidence under its own custody. Reopening a saved definition
is not reopening original source observations or a combined cross-signal snapshot.

## Native uncertainty and limits

Interval observations show at most the latest 200 supplied intervals; heatmap shows
at most the latest 32, with columns representing intervals rather than elapsed time
or rates. Full evaluation input remains independent of those display limits.
Compatible bucket-layout selection remains explicit; changed/incompatible layouts
and unavailable intervals stay outside/unknown, not zero or silently rebinned.
Keep exact units, temporality and stream identity independent; no fleet percentile,
merged delta/cumulative stream, request rate or complete-window guarantee.

A histogram export interval is distinct from an aggregate histogram time bucket.
The first contract should reject aggregate time-bucket questions unless separately
closed; the display label cannot erase that grain. Complete-input requirements and
unknown whole-window coverage remain the existing typed model's facts. A preference
must not force an unavailable view to fabricate values or disguise a refusal.

## Readiness decision and genuine acceptance

Implement only if one bounded versioned presentation field can reuse existing
package/configuration validation, immutable save/open and current destination
admission without a new metadata/query framework. Otherwise explicitly defer and
advance the next selected roadmap item; do not ship a nominal native template that
is really a count view.

Genuine journey: inspect a valid native histogram, choose one display preference,
save/reopen the definition, apply it to a second exact compatible current destination,
preview source-free and Run explicitly. Verify destination membership/units/intervals
and independent read clock, then reopen held/captured evidence without relabeling it.
Exercise wrong kind/unit, changing bucket layouts, unknown interval/coverage and
stale Apply/late response. Actual desktop/narrow frames should show the representation
choice, destination scope and native result/limits; source budgets and immutable
question checks establish authority independently of appearance. No new workspace
or broader grammar migration belongs in this cut.


## Closed root contract — accepted for implementation

Use the smaller existing selected-export **Native histogram display** choice,
Histogram/Table, backed by the typed native histogram snapshot and full supplied
bins. This supersedes the earlier timeline/heatmap preference: it reuses one dataset
and existing scoped plotting declaration, not two different projections.

One content-addressed native-histogram view contains the whole measurement question
and display declaration only. It contains no selected stream, interval bounds/index,
observed temporality, bins or result. Existing workbench publication supplies three
bounded metadata operations (create/load/reuse), without listing/registry/router.
Exact s/ms unit, per-stream histogram and supported declaration compatibility are
checked; aggregate histogram time buckets/resource-native shapes are excluded.
Actual delta/cumulative temporality remains post-Run evidence, never metadata proof.

The exact-ID pilot is acceptable with **Native histogram view ID** visibly distinct
from **Measurement package revision**. Save in existing Save/share or selected-known-
export context names configuration only, not captured bins. Load/Inspect shows its
title/kind and originating definition. Explicit Apply binds the display to the
CURRENT validated whole question, leaves that question unchanged, clears old result
and stages explicit Run. Wrong compatibility keeps destination/draft intact.

After Run, suppress automatic latest-interval selection for the applied view. Prompt
**Select an export interval to display**; the declaration renders only after an
explicit actual destination interval choice. Unknown interval means Distribution
unavailable, not an empty chart/zero. Inspect exact units, interval/stream authority,
full-bin versus displayed-bin limits and independent clocks through existing details.
No catalogue or new authoring surface is needed for this bounded pilot.

Acceptance must cover genuine Save/load/Inspect/Apply source0, whole destination
query equality, old result cleared, explicit Run and explicit destination interval,
Histogram/Table over identical bins, wrong unit/kind/grain refusal, unknown intervals
and stale metadata responses. Exact-ID friction is an honest pilot limit, not a
claim that team discovery/catalogue is complete.

## Frozen candidate source acceptance

Accepted the candidate source pending genuine runtime/frames. The existing selected
known-export disclosure saves Histogram/Table configuration only, with clear view
ID/title and no claim to retain bins. Native histogram view ID is explicitly distinct
from a measurement package revision. Load reports title/kind/unit and compatibility;
exact configuration remains supplemental. Apply preserves the current validated
question, clears old result, and stages explicit Run. After Run the applied view
requires a user-selected actual export interval; unknown distribution is unavailable,
not zero. Existing passive bin snapshot/admission is reused.

Metadata actions bind current plan generation/question identity and visible controls;
replacement/input/navigation invalidates loaded ownership. Save result cannot publish
to a departed owner. No material source UX blocker found. Exact-ID discovery friction
and raw configuration inspection remain honest pilot limits, not a new catalogue
requirement. Actual destination membership, old-handler/late-response refusal and
complete-bin Histogram/Table consistency require native/semantic evidence, not
source presentation review alone.

## Initial genuine-frame review

Viewed current original371-reused-native-display.png and371-unknown-native-export.png.
The reused Table visibly shows destination namespace/team-target, target attributes,
seconds boundaries0.02/0.2 and three bins with2/3/2 observations. Configuration-only
Save wording is explicit; retained intervals independently show known7 versus Not
recorded/unknown. Unknown export shows hatched unavailable timing, exporter-absent
not-zero explanation and selected interval Unavailable/unknown/no-recorded-value.
Its exact Distribution unavailable message is below this crop, so final visual proof
of that text awaits the requested task scroll. These initial frames support bounded
native semantics, not complete runtime acceptance or above-fold usability. Strict
input correction/rebuild and remaining canonical agent gates are still pending.


## Final source/frame acceptance

Re-viewed both regenerated original PNGs after the final scripted nine-fact flow.
Known destination Table retains three complete bins,2/3/2 observations and seconds
boundaries0.02/0.2, with target identity and configuration-only Save wording. The
final unknown frame now visibly includes **Distribution unavailable for this export
interval. Unknown is not zero**, alongside selected unavailable interval and exporter-
absent diagnostics; no zero distribution/chart is fabricated.

Bounded source/frame UX acceptance is final. These are desktop task-positioned
frames with internal inspector scrolling, not narrow-layout or first-entry proof.
Exact-ID reuse remains a pilot limitation. Whole destination identity, source budgets,
late-owner refusal and runtime/actor/model results remain independently recorded by
root in VALIDATION.md; no actor completion claim is made here while its infrastructure
rerun is pending. No further source UX change is requested.
