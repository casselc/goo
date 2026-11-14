# 436.3 independent Go, architecture and ownership review

Accepted for the bounded representation and pure authoring-owner slice. No open
correctness findings remain in the reviewed production code. This does not claim
a mounted editor, browser usability acceptance or the whole-product 436 journey.

## Contracts reviewed

- `querycel.PrintPredicate` admits the complete ordinary predicate before recursive
  printing. It shares the compiler's fixed declarations, quotes literals, then
  checks and lowers the generated source through the actual CEL compiler. Exact
  typed predicate JSON must agree before any text is returned. No unsupported
  sibling can disappear behind a printable leaf.
- That comparison preserves order and nesting. Singleton and non-round-tripping
  n-ary shapes refuse instead of claiming Boolean equivalence. Fixed supported
  fields exclude numeric bindings, presence and map expressions; the resulting
  omission rules agree with the existing `QuestionSession.identity` contract.
  Noncanonical uint spelling such as `01` refuses rather than silently changing
  an identity-bearing literal. Size, cycle and ordinary depth/node checks remain
  bounded. Refusals are typed and return no partial text.
- `AuthoringSlot` owns an immutable original base separately from execution state.
  The supplied server content hash remains opaque; it is not substituted with
  browser normalized question identity. Local parent revision, slot, mount epoch
  and monotonically increasing exact string draft revision remain distinct.
- Each check has a private invocation token and detached originating request.
  Success checks the full compiler binding, version, source and ordinary complete
  base-AND-predicate identity. Failure correlates through its originating token;
  diagnostic HTTP envelopes do not need to echo a binding. Repeated checks,
  text/parent A→B→A, hide/restore and disposal retire old callbacks.
- Staging requires the current checked token. Last valid meaning can remain
  inspectable but cannot stage for an invalid or changed draft. Edits repeatedly
  target the original base, so they replace one owned additional condition
  without growing nested ANDs. Explicit clear returns a detached parent. Mode
  changes preserve invalid text; overwriting an existing draft through visual
  adoption requires a deliberate discard choice.
- The module owns no DOM, timers, requests, source execution or held evidence.
  It does not merge `QuestionSession` execution generations or introduce a global
  coordinator. Existing Apply/Run and evidence owners retain their roles.

## Independently executed checks

Managed Go toolchain, focused final printer tests:

```sh
go test ./querycel -run TestPrintPredicate -count=1
```

Pass: 0.153 seconds. Includes fixed-field round trips, Unicode/control/quoted
literals, max uint64, canonical-text idempotence, supported same-kind nesting,
unsupported sibling/singleton/n-ary/presence/map/numeric/cyclic/oversize controls.

```sh
node --test tests/cel-authoring.test.mjs
```

Pass: one file runner containing nine explicit cases, zero failures/skips,
0.320 seconds. Cases cover detached inputs/results, repeated replacement, stale
success/error ABA, invalid text/mode/adoption, hide/parent/dispose direct callbacks,
receipt tampering, originating-request mutation and bounded well-formed tokens.

The parent records generated native checks of the actual browser identity and
broader suites separately. This report does not claim a fresh source read,
ClickHouse parity run, browser screenshot, physical assistive-technology test,
cross-engine run or OpenAPI regeneration. No visible editor is mounted here.

## Follow-on integration consideration

The 436.4 adapter should report malformed UTF-16 source locally before transport,
while preserving the draft. Current exact-source receipt checks prevent accepting
JSON replacement text, but those checks alone are not useful user feedback.

## Observed ownership model review

The added `cel_authoring436.qnt` projection and its Go/Node adapter are consistent
with the stated bounded claim. The script uses the real `AuthoringSlot` and sends
its actual requests to Go `PlanRecord`, receiving actual receipts/diagnostics.
Twenty-six events project method return values, stage availability and snapshots;
these are observations, not prefilled expected state. The trace exercises text
ABA, an old stage callback, invalid text, hidden replies, parent retirement,
clear and disposal.

The model conserves readiness, pending/active/alive state, last-valid presence,
acceptance and the local effect counter. It does not independently model binding
contents, text semantics or token identities, and is not an exhaustive concurrent
browser campaign. Those contracts have separate compiler/owner tests. Six
negative controls forge acceptance after ABA, an old stage callback, readiness
with invalid text, hidden acceptance, disposed failure acceptance and a final
local read count; their selected indices match the actual script actions.

The observed read counter intercepts `globalThis.fetch`; source review establishes
that the pure owner has no other I/O dependencies. This is not a measurement of
arbitrary process I/O, browser requests or ClickHouse reads. Execution results for
the native replay are recorded by the parent validation run; this section records
independent adapter/model/negative-control source review.
