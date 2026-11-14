# Numeric severity acceptance — 435

Status: accepted. [Contract](../ACTIVITY_NUMERIC_SEVERITY.md),
[storyboard](../journeys/activity-severity.md).

Independent architecture and UX/accessibility/standards reviewers accept the
bounded extension. They reviewed source and focused tests; the UX reviewer also
inspected genuine 1440/390/320 frames. Implementation and native-verification
owners separately exercised their respective boundaries.

## Accepted behavior

- One logs-only grammar flag carries exact numeric partitions through count,
  groups, time buckets, ordinary results, captures and packages. SeverityText
  remains independent. Unrequested/malformed summaries fail admission; every band
  must conserve across independent reads before the snapshot is actionable.
- The existing volume chart, native selector and shared controls provide local
  band inspection with a common all-log scale. A separate Filter action stages
  exact numeric predicates; Apply is the source-read boundary. Unsupported
  predicate budgets withdraw actions without hiding supplied counts.
- Explanation permits only a declared numeric-band refinement, reconstructed as
  the complete original parent AND the known numeric clause. Arbitrary replacement
  selectors remain rejected. Post-pause guards and opener focus return remain intact.
- Standard log rows carry physical SeverityNumber, with matching strict UInt8
  readiness and deterministic ordering. Held guides count only matching reported
  values, without synthesizing missing numbers or borrowing source totals.
- Unsupported generic datasets, cohort and rate profiles withdraw or refuse
  explicitly. Their later adoption is separate work, not an inferred capability.

## Integration findings resolved

The real browser exercise exposed two contracts that isolated controller tests did
not initially connect: the explanation owner rejected the refined selector, and
the ordinary row adapter omitted the numeric value needed by the held guide.
Both are fixed with explicit validation and actual reported data, rather than a
generic selector bypass. Focused tests now execute the volume-to-explanation
handoff; the real journey verifies a nonzero band-filtered guide.

The fixture distinguishes 125 matching source logs, 100 held rows, 112 numeric
Error source logs and 90 held Error observations. The independent actor checks
local read budgets, exact scope and explicit Apply. Native replay rejects eleven
altered traces, including equal-total band drift and incorrect held-guide facts.
See [validation](../../VALIDATION.md) for executed checks.

## Limits and next work

No physical assistive-technology speech or cross-engine acceptance is claimed.
Agreement between independent reads does not prove ingestion completeness or a
transaction. Zero cannot distinguish omitted producer severity from explicit zero.
The next consolidation slice should retain these closed refinement and custody
contracts while aligning log-volume and span-latency interaction where appropriate.
