# Worker and queue adapter

The sample now exercises the `goo.worker-queue` prebuilt family with actual Redis
Streams and OTel telemetry. Open **Queue and workers (USE)** or choose the USE
guide. This is a reference adapter for the semantic-package design, not a generic
claim that every broker's queue-depth metric has the same meaning.

| Instrument | Meaning | Unit |
| --- | --- | --- |
| `lab.queue.depth` | Redis inventory group's undelivered lag from `XINFO GROUPS` | `{message}` |
| `lab.queue.pending` | Delivered but unacknowledged entries, including active and stale entries | `{message}` |
| `lab.worker.active` | Tasks currently executing in this consumer process | `{worker}` |
| `lab.worker.capacity` | Actual dispatch bound: seven tasks | `{worker}` |
| `lab.worker.utilization` | Active / capacity in one locked observation | `1` |
| `lab.queue.wait` | First worker start minus Redis-assigned stream-ID enqueue time | `s` histogram |

These are **custom names**, emitted with scope schema `goo://demo/queue/v1`.
Identity uses standard `messaging.system`, `messaging.destination.name` and
`messaging.consumer.group.name`. Resources carry `service.instance.id` so process
restarts and different consumers are not merged into one gauge stream.

Queue wait uses the millisecond Redis enqueue clock and the consumer wall clock.
It includes delivery/scheduling delay before the first attempt; it excludes
later retry processing. It is not acceptance-to-completion duration or immutable
custody time. Negative elapsed time produces a clock-violation diagnostic and
no wait observation. Cross-host clock synchronization remains an assumption.
The first attempt span records exact decimal `lab.queue.wait_ns`; subsequent
attempts do not duplicate it. A processing-start event provides supporting logs.

Broker observation errors or unknown Redis lag produce no gauge point, never a
healthy zero. Group snapshots are approximate and separate calls are not a
transactional broker view. In a multi-consumer deployment the broker values are
group-wide; do not sum repeated group observations across consumer instances.
Pending includes stale entries left by crashes; it is not equivalent to backlog.
No pending entries are reclaimed or deleted by this change.

Worker active count is protected by a lock and released in `finally`, including
exceptions. Capacity matches both the executor and the maximum fetched batch.
The demo waits for a whole batch before reading the next one, so backlog can grow
while some slots are idle. The chart can expose that scheduling limitation; it
does not infer that the queue is healthy from low worker utilization.

The view plots one-minute **last-observed gauge snapshots**, separately by
instance. It does not average utilization or sum cumulative metric points.
Queue-wait p95 is calculated directly over raw first-attempt span observations
within each minute/instance, not by merging percentile summaries. The exported
histogram remains available for future temporality-aware operators.

Points older than 30 seconds are labelled stale and drawn hollow; no-data is
shown as unknown. Connecting lines are dashed aids between adjacent observed
minutes, not claims about unobserved intermediate states. The 1000-row view cap
is labelled; it is an inspection view, not a production objective calculation.
CPU/memory/pool pressure and complete error rates remain separate capabilities.

Checks cover bounded active state and exception release, missing broker evidence,
clock violations, per-instance snapshots and raw wait percentile computation.
The live browser journey creates slow work, observes active workers through the
collector, inspects all four panels and opens the reusable prebuilt catalog.
