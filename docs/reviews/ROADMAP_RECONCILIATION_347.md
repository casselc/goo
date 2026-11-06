# Roadmap reconciliation — 347

Status: proposed next execution order after the 347 whole-product checkpoint.
This document proposes work; it includes no new implementation. Final 346 native and
regression acceptance remains with its own validation record. This review adds
no production code, measurements or compatibility commitments.

Read with the [architecture review](WHOLE_PRODUCT_ARCHITECTURE_347.md),
[design review](WHOLE_PRODUCT_DESIGN_347.md), [active November 5 roadmap](../ROADMAP_2026_11_05.md)
and [known limits](../KNOWN_ISSUES.md). The coordinator should adopt one current
order after accepting those reviews; historical plans are not parallel queues.

The next round should broaden the useful investigation loop rather than continue
iterating on facets. Goo already supports Activity → fixed population → comparison
→ supporting records → retained Back → preservation. The three cuts below connect
existing capabilities across that loop, remove an unrelated-cardinality blocker,
and restore grammar authoring for the typed scopes now used by ordinary entry.

## Reconcile the completed work

| Recent work | Carry forward | Remove from the active queue |
| --- | --- | --- |
| 332, 335: captured evidence and saved-view visit ownership | One admitted visit, explicit rereads, retained task state, retire before abort | The old general capture/saved-view restoration defects; another universal navigation owner |
| 333–334, 337–341: Activity predecessors, operand labels, fresh paired facets, typed trace search, ordinary mean comparison | Whole questions/policies, full-population denominators, exact sums/counts, independent clocks | Building a second comparison executor or treating inspection rows as totals |
| 339, 342: compact provenance and retained controls | Passive components, current owner admission, old callback refusal, deliberate dirty-draft guards | Reopening the same presentation/Back problem without a new failing journey |
| 343, 346: local filtering and measured publication cost | Exact literal identity, same-method styled measurement, explicit owner/model adoption | More setter/filter micro-optimizations as the default next product task |
| 344–345: ordinary and contextual source recovery | Closed typed reasons, bounded safe details, explicit user reacquisition, unavailable stays unavailable | A second generic error parser, text-matching database exceptions, automatic retry |
| 334, 336: Go/runtime and dependency maintenance | Go 1.27.3, current checked dependency graph, meaningful standard-library use including `errors.AsType` | Gratuitous generics/framework replacement; a claim that the blocked vulnerability assessment is clean |

346 does not eliminate the approximately 32,000-element paired tree. 343 label
samples do not establish a maximum supported label size; `/compare` request-body
admission and counted-value response budgets remain distinct. These remain honest
limits, not reasons to postpone the following independent product work.

## Next breadth cut: explain a fresh comparison where it is used

**Task:** After comparing two fixed populations, identify which attribute
populations differ and inspect the relevant side.

The existing `ui/explanation.mjs` owner and `/population/explain` API already serve
Activity and Distributions. The authored fresh-pair population workspace does not
mount that capability. Add a contextual action to an applied fresh count or mean
pair, using both complete original questions, rather than starting another
explanation engine or new sidebar. Reuse the existing explicit field selection,
read, ranked display, side filter and observed-explanation custody paths.

Close the task contract before implementation: `PlanExplanation` preserves the
original request while explicitly deriving count questions without duration or
numeric summaries. For a mean pair, explain **population composition changes**,
not the arithmetic cause of its mean change. Show this distinction at the action
and result. Do not silently interpret the mean as a weight or promise causal
attribution, significance, unique requests or upstream completeness.

Acceptance is one genuine Activity → fixed pair → explicit explanation → chosen
side records → Back → preserve/reopen observed explanation journey. Check the
whole typed selectors, namespace predicates, independent windows and source
clocks, explicit 1–6 field limit, missing versus empty values, partial/unavailable
fields, failed-side behavior and late owner retirement. Opening/staging/Back must
not issue hidden reads. Saving the observed explanation uses existing custody,
not a rerun or a transient result token as durable evidence. Generated semantic
checks should validate the actual planner/intent path; native tests establish
transport, focus and first-viewport behavior.

**Stop:** fresh fixed pairs only. Captured-baseline generalization, automatic
anomaly scanning, mean attribution and expanded statistical scoring need separate
contracts and demonstrated tasks.

## Next depth cut: acquire operation inventory for the selected service

**Task:** Inspect a small service even when unrelated services have more than
1,000 operation pairs in the same environment.

