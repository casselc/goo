# Adversarial review: what must Goo actually be?

Reviewed 2026-11-17 at `9119a967`, after the component and holistic reviews.
This is a deliberately contrary product/code judgment, not an implemented cut,
usage study, benchmark or proof that registered code is unreachable.

## Verdict

**The reviews are too protective of the existing product boundary.** They correctly
protect scope, clocks, provenance and publication outcomes, but repeatedly answer
removal questions with “it has active callers.” Active code can still be unnecessary
for this product. Formatting and extracting every existing feature would make an
oversized product more legible without making it smaller.

Goo currently carries at least four ambitions: an ordinary observability
workbench, an application-contract/measurement system, an annotation/review and
temporal evidence platform, and a verification/instrumentation laboratory. They
can reinforce one another, but they should not all determine first use, default
navigation, core bootstrapping and the default developer workload.

The essential product is: **turn existing telemetry into a useful, reusable
investigation without requiring a bespoke instrumentation or data-governance
project first.** Its distinctive reusable grammar matters. A new generic log
viewer alone would throw away the strongest reason to build Goo.

## Smallest coherent product

Start with **Explore, Services and Saved** as the default conceptual surfaces:

1. See recent ordinary logs, spans and useful native metrics, even when labels
   are unfamiliar. Show actual scope, limits, failure and unknown states.
2. Narrow a population, inspect a record/trace and compare against another scope.
3. Return locally without surprising source reads.
4. Save a question/view, or deliberately preserve observations and attach a note
   or label. Keep those meanings distinct behind a consistent interaction.
5. Reuse the same checked query/data/view primitives in a team-contributed view.

The essential implementation is typed scope/questions, bounded source adapters,
correct native kernels, a small supplied-data grammar, ordinary application owners
and reliable configuration/observation storage. CEL is an optional expression
editor over that model, not another execution platform. Team authoring remains
real; it need not be another top-level product.

## Cuts worth making decisions about

| Action | Evidence and proposed cut | What is lost / counterargument / gate |
| --- | --- | --- |
| **COMBINE default surfaces** | [`ui/index.html`](../../../ui/index.html) has separate Explore, Services, Operations, Saved and Definitions navigation. [`app.mjs`](../../../ui/app.mjs) imports and wires prepared operation, fleet, ratio, histogram, package, retention and generic exploration owners together. Put reusable definitions beside Saved views and contextual Edit/Build actions; make declaration-specific Operations an explicitly enabled entry. | Power users lose a dedicated shortcut unless retained contextually. This is not permission to remove shared grammar or instrumentation. Prove a team can create, share, reopen and edit the same view without a parallel authoring workspace; compare task paths before changing navigation. |
| **REMOVE the prepared-demo shell from the ordinary product path** | [`serve_http.go`](../../../cmd/goo/serve_http.go) separately composes prepared HTTP/declared-operation GET routes, capability responses and `registerInvestigations`; the generic startup still registers that machinery and the browser still wires its disabled surfaces. Move declaration/conformance UI and its separate operation-evidence format behind the existing explicit package configuration, outside normal composition. Ultimately retire the parallel shell if Services plus ordinary typed captures can express its useful outcomes. | Declared-versus-observed conformance, versioned operation evidence and some example flows are unique today. Preserve them as an optional capability or port them to the common path first. Registration proves they work, not that every ordinary user should understand them. Do not delete stored artifacts; name their export/retention consequence before retiring a format. |
| **MAKE business lifecycle and warehouse compilation optional** | [`measurements/source.go`](../../../measurements/source.go) branches between ordinary numeric evaluation and lifecycle custody; [`lifecycle`](../../../lifecycle) and [`subjectfacts`](../../../subjectfacts) introduce prepared subjects, deadlines and warehouse aggregation. Keep those as an explicit domain extension/CLI path, not prerequisite concepts for understanding a native counter or inspecting telemetry. | Correct business-outcome evaluation and Databricks metric-view output are valuable requested directions. Do not replace them with span guesses. Isolation loses no outcome; outright removal would. Give the extension a concrete owned example and keep its custody tests in its tier. Avoid inventing a plugin platform to accomplish separation. |
| **KEEP annotation; MAKE adjudication a separate commitment** | [`workbench/meaning.go`](../../../workbench/meaning.go) covers useful notes/tags/revisions. [`capture_review.go`](../../../workbench/capture_review.go) and [`capture_reviews_http.go`](../../../cmd/goo/capture_reviews_http.go) add tasks, decisions, supersession and manifests. Prioritize natural annotation of an exact record/population and a useful export; stop expanding a review-management product until a real training workflow requires it. | The user explicitly requested training-useful annotations. That is a reason to preserve stable targets, labels and corrections, not proof that all task/adjudication machinery belongs in core. Removing the queue loses policy-controlled multi-step review; retain it optional or demonstrate that outcome before promoting it. |
| **KEEP XTDB experimental, stop its gravitational pull** | [`retention/store.go`](../../../retention/store.go) and [`tail.go`](../../../retention/tail.go) add connection/schema/transaction/await semantics; [`app.mjs`](../../../ui/app.mjs) and runtime composition know the adapter. It is **already configuration-gated**. The new cut is to keep its UI/composition/test ownership explicitly outside ordinary investigation work, not merely add another unavailable-state message. | Bitemporal corrections and retained training records were explicitly requested research. Local immutable revisions do not automatically replace native bitemporality. Require one correction/as-of/export journey showing what XTDB adds before broader federation or profile expansion. Core Save/Keep/Capture must remain understandable and testable without learning XTDB. Do not add pooling/migrations until the experiment's ownership and deployment target justify them. |
| **KEEP a small display vocabulary; COMBINE encoding implementations** | [`display.mjs`](../../../ui/display.mjs) exposes many related timelines, comparisons, trace breakdowns, raw plots and kind helpers. [`population-workspace.mjs`](../../../ui/population-workspace.mjs) owns many accompanying editors. Organize around records, temporal series, distributions, waterfalls and comparison; express variations as checked kinds/settings where encoding and interaction truly agree. Stop adding a new bespoke editor per chart. | Different histogram, gauge, completeness and provenance meanings must survive. Module count alone does not prove a chart redundant. Keep team-contributed kinds/views; defer a general WYSIWYG canvas from this pruning slice, not from the user's accepted direction. Require a reuse/extension example before adding another base abstraction. |
| **REMOVE reservation infrastructure from the default experience** | [`run_journeys.sh`](../../../scripts/run_journeys.sh) always drains Redis reservations/inventory before any requested journey. [`Makefile`](../../../Makefile) defaults to the full lab while generic-runtime fixtures can own their source/runtime. Make the smallest useful lab one ordinary OTel workload, collector, ClickHouse and Goo; keep business stress scenarios as named optional fixtures. | Reservation failures are useful distributed-system tests. Keep their unique assertions, but do not make their services or domain model necessary for a generic first investigation. Preserve cleanup of fixtures actually owned by each test; do not reset shared data to simplify scripts. |
| **STOP verification and documentation expanding per milestone** | [CI](../../../.github/workflows/api-contract.yml) runs root tests, focused repeats, API observation checks and a gotogether shadow lane; [`verification`](../../../verification) contains several independent runners/models. The walkthrough census records 969 Markdown files and 126,581 documentation lines before this audit adds more. Keep one current contract/check map and behaviour-named suites; retain extra campaigns only for distinct failure power. | Hegel, Quint, Bombadil, Playwright and gotogether were explicitly requested and can catch different failures. Keep that division of responsibility, not every historical wrapper/ledger. Upstream gotogether experiments should not dictate Goo's production design or block unrelated changes without a demonstrated contract. Port unique assertions before deleting an offline oracle. Do not let a test passing only because a tool skipped count as pruning success. |

