# Compare reported service versions

Executable: [service-versions.spec.js](../../journeys/service-versions.spec.js).
This storyboard follows ordinary OTLP spans in an isolated Go runtime backed by
real ClickHouse telemetry. The validation record distinguishes executed checks
from planned behavior.

An operator has selected one qualified service and wants to compare two reported
versions in the same fixed window. Version is the resource `service.version`;
a similarly named span attribute does not substitute for it. Missing and empty
values remain separate, and deployment causality is not inferred.

| Step | Operator interaction | Observable contract |
| --- | --- | --- |
| Select service | Open an exact namespace/service from the inventory. | Fixed environment policy and UTC window bind the overview. |
| Discover | Open Compare versions and choose Discover versions. | One explicit aggregate request; full population count and version groups, no record projection. |
| Choose | Pick baseline and comparison versions. | No defaults or automatic execution; resource missing/empty remain visible. |
| Stage | Stage captured-count comparison. | Existing population workspace opens with two fixed version predicates; no source read. |
| Compare | Explicitly Run. | Full counts match independent source selection even though the delivered page is bounded to 100. |
| Return | Back to service. | The same parent widget DOM/evidence returns without rereading. |
| Inspect duration | Choose mean duration, Stage, then Run. | Both sides use independent full-query duration sums/counts; no averaged scalar percentile. |
| Inspect Error context | Choose Recorded Error count or fraction, Stage, then Run. | Independent source partitions conserve Error/Ok/Unset/Unknown counts; fractions use each full population. |
| Empty denominator control | Temporarily compare a deliberately supplied empty baseline. | Error count remains known zero, fraction/change become unavailable; no source-empty claim. |
| Recover | Discover with an unavailable source, then explicitly retry. | Failed discovery withdraws prior choice authority; no automatic retry. |
| Leave | Depart while discovery is pending. | A retired request cannot restore choices or stage a destination. |

The fixture includes 132 baseline and 168 comparison spans in the selected
service/namespace, plus 12 empty-version and 8 missing-version spans. Distractors
cover empty/missing/foreign namespaces, current/legacy environment precedence,
other services, Client spans, boundary timestamps and conflicting span-attribute
version values. Independent raw-row selection checks the exact population and
full duration accounting. Baseline status is 22 Error, 22 Unset and 88 Ok;
comparison status is 42 Error, 42 Unset and 84 Ok. Exact recorded Error
fractions are 1/6 and 1/4, a difference of 1/12 (approximately 8.33 percentage
points), independently computed from all matching source rows. HTTP response
attributes are distractors rather than an error classifier.

Actual documentation frames are generated into
`docs/journeys/screenshots/service-versions-picker-{1440,390,320}.png` and
`service-versions-count-comparison.png`. Ignored report files remain under
`artifacts/service-versions419/`; the test stops its isolated runtime in `finally`
without deleting existing telemetry or captured evidence.

The discovery's **Read details** disclosure exposes acquisition time and the
independent count/group read clocks without issuing another request. The compact
status keeps the unreported/empty counts visible after ordinary versions are
chosen. Count, mean and status comparison runs are checked against a six-SQL budget;
choosing a measure and staging its intent remain local operations.

With `GOO_SERVICE_VERSIONS_BOMBADIL=1`, a 45-second campaign independently opens the
stored inventory scope, discovers versions, chooses cohorts and recorded Error fraction, runs a comparison
and returns to the retained service. It rejects unexpected source fanout,
publication and replacement of the held parent evidence.

The presentation refinement also checks primary values at 1440/390/320px: paired
numbers align even when labels wrap, chart chrome stays bounded, exact read facts
remain keyboard-accessible, and calling the folded baseline-refresh handler cannot
read telemetry. Additional 390/320px comparison frames complement the picker
frames. The shared timeline retains all exact unknown intervals without repeated
question-mark glyphs.

## Recorded Error extension (419)

The executable extension adds count and fraction runs to the existing flow,
followed by held Back after each. Fractions show each exact Error numerator and
population denominator, with unset status retained explicitly. Display changes
use percentage points rather than relative percentage; counts use records.
The controlled empty baseline is produced from supplied comparison data and is
labeled in the report. Discovery itself cannot discover a zero-record version.
This is a denominator-state control, not a claim that the source changed.

The extended run regenerates picker/count frames plus
`service-versions-error-fraction-{1440,390,320}.png` and
`service-versions-error-zero-390.png`. The live journey and 45-second actor passed in 1.3 minutes: 27 main facts,
zero page errors, four explicit population-comparison runs and one pure supplied-
zero control. The actor completed across 816 snapshots, retaining the parent
widgets/evidence and avoiding publication. The isolated runtime stopped cleanly.
The independent actor runs after all main facts, keeping that ledger contiguous.
The recorded per-gesture ledger distinguishes initial dashboard source reads
from the comparison runs; native replay and independent visual acceptance are
recorded in [validation](../../VALIDATION.md).