`service-browser.mjs` currently requests `/api/inventory` with `activeScope`, then
filters rows by service in the browser. `InventoryQuestion` has environment,
namespace and fixed bounds but no service selection; SQL groups the environment
before its 1,001-row sentinel. This makes unrelated cardinality an avoidable
failure for the selected-service task. Reuse that existing question/compiler and
service browser; add checked exact service scope **before aggregation**.

The implementation contract must settle omitted versus exact service identity,
including any supported empty name, and preserve the established typed environment
and namespace unions. The selected-service plan, echoed question, continuation and
capture must agree. Capture must not quietly include other service names. Keep an
intentional whole-environment inventory path and its existing refusal; do not
increase caps or relax source stream/entity authority.

Acceptance seeds ordinary OTel service/operation data whose unrelated population
exceeds the original cap while the selected service remains small. The exact
selected-service inventory succeeds; whole-environment refusal remains truthful.
Include namespace collisions, qualified policy cases, literal service names,
empty/no-data and unavailable states, exact record continuation, capture/reopen
and retained Activity Back. Assert parameterized SQL scope and raw response proof
with an independent reference, not a browser-side filter alone.

**Stop:** selected-service operation inventory and its preservation boundary.
New protocol families, Kubernetes dashboards, a discovery engine replacement and
arbitrary native-stream composition are separate proposals.

## Next consolidation cut: typed scope parity for telemetry package authoring

**Task:** Turn an ordinary discovered/interactive question into a reusable grammar
package without losing its environment policy.

The current Explore paint path disables `exploreCreatePackage` whenever
`environmentSelector` is present; the reusable telemetry-package source definition
still has a scalar environment contract. Ordinary Activity and Services now
naturally supply typed scopes. That mismatch blocks the user's manifest-free
package-authoring goal at a mainstream entry point.

Carry the shared typed environment selection through the existing package draft,
YAML definition, validation, planning, fixed preview, save/reopen and fresh run.
Reuse the existing identity/record-environment AST; do not weaken an effective,
empty or unreported scope into a scalar name. Explicit namespace/service/attribute
predicates must survive unchanged. Keep a reusable relative lookback separate from
the original absolute preview and from a later live execution window. Negotiate
any prerelease format change directly; no compatibility alias or legacy branch is
needed.

Retain the existing package profiles: `TelemetryPackage` already carries numeric,
histogram, duration, span-status and grouped-time directives, despite its
`captured-record-count` measure label. Do not implement these a second time or
repeat older count-only documentation as a current restriction. Consolidate the
typed-selector gap and correct capability descriptions; a canonical replacement
measurement DSL remains separate. Source-review
annotation generation separately: a code-local annotation may not represent an
arbitrary external resource/environment scope. Enable it only with an exact
meaningful projection, or give a clear unsupported explanation while the package
remains usable for non-Go instrumentation. Scope parity must not become full
WYSIWYG, a new package catalogue or a measurement-language rewrite.

Acceptance follows Activity/Services → question → package draft → fixed preview →
YAML → save/reopen → explicit fresh run, using named current-only/effective, empty
and unreported environments plus namespace predicates. Generated Go/JS roundtrip
and planner checks should expose any policy weakening; native records establish
actual current/legacy-key behavior and independent preview/run clocks.

Use the design review's end-to-end entry/preserve audit as this cut's usability
acceptance: inspect natural desktop/narrow first viewports, keep data and the next
useful action visible, and remove duplicated prose/settings only where actual
frames demonstrate a problem. This is a concrete journey check, not another
unbounded component or navigation redesign.

## Deferred work and reevaluation

A durable investigation path is a worthwhile later candidate, but first reconcile
its ordered steps/citations with existing investigation cards and custody. Do not
claim retained child history or same-tab draft recovery already persists branches.
A proposed contract can follow these three cuts; implementation should wait for
that schema/authority review and a clear task.

Also retain, without turning them into the immediate queue: large-tree/byte-budget
work justified by new measurements; specialized/native/live recovery beyond the
ordinary owners; newer OTel convention and resource-family defaults; fractional
counter semantics where a selected panel needs them; production deployment and
shared collaboration. Existing automatic HTTP/RPC/database-client/messaging,
host/runtime and native interval tooling are implemented capabilities, not blank
backlog items merely because broader families remain incomplete.

The vulnerability assessment remains blocked by official database access. Complete
it when that access becomes available; avoid changing packages just to create a
maintenance milestone. Current runtime/library features are tools for simpler
contracts, not a requirement to expose clever type machinery to users.

After each selected cut, review source, actual journey and compact validation
before advancing. After the third, reconsider whole-product priorities. If one
cut exposes a concrete correctness blocker, repair that blocker and rewrite the
remaining order; do not silently turn this three-cut plan into a long series of
facet or ownership minutiae.
