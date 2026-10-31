# Record inspection source work — 289

## Measured boundary

`TestRecordPagingSourceWorkMeasurement289` runs the actual Go source acquisition
against a deterministic recording HTTP transport. It records executed SQL requests,
not planner intentions or browser POST counts. No ClickHouse timing, physical scan
volume or production performance claim is made by this experiment.

| Query | Initial full run | Existing full paging | New page profile |
| --- | ---: | ---: | ---: |
| Records without grouping | 3 | 3 | 2 |
| Grouped records | 4 | 4 | 2 |
| Grouped records with timeline | 5 | 5 | 2 |

Full reads acquire count, rows and field discovery, plus grouping and timeline
when configured. Changing only offset does not remove those acquisitions. The
new profile acquires a real plain count and a bounded row page only. The grouping,
duration/percentile, numeric, status and timeline declarations are removed from a
private count-acquisition question; they remain in the original returned question.
Rows use the original compiler output and its own parameters, preserving requested
derived fields and reported-edge/filter/environment semantics.

## Why the rollout is composition-only

`pageSignal` already holds aggregate and timeline results in `entry.result` for a
composition and swaps only `entry.recordResult`. Previously it discarded newly
read aggregates and discovery from the page response. In an ordinary population,
paging replaces the whole result, including aggregates; changing that behavior
would require a separate visible freshness/ownership design. Its path remains
unchanged.

The page still carries its real fresh count, state, basis and read clock. None of
these replaces the held chart denominator or chart basis. Count and rows are
separate, nontransactional reads; late arrival can change a page or produce
`changed_between_reads`. No-data is not health, and upstream coverage stays unknown.

## Delivered contract

`POST /explore/page` accepts the same validated Query and follows ReadAndSnapshot
admission. Its narrow `RecordPageResult` contains only question, capturedAt,
basis, count, state, rows, hasMore, optional resultToken and warnings. It exposes
no plan, aggregate datasets, groups, fields, percentile summary or completeness.

The original question is validated before stripping any acquisition requirements.
Invalid original summaries or other declarations cannot gain admission by being
removed. Exact original row SQL is checked for group-only exception_type,
body_pattern_v1 and severity_number projections. Count has no quantile or grouping
work. Source failures do not become successful empty pages.

Existing server-side ResultSnapshots retain the exact bounded row result, with
its existing TTL and capacity semantics. The capability supports Keep selection
without trusting client rows or manufacturing a population total. The response
projection hides internal Result compatibility fields; Keep reads exact stored
indexes and does not gain aggregate authority. No new token family or durable
record schema is needed.

## Acceptance and remaining limits

Focused Go tests cover executed read counts, full-query preservation, independent
snapshot loading, exact kept-row custody, narrow wire slots, invalid-query rejection,
source failure and original derived projection parity. HTTP tests cover route
admission and valid token publication. Browser/model/property verification is
recorded separately in VALIDATION.md by its execution owner.

This removes demonstrably unused work; it does not establish a latency percentage.
Reported-edge row/count SQL may still each resolve parent ownership, and plain
count is still a population scan. A rows-only profile would need a separate
custody/data-model review rather than inventing totals or silently disabling Keep.
