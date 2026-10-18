# Duration breakdowns and investigation polish

Status: implemented locally. An Explore query with `summary: span-duration` and
one or two `groupBy` fields now returns full-population duration measures for each
group, using the same aggregate expression as the total/time summary. This extends
the existing AST; packages and general captures use the same definition.

## Measurement contract

Each nonempty group includes `duration` with exact count, UInt128 nanosecond sum,
min/max and upper p50/p95/p99. The exact mean is sum/count. Missing map keys remain
null dimensions; present empty strings remain separate values. Each group's count
must match its summary count. Malformed summaries, duplicate dimension tuples and
more than 1,000 groups fail closed. This is not a top-N approximation.

Group counts/sums and extrema are checked against the separate total aggregate
read. A disagreement preserves `changed_between_reads`; it does not become no-data,
a health conclusion, or a silently normalized distribution. Agreement still does
not establish a transactional source snapshot. No aggregate percentile is computed
by averaging group percentiles. The existing exact-upper percentile rule, units,
source limits and sampling/clock caveats in [core basics](CORE_WORKBENCH.md) apply.

The ClickHouse adapter normalizes physical aggregate aliases into the public
`group.duration` shape. The UI consumes semantic results, without map-key or SQL
binding knowledge. Old count-only groups remain readable.

## User workflow

1. In Explore, choose **Span duration summary**, a breakdown, optional time width,
   then **Apply question**. The controls are grouped together. Environment edits
   invalidate evidence immediately, before another click can shift the layout.
2. Open **Distribution**. **Show breakdown by** selects captured count, exact mean,
   p95, p99 or maximum span duration. Count bars retain population percentages;
   duration bars scale to the largest captured group value. Sorting/ranking uses
   integer cross multiplication, including rational means beyond JavaScript's
   safe integer range. Display milliseconds are rounded to three decimals.
3. Local value search hides rows without changing the captured denominator or
   duration scale. Selecting a group refines the query explicitly (missing stays
   missing, empty stays empty). Existing AND/OR drilldown constraints still apply.
4. Save a general capture and reopen/page its stored records. The complete group
   measures and source basis remain fixed. The Records table also exposes group
   mean/p95/max; narrow screens scroll that table and use two-column summary cards.
5. Create a reusable telemetry package from the question. Validation, preview and
   future fresh-window execution retain its grouping/summary; previews expose
   total and group measures. Configuration does not archive preview evidence.

**Pin as baseline** retains complete supported measures. [Error and latency
comparisons](ERROR_COMPARISONS.md) add descriptive mean/p95/p99 changes alongside
record shares; statistical percentile significance is not inferred.

## Capture/compiler compatibility

New captures retain normalized group measures. Older captures may have a global
duration summary but count-only groups; those show counts and an explicit
“not recorded” explanation rather than fabricated duration measures. Existing
immutable receipts are not upgraded in place.

Adding grouped aggregates changes the source plan for summary+group questions.
The existing fixed-plan refresh guard may reject refreshing an older such capture
with 409. Reopening remains source-free. Open it as a live question and create a new
capture to compute the current plan; this is explicit new evidence, not a rewrite
of its history. Count-only plans retain their existing source shape.

## Executable evidence

[Journey 34](journeys/README.md#34-duration-breakdowns-and-saved-measures) ingests
ordinary OTLP spans: 100 fast, 20 slow, two with missing route and two with an empty
route. It checks full-population p95 versus a 100-row page, two-field grouping,
exact zero/missing distinctions, ranking, stable search scale, explicit refinement,
cohort measure accounting, package preview, immutable pages and narrow screens.
Quint validates measure/save/page observations and rejects three corrupted ledgers.
Hegel checks exact group count/sum conservation across generated populations,
including uint64 extremes; native source tests cover malformed/duplicate groups
and same-count duration disagreement. Bombadil explores saved ranking/search/pages
without reading live telemetry, saving resources or generating application work.

Next: arbitrary typed numeric measures, grouped time-series comparison and native
metric exploration. Production identity and actual ingestion-health evidence remain
separate gaps; these span measurements do not establish request or trace latency.

[Error and latency comparisons](ERROR_COMPARISONS.md) extend baselines to complete duration/status evidence and allow different fixed windows. Count-only evidence still uses population-share comparison.
