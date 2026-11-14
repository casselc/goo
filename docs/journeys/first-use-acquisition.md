# From an unknown metric source to held observations

An operator has ordinary OTLP gauge data and no manifest or instance identity.
They want to find the relevant service, read its observations, inspect a subset,
and return to that subset without another source read.

## Executable journey

`journeys/first-use-acquisition.spec.js` sends **240 actual OTLP points** with
separate partition attributes to the collector. It waits for the exported rows,
then copies this run's gauge rows into an isolated configured source. Other
metric-family tables are absent. The generic runtime has no demo definitions.
A transparent source meter forwards actual ClickHouse bytes unchanged.

1. Open **Find metric scopes** over one fixed event-time window. No source query
   runs until **Read metric scopes**.
2. Read reported scopes, then choose the service. Choosing stages locally and
   preserves that fixed window.
3. Choose **Discover** to read instrument metadata. Missing instance identity
   continues to prevent evaluated measurement construction.
4. Choose **Read observations**. The response contains all 240 occurrences while
   the shared table renders a local page of 100.
5. Page locally, Find `beta` across the whole held bank, and press Enter on
   **Inspect**. The subset contains 120 held observations, not just the first
   page's matches.
6. Switch to Captured records and return to Measurements. The exact source-read
   clock, fixed question, Find text, page, occurrence keys and selected detail
   remain. No telemetry request or SQL runs on return.
7. Reread explicitly. The same fixed question is used; fresh source-read time is
   a new receipt. Cancel a subsequent delayed read and retain the previous bank.
8. Start another read, leave and return while its completed real HTTP 200 is
   held at the browser boundary. Deliver those exact bytes afterward. The retired
   completion changes neither held observations nor focus.

## Measured acquisition

This run uses one gauge-only source and the same fixed question throughout.
Browser request counts below mean initiated POSTs, not delayed completions.

| Interaction | POSTs | SQL | Source response bytes | API response bytes | Rendered rows |
| --- | ---: | ---: | ---: | ---: | ---: |
| Read metric scopes | 1 | 6 | 4,136 | 1,644 | 0 |
| Discover instruments | 1 | 6 | 105,929 | 13,924 | 0 |
| Read observations | 1 | 3 | 102,051 | 229,783 | 100 of 240 held |
| Local page, Find, Inspect (each) | 0 | 0 | 0 | 0 | 100 |
| Retained return | 0 | 0 | 0 | 0 | 100 |
| Explicit reread | 1 | 3 | 102,051 | 229,783 | 100 of 240 held |

Cold acquisition totals **3 source POSTs, 15 SQL requests and 212,116 source
response bytes**. The three source families absent from this isolated database
are probed and refused during discovery; this is not a four-family benchmark.
The raw result deliberately carries both raw and typed inspection facts, so its
API representation is larger than the physical projection response. Returning to
the settled bank requires neither the explicit reread's three SQL requests nor
its 102,051 source response bytes. This is a logical cost comparison within one
run, not a before/after latency benchmark.

Cancellation separately initiates a raw POST and the existing query-cancel POST.
The delayed source boundary records downstream closure; the fixture then permits
the upstream request to finish. Its source bytes are still charged to that read,
so cancellation is not claimed to eliminate already-started source work. The
late-success adversary similarly charges source work at initiation and proves
that later completion acquires no new authority.

Final browser/native acceptance is recorded in [validation](../../VALIDATION.md).
The cleaned-up report is `artifacts/first-use439/13b93837eef9/report.json`.

## What the meter means

Counts distinguish browser request initiation from completion: delivery of an
older response is not another source request. SQL entries record status, elapsed
time, response bytes and returned rows. Schema descriptions and zero-row probes
are separate from observation rows; excluded-family errors are retained as
source responses. These are response bytes, **not physical scanned bytes**, and
local elapsed times include browser/test scheduling. This is an acquisition
accounting experiment, not a production latency benchmark.

Raw value reads retain their own schema probe. Scope discovery and instrument
metadata have different questions and result profiles; their probe results do
not establish current value readiness. No cross-request readiness cache is added.
The bounded consolidation is retaining an already settled bank within its owner.
Scope edits, new discovery and Close still discard it; no durable capture or
cross-session cache is implied.

## Runtime frames

- [Desktop held observations](screenshots/439-acquisition-1440.png)
- [390 px held observations](screenshots/439-acquisition-390.png)
- [320 px held observations](screenshots/439-acquisition-320.png)
