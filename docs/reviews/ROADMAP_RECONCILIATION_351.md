# Roadmap reconciliation — 351

Proposal after the finite 348 breadth → 349 depth → 350 consolidation sequence.
Finish 350's own final native/actor acceptance before starting a new implementation.
This review changes no production code, active roadmap, source budgets or
compatibility commitments. Historical plans are research, not parallel queues.

Read with the [architecture checkpoint](WHOLE_PRODUCT_ARCHITECTURE_351.md),
[design checkpoint](WHOLE_PRODUCT_DESIGN_351.md),
[347 reconciliation](ROADMAP_RECONCILIATION_347.md),
[active roadmap](../ROADMAP_2026_11_05.md) and [known limits](../KNOWN_ISSUES.md).

The product remains **see current telemetry → isolate a population → follow its
supporting signals → preserve/reuse**. The next round should improve ordinary
service investigations under the policies people now use, then make their
independent widgets recover naturally. Do not restart explanation, inventory or
package engines that the last sequence already connected.

## Reconcile the work and lessons

| Work | Carry forward | Retire from the active queue |
| --- | --- | --- |
| 348: fresh-pair explanations | Whole original questions/profiles, independent count/share reads, exact side planning and retained Back | “Fresh pairs cannot be explained”; a second explanation engine or mean-causality claim |
| 349: service-local operation inventory/capture | Bind selected service before grouping, verify raw scope/membership, exact empty names, truthful budget refusal | Whole-environment acquisition plus browser filtering as service authority; raising cardinality caps to unblock small services |
| 350: typed package authoring and Stage → Run | Full checked source/profile through JSON/YAML/hash/storage, detached definitions, explicit planning/read clocks, current-owner admission | Scalar-only package restrictions, count-only descriptions, ambiguous Use-means-read behavior and test expectations that click Stage then await telemetry |
| 332–346: custody, saved-view/parent ownership, comparisons, source recovery and measured facets | Existing source/receipt distinctions, explicit reads, passive shared components and meaningful negative controls | Reopening delivered generic ownership/facet/error work without a new failing task |
| Go maintenance | Go 1.27.3 and useful standard-library features inside straightforward libraries | Gratuitous generics/framework replacement; claiming the still-blocked vulnerability assessment is clean |

Recent failures reinforce bounded contracts: shared DTO fields must be preserved
or refused by every consumer; exact selected definitions matter beyond a claimed
revision; consumed display shape must be checked before adoption. Library actions
and pending metadata need current ownership too. Keep those lessons in the next
task's tests rather than launch a universal scope/owner framework.

## Proposed finite sequence

### 352 — breadth: continue and reuse from a typed service workspace

**User task:** discover services in the selected environment policy, follow a
related child population or find a saved Activity view, and return without
rebuilding the scope.

There are two demonstrated entry gaps. `ui/distributed-services.mjs`
`widerQuestion` and `edgeQuestion` copy only `scope.environment`, although service
discovery already supports typed selectors. The view eagerly constructs an exact
edge link; an unsupported question can therefore prevent useful wider links from
rendering. Separately, `ui/service-browser.mjs` gates Activity's saved-view library
on a scalar `currentName`, while `ActivityViews.library` already accepts a typed
scope object through `scopeFields`/`scopeQuery`.

Reuse shared record-environment fields for the supported wider child question,
retaining exact child service/namespace and fixed discovery bounds. Refuse or
explain unsupported exact typed edges **before** constructing a step link;
`Query.Validate` currently rejects `EnvironmentSelector` plus `ReportedEdge`.
This cut must not claim the next backend extension is already implemented or
silently weaken an effective policy to current-only naming.

Mount the existing Activity library under its checked environment policy instead
of the unrelated native-metric capability gate. Label it as environment-scoped
configuration: neither that listing nor a matching name proves a saved selector
targets this service. Do not inject the selected service, overwrite the saved
selector, or claim a rolling Activity definition retains the fixed discovery
window. Opening stages configuration; Start remains explicit. Existing scalar
entity annotations/deployments and native-stream restrictions remain separate.

**Acceptance:** genuine current-only/effective named, empty and unreported service
discovery with current/legacy conflicts, reported relationships and missing/empty
namespaces. The view renders despite unsupported exact-edge capability; wider
records match an independent source oracle and retain the fixed question. A saved
view from the same checked policy lists/stages with its own unchanged selector;
a foreign-policy response is refused. Display, staging and retained Back do not
read telemetry. Keep exact-versus-wider labels and inspect desktop/narrow actions.

**Stop:** service-workspace continuation/reuse parity. No new catalogue, native
scope projection, graph layout or backend exact-edge claim. This is the best
first item: it fixes broken mainstream entry points using capabilities already
implemented, while making the remaining limitation explicit.

