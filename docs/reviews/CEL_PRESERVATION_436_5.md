# 436.5 independent architecture, Go and UX review

The reviewed source and focused checks are accepted for the bounded saved-question
provenance contract. This is not acceptance of global revision history, migration,
authenticated attribution, physical assistive technology or whole-product 436.6
conformance.

## Ownership and meaning

- The existing `workbench.ExplorationView` owns optional per-signal record
  authoring. Its store identity, create/load validation and portable `ViewArtifact`
  YAML paths remain the publication and transport owners. No CEL artifact family,
  source executor, capture arm, proxy or second Save workflow is introduced.
- Layering is direct: `workbench` can call `querycel`, which already depends on
  ordinary `exploration` meaning. No `exploration` → `querycel` import cycle or
  generic validation callback API is needed.
- Source, representation, complete original base, checked binding, parent revision
  and lowering digest accompany the ordinary query. Every admission recompiles
  source with the pinned contract, validates the digest and compares the complete
  resulting typed question. Validation runs before composition/template early
  returns. Unsupported parent predicates remain intact.
- `RecordPlan.LoweringDigest` is computed by Go. Browsers validate shape and
  complete-parent conservation but do not imitate Go JSON hashing or claim to
  independently verify compiler lowering.
- Exact authored bytes and admitted meaning have separate identities. Formatting
  changes alter definition identity while unchanged typed lowering retains its
  digest. Revision/slot/epoch labels are provenance, not signatures: a legitimate
  newly labeled definition is allowed; modifying stored metadata under an old
  content ID refuses. Wrong library/compiler bindings refuse without migration.
- Metadata JSON is closed and rejects lossy Unicode repair. Repeated outer
  authoring objects follow last-value replacement rather than Go nested-map merge;
  case-folded authoring aliases cannot bypass admission. The portable YAML owner
  retains its existing bounded, strict conversion.

## Interaction and preservation

- Reopen restores exact text and its original base as supplied definition facts.
  It does not deserialize a private check/stage lease. A fresh edit/check obtains
  fresh invocation ownership. A restored ordinary question still requires its
  existing explicit Run to acquire telemetry.
- Native textarea display normalizes line endings. Preserved raw source remains
  distinct until an actual input event; merely checking a restored CRLF source
  does not rewrite it. Caret/undo behavior remains with the existing persistent
  textarea.
- Explicit discard relinquishes source ownership while preserving ordinary
  filters. New source then adopts the current typed question, rather than an old
  restored base. An authorized clean recall resets the prior representation and
  clears stale editor errors. Both unapplied text and valid but unsaved source
  retain a recoverable recall choice. Only the exact currently acknowledged
  saved query/source pair permits self-recall without discard; stale or foreign
  acknowledgements cannot bypass it. Authoring map insertion order is not meaning.
- Query/window/projection changes do not silently remove provenance to permit a
  save. Stale source blocks preservation until checked or explicitly discarded.
  The existing Preserve menu exposes source-only discard so the recovery is
  reachable even where the full expression editor is not mounted.
- Applying presentation keeps destination query/source ownership. Opening an
  original definition intentionally restores the donor question. Template
  expansion does not transfer the donor expression to a different destination.
  Existing captures and fresh-read/custody owners are not retrofitted.

## Independently executed checks

Managed Go 1.27.3, workspace module/build caches and CGO disabled:

```sh
go test ./workbench ./querycel -run 'TestCELAuthoring|TestRecordPlanning|TestPlanCEL' -count=1
```

Both packages pass (0.196s and 0.019s). These include actual stored/portable
round trips, source-versus-lowering identities, mismatched pair refusals,
strict Unicode/wire controls, stored-content tampering, duplicate-map replacement
and template destination authority.

```sh
node --experimental-test-module-mocks tests/cel-expression-editor.test.mjs
node --experimental-test-module-mocks tests/cel-provenance.test.mjs
```

Final direct runs pass **13 editor cases** (0.131s) and **3 provenance cases**
(0.055s), zero skips/failures. They cover lease retirement, exact restored CRLF
through Check, A→B→A invalidation, typed-default portable response normalization,
full source preservation and refusal of a receipt missing its lowering identity.
A further final run of seven focused file runners passes in **1.652s**:

```sh
node --experimental-test-module-mocks --test tests/cel-expression-editor.test.mjs tests/cel-preservation-integration.test.mjs tests/display-modules.test.mjs tests/held-presentation-396.test.mjs tests/preservation-owner-426.test.mjs tests/cel-authoring.test.mjs tests/cel-provenance.test.mjs
```

The integration file also passes all **5 direct production-callback cases**
(0.082s), independently exercising joined dirty-recall notifications, retired
callbacks, complete checked bounds, destination provenance, explicit discard and
valid-unsaved-source protection with matching acknowledgement exceptions. Its
acknowledgement control covers reversed signal-map insertion order. These are
mocked-DOM/controller checks, not browser or source integration proof.

## Visual and observed-driver review

Actual saved-definition and restored-source frames were inspected at 390px,
including the later 15:28 frames and dirty-recall state. The existing Preserve
menu and Save view action remain cohesive. A compact CEL source disclosure and
explicit source-only discard make preservation state clear. Restored comments and
exact unsigned literals are readable inside shared Filters; Keep current draft /
Open saved view and discard draft clearly separate recovery choices. There is no
new global toolbar, navigation or duplicate save action.

The new Bombadil actor was source-reviewed: its initial restoration witness reads
expected source independently from the stored definition; later actions deliberately
edit text, so that witness is not a continuous text-equality claim. Actual fetch
instrumentation and a positive plan-call requirement avoid accepting an unobserved
zero counter. Native line-ending display normalization is explicitly separate from
raw metadata identity. Scripted journey and parent validation own executed actor,
publication/import, dirty-recall, destination reuse and final frame results.

## Resolved review findings

Outer JSON duplicate/case behavior, normalized authoring-base defaults in portable
responses, old-base reuse after explicit discard, stale source being silently
pruned at Save, restored editor errors after recall, and CRLF normalization on a
check were corrected. No blocking source correctness finding remains in the
reviewed scope.
