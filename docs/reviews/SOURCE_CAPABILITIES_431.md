# Signal availability and partial discovery — 431 acceptance

[Contract](../SOURCE_CAPABILITIES.md) · [Storyboard](../journeys/source-capabilities.md).

## Architecture/data review

An independent OTel/data reviewer accepted the fixed-family SQL boundary,
readiness/presence separation, header-only missing-source classification,
cancellation, partial-family inventory and refusal when no source is usable.
The review found one real mismatch: readiness accepted numeric `UInt8` monotonic
flags while discovery decodes JSON booleans. The profile now accepts only
`Bool`/`Boolean`; regression vectors reject `UInt8`. Follow-up source review
confirmed the correction and accepted the bounded design.

Metric discovery compatibility remains distinct from evaluation eligibility.
Missing metadata and permission denial never establish table absence. A later
catalog read failure does not silently remove that family and publish the rest.

## UX follow-up

The widget author's integration review accepted the shared Find data disclosure,
explicit local signal selection, preserved filters and focus return. This is not
presented as independent review of that author's own module.

The independent UX reviewer checked source and real generated frames. The first
narrow frames revealed a native table-caption sizing failure: captions wrapped one
character at a time despite no horizontal page overflow. The correction gives the
caption an explicit block width in the stacked mobile table. The journey now
measures caption width, height and its gap to the first row, in addition to page
containment. Metric state labels also say Points observed / No exported points
in window instead of using log/span terminology.

Independent follow-up accepted the refreshed logs/spans 390/320 frames and
gauge390 frame: captions use the available width, states/actions are readable and
metric wording matches exported points. No remaining bounded UX blocker was found.

Final executed checks are recorded in
[VALIDATION.md](../../VALIDATION.md).

## Scope and evidence limits

The partial-source journey uses real ClickHouse databases cloned from the lab's
exporter schemas and deterministic exporter-shaped inserts. The separate generic
and unknown-data regressions send real OTLP through the existing collector.
Denied metadata, malformed source replies and post-check failures are verified by
structured transport tests; this milestone does not claim a real restricted-user
ClickHouse deployment test.

Hegel covers generated readiness classifications and cancellation. Native Quint
validates observed readiness-probe transitions with negative controls; it does not
model the browser or imply that presence establishes a complete population.
Bombadil independently observes capability results, local selection, exact scope,
records and keyboard Pause before declaring journey completion.

Metrics-only environment/service bootstrap and sparse/delayed event recovery
remain 432; whole-product discovery/recovery consolidation remains 433.
