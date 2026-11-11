# Captured-record activity rates

Choose **Captured records/s** in an ordinary fixed summary or ungrouped timeline.
It is an alternate measure; existing defaults and card counts stay unchanged.
Summaries use the full captured-query count. Timelines use each supplied bucket's
count and actual start/end bounds, including a shortened final bucket.

The supplied field is `record.rate`, with unit `records/s` and a distinct rate
aggregation. Its exact rational value is:

```text
captured count × 1000 / elapsed milliseconds
```

The cell retains the exact count and elapsed milliseconds. Validation checks them
against the same row's count and the declared query window or interval bounds.
The fixed question window supplies summary exposure even when a predicate narrows
which records match. Rounded display values keep the exact value and facts in
existing tooltips/disclosures. Rate is not percentage or bounded-fraction metadata.

A known zero count over positive elapsed time is a valid zero rate. An interval
marked unknown stays unknown, with no invented count or rate. Sampling, duplicate
observations, collection coverage and source-read consistency retain their existing
meaning. Captured log/span records per second do not establish request throughput,
unique transactions, health or success.

The initial profile supports ordinary fixed log/span summaries and ungrouped
count timelines. Grouped/category rates, rate comparisons, live-buffer rates,
native metrics and histogram/reported-edge profiles remain outside this arm.
There is no additional source SQL or renderer-side rate calculation.

Selection, field order and measure changes are local presentation. Save/reopen
and compatible reuse keep the destination's complete query and explicit **Run**.
[Ungrouped summary templates](UNGROUPED_SUMMARY_TEMPLATES.md) can retain the chosen
rate field under their existing profile and page-zero gates. A donor contributes
presentation, not its values, exposure, completeness or selection authority.

See [summary authoring](AUTHORED_POPULATION_SUMMARIES.md),
[fixed timelines](SAVED_COUNT_TIMELINE.md), [shared components](UI_COMPONENTS.md)
and [validation](../VALIDATION.md).
