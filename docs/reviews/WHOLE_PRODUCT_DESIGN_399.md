# Whole-product design and product review — 399

> Acceptance closure: 399 is validated separately in [the validation record](../../VALIDATION.md).
> [The adopted checkpoint](../ROADMAP_2026_11_10_CHECKPOINT_399.md) reconciles the
> recommendations below. The original independent review scope remains unchanged.

November 10, 2026. Independent read-only review after accepted 397 quantitative
customization and 398 grouped mean-duration reuse. The 399 helper extraction is
underway; this review does not certify its runtime acceptance or adopt a roadmap.
No new browser, source, performance or heavy test campaign was run for this review.

Inputs: the [capability index](../CAPABILITY_GUIDE.md),
[396 reconciliation](ROADMAP_RECONCILIATION_396.md), prior
[design](WHOLE_PRODUCT_DESIGN_396.md) and [frontend](WHOLE_PRODUCT_FRONTEND_396.md)
reviews, current Live/Activity, Services, Measurements, Saved, population and
composition source, and personally inspected actual 384, 390, 397 and 398 frames.
Those frames have their recorded journey scope; a fixed-result frame is not a
fresh measurement of the Live entry screen. The 390 passive component frame is a
synthetic supplied-data fixture, not telemetry evidence.

## Product direction

Keep **see current telemetry → scope → isolate → inspect or compare → preserve
and reuse** as the primary operator loop. Live logs and spans remain the natural
entry, with contextual investigation gestures. A manifest can enrich this loop;
it should not be a prerequisite or a reservation-domain dependency.

397 and 398 materially improve that loop. An accepted quantitative breakdown now
leads to the existing editor without losing its selected measure, units or query.
Grouped duration means now share supplied-data Small multiples, Table and Heatmap
with exact sum/count support. Count and mean are visibly different; an empty
interval is not zero latency. This is useful general-purpose observability depth.
Do not treat the completed work as a reason to add another query engine, editor,
permanent toolbar or universal lifecycle manager.

The useful next improvement is **a small number of reusable views with natural
follow-through**, rather than more chart types. Common source and state semantics
must continue to make a new view easy to trust. 399's dependency cleanup supports
this direction, but establishes no performance improvement by itself.

## What is already present

Live selection and Keep, typed filters and field columns, trace waterfalls,
comparisons and explanations, service/resource dashboards, native instrument
profiles, saved questions and captures, captured-record review and export all
exist. Do not reopen these as missing feature categories. Saved configuration,
a kept occurrence and a fresh capture are intentionally different operator tasks;
keep their short action labels and source/custody distinctions.

The meaningful comparison with established observability products is task-based:

| Familiar product pattern | Goo's current useful path | Relevant next refinement |
| --- | --- | --- |
| HyperDX-style logs/spans exploration | Live, scoping, structured inspection, trace expansion and retained selection | Preserve fast contextual progression; no replacement home dashboard. |
| Honeycomb-style population investigation | Exact questions, descriptive comparisons, explanations and shared scales | Keep comparison assumptions visible; do not imply statistical significance or causality. |
| Datadog/Grafana metrics-to-record investigation | Discovered typed instruments, service-native views and ordinary measurement context | Complete the proven gauge gesture across Services/Saved with correct origin semantics. |
| Grafana-style reusable presentation | Checked declarations, composition and destination compatibility | Make existing compact summaries authorable; improve reading shared wide tables. |

This is a comparison of interaction patterns, not new vendor research or a parity
claim. Alerts, tenancy, presence, arbitrary executable plugins, warehouse execution
and new stores are separate product decisions, not prerequisites to this queue.

## Recommended finite cadence

These are ranked planning inputs, not three concurrent implementations. The
planner should reconcile them with architecture findings and 399 acceptance.

### Breadth: author a compact population summary

**Task:** “Keep the few count, latency or numeric facts that explain this question
beside its records, then use that presentation on another compatible population.”

The gap is concrete: `view-composition.mjs` admits aggregate, records, timeline,
comparison and facet roles, but no summary. `summaryDataset` and
`goo-measure-summary` already exist; `distributions.mjs` renders them outside the
authored composition. Another summary widget or another SQL aggregate is not
needed. The missing part is checked authoring, persistence and reuse.

Start with one optional summary child selecting a short ordered list of fields
already supplied by an ordinary record query's summary dataset. Retain the
records child and one query owner. Count and supported numeric/duration fields
provide a useful first cut; status can join only with its existing count/fraction
support contract. Native, grouped, comparison-summary and arbitrary calculated
fields need not join the initial arm. State the admitted combinations explicitly.

Use the existing compact summary display. Put field choices in folded composition
settings; do not add a second row of authoring controls above every result. The
summary's selected fields and units should explain themselves without a prose
paragraph. Do not render a duplicate fixed summary beside an authored one.

