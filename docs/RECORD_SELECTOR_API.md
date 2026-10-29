# Checked record selector APIs — 250

Logs and spans without an environment label can now use the main **record APIs**,
not only the separate adoption preview. Fixed reads, rolling tails, context
constructors, population/comparison/explanation planning, server-held results,
filesystem captures, exact reopen/refresh and live comparison/freeze preserve the
explicit selection policy. This is the server portion of A0, not full UI adoption.

```json
{
  "signal": "logs",
  "environmentSelector": {
    "version": "goo/environment-selector/v1",
    "policy": "effective",
    "scope": {"mode": "unreported"}
  },
  "fromMs": 1793275200000,
  "toMs": 1793275260000,
  "limit": 100,
  "offset": 0
}
```

Submit a fixed question to `POST /api/explore`, or as
`question: {kind: "explore", explore: QUESTION}` to `POST /api/captures` with a
title. The latter performs a fresh bounded source read; it does not save an
already displayed page. Exact `GET /api/captures/{id}` reads retained data.
`POST /api/tail` accepts signal, environmentSelector and optional where, without
fixed time bounds; each poll uses the existing rolling 60-second event window.
The [wire contract](ENVIRONMENT_SELECTOR_PROTOCOL.md) defines policies and modes.

## Identity and selection

Choose a nonempty legacy `environment` name **or** `environmentSelector`. With
the new field the legacy name must be empty or omitted. The Go output still
serializes `environment: ""`; that value alone is invalid, never a wildcard.
Unknown versions, invalid/ambiguous combinations and malformed DTOs refuse before
source access. Existing legacy inputs, canonical bytes, plans and stored IDs stay
current-key-only and unchanged. No old saved question is reinterpreted.

Effective selection distinguishes unreported (neither key), present-empty and
named values. Current-key presence wins even when empty or conflicting; legacy
is used only when current is absent. An explicit `current-only` named selector
means the same population as its legacy name. Semantic compatibility accepts
that equivalence, while immutable identities retain the actual wire choice.
Effective named `prod` remains different from current-only `prod`.

New-selector rows are checked against signal, half-open event window and raw
resource map presence/values, including overflow rows before truncation. Missing
or null resource maps/values do not become unreported data. The same checks apply
to result snapshots, tail membership issuance and stored capture verification.
Counts/groups remain independent full source reads; this does not establish
transactional snapshots, ingestion completeness, unique requests or healthy data.
Existing source/row/byte/time and capability lifetime limits still apply.

Derived context/followup questions copy the full selector, including policy.
Comparison, explanation and companion checks use checked semantic equality.
Returned selector DTOs are detached where queries, normalized captures, scopes and
capabilities cross ownership boundaries. Membership resolution binds the full
serialized selector; a token for effective named data cannot be reused under a
legacy name. Snapshots retain their own exact question rather than caller scope.

## Captures and catalog

Filesystem capture serialization includes the selector; capture refresh cannot
change it. Exact reopen and explicit catalog rebuild verify selected record
membership. The catalog adds optional `environmentSelector`; its legacy
`environment` is null for these entries. Null alone still means unknown metadata,
not unreported. Policy/scope are explicit in summaries, including comparison
members. Existing named environment filters match legacy summaries only; use
unfiltered/title searches to discover selector captures in this cut.

The v2 metadata extension preserves old entry bytes/digests and supports mixed
catalog pagination/rebuild without rewriting capture payloads. A pre-250 reader
may reject the **whole catalog** after selector captures are added; it will also
reject selector-bearing exact captures. Rollback requires a compatible reader;
do not delete or downgrade immutable evidence to make an old reader accept it.
No index migration runs automatically.

## Explicit unsupported boundaries

| Surface | Current behavior |
|---|---|
| Browser Activity/fixed-question authoring | Find environments remains local preview. No selector handoff, live selector control or question link is enabled. |
| Browser selector-capture inspection | Catalog/linked facts label policy/scope. Opening refuses before mounting legacy investigation owners and offers the exact captured JSON read link. Existing legacy capture UI continues working. |
| Saved Activity/exploration views, authored steps and reusable package generation | Server validation refuses selector-bearing intent before publication. Known selector-capture baselines are refused by view HTTP owners. |
| Native/metric discovery, service/inventory/dashboard/pod/trace-search request families | Existing named-only schemas remain; passing the new field is unsupported, and an empty name is not accepted. |
| Reported parent-edge queries | New-selector queries refuse before compilation/source access; parent ownership SQL has not been migrated. Exact trace/span record followups retain the selector and remain supported through ordinary record queries. |
| XTDB result/tail retention | Explicit refusal before publication; current named entity/catalog/annotation contracts are not reinterpreted. Exact filesystem capture references remain immutable IDs. |

These are server gates where applicable, not hidden UI controls as the only
protection. A valid standalone DTO does not grant support to another family.

```mermaid
flowchart LR
  SDK[Standard OTel logs / spans] --> CH[ClickStack exporter tables]
  Select[Legacy name OR versioned selector] --> Query[Checked record Query / Tail]
  Query --> CH
  CH --> Rows[Bounded reads / selected-row verification]
  Rows --> Hold[Exact result / tail membership capabilities]
  Rows --> Capture[Filesystem capture / live freeze]
  Capture --> Reopen[Verified exact reopen / policy-explicit catalog]
  Query --> Gate[Unsupported family / persistence gates]
  Reopen -. browser adoption pending .-> UI[Activity / fixed / Saved owners]
```

Next: adopt the shared selector protocol throughout browser applied/draft state,
observation-to-fixed handoffs, routes, return links, record/context and saved
owners together. Only then remove the browser gates. Service/native/package/
retention scope adoption remains separate capability work.

The expanded [journey 247](../journeys/247-environment-adoption-storyboard.md),
Go properties and bounded observed Quint traces exercise the implemented cut.
[VALIDATION](../VALIDATION.md) records actual runs and limits.

## Browser constructor foundation — 251

The pure record-environment module now provides `selection(question)` for checked
semantic scope, `questionFields(question)` for detached representation-preserving
copying, and `sameQuestionEnvironment(a, b)` for compatibility. Legacy names map
to current-only policy for comparison; this never rewrites stored wire identity.
Missing/empty names without a selector and ambiguous combinations refuse.
Trace/span-log, surrounding-log, retained-tail inspection and tail preparation
use the shared copy boundary. Tail preparation validates before changing owner
state or clearing a retained buffer.

```mermaid
flowchart LR
  Q[Legacy name or explicit selector] --> Check[Checked browser selection]
  Check --> Copy[Detached original environment fields]
  Copy --> Trace[Trace and span logs]
  Copy --> Context[Surrounding logs]
  Copy --> Tail[Tail and original-window inspection]
  Check --> Equal[Policy-aware semantic equality]
```

This is a constructor foundation, not enabled browser selector support. Existing
Activity/session/link and Saved gates remain. Next migrate Activity builders,
owner identity/invalidation and fixed/Saved handling together. In particular,
SSE reset checks must reconcile server omission defaults (`null` selector and
explicit empty scope values) without changing immutable question identity.
Unsupported saved intent, native/service and XTDB paths stay gated.
