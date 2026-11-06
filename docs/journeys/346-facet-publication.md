# Measure whole facet publication operations

Status: baseline measured against commit `537c879`; the final implementation
and genuine native verification pass. The executable
[journey](../../journeys/346-facet-publication.spec.js) measures the styled running
application over genuine OTLP data. Successful results are not mocked.

One fixed log population has 1,000 category values: missing, present empty and
998 labels of 384 UTF-8 bytes. The previous window has 1,000 disjoint labels.
Grouping by one service keeps unrelated chart size small. The same fixture is
used at 1280px and 390px, first as an ordinary facet, then as a fresh fixed-window
comparison with a 2,000-value union. Effective typed environment policy, tenant
predicate, UTC bounds and independent read receipts remain exact.

The user configures a facet, explicitly reads its values, finds a terminal label,
stages that exact full value, opens a child, and returns with Back. The paired
version repeats the same flow using Filter both windows. No child Run is needed
to exercise the retained-parent owner transition. Staging and Back must perform
zero source IO, preserve Find and receipts, and reject the actual retired action
callback even if its old node is reattached.

## Measurement boundary

A test-only prototype wrapper records each categorical-facet render, connected
state, component identity, cause, output row count and exact state/receipt
signature. It separately records data/interaction setter calls. Render CPU is
summed only across render calls; setter CPU includes nested rendering and is never
added to that sum. Loading and partially ready publications are distinguished
from complete ready builds.

Operation wall time extends through two animation frames after the action
settles. It includes browser-driver scheduling and, for reads, source latency.
This is a paint opportunity, not compositor presentation or a portable latency
SLO. Numbers are bounded integration samples, not a statistical speedup study.
Source and CSS hashes match the immutable reviewed binary; components.css is
verified loaded through the CSSOM. The probe and fixture remain the same for the
final measurement.

## Baseline

The baseline passed in 1.6 minutes (1.7-minute Playwright runner), with 18 source
API requests, two pure comparison calls and no browser errors. Each viewport
uses four setup reads, one ordinary value read, two fixed-comparison reads and
two paired value reads. Local gestures perform no source reads.

Both viewports have the same deterministic build counts:

| Operation | Render calls | Constructed category rows |
| --- | ---: | ---: |
| Ordinary owner setup | 2 | 0 |
| Ordinary source publication | 5 | 1,000 |
| Ordinary local staging | 1 | 1,000 |
| Ordinary child staging | 5 | 1,000 |
| Ordinary retained Back | 5 | 5,000 |
| Paired owner setup, including explicit Run | 14 | 0 |
| Paired source publication | 8 | 3,000 |
| Paired local staging | 1 | 2,000 |
| Paired child staging | 5 | 2,000 |
| Paired retained Back | 5 | 10,000 |

Source publication is not five or eight duplicate complete builds. Ordinary
publication builds the ready 1,000 rows once. Paired publication builds an
independently ready 1,000-row side, then the complete 2,000-row union once. Earlier
calls represent unloaded/loading states.

Back is the concrete repeated-work finding: each return rebuilds the identical
ready rows five times, from reconnect, interaction assignment and three data
assignments. The exact state, receipt clocks, disabled flag and staged intent
are unchanged across those calls. Observed total render CPU on Back was
477.7/1,590.8ms for ordinary/paired desktop, and 490.8/2,059.5ms at narrow width.
These samples motivate removing repeated publication, not suppressing independent
source observations or introducing a query/DOM cache.

Baseline frames:

- [Ordinary desktop return](screenshots/346-baseline-1280-ordinary-returned.png)
- [Paired desktop return](screenshots/346-baseline-1280-pair-returned.png)
- [Ordinary narrow return](screenshots/346-baseline-390-ordinary-returned.png)
- [Paired narrow return](screenshots/346-baseline-390-pair-returned.png)

## Final verification

