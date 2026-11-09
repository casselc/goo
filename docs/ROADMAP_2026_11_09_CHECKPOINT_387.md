# Roadmap after the 387 checkpoint

Adopted November 9, 2026 after the 387 runtime and whole-product checkpoint.
This replaces the completed [384 checkpoint queue](ROADMAP_2026_11_09_CHECKPOINT_384.md).
The [current capability index](CAPABILITY_GUIDE.md) summarizes delivered behavior;
[validation](../VALIDATION.md) records actual checks. This is one finite next
queue, not parallel execution plans.

Inputs: independent [whole-product design](reviews/WHOLE_PRODUCT_DESIGN_387.md)
and [architecture](reviews/WHOLE_PRODUCT_ARCHITECTURE_387.md) reviews, the
[current capability index](CAPABILITY_GUIDE.md), source inspection, and
[validation](../VALIDATION.md). The reviews and older plans are inputs; this
adopted queue owns execution order.

## Product direction and lessons

Keep the main loop **current telemetry → isolate a population → follow evidence
→ preserve or reuse the question**. Ordinary OTel is sufficient; manifests are
optional. Existing Live Activity, filters/facets, traces/waterfalls, comparisons,
explanations, service/resource observations, review labels and scoped display
kinds are delivered capabilities, not missing product categories to recreate.

385 added an outgoing HTTP observation lens using ordinary typed record questions.
386 extended one native display owner to gauge samples while retaining histogram
interval semantics. [387 capture admission](CAPTURE_FAMILY_ADMISSION.md) closes six
capture envelopes without replacing family evidence validation. Receipt-only
references, nested null data, source clocks and immutable custody stay distinct.

The adopted cadence is **388 breadth (delivered) → 389 facet depth (next)
→ 390 consolidation**. Finish
one acceptance cut before implementing the next. This prerelease has no legacy
route, format or API floor requiring aliases or compatibility machinery. Use the
current Go 1.27.3 runtime and standard-library facilities where they simplify a
touched path; no speculative runtime upgrade or type-system project is needed.

Recent checks also change how work is accepted:

- Metadata, authored configuration, full evidence and execution authority remain
  separate. Receipt summaries cannot acquire result or paging authority.
- Keep accepted evidence and its page address separate from pending/failed request
  progress. Validate requested ID and cursor before adoption; refused loads retain
  the old evidence, citation and comparison capability.
- Source-free staging from accepted held evidence is a distinct capability from
  native source execution. Back can retain the unchanged parent and allow another
  nearby question while source actions stay paused. Edits, Apply, new measurement
  work or result replacement retire that capability; hidden owners cannot stage.
- Missing/unknown values must remain distinct from recorded zero in both model
  and paint. A valid model with a misleading geometric position is insufficient.
- Test physical coordinates under real application styles. The unavailable rail
  and numeric timeline must share the same time axis; equal percentage attributes
  did not prove that when global section padding changed their containing boxes.
- Observe completed UI outcomes rather than clicks or an actor phase alone. Keep
  actual source/population admission separate from model replay and screenshots.
- Collector acceptance is not source settlement. Confirm the physical source bank
  before claiming a successful telemetry journey. Bound disposable build caches
  and diagnostics while preserving lab stores, volumes and retained evidence.

## 388 — breadth: records around a gauge observation

Delivered in ordinary Measurements: [contract](GAUGE_OBSERVATION_CONTEXT.md) and
[actual checks](../VALIDATION.md). The adopted fixed radius is ±60 seconds, clipped
to the original exact bounds. Saved/service continuation remains deferred; the
remaining active queue starts at 389.

User task: “This observation is unusual or unavailable; what logs or spans were
reported nearby by this instance?”

Before 388, the [native record-context contract](NATIVE_RECORD_CONTEXT.md)
offered counter and histogram interval follow-through, while the UI suppressed
gauge context actions. `PlanNativeRecordContext` still represents a gauge only as
`[atNS, atNS+1)`; 388 adds a distinct surrounding-observation projection rather
than relaxing that contract or silently treating a point as a measurement interval.

- Add a distinct observation-context projection with an explicit, documented
  fixed radius around a selected reported gauge timestamp. Clip the resulting
  half-open interval to the original exact native question bounds; never use the
  current clock or expand beyond the held question implicitly.
- Begin in ordinary Measurements with validated per-stream service gauges.
  Saved/service surfaces may adopt the same projection only after their explicit
  retained-evidence admission is covered; they are bounded follow-ups if that
  would widen this cut. Preserve selected service,
  instance, missing/empty/named namespace and environment policy. Metric point
  attributes, instrumentation/schema identity and unrelated resource attributes
  must not become record filters merely because they appear on the metric.
- Present actions as **logs/spans around this observation** and explain time
  association. Do not infer cause, exemplars, continuous gauge coverage or trace
  completeness. Unknown/no-recorded-value observations may supply a reported
  timestamp, but cannot supply a fabricated magnitude or timestamp.
- Prove that the selected observation belongs to the admitted exact retained
  stream before offering held/captured continuation. Preserve receipt, source
  basis, evaluation window, selected point and parent evidence.
- Stage an ordinary typed record question without telemetry. Explicit Run reads
  records; held return restores the native parent without source/store rereads.
  Keep existing interval and reported-exemplar planners strict and separate.

Acceptance includes independent integer-nanosecond membership and clipping
oracles; boundary points, missing/empty namespaces and foreign-instance/env
controls; reported stale/unknown samples; current fixed evidence and measured
held-parent membership; exact staged requests; explicit read/refusal/retry and no-read return. Real logs
and spans inside/outside the chosen radius must distinguish the new context from
both a one-nanosecond point and a whole-window query. Refresh current wide/narrow
frames and executable Playwright/Bombadil storyboards, with generated properties
and measured Quint replay plus negative controls.

