# Tags, notes and retained evidence

Goo's shared annotation contract attaches authored context to an exact target.
It does not turn a note into an observed fact, service-health verdict or classifier
label. The [whole-product design](PRODUCT_ANNOTATIONS_DESIGN.md) remains the roadmap
for richer review, relationships, training vocabularies and dataset exports.

## Cross-product metadata

Open **Tags and notes** on an exactly scoped service, saved exploration view,
investigation revision or saved telemetry package. Enter lowercase tags and a
note, then explicitly save. Service annotations apply to the displayed fixed
half-open time window and exact environment/service/namespace identity. Missing,
present-empty and named namespaces are distinct; an all-namespaces selector is
not an annotation target. Saved artifacts refer to exact content revisions.

A correction appends a child revision. Two corrections of the same parent remain
independent branches; the interface never uses last-write-wins to pick a reviewer.
Withdrawal appends another revision and retains the earlier content. Show history
or open an exact revision link to inspect older assertions. Correcting an older
service assertion retains that assertion's applicability window.

The **Tags and notes across the workbench** disclosure in Saved investigations
finds current active revision heads by exact tag. This is a metadata catalog,
not a telemetry predicate, inherited label or complete service inventory filter.
Opening, editing, cancelling and saving annotations perform no telemetry reads.

The initial file store has bounded catalog scans (10,000 directory entries),
20-item pages and explicit cursor restart errors when the selected catalog changes.
Each artifact is content-addressed and integrity-checked. Attribution is explicitly
`local-user-unverified`; this local lab has no authenticated reviewer identity or
multi-tenant authorization. Notes are bounded to 4,000 UTF-8 bytes and tags to 16
unique lowercase slugs of up to 64 ASCII characters.

## Optional XTDB retained-evidence experiment

The optional backend is a Go adapter using pgx and XTDB 2.2.0's PostgreSQL wire
protocol. It is an experiment alongside the normal annotation store, not a
migration of all metadata to XTDB or a native ClickHouse external-source connector.

Start the ordinary lab, build the query/packages image, then enable the overlay:

```sh
docker compose -f compose.yaml -f compose.xtdb.yaml up -d --no-build xtdb packages query
```

The overlay pins XTDB's image digest, binds ports to loopback, limits memory/CPU,
and preserves its data in a dedicated volume. It sets `GOO_XTDB_URL` only on the
Go packages service. The UI adapter checks configuration; passive display elements
never access the database. The standalone database configuration is for the local
experiment, not a production topology.

1. Run a question or **Show recent activity**, then select up to 20 displayed
   inspection rows. The Go API has already held the exact page; retention does not
   reread the query. Saved captures still support selecting their stored rows.
2. Expand **Retain selected evidence in XTDB**, enter tags/note, and explicitly save.
   Selection and drafts publish nothing. Refreshing, paging or changing scope clears
   row selection; returning to a paused parent retains its original page selection.
3. Follow **Open retained evidence**. Inspect attributes and correct or withdraw the
   review using the shared annotation panel. Reloading reads XTDB's retained data.

Live result pages have opaque 256-bit tokens, a two-minute TTL and process-local
custody. The cache admits at most 64 pages / 64 MiB, with a 2 MiB serialized page
limit. It rejects new snapshots at capacity rather than evicting an unexpired page;
ordinary query results remain available with a retention diagnostic. Expired or
restarted-process tokens cannot publish new evidence. A confirmed absent request
returns 410 after expiry; an unavailable recovery store returns a retryable error,
keeps the original draft/key, and never claims the earlier save was absent. These tokens are capabilities
in the unauthenticated lab, not tenant isolation or authenticated provenance.

`POST /explore` returns `resultToken` when the page was held. Publication accepts
exactly one `captureID` or `resultToken`, selected `indexes`, `requestID`, tags and
note. Client record payloads are rejected. Live indexes are relative to the returned
page; the original query offset is retained separately. Saved-capture indexes are
relative to its complete stored inspection bank. A live manifest has
`origin: "result-page"`; its structurally shared `capture` receipt identifies the
original page token/question, not a saved-capture file. The UI labels that origin
and offers no fabricated saved-capture link.

The native transaction publishes the manifest, selected member projections,
initial annotation and request receipt together. A stable request key lets the UI
recover the same initial annotation after a lost response, including after its
ephemeral page expires; reusing that key for a
changed intent fails. Member identity includes its original page or capture row position,
so identical payloads at two positions remain two observations. Content hashes
check integrity; they do not attest origin or establish unique physical events.

The manifest names the original page or capture receipt/question, evidence basis, source
count, retained-record count and selected indexes. It expressly describes unknown
population completeness. Record JSON is retained without conversion through
floating-point values; original nanosecond strings and JSON numeric tokens survive.
API receipts and manifest reads also include `manifestJSON`, the exact serialized
artifact as a string. This survives JSON proxies and browser decoding; rebuilding
the artifact from display objects may change escapes, key order or numeric tokens
and must not be used to verify its identity. Native member columns project
service, reported trace/span IDs and event-time
nanoseconds for joins within an explicitly selected manifest. The projected service
name alone is not a full service identity; environment and missing/empty namespace
semantics remain in the capture scope and resource attributes. Arbitrary attributes remain in the retained JSON payload;
this spike does not provide an arbitrary-attribute XTDB query compiler.

