# Keep exact Activity observations — 257

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

## Deliberate limits and next slice

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
