# Live log observations in Explore

An applied log question offers **Tail matching logs**. The app retains its exact
environment and predicate tree, and explicitly replaces the fixed bounds with a
rolling `[now − 60 seconds, now)` event-time window. The fixed question/result
stays separate; tail rows never supply its count, measures, capture or comparison.
Saved/read-only questions have no automatic subscription. Running/editing a
question, changing scope, leaving Explore or selecting Measurements stops the
stream and clears the local buffer.

The Go `/tail` endpoint accepts POST JSON `{environment, where?}` with the same
bounded log AST validation and parameterized physical adapter as ordinary Explore.
The Python app proxies `/api/tail` as flushed SSE rather than buffering JSON. The
app uses fetch streaming so selectors remain in the request body, not a URL.
Components have no network behavior.

Each connection emits `reset`, sequentially numbered `batch` events (including
empty heartbeat batches), then an explicit `gap` when possible. Reset identifies
the selector, 60-second window and `replay:false`. A poll every two seconds reads
up to 1,000 newest records using descending event time and deterministic tie
ordering. Each read keeps the existing source timeout/byte/memory budgets, and
limits emitted raw records to 512 KiB per batch. Budget limits set `limited`;
the UI retains a continuity warning. There are four concurrent tails per Go
process and each connection lasts at most 60 seconds. Source failure or identity
budget exhaustion ends a connection. Writes have a five-second deadline.

This exporter has no ingestion cursor or immutable log event ID. A connection
tracks the maximum observed multiplicity of each canonical full-record SHA-256
fingerprint; fingerprint/occurrence pairs are **observation handles**, not event
identity. Exact duplicate rows can appear multiple times. Repeated overlap polls
do not append those same occurrences again. Deletions/reinsertions of identical
records, retention and late arrivals can defeat any inference about unique events.
Seen fingerprints are bounded to 20,000 and discarded with the connection.
`late` means an event time earlier than a previously observed maximum, not an
export/ingestion delay measurement or a completeness watermark. Records outside
the rolling window and records beyond read budgets can be missed.

The browser retains at most 100 observations and 2 MiB, in observation order.
Each bounded source batch is presented in event-time order, so batch eviction
keeps its newest records; later polls can append older late events. Retained
observations can age beyond the rolling source-read window.
The UI labels retained/evicted/observed counts as local connection counts, never
population totals. DOM updates append/remove caller-owned rows without rebuilding
existing rows. Follow-scroll can be disabled, and scrolling away from the end
disables it. Keyboard focus inside the viewport pauses the stream before a focused
row could be evicted. Attribute inspection reuses the shared typed viewer.

**Pause stream** cancels collection while retaining inspected rows. **Resume with
recent logs** starts a new recent window and resets the buffer after the new
server reset. Reconnects retry after two seconds and visibly retain a continuity
gap; there is no Last-Event-ID replay or durable cursor. Loading, stale (no
completed read for six seconds), disconnected and limited-source states remain
separate from application health. A scope generation guard prevents late frames
from an old connection publishing into the current selector.

```mermaid
sequenceDiagram
  participant User
  participant Explore as Explore tail adapter
  participant Proxy as Python SSE proxy
  participant Go as Go tail source adapter
  participant CH as ClickHouse logs
  User->>Explore: Tail applied log selector
  Explore->>Proxy: POST /api/tail {environment,where}
  Proxy->>Go: POST /tail
  Go-->>Explore: reset (no replay; rolling window)
  loop Every 2 seconds, bounded connection
    Go->>CH: Parameterized recent-row read
    CH-->>Go: Bounded captured records
    Go-->>Explore: batch (new observed multiplicities, limits, clocks)
    Explore->>Explore: Patch retained rows; cap count and bytes
  end
  User->>Explore: Pause or inspect with keyboard
  Explore->>Proxy: Abort stream
  User->>Explore: Resume with recent logs
  Explore->>Go: New connection via proxy
  Go-->>Explore: Reset; show continuity gap
```

Journey 66 injects ordinary OTel logs, including identical duplicates, shared
trace/span IDs, a foreign environment, late event times and enough rows to evict.
It checks focus pause, source separation, explicit resume gaps, scope cancellation,
mobile containment and Quint replay with rejecting mutations. Hegel generates
changing overlapping multiset polls and late arrivals against the Go tracker.
The existing source is not a durable append log; guaranteed catch-up, ingestion
clocks, production tenancy and baseline-versus-live measures remain future work.

## Compact Live Activity composition

Fresh entry reuses this transport with compact supplied rows and application-owned
inspection: [Live Activity](LIVE_ACTIVITY.md). Scroll-away pauses transport; Resume
shows the new-window gap. Reconnect retains the last-good buffer until the first
successful replacement batch. Explicit scope changes clear the old buffer. The
HTTP proxy closes upstream reads on downstream disconnect; the four-reader
admission limit is unchanged.
