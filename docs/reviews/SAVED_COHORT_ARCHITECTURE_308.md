# Saved fixed cohort questions — architecture 308

Status: bounded contract; implementation review pending. The concrete task is to
save/share the useful selected-versus-rest question discovered in 307, without
turning a held observation into a fresh query or adding another comparison owner.

## Representation and admission

Extend the existing single comparison child with a pinned fixed-query count kind,
for example `query-count-comparison`, owning one explicit `baselineQuery`. The
current view root remains the selected question. The captured-count variant keeps
its exact capture-ID contract; each declaration must choose one variant and refuse
unused/conflicting fields. One comparison role per composition remains unchanged.

Both queries must satisfy the existing ordinary single-group count profile and
count-cohort compatibility, with exact typed environment, fixed bounds, grouping
and remaining profile directives, and page zero for both roots. Predicates may
differ. This is a deliberate bounded query-plan operand in the existing role;
update the formerly query-free composition description rather than leaving a
contradictory public contract. Do not admit composition plus legacy flat breakdown
baselines or create generic dataset/recipe roles.

Manual or reopened fixed queries do not establish logical complement. Use generic
Baseline query/Selected query labels. The UI-generated 307 remainder is a known
construction at that moment, but saving arbitrary predicates is not a persistent
proof that the sides partition another population. Independent reads also do not
establish a simultaneous complete upstream snapshot.

## Explicit transition and execution

Offer **Include comparison in view** only for the runtime remainder mode. It
validates and copies fixed query intent into the existing comparison declaration,
without reading, publishing, capturing or copying supplied cohorts. Then existing
Save/export persist the definition. The held-parent mode stays runtime-only; no
silent fresh-query conversion, implicit capture or new evidence wrapper is allowed.

Reopen/import stages the saved view. Explicit Run acquires selected and baseline
through existing full exploration and comparison admission, with independent read
clocks and existing changed/no-data states. Use the current controller/serial and
count-pending owner; query-kind switching must retire incompatible transient or
captured comparison state. Existing immutable captured baseline behavior remains
separate and unchanged.

## Reuse, transport and ownership

Portable artifact admission allows the fixed-query variant while continuing to
reject captured references. Strict YAML/JSON admission, normal Save/store identity
and catalog Open remain the same owners. No rows, snapshots, result tokens, source
basis or stored capture receipts are added to the declaration.

Refuse ordinary presentation Apply and category-template expansion for the fixed
query-pair variant in this first cut. Copying a source baseline into a destination
would quietly import its predicate/environment/window. Even a same-looking target
must not imply a rebinding contract. A future explicit transform/input can define
cross-scope reuse; this milestone provides durable original-question Open/import.
Ensure Go, browser checks and OpenAPI agree on that rejection.

## Acceptance and source review targets

Use 307's real authored group selection → Compare with rest → explicit Run →
Include comparison in view → Save/export → import/reopen → Run. Prove local
inclusion reads nothing, the persisted pair contains intent only, both fresh scopes
retain exact filters/policy/bounds/grouping and read clocks, and the selected
records kind/options survive. Confirm generic labels after source editing/reopen,
held-parent absence, captured-ID refusal in transport, ordinary Apply/template
refusal, malformed/mixed variants, stale Include controls and read failure.

Review validators, discriminated strict decoders, deep-copy/reuse code and schema,
not just rendering. The older captured declaration has shallow-copy paths because
it formerly contained no query pointer; new baseline query state must not alias a
caller or slip into captured template/reuse branches. Maintain one comparison
execution owner and current Back/pending-record contracts.

## Final implementation source review

Go/browser/schema use the explicit `query-count-comparison`/`baselineQuery` arm
of the existing count-comparison declaration. Strict wire admission refuses mixed
arms including null/empty supplied fields. Query validation enforces fixed shared
bounds, page zero and ordinary count context. The renamed count declaration and
composition descriptions reflect the operand union. Query-pair reuse and templates
are refused even for identical destinations or hidden children; portable codecs
allow fixed intent while rejecting captured references.

The runtime remainder action explicitly includes the comparison, drops category
template binding with a message and stages Run after clearing transient results.
The existing count executor/controller reads the fixed baseline independently and
uses normal comparison admission/components. Labels identify fixed baseline and
selected questions without a persisted complement assertion. No held-parent
conversion or captured evidence transfer occurs.

Source review found root-only scope controls could invalidate a fixed pair. These
are now disabled/omitted with an explanation to edit both questions together;
handlers guard the same policy. Execute validates the proposed fixed context before
mutation or any source request, and baseline preparation validates before reading.
This prevents stale scope controls from silently leaving an invalid pair. Current
artifact normalization/preview includes the baseline question, retaining full AST
matching and explicit independent read clocks.

Final architecture source review has no remaining blockers. `git diff --check`
passed. Actual browser, Go, OpenAPI and emitted-response/model results belong in
validation and are not inferred from this review.
