# RFC 0002 Populations time and evidence

Status: Proposed. Date: 2026-10-16. This RFC specifies target behavior; current deviations are listed in KNOWN_ISSUES.md.

Implementation note: the separate Go lifecycle pilot now exercises the counting,
cohort, exact-deadline and frozen-basis subset on normalized events. See
[LIFECYCLE_PREPARATION.md](../LIFECYCLE_PREPARATION.md) for implemented boundaries;
SQL preparation, durable correction publication and UI integration remain pending.

## Counting and identity

Declare the subject before aggregation: server call, client attempt, accepted reservation, processing attempt, message, or batch. Identity must include its scope, such as environment, tenant, and business ID. Retries may increase attempt counts while leaving reservation counts unchanged. A batch span cannot establish per-message outcome without per-message evidence.

Declare immutable event identity separately from subject identity. Duplicate delivery of the same event contributes once. Different immutable events with conflicting acceptance or terminal facts invalidate the subject under the pilot's strict policy. If an application supports revisions, it must supply ordering and conflict-resolution semantics explicitly.

Missing identity or malformed timestamps are invalid records. Zero is a valid measurement and cannot stand in for missing or unparseable evidence. Invalid-record totals and their influence on eligibility must be visible. Excluding all malformed acceptance records silently can make a service appear healthier.

## Mature deadline cohorts

Capture evaluation time T once. Let D be the completion deadline and L the configured ingestion allowance. For a requested accounting window W, use cohort end C = T - D - L and acceptance interval [C-W, C). Every eligible accepted subject contributes once after maturity.

Fetch terminal evidence for those subjects through the relevant evidence cutoff; do not restrict outcomes to the same interval as acceptances. A late terminal can distinguish 'completed late' from 'no completion' while remaining a deadline failure. An inspector may display recent pending subjects separately from the mature SLI.

| State | Condition |
| --- | --- |
| Good | Successful terminal outcome at or before accepted time plus D |
| Bad | Failed terminal outcome, late success, or missing completion after maturity when coverage permits the inference |
| Pending | Deadline and allowance have not yet elapsed |
| Unknown | Evidence loss, ambiguous records, violated clock assumptions, or unresolved conflicts prevent classification |
| No data | No eligible subjects exist in the selected mature cohort |

Compare deadline timestamps with an explicitly inclusive success boundary. Prefer integer epoch nanoseconds or microseconds over floating-point equality for the reference semantics. Tests should include one representable tick before, exactly at, and one tick after the deadline.

Event time, ingestion time, and evaluation time are distinct. Collector receipt does not automatically provide a reliable business event time. State clock assumptions and expose violations. Future timestamps, negative durations, and excessive skew must not become ordinary good results.

## Coverage and corrections

An unsampled SDK or exporter does not prove complete collection. Delivery loss, restarts, and missing instrumentation remain possible. Coverage policy needs observable inputs such as reconciled business counts or authoritative lifecycle state. Label the sample's current evidence demo-unverified.

When evidence is impaired, the pilot's strict policy returns an unknown SLI for the affected scope and preserves counts and diagnostics. A later policy could provide bounded estimates or partial scopes, but must not silently drop unknown subjects from the denominator.

Lateness allowance defines a closure policy, not certainty. Recompute within a correction horizon and version revised contributions. Deduplicated subject results need replacement/retraction or generation-based publication; a second insert cannot be counted as another subject. Historical changes do not replay old notification pages. Retain the original evaluation receipt and add corrected evidence.

## Measurement and inspection

Compute full-population totals in an aggregate query without a display LIMIT. A separate evidence query can page subjects from the same normalized population and evaluation context. Changing page size or order must not change the SLI.

Record numerator, denominator, good, bad, pending, unknown, and invalid counts with window bounds. Aggregate ratios from aggregate counts. Define zero denominator as no-data. Do not replace unknown with success. Preserve underlying counts for 30-day objectives when raw retention is shorter.

Scalar quantiles cannot be averaged into a fleet percentile. Use underlying observations or compatible mergeable distribution states. Counter rates need instrument identity, temporality, start time, and reset handling. Histograms need bucket/sketch compatibility, units, and aggregation temporality. Consult OTel's data model rather than treating every exported Value as interchangeable.

## SLO and alarm semantics

An SLO adds a target and accounting window to an SLI. Burn rate = bad fraction / (1 - target). An alarm adds evaluation cadence, windows, condition, pending/recovery rules, and notification route references. Its full evaluation and evidence state are distinct from notification delivery state.

Each evaluation is true, false, or unknown. Unknown cannot establish recovery. No traffic and telemetry outage remain distinguishable. Minimum-volume guards, if used, affect paging conditions rather than removing failures from SLO accounting. Native metric evidence and trace exemplars remain different measurement bases.

An incident receipt records expression and package/profile revisions, evaluation time, bounds, parameters, counts, quality state, and notification episode. Reproducing exactly the original numbers additionally requires retained input snapshots or pinned results.

## Required semantic fixtures

Use a fixed clock and manually specified expected outcomes for duplicates, retry success, no completion, late completion, exact boundaries, acceptance-window edges, terminal records outside acceptance bounds, malformed fields, cross-tenant identity, conflicting facts, zero traffic, and evidence outage. Include more than 1000 subjects and multiple inspector page sizes.

Run the same normalized expression and fixture data through reference and ClickHouse implementations. Compare population identities and metadata as well as numerical totals. Matching SQL and reference bugs is possible; the independent expected results are the third oracle.
