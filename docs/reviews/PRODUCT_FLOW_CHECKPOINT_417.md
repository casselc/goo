# Product-flow checkpoint after 417

Status: reviewed proposal, reconciled into the [adopted 417 checkpoint](../ROADMAP_2026_11_13_CHECKPOINT_417.md). This independent
review read the current capability/gap index, feature contracts and application
source, and inspected the actual Activity desktop, log-volume mobile and version
comparison mobile frames. It ran no new browser or vendor campaign. Milestone
417 execution acceptance belongs to its own validation record.

## Product direction

Goo's useful center remains **current telemetry → narrow a population → inspect
or compare → preserve the question or evidence**. The stream should remain the
initial useful content. Services, Measurements and Saved are continuations of
that loop, not separate collections of demonstration features.

The recent work materially improves that loop: Service triage helps choose an
entry; log volume provides a full-scope clock beside the bounded buffer; scope
recall separates configuration from held results; version cohorts and exact
exception groups offer useful deeper questions. The latest exception polish is
a good pattern: compact active context, explicit acquisition, real data, then
native disclosures for configuration and exact inputs. Do not restore large
setup cards or disclaimer paragraphs ahead of data.

The largest near-term opportunities are connections between existing capabilities,
not new top-level destinations, storage engines or chart libraries.

## Recommended next finite cycle

| Slot | Operator gap | Natural interaction | Reuse |
| --- | --- | --- | --- |
| 418 — breadth | Ordinary comparisons show captured counts even when a saved baseline and selected question cover different durations. The larger bar can simply represent longer exposure. | In the existing comparison's measure control, choose **Captured records/s**. Keep aligned bars/table, exact supporting counts and window durations, and each side's scope. Changing this measure is local. | Existing `record.rate` rational count/exposure model, captured-count comparison carrier, shared paired comparison display and exact-value disclosure. |
| 419 — depth | The guided version flow answers count and mean-duration questions but cannot directly answer “did this version report a larger Error share?” | From the already selected version pair, choose **Recorded Error share** beside count/mean. Show paired fractions and captured numerator/denominator; selecting a side can stage that side's exact Error records. | Existing span-status summaries, exact version cohorts, checked population comparison and held service return. |
| 420 — consolidate | Contextual actions have coherent underlying ownership but inconsistent reachability and test patterns as controls move into inspectors/disclosures. This increases interaction fragility and makes features harder to discover. | Keep one visible primary action, native **More** for secondary actions, and restore focus to the actual opener. Open the relevant editor/disclosure before modifying or saving a question. | Existing `nativeControlReady`, `completionControlReady`, `actionVisible`, `revealFocus`, shared action menu and inspector. Adopt these patterns; do not invent an action registry or navigation manager. |

### 418: fair exposure before more comparison dimensions

The existing [captured-rate contract](../CAPTURED_RECORD_RATES.md) explicitly
excludes ordinary rate comparisons, while
[log-pattern time comparison](../LOG_PATTERN_TIME_COMPARISONS.md) already proves
the usefulness and semantics of unequal-window rates. That specialized feature
is delivered; this proposal connects the same idea to ordinary log/span count
comparisons and their display grammar.

The smallest slice should start with the existing captured-baseline comparison
path, whose compatibility rule already allows different fixed windows. Do not
relax the fresh query comparison's deliberate same-window/adjacent equal-window
rule merely to add this display measure. Each side uses its declared queried
window as exposure, even if matching records occupy only a small part of it.
Reuse exact rational arithmetic; avoid renderer-local floating point division.
A known zero count over positive time is a known zero rate. Missing/unavailable
counts or exposure do not become zero. This remains captured records per second,
not request throughput, sampling correction, anomaly significance or health.

Acceptance should include unequal windows with equal rates but different counts,
one side with a known zero, unavailable data, exact counts beyond JavaScript's
safe integer range, saved/reused display configuration, local measure changes,
explicit source execution, and held Back. Matched desktop/mobile frames must make
the two values comparable to a human without reading a provenance paragraph.
No new SQL, ingestion profile or evaluator should be required for this slice.

### 419: the next question after choosing two versions

[Version comparison](../SERVICE_VERSION_COMPARISONS.md) explicitly limits its
first carrier to count and mean duration. Adding recorded Error share is a useful
bounded depth improvement to a current entry, rather than another comparison
workspace. Preserve the qualified service, environment policy, original operation
predicate/grouping, fixed window and literal selected version values.

Use reported span Error evidence with an explicit captured-span denominator.
Unknown/unset status must stay distinguishable in supporting facts; the complement
of Error share is not a success or health claim. A zero denominator gives an
unavailable fraction. Keep each side's independent read basis. Do not average
fractions across operations, derive a percentile, infer a deployment cause or
expand scope from version metadata. Any drill should add an exact status predicate
to one side's original question and stage without reading.

Acceptance should include different operation mixtures, zero records, unknown
status, identical Error shares with different counts, per-operation conservation,
retired selection controls and no-read held return. The same supplied comparison
component must handle count, mean and fraction with appropriate unit/support.

### 420: action consistency as a workflow contract

The regression work around 416/its polish found old drivers looking for trace
expansion inside a record row after it had moved to the shared selected-item
inspector; expecting Back to focus a menu summary instead of the actual opener;
and trying comparison/package controls inside closed native disclosures. These
are evidence of interaction surfaces evolving, not by themselves proof that the
product's current source semantics are broken.

Choose a finite set of representative paths: Activity row → inspector → trace →
Back; service comparison → records → held service; fixed question → comparison
context; and fixed question → Edit → reusable package. Align visible naming and
use existing readiness/focus helpers at those boundaries. Add browser assertions
that follow the actual native disclosure ancestry, including Escape/close,
hidden retained callbacks, and return after an origin row leaves the viewport.
Avoid bespoke convenience helpers that click hidden controls or set `open=true`
without testing the gesture. Preserve source owners and staged/applied distinction.
The deliverable is fewer inconsistent paths and focused conformance tests, not a
new generic command framework.

## Capabilities not missing

- Activity already has full-scope counted field values. `FilterValues.mount`
  pauses the actual stream and freezes its completed window; choosing a value
  stages a condition around the full existing selector. A **Facet by…** entry or
  clearer include/exclude gestures may be useful later, but this would improve
  discovery of an existing capability. Do not schedule a second facet engine.
- Activity log-volume bars already support pointer/keyboard range selection and
  exact interval explanation. Span latency, fixed distributions, waterfalls,
  heatmaps, shared tables and attribute inspection also exist. Choose concrete
  workflow gaps before adding another chart type.
- Native measurements, automatic OTel service/resource profiles, view authoring,
  immutable captures and review labels are already substantial capabilities.
  Their deeper or broader arms should remain evidence-driven rather than being
  described as entirely missing product categories.

Monitoring/alert delivery, authenticated shared deployment, multiplayer presence,
arbitrary executable kinds, broader database/broker health and new storage/warehouse
backends remain larger gated initiatives. They should not displace this finite
operator-focused cycle without an explicit change of product priority.

## Immediate documentation reconciliation

The capability index says it is reconciled through 416 but its lower summary
still calls 416 “Next.” Update that and the active checkpoint link as part of the
417 delivery. Preserve historical proposals as history. This is routine delivery
hygiene, not another roadmap milestone or reason to postpone operator value.
