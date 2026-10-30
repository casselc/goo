# Recorded status fractions use their own denominators — 268

[Executable journey](268-status-fractions.spec.js) sends 103 real OTLP spans:
100 frontend spans (one Error, 99 OK), one worker Error, and two spans with explicit
Unset or omitted status. The ClickHouse exporter is independently checked: the
omitted status becomes Unset, so the observed classes are Error 2, OK 99, Unset 2 and
Unknown 0. Unknown-class behavior is tested separately in generated DTOs; the
journey does not pretend to have observed an unknown exporter value.

| Interaction | Result | Meaning |
|---|---|---|
| Choose Recorded error fraction | Frontend 1/100 is 1%; worker 1/1 is 100% | Each row uses its own captured count, not the query total 103 |
| Find frontend, sort, or switch table/bar | The bar remains 1% on a fixed 0–100% axis | Hidden groups never normalize a fraction to the largest visible value |
| Inspect status class counts | Both Unset spans remain visible in their class | Zero recorded Errors is not a healthy/success classification |
| Inspect empty timeline intervals | Count/class counts are 0; fraction is unavailable/no-data | An empty denominator does not become 0% |
| Pin baseline and rerun | Same data revision, new read clock; old callbacks refuse | Per-row facts and current execution ownership are separate |
| Stage frontend, then Run | Staging reads nothing; frontend still 1/100 and 1% | The query total changes 103→100 while the row denominator remains 100 |
| Inspect comparison | Native comparison retains −0.94 percentage points | A difference in percentages is not another bounded fraction |
| Open a status-only Saved capture | Canonical summary and read-only fractions remain available | 100 displayed rows do not replace 103 captured/grouped records |

## Documentation frames

- [Independent row denominator and fixed axis](../docs/journeys/screenshots/268-independent-status-fraction.png)
- [Known empty intervals do not imply 0%](../docs/journeys/screenshots/268-known-empty-status-intervals.png)
- [Status-only Saved population summary with all four classes](../docs/journeys/screenshots/268-status-summary.png)
- [Read-only Saved status fraction](../docs/journeys/screenshots/268-saved-status-fraction.png)

The Hegel suites run 100 generated cases over actual Go source-produced group,
time and summary DTOs. They check class partitions, uint64 extremes, exact reduced
fractions, zero denominators, generic field names, unit 1 without percentage meaning,
malformed references/cells/unsafe numbers, and retained row fractions in held or
projected data. The unknown-interval case is explicitly a held Go DTO derivative
of source data; it tests the consumer contract with null operands and unknown
cells, not a fabricated source observation. Fraction metadata identifies same-row
numerator and denominator fields; no support partition is invented for a ratio.

The observed eight-step Quint model projects source → Find → sort → table → rerun
→ retired callback refusal → Stage → explicit Run. It reads operand/fraction
facts through the public API over the actual chart handle and effects from the
DOM. The browser separately asserts the visible fixed-axis geometry. It does not prove private
source authority, ownership, atomic source reads or complete telemetry collection.
Seven altered traces and an unfinished prefix must fail. Read counts refer to
Explore requests; Saved capture is a separate explicit source read and comparison
uses the independently held operands.

The optional 35-second Bombadil campaign passed with 206 snapshots across all
11 phases and three explicit Explore requests, ending at the unchanged 1/100 row
fraction in the frontend-only scope. Retain compact summaries/model traces and
current frames; successful raw campaign logs,
screenshots and disposable capture stores are not needed after validation.
