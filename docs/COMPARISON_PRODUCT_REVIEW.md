# Comparison product and usability review

Review date: 2026-10-20. This is a design review, not a claim that its proposals
are implemented. It examines `ui/evidence-comparison.js`, `ui/fresh-comparison.js`,
the exploration/capture contracts, and the generated [fresh comparison frame](journeys/screenshots/57b-current-evidence-and-stored-baseline.png).
No new browser campaign was run for this review; screenshot findings refer to the
last recorded journey, not a newly observed deployment.

## Product essence

Goo helps an engineer turn observed telemetry into a question, retain the evidence,
explain a change, and share a finding someone else can inspect. The useful next
step is **locating a change within the observed population**, using the same
question, capture, comparison and investigation primitives.

The interface should answer these questions in order:

1. What changed between these two scopes?
2. Which observed groups account for the difference?
3. Can I inspect or pursue one group?
4. What evidence and limitations support this finding?

## Existing strengths and friction

Implemented: saved fixed questions, explicitly previewed fresh scopes, immutable
captures, stored comparisons, source-free sharing, and reusable question cards.
These already form one coherent investigation workflow. Exact native boundaries,
missing coverage and retained IDs make its conclusions inspectable.

Observed in frame 57b: the first result table follows two expanded provenance
cards, capture selectors and the fresh preview. The engineer scans roughly 800
pixels of metadata inside the comparison before seeing its numeric result.
The table displays mean duration as `20000000/2`, p95 in nanoseconds, and error
share as `1/2`; this preserves arithmetic but slows ordinary interpretation.
Nested panels repeat the fresh scope, comparison scope and capture identity.
Adding another independent feature panel would compound that friction.

## Recommendations for this milestone

All recommendations below are proposed at review time. The implementation and
validation record must identify which are delivered in the current round.

| Priority | Decision | Acceptance condition |
| --- | --- | --- |
| P0 | Compare full stored group aggregates, never the first 1,000 inspection rows. | A capture with more than 1,000 records yields the same group totals regardless of its retained page. Unsupported or missing aggregates produce an explanation, not an inferred total. |
| P0 | Expose the captured breakdown before suggesting other dimensions. | Show `Breakdown: service` or the qualified attribute path. A dimension absent from either capture is unavailable. Any future new grouping requires a clearly previewed new question and explicit capture. |
| P0 | Align the union of groups and preserve absence. | Missing attribute, present-empty value, absent group, zero records, no data and failed evidence remain distinct. No percentage or latency is invented for an empty denominator. |
| P0 | Follow a group into a typed question. | `Explore comparison group` stages that side's scope and the original predicate AND the selected group; it makes no implicit source read. Missing dimensions become a missing predicate. Existing filters, including OR expressions, remain intact. |
| P1 | Put the result before full custody details. | A compact baseline/comparison scope line, result summary and group table precede an expandable evidence-basis section. Full IDs and exact clocks remain accessible. Critical incompatibility/unknown warnings remain visible. |
| P1 | Render human units while retaining exact values. | Display `10 ms → 100 ms`, `+90 ms`, and `50% → 25% (−25 pp)`. Preserve exact source strings in an accessible disclosure. Do not convert exact counts through unsafe JavaScript numbers. |
| P1 | Give the breakdown one clear table and a next action. | The table has group, baseline, comparison and change columns; volume is available for both signals, with error share/duration only when actually captured and meaningful. A selected group has explicit baseline/comparison follow-up actions. |
| P1 | Preserve the explanation on sharing. | Saved comparison reopening keeps its groups, states and chosen follow-up scope source-free. The recipient sees the same evidence, rather than a silently rerun breakdown. |

Recommend count change as the first simple ordering, clearly labelled, rather
than a generalized "most anomalous" rank. If another ordering is offered, identify
its measure. New groups can rank highly because volume changed; that does not
establish causality, a deployment regression, or statistical significance. The
sum of group count differences can explain the total count change only for a
complete partition. Group percentiles are not additive contributions to a fleet
percentile.

For logs, use **records** and reported severity, not request error rate. For spans,
say **spans** and reported error status, not unique requests. Follow the actual
OTel field namespace: `resource.service.version`, `resource.cloud.region` and
`attribute.http.route` must not collapse into ambiguous bare keys. Suggested
dimensions should come from retained inventory, with resource and record fields
visibly distinguished; an OTel recommendation does not prove the field exists.

## Proposed interaction storyboard

An engineer opens a saved HTTP-span question and explicitly compares a fresh
window against prior evidence. The summary shows volume, error share and latency
with both scopes. Below it, the captured route breakdown shows `/checkout`
increasing while another route remains stable. The engineer chooses **Explore
comparison group**, sees the selected side's fixed scope and filter, and runs it
explicitly. They inspect a trace, retain the resulting evidence and save a finding
in the same investigation. A colleague reopens the finding without telemetry
reads and can independently open either side as a question.

The log variant uses a service or resource-region breakdown, preserves records
with no such attribute, and opens that missing group with a missing predicate.
An older capture without grouped aggregates explains that it cannot supply this
breakdown; it remains usable for its existing overall comparison. These are
proposed storyboards until executable journeys and generated frames exist.

## Verification required for implementation

Use a live OTLP journey with more than the inspection-row limit, multiple groups,
a missing attribute, a present-empty value, and a group appearing on one side.
Assert complete totals and unchanged original capture bytes; group follow-up
stages a question without a source read and preserves existing AND/OR semantics.
Save and reopen in a fresh browser. Cover log labels, unsupported older captures,
incompatible grouping, source inconsistency, and 390-pixel containment.

Useful properties: input immutability, group-key identity, order independence,
partition count reconciliation, exact arithmetic, and one-sided denominator
handling. Stateful checks should cover repeated group selection, changed active
investigation, failed publication/retry, and source-free saved reopening. Bombadil
actions should select groups and stage follow-ups with invariants that prevent
implicit reads and cross-document mutation. Quint observed replay should reject
mutations that change a selected group, scope, source-read count or frozen result.

## Later coherence work

After this bounded milestone, consolidate the investigation entry points around
three verbs: **Explore**, **Capture**, **Save finding**. Clarify when an action
stages a draft versus publishes a revision; avoid adding a competing capture or
comparison library. A compact investigation outline could expose question →
evidence → finding relationships without displaying every receipt at once.
Automatic dimension discovery across old captures, causal explanations,
statistical significance and fleet-native regrouping are separate future work.

## Implementation follow-through

[Captured breakdowns](COMPARISON_BREAKDOWNS.md) implement full-group reuse, explicit
recapture for other dimensions, missing/empty/one-sided semantics, fixed-side
focused questions, readable group means and results before receipt disclosure.
The recommendations above remain the review snapshot; broader dashboard and
navigation proposals are tracked in [whole-product review](WHOLE_PRODUCT_REVIEW.md).

The final source/frame review found no critical correctness issue in the bounded
implementation. Two follow-ups remain: provide error/latency ranking alongside
explicit population-share ranking (a large error regression can have a small
share shift), and hide span-only measure columns for logs instead of repeating
Unavailable. These are proposed polish, not implemented controls.

## Visual follow-through

The [visual comparison pass](VISUAL_COMPARISONS.md) adds paired bars on common
zero-based scales, equal-width group tracks and aligned native histogram bins.
Scope and arithmetic remain inspectable in disclosures. Updated executable
frames 35, 52, 54 and 58 show the delivered behavior; this supplements the
historical screenshot observations above. Signed/native scientific-number formats
still retain exact tables rather than an invented chart value.
