# Whole-product simplicity and understandability review

Reviewed 2026-11-17 against implementation `9119a967`, after the seven component
walkthroughs and reviews were complete. This is a product/code assessment, not a
new usability study, competitive benchmark or runtime performance experiment.
Recommendations below are proposed work; this review changes no implementation.

## Product judgment

Goo's useful center is already coherent: **see telemetry → isolate a population →
follow context → preserve observations or reuse a question**. The [README](../../../README.md)
and [capability guide](../../CAPABILITY_GUIDE.md) describe that loop, and the code
has real supporting boundaries: one Go runtime, typed questions, explicit source
owners, passive components, and separate configuration/observation storage.
The product is substantially broader than a reservation demo. Another chart or
semantic profile will not, by itself, make it easier to understand or maintain.

The main cost is discovering which owner and lifetime a change belongs to.
Compressed authored JavaScript and large application closures obscure the useful
architecture. Historical documentation and test wrappers add a second learning
burden. An afternoon should produce a working map of one component, not require
reading every profile, milestone or verification receipt. The new
[walkthrough index](../../walkthroughs/README.md) is the right entry point, but
readability of the source itself must catch up.

I read all seven walkthroughs and component reports, then checked representative
paths in [`app.mjs`](../../../ui/app.mjs),
[`workspace.mjs`](../../../ui/workspace.mjs),
[`investigation-trail.mjs`](../../../ui/investigation-trail.mjs),
[`numeric-plot.mjs`](../../../ui/numeric-plot.mjs),
[`serve_http.go`](../../../cmd/goo/serve_http.go),
[`Makefile`](../../../Makefile), and
[`run_journeys.sh`](../../../scripts/run_journeys.sh). The findings below combine
those source checks with the explicitly scoped component reviews.

## Ranked outcomes

| Priority | Outcome and evidence | Bounded change and acceptance |
| --- | --- | --- |
| **1 — contributors can follow the normal operator loop without reconstructing compressed code** | `population-workspace`, `exploration`, `stream-entry`, `captures` and `documents` combine UI construction and several independent lifetimes. The Activity/presentation/preservation reviews show dense authored lines and broad nested closures. | Start with a few representative owners: ordinary formatting, named phases and an explicit callback/owner map. Then extract one cohesive editor or publication controller. Keep each patch reviewable; do not rewrite the UI or adopt a state framework. A non-author should locate staging, execution, retirement, local Return and write acknowledgment without reading unrelated forms. Existing request-count, hidden-action, mutation and uncertain-write cases must remain unchanged. |
| **2 — Return is an owner contract rather than a custom-element implementation trick** | `investigation-trail.leave/restore/select` scans descendants, saves child nodes of every `goo-*` element and puts those nodes back after reconnect. The 461 bug showed why selected identity must not depend on visible rows; 462 supplies a smaller immutable reading-state precedent. This review establishes coupling, not a demonstrated leak. | Characterize one retained result owner, then give it an explicit reading-state/restore interface. Keep source suspension and publication acknowledgment separate. Acceptance includes eight visits, absent selected rows, preserved Find/page/inspector, no source reads on Return and measured retained nodes/listeners. Do not replace all component state with one universal serializer. |
| **3 — Save, Keep, Capture and investigation mean the same things everywhere** | These operations have useful distinct effects, but multiple stores and similarly named surfaces make the code and product harder to navigate. `registerInvestigations` actively serves operation-package evidence; it is not the narrative document store. | Publish one short operation/store map and use consistent action language and named owners. Rename ambiguous internal operation-evidence ownership if needed. Keep reusable definitions, held promotion, fresh capture, notes/assertions and exact review decisions distinct. Test effect and recovery semantics, not merely menu labels. Avoid a universal Save/retry/store wrapper. |
| **4 — a fresh checkout has a truthful, reservation-independent development loop** | `make test` omits nested shipped modules; CI separately checks them. `run_journeys.sh` always drains the reservation workflow although generic runtime journeys own independent fixtures. Giant append-only verification narratives obscure the right command. | Provide a small current check matrix and a shared explicit application-module list; separate generic journey execution from reservation lifecycle. Keep optional native tools explicit and required verdicts strict. A generic journey must not require reservation Redis, and a deliberate nested-module failure must fail the documented default checks. Measure CI savings before deleting repeated runs. |
| **5 — the operator can distinguish available data from supported computation without reading a policy essay** | Source compatibility, observed instruments, evaluation eligibility and display certainty are distinct in `sourcecap`, `instrument-capabilities`, raw profiles and view admission. Capability gaps are real, but repeated prose is not the best explanation. | Keep that capability ladder explicit in small consistent presentations and refusals. Extend the existing shared components where the meanings agree. Do not flatten it into `supported: true`, or build a global registry to eliminate purposeful distinctions. Verify absent-table, denied-source, raw-only, evaluable and unknown-coverage cases across Activity, Services and Measurements. |
| **6 — optional advanced work has a named purpose and an exit decision** | XTDB retention, training/review exports, business lifecycle packages and operation annotations all have active references. No component review proved a whole production feature dead. Several are wider than the core first-use loop. | Keep these as explicit optional capabilities. For each, name its user outcome, owner, test cost and graduation/removal criterion. Decide whether to retain or retire an experiment from actual intended use and dependency coverage, not package names. Remove only confirmed aliases/indirections opportunistically now; do not delete an active workflow solely because it is not the default screen. |
| **7 — performance and reliability work follows user-visible budgets** | Acquisition can make independent sequential reads; catalog filtering can scan bounded histories; retained visits can traverse DOM. Reviews identify plausible costs, not measured regressions. | Capture source request/byte budgets, first useful rows, local-interaction work and suspended-view resources on representative bounds. Improve the dominant measured cost while preserving independent-read uncertainty and exact identity. No blanket parallelization, caching or removal of validation/cancellation guards is justified yet. |

