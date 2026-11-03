# Typed native measurement environment — architecture 319

Status: final source review accepted; execution is tracked in VALIDATION. Deliver actual native
values under the same checked service environment policy, using existing instrument
picker/planner/evaluator owners. Entity/deployment/service annotation semantics stay
current-only named and are not broadened by this milestone.

## Useful first family and strict boundary

Support typed environment selectors for service-bound **per-stream gauges** first.
Add the selector union to native Definition and DiscoveryQuestion, preserving legacy
`environment: name` as current-key-only. Effective named, empty and unreported must
use the same current-key precedence as record scope, including present-empty
shadowing a legacy name. Never enable a native value by replacing its selector with
a name string.

Unsupported typed counter/stock/histogram/state/lifecycle and resource-specific
families must fail before source IO or appear explicitly unsupported in discovery.
Do not infer family support from metadata presence or leave a Definition attached
to an instrument marked unsupported. Legacy named definitions and their existing
families retain their behavior. If equivalent typed current-only named support is
expanded beyond gauge, it requires explicit validation/executable parity rather
than an accidental string fallback.

## Concrete boundary map

- `measurements/model.go` currently requires nonempty scalar environment in
  Definition validation. Introduce the closed mutually exclusive union and checked
  environment selection once, enforce the family boundary there, and carry it
  through YAML/JSON package revisions. Preserve exact stream/resource binding rules.
- `measurements/discovery.go` has scalar DiscoveryQuestion and three-family SQL
  hardcoded to the current resource key. Compile the actual checked policy, preserve
  it in returned question and generated supported gauge definitions, and detach
  new selector pointers. Returning metadata for unsupported families is useful only
  with truthful capability/reason and no executable typed definition.
- `measurements/source.go` hardcodes the same current key in Compile and decoded
  row membership. Both must use the checked selector; SQL-only filtering is not
  sufficient. `measurements/exact_stream.go` also checks scalar environment and must
  enforce selector membership for gauge exact streams. Keep all original raw
  ResourceAttributes in stream identity: two current/legacy resources matching the
  same effective name remain distinct streams, not a fleet merge.
- Common capture Scope/normalization, definition revision/package storage and
  discovery-to-measurement transforms must preserve selector identity. Labels and
  exported packages must not show empty environment or silently discard policy.
- Browser native definition/stream validators, measurement discovery retained scope,
  service dashboard native matching, catalog/evidence/return labels must use the
  same scope representation and capability rules. Existing histogram validators
  must continue to refuse unsupported typed histogram definitions rather than
  accepting new fields broadly.

## Service integration and ownership

The existing service instrument discovery and picker should expose supported typed
gauges and make unsupported native families clear. The explicit native read uses
its existing evaluator with exact unit, stream identity and original fixed window.
Raw metric-record counts remain separate from values and do not become a fallback
when no typed gauge can be evaluated. Keep the same cancellation/return/held-data
contracts established in 316–318.

Service native capability can no longer be a blanket rejection of every effective
scope once gauges work; it must describe the supported gauge subset and mark other
families unavailable. Do not turn empty supplied instruments into a healthy zero.
Current-only entity annotations, deployment context and declared operation bindings
remain independently gated, even when native gauges become supported.

## Implementation constraints and acceptance

Use shared logical environment selection/membership from identity. Measurements
cannot import exploration's private SQL lowering without a package cycle; a small
source-specific lowering with executable parity tests is preferable to moving SQL
into logical identity or creating a general query engine. Preserve typed parameters,
source budgets, current-only legacy plans and exact raw stream keys.

Demonstrate effective legacy-SDK service scope → instrument discovery → explicit
gauge value with exact namespace/unit/window → capture/package/reopen and paused
return. Test current/legacy equal names remain distinct streams, present-empty
shadowing, effective empty/unreported, typed/legacy current-only parity, malformed
union/family refusal before SQL, wrong resource-policy membership and returned-scope
aliasing. Go/browser/schema and actual emitted responses must agree. The final
value/quality must be a real gauge result, not an instrument record count or source
fixture rendered as if evaluated. No new standalone native explorer is needed.

