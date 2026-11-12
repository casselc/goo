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
| Recover | Discover with an unavailable source, then explicitly retry. | Failed discovery withdraws prior choice authority; no automatic retry. |
| Leave | Depart while discovery is pending. | A retired request cannot restore choices or stage a destination. |

The fixture includes 132 baseline and 168 comparison spans in the selected
service/namespace, plus 12 empty-version and 8 missing-version spans. Distractors
cover empty/missing/foreign namespaces, current/legacy environment precedence,
other services, Client spans, boundary timestamps and conflicting span-attribute
version values. Independent raw-row selection checks the exact population and
full duration accounting.

Actual documentation frames are generated into
`docs/journeys/screenshots/service-versions-picker-{1440,390,320}.png` and
`service-versions-count-comparison.png`. Ignored report files remain under
`artifacts/service-versions415/`; the test stops its isolated runtime in `finally`
without deleting existing telemetry or captured evidence.

The discovery's **Read details** disclosure exposes acquisition time and the
independent count/group read clocks without issuing another request. The compact
status keeps the unreported/empty counts visible after ordinary versions are
chosen. Count and mean comparison runs each have the observed six-SQL budget;
choosing a measure and staging its intent remain local operations.

With `GOO_SERVICE_VERSIONS_BOMBADIL=1`, a 45-second campaign independently opens the
stored inventory scope, discovers versions, chooses cohorts, runs a comparison
and returns to the retained service. It rejects unexpected source fanout,
publication and replacement of the held parent evidence.

The presentation refinement also checks primary values at 1440/390/320px: paired
numbers align even when labels wrap, chart chrome stays bounded, exact read facts
remain keyboard-accessible, and calling the folded baseline-refresh handler cannot
read telemetry. Additional 390/320px comparison frames complement the picker
frames. The shared timeline retains all exact unknown intervals without repeated
question-mark glyphs.