Acceptance: independently checked source summaries; known count zero versus
unavailable numeric/duration values; exact field/method/unit compatibility; a
whole destination question unchanged by Apply; no telemetry reads for local field
choices, Save, reopen or Inspect/Apply; explicit Run; repeated held return with
visible intent equal to saved intent. Validate the complete saved-view carrier in
Go, JavaScript and API schemas as well as the child declaration. 398 exposed why
leaf-only schema checks are insufficient.

### Depth: follow gauge observations from Services and Saved

**Task:** “This resource value looks interesting; inspect nearby records, then
return to the exact metric evidence I was viewing.”

The proven gesture already ships in ordinary Measurements. The remaining source
gap is visible: `service-dashboard.mjs` enables native record-context options for
histograms; `saved-native-record-context.mjs` renders retained measurements without
the source-question/observation options used by the gauge gesture. Reuse the
existing observation planner and child population owner, with separate parent
admission. Do not add another inspector or invent a metric-label join.

Services supplies a live discovered parent; Saved supplies frozen evidence and a
custody receipt. Preserve that distinction. Start with the admitted service-bound
gauge profiles and the existing timestamp-centered, clipped ±60s question.
Explicit resource/environment/namespace policies travel with the plan. Unknown
or stale source state remains visible. A fresh child is time-associated context,
not cause, not full coverage and not part of the saved receipt.

Acceptance: actual service and saved parents; zero and stale/unknown observations;
exact staged bounds and predicates; explicit source execution; source-free repeated
Back; hidden/inert/retired controls refused; the original metric clock, selection
and Saved receipt retained. If the two parents cannot share a finite validated
plan, split delivery rather than generalize their lifetime machinery.

### Consolidation: make shared wide tables readable beyond the first viewport

**Task:** “Read the value and support of this row without losing its group or which
field I am looking at.”

The actual 397 narrow table puts additional quantitative columns outside the
initial viewport. The 398 narrow heatmap shows a useful first interval and part of
the next, but subsequent intervals require horizontal movement. Internal scrolling
is correct and prevents page overflow; discovery and retained context are the
remaining issues. This is more useful now than another chart family.

`goo-table` already supplies a focusable labelled region, and component CSS already
owns horizontal overflow and optional sticky headers. Do not present these as
missing features. Extend the existing wrapper with a restrained overflow cue and
row/header context where it improves these concrete embeddings. A frozen group
column is a candidate, not a decision: prove it does not cover focused cells or
consume the narrow viewport. Keep heatmaps as matrices. Ordinary short tables may
use a priority/stacked variant only when labels, units and row relationships remain
clear.

Acceptance: real ordinary numeric/duration Table, grouped Heatmap and native
bucket Table; standalone and nested composition; keyboard scrolling and focus;
long/Unicode/missing/empty literal labels; narrow/zoom geometry; no clipped active
control or page overflow; exact values and zero/unknown states unchanged; zero
source reads. Before/after frames must show a reading improvement. No global CSS
reset or new table framework belongs in this cut.

## Restraint and follow-up

- Preserve the stream-first entry; new authoring capabilities surface from the
  current evidence and folded settings, not another blank landing workflow.
- Keep exact support, common scales and uncertainty in the data contract. Do not
  reintroduce averaged means or scalar percentiles for visual convenience.
- Keep field-specific compatibility errors close to Apply. A generic disabled
  action without a usable reason is not a successful extension mechanism.
- The passive library entry/gallery remains a valid later developer-experience
  candidate: `display.mjs` does not yet provide the current grammar elements as
  explicit entries. Direct imports work; this does not block the operator tasks
  above or justify a new package/build ecosystem.
- Grouped P95, statistical anomaly claims, arbitrary WYSIWYG kinds and broader
  broker/database health need their own source/support gates. No evidence here
  promotes them ahead of the finite tasks above.
- Do not accumulate screenshots or new verification infrastructure as product
  scope. Reuse current real journeys and keep only the artifacts that explain a
  delivered feature or a meaningful regression.

Source anchors: [composition](../../ui/view-composition.mjs),
[summary element](../../ui/measure-summary.mjs),
[ordinary distributions](../../ui/distributions.mjs),
[service native rendering](../../ui/service-dashboard.mjs),
[Saved native context](../../ui/saved-native-record-context.mjs),
[ordinary observation gesture](../../ui/measurements.mjs),
[table element](../../ui/components.mjs),
[table styles](../../ui/components.css),
[398 actual storyboard](../../journeys/398-grouped-mean-storyboard.md), and
[397 actual storyboard](../../journeys/397-quantitative-customize-storyboard.md).