### 353 — depth: inspect an exact reported edge under typed policy

**User task:** choose an observed parent → child relationship and inspect its
reported child spans under the discovery's original checked policy/window.

This is a real compiler/source/admission extension, not another link change.
`Query.Validate` currently refuses typed reported-edge questions, and
`compiler.reportedEdge` in `exploration/reported_edge.go` hardcodes current-key
environment membership for parent ownership. Reuse the existing bounded join and
checked environment predicates for parent ownership **and** child membership.

Resolve ownership before narrowing child records. Preserve reported trace/span
IDs, qualified namespaces, ambiguous/unresolved diagnostics and independent read
clocks. Same-named services and duplicate parent IDs must not manufacture a
unique owner. Counts describe child-span records, not unique calls, complete
traces, causal business transactions or end-to-end latency.

**Acceptance:** scalar/typed current-only equivalence plus effective named, empty
and unreported partitions; conflicting labels on both parent and child;
namespace collisions, ambiguous duplicate parents and outside-window controls.
Assert parameterized parent and child scope before ownership, using an independent
membership oracle. Follow edge → records → capture/reopen → Back with exact
questions and no hidden rereads. Apply the new capability to 352's explicit
unavailable action only after the backend contract and proof are complete.

**Stop:** the existing reported-parent relationship. No async span-link inference,
dependency engine, fleet metric pooling, service health or new topology store.

### 354 — consolidate: retry one failed service widget

**User task:** an overview has useful observations, but one widget failed; retry
that widget without rereading the healthy siblings.

`ui/service-dashboard.mjs` already owns bounded two-worker execution and per-widget
results. Its worker currently converts failures to `Error(data.error)`/strings,
and error cards offer no targeted retry. Reapplying the overview rebuilds and
rereads the successful widgets too. Reuse the existing typed refusal parser,
passive reason display and exact widget questions inside this owner.

**Acceptance:** fail one real record-aggregate widget and one supported native
read. Healthy siblings keep their result/DOM identities and read clocks; explicit
Retry acquires only the failed original fixed query/measurement. Replace that
widget's result rather than append a duplicate. Recompute the overview's state
without implying atomic clocks, turning unknown into zero, or retrying unsupported
prerequisites. Selection/profile/window changes retire pending work and old Retry
handlers. Keep scalar/typed scope and native per-stream identity/temporality
contracts intact. Actual fault/recovery and narrow/desktop frames complement
state/negative tests; no automatic retry loop is selected.

**Stop:** independent recovery for already-supported widgets. Do not broaden native
families or introduce a retry/execution framework, global resubscription manager
or replacement overview planner.

## Material work retained outside this sequence

- **Current RPC defaults:** a useful next protocol-depth candidate. Existing
  `rpc-server` is explicitly a legacy `rpc.system`/`rpc.method` lens; current keys
  are not aliases. First research and pin the applicable OTel convention, then
  propose a separately named profile and exact record/native predicates. Keep it
  outside the first three cuts so resolving the distributed continuation does not
  stop halfway at an unavailable exact edge.
- **Team view from a service context:** verify a bounded applicability/discovery
  bridge using existing authorable templates and service Customize. A template is
  presentation intent, not authority to import its specimen service/window or run
  automatically. This is not a mandate for whole-view WYSIWYG or a new view store.
- **Durable investigation paths/collaboration:** ordered question cards, evidence
  references and same-tab return/recovery already exist. Persistent branching,
  shared workspaces and collaboration require a separate schema/authority task;
  do not rename the current local trail as a durable whiteboard.
- **Measured DOM/byte work:** the roughly 32,000-element paired tree and independent
  comparison/facet byte budgets remain real. Select a follow-up only for a measured
  legitimate user case; 346's duplicate-publication fix did not solve all cost.
- **Operations:** complete the vulnerability assessment when official database
  access is available. Production tenancy/deployment, broad Kubernetes/USE/broker
  coverage, additional native fractional semantics and specialized live recovery
  retain their own contracts. Existing HTTP/RPC/DB/messaging/resource/native
  defaults are implemented foundations, not blank backlog headings.

Keep proposals 260/checkpoint 274 and older roadmaps as dated design background;
they must not gate all product work on reimplementing existing declarations or
converting every presenter. Current work does not require DataFusion, an XTDB
replacement, a new grammar, arbitrary CEL/KCL calculations or broad competitor
parity. Keep research available without silently adding those projects to this
finite queue.

Review source, genuine entry/return journey, model/negative outcomes and actual
frames after each cut. If a concrete correctness blocker appears, repair it and
rewrite the remaining order explicitly. After the third cut, reconsider the whole
product again instead of extending this plan indefinitely with ownership or
styling minutiae.
