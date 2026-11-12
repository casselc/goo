# Independent service-triage review — 412

Date: 2026-11-12. Disposition: **accepted in the exercised scope**.
The reviewer did not author implementation or run a concurrent browser/test
campaign. Source inspection and actual runtime evidence were independent of the
implementation agents; the coordinator kept heavy execution serial.

Reviewed candidate Go binary SHA-256:
`af4fb99e7d755d2d41b3622370ea3471aaddaeb34a1cb67eb57f808739308301`.
Scope: [fleet service triage](../SERVICE_TRIAGE.md), aggregate-only two-field
admission, the shared DTO protocol, catalog integration and compact presentation.

## Architecture and semantic acceptance

- The feature extends the existing Go compiler/executor rather than creating a
  fleet evaluator or querying each service dashboard. Aggregate-only admission
  now accepts one or two string grouping fields, validates every field, and
  retains complete-group and source-budget behavior. The normal triage operation
  uses exactly two HTTP aggregate requests and four SQL statements: count and
  groups for each fixed window. Opening a service is a separate existing read;
  its cost is not mislabeled as a local action or part of the four-statement cap.
- The questions preserve the accepted environment policy, namespace and current
  fixed bounds; the baseline is the adjacent equal-duration interval. Both group
  by service and reported namespace and filter the exact reported Server role.
  Recorded span Error, mean span duration and captured counts do not become
  unique requests, application outcomes, SLOs or a service health score.
- Shared admission checks question identity, required tuple dimensions, duplicate
  tuples, provenance shape and count/status/duration conservation. Changed or
  partial observations do not support the ranked table. Missing and present-empty
  namespaces remain distinct. The 1,000-group bound is now reflected in OpenAPI,
  including conditional `g1` shape and rejecting controls.
- Snapshot union keeps current, previous-only and catalog-only services visible.
  Zero captured Server spans is distinct from unavailable Error fraction/latency.
  Unqualified declarations retain unavailable summaries instead of aliasing a
  reported missing namespace or claiming a fabricated zero. This declaration
  case was inspected in source and focused tests, not the ordinary live fixture.
- Source execution, cancellation and navigation remain in application adapters.
  Scope/read revision and creating-draw leases prevent late adoption and prevent
  reattached retired buttons from acquiring the newer scope. Local sorting,
  search, mode switching and held return do not acquire telemetry. The shared
  catalog-search predicate preserves service/signal/declaration search semantics.

## Corrections required during independent review

These were addressed before this acceptance:

1. The UI offered a name ranking that the initial model did not support.
2. Signal counts and Cancel replaced focused native controls without restoring a
   visible replacement. They now preserve focus, while completion does not steal
   focus if the operator moved elsewhere.
3. A DOM containment check alone allowed an old callback to inherit authority if
   its old control was reattached inside a replacement host. A creating-draw
   lease and the stronger reattachment negative control close that path.
4. The catalog's shared search originally included signals/declarations while
   summary search matched only service/namespace; the two now share one predicate.
5. Wildcard declaration identity and summary-origin text needed to remain
   distinct from observed missing namespaces and previous-only observations.
6. The first frames placed an ambiguous “8 shown” catalog fact beside nine
   summary rows and repeated view controls. The retained discovery fact is now
   explicitly labeled Catalog, the summary caption owns its own shown count,
   and the original mode row is hidden while summary controls are active.
7. Initial mobile cards placed Mean and the action on separate tall rows. The
   final native table uses one qualified identity, three aligned statistics and
   one action row, preserving readable labels at 390 and 320 pixels.

## Actual evidence inspected

The coordinator's real OTLP/ClickHouse journey passed in **49.4 seconds**. The
reviewer inspected `artifacts/service-triage/report.json`, its independent raw-row
oracle and the actual [1440px](../journeys/screenshots/service-triage-1440.png),
[390px](../journeys/screenshots/service-triage-390.png) and
[320px](../journeys/screenshots/service-triage-320.png) frames.

The initial Measure records HTTP POSTs `1 → 3` and SQL statements `6 → 10`.
The two accepted populations contain 13 current and 10 previous Server spans;
the oracle checks exact qualified counts, reported Error numerators and duration
sums. The table has nine union identities, including four separately qualified
checkout identities, a previous-only service and Client/log/metric-only cases.
Client-role and foreign effective-environment controls do not contaminate the
Server population. The named namespace “foreign” is correctly included under
All namespaces; it is not an environment-scope exception.

Sort, search and cached mode switching retain identical request/SQL counters.
After the explicitly read service view settles, Back records HTTP POSTs
`11 → 11` and SQL statements `26 → 26`, with opener/ranking restored. An upstream
failure yields zero rankable rows; the held retired read cannot adopt or restart
through its retained callback. The report records no application JavaScript
errors. Its optional Bombadil campaign records 837 snapshots and a completed
triage journey. This is an exploration/check campaign, not an exhaustive proof.

## Limits

This acceptance covers the bounded Server-span lens and the tested Chromium
paths. It does not certify full-fleet membership, telemetry completeness,
statistical anomaly detection, ingestion health, physical screen-reader behavior
or cross-engine accessibility. High-cardinality refusal, large-integer arithmetic
and artificial reattachment are covered by focused/property/model checks rather
than claiming that those situations were all generated in the live fixture.
The verification package separately distinguishes its synthetic protocol model
from replay of actual browser read budgets. The reviewer inspected the final
serial `/tmp/goo-triage-properties.log`: the strict Quint protocol gate passed
(15.78s), the observed ledger replay against this actual browser report passed
(2.28s), and the 64-case Hegel test of the real JavaScript model against independent
Go arithmetic passed (7.46s). All three tests passed with no skips (25.593s total),
including their rejecting controls. These bounded checks strengthen acceptance;
they do not certify a complete source population or exhaustive state-space coverage.
