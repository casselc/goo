# Exact fixed bins share the canonical display contract — 269

[Executable journey](269-scoped-histogram.spec.js) ingests 18 real OTLP numeric
attributes: 13 valid signed integers, one missing, one empty and three invalid.
Values include signed limits, a value above 2^53 and the invalid max+1 sentinel.
The collector/ClickHouse population is independently checked before using the
isolated Go application. No source response or source witness is mocked.

| Interaction | Outcome | Boundary |
|---|---|---|
| Inspect the distribution | 13/18 valid values with explicit missing/empty/invalid counts | Bin counts sum to valid values, not all captured records |
| Select the final signed bin with the keyboard | One record in [2^62, 2^63) | Endpoints are stable bin keys and exact geometry; max+1 is not a valid input value |
| Open Exact values | All 128 bins retain exact bounds and counts | Equal ordinal bar widths represent counts, not numeric density |
| Rerun the same fixed population, then replay old actions | Same content revision/new read clock; no planner or source request from old callbacks | Host execution ownership is separate from immutable contents |
| Select current maximum bin and Run selected population | One source-free plan, then one explicit child read | Go omits the illegal max+1 predicate threshold while retaining the valid-value guard |
| Inspect the child | One visible nonzero tail bin, with all 128 bins still supplied | Outer zero-bin trimming changes presentation, not measurement geometry |
| Open Saved data | The original distribution is read-only, including on mobile | Source capture and runtime proof remain distinct from serialized data |
| Read invalid-only and absent populations | Valid 0/total 3 differs from valid 0/total 0 | Empty bins never imply an empty or healthy captured population |

## Documentation frames

- [Exact maximum-bin selection](../docs/journeys/screenshots/269-exact-sentinel-range.png)
- [Selected numeric population](../docs/journeys/screenshots/269-selected-population.png)
- [Read-only Saved histogram on mobile](../docs/journeys/screenshots/269-saved-fixed-histogram.png)

The 100 Hegel cases use actual Go source-produced histogram DTOs and actual JS
contracts. Existing independent integer/fixed-decimal oracles predict 65 unsigned
or 128 signed bins, reduced human-unit edges, max+1 sentinels, and numeric support.
Cases include int64/uint64/decimal64 boundaries, missing/empty/invalid values,
empty/no-valid populations, malformed geometry/counts/support and exact intents.
Imported data retains the full fixed schema but loses the bins guarantee. Row
projection is refused; removing bins while keeping full-population support is not
a supported operation. Source fixtures use bounded fake physical reads, separately
from the real collector/ClickHouse browser journey.

The small observed Quint model checks source → keyboard selection → exact table
→ rerun → retired callback refusal → current selection → explicit execution. It
reads exact support/bin facts from the actual chart and counts observed Explore
and planner requests. Its `selected` field records the fixed candidate bin's count;
the browser separately asserts the actual gesture and selection tray. Seven altered traces and an unfinished prefix must fail.
The model does not prove private source authority, upstream collection completeness
or every possible range. Additional invalid-only/empty reads and Saved capture are
explicit scenarios outside the model's three-request main flow.

The optional 35-second Bombadil campaign passed with 241 snapshots across all
eight phases, three Explore requests and one source-free planner request. The
real journey also passes the Saved/mobile, invalid-only and absent-population
checks. Retain compact summaries/model traces and current frames; raw successful campaign
logs/screenshots and disposable capture stores are removed after validation.
