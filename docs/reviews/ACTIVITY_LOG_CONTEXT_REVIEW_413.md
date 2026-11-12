# Independent Activity log-context review — 413

Date: 2026-11-12. Disposition: **accepted in the exercised scope**.
The reviewer did not author the implementation or execute concurrent browser,
Go, property or model campaigns. Source inspection and inspection of actual
runtime reports/frames were independent; the coordinator kept heavy execution
serial.

Reviewed candidate Go binary SHA-256:
`cb8d15240fd6f42f94ff0b56b0413bea6beafc7e23f2673f167aad4b08393728`.
The executable is [activity-log-context.spec.js](../../journeys/activity-log-context.spec.js),
with its [storyboard](../journeys/activity-log-context.md).

## Architecture and measurement acceptance

The optional log context uses the existing aggregate-only executor with one exact
reported severity-text group and six 10-second whole-population buckets in a
completed 60-second window. It adds no evaluator or query per severity value.
The shared DTO protocol checks question identity, independent read receipts,
bounded groups and conservation. Changed or malformed reads are non-actionable.

I inspected the final `artifacts/activity-log-context/report.json` and independently
recounted its raw records using the fixed time bounds, effective environment
policy and original OR predicate. There are **624 matching records**, against
**100 retained stream rows**. Exact severity-text counts are 156 each for empty
text, `INFO`, `ERROR` and `notice.custom`. The six bucket counts are
24, 24, 0, 24, 0 and 552. The raw fixture includes numeric severity 21 with text
`INFO`; text was not reclassified or replaced with a numeric interpretation.

The timeline represents total volume. Its severity table describes the whole
window, not severity within each time bucket. Literal labels are quoted and
pre-wrapped, preserving case and whitespace; empty text is not a missing numeric
severity. Captured-record rates use integer/rational arithmetic before display
rounding. These are not unsampled traffic or service-health measurements.

## Ownership and interaction acceptance

The final report records 17 facts and no page errors. Show and three completed
Refresh gestures each add **one API request and three SQL statements**. Severity
inspection, cadence selection, real pointer/keyboard selection, explanation
open/close/reopen and settled return add no API POST or SQL. The actual selected
question keeps the typed environment, original OR predicate and exact half-open
subwindow of the chart's admitted parent, even after the stream window advances.

The report separately exercises an in-flight Refresh frozen during selection:
one attempt begins, its late response is not adopted, and the older parent and
selected bounds remain. An upstream failure yields no actionable result; a
second refresh cannot overlap pending work. Retired callbacks add no requests
or SQL and cannot adopt after scope replacement. A dirty draft withdraws data
and source actions while leaving the local Hide volume control usable.

Acquisition, cadence, cancellation and navigation remain in application adapters.
Time bars receive supplied counts and emit intent. The small pure read gate is
shared by log and span controllers without importing one signal's controller
into another. The accompanying real span smoke report records two observed
six-span latency results and no errors, exercising the existing consumer after
that extraction.

The source-free anchored explanation validates the same selector/environment,
requires the selected interval inside its parent and normalizes benign Go wire
omissions. It preserves the admitted parent rather than substituting the latest
stream window. Opening focuses its heading; explicit Close returns to the visible
volume opener. Ordinary explanation entry retains its current-window contract.

## Corrections required during review

- A draft assignment cleared the shared chart's pointer-range handler. It was
  removed; capture listeners pause work without replacing the component gesture.
- Retained native bars could be reattached after data replacement. Shared internal
  render revisions now guard both native selection and range emission; the real
  journey tests an old bar and a whole old figure before the current figure.
- A single-bar choice after a range now clears old range highlighting before
  updating controlled `selectedKey`/`aria-pressed`, without replacing gesture DOM.
- Literal text preserves whitespace, and anchored question comparison normalizes
  omitted defaults without relaxing actual selector/window differences.
- Anchored explanation has heading focus and close-return behavior.
- Initial mobile context was too tall. Short controls, hidden disabled manual
  Pause, a 48px chart and concise facts remove duplicate window/status prose.
  Exact dates remain in bar labels and read details; visible axes are UTC times.

## Actual visual and exploration evidence

I inspected the final 1440, 390 and 320px documentation frames. Desktop remains
readable alongside the stream. At 390px, the optional context and a whole first
record—timestamp, service, body and actions—fit, with the next record entering the
viewport. The contained 320px frame reaches the first timestamp/service; existing
scope-toolbar wrapping remains a small-screen constraint, not a reason to add
another layout system. Documentation scrolls to a whole record; ordinary
chart-induced pauses do not claim the explicit Pause-button alignment contract.

The successful two-journey campaign includes the six-span regression and a
30-second Bombadil run with **493 snapshots and completion observed**. This is a
bounded exploration/check campaign, not exhaustive proof. I also inspected the final serial
`/tmp/goo-log-properties.log`: strict Quint protocol passed (21.84s), actual
observed-ledger replay passed (2.22s), and the 64-case projection property plus
64 × 60 commands against the actual shared read gate passed (6.36s). All three
tests passed without skips in 30.435s, including rejecting controls. These model
and property checks remain distinct from actual browser/frame inspection.

## Limits

This acceptance covers the log-context slice and tested Chromium paths. It does
not certify telemetry completeness, ingestion health, severity-number/text
reconciliation, statistical anomaly detection, a transactional snapshot,
physical screen-reader behavior or cross-engine accessibility. High-cardinality,
large-integer and synthetic lifetime cases have focused/property/model coverage;
they were not all reproduced as real telemetry in the browser fixture.
