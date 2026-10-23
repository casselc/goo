# Exact Activity selection: membership prerequisite

Status: a bounded Go membership primitive is implemented and tested, **unwired to
SSE and HTTP**. Existing tail UI and retention origins are unchanged. This document
specifies the next integration; exact Activity retention and its browser retry
journal are not yet available.

The concrete task is to preserve up to 20 observed Activity log/span occurrences
without rereading telemetry. This is observation custody, not unique physical-event
identity, complete trace capture, a population denominator or classifier training.

## Existing contracts and the gap

| Existing path | What can be reused | What cannot be assumed |
| --- | --- | --- |
| `Source.PollTail` / `TailTracker` | Bounded rolling read, exact record fields, per-connection multiplicity handles and late-event diagnostics. | A hash/ordinal handle is not durable membership. The same handle can recur in a new connection. Polls have distinct windows/read clocks; no ingestion cursor or replay. |
| `ResultSnapshots` | Immutable server-owned bytes, short-lived capabilities and explicit capacity refusal. | It holds a whole fixed `Result` page. It cannot represent a mixed-poll tail selection or supply fabricated aggregates. |
| `retention.Manifest` v1 | Saved/result-page selection, exact JSON payloads, immutable publication and meaning revisions. | It requires an actual capture/result-page origin. The validator deliberately rejects a tail origin. |
| `Store.Publish` / `RecoverSelection` | Transactional request-intent binding, atomic receipt publication, committed-request recovery after source-page expiry, and distinct read/custody clocks. | `RecoverSelection` verifies capture/result-page indexes. It cannot recover tail references without a new origin-specific contract. |
| Native browser retry journal | Reuse the immutable-intent lifecycle after a server publisher supports it. | Browser state does not establish server membership or make arbitrary writes idempotent. |

Current `/tail` reset/batch payloads issue no membership capability. Current
`GooTail.snapshot()` omits per-row provenance from its public handoff. `/retained-evidence`
accepts only saved capture IDs or fixed result tokens and indexes; it rejects client
records and tail-reference fields. These boundaries remain enforced.

## Implemented Go primitive

`exploration.NewTailMembershipSnapshots(clock)` creates an isolated process-local
bank. It is constructed explicitly; no global bank or route is registered. Its
only issuer is intended to be the server's existing validated source-poll path.
Calling it with arbitrary browser record payloads is not an admissible integration.

`Freeze(TailMembershipBatch)` freezes connection ID, normalized selector, original
half-open read window, observed/read clock, limited flag and emitted `TailRow`s.
The batch is copied; later caller/response mutations cannot change the bank.
Canonical record encoding uses the existing tracker convention: sorted object keys,
JSON number tokens retained through `UseNumber`, no float coercion. A valid handle
must match the canonical payload hash plus its positive occurrence ordinal. Event
timestamps must lie within the original poll window.

`Resolve(connection, selector, refs)` accepts only server-issued tokens and handles.
It resolves all references under one lock or returns no draft. Selector normalization
covers the default signal and Go JSON defaults, not arbitrary Boolean equivalence. The result retains
per-member window, read clock, basis, late/limited flags and canonical payload.
Sorting references makes retries/order changes deterministic. Identical payload
occurrences with different ordinals stay distinct; selecting the same occurrence
through two tokens is rejected. Mixed connections or selectors are refused.

The `TailSelectionDraft` version/digest describes ephemeral proof, **not a persistent
retained manifest or receipt**. It has no manufactured `SavedCapture`, aggregate
count, custody clock or physical event ID. Source corrections cannot change already
frozen bytes; later observations require their own provenance.

| Bound | Current primitive |
| --- | --- |
| Read window | Original 60-second half-open tail window. |
| Capability expiry | Two minutes after the original observed/read clock, not after issuance or last access. Delayed issuance cannot renew it. |
| Batch | 1..1,000 emitted observations; at most 1 MiB serialized provenance/payload. Raw input is bounded before decoding. |
| Bank | At most 64 batches and 8 MiB. Expired entries are pruned; capacity refuses new issuance without evicting valid proof. |
| Selection | 1..20 unique occurrences, one connection/selector, at most 2 MiB. |
| Failure | Missing/expired/unproven members reject the whole selection. No partial draft, source query or reconstructed payload. |

The bank is thread-safe; its injected clock must itself be safe for its callers.
Tokens are process-local opaque capabilities, consistent with the existing lab's
result-page bank. They are not authenticated user identities. Restart loses
uncommitted proof. Disconnect does not renew or revoke a capability; a selection
can still name its original connection until its original expiry. A different
connection must not silently rebind it.

