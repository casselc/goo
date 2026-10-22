# S2/A3 acceptance reconciliation

This audit reconciles the active queue with delivered work as of 2026-10-22.
The bounded A3 record/one-service cut and S2 fixed-editor mount gate are delivered. The authoritative order remains [the active queue](IMPLEMENTATION_PLAN.md).
Comparison work belongs to D1; it is not the immediate successor to these cuts.

| Acceptance | Implementation and evidence | Disposition |
| --- | --- | --- |
| Stage/apply typed live scope; one active owner; no hidden inspection reads | `stream-entry.js` owns draft/applied scope and subscription transitions. Journeys 103/104 and the Live attribute filters validation section cover real signal changes, unavailable/gap states, obsolete results and explicit fixed staging. | Delivered bounded S2. |
| Held ancestry, honest missing/ambiguous/cycle/depth states | `record-presentation.js` resolves only supplied same-trace parents. Journey 106 and 200 Hegel supplied-parent cases cover cross-trace independence and non-mutation. | Delivered; wider trace context remains an explicit read. |
| Return to paused parent with anchor/focus, no rerun | Activity owns its tab-local return anchor; journey 105 and the paused-return validation section cover cancellation, scroll/focus and explicit resume gaps. | Delivered; reload/cross-session resume is outside this cut. |
| Two real surfaces share facts without flattening source types | Live/fixed/Saved records and held Services rows share the record/table presenter; journeys 106/59 cover source-free inspection. Independent service overview reads keep their own scope and basis. | Delivered A3 one-service adoption. |
| Record receipt adapter in captures and documents | `GooEvidenceScope.record` supplies count/window/source/custody facts to both, with unknown count/page disagreement. Journeys 32/41/45/50 and their validation entries cover fixed continuation and source-free reopen. | Delivered A3 receipt adoption. |
| Remove duplicate editor/details after parity | `saved-records.js` replaces the hidden query editor; `record-results.js` shares fixed/Saved context. Journey 107 and the single-inspector validation campaigns cover paging/context, delayed response rejection, one inspector and hidden unusable comparison controls. | Delivered; typed fixed query authoring is still a real consumer, not an unused compatibility editor. |
| Remove eager fixed editor under Activity | Trail continuation registration is independent of editor construction. Journey 109 proves source-free explicit mount and same draft/DOM restoration; 102/105 check the boundary with real stream/fixed reads. | Delivered S2 mount gate; bounded B1 consolidation is also delivered. |
| Contextual lookback | Navigation hides the default on Saved and labels new discovery/operation/example reads and the Definitions HTTP preview. Journey 107 checks route labels, unchanged default and no added source reads. Fixed questions retain their own bounds. | Delivered in `9e80f9c`. |
| Mobile, passive component lifecycle and formal evidence | Journey 106 covers SSR, detached-property/reconnect and no component API reads. Subsequent service/Saved/inspector journeys cover narrow layouts. Existing Hegel state/swarm and Quint observed ledgers with rejection controls protect ownership boundaries. | Evidence is cumulative; this audit does not claim every historical campaign was rerun. |

## Remaining work has an owner

- **B1, bounded consolidation delivered:** [Definitions, fixed reusable questions,
  document revisions and compatible retained metadata rows](DEFINITION_LIBRARY.md)
  share passive rows/inspectors. Retained snapshots preserve current-head tag clocks;
  pinned review citations, original retry intents and read bounds retain their own
  owners. Client/basis, catalog, citation and explicit recovery are separated.
- **B2, active:** the first [record-field presentation and Go/JS wire fixtures](FIELD_PRESENTATION.md) are delivered; broader qualified identity/protocol work remains. Request-owner extraction requires matching adapter lifetimes.
- **D1:** remaining aggregate/native comparison compositions and fitting inspector
  adoption. The recent native presenter is useful groundwork, not completion of D1.
- **E1:** broader native instrument discovery and explicit convention capabilities.
- **S3:** exact tail custody stays blocked on server-proven membership/expiry/retry
  contracts. Existing live-page retention is not a substitute for tail membership.

The `entry-advanced` disclosure remains an explicit authoring entrance. The trail
registers continuation handlers without constructing the fixed editor on fresh
Activity. Opening authoring or a supported fixed continuation mounts it; returning
to Activity suspends/detaches the retained editor, and reopening restores the same
draft. Old record question/step/view links still restore without execution. Initial
Services/Definitions/Saved routes also avoid constructing an unused editor. Native
measurement links retain their own owner. The first legacy service/package/trace
continuation is delivered once after its editor is mounted.

Journey 109 checks absent editor DOM on fresh Activity, source-free explicit mount,
retained draft/DOM identity and inert deep-link restoration. Journeys 102/105 also
check zero eager editors and exactly one on fixed continuation; existing real-read
and rejecting Quint ledgers protect paused return and execution boundaries. This
closes the eager-mount removal gate, not every future control/layout recommendation.

```mermaid
flowchart LR
  S2[Delivered: live scope / inspect / paused return] --> Gate[Delivered: explicit fixed editor mount]
  A3[Delivered: shared service records / receipt facts] --> Gate
  Gate --> B1[Delivered: selected revision libraries / retained snapshots]
  B1 --> B2[Protocol and field consolidation]
  B2 --> Later[Later queue: C / D / E]
  Capability[Server-proven tail membership] -. required .-> S3[Exact tail custody]
```

See [actual validation](../VALIDATION.md), [record presentation](RECORD_PRESENTATION.md),
[saved record ownership](SAVED_RECORD_PRESENTATION.md) and [Live Activity](LIVE_ACTIVITY.md).