Risk: an arbitrary lookaround can look like causal correlation. Keep radius,
clipping and omitted metric-only selectors visible. Do not broaden resource
bindings, synthesize missing observation times or relax interval admission to
make this work.

## 389 — depth: reuse evaluated counter displays

User task: “Use this useful counter-rate timeline or table on another compatible
counter question without inheriting its reset or freshness policy.”

[Counter evaluation](NATIVE_COUNTERS.md) and interval presentation already exist.
[NativeView](NATIVE_GAUGE_VIEW_REUSE.md) currently admits only gauge and histogram
configuration; `nativeViewQuestion` rejects counters and `counterDomain`.

- Extend the same finite native-display contract and owner with a closed counter
  timeline/table arm over evaluated intervals/rates. Do not coerce intervals into
  gauge samples, recompute rates from displayed totals or create another store.
- Define applicability against the supported counter grain, exact literal unit,
  declared integer/decimal domain and evaluated rate/reset semantics before
  implementation. The destination retains its entire question and policy.
  Donor `maxGap`, source identity, values, selected intervals, clocks and custody
  must not transfer.
- Preserve resets, epochs, gaps, stale/unknown/unavailable states and exact-rate
  information supported by the existing evaluator. No interpolation, missing-to-
  zero conversion, fleet sum or scalar-percentile aggregation is introduced.
- Browse/Inspect/Apply remain configuration operations without telemetry. Apply
  clears the prior result/selection and waits for explicit Run. Reject unsupported
  families, units, grains or counter policies before applying.

Acceptance includes immutable configuration round trips and byte/hash checks;
generated compatible/refused applications across integer/decimal and reset/gap
cases; exact destination request/calculation identity; honest supplied-data
rendering; and an actual save → different destination → Apply → Run journey.
Preserve existing gauge/histogram reuse through focused regressions. Use the
current shared component/model verification stack, not a new counter-view engine.

Risk: equal units alone do not establish compatible rate semantics. Finalize the
smallest explicit applicability contract in design/architecture review before
adding the counter arm; do not add universal policy descriptors speculatively.

## 390 — consolidate: data-first native entry and component boundaries

User task: “Show the observation first, with its identity and uncertainty clear;
let me inspect full provenance or save/reuse controls when I need them.”

[384 data-first fixed results](DATA_FIRST_FIXED_RESULTS.md) covered checked record
surfaces. It did not establish a natural-top contract for native measurements.
386 frames still spend substantial space on identity, diagnostics and save
controls before the primary visual. Broad application `section` styling also
changed reusable timeline geometry. Ordinary widget links and Customize/native
contexts expose different return affordances.

- Reuse a compact native identity/status summary and one natural disclosure for
  full stream identity, diagnostic explanation and configuration. Keep active
  errors, unknown/stale state, source refusal and fixed-window freshness visible;
  do not hide uncertainty to gain density.
- Put the selected native visual or first observation in view at natural initial
  scroll position on desktop and 390-pixel screens. Measure the real scroll
  container and actual geometry; forced scrolling is inspection, not entry proof.
- Bound component chrome explicitly and scope application selectors to their
  owned children. Verify standalone/server-rendered and nested SPA composition,
  numeric/unavailable time-axis alignment, small/large labels and keyboard focus.
  Do not switch frameworks or rewrite the entire stylesheet.
- Reconcile existing record/native return affordances through the smallest shared
  delegation contract that preserves held ownership and focus. Reuse current
  question sessions and retained visits; do not introduce another navigation
  manager, hidden rerun or generalized generation-counter abstraction.
- Remove a duplicate helper only when behavior and retirement policy actually
  match. Preserve reads, cancellation, immutable custody and source-family limits
  in their existing owners.

Acceptance is a before/after usability and geometry baseline across live native,
service-expanded native and saved native surfaces, including empty/unknown/refused
states. Existing plots, local kind selection, staged context, explicit Run and
retained return must keep their exact scopes and no-read budgets. Keep actual
frames, accessibility checks, storyboards and actor invariants current.

Risk: compacting prose can erase essential scope or failure information. Separate
passive presentation consolidation from read/session authority and validate both
standalone and composed components before removing old layout code.

## Reconciled deferred work and checkpoint

Contextual review-task discovery → verified kept-occurrence review remains a
coherent later pair; captured-record review, labels and export already exist. It
is not a prerequisite to this core-observability queue and is not silently removed.
Checked log-pattern display admission and native HTTP-client duration adoption are
other bounded later alternatives. Existing comparisons, histogram evaluation and
service discovery should not be reimplemented under those labels.

The parameter-escaping duplication identified in architecture review is an
investigation candidate, not a fourth active milestone. First establish a real
correctness/performance difference with literal values and source tests; then
share only the matching mechanics, preserving family source limits and failures.
Do not create a universal query/transport layer to resolve it.

General exponential-histogram execution retains its
[separate source/mathematical gate](EXPONENTIAL_HISTOGRAM_READINESS.md). Broad DB or
broker server health, authenticated tenancy/presence, alert delivery, warehouse
lowering/new stores and arbitrary contributed executable kinds/WYSIWYG authoring
remain wider initiatives. They are not parallel implementation queues.

Reassess the whole product after 390, or earlier if actual usage invalidates this
sequence. Update this one queue from evidence rather than append plans indefinitely.
