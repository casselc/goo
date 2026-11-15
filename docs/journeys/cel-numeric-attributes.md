# Filter ordinary OTLP data with exact numeric conditions

Executable: `journeys/cel-numeric-attributes.spec.js`.

An operator has logs and spans with exporter string attributes, without Goo
manifests. They need to isolate a numeric population while retaining the current
service/environment/time question.

1. Stage the fixed parent, open **Filters → Expression (CEL)** and check an
   explicitly declared uint64 condition. Checking reads no telemetry.
2. Run to isolate the exact value `9007199254740993`, then a broader valid range.
   Record IDs are checked against independently read exporter records, never
   inferred from the displayed page or JavaScript Number conversion.
3. Try signed negatives and valid-only inequality. Missing, empty, overflow,
   leading-zero and fractional strings are distinct cases, rather than zero.
4. Use explicit missing/empty/invalid methods. Declare decimal64 scale 2 and
   compare `1.20` or the maximum coefficient value `92233720368547758.07`.
5. Change only the attribute namespace. A resource key and a record key with the
   same name deliberately select different records.
6. Check invalid thresholds, computed scale/key and an invalid unit. The editor
   retains editable source and a positioned diagnostic; execution remains
   disabled until a valid check recovers.
7. Switch to Conditions, discard the expression through the existing ownership
   control, and edit the same typed filter. Its uint64 type and exact threshold
   survive; changing representation reads no data.

The fixture emits ten cases and two outside-parent/environment controls per
signal through real OTLP. It independently verifies their physical exporter map
strings before the browser test. There are twelve positive examples per signal,
six refusals per signal and one valid recovery per signal. All explicit source
runs preserve the complete parent and fixed clock. A separate property suite
covers all six comparisons and all four parse states across all three types and
both maps/signals, including min/max coefficient boundaries.

## Current interface

The same shared Conditions editor displays the checked binding and exact
threshold. No extra toolbar, query mode or numeric chart is introduced.

- [Desktop](screenshots/443-numeric-condition-1440.png)
- [390 px](screenshots/443-numeric-condition-390.png)
- [320 px](screenshots/443-numeric-condition-320.png)

The existing `journeys/cel-editor.spec.js` also runs a 25-second Bombadil campaign.
Its valid candidate must now be a typed uint64, decimal64 or invalid-state
condition. It retains invalid source across Conditions/Expression changes and
rejects a hidden stale callback, with zero telemetry reads or forbidden POSTs.
The terminal typed-condition marker prevents empty campaign acceptance.

Executed results and local run receipts are in [validation](../../VALIDATION.md).
See [syntax and semantics](../CEL_NUMERIC_ATTRIBUTES.md) and
[independent review](../reviews/CEL_NUMERIC_443_ACCEPTANCE.md).


## Executed acceptance

The final run passes **1.1m**: 24 source runs, 38 condition plans and 28 local
source-free steps across logs/spans. All twelve deliberate bad expressions return
positioned HTTP400 diagnostics; recovery and visual editing pass. Browser console
refusal messages are accounted for separately; no unexpected browser errors
remain. Owned runtime cleanup succeeds. Receipt:
`artifacts/cel-numeric443/fcb86adaeb84/report.json`.

The contextual editor/Bombadil journey passes **34.9s**. Its campaign records
**513 snapshots**, completes with a typed `int64(...).invalid()` condition, and
makes no source reads or forbidden POSTs. It does not establish that each random
numeric candidate was chosen. Receipt:
`artifacts/cel-editor4364/14503da776/report.json`.

The source lab has a day-based TTL, so a run crossing midnight can lose yesterday's
fixture rows. Within two minutes of midnight this synthetic test places events
just inside the next UTC day, retaining bounded fixed questions and unchanged
source assertions. This is fixture stability, not a source-completeness claim.
