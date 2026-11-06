# Current RPC observation profile — architecture contract 357

Depth cut selected by 355. The source pin and evidence are recorded in
[official convention research](../CURRENT_RPC_CONVENTION_357.md): published
semantic-conventions v1.45.0, commit
`fb30ca17193b551473b1f43c315482918ab6987c`. This contract is closed before code;
implementation and runtime validation remain separate.

## Two deliberate observed populations

Add `rpc-server-current`, UI hint `rpcServerCurrentObserved`, rule set
`goo/rpc-server-current-span-observations/v1`. Its existing service/namespace and
environment/window scope is narrowed by **exact Server kind and nonempty
`rpc.system.name` only**. Preserve all existing predicates. Group by the reported
`rpc.method` attribute using the existing grouping machinery. Method is optional
in the pinned convention: missing/null group, reported empty string and `_OTHER`
remain separate. A method on another row is not required evidence. Do not join
`rpc.service`, use span name as a substitute, synthesize `_OTHER`, parse methods
or union `rpc.system` into the current predicate.

Keep the existing `rpc-server` as visibly **Incoming RPC (legacy)** because it
measures a different reported population: Server kind and nonempty `rpc.system`
and `rpc.method`, with its legacy instrument in milliseconds. This is useful
mixed-data support, not an API compatibility floor. The current visible title is
**Incoming RPC**. Labels, timeline and method diagnosis must explain their
population and inclusive durations rather than imply protocol outcomes.

A dual-key record may belong to both explicitly selected profiles. Auto chooses
one; it does not union or deduplicate them. Separate records from duplicate
instrumentation are also separate captured records. No unique-call/request claim
is introduced.

## Candidate, positive witness and auto choice

Field metadata can make the current family a candidate. It cannot establish that
one Server record has the required current system value in the selected scope.
The UI's positive hint must inspect a **single held row** with matching exact
service, namespace presence/value, checked environment policy, half-open fixed
window and Server kind, plus a nonempty current system value. Retain full expected
question/result admission. Do not combine role from one row, system from another,
or method metadata from unrelated telemetry. A false hint means unknown, not
absence or failure.

Auto precedence is explicit: HTTP → current RPC → legacy RPC → DB client →
messaging consumer → messaging producer → All spans. Current/legacy metadata-only
candidates require explicit selection unless a real positive row witness or
eligible native instrument supports auto choice. A native-only choice remains a
candidate record profile until record evidence is acquired; the profile summary
must not claim observed spans from metric metadata. Manual selection and no-data
read remain truthful and keep All spans available.

## Native evidence

Recognize only `rpc.server.call.duration`, histogram, unit `s`, under the existing
supported native duration admission. Keep exact service/namespace/environment,
instrument scope/version/schema identity, existing definition checks and supported
OTLP temporality 1/2. The latter is a Goo transport/evaluator capability, not an
OTel requirement imposed by the RPC document. Reject wrong names/units/kinds,
client instruments and unsupported temporality; do not convert legacy milliseconds
or blend current/legacy/native/span series.

The native metric's required attributes remain reported dimensions. Instrument
name/type/unit recognition does not certify complete producer attributes or
convention conformance. Metadata dimensions alone cannot prove their values, nor
can native success act as the same-row positive span witness. Native and record
results retain separate clocks, per-stream identity, interval/coverage/reset
semantics and source refusal states.

## Integration and reliability

Reuse existing Go dashboard planning/compiler predicates, record aggregates,
independent native evaluator, passive components and service owner/queue. Add the
new hint and profile to existing validation/selection branches, including
current-aware ranking of native options and fixed-group diagnosis. Keep typed
current/effective named, empty and unreported record policies intact; do not expand
unrelated native environment capabilities under this profile.

Recorded span Error/Unset/Unknown remain existing counters. Do not classify
`rpc.status_code` into a generic success/error policy or infer gRPC completion,
business success, missing exports or complete traces. Current method semantics can
have bounded source-cardinality refusals; do not raise limits for a high-cardinality
method or silently drop groups. Existing per-widget recovery and fixed-question
capture remain independent explicit reads.

## Verification contract

Independently enumerate raw rows for current-only, legacy-only, dual-key, missing/
empty system, missing/empty/`_OTHER` method, Client/Internal role, foreign service,
namespace missing vs empty and environment precedence. Verify current membership
with actual UI hint and Go predicate/SQL rather than generating both oracles from
one descriptor. Include split-row false positives and native-only candidates.
Test native name/unit/kind/temporality/scope mismatches and mixed current/legacy
selection without percentile or count pooling. Native browser evidence should
show auto choice, explicit legacy selection, truthful no-data, method categories,
separate clocks and retained Back. Reuse observed lifecycle/model protocols only
for actual actions; no new profile engine or normalization framework is needed.

## Candidate source acceptance

Read-only review of the frozen Go/UI/parent candidate found no contract blocker.
The Go current helper adds only Server and nonempty current system predicates;
its method dimension is created even without method metadata, preserving missing
and empty groups. Auto checks current before legacy after HTTP. Native current
recognition reuses the existing exact binder, while independently selected legacy
streams retain their own family/unit rather than being relabelled or converted.
The UI hint uses the existing same-row scope/window checker with only the current
system admission key. Browser initial and rerun inputs carry that hint, and
`onPlan` retains the current profile through explicit narrower-window rereads.
Visible labels are **Current RPC** and **Legacy RPC**, consistently distinguishing
the two profiles; these replace the earlier provisional Incoming RPC wording.

New Go tests inspect literal membership, typed policy/namespace preservation,
optional methods, auto precedence and native eligibility. The parent owns their
execution plus generated/native/model gates. Source acceptance is not a claim that
those runtime gates have already passed.

Subsequent canonical execution passes the full Go/UI/API gates, independent
generated/read-only SQL comparisons and genuine native/actor/observed replay.
Successful delta/cumulative results retain separate instrument identities; invalid
mixed-temporality fixture streams were corrected without changing the evaluator.
See [validation](../../VALIDATION.md) and the
[storyboard](../journeys/357-current-rpc.md) for final results and limits.
