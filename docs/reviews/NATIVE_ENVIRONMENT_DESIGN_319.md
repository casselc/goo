# Native values under a checked service environment — design 319

Status: bounded implementation source and all three final actual frames accepted.
Read with [typed record-service adoption](SERVICE_ENVIRONMENT_DESIGN_318.md) and
[existing native instrument discovery](../NATIVE_INSTRUMENT_DISCOVERY.md).

## Useful outcome

A service using a legacy/effective, empty or unreported environment can inspect
actual supported native values without changing the checked population to a
name-only scope. Metric data-point presence from 318 is not a metric value. This
cut should deliver a useful end-to-end native family, not merely enable the
instrument list or replace an unavailable note with zero.

A finite Gauge-first family is acceptable if the whole path preserves policy,
exact resource/stream identity, units, time and quality. Other families remain
explicitly unavailable until admitted end to end. Entity-bound metrics, deployment
and annotation identity are not implicitly migrated by native Gauge support.

## Existing flow, explicit acquisition

Reuse the current service Native instruments discovery/picker and measurement
owners. The existing explicit Read service overview may perform clearly declared
bounded discovery/default-value reads under its current policy; a separate
Discover button is not required merely for symmetry. Existing Apply and reread
overview remains explicit when changing instruments and must state the actual
read effect. Mounting a disclosure or inspecting metadata cannot acquire values.

The checked service/namespace/environment policy and original fixed window stay
visible in current scope/read facts. Discovery counts remain labelled data-point
records or observed streams, never a Gauge value. A supported native result must
show the actual value and unit, latest retained observation endpoint (or family
interval), read clock, support/coverage and applicable diagnostics. Do not turn a
latest Gauge observation into a whole-window mean, rate or service-health verdict.

Open stream question, retained details and return continue through their existing
owners only where the typed definition/share codec can honor the same policy.
Unsupported downstream actions must refuse before source work rather than open
a legacy-string question. No second native explorer or catalog is needed.

## Independent streams, not alias-normalized values

Effective environment selection uses current-key presence and legacy fallback;
it does not authorize coalescing raw resource identities. Current-key and legacy-
key resources remain separate exact streams even when service, instance,
instrument, attributes and effective environment name otherwise match. Never
average, sum, deduplicate or pick one as “the same instance” without an explicit
existing compatible aggregation contract.

The default rows should be distinguishable enough to understand their values.
Existing exact stream/resource details must expose the actual raw environment
keys and identity. If compact summaries would be identical, use actual supplied
resource context to distinguish current versus legacy key rows; do not infer a
source key from a normalized name or assign a synthetic identity. Values for
empty and unreported resources remain distinctly labelled, not blank/wildcard.

Units and exact measurement grain remain visible. Keep no-data, invalid,
missing/stale endpoint, conflict and partial support as their existing diagnostic
states. An unread/unsupported stream is not zero, and independent native values
are not atomic observations with the record overview.

## Deliberate family and identity limits

If this cut supports Gauge only, discovery may show Counter/Histogram metadata
as unavailable with a precise reason, but neither UI, planner, direct endpoint nor
an imported definition may execute a falsely supported typed-policy family. Reject
before source acquisition; do not try a string fallback after failure. An existing
current-only family path can remain supported under its own checked contract.

Likewise entity-bound instruments, annotation/deployment metadata or another
identity family remains unavailable when it cannot preserve selector semantics.
Keep those capability facts in the existing prerequisite/subpanel area rather
than adding a warning checklist before the useful values. Do not show Not observed
for deliberate nonacquisition.

## Acceptance as a real task

Start from an effective legacy-SDK service in Activity, stage/read the service
context and discover/select a supported native instrument. Read actual values
explicitly through the existing owner. Show current-key and legacy-key exact
streams with otherwise matching identity where possible; verify their actual
values remain separate and their resource difference is inspectable. Include
current-empty precedence, conflicting current/legacy values and unreported scope
in the admitted policy checks.

Verify actual value/unit/endpoints and support, not metric record counts. Exercise
instrument selection, metadata/disclosure zero-IO, pending request cancellation,
stream continuation where supported and Back to the same paused service/Activity
with no read or tail resume. Counterexample families/definitions must refuse
before reads. State native versus unit/property coverage accurately rather than
claiming a full policy/family matrix from one screenshot.

Review actual compact discovery/value/identity frames before accepting the UI.
Keep source/read-effect and exact membership assertions separate from static
presentation. No fleet aggregation, generic entity migration, new metric evaluator
or broad suite parity belongs in this finite round.

## Frozen source review and metadata frame

The source uses the existing service instrument discovery → source-free Preview
measurement → explicit Run measurement path. It admits the Gauge-only typed scope
through the same definition/source/stream codecs rather than creating a new
explorer. Current-only compatible service paths retain their existing bridge;
other typed native families/entity bindings refuse before source reads. Raw
current/legacy resource keys are included in compact stream identity labels and
remain available in structured full identity. Actual observations are not merged.

The actual service-instruments frame is useful discovery proof: an inventoried
Gauge with `items` unit and Preview measurement appears in the existing Native
instruments panel, with observed-metadata limits explicit. It is not value proof.
The first value screenshots still showed the open editor after Run; they require
reanchoring to actual value/unit/identity/read context before final visual
acceptance. No product scrolling or extra control change was requested while the
native run/campaign is active. Semantic source/oracle results remain separate
from screenshot presentation.


## Final actual 319 acceptance and proof limits

All three final reanchored frames are accepted. The
[service instruments](../journeys/screenshots/319-effective-service-instruments.png)
frame shows discovered Gauge metadata, `items` unit and the existing source-free
Preview measurement continuation. The
[effective Gauge values](../journeys/screenshots/319-effective-gauge-values.png)
frame now shows an actual selected observation **456 items**, its UTC endpoint /
observed state and the retained **Recorded value 456** row. It also shows the next
current stream collapsed and **two independent streams**, with no automatic
interpolation or merging claim. This is actual value proof rather than a metric
record count or an editor-only screenshot.

The [unreported values](../journeys/screenshots/319-unreported-gauge-values.png)
frame shows the actual **567** recorded value and UTC observation with one
independent stream. Its current crop does not show the unit/policy heading or full
raw resource identity; likewise the effective crop does not show all scope/raw
legacy keys. Policy membership, raw stream separation, units and exact fixed/read
scope are established by source/native/API assertions and inspectable details,
not by pretending these cropped images contain every fact. No additional image
or product-polish pass is needed for this finite acceptance.

The existing selected-Gauge inspector still says the observation has no interval
and exposes interval-only context navigation. The exact-point Gauge context
projection supported by API/protocol tests is **not** a newly delivered Gauge UI
navigation button. Keep that boundary explicit in docs rather than imply a
counter/interval workflow exists for point observations.

The verification owner reports final main-only native replay green in **31.8s**
(**34.3s** including runner), after the four actual Gauge-policy/value oracles had
passed. Root reports the optional actor separately green with **266 snapshots**
and one explicit Read. These reports establish effects/completion separately from
the three frames; the review does not infer a full native-family/mobile matrix.

No visual blocker remains. No further source changes, scrolling behavior,
explorer controls or styling were requested. The original policy and exact-stream
contracts remain authoritative; this is Gauge-family delivery, not entity or
counter/histogram migration.