## Keep, defer, stop

**Keep:** the stream-first experience; ordinary OTel usefulness without manifests;
shared native controls and passive data displays; checked predicates/CEL as an
optional authoring layer; explicit scope, uncertainty and clocks; original-window
continuations; separate read and publication lifetimes; and narrow schema/profile
admission. These are the product's reliability, not accidental bureaucracy.

**Outside this readability slice:** arbitrary executable kinds, a general
WYSIWYG canvas, whole-query CEL/optimizer work, warehouse execution, expanded
training queues, multiplayer presence and broad XTDB federation. Retain the user
requested directions; this review does not cancel them or imply absent interest.
Expanded implementation needs a scoped near-term use case and the corresponding
design/admission gate. Current checked compositions, native forms, notes and saved
views provide a useful base for that work.
Production auth/tenancy, durable monitors and alert delivery remain genuine gaps
for a production service; do not imply they are solved by local saved questions.

**Stop doing by default:** adding another toolbar, format or evidence artifact to
complete a historical checklist; treating physical line count as comprehensibility;
explaining every policy permanently on screen; making generic tests depend on the
reservation sample; and reporting optional skips as successful validation. Preserve
reproductions, models, screenshots and receipts only where they serve an owned
contract, documentation or future feature work.

There is no evidence here for deleting an entire Go package, chart family or
product surface. The confirmed small alias/lookup leftovers in the component
reviews are routine cleanup, not a substitute for the larger ownership work.
The product has no Python proxy; active Python reference tests and independent
lab workloads are different decisions.

## Sequencing relative to 463–465

The [adopted cycle](../../ROADMAP_2026_11_14_CHECKPOINT_429.md) remains a sensible
breadth/depth/consolidation sequence, but this user-requested review should result
in a **bounded readability and ownership step before more feature work**:

1. Establish the walkthrough/check front door, format a representative owner and
   record its effects/lifetimes. Characterize the retained-visit contract before
   extracting one owner. Limit this to reviewable patches, not an open-ended rewrite.
2. Use **463 All-environments ordinary spans** as the next general-purpose proof:
   an operator with unknown labels should get useful spans, then narrow explicitly.
   Review the carrier matrix first; do not globally remove narrow-scope guards.
3. Proceed with **464 raw exponential-histogram inspection** only after the real
   source profile proves trustworthy occurrence/count facts. It is a distinct
   coverage outcome, not permission to invent zero-threshold/statistic presence or
   implement an evaluator/chart prematurely. Record a source blocker if necessary.
4. Make **465 capability consolidation** address the duplication demonstrated by
   those two flows. Keep it bounded. Then perform the planned whole-product
   checkpoint, including whether operational deployment/monitoring now outranks
   more profile depth.

The architecture and Go reviews should choose the exact internal refactoring
order; product acceptance is that these changes shorten the route from a user
question to its owning code and make the normal investigation loop more reliable.

## How to judge the afternoon goal

Ask a developer who did not author the component to follow its walkthrough and:
identify the next explicit read; find the owner that cancels it; explain why Return
cannot query; distinguish saving a question from keeping/capturing observations;
and point to one focused happy-path and one refusal/withdrawal test. Record where
they had to search outside the map. This is a proposed comprehension exercise,
not an executed study or a claim that all seven components can be mastered in one
afternoon.

No implementation, live-source campaign, browser study or performance benchmark
was run for this holistic product review. Component checks and their limits remain
recorded in their individual reports.
