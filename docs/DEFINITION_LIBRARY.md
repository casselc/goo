# Compact definition libraries

B1 adopts one supplied-data presenter in the record telemetry-package and native
measurement catalogs on Definitions, the fixed reusable-question list and the
investigation revision catalog. Each compact row retains its exact
revision; selecting it opens one detail with explicit Edit/Explore or Preview actions.
Selection itself does not validate, execute telemetry, publish annotations or replace
the package editor draft. Native Preview retains its existing source-free compiler
step; record Explore retains its existing explicit execution contract.

The application adapters own list requests, refresh generations, authoring and
execution. `GooDefinitionLibrary` owns only loaded-row paging/search and the shared
inspector. There are 25 displayed rows per page; search covers the loaded catalog,
not a global index or all evidence. Its pure page model preserves order, exact IDs,
missing matches and typed record/measurement/document identity without merging titles/names.
Native and record refreshes reject older responses. A failed refresh retains old
rows and the selected detail, with an unavailable notice.

Only the selected record revision has an attached annotation panel. The existing
inspector caches up to 32 visited details in this tab, so switching revisions can
restore an unsaved annotation draft. Filtering/paging does not silently change the
selected target; a notice identifies an out-of-page/filter selection. Close returns
focus to its row when that row is available. This cache is not durable draft recovery;
eviction/reload loses local drafts. Annotation history, explicit parallel heads,
pinned old revisions and native retry/await-token semantics remain in their existing
owners. No new retention or review semantics are introduced.

```mermaid
flowchart LR
  Records[Record package adapter] --> Data[Supplied typed immutable revisions]
  Native[Native measurement adapter] --> Data
  Fixed[Fixed reusable-question adapter] --> Data
  Documents[Document catalog / server cursor owner] --> Data
  Data --> Rows[Compact loaded-catalog rows / paging / local search]
  Rows --> Inspect[Explicit selection: one cached inspector detail]
  Inspect --> Meaning[Selected record annotation panel]
  Inspect --> Actions[Explicit Edit / Explore / Preview / Open revision]
  Actions --> Owners[Existing typed compiler / editor / source owners]
```

Journey 110 creates unrelated same-title record/native definitions through the real
Go APIs. It checks exact selection, one attached annotation panel, editor and
annotation-draft preservation, no source/publication reads on selection, mobile
containment and a delayed real catalog response overtaken by a newer refresh.
Journeys 40 and 98 protect native continuation and existing annotation branches/
history. The library Bombadil campaign inspects, filters and pages without execution.
Hegel generates loaded catalogs/pages to protect exact identity, complete paging,
literal search and non-mutation. Existing annotation Quint ledgers protect unchanged
publication semantics; the DOM composition is not exhaustively modeled.

The fixed reusable-question list now uses these same rows and selected detail.
Selection/close invalidate its pending compiler plan; workspace/mode changes and
fixed-owner suspension do the same, with explicit non-loading status on return. A plan must still belong to the current owner
and resolve the exact selected revision before its explicit Run can dispatch a
source query. Inspection/filtering/paging never run a source. The existing Run
semantics still use a fresh default-lookback window, not captured evidence.

Journey 111 holds real compiler responses across selection/close/workspace changes
and fixed suspension into Activity, proving zero obsolete queries. A mismatched returned revision is refused. A subsequent
explicit Run reads exactly one fresh question. Core journey 33 and its Bombadil/Quint
ledger protect full-population measures and real definition reuse.

Bounded B1 consolidation is delivered across Definitions, fixed reusable questions,
document revisions and compatible retained metadata rows. Existing server cursor
contracts remain independent; no new family/lineage API, inferred head, global search
or store migration is introduced. B2 owns the next protocol/qualified-field work.

Investigation rows inspect already supplied revision metadata without opening the
editor or loading referenced evidence. The separate **Open investigation revision**
action retains the document owner's pending-note/dirty/busy guards. Inspection can
therefore coexist with a pending draft; an explicit refused open leaves it intact.
Rows show UTC custody time; the inspector carries the full selected and parent IDs.
Titles never imply family membership or a latest head. The server's **More revisions**
cursor remains separate from the presenter's 25-row local pages. Search explicitly
covers only the current server page. Selected detail survives local filtering or a
server-page change; unavailable refreshes preserve existing rows. Retained evidence uses the compatible row/inspector primitives described below.

## Retained catalog snapshots

The retained-evidence catalog reuses the compact supplied-data `row` primitive and
shared inspector, rather than the immutable-definition catalog controller. Its
server owns 20-item cursor pages and tag/environment/literal-service/UTC capture-time
filters. Selection inspects only the supplied catalog metadata; the explicit **Open
retained selection** link loads observations and review history in the retained view.
Local inspection, filter edits and disclosure issue no manifest/review detail,
source or publication request. Explicit Browse reads catalog metadata, including
current review-head tags, through its existing endpoint.

A manifest is immutable, but catalog tags reflect active review heads at the time of
the request. Inspector keys therefore include a catalog-load generation and manifest
ID. Selected metadata is explicitly a snapshot, never a pinned review. Filtering or
refreshing preserves that snapshot with an earlier-load notice until another row is
inspected; a failed refresh keeps the existing rows/detail. Old responses cannot
replace a newer filter/page. A new explicit inspection after refresh uses the new
snapshot. Exact review citations continue to use their own revision ID and read bound.

Retention ownership is now split without changing native publication contracts:

| Owner | Responsibility |
| --- | --- |
| `retention-client.js` | Native transport, tab-local read bounds, original request journaling before send and acknowledgment |
| `retention-catalog.js` | Server filter/cursor requests, stale-response guards and supplied metadata selection |
| `retention-citations.js` | Exact pinned review reads, explicit review choice and guarded investigation attachment |
| `retention-recovery.js` | Explicit retry/forget UI over the native-only browser journal |
| `retention.js` | Selection authoring and retained observation view composition; existing public facade |

```mermaid
flowchart LR
  Selection[Explicit retain / review] --> Client[Retention client and read bounds]
  Recovery[Explicit retry / forget owner] --> Client
  Client --> Journal[Original native request journal]
  Client --> Native[Native retained API / XTDB]
  Catalog[Catalog filter / server cursor owner] --> Client
  Catalog --> Row[Shared passive row / inspector]
  Row --> Snapshot[Selected catalog snapshot / no records read]
  Snapshot --> Open[Explicit open retained selection]
  Open --> View[Retained observation view]
  View --> Client
  Citations[Exact review citation / attachment owner] --> Client
  Citations --> Documents[Guarded document adapter]
```

Journey 101 now checks source-free catalog inspection, actual delayed response
rejection across a filter change, failed-refresh retention, pending request recovery
and old pinned citations after review correction. Its Bombadil actor also selects
and closes catalog metadata while retaining publication/reference invariants.
