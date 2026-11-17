# Preservation component: understandability review

Reviewed 2026-11-16 at the 462 baseline. This is a source review and current-code
walkthrough, not an implementation change, storage migration or live XTDB audit.
[Walkthrough](../../walkthroughs/preservation.md) maps the afternoon reading path.

## Assessment

The product model is sound and explainable: save configuration, keep server-held
observations, explicitly capture a new result, and attach authored interpretation
without rewriting evidence. Understanding the implementation takes longer than
it should because those four verbs cross several store/receipt types and dense
UI owners. The answer is clearer boundaries and readable code, not one universal
“save” operation. Conflating these paths would remove correctness.

## Findings and bounded proposals

| Priority | Evidence in current code | Simplification and acceptance |
| --- | --- | --- |
| High | [captures.mjs](../../../ui/captures.mjs) `mount` holds visit generations, abort controllers, layout writes, stored paging, native context, kept view return and catalog state. [documents.mjs](../../../ui/documents.mjs) combines editor, publication recovery, cards, capture attachment and recall. Source lines reach roughly 5,000 characters. | Format these owners into ordinary readable statements first, with no behavior change. Then extract named visit/publication state transitions only where a state table can replace interacting flags. Keep DOM/focus in the owner. Acceptance: a reader can identify state, entry, retirement and side-effect boundaries without unpacking a giant closure; existing request-count, stale-action, hidden-return and uncertain-save tests remain unchanged. File-count reduction alone is not the goal. |
| High | Capture-family shape, normalization and stored-delivery rules are distributed across [capture_store.go](../../../exploration/capture_store.go) `CaptureQuestion.Validate/Normalized`, `Create`, `validateCapture`, `Page`; [capture_envelope.go](../../../exploration/capture_envelope.go); and [saved-capture-protocol.mjs](../../../ui/saved-capture-protocol.mjs). `Create` repeats family dispatch in parent-plan checking and execution. | Write one compact carrier matrix for each family: source read versus held promotion, row-bank versus aggregate, paging, original-question equality and refresh eligibility. Use it to isolate family-specific execution/admission functions and remove repeated switches where it helps. Retain a closed typed union; no dynamic registry or general storage framework is justified. Acceptance: adding a family has an obvious small checklist, and cross-language rejecting fixtures cover mixed payloads, changed questions, cursor mismatch and unknown outcomes. |
| Medium | Local stores are `workbench.Store`, `exploration.CaptureStore`, `KeptObservationStore`, and operation `investigations.Store`; optional `retention.Store` adds XTDB. Naming alone does not tell a new reader whether “investigation” means a narrative or an operation capture. `serve_http.go` actively registers the operation path. | Rename/package-document the operation store around operation evidence and show the five stores in one architecture map. Assess whether the operation envelope can become a typed capture family only after deciding its product use; do not delete it as dead code or force a common payload to save a directory. Acceptance: all registered routes have one named owning store and each stored type says whether it contains question, observation, assertion or reference. |
| Medium | [workbench/meaning.go](../../../workbench/meaning.go) `Meanings` loads the full bounded meanings catalog and builds parent/head sets before filtering. [activity_views.go](../../../workbench/activity_views.go) `ActivityViewsForEnvironment` loads definitions until it finds matching page items. `Store.ids` permits 10,000 entries, with each workbench artifact up to 128 KiB. | Measure filtered catalog reads with realistic revision history. If they are material, reuse the existing derived-index pattern, not an in-memory “latest head” authority. Keep artifact files authoritative and direct-ID recall available during index repair. Acceptance: read/byte budgets are explicit, branches and withdrawal semantics are preserved, and indexed results match a full-scan oracle. This is a scaling opportunity, not a measured production slowdown. |
| Medium | [investigations/store.go](../../../investigations/store.go) `Load` uses `json.Unmarshal`; `workbench.Store.read`, generic captures and kept resources use strict one-document decoding. [immutablefile.Publish](../../../internal/immutablefile/publish.go) is already the shared publication primitive, with different collision/durability choices deliberately exposed. | Standardize the small bounded JSON-reading primitive and explicit error categories where callers truly share a contract; make the operation envelope reject unsupported fields consistently. Keep family validation and content-digest rules local. Do not build a generic repository abstraction or erase `PublishNew` versus idempotent winner semantics. Acceptance: corruption, unknown fields, oversized input and trailing JSON tests describe a single intentional policy per envelope. |
| Medium | Publication/retry policies differ intentionally: keeps bind `RequestID` to intent, capture-review `*Once` reserves a stable intent clock, XTDB uses transaction assertions and await tokens, while [investigation-draft.mjs](../../../ui/investigation-draft.mjs) journals unknown publication and [retention-client.mjs](../../../ui/retention-client.mjs) tracks pending exact intents. | Put a short publication-state table beside these APIs: before-send failure, rejected write, acknowledged receipt and uncertain response. Reuse small transport/error helpers only after that table is explicit. A universal retry wrapper would currently hide different guarantees. Acceptance: retries never substitute a fresh source read or a different selection, and an uncertain response never silently becomes “not saved.” |
| Medium | [retention/store.go](../../../retention/store.go) `Publish` connects via `pgx`, checks request intent, creates schema tables and commits evidence members plus annotation; `PublishTail` and `CorrectOnce` have related transaction paths. The feature is live but optional and rejects profiles other surfaces support. | Keep the experiment behind its existing explicit adapter and capability statement. Decide from actual annotation/training needs whether to graduate it or remove it, rather than expanding XTDB to cover every local format. If graduated, move schema setup to an explicit startup/migration boundary and measure connection/transaction costs before choosing pooling. Acceptance: local Save/Keep/Capture still works without XTDB; unavailable/unsupported profiles remain explicit; instance-backed tests prove retry and read-basis behavior. |

## Work that should stay separate

- Local draft recovery, immutable artifact identity and XTDB await tokens solve
  different problems. Replacing them with one generic version number would make
  failures harder to understand.
- Training-review decisions in [capture_review.go](../../../workbench/capture_review.go)
  are not ordinary note tags: they resolve exact captured-record references and
  enforce task/label/supersession rules before bounded export. Keep that domain
  explicit if the product retains training workflows.
- Existing capture metadata indexes and kept publication gates are deliberate
  bounded-work mechanisms. Do not remove them as “evidence machinery” solely
  because the user does not want accumulating diagnostic artifacts.

## Dead/unnecessary work conclusion

No production path in this component is proven unused by this review. In
particular, the operation capture store is registered in
[serve_http.go](../../../cmd/goo/serve_http.go), and optional retention is exposed
through explicit configured-status and refusal routes. Suspected product overlap
requires route/usage and capability decisions before removal. Historical milestone
notes are not evidence that their implementations are dead.

The first recommended change is readability of the two UI owners and the explicit
storage/carrier map. The next is a measured catalog-cost check. Avoid expanding
publication evidence or inventing a new persistence platform merely to complete
this review.

## Checks

The review followed active registrations, caller imports, family validators,
publication primitives, catalog loops and browser recovery state. Four focused Node test files pass, zero skipped:
`preservation.test.js`, `investigation-draft.test.js`,
`saved-capture-protocol.test.js` and `retention-client.test.js`.
`git diff --check` passes. No Go suite or live XTDB integration was run for this
documentation-only review. No deployment,
live source query or live XTDB instance was needed to establish these findings.
