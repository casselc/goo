# Supplied trace-page comparison

`exploration.CompareTracePages(baselineRows, selectionRows, baselineTruncated, selectionTruncated)` is the normative Go projection. `GooTracePageComparison.compare(...)` mirrors it for supplied-data display. Both accept at most 1,000 observations per page. They perform no source reads, HTTP requests, persistence or capture mutation. The shared fixtures in `tests/fixtures/trace-page-comparison.json` check the complete Go/JavaScript result.

The version is `goo/trace-page-comparison/v1`. Each side includes the existing `goo/trace-page-breakdown/v1` analysis, its timing extent, and reported trace classification. All durations and signed differences are exact integer nanosecond strings. Missing numeric properties and explicit nulls mean unavailable, while `"0"` means known zero.

## What aligns

Groups use exact nullable `(service, operation)` tuples. Missing/null labels, present empty labels and literal display-placeholder text remain distinct. Ordering puts missing labels first, then uses UTF-8 byte order for service and operation. Groups present on only one supplied page remain one-sided; absence on a page says nothing about the full trace or another source window.

Group sums include supplied observations without deduplicating IDs. Inclusive and outside-observed-children differences are computed only when the corresponding group metric is known for every row on both sides. An absent group never contributes an invented zero duration. Unknown group members withhold that metric's difference, while another fully known metric can remain comparable. Summed spans can overlap and are not wall-clock trace latency, CPU time, self time or a critical path.

A row-level `unique-label-match` is only a descriptive alignment candidate. It requires all of the following:

- Exactly one row in the service/operation group on each page.
- Both service and operation are reported strings. Present empty strings remain distinct reported values; missing labels withhold the candidate.
- Valid, unique reported `(TraceId, SpanId)` identities within each page, using exact lowercase OTel hexadecimal IDs without case normalization.
- No malformed-row diagnostic from the underlying page model.
- Both rows report nonempty string `SpanKind` values that are exactly equal. There is no allowlist or inferred role; equal custom reported strings are accepted. Missing, empty, wrong-type or differing kinds withhold the candidate.
- Each page contains exactly one distinct valid reported trace ID. A mixed or unreported trace page withholds row alignment.

A unique label match does not establish the same invocation, causal equivalence, equivalent instrumentation or corresponding parent topology. IDs do not need to match across the two pages. Repeated label groups are `ambiguous`: observations are never zipped by time, sorted position or ID, even when the two pages contain equal numbers of rows. Their complete known group sums can still be compared.

`matchState` is `unique-label-match`, `ambiguous`, `baseline-only`, `selection-only` or `unavailable`. Withheld candidates may include `matchReason`: `repeated-label`, `unreported-label`, `mixed-or-unreported-trace`, `invalid-or-duplicate-identity`, `malformed-row` or `span-kind-mismatch-or-unreported`.

## Timing and partial observations

Each side's `startNs` and `endNs` bound its valid supplied intervals, with `extentNs = endNs - startNs`. Timestamp and duration must be canonical uint64 strings, and their sum must not overflow uint64. Zero-duration intervals are valid. Timing bounds do not require valid span identity or labels: they describe reported interval geometry only. `traceState: single-trace` means one distinct valid reported trace ID was found; it does not assert every observation has valid identity. Invalid observations remain visible in the underlying breakdown.

A matched row's relative start is measured from its own side's earliest valid supplied interval. `relativeStartDifferenceNs` is selection-relative start minus baseline-relative start. The two pages have independent origins; equal horizontal positions do not prove concurrent execution, common clocks or aligned epochs. A valid reported duration may still yield an inclusive difference when a malformed timestamp makes relative placement unavailable. Direct-child union and outside-child measures retain the conservative rules of the underlying page analysis: overlapping children count once; outliving children, ambiguous edges and unusable intervals withhold the affected union rather than clipping or repairing it.

`truncated` remains an explicit per-side input. Application owners should set it for a nonzero page offset or `hasMore`, and keep the original trace selector, environment, fixed window, basis, read time and page position alongside the display. An untruncated supplied page is not a completeness guarantee. Independent aggregate queries must not be presented as sums derived from these row pages.

## Custody and integration

The model adds no API, SQL, query flag, result-bank type or stored capture fields. Reopening a capture or returning to a retained application visit can project the same supplied observations without rereading telemetry. If an owner explicitly fetches another page or trace, it must retain that read's own provenance and replace the corresponding side deliberately. No model computation silently refreshes either page.

Shared fixtures cover exact differences above JavaScript's safe integer range, independent relative origins, partial pages, unknown residuals, missing groups, ambiguous repeats, duplicate identities, mixed traces, malformed labels/timestamps, nullable versus empty labels, and SpanKind eligibility. JavaScript also rejects accessor-backed inputs without invoking getters and rejects whitespace in canonical nanosecond values. The model makes no anomaly, health, statistical significance or causality claim.


## Application flow and boundaries

In the trace inspector, **Pin trace page as reference** freezes the currently
accepted supplied page in that record-results owner's memory. Close the
inspector, explicitly expand another trace, then **Compare with reference**.
Each side retains its own environment, fixed UTC window, trace selector, source
read time, basis and page position. Unique-label candidates share an elapsed-time
scale; the pages' timestamp origins remain independent. Repeated label groups
show aggregate values without invented individual pairing.

**Inspect pair attributes** displays only supplied resource/span differences,
using qualified namespaces and the shared typed field-value component. Object
key order does not create a difference; arrays and value types remain distinct.
Missing keys in a supplied map differ from an unavailable map. This action does
not select the native waterfall or trigger its linked-log read.

Pin/Compare/attribute inspection and Clear are local. Closing the inspector keeps
the held reference; disposing its application owner discards it. New or paged
trace reads temporarily revoke current-page action authority, and obsolete
responses cannot restore it. Hidden, closed or detached controls are inert. A
failed read can still discard the reference through its current visible Clear
control. No new API, persistent paired-comparison capture or automatic typical
trace selection is introduced. Existing saved captures keep their original
bytes; expanding their trace context remains a fresh read.

Actual journey, desktop/mobile frames and final acceptance receipts are linked
from [the journey index](journeys/README.md) and [validation](../VALIDATION.md).
