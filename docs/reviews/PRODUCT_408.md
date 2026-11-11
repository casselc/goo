# Product review at the 408 checkpoint

Status: read-only product/roadmap review. This review ran no browser, source,
model or performance checks. 406 and 407 are delivered; 408's passive exports and
embedding reference are being validated separately. The sequence below is a
proposal for reconciliation with architecture and design, not another queue.

## Product judgment

Goo already has the essential categories of an observability investigation tool:
live logs/spans, typed scope/filtering, distributions, population explanations,
independent comparisons, trace search/waterfalls, service/resource defaults,
immutable evidence, annotations/review and reusable questions/views. Those are
implemented paths in [the capability index](../CAPABILITY_GUIDE.md), not missing
features to recreate under vendor terminology.

The useful comparison with Honeycomb, Datadog, HyperDX and Grafana is their short
operator loop: notice change, scope the affected work, inspect examples, compare,
and keep a reusable explanation. Goo's strongest distinction is that definitions,
exact scope and evidence meaning survive that loop. Its next investment should
make existing powerful operations natural at an observed service, then improve
one meaningful service lens, then consolidate presentation. Another generic
chart family or new storage engine would contribute less to this loop now.

## Proposed finite sequence

### 409 — breadth: compare captured service records with the previous window

Task: “This service or operation looks different. Show me what changed from the
preceding interval without rebuilding its scope in another editor.”

`ui/service-dashboard.mjs` currently exposes Open widget question, Customize,
Capture, Add question and Use saved template. Its widget-action construction has
no direct previous-window comparison continuation. Meanwhile
[preceding-window comparisons](../PREVIOUS_WINDOW_COMPARISON.md),
[paired windows](../PAIRED_COMPARISON_WINDOWS.md) and
[aggregate-only baselines](../FIXED_BASELINE_READS.md) already execute this task
from authored populations. This is an entry/continuation gap, not a missing
comparison engine.

Add a compact **Compare captured records** action to eligible completed ordinary
single-group record widgets. A checked pure projection should preserve the
entire service/namespace/environment policy, profile predicates, grouping and
selected window, then stage the existing adjacent equal-duration count pair.
It must explicitly select the count measure rather than pretend to compare the
whole service widget. Existing service questions frequently request `spanStatus`;
`ui/count-cohort-context.mjs` deliberately excludes it from this comparison arm.
Omitting an unused projection is acceptable only as an explicit count recipe;
no predicate, population or source observation may be dropped. A native interval,
recorded-error fraction or captured-record rate must not be silently converted
into a supported comparison measure.

Acceptance:

- Service and operation/destination group examples keep exact selectors, including
  empty namespace and effective/current environment precedence.
- Opening stages configuration only. Run independently acquires both operands;
  the ordinary baseline uses its existing aggregate-only path where admitted.
- Both windows, full counts and independent clocks remain visible. No significance,
  causality, application-throughput or health conclusion is inferred.
- Failed/pending/dirty/retired widgets cannot stage an old pair. Back restores the
  held service observations without overview reload. Save/reopen retains the pair.
- Native/specialized or incompatible widgets have no misleading action. No new
  comparison endpoint, evaluator or subscription is added.

Risk: directly reusing a service result's current `spanStatus` question will fail
existing admission. Settle the explicit projection and user label first; do not
relax comparison guards merely to make the button work.

### 410 — depth: reported incoming HTTP response-status context

Task: “Which recorded HTTP responses accompany this server-side population, and
which route or instance should I inspect?”

Incoming HTTP already has a checked Server-span population, route breakdown,
duration and recorded span-status facts. Those facts explicitly do not establish
an HTTP response outcome policy. Outgoing HTTP already has literal response
status, method and port context (`exploration/dashboard.go`'s
`http-client-responses`, and `ui/service-dashboard.mjs:dimensionProfile`), but the
incoming HTTP profile does not offer corresponding response-status context.

Reuse that existing typed-question and grouped-context path for
`attribute:http.response.status_code` within the incoming HTTP profile. Show
literal reported status values with counts, missing and present-empty groups;
noncanonical/custom values remain visible. Keep recorded span Error separate.
Do not introduce a configurable outcome classifier, HTTP success rate, inferred
health, legacy alias fallback, deduplication or fleet percentile. If response
classes are wanted later, they require their own explicit typed transform and
policy; a renderer must not calculate them from labels.

