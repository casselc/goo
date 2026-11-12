# Roadmap after the 408 checkpoint

The [405 cadence](ROADMAP_2026_11_11_CHECKPOINT_405.md) delivered captured-record
rates, reusable numeric histograms and the passive display entry/reference.
[Validation](../VALIDATION.md) distinguishes actual checks from proposals.
The [review reconciliation](reviews/ROADMAP_RECONCILIATION_408.md) records the
architecture, design and product tradeoffs behind this next finite queue.

Keep the loop: **current ordinary telemetry → narrow a population → inspect or
compare → preserve/reuse**. Manifests remain optional. This prerelease needs no
legacy formats or compatibility layer. Use the current Go toolchain (1.27.3 at
this checkpoint); choose newer facilities for a demonstrated benefit, not novelty.

## 409 — breadth: compare captured records from Services

Implemented: [service-window comparisons](SERVICE_WINDOW_COMPARISONS.md) now
stage the existing fixed pair from an accepted widget, preserving exact population
scope and the held service return. See the executable
[journey](../journeys/409-service-count-comparison-storyboard.md) and [validation](../VALIDATION.md).
410 and 411 are also implemented below. The [411 checkpoint](ROADMAP_2026_11_12_CHECKPOINT_411.md) now owns the next finite queue.

User task: “Has this service's captured activity changed from the preceding
window?” Add a contextual action to supported ordinary log/span widgets that
stages the existing preceding-window comparison. Reuse the existing comparison
owner, planner and explicit Run. No new query engine, comparison store or native
measurement comparison is required.

Call the action **Compare captured records**. A checked count projection may omit
unused measure directives such as recorded span-status summaries; preserve every
population predicate, qualified grouping, environment/namespace policy and page
limit. Make this intentional projection explicit rather than promising comparison
of every widget measure. Preserve the selected fixed window; derive the previous
half-open window through the existing checked recipe. Refuse invalid/underflowing
windows and unsupported numeric-measure, histogram, reported-edge, grouped-time
or native profiles. Retain existing single-category-group, page-zero admission;
do not invent grouping or reset a continuation. Supported typed numeric predicates
inside `where` remain population filters and must stay intact.

Acceptance: actual discovered-service entry, stage without telemetry reads,
independent selected/previous-window reads after Run, exact raw membership and
window oracles, unchanged service parent and source-free held Back. Pending,
retired and foreign-widget actions must refuse. Share controls/components and
reuse existing semantic proofs; do not add another acquisition owner.

## 410 — depth: incoming HTTP response context

Implemented: [incoming HTTP response context](INCOMING_HTTP_RESPONSES.md) adds
one independent literal-response grouping under the existing current Server-span
population, with guarded record continuations and held return. See the executable
[journey](../journeys/410-http-server-responses-storyboard.md) and [validation](../VALIDATION.md).
411 quantitative view consolidation and the whole-product review are delivered below.

User task: “Which reported response codes make up this service's incoming HTTP
population?” Adapt the outgoing HTTP context pattern to the finite current OTel
HTTP-server profile. Expose literal `http.response.status_code` populations and
natural record continuations under the accepted service scope.

Keep reported response status separate from recorded span Error. Missing, empty
and unexpected literal values remain explicit. Do not infer availability,
success, request identity, an SLI denominator or backend failure from these
counts. Use qualified current semantic-convention fields; unsupported/legacy
profiles remain visible boundaries rather than hidden aliases.

Acceptance: discovered ordinary HTTP-server spans, foreign service/environment
controls, literal and missing/empty code groups, independent full counts despite
record/group presentation limits, current-owner filter/drill/save/capture paths,
and held Back. Reuse source adapters, existing grouping/selection grammar and
shared attribute/record displays. Keep labels compact and avoid another wizard.

## 411 — consolidation: compact quantitative view composition

Implemented: [compact histogram composition](QUANTITATIVE_VIEW_COMPOSITION.md)
retains authored titles and exact values, adds named native scrolling and shares
invocation/completion readiness without renewing held source authority. The
[actual storyboard](journeys/411-quantitative-chrome.md) and [validation](../VALIDATION.md)
record acceptance. [Whole-product review](reviews/ROADMAP_RECONCILIATION_411.md)
adopts the next [412–414 queue](ROADMAP_2026_11_12_CHECKPOINT_411.md).

User task: “Read the chart and its scope without stepping through repeated
headings, borders and instructions.” Consolidate repeated chrome in a bounded
set of existing quantitative owners, starting with numeric histograms and one
service/native composition. Keep reusable display styles separate from page
layout; do not hide units, coverage, source consistency or exact-value access.

Provide one clear heading/frame per visual hierarchy and a discoverable narrow
scroll region. Keep local display changes separate from source-capable gestures.
Codify the invocation/completion distinction exposed by 407: a control can be
correctly disabled during an asynchronous operation while its checked completion
still has a valid handoff. Retained redraw must not renew original-read authority.
Use an existing owner/helper boundary, not a global view/source manager.

Acceptance: real desktop/390px before/after geometry, keyboard/focus and exact
value disclosures, unchanged source/read/write budgets, usable local held
controls and source-free Saved highlighting. Reuse the passive reference for
embedding conformance. Prefer focused owner/semantic tests over another numbered
campaign and do not introduce new dependencies on private milestone helpers.

## Review gates and deferred candidates

Finish/review each milestone before advancing; keep heavy checks serial and
preserve telemetry/stores. Review the whole product after 411.

A contextual review-task chooser, captured-occurrence navigation, fixed-summary
rate comparisons, a bounded verification-family extraction, compact record
operations and raw-hash disclosure cleanup remain relevant candidates. They are
not silently appended to this queue. Exponential histogram execution, grouped
P95, broad host/platform expansion, authenticated collaboration and arbitrary
executable kinds remain outside it. No runtime performance rewrite is justified
by this review; profiling should precede such work.
