# Observed service handoff scope gate — 316

Status: architecture source review accepted; executed checks are tracked in
`VALIDATION.md`.
The proposed user task is to move from an observed Activity record to its service
context without changing environment policy, namespace meaning or fixed window.

## Minimal checked intent

An observed-service handoff should carry a versioned intent with exact nonempty
service name, explicit namespace selector, complete typed environment selection,
and original per-observation fixed bounds. The source action is bound to the
current retained row/owner and its read window, not a global latest tail window.
No rows, result token, source authority, protocol family or query executor belong
in this intent. It stages destination configuration; explicit Read service overview
would acquire data later.

Namespace must distinguish reported missing, reported present empty and reported
literal value. A missing key in an actually supplied ResourceAttributes map can
support missing mode. An omitted/unknown resource-map payload cannot establish
missing namespace and must refuse the action or retain an explicit unsupported
reason. Empty/unreported service names do not become synthetic named services.
Existing `identity.Namespace`/`service-identity` selectors already distinguish
missing/value/empty; do not convert them to nil/all namespaces.

Copy environment via `RecordEnvironment.questionFields`/the typed selection model.
Effective/current-only, named/empty/unreported scopes must not silently become a
legacy environment string. The bounds come from the selected item's admitted
`readWindow` (`ui/tail.mjs`), retaining the exact fixed interval. A record's event
timestamp is not an alternative read window.

## Cross-signal scope semantics

Service overview is context, not a replay of the Activity query. Retain only exact
service/namespace/environment/window; Activity message/severity/trace/span/other
filters cannot be assumed meaningful for spans, logs and metrics alike. Clearly
state that those source filters are omitted rather than guessing portable clauses.
Do not infer HTTP/DB/messaging profile from one row or choose a protocol template
until actual existing discovery semantics support it.

The Activity owner retains paused buffer, original opener and scroll. Destination
staging must perform no source IO. A named return uses that existing owner callback
and restores the paused buffer/focus without Start. No universal resource/trail
owner or automatic cross-signal read is required.

## Confirmed API and application gaps

`exploration.InventoryQuestion` currently contains only Namespace, Environment
string and From/To (`exploration/inventory.go:20`). Its Query constructor creates a
legacy environment query. Service discovery and dashboard planning share this type;
`DashboardRecordQuestion` also copies the string (`exploration/dashboard.go:91`).
Thus Inventory/Discovery cannot currently preserve a typed environment selector.

`ui/service-browser.mjs:18` reconstructs record questions from `scope.environment`.
Its `open` method immediately issues discovery, full spans, metric discovery and
inventory requests, then plans/reads the dashboard. The same path initializes
native instruments, annotations and Activity-view libraries using string environment
scope. Adding a selector field to inventory alone would leave downstream queries
or entity keys downgraded and would not make staging source-free.

These are concrete blockers to wiring an arbitrary typed Activity observation into
the current service open path. The existing service code is not merely a passive
view accepting the proposed intent; it is a multi-source owner with a narrower
scope contract. Do not route a selector through `environment || ''`, invent a
sentinel or run legacy discovery to approximate the requested policy.

## Recommended bounded decision

Keep 316 contract-only unless the team deliberately scopes an end-to-end supported
service entry. Two legitimate next choices are: propagate typed environment scope
through all actual discovery/dashboard/native consumers with explicit unsupported
states, or introduce a clearly bounded record-only service context using existing
record queries while refusing unsupported native/annotation surfaces. Neither may
pretend to be complete existing Services parity. Choose from a concrete task and
request budget before product code, not by adding fields opportunistically.

For any eventual implementation, require constructor tests for namespace missing/
empty/literal versus unknown map, typed environment policies, original row window,
omitted filters and stale record action; emitted API requests must prove exact
scope. Native acceptance must show staging zero reads, explicit destination read,
truthful unsupported states and paused Activity return zero reads. This review
adds no source endpoint, profile or execution owner.

## Settled bounded implementation gate

An end-to-end cut may support only legacy/current-only **named** environment
selection through an explicit semantic encoding. The intent keeps the original
validated typed policy. A destination adapter must prove policy is current-only,
scope is named and name is nonempty, then encode that exact selection using the
existing legacy Services Environment field. Existing environment selection/SQL
contracts define that spelling as current-key-only, so this is equivalent encoding,
not a fallback from an unsupported policy.

