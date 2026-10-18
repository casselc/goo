# Error and latency comparisons

Explore can summarize recorded span status, rank error percentages by service,
span name or any available attribute, and compare independently captured
populations. The existing duration summaries provide mean, p95 and p99 changes.
This works with ordinary OTel spans; goo manifests and application annotations
are optional.

## Status contract

`spanStatus: true` is an opt-in spans-only question/package setting. The
ClickHouse OTel adapter counts exact exporter values `Error`, `Ok` and `Unset`;
empty/unrecognized values form a fourth `unknown` category. The semantic result
uses decimal strings: `count`, `error`, `ok`, `unset`, `unknown`. The four categories
must account for the entire matching population. Unset is not explicit success;
unrecognized data is never coerced to OK. OTLP's absent status normally becomes
exporter `Unset`; the adapter cannot recover whether the producer omitted it.

Recorded error percentage is **Error spans / all matching captured spans**.
This describes captured telemetry, not unique requests, HTTP response-code errors,
business outcomes, completeness or a failure SLI. Choose server/client/internal
span selectors deliberately to avoid interpreting every span as a request.
Zero captured spans have no error percentage. Status classifications are exposed
as adapter field `column:status_class` (`error`, `ok`, `unset`, `unknown`) for
filtering and grouping. General custom outcome policies remain future work.

Status counts are computed in the same read as totals, duration summaries and
each group/time interval, independent of inspection-page limits. Missing or
invalid category counts fail the read. Separate total/group/interval reads with
different category totals yield `changed_between_reads`, even when overall record
counts agree. Their comparison is unknown. These checks do not create a source
transaction or establish completeness.

## Comparison workflow

1. Apply an explicit span question with duration summary, span status/errors and
   one breakdown. Select **Recorded error percentage** in Distribution to rank
   the complete groups. Search only hides rows; it preserves the captured scale.
2. **Pin as baseline** retains a browser-held population while filters change
   within its fixed window. **Compare previous window** reads an adjacent earlier
   window of the same width with the current selectors, grouping and measures.
   The current inspection window stays fixed while investigating changes.
3. Compare population shares, mean, p95, p99 or recorded error percentage. Whole
   population and group differences are current minus baseline. Ranking uses exact
   rational/integer absolute differences; rounding is only for display. Each
   percentile comes from its own population; group percentiles are never averaged.
   API mean differences are rounded to 0.001 ns and error changes to 0.01 pp;
   exact sum/count evidence remains available for rational calculation and ranking.
4. **Inspect current** adds the changed group's predicate and opens Records.
   The baseline is retained. Follow a record's reported trace into a bounded
   waterfall and exact linked logs. Trace expansion reads across services in the
   selection's original environment/window and identifies its separate evidence.
5. **Capture comparison** rereads both fixed questions on the server. The dialog
   explains that this captures new evidence and may differ from the live browser
   baselines. It stores two normal bounded record captures and an immutable pair
   receipt. Open Records on the pair to open either underlying capture; these
   retain ordinary saved paging and explicit live trace/log expansion.

Both cohorts must use the same signal, environment and one grouping field.
Different filters and fixed windows are allowed; equal window lengths are not
required by the API. Counts are not rates per second. Source bases on the live
`/api/compare` input are caller-supplied metadata, not authenticated provenance.
The API verifies canonical counts, unique groups, category accounting and supported
duration summaries. Saved pairs are derived from server source reads.
Missing groups are absent populations: their measures and differences are
unavailable, never substituted with zero. Old count-only captures remain count
comparisons. No-data or source disagreement suppress normalized differences.
Comparisons are descriptive leads, with no causal or statistical significance claim.

## Persistence and compatibility

The existing capture API accepts `kind: comparison` with `comparison.baseline`
and `comparison.selection` query objects. Creation publishes two child Explore
captures and then the pair. If later work fails, already published children may
remain in the catalog; no incomplete pair is published. Publication is local
immutable file storage, not a transactional multi-object database.
The pair includes both aggregate cohorts, their comparison and child capture IDs.
Reopening verifies stored integrity and recomputes the comparison locally, without
telemetry access. Refresh retains both fixed questions and checks each child's
compiled plan before creating successors. Failure retains the previous receipt.
Opening a pair as a single live question is disabled; open the appropriate child.
The selected display measure and browser layout are not stored in the receipt.

Existing count-only SQL and receipts retain their shapes when `spanStatus` is
absent. Status-aware YAML packages retain the setting through drafting, preview,
publication and fresh-window execution. This adds no general numeric expression
language or native metric error-ratio policy.

## Executable storyboard and validation

[Journey 35](journeys/README.md#error-and-latency-comparison) ingests 80 prior and
124 current spans through real OTLP. Only 100 current rows are displayed while
all 124 contribute to status and duration measures. It verifies 2.50% → 16.13%
recorded errors (+13.63 percentage points), +90 ms p95, missing/empty routes,
saved pair reopening, route-to-trace-to-log navigation and mobile containment.

`verification/error_comparison_test.go` adds Hegel state/swarm checks for exact
changes, window moves, scope changes and disagreement. The Quint ledger adapter
checks actual measure/compare/save/reopen observations and rejecting controls for
wrong denominators, wrong p95 changes, window drift and modified saved evidence.
The opt-in Bombadil actor changes saved comparison measures while checking
accounting, display and immutable evidence. These are bounded observed checks,
not a proof of telemetry completeness or universal browser scheduling.

Saved comparison pairs can now be added to [ordered investigation notes](WORKBENCH_CONTEXT.md). The narrative cites the exact pair ID and retains its own revision history. Authored deployment receipts provide separate context rather than changing either comparison population.
