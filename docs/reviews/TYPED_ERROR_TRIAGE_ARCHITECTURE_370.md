# Typed error-triage follow-through — architecture 370

Status: source, native, generated, actor and observed-model acceptance passed.

## Problem and retained boundaries

Error triage already constructs complete checked questions. `inspectGroup` adds
exact service/type predicates; `compareInService` narrows the service and retains
type grouping. Typed policy, fixed bounds, original predicates and namespace are
preserved. Missing and present empty values remain distinct. These are reported
error populations, not deduplicated issues or inferred application failures.

The fixed exploration owner previously omitted these actions for every typed
environment selector. Removing that guard alone was unsafe: the group-table
adapter created bare hrefs without action-time ownership, and `canRefine` ran
only while constructing the link. A stale parent anchor could reach the trail's
delegated navigation after its source owner was hidden or replaced.

The href itself does not auto-read. The investigation trail intercepts ordinary
local question links and explicitly stages them; reload/new-tab restoration also
stages. Fresh modified-click and sharing behavior therefore retain normal links.
The repair addresses origin authority, not a missing compiler or router.

## Implemented source contract

Only the triage group typed-policy barrier is removed. Actions require an
observed/no-data two-field service/type result, active applied session, captured
generation, exact current result object and membership of the original group.
Read-only, changed-between-reads/refused results and invalid drafts cannot act.
Builders validate the whole child question; raw service labels do not become
globally unique namespaced service identities.

The existing group adapter retains anchors under a local owner. Each onclick
captures an immutable lease and checks connected/contained, visible/non-inert
controls and the caller's current predicate. Stale callbacks prevent default and
stop immediate propagation before the trail delegate, including when an old
anchor is reattached. Replacement/clear/suspension retire the lease; resume
rebinds the same anchors without repainting observations or issuing reads.

A dedicated `triageGroupLive` flag renews this authority on retained Back only
for an active connected parent outside recent/population presentation, with an
applied session, observed/no-data held result, unchanged evidence basis and
canonical whole question, and valid filters. It does not renew the broader
revoked action context or unrelated distribution actions. Existing anchor DOM,
receipts, focus and scroll remain held by the trail.

Stage reads no telemetry. Explicit Run alone acquires the fixed child. Back
suspends/cancels that actual child owner, so late completion cannot replace the
retained parent. Source/store/SQL counters establish these bounded operations;
they do not establish ingestion completeness or transactional source snapshots.
Shared questions do not grant capture custody or execute automatically.

The separate typed distribution population barrier remains unchanged. Its rest/
view/template destination arms need their own contract before removal. No blanket
typed-capability switch, scalar downgrade or new source/session framework exists.

## Verification and final acceptance

Focused tests exercise the actual anchor owner and typed builders. Generated
cases execute actual JS, admit output in Go and compare independent raw predicate
membership across policies and nullable service/type/namespace values. The native
ledger checks parent/child record IDs and exact grouping tuples/counts against
independent fixture/raw-query facts, including literal, empty and missing types.
Parent evidence and pending late-response snapshots are separately checked.

Observed replay consumes actual fact names and cumulative source counts, not a
fixed success script. Depth is a logical action projection; the model proves
read/depth safety, while browser/Go checks prove membership, custody and owner
state. Direct source/order/hidden-false corruption controls are rejected.

The final genuine script passed all policy cases in approximately 2.0 minutes
(runner 2.1 minutes), including actual detached-source preventDefault, unchanged
state/history and narrow 390-pixel action-focus visibility with zero reads.
Final observed replay passed in 13.19 seconds (package 13.201), zero skips. The
initial full script/actor/model campaign took approximately 2.8 minutes; its model
passed in 15.59 seconds. Its positive actor completed in 45.151 seconds with 688
snapshots across 13 phases, three reads, one SSE stream and zero writes/store work.

Root reports full Go passing (cmd 8.612 seconds; workbench cached), 1,276 UI checks
in 19.786 seconds and Hegel 40 in 3.90 seconds. No acceptance gate remains pending.
These results cover the bounded typed triage continuation and retained Back;
they do not imply global query-history persistence or broader distribution parity.