```mermaid
sequenceDiagram
    participant U as Investigator
    participant G as Goo Go API
    participant S as ClickStack
    participant P as Bounded page custody
    participant X as XTDB
    U->>G: Run fixed question
    G->>S: Explicit source reads
    S-->>G: Rows / independent counts and groups
    G->>P: Hold exact page bytes, question and basis (2 min)
    G-->>U: Display rows + opaque result token
    U->>G: Retain token + selected indexes + tags/note + request key
    G->>P: Resolve exact held observations
    G->>X: Atomic manifest + members + annotation + request receipt
    X-->>G: Commit and await token
    G-->>U: Retained evidence link + committed read lower bound
    U->>G: Retry same intent after lost response
    G->>X: Recover original request receipt
    G-->>U: Same annotation and manifest
    U->>G: Correct annotation + retry key + await token
    G->>X: Await prior commit, atomically append correction and request receipt
    X-->>U: Stable revision + new await token
    U->>G: Reopen evidence / history + await token
    G->>X: Await lower bound, read retained evidence and branches
    X-->>U: Exact evidence and authored interpretation history
```

Captured event time, original custody/basis, capture time, metadata recorded time
and XTDB system time remain separate. Native system-time SQL can reconstruct which
assertions existed at an earlier cutoff. The application still derives branches
from explicit parent edges. An immutable evidence target does not receive a
fabricated business-valid interval merely because the database is bitemporal.

## Retry and read-basis contract

Native annotation POSTs require a 64-hex `requestID` alongside the existing typed
annotation input and optional `awaitToken`. Repeating the same key and intent
returns the original revision. Concurrent different review requests preserve
branches. Changed intent under a used key fails. Correction and request receipt
commit atomically; interrupted transactions leave neither. The UI keeps a key
for a failed draft, creates a fresh one for a new/cancelled draft, and disables
editing while a save is pending. Reload/navigation recovery of an unacknowledged
request draft remains future work.

Publication and correction return an XTDB `awaitToken`. Native evidence/history
GETs accept `?awaitToken=...`; native correction POSTs accept the same field. Each
new pgx connection applies `SET AWAIT_TOKEN` before reading. The UI remembers the
latest successful token per evidence ID in tab-local session storage and sends it
on immediate reads and reloads. API clients must propagate the receipt themselves;
the bare shared evidence link does not carry a commit basis into another tab or
client. This is a read-your-write lower bound, **not** an exact historical snapshot
or system-time cutoff. An invalid/unavailable basis fails rather than silently
reading an older state. No distributed XTDB replica-lag experiment is claimed;
integration tests exercise fresh connections against the local native instance.

## Prototype limits

The retained-evidence catalog is separate from the file-backed global tag catalog;
there is no cross-store search, production permissions/redaction/deletion policy,
training export or historical-cutoff UI. Live pages remain ephemeral until an
explicit save. Selection retention covers Explore/recent-activity result tables
and saved captures; SSE tail rows and waterfall span selections do not yet offer
this action. Native initial and correction retries are idempotent; file-backed
metadata writes retain their existing non-idempotent API. Retained annotation
history is bounded to 1,000 revisions and fails explicitly beyond that bound.
The evidence view links the retained manifest; exact file-annotation deep links
are not reused as native annotation links.

See the [storage assessment](ANNOTATION_STORAGE_SPIKE.md) for Minigraf/XTDB tradeoffs,
external-source clock semantics and the next conformance requirements.

## Executable evidence

[Journey 98](../journeys/annotations.spec.js) covers service/view/document/package
annotations, cancelled drafts, exact links, parallel corrections and withdrawals,
with no additional source reads. Its observed ledger is checked through quint-go,
including negative controls for hidden source reads, implicit publication and
incoherent scope/history. Bombadil explores draft/cancel/history interactions.

[Journey 99](../journeys/retained-evidence.spec.js) emits real OTLP spans, creates an
explicit source capture, selects its stored records, loses a successful native
publication response, retries and reopens the retained evidence with browser source
API reads blocked. This is not a claim that the ClickHouse process was stopped.
Bombadil also explores draft/cancel/history on the retained-evidence surface.
Native Go integration tests exercise rollback/retry, concurrent request recovery,
temporal cutoff queries and immutable record
round-trips against a real optional XTDB instance. Hegel state/swarm campaigns
check the common annotation branch/history contract. Screenshots are in
[the journey index](journeys/README.md).

[Journey 100](../journeys/live-retention.spec.js) selects a real OTLP observation
from recent activity, inserts a later observation, and retains the original page
without rereading it. Both initial-save and correction responses are deliberately
lost after actual commits; retries recover the same revisions. Reloads carry the
committed read basis while browser telemetry APIs are blocked. Pagination/refresh
clear selection. The observed ledger is checked against the shared Quint model,
including original-token, scope and selected-member negative controls. Bombadil
also exercises live selection, unpublished drafts and clear. TTL/capacity and
expired-page recovery are checked in Go with a controlled clock/native adapter;
the browser test does not delete source telemetry or simulate replica lag.

The release check additionally restarts the Go service and XTDB: the unpublished
page cache disappears, while the same committed request recovers its original
annotation and exact manifest. An unknown request against that unavailable page
returns 410 without adding an assertion. Native history remains unchanged; see
[validation evidence](../VALIDATION.md).