Acceptance:

- A fresh standard OTel bank includes Server and Client spans, missing/empty/
  noncanonical codes, a reported 500 with non-Error span status, and the converse.
  Current profile predicates and complete group accounting determine membership.
- Context and drilldown preserve service/namespace/environment/window and the
  literal qualified value. An absent dimension is unknown/unreported context,
  not a successful response population.
- The response question can use the 409 count comparison path where compatible,
  giving a real second use of the general continuation.
- Native HTTP duration retains independent stream identity and clocks. No join
  between native counts, span records and HTTP response values is asserted.
- New context stays under existing widget/context controls; it adds no top-level
  destination or automatically expanded wall of cards.

Risk: this is a response-observation lens, not full RED. Label it accordingly.
Pin the current supported OTel convention rather than inferring semantics from
an arbitrary similarly named attribute.

### 411 — consolidation: one compact chart frame across supplied displays

Task: “Keep the plot and its scope readable when I switch view or return to a
held investigation.”

The actual 407 design review records nested chart borders/headings and weak
horizontal-scroll affordance at 390 px. Recent 406/407 integration also exposed
how easy it is to confuse local presentation renewal, successful async completion
and renewal of source authority. Consolidate these through existing component
and owner patterns, rather than adding another manager or renderer.

Use the shared chart-frame density/disclosure primitives to eliminate repeated
headings and borders in representative fixed numeric, native and Saved plots.
Keep one actual plot title, compact units/scope/state and disclosed exact read
facts. Add a visible, accessible horizontal-overflow affordance where categorical
bins cannot fit; preserve exact original bins and keyboard navigation. Local
Saved highlighting remains source-free; held displays retain local controls while
source actions remain paused until their existing explicit Run.

Acceptance uses matched actual desktop/390 px frames, first-plot geometry,
keyboard/focus and held-return checks, not a reduction in DOM nodes alone. No
query, denominator, completeness guarantee, evidence clock or export changes.
Prefer a narrow shared-component/adapter cleanup justified by these concrete
surfaces over a broad rewrite of `population-workspace.mjs`.

## Reconcile the old candidates

| Candidate from older reviews | Current decision |
| --- | --- |
| Save/reopen trace-search recipes and captured search witnesses (381) | Delivered, including continuation in 403. Remove from proposed execution queues; preserve history links. |
| Log-pattern display admission/reuse (381) | Delivered in 391. Do not propose clustering or patterns as missing. |
| Fixed/native first-data visibility (381) | 384/390 delivered the broad cuts. Retain only demonstrated remaining chrome/overflow issues for 411; no fresh whole-app redesign prerequisite. |
| Generic summary/rate and authored numeric histogram grammar | 400/404/406/407 delivered. 408 closes public entry/examples; it is not a new interpreter. |
| Review-task chooser and exact kept-occurrence review | Still real, source-confirmed gaps: `capture-record-review.mjs` requires exact task ID, backend has create/exact reads only; `kept-saved.mjs` does not mount review. Keep as the next annotation/training cadence candidate, not forgotten or relabeled delivered. |
| Grouped P95 | Defer until raw-cell/population quantile support is explicit. Existing scalar P95 values are not composable percentiles. |
| Exponential histogram execution | Defer under source/export and mathematical readiness gates. Discovery already exposes unsupported instruments. |
| Host sample context | Defer; existing full-window links are delivered. A narrower association needs a unique identity and concrete user need. |
| Multiplayer, tenancy, alert delivery, arbitrary executable kinds/WYSIWYG, new stores and warehouse execution | Separate initiatives. None is a prerequisite for this three-step investigation cadence. |

## Lessons to retain

A declaration's validity is not evidence that its complete saved/HTTP carrier or
its actual interaction works. Keep whole-carrier checks and real visible-control
journeys, but reuse independent source oracles and avoid an ever larger generic
“proof platform” for each small display change. Recent tests caught concrete
resolver, async-completion and held-renewal mistakes; those are useful regression
cases, not reasons for another source or navigation architecture.

Prerelease status allows simpler contracts and removal of stale formats. It does
not justify losing a working passive Saved interaction while tightening a held
source-action guard. Keep user-facing outcomes and current scope explicit, and
add advanced machinery only when it improves the task's reliability or usability.
