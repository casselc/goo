# Whole-product review — 2026-10-28

> Historical same-day review. The [post-234 refresh](WHOLE_PRODUCT_REVIEW_2026_10_28_REFRESH.md)
> reconciles delivery and owns current recommendations; this file preserves its
> original observations and external-research limits.

Reviewed source at `9154922`, earlier product/architecture/usability/library reviews,
and the running local Chromium app. This combines product, design, architecture,
performance and reliability findings. It is a technical/design review, not a user
study, load benchmark, production-readiness assessment or multi-browser campaign.
The lab frontend was inspected; this does not establish that its deployed backend
has every newer source-only refactor. Current order and acceptance are in the
[revised roadmap](ROADMAP_2026_10_28.md).

## Product judgment

Goo's job is to help an engineer see current observations, isolate an interesting
population, follow its logs/traces/metrics, and reuse the resulting question or
explanation. This already works across ordinary OTel and declared application
operations. The next gains are completing ownership and reducing interaction cost
across those existing tasks. More isolated chart types, evidence object types or
infrastructure-specific one-off panels are lower priority.

The recent Go baseline, native ESM, HTTP family policies, shutdown, explicit
instrumentation runtime and pure native evaluation are shipped foundations. They
are not a reason to keep refactoring indefinitely before improving the product.

## Findings from the current app and source

| Priority | Finding and evidence | Consequence / bounded response |
| --- | --- | --- |
| 1 | `ui/app.mjs` uses generation checks but uncancelled raw GETs for operation evidence and its sequential rates/histograms. A real Chromium run held the operations response, switched to Services, then released it: both `/api/counter-rates` and `/api/histograms` started **after leaving**. | Complete browser ownership for the four newly admitted legacy GETs. Abort through JSON, cancel server work by query ID, invalidate on owner/scope change, and do not start hidden follow-on reads. Keep publication separate. |
| 2 | Fresh Activity shows observations promptly and keeps controls compact. At 1440×1000, short records still use separate severity, body and action lines, leaving only about seven ordinary rows visible. | Make shared compact record summaries and actions scan densely while retaining full details, UTC tooltips, keyboard access and touch targets. Avoid adding another global control bar. |
| 3 | One fresh local load made 14 API requests, including duplicate measurement-package, example-question and retention-status reads. It loaded 136 native modules and had 3,252 DOM nodes after about 2.2 seconds. Navigation DOMContentLoaded was about 1,035 ms in this single run. | Remove demonstrated duplicate declaration work; measure again. These are local observations, not a performance baseline or an argument for a bundler/framework. Then defer hidden-owner work only where lifecycle/continuation contracts are preserved. |
| 4 | Saved opens with several vertically stacked disclosure/container headers before useful capture cards. Definitions has bounded paging/search, but the test-used volume contains 181 record revisions and 49 native revisions, many with similar titles. | Simplify task-level entrance and make kind/scope/origin discoverable with existing catalog primitives. Do not merge distinct definitions by title, invent revision chronology or delete stored work. |
| 5 | Services/resources has explicit discovery and exact role/unit semantics; generic startup and native interval-to-record continuation are delivered. The older roadmap's top summary still calls several now-shipped resource deliveries planned. | Reconcile status instead of rebuilding dashboards or creating parallel queues. Preserve explicit read windows and unknown coverage; improve shared scope/return gestures before adding unrelated profiles. |
| 6 | Native `Evaluate` is now storage-free and runtime ownership is explicit. Prepared `measurement` versus native `measurements` remains confusing; unchecked `Hash(any)` remains public. | Retain the package/revision cleanup as a finite library task. Inventory real callers and saved identifiers before replacing hashes; do not treat required saved-reference IDs and unused bookkeeping alike. |

The startup requests above include configuration/catalog reads, not fourteen
telemetry queries. The live tail is the only automatic telemetry stream in that
recording. No throughput, allocation reduction or production scan-cost claim is
made from the DOM/module count.

## External comparison and limits

Current public source pages were read on 2026-10-28:

