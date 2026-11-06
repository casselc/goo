# Recover trace and linked-log context

Status: implemented and verified against the reviewed 345 build. The executable
[journey](../../journeys/345-context-recovery.spec.js) uses genuine OTLP spans and
logs, the Go server, and ClickHouse through the scoped fault proxy from 344.
The screenshots below were captured during the genuine native journey.

A user opens a fixed span question, inspects an operation and asks for matching
traces. The fixture has one reported two-service trace: a current-environment
root and legacy-environment worker. A second span condition finds the worker's
error independently of the root operation. Effective typed environment policy,
fixed bounds, conditions and trace identity must survive every recovery.

1. Trace-wide search encounters a genuine upstream unavailable response. The
   reason and collapsed Technical details explain that acquisition failed rather
   than suggesting zero matches. Keyboard disclosure performs no source IO.
   Run trace search explicitly retries the exact request, with one source call.
2. Open trace (fresh read) fails acquiring the span page. Retry trace context
   retries that failed question and then reads linked logs, matching its stated
   two-read boundary. No linked logs are read after the failed span acquisition.
3. The span retry succeeds, while exact linked logs fail independently. The known
   two-span trace remains visible; logs are unavailable. Find linked logs retries
   only the failed log question, with one source call. It returns the one exact
   linked log, excluding an unlinked neighbor and a foreign-environment decoy.
4. The retired trace retry callback cannot issue another read or replace context.
   Closing the inspector restores visible, unobstructed focus to the current trace
   opener and leaves the original question and rows intact.
5. Closing a genuinely pending trace search retires its request. Its late refusal
   cannot reopen results or publish a failure into the closed owner. Close and
   local details perform no replacement reads.

The proxy forwards all healthy source requests unchanged. Only the deliberately
failed acquisitions are planted; no successful telemetry or UI response is mocked.
No new lifecycle model is proposed because existing owner and cancellation
protocols remain unchanged. The existing typed trace-search regression provides
broader policy/model/actor coverage after the bounded new journey.

## Results

The new journey passed on its first run in **11.2 seconds** (15.9-second
Playwright runner), with **8 source API requests**, 17 total API requests and no
browser errors. The scoped proxy recorded 21 upstream requests, including fixture
readiness and ordinary multi-query reads. Three explicitly planted upstream 503
responses became safe, typed unavailable refusals at the Go boundary. The fourth
held request closed on cancellation before any late response was accepted.

| Gesture | Source API requests |
| --- | ---: |
| Initial fixed span question | 1 |
| Failed trace-wide search | 1 |
| Explicit Run trace search retry | 1 |
| Failed initial trace expansion | 1 |
| Retry trace context: spans and linked logs | 2 |
| Find linked logs: failed log question only | 1 |
| Pending search retired on close | 1 |
| Technical details, retired retry and close | 0 |

Both search attempts submitted the identical whole request. Recovery retained
effective typed environment policy, fixed UTC bounds, both span conditions and
exact trace identity. The admitted trace contains the root and worker span IDs;
its successful basis and capture clock remain separately recorded when log
recovery returns one exact linked record. The unlinked neighbor and
foreign-environment linked decoy are excluded.

The old actual trace-retry callback performs no IO or local replacement after
successful recovery. Inspector close returns focus to **Open trace (fresh read)**,
connected, within the viewport and unobstructed at its center. The pending search
is retired on close, and its late refusal neither opens results nor publishes
into the closed owner.

- [Trace search refusal and keyboard-open details](screenshots/345-trace-search-refusal.png)
- [Trace expansion refusal and explicit recovery boundary](screenshots/345-trace-context-refusal.png)
- [Known spans with unavailable linked logs, desktop](screenshots/345-linked-logs-refusal-1280.png)
- [Known spans with unavailable linked logs, narrow](screenshots/345-linked-logs-refusal-390.png)
- [Independent exact linked-log recovery](screenshots/345-linked-logs-recovered.png)

Compact fixture questions, real results, refusal provenance and runtime identity
are retained under `artifacts/context-recovery345/`. The unchanged 340 typed trace-search regression also passed: **2.5 minutes**
for the test (2.6-minute runner), observed Quint check **9.86 seconds**, no skips
and no timeout override. Its 40.162-second Bombadil campaign produced 739
snapshots through phase 8, with source 4, verified 4, preserved policy and no
overflow. Four typed-policy profiles and their current maintained frames were
regenerated.

The separate driver-image audit found 739 images with three unique hashes;
729 repeat the terminal image. That image shows a real three-span waterfall but
still says “Reading exact linked logs…” with Find disabled. The final observed
DOM state explicitly reports captured logs, and the independent response oracle
verified one linked log. The driver image is stale at acquisition and cannot
establish settled log appearance. The new 345 native frames show both the partial
and completed log states. One labelled terminal sample and the audit remain with
compact logical receipts; raw campaign/report bulk was pruned after review. The
lab's telemetry, saved views and 51 pre-existing pending stream entries are
preserved.
