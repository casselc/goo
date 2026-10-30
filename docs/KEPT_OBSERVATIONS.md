# Keep exact observations

258 supports Activity, fixed Explore results and Saved record selections. The
[fixed/Saved extension](#fixed-results-and-saved-records--258) below specifies the
current index and provenance contracts; the initial Activity slice is recorded first.

Select one or more Activity rows, then **Keep selected**. Selection pauses the
stream. Keeping requires neither a title, tags, an interpretation nor XTDB.
**Open saved** opens the selected records and their observation provenance.
Inspection, reopening and retrying do not query telemetry.

This first slice supports 1–20 server-attested log or span occurrences from one
Activity connection and selector, including checked environment scopes. Duplicate
records retain distinct occurrence keys. These are the server's canonical observed
records, **not original OTLP bytes**, proof of unique physical events, a complete
population, or reviewed classifier labels. Record time, original read window,
observation time and immutable keep time remain distinct; the UI displays UTC.
Fresh **Capture evidence** still executes a fixed question and has different semantics.

```mermaid
sequenceDiagram
    participant U as Activity UI
    participant B as Go membership bank
    participant F as Immutable local file
    U->>B: Tail poll (explicit Start)
    B-->>U: Observed records + short-lived membership references
    U->>U: Select occurrences; pause stream
    U->>B: Keep selected (request ID + references only)
    B->>B: Verify connection, selector, occurrence membership
    B->>F: Copy trusted records; atomically publish one envelope
    F-->>U: Keep receipt + exact ID
    U->>F: Retry same intent or Open saved
    F-->>U: Original receipt and records (no source query)
```

## API and storage

`POST /api/kept-observations/tail` accepts `goo/keep-tail/v1`, a 64-hex
`requestID`, connection, selector and 1–20 token/key references. It rejects record
contents, tags and unknown fields. `GET /api/kept-observations/{id}` returns the
immutable `goo/kept-observations/v1` resource. The OpenAPI specification includes
both routes. The resource lives under `kept-observations/` in the configured
investigations directory, separate from existing captures and XTDB revisions.
No old identity or catalog format changes.

Each request ID has one intent-bound winner. The server checks that winner before
resolving short-lived references, allowing an identical retry after expiry or
process restart. A changed intent returns 409; unavailable original membership
returns 410 rather than rereading or silently substituting records. Resolve copies
release the membership lock before disk IO. Four admissions bound simultaneous
new publication; each durable envelope is limited to 3 MiB. A fully written and
synced temporary file is linked atomically without overwriting a winner, then the
directory is synced. Live proof tokens are omitted from durable/public records;
only their intent digest remains. Digests detect accidental changes, not malicious
filesystem rewriting. The lab has no authentication or authorization boundary.

The browser journals the exact reference-only intent before POST. An uncertain
acknowledgement freezes the submitted selection; **Retry unchanged keep** reuses
that intent. Changing screens does not let a late result replace another view.
Saved captures' **Kept observations and pending keeps** disclosure exposes recent
keeps (16) and pending requests (16) for this browser. **Forget pending request**
removes only local retry state, never a committed server file; preserve the exact
link before forgetting an uncertain request. Storage failure refuses publication.

## Limits of the initial Activity slice (257; extended below)

There is no server-side keep catalog, store quota/GC, cross-browser discovery,
annotation attachment, fixed-result/Saved-record selection, or full trace graph
keep in this slice. Durable exact-ID links work across browsers using the same
server storage; browser recent history is only a convenience. An expired, never
committed observation cannot be kept later without a separately identified fresh
read. Optional XTDB retention remains separate and retains its selector gates.
Future origins must reuse this honest selection/provenance contract rather than
pretend a newly executed question was the earlier observation.

See the [executable storyboard](../journeys/257-kept-observations-storyboard.md)
and [validation](../VALIDATION.md).

## Fixed results and Saved records — 258

The same **Keep selected → Open saved** action now works in the existing record
selection panel for fixed Explore results and Saved record captures. Checked
environment selectors and ordinary named environments use the same path. No label
or XTDB is required. Changing pages or screens invalidates the old selection owner;
a late acknowledgement can update browser recovery but cannot replace the active
view. Clear selection leaves an uncertain request in explicit recovery.

`POST /api/kept-observations/records` accepts `goo/keep-records/v1`, a request ID,
exactly one `resultToken` or `captureID`, and 1–20 distinct indexes. Result indexes
are **local to the original held page**, preserving its nonzero question offset.
Saved indexes are **absolute in the retained capture row bank**; the UI adds its
Saved page offset. Duplicate identical rows retain distinct indexes, even though
their content digests match. This is occurrence identity, not unique physical events.

The optional `expected` question/basis is a precondition, never browser authority
for record contents. The UI persists it in the same retry intent; both Go and the
browser validate it before publication/acknowledgement. A different source scope
returns 409, and a mismatched reply remains pending without being added to Recent.
Browser retry recovery uses the identical precondition after reload.

The new immutable `goo/kept-records/v1` receipt retains origin, exact original
question, basis, read clock, original result state, retained-page/bank size and
limitation flag, selected indexes and record hashes. Saved origin also retains
capture ID and its separate custody clock. Counts/groups are not promoted to a
population claim. The synthetic live snapshot receipt contains the capability
itself and is deliberately not copied; live result tokens never appear in the
public/durable receipt. `GET /api/kept-observations/{id}` strictly dispatches either
receipt version; existing Activity v1 bytes and IDs are unchanged.

Both origins share the existing four-admission, 3 MiB atomic file publisher and
winner-before-resolution retry behavior. The result bank detaches bytes while
locked, then releases its lock before decoding, validation and filesystem work.
Saved origin resolves immutable storage without telemetry. An existing winner
still reopens after process restart, proof expiry or original capture deletion.
Cross-origin reuse of a request ID conflicts. Only record results are eligible;
explanation, measurement and inventory tokens/captures are refused.

```mermaid
flowchart LR
    A[Activity membership references] --> P[Shared atomic Keep publisher]
    R[Fixed result token + local indexes] --> B[Detach original server page]
    S[Saved capture ID + absolute indexes] --> C[Verify original stored capture]
    B --> P
    C --> P
    P --> F[Immutable versioned keep file]
    F --> UI[Shared Saved inspection and exact link]
    J[Browser original intent journal] --> P
```

The [258 executable storyboard](../journeys/258-keep-records-storyboard.md) covers
both origins. The earlier first-slice limits above are historical: fixed/Saved
origins are now implemented. Server-wide discovery, quotas/GC, optional later
annotation attachment and whole-trace keeps remain separate work.

## Server discovery and bounded repair — 259

In Saved → Browse evidence → **Kept observations and pending keeps**, choose
**Find kept observations** to discover keeps from every browser using this server
storage. Filter by origin (Activity, fixed result, Saved capture) and signal. Each
page shows up to 20 selections, newest keep time first with ID as a stable tie
breaker. Scope displays its original environment policy; unreported, empty and
named scopes are never merged. Selection counts are not telemetry population counts.
Local browser recent links and unresolved requests remain separate below.

`GET /api/kept-observations/search` reads derived metadata only, never telemetry
or full record payloads. Metadata retains receipt version, exact ID/content digest,
origin, signal, versioned environment policy/scope, selected occurrence count and
keep time. It contains no original question, source capability, record contents or
labels. **Open kept observations** uses the existing exact-ID verifier. Summary
checksums and file size/mtime stamps detect common staleness; they do **not**
establish record integrity or authenticity. Exact reopen checks authoritative
contents even if same-size corruption escaped discovery stamps.

Origin/signal changes only stage filters. Find applies them; More/Previous continue
the applied query. Cursors bind filters and the full metadata snapshot, so changed
catalogs refuse continuation rather than silently duplicating/skipping selections.
History advances only after successful reads. Collapsing either containing
disclosure, leaving Saved or hiding the tab aborts/invalidate work; no background
repair or automatic subscription resumes. A local Keep marks loaded metadata stale; a change from another browser
invalidates the next cursor read. Find is explicit.

### One admission contract

Publication and discovery share a **1,000 physical root-entry** limit; the real
`.catalog-v1` derived directory is reserved separately. Junk and unrelated entries
consume capacity. Every accepted resource must produce a valid summary fitting
**4 KiB** before its payload is published. Search reads at most 1,000 such metadata
entries (under 4 MiB) and returns at most 20. Two shared admission slots bound
concurrent discovery/repair requests. Existing four-slot Keep admission and 3 MiB
per-resource limits remain in force.

The shared `internal/directorylock` helper was extracted from native catalog
locking. Supported Unix platforms use shared directory locks for discovery and
exclusive locks across Keep admission, scratch publication/cleanup and metadata
indexing. Independent processes cannot both take the last slot. Unsupported
platforms fail closed for coordinated admission. This assumes a stable local
filesystem root and cooperating writers; it does not extend locks to external
editors or unsupported network filesystems.

Index publication is best effort **after** the authoritative keep commits; an
index failure does not turn a successful Keep into a failed save. Exact original
retries still succeed at/above capacity and do not rewrite metadata. Existing
unindexed keeps and interrupted metadata updates use explicit repair. GET never
creates directories or repairs anything. An existing over-limit store refuses
full discovery and new keeps without deleting resources; known-ID open/retry still
works. Repair cannot remove excess files or fix corrupt authoritative payloads.

### Repair across large stores

Missing/corrupt metadata refuses discovery visibly. When applicable, **Repair kept
index** performs explicit source-free `POST /api/kept-observations/repair` requests.
Each request verifies and indexes at most **64 resources / 32 MiB of payload**;
continuation binds payload IDs, sizes and mtimes, not the derived files it changes.
A 36 MiB store therefore repairs across batches instead of becoming irreparable
at a whole-store byte ceiling. Payload corruption or a changed continuation refuses
honestly. Completed metadata remains; original keep bytes are never rewritten.

The app continues bounded batches only while the task remains visible. **Stop index
repair**, disclosure collapse, navigation and tab hiding stop it. Restart is explicit
and safe; a batch may have completed before its acknowledgement was cancelled.
Capacity refusals do not offer repair as a false solution. Find explicitly reads the
repaired index. Disk-byte quotas, garbage collection, annotation attachment and
whole-trace graph keeps remain separate work.

```mermaid
flowchart TD
    Keep[Verified Keep candidate] --> Bound[Summary size + shared physical admission]
    Bound --> Commit[Atomic authoritative keep file]
    Commit --> Index[Best-effort bounded metadata]
    Find[Explicit Find / filters / page] --> Index
    Index --> Links[Discovery summaries and exact links]
    Links --> Open[Exact payload verification and shared inspector]
    Missing[Missing or corrupt metadata] --> Repair[Explicit resumable repair]
    Repair --> Batch[Verify at most64 resources /32 MiB per request]
    Batch --> Index
```

[259 storyboard and actual frames](../journeys/259-kept-discovery-storyboard.md).
Earlier sections preserve their initial-slice limits; server discovery and count
admission are now implemented.