These are product-boundary recommendations, not claims that a directory can be
removed safely with one `rm`. Prefer removing an unwanted capability from the
normal dependency/learning path over retaining its complexity under nicer names.
A simple explicit optional composition is enough; more feature flags, runtime
registries or build-profile matrices would defeat the cut.

## What I would not cut

Ordinary OTel usefulness; named/missing/empty scope distinctions; counter and
histogram correctness; exact observation identity; a coherent reusable view model;
accessible shared controls; explicit read versus local interaction; original-window
Return; and reliable saved work. These are essential, even where their code is
longer than a permissive alternative.

Save configuration, Keep admitted observations and Capture a new read can share a
menu and some storage mechanics. They cannot become one ambiguous operation.
Likewise, physical source compatibility and evaluation eligibility are different
facts. Pruning their duplicate presentation is useful; deleting their meaning is not.

## Challenge to the proposed order

I disagree with “format and decompose the current product, then decide what stays.”
**Decide the core/optional boundary first**, including which UI owners will stop
being part of default composition. Then format/extract the retained core one owner
at a time. Otherwise the first refactor entrenches everything it makes beautiful.

I support **463 All-environments ordinary spans** because unknown-label first use
belongs to the essential flow. I would not commit to 464 merely to honour a
breadth/depth rhythm. Advance raw exponential-histogram inspection if actual source
coverage makes it the next first-use blocker; otherwise prioritise a confirmed core
failure or deployment boundary. Use 465 only to consolidate demonstrated paths,
not to build a capability framework in anticipation of every extension.

Before another broad feature cycle, make one new contributor complete a generic
investigation, create a reusable team view, label a captured item and reopen it,
without the reservation lab or advanced packages. Where that fails is a stronger
priority signal than the age or number of a roadmap item.

## The audit itself needs an exit

The seven walkthroughs are useful as one current handbook. Raw reviewer reports
are supporting material, not seven more normative specifications plus four new
backlogs. Publish one reconciled cut decision and one bounded next change. Keep
one capability guide, current architecture/owner map, API/check reference and
walkthrough index authoritative; archive the rest as historical rather than
requiring every newcomer to arbitrate it. Do not commission another full-product
review after every small slice.

This critic read the current guide, roadmap, seven walkthroughs, holistic reports
and representative composition, navigation, measurement, preservation, display
and delivery code. No usage telemetry or benchmark was collected, and no product
code or data was removed. The disagreement is over what deserves to be core,
not over whether the existing acceptance evidence is real.
