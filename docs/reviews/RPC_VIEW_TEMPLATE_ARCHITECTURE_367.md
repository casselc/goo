# Current RPC view templates — architecture 367

Status: final integrated source and genuine native/actor/observed acceptance
accepted. No source blocker found.

The existing configuration-only expansion boundary now recognizes shipped
`otel.rpc-server-current-methods` and authored binding
`otel.rpc-server-current`. Its dimension widget family is `rpc-server-current`,
grain `server-span-record`, with `rpc.method` grouping. Admission requires an exact
nonempty service, Server kind and nonempty current `rpc.system.name` predicate.
Method presence is not admission: missing, empty and `_OTHER` remain literal
populations. Legacy RPC is a distinct observed profile and is not an alias arm.

The Go descriptor and UI profileRules agree. UI eligibility is a presentation
hint; server expansion remains authoritative. The existing authored destination
binding retains whole checked policy, namespace, predicates, fixed window and
chosen destination grouping; donor presentation does not supply donor scope or
observations. Existing page-zero normalization and specialized/native refusal
boundaries remain unchanged. Shipped specialized behavior was not relaxed.

The API enum/reference/profile declarations agree with actual output. Root HTTP
tests exercise shipped expansion, immutable save/reopen and reuse into another
typed policy, namespace, service, predicate and grouping, with wrong-role/current-
key refusal and no telemetry reads. Current-profile CSS now uses the same compact
service layout and comparable fixed-width tracks as the other role profiles;
it changes no data or query authority.

Focused workbench tests passed in 0.033 seconds, including optional-method
admission, whole destination identity, detachment, save/load and mandatory-role
refusals alongside existing HTTP/database/authored regressions. Root reports full
Go, 1,256 UI checks, API validation and generated/observed-wrapper checks passing.
Runtime acceptance is intentionally not inferred from these source/code gates.

## Genuine-entry correction reviewed

The first native entry exposed a pre-existing record-query echo mismatch:
expected typed questions omit the scalar environment while Go echoes its empty
zero value. A strict object comparison withheld the positive RPC witness despite
returned Server rows. The bounded correction in `service-dashboard.mjs` checks
both environment selections and uses the existing record question identity only
for the record branch. This normalizes accepted Go defaults while retaining the
complete policy/window/predicate/profile identity. Conflicting nonempty scalar
plus selector refuses before normalization. Native measurement comparison keeps
its prior strict identity boundary. Source accepted; root reports the added
four-policy and genuine-scope-change regression checks passing within 1,257 UI
checks. This does not broaden the RPC predicate or treat key metadata as a witness.


Final execution reported by root: native 60.521s, required observed model
replay 4.85s with zero skips, and 602 actor snapshots across ten phases with
positive visible service Return and actual List staging. Exact target scope, six raw-oracle IDs/210ms,
role refusals and stale owner controls all pass. Acquisition during discovery and
service loading remains explicit and separate from metadata-only reuse. See
[validation](../../VALIDATION.md) and [storyboard](../journeys/367-current-rpc-template.md).