- [Grafana Explore](https://github.com/grafana/grafana/blob/main/docs/sources/visualizations/explore/_index.md)
  puts querying/analysis before dashboard construction and links query management,
  inspector, logs, traces and correlations. Adopt continuity between these tasks
  and a useful optional query inspector; do not copy a full dashboard platform.
- [Datadog Log Explorer](https://github.com/DataDog/documentation/blob/master/hugo/content/en/logs/explorer/_index.md)
  organizes the same workspace around search/filter, grouping, visualization and
  export/reuse, including entry from saved views and contextual links. Goo already
  has much of this; consistent density, scope and return behavior matter more than
  a new editor for each task.
- [ClickStack](https://github.com/ClickHouse/ClickStack/blob/main/README.md)
  emphasizes search/property filtering, live logs/traces, patterns, event deltas,
  dashboards and correlated OTel signals on ClickHouse. Keep Goo's typed reusable
  semantics above that source platform and reuse existing pattern/comparison work.
- Honeycomb's BubbleUp remains a comparison input from the
  [earlier survey](CORE_EXPERIENCE_PRODUCT_SURVEY.md), with its original explicit
  verification limits. Its public docs repository was unavailable in this review;
  no new Honeycomb feature, algorithm or collaboration capability is claimed.

Old Grafana/Datadog document paths had moved; the links above are the successfully
read replacements. GitHub API access was denied by the environment policy. Public
GitHub documentation pages were available; no access policy was changed.

## Architecture and reliability decisions

- Keep passive framework-neutral display elements. The table/record summary
  already serves live/fixed/saved contexts; compactness belongs there, with owners
  retaining selection/navigation/source effects.
- Extend the existing JSON query transport narrowly for admitted GETs. Keep source
  reads, declarations, publication and SSE distinct. Each owner must cancel its
  own active work and prevent follow-ons; a stale-result guard alone is insufficient.
- Keep explicit typed source families and the new pure evaluation API. Do not
  introduce a universal result/store/query hierarchy, runtime exporter ownership
  in the instrumentation library, or global request cancellation that kills saves.
- Maintain exact OTel resource/point identity, units, reset semantics and missing/
  unknown distinctions. Existing dashboard acceptance does not imply all convention
  versions, controller graphs, service health or ingestion completeness.
- Preserve useful immutable captures and authored work. Training export,
  collaboration and optional bitemporal experiments need concrete workflows;
  evidence bookkeeping is not the default product experience.

## Reconciliation and review trigger

The previous roadmap is historical delivery detail. Its C1, bounded C2 resource
slices through 216, C3 and measured C4 slices are not reopened. Library cuts 2/3/7
are implemented; cut 5 has substantial bounded ownership/admission/shutdown work,
and cut 4 has the first pure evaluation path. Remaining package/revision cleanup,
metadata simplification and measured optimizations stay scoped rather than blocking
all product work. Compiler/backend migration, live Databricks, new stores, alert
scheduling, RUM/profiling and shared canvas remain dependency-gated.

Review again after the browser-ownership fix, compact observation pass and first
startup-work reduction, using one real Activity → service/resource → record/detail
→ saved/reusable question journey. Check mobile and keyboard flows, request counts,
late/aborted responses and an unavailable source. That is a better next review
point than another long sequence of tiny specialized capabilities.

## Follow-up at the R1–R3 gate

R1 (217) removes the reproduced hidden operation follow-on reads; R2 (218) gives
short ordinary-OTel rows one scan line while preserving full inspection; R3 (219)
reduces measured startup requests from 14 to 7 without changing the live stream or
module count. Direct workspace links, explicit refresh, isolated generic startup,
keyboard/mobile display checks and bounded Bombadil/observed-entry checks pass.

This supports continuing with Saved/Definitions cohesion rather than introducing
another backend or chart family. The 220 Saved slice puts typed capture rows first,
preserving store order and independent drafts/publication. Definitions still has
separate record/native/application families stacked in one long entrance; making
those families easier to choose is the next bounded opportunity. Mobile navigation
outside Activity still consumes substantial vertical space and deserves a shared
chrome pass, not another Saved-only menu. No new vendor/latency/load claim is made
by this follow-up; the external research and its limits above remain the basis.

## R4/R5 review gate after 226; first L1 cut (227)

This follow-up reviews current source, the 220–226 task checks recorded in
`VALIDATION.md`, and their current screenshots. It does not claim a new vendor
survey, production deployment, browser campaign or latency benchmark.

The bounded scope-coherence gate is accepted:

| Task | Current behavior and basis |
| --- | --- |
| Enter and scope | Activity remains the ordinary-OTel entry. Definitions has explicit families; global defaults appear where consumed, and empty toolbar chrome collapses. 222/223/226 cover direct entry, mobile/keyboard and delayed initialization. |
| Follow and return | Resource origin uses one latest-owner presenter with qualified identity and UTC scope. Native interval return retains its distinct parent chain. 192/224/225 cover exact windows, retained data/focus, stale actions and cancellation. |
| Change a default | Next-discovery lookback does not discard completed service/host/pod observations. Unchanged native fixed re-preview retains nanoseconds; fresh or edited windows are explicit. 224/226 and 55/56 cover these transitions and reusable exact questions. |
| Save and reuse | Saved capture rows precede secondary controls; definitions keep distinct revisions and drafts. Existing store order, independent publication/recovery and receipt formats remain. 110/117/220/221 retain these separate contracts. |

This closes the bounded entry/scope/return queue. It does not establish complete
trace coverage, a universal navigation stack, production isolation, cross-store
search or all OTel profiles. Query-cost presentation stays deferred until usable
source diagnostics exist; returned row counts are not a cost estimate.

Two follow-ups remain concrete. First, the 390px native workbench frame in 226
still stacks secondary save/share/capture actions before results. A bounded action
bar/disclosure pass should improve that layout while keeping Run, preview and
pending/retry state obvious. Native/resource ancestry composition also merits a
real nested task before introducing any shared history abstraction. These are
named UI follow-ups, not reasons to reopen every delivered catalog or dashboard.

The first L1 cut renames the small prepared-fact compiler to `subjectfacts`, using
the actual eight importing files in CLI, lifecycle and prebuilt adapters as its
migration boundary. The native library stays `measurements`. No alias package,
new expression hierarchy, stored-format migration or hash change is introduced.
External-package examples demonstrate general request facts/offline compilation
and supplied native gauge evaluation with a fixed clock. Sample Docker source recipes also include their actual root CLI
package dependencies; isolated copied-source builds check those recipes without
building images. Broader unchecked-hash removal requires its own consumer map.
