# Saved capture search contract — 242

The immutable files remain authoritative. Their bytes, IDs, publication and direct
ID reopen stay unchanged. Search uses derived per-capture metadata in `.catalog-v1`.
The old list cursor remains an ID position; the new search endpoint's cursor binds
its query, admitted capture ID set, metadata digests and last result ID.

Search is literal case-insensitive title containment, exact question kind,
known signal, and optional exact environment. Omitted environment means any;
present-empty means exact empty, never unknown. Explore/inventory name an
environment; comparison/explanation require agreement between both sides.
Search reuses the existing `CaptureQuestion.Scope` primitive: native measurements
report metrics and their declared environment; lifecycle measurements report logs.
Inventory is span-backed. Signal and environment must each agree across pair members;
a mixed environment does not erase an otherwise known signal. Arbitrary selectors
are not mined for scope. There is
no service filter: OR, negation and multi-service populations make inference unsafe.

Order uses parsed receipt capture-time instants, newest/oldest, with ascending ID
ties. It is not event time, file mtime or ID order. Same titles remain distinct.
Search covers the whole admitted catalog before paging 20 results. A changed query,
catalog membership or metadata digest invalidates pagination; restart page one.
A bounded directory read is not a transactional filesystem snapshot.

## Publication and rebuild

After publishing and fully verifying an immutable capture, the writer atomically
replaces its derived metadata entry (sync temporary file, rename, sync directory).
An index write failure does not turn an already committed capture into an apparent
failed publication. Search explicitly refuses missing metadata instead of omitting
it. An immutable duplicate is verified before metadata is repaired. No index file
is required to reopen by ID. Concurrent writers/rebuilds for the same immutable ID
write the same metadata; no shared append log or whole-index rewrite is needed.

Existing stores require an explicit administrative rebuild, not a hidden UI source
read or migration:

```sh
goo rebuild-capture-catalog -dir /path/to/investigations/exploration
```

Rebuild admits at most 10,000 physical directory entries and 256 MiB aggregate payload
bytes, with the existing 24 MiB per-capture bound. It fully decodes, digest-checks and
semantically validates each capture before replacing its metadata. Failure can
leave a valid partial derived catalog; the search still refuses any missing/bad
entry. Rerun is idempotent. It never rewrites, repairs or deletes capture payloads.
Orphan metadata is ignored because enumeration starts from authoritative IDs.

## Query budgets and integrity

A search admits at most 10,000 physical capture-directory entries (including the
metadata directory), 16 MiB accepted aggregate metadata bytes, 64 KiB per metadata
entry, and 8 MiB retained receipt JSON. Readers allow one overflow-detection byte.
Preflight sizes and actual reads both enforce limits. These are bounded input/
retained-metadata budgets, not hard RSS or timing guarantees. Search does not open
payload bodies. It validates entry version, checksum, ID, question/title/time and
payload stat stamp (size and mtime); nonregular files are refused. Cursors include
metadata digests, so changed metadata cannot silently reorder a continued page.

A checksum is local consistency, not authentication. Stat stamps catch ordinary
payload changes but cannot establish arbitrary content integrity. Search therefore
says **metadata only; opening verifies captured contents**. A same-size mutation
with restored mtime can remain listed; reopen still fully rejects invalid content.
Rebuild also fully verifies content. The legacy list retains its original full
verification. This split is deliberate, not a claim that an index verifies payloads.

Missing metadata is 503 with rebuild/direct-ID guidance; invalid metadata or a
changed payload stamp is an integrity refusal, not an empty result. Scan/byte limits
are 503; invalid filters, cursors and incompatible pagination are 400. Empty matches
are a successful empty page. Missing capture directories are empty; inaccessible
storage is unavailable. No catalog action or reopen queries telemetry.
