# Source refusal and explicit recovery

Status: implemented and verified against the reviewed Go server. The executable journey is
[`344-source-refusal.spec.js`](../../journeys/344-source-refusal.spec.js).
The screenshots below come from genuine OTLP-backed source reads and deliberately planted upstream failures.

A user opens a linked, fixed log population and customizes it into a service
comparison with HTTP-method facets. The question uses an effective typed named
environment, an exact service and tenant predicate, and two fixed UTC windows.
Nine selected logs and five previous-window logs enter through ordinary OTLP;
successful observations and retries read those records from ClickHouse.

The test's disposable Go runtime connects through a scoped local HTTP proxy.
Healthy SQL and parameters forward unchanged to ClickHouse. Explicitly planted
upstream failures exercise the real Go acquisition and admission boundary; the
browser never receives fabricated successful telemetry or a mocked API refusal.
The proxy records which source request was affected.

1. The previous-window aggregate receives 1,001 planted group rows, exceeding
   the application's complete grouping limit. Selected records remain known;
   the baseline is unavailable and there is no persuasive comparison or invented
   zero. The concise reason and collapsed Technical details disclose the refusal.
   Opening details performs no source read.
2. Read fixed baseline explicitly retries that same frozen question. One source
   API call recovers the real five-record baseline; selected observations remain
   unchanged. Independent read clocks retain their existing meaning.
3. Reading HTTP-method values independently fails the baseline with an upstream
   503. Selected counts remain available; baseline cells show unknown. Desktop
   and narrow frames explain the partial state. Retry baseline values reads only
   that window, retains selected counts and receipts, and recovers real values.
4. A held baseline request is retired by hiding and restoring the facet. Its late
   refusal cannot replace the restored owner or leave Loading stuck. Local
   controls do not reacquire data; explicit Read values rereads both windows,
   matching the existing contract. A retained old retry callback stays inert.
5. Filtering both windows stages a child without reading. Back restores exact
   held facet/comparison observations, performs no source IO, and returns focus
   to a connected, visible, unobstructed control.

Timeout, malformed-response and other taxonomy variants belong to the bounded Go
semantic tests rather than a browser fault matrix. Cancellation is owner
retirement, not a new source refusal. The lab's telemetry, stored views and pending
stream entries remain untouched.

## Results

The native journey passed in **17.6 seconds** (19.5-second Playwright runner),
with **14 source API calls**, one pure comparison request and no browser errors.
The source proxy recorded 52 upstream requests, including ingestion-readiness
checks, ordinary multi-query acquisitions and the three explicit fault plants.
The budget plant returned HTTP 200 with 1,001 group rows and was rejected by the
real local admission limit; the unavailable plant returned upstream 503. The held
request's connection closed on cancellation before any late response was accepted.

| Gesture | Source API calls |
| --- | ---: |
| Linked population and customization | 4 |
| Selected read plus refused aggregate baseline | 2 |
| Explicit Read fixed baseline | 1 |
| Partial counted-value read | 2 |
| Explicit Retry baseline values | 1 |
| Held counted-value reads, then owner retirement | 2 |
| Explicit Read values after return | 2 |
| Technical details, retired retry, staging and Back | 0 |

The healthy selected chart retains the exact full snapshot across baseline-only
recovery. Its revision and scope generation match the actual source response
basis; its capture timestamp matches the actual response clock. The facet retry
retains selected values and their receipt while acquiring a new baseline receipt;
the held summary comparison keeps its separate receipts. Failed-side values
remain unknown, with no unsupported count difference. Back retains the legitimate
staged POST choice and exact held observations; focus returns to **Filter both
windows**, connected, inside the viewport and unobstructed at its center.

Three preliminary runs corrected test expectations, with no production changes:
the authored query explicitly adds its canonical empty legacy environment and
one-second bucket; chart source authority uses revision/scope generation rather
than the comparison-only sourceBasis field; the parent snapshot is captured after
the user's POST selection, because that legitimate staged intent survives Back.

- [Aggregate refusal and keyboard-open details](screenshots/344-budget-refusal.png)
- [Partial values at desktop width](screenshots/344-partial-values-1280.png)
- [Partial values at narrow width](screenshots/344-partial-values-390.png)
- [Recovered independent values](screenshots/344-recovered-values.png)

Compact source questions, fault provenance, counts, receipts and runtime identity
are retained under `artifacts/source-refusal344/`. These are local verification
artifacts rather than product captures. The existing 338 paired-facet regression also passed in **2.5 minutes**, including
its independently observed Quint ledger check (**17.99 seconds**, no skips).
Its Bombadil campaign completed in 35.135 seconds: 581 snapshots, phases 0–7,
source 6, lookups 2, comparisons 2, writes 0, verified 4, immutable true and no
overflow. These establish logical campaign results.

The independent driver-image audit found **581 identical JPEGs**, all showing
initial “Loading saved exploration view…”, although DOM observations reached the
settled applied child. Those stale driver captures do **not** validate terminal
appearance. One labelled sample and the audit remain; current maintained
Playwright frames establish appearance. Raw campaign/report bulk was pruned after
review, preserving model input/log, compact run/campaign summaries and replay
inputs. Genuine lab telemetry, views and all 51 pre-existing pending stream entries
were preserved.