Require round-trip Go `SameEnvironment`/selection equality and SQL scope controls
for typed current-only named versus its Services spelling. Effective named—even
with the same name—plus empty/unreported/unknown/invalid policy scopes remain
explicitly unsupported. Never use truthiness to select a default environment.
This boundary permits current OTel defaults without pretending full typed Services
support or migrating unrelated APIs.

The source metadata and zero-read staging requirements above remain unchanged.
Counterexamples include effective named, empty/unreported scope, invalid selector
version/policy, empty legacy name, absent ResourceAttributes versus known missing
namespace key, malformed/null/nonstring namespace, unrepresentable namespace/service
values, replaced row occurrence and global-window drift. Unsupported cases must
show a reason and retain a manual Services path, not attempt approximate discovery.


Use the intersection of actual downstream admission bounds, including the native
measurement ceiling `toMs < 4102444800000`, rather than a looser record-only maximum.
A handoff advertised as supported must not stage a scope that one mandatory
service pipeline immediately rejects. Backend inspection confirms named nonempty
current-key environment equality across record, inventory/discovery and native
metric paths; executable plan/SQL parity remains the validation gate.

## Initial implementation source review

The pure protocol retains the complete original question and canonical current-only
named selector, verifies reported resource environment and a plain present resource
map, preserves exact namespace reporting and checks row event time lies within the
original completed window. Its Services scope adapter deliberately omits Activity
filters and proves equivalent legacy encoding rather than guessing policy.

The destination prevalidates before replacing held state, stages with zero reads,
and routes explicit Read through the existing service owner. Current lease,
generation and actual visibility/inert guards protect Back/Read. Cancel and pending
suspension abort then restage the same checked return lease, retaining paused
Activity ancestry. No source API was broadened.

The source lane binds per-row occurrence, selector, anchor and scope generation;
false destination dispatch restores the source. Review flagged thrown dispatch as
well: it must restore only a still-current paused source rather than leave it hidden
with a status message. That correction is pending final source confirmation.

A Hegel property file (`verification/service_handoff_property_test.go`) generates
legacy/typed current-only scope, missing/empty/literal namespace, quoted/Unicode
values and per-row windows. JS intent/scope equality is checked before existing Go
inventory/discovery admission and executable record-plan parity. Counterexamples
cover unsupported policies, unknown namespace reporting, empty service, out-of-window
observation and malformed intent. At authoring time it was **not run**: the resumed
machine lacked the project Go toolchain (`/usr/bin/go` was unrelated). Do not report
that property as passing until the repaired toolchain executes it.


## Final source review

The source throw-dispatch edge is resolved: a restore closure exists only after a
valid private lease is established, and failure restores only that still-current
paused source. App wiring navigates to the existing Services owner for staging and
back to Explore before invoking the private Activity return; rejected staging
restores route and source without starting a read or merging owners.

Destination staging displays current-key-only scope and an explicit omitted-Activity-
filters disclosure, with original signal/predicate retained as context. Competing
catalog controls are hidden only during this staged owner and restored on return.
Cancel/suspend restage retains named Back, and hidden-tab restaging cannot steal
focus. Per-row timestamp input is bounded before BigInt conversion. Final source
review has no remaining architecture blockers; `git diff --check` passed.

The toolchain was subsequently restored and the property file formatted. Its first
execution attempt was blocked by sandbox network access while dependencies were
missing; an authorized retry is fetching the verification dependencies. Update the
property result only after that command completes.

The generated property subsequently **passed** with the restored pinned Go 1.27.2
toolchain: `TestHegelObservedServiceHandoffPreservesExecutableScope`, 30 Hegel
examples, verification package result `ok` in 65.715 seconds. It checks actual JS
constructor/scope output against generated Go-admitted service scopes and exact
current-only typed/legacy executable record plans. The earlier unavailable/network
notes document the attempted runs, not the final result.

Native first-Back validation exposed a shared visibility defect: the first summary
of its own closed row-actions disclosure was treated as hidden content, so return
focus fell back to Resume. The corrected visibility helper walks every closed
ancestor and permits only membership in that ancestor's first direct summary.
Nested closed outer content and second summaries stay unavailable; hidden/aria-
hidden, disconnected, document-hidden and nonrendered checks remain. Source review
accepts this generic focus fix without opening disclosures or changing source
ownership. Actual native rerun outcomes remain in validation.