## Settled family and package boundary

The accepted cut admits `environmentSelector` only for service-bound per-stream
gauge definitions, rejecting typed-selector non-gauge and resource-binding
variants. Legacy scalar definitions retain all existing family behavior. Services
may continue its already checked current-only named scalar projection for the
older native families; effective policy must never use that projection. Root owns
common capture Scope/native-record context propagation.

The package catalog already validates Definition revision, serializes its typed
YAML and verifies bounded read/revision on reopen. JSON digest naturally includes
new selector policy; omitted optional selector leaves legacy identity unchanged.
No extra workbench schema or storage owner is needed. Identity selector/scope
already have YAML tags. However strict JSON union decoding alone cannot reject
YAML `environmentSelector: null` beside legacy environment (or null environment
beside a valid selector); Read must preserve/check wire presence before those
values disappear. Retain 32 KiB package, strict field/depth and exact-document limits.

Source.Run currently compiles before reading and clones the question only later
inside result preparation. The new selector pointer must be detached before source
IO so a caller/transport callback cannot alter the policy between compiled SQL and
evaluation/returned question. Apply the same principle to instrument discovery
metadata. Existing checked clone patterns suffice; this does not require a new
serialization or immutable-object framework.

## Initial UI and store audit

The frozen UI retains checked discovery environment fields independently of their
display label, clears them only on actual environment editing and carries policy
through native package/evidence/return labels. Native-record context explicitly
admits typed service per-stream gauges and checks stream resource membership via
shared environment selection; histogram scope admission remains strict and rejects
the new field. Raw current/legacy resource-key labels distinguish streams that
otherwise share an effective name. Remaining scalar lifecycle/pool-family paths
are outside the typed gauge family and must stay refused before they execute.

Catalog storage uses the existing Definition Revision/typed YAML/read path; no
separate workbench definition transport exists to update. The new selector's
presence must be included by the current typed serializer rather than copied into
metadata-only labels. This lane has no initial source-review blocker, pending
actual backend union/Read/Run-entry fixes and emitted-response parity checks.

## Backend source review

The landed backend addresses the package and mutation boundaries: closed JSON
Definition/Discovery decoding refuses null environment fields, YAML Read checks
those fields before decoding, and typed Definition admission is service per-stream
gauge only. Run detaches the complete question before compile/IO; result preparation
uses the checked clone path. Discovery detaches namespace/selector and each emitted
definition gets its own selector copy.

Compile preserves the legacy SQL branch while typed selection lowers through
validated parameterized metric predicates. Evaluation and exact-stream binding
check the same logical selection against the actual resource map. No resource key
normalization or effective-label stream merge was introduced. Unsupported typed
families retain metadata but have Supported false, no executable Definition and an
explicit reason; capability planning derives eligibility from that state. Current
source review finds no backend architecture blocker, pending final scope/dashboard
integration and actual tests.


## Final integration review

Final Go/UI/schema source review has no architecture blocker. Dashboard native
options compare checked environment selections, validate executable definitions and
retain literal service/namespace/instrument identity. Unsupported definitions cannot
enter a widget through a missing-definition shortcut. Only declared entity bindings
keep the separate current-only named gate; effective scopes truthfully expose the
gauge-only capability instead of claiming global native availability.

Measurement capture normalization already round-trips the typed question through
JSON, so the new selector is detached and retained without a second catalog model.
Common Scope, discovery transform and native-to-record context now preserve the
selector and namespace. Definition revision and immutable package storage include
the selector through the existing typed JSON/YAML paths. No new acquisition owner,
stream merger, scalar policy fallback or family expansion was found.

The earlier null-YAML and before-IO detachment findings above are resolved by the
landed source. Focused UI checks were reported passing by their owner; this review
lane did not independently execute the 319 test suite. Actual Go, cross-language,
native UI and source-budget execution results belong in VALIDATION.
