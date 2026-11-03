# Typed service environment scope — architecture 318

Status: architecture source review accepted; executed API/browser/property checks
are tracked in `VALIDATION.md`. Deliver typed record-family Services support rather than
another contract-only round or a partial environment-name fallback.

## Useful outcome

An Activity observation selected under effective environment policy can stage and
explicitly read its service overview, customize a record-family view and return to
the same paused Activity buffer/window. Preserve exact namespace reporting, service,
fixed occurrence window and typed selection through every record request. The same
checked scope model should support effective named, empty and unreported modes.
Current-only remains named-only under the existing identity contract; do not invent
current-only empty/unreported variants.

A reported current environment key takes precedence even when present empty; do
not OR current/legacy values together. Legacy `environment: name` requests retain
current-key-only meaning. They must not be reinterpreted as effective policy merely
because the new field exists. The 316 observed-service intent retains original
source context and its explicit omitted-Activity-filters semantics.

## Concrete boundary map and implementation

1. Extend `exploration.InventoryQuestion` with optional typed selector, retaining
   the existing legacy field. Reuse Query environment selection/validation so both
   representation exclusivity and policy semantics have one owner. Inventory and
   service discovery compile the actual checked selection, not a derived string.
   Deep-copy both namespace and new selector in returned question/plan metadata.
2. Preserve that scope in `DashboardRecordQuestion` and record-family planning,
   including HTTP/DB/RPC/messaging observed-profile checks. UI record constructors
   use `RecordEnvironment.questionFields`; exact scope comparison uses typed
   equality; row-based profile support uses `matches(selection, resource)` rather
   than current-key string equality. No protocol profile is guessed from the
   environment, and actual service/namespace requirements remain.
3. Service discovery may count raw metric data-point presence under the same typed
   ResourceAttributes selection across metric tables. That is observed record
   presence, not support for native metric computation, unique requests or health.
   Keep this distinction visible in overview and source facts.
4. Broaden the handoff adapter to carry supported selectors intact into the new
   record scope. Keep stage zero-IO, explicit Read, source occurrence lease,
   cancellation and named paused return. Do not add another executor/navigation
   owner or use a name-string sentinel for unsupported scope.

## Explicit native and entity boundaries

Native instrument discovery/evaluation, deployment context, service annotations,
Activity-view library bindings and declared-operation hints currently use named
current-key environment contracts. Do not broaden them incidentally. A small
`current-only named` capability extraction may produce their existing environment
string only after exact policy proof, including typed current-only named input.
An effective named selector with the same spelling is not equivalent.

For unsupported policies, omit those requests and render explicit unavailable
subpanels/reasons while the record service overview remains usable. Do not send
empty instrument lists and then label the service as no metrics/no native data.
Record-level raw metric presence and unavailable native evaluator can coexist.
Do not publish annotation/entity keys under a guessed environment or silently
suppress the entire overview because one optional family is unsupported.

## Admission, certainty and compatibility

Go/browser/OpenAPI must admit the same request representations. Existing legacy
queries/SQL retain their meaning. Typed source scope, capture/catalog questions and
Customize/save/export must not lose selector policy on intermediate serialization.
Generated record views use existing validators and artifact/catalog owners; no
compatibility shim or second query AST is needed.

The new selector pointer introduces the same mutation risk corrected in 312:
returned inventory/discovery/dashboard scopes must be detached from caller input.
Namespace missing, present empty and literal value remain distinct. Do not infer
missing environment from a missing key in an unavailable resource-map payload.
Observed/no-data/changed and independent-read guarantees remain the source adapter's
existing semantics, not a new completeness statement.

## Required finite acceptance

Run actual effective-named legacy-SDK Activity → staged service → explicit Read →
record overview/Customize → paused return at desktop and narrow width. Assert exact
policy/window/namespace on all actual record and raw-presence requests, source-free
stage/return, current-present-empty shadowing legacy value, and visible zero-request
native/entity refusal. Exercise effective empty/unreported with fixed source
fixtures or native journeys where available; no unsupported mode may silently
fall back. Include legacy current-only plan parity and typed current-only bindings
for existing native paths.

Add generated Go/browser field-policy/namespace cases, strict request/schema
counterexamples, returned-scope immutability and stale/canceled handoff actions.
This milestone need not migrate native evaluators or all entity identity domains.
Finish the supported record task and make the remaining boundary explicit.

## Initial source review

The current implementation adds a closed environment selector union to inventory
scope, compiles inventory/discovery from actual checked selection, preserves
selector/namespace clones and carries typed record dashboard questions. Raw metric
presence uses the same policy predicate. A distinct native-capability boundary
rejects incompatible supplied instruments/selections and prevents scalar declaration
matches. UI current-name extraction admits only provably current-only named scope;
effective native/deployment/annotation/library calls are omitted with reasons.

The browser record constructors, profile resource witnesses and returned scope
checks use shared typed helpers. Manual inventory rediscovery retains private
checked environment fields until genuine input editing, avoiding conversion of a
policy display label into a literal name. Handoff now preserves its original scope
representation and existing stage/return ownership.

Review identified two remaining presentation-only boundary gaps: ResourceReturn's
origin environment string fallback reports typed scope unavailable, and the trail's
generated visit title uses `q.environment` directly. Both should use the shared
checked policy label for record scope while preserving native definitions and
literal authored titles. These do not justify broadening entity/native contracts.


## Final boundary review

The two policy-label gaps are corrected through shared checked labels: resource
origin context handles typed record scope without changing native definition
meaning, and generated trail labels no longer interpolate an absent legacy name.
Authored titles remain untouched. Manual rediscovery also retains its checked
private selection until actual environment input editing.

Final source review accepts typed record scope/admission/SQL, exact scalar-capability
gates, native unavailable representation and selector detachment. No guessed name,
empty-global declaration match or effective-to-current fallback was found in the
inspected continuations. Source review has no remaining architecture blockers;
`git diff --check` passed. Actual all-mode source, property and native journey
results remain verification evidence rather than inferred here.

Targeted inventory-capture follow-up is accepted. `CaptureQuestion.Normalized`
uses the new scope clone, detaching namespace and selector. Common Scope derives
inventory through its typed Query and copies the selector. Catalog summaries mark
legacy environment known only when no selector exists, otherwise retaining the
exact typed selector; effective scopes do not become empty legacy names. Stored
inventory checks compare the complete question and refresh through typed
CompileInventory. No policy identity drop was found in this newly admitted path;
this is not a reopening of unrelated historical capture behavior.
