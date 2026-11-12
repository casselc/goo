# 411 · Read a histogram without stacked chrome

The operator has already scoped a population or selected an instrument export.
The next task is to read its distribution, inspect an exact bucket, and continue
without another source read.

| Scene | Operator action | Expected result |
| --- | --- | --- |
| Numeric population | Run the scoped span question, open Distribution, customize, run the population view. | The original numeric binding, coverage and exact signed 64-bit bin counts remain supplied to the shared chart. One chart surface retains its meaningful title. |
| Compact viewport | Resize to 390 px, focus the named bucket region, press Right. | The buckets scroll locally; the page does not overflow. Unit, scope and observation coverage remain visible. |
| Exact numeric values | Focus Exact values and press Enter, then close it. | Every supplied bin remains in the table, including hidden outer empty bins. Focus stays on the disclosure. |
| Saved numeric evidence | Explicitly capture the fixed question, reopen its stored Distribution, highlight a bin with Space/Enter. | The highlight is local, the frozen snapshot is unchanged, and no source-range action or telemetry read is introduced. |
| Native export | Run the native measurement question, select an export interval, open Selected native bucket counts. | Selected interval observations displays the same instrument bucket bounds and observation counts, retaining native histogram semantics. |
| Exact native values | Open Exact values, then close it; use the same keyboard scroll gesture on mobile. | Exact bounds and counts remain available. These gestures issue no API POST, SQL query or save. |

`journeys/411-quantitative-chrome.spec.js` runs both binaries against one persisted
fresh OTel bank. Only the before pass ingests it. The after pass reuses its exact
questions and verifies raw records, native observations and supplied chart values
against the before pass. The report records visible headings, bordered inner
frames, chart-to-plot offset, usable width and page overflow at 1280 and 390 px.
Screenshots and both compact reports are in `artifacts/quantitative411/`:
`{before,after}-{numeric,native}-{1280,390}.png` and `{before,after}.json`.

Run before with `GOO_411_PASS=before GOO_GENERIC_BINARY=/tmp/goo-411-before`,
then after with `GOO_411_PASS=after GOO_GENERIC_BINARY=/tmp/goo-411-reviewed-goo`,
each followed by `npx playwright test journeys/411-quantitative-chrome.spec.js`.
Keep the bank between passes; a new before pass intentionally creates a new bank.

The focused comparison establishes presentation, local keyboard behavior and
source-free highlighting of independently captured evidence.
The existing 407 authored histogram and 371 native reuse journeys separately
exercise source ownership, held controls, saved configuration and stale responses;
411 does not replace those acceptance tests.

## Checked result

The same supplied OTel observations passed both runs on November 12, 2026.
Each chart now has one visible title instead of two; the inner bordered frame
is removed. At 390px, usable numeric plot width increased **216 → 276px**, and
native width **224 → 284px**. Plot offset decreased **191.6 → 121.6px**. Neither
page overflowed; named regions accepted focus and ArrowRight scrolled them.
Exact-value disclosure and viewport changes added **zero SQL queries or API POSTs**.
Read clocks remain independent between runs; supplied values and fixed scope match.

| View | Desktop | 390px |
| --- | --- | --- |
| Numeric population | [Frame](screenshots/411-numeric-1280.png) | [Frame](screenshots/411-numeric-390.png) |
| Native interval | [Frame](screenshots/411-native-1280.png) | [Frame](screenshots/411-native-390.png) |

[Before: numeric at 390px](screenshots/411-before-numeric-390.png) preserves the
same-data density comparison. These frames deliberately scroll to the visual
under review; they are not first-data entry measurements.