The final journey passed in 3.2 minutes, with the same 18 source requests,
two comparisons and no browser errors. Atomic publication installs the model and
interaction together; an owner-local exact-input guard skips redundant refreshes.
A retained parent is not removed and reinserted. A fresh owner still retires old
callbacks and publishes the current state synchronously.

Both viewports show the same deterministic final counts:

| Operation | Baseline → final render calls | Baseline → final constructed rows |
| --- | ---: | ---: |
| Ordinary owner setup | 2 → 1 | 0 → 0 |
| Ordinary source publication | 5 → 4 | 1,000 → 1,000 |
| Ordinary local staging | 1 → 1 | 1,000 → 1,000 |
| Ordinary child staging | 5 → 1 | 1,000 → 0 |
| Ordinary retained Back | 5 → 1 | 5,000 → 1,000 |
| Paired owner setup, including Run | 14 → 5 | 0 → 0 |
| Paired source publication | 8 → 7 | 3,000 → 3,000 |
| Paired local staging | 1 → 1 | 2,000 → 2,000 |
| Paired child staging | 5 → 1 | 2,000 → 0 |
| Paired retained Back | 5 → 1 | 10,000 → 2,000 |

Independent partial and complete value publications remain intact. Back builds
one current ready surface, preserving exact receipt fields and Find. All four
returns restore connected, visible, unobstructed opener focus. The retained
Read details disclosure remains closed in this journey; opening-state retention
is not claimed. Current Find focus and End-caret behavior work. Arbitrary caret
retention is not claimed. The final negative check captures an actual current
Include callback immediately before child departure, then reattaches and invokes
it after Back: neither source requests nor local selection/model change.

Observed Back samples illustrate why build counts and timing claims are separate:

| Return | Render CPU ms, baseline → final | Wall through paint opportunity ms, baseline → final |
| --- | ---: | ---: |
| Ordinary desktop | 477.7 → 77.6 | 2,154.7 → 1,742.4 |
| Paired desktop | 1,590.8 → 428.1 | 5,723.0 → 11,157.9 |
| Ordinary narrow | 490.8 → 87.6 | 1,905.4 → 1,761.2 |
| Paired narrow | 2,059.5 → 407.9 | 6,921.8 → 6,452.8 |

Some wall-time samples increase despite fewer ready builds. These measurements
establish eliminated repeated construction, not a general end-to-end latency
improvement. The final run adds semantic assertions outside the measured render
payload; its overall duration is not compared as a speedup.

A small receipt-admission helper now guards future final runs: each of four Back
and staging operations must build once with zero source/comparison IO, and source
publications must retain the independently ready 1,000-row side and 2,000-row
union. Post-run admission accepted the actual final receipt and rejected a cloned
receipt containing a duplicate connected Back build. This check does not rerun
the browser and imposes no CPU threshold.

Final frames:

- [Ordinary desktop return](screenshots/346-final-1280-ordinary-returned.png)
- [Paired desktop return](screenshots/346-final-1280-pair-returned.png)
- [Ordinary narrow return](screenshots/346-final-390-ordinary-returned.png)
- [Paired narrow return](screenshots/346-final-390-pair-returned.png)

The unchanged 338 paired-facet regression passed in 3.8 minutes (3.9-minute
runner), including partial-side failure/retry, pending-owner retirement and stale
controls. Its observed Quint check passed in 19.66 seconds with no skips. The
35.187-second Bombadil campaign produced 544 DOM snapshots through phase 7,
source 6, lookups 2, comparisons 2, writes 0, verified 4, immutable true and no
overflow.

The independent driver-image audit found 544 images with four unique hashes;
535 repeat the terminal image. That image shows an applied child but still reads
“Reading fixed baseline…” with disabled controls, whereas actual final DOM reports
the pair ready. The image is stale at acquisition and is not settled comparison
visual evidence. Current maintained Playwright frames establish appearance; the
labelled terminal sample, audit and compact logical receipts remain. Compact baseline/final inputs and operation receipts are retained
under `artifacts/facet-publication346/`. Genuine lab data, saved views and all
51 pre-existing pending stream entries remain untouched.