```mermaid
flowchart LR
  Poll[Existing bounded source poll] --> Tracker[Existing TailTracker occurrences]
  Tracker -. planned server issuer .-> Bank[Implemented membership bank: unwired]
  Bank --> Draft[Exact ephemeral selection draft]
  Bank --> Expiry[Read-anchored expiry or capacity refusal]
  Draft -. next integration .-> Origin[New typed tail manifest adapter]
  Origin -. planned .-> Publish[Idempotent durable publication and receipt]
  Publish -. explicit integration .-> Catalog[Origin-aware catalog and exact citations]
```

## Publication and retry contract — proposed next cut

Use a distinct typed tail manifest/receipt alongside current v1 capture/result-page
receipts. Preserve old bytes, hashes and validators. Reuse transactional publication
mechanics behind origin-specific validation; do not relax v1 to accept missing
captures. Catalog, exact-reference/citation resolution and pinned review reads must
explicitly understand the new origin before the UI advertises durable evidence.

A proposed `POST /retained-tail-evidence` request contains a version, client request
ID, original connection/selector, bounded token/handle references, tags and note.
It contains no client rows, hashes, basis, read windows or clocks to trust. The
server constructs those fields from its bank. The request ID binds the normalized
selector, sorted exact references and original tags/note; editing the intent
requires a different request, never silently changing a pending retry.

| Server outcome | Required behavior |
| --- | --- |
| Same request already committed | Return its original durable evidence/annotation identity even after capability expiry or process restart; do not resolve the bank or reread telemetry. |
| Same request ID, different intent | Refuse a conflict. Do not overwrite or republish the first intent. |
| Definitely uncommitted and proof valid | Resolve all members, validate the new manifest, atomically publish evidence + request intent + receipt. Record custody time independently of each read clock. |
| Definitely uncommitted and proof unavailable | Refuse expired/unavailable membership. Do not create a replacement source read or accept a client payload. |
| Request store unavailable / commit outcome unknown | Preserve the same pending intent and return retryable unavailability. Absence of a response is not proof of no commit; do not generate a new ID or claim simple expiry. |
| Issuance capacity exhausted | Future SSE should mark exact-preservation capability unavailable while normal observation can continue. Never advertise a capability that was not frozen. |

The existing saved/result-page transaction tests verify the reusable mechanics:
interruption leaves no partial publication, concurrent attempts converge, and an
already committed selection can be recovered after page expiry. They do **not**
prove durable tail publication. A tail publisher must separately pass lost-ack,
changed-intent, expiry, unavailable-store and concurrent retry tests against its
new manifest before it is connected to the browser journal.

## Planned user storyboard — not implemented

1. Activity receives server-issued connection and optional batch capability metadata.
   Each buffered observation retains its own reference/window/read clock; the latest
   global poll window must not replace earlier members' provenance.
2. Pause and select up to 20 observations from one original connection/selector.
   The held selection is local, with an explicit exact-preservation expiry. Eviction
   from the scrolling buffer does not turn a handle into saved evidence.
3. Preserve selected observations prepares one immutable request intent and sends
   exact references. A pending journal retains that intent across lost responses.
4. A committed receipt promotes the selection to saved evidence with independent
   read/custody clocks. Reload retry can resolve that receipt without source reads.
5. Expired, definitely uncommitted proof stays unavailable. Inspection can retain
   local context; an explicitly named new fixed read is a separate action.
6. Catalog, documents and pinned reviews reference the new origin explicitly.
   Corrections create new meaning revisions; original observation payloads stay fixed.

Future Playwright/Bombadil acceptance must exercise duplicate payload occurrences,
late observations across polls, limited batches, buffer eviction, disconnect and
reconnect, scope changes, expiry, lost acknowledgement/reload and store outages.
No UI screenshots are fabricated for these planned steps. Existing live-tail
storyboards remain the implemented interaction.

## Executed evidence

- Go unit checks cover duplicate occurrences, exact numeric tokens, per-poll clocks,
  late/limited flags, input/output isolation, stable resolution, wrong scope,
  duplicated/forged/partial references, read-anchored expiry, delayed issuance,
  restart, count/byte capacity and concurrent issuance/resolution.
- Go race detection passes for the membership bank.
- Hegel state/swarm executes the actual bank across issue/advance/resolve sequences,
  with TTL invariants checked after every action; existing tail multiplicity
  properties remain separate.
- Observed Quint replay executes issue → resolve members → expire → refuse through
  the real bank. Four controls reject wrong counts, forged availability, changed
  member windows and accepting expired proof. This is a membership model, not a
  model of an unimplemented durable publisher.
- Existing XTDB interruption, concurrent recovery and expired fixed-page recovery
  integration checks pass. Legacy origin/HTTP guards remain strict.

Actual commands/results and retained artifacts are in [VALIDATION.md](../VALIDATION.md).
