# Saved capture search contract — 244 (supersedes the 242 index)

The immutable files remain authoritative. Their bytes, IDs, publication and direct
ID reopen stay unchanged. Search uses derived per-capture summaries in `.catalog-v2`. Search pages declare `goo/capture-search/v2`.
The old list cursor remains an ID position; the new search endpoint's cursor binds
its query, admitted capture ID set, metadata digests and last result ID.

## Bounded discovery projection

Search items are `CaptureSummary`, never incomplete `CaptureReceipt` objects.
They carry ID, optional parent, title, canonical capture timestamp, question kind,
optional measurement kind, declared signal and nullable environment. The canonical
`CaptureQuestion.Scope` supplies signal/environment; full selectors, measurement
bindings/definitions and paired questions remain only in the immutable payload.
The display subtype distinguishes gauge, counter, histogram, lifecycle, stock
and state without retaining a definition. Unknown environment is null; exact empty
remains distinct. No service predicate is inferred.

Accepted title/environment strings are at most 120/128 UTF-8 bytes. IDs are fixed
SHA-256 strings; kinds/signals are closed enums. Parsed capture timestamps are
formatted with bounded nanosecond precision (normally UTC; boundary-year offsets
are preserved when UTC would leave RFC3339). Thus HTML escaping or arbitrarily
large valid questions cannot inflate an entry. Full original receipt timestamps
and questions remain unchanged on exact-ID responses. Capture-time sorting still
uses instants, with ID ties.

This replaces v1's full-receipt index, which could not index some accepted
questions after JSON escaping. Existing `.catalog-v1` contents are ignored and
left untouched. Rebuild explicitly verifies authoritative files into v2, including
already committed captures that v1 could not index. Old search cursors are refused
until the caller restarts. No payload/ID migration or automatic search-time scan
is performed. The 256 MiB whole-store repair limit still applies; targeted repair
is a separate roadmap item.

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

Rebuild admits at most 10,000 nonreserved physical directory entries and 256 MiB aggregate payload
bytes, with the existing 24 MiB per-capture bound. It fully decodes, digest-checks and
semantically validates each capture before replacing its metadata. Failure can
leave a valid partial derived catalog; the search still refuses any missing/bad
entry. Rerun is idempotent. It never rewrites, repairs or deletes capture payloads.
Orphan metadata is ignored because enumeration starts from authoritative IDs.

## Query budgets and integrity

A search admits at most 10,000 nonreserved physical capture-directory entries, 16 MiB accepted aggregate metadata bytes, 64 KiB per metadata
entry, and 8 MiB retained summary JSON. Readers allow one overflow-detection byte.
The known `.catalog-v1` and `.catalog-v2` directories are separately reserved
(at most two); unrelated directories, links and temporary files still count.
Enumeration reads at most 10,003 entries including its overflow sentinel. Search,
rebuild and the legacy list share this policy so retaining v1 during migration
cannot exhaust the directory budget. Preflight sizes and actual reads both enforce limits. These are bounded input/
retained-metadata budgets, not hard RSS or timing guarantees. Search does not open
payload bodies. It validates entry version, checksum, ID, bounded summary fields/title/canonical time and
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
