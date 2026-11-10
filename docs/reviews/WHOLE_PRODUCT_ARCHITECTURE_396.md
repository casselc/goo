# Whole-product architecture checkpoint — 396

> Root acceptance closure: 396 actual main/actor, property and required observed
> model checks are accepted; see [validation](../../VALIDATION.md). The review
> findings below retain their original read-only context. The adopted queue is
> the [396 roadmap](../ROADMAP_2026_11_10_CHECKPOINT_396.md).

Draft architectural input, November 10, 2026. 396 runtime acceptance was pending at review time.
Its source is frozen and focused checks have been reported; the final browser,
actor and observed-model results must be accepted before the planner adopts a new
queue. This document proposes no parallel implementation work. The review starts
no builds, benchmarks or browser campaigns and makes no production edits.

## Product center and existing boundaries

The useful loop remains **see current telemetry → isolate a population → inspect
related evidence → preserve or reuse the useful scope and presentation**. Ordinary
OTel is enough to begin. Manifests add meaning; they are not an ingestion or
exploration prerequisite. Current Activity, typed filters, trace search,
comparisons, discovered dashboards, native measurements and retained observations
are delivered capabilities, not categories to recreate from a vendor checklist.

The layers still fit that product:

| Layer | Responsibility to retain |
| --- | --- |
| OTel discovery and typed identity | Reported signals, qualified fields, namespaces and environment policy; missing, empty and literal values remain distinct. |
| `exploration` source planning | Validate the complete question, construct typed SQL parameters and bound physical reads. Source results retain their exact question, basis and read clock. |
| `metrics` / `measurements` | Per-stream arithmetic and explicit unknown/reset/coverage semantics. Numeric evaluation is pure; lifecycle custody is an explicit additional effect. |
| `scopedview` and supplied-data adapters | Exact values, units, support and finite dataset geometry; a named completeness witness concerns that captured query aspect. |
| Passive components | Display coordinates, labels, local presentation and typed selection. No fetching, source publication or navigation ownership. |
| Application owners | Applied versus draft question, local-definition validity, requests, current capabilities and retained return. |
| `workbench` | Immutable configuration/evidence publication and verified exact-ID loading; configuration reuse does not transfer a donor query or observation authority. |

Record pages and trace geometry retain their checked family contracts. The finite
dataset profiles do not yet express every inspector or arbitrary contributed
widget; that wider product goal should not be described as already delivered.

The display grammar is a supplied-data contract above those layers. It is not a
replacement query planner, optimizer or measurement kernel. Nothing in the recent
work justifies a universal query engine, validator, renderer registry, navigation
manager or generation-counter framework. Distinct return owners have useful
contracts; common styling and pure control predicates need not merge them.

## What 393–396 established

393 exercised legal literal transport through actual public source calls and a
settled ordinary OTel bank. Different encoder spellings round-trip correctly;
there is no correctness reason to merge the record and native HTTP clients. The
actual typed-input failure was elsewhere: JSON cloning repaired malformed UTF-8
before validation. Validating `Question` before detaching it protects both source
execution and offline evaluation. This is not proof that HTTP JSON ingress rejects
unpaired surrogates; that is a different boundary.

394 removes an unnecessary record-inspection detour. A compatible ordinary count
distribution can stage its exact query and existing flat display configuration in
an investigation-trail child. It does not republish the old snapshot as a fresh
read. Nonzero offsets and unsupported quantitative profiles are explicitly refused
by that shortcut rather than silently normalized.

395 adds one finite grouped-count dataset and declaration profile. Its typed
category and exact group-by-interval identity distinguish null from empty strings.
Eight groups and 1,000 observed source cells can produce at most 1,920 dense cells
across 240 intervals. An absent bin becomes known zero only with coherent,
exhaustive coverage; disagreement retains positive diagnostics and makes filled
absences unknown. The last partial interval retains its actual bound. Heatmaps
refuse more than 32 columns without truncating the population. These are captured
query guarantees, not proof of complete ingestion or a transaction across reads.

395's executed verification included independent raw-record count matrices,
configuration/store reuse, actual known-zero staging and no-read Back. The review
agent previously ran 32 generated Go/JS projection/reuse/store cases and 13 wire
admission cases, plus the required actual-ledger/Quint check. Actual evidence and
generated source fixtures remain explicitly separate in [validation](../../VALIDATION.md).

396 has reproduced two different failures: after outer trail Back/Forward, visible
local grouped edits could redraw but Save would publish the old declaration; a
fresh invalid UTF-8-byte-length Find draft could also leave Save enabled. The
source fix separates local-definition publication from retired acquisition epochs,
while preserving current view/query membership and source-staging guards. Invalid
text stays visible, the plot retains its last valid query, and owner validity
blocks publication until correction. Ranked display callbacks now use the same
owner predicate where the ownership semantics actually match.

The distinction between nested population return and outer trail return matters:
the former renews definition controls; the latter can retain an existing paused
adapter. A global resume or generation reset would conceal that distinction.
396 also gates native DOM callbacks independently of explicit programmatic display
updates, including the unsplit interval controls affected by the shared owner
predicate. Runtime acceptance of the final fix is still pending here.

## Ranked inputs to the next finite queue

These are recommendations for the planner, not adopted milestone numbers.

### 1. Breadth: customize an existing numeric or duration distribution directly

User task: “This mean payload size or endpoint latency breakdown is useful; keep
and reuse this presentation without reconstructing the view.”

There is a concrete admission gap. `ui/distributions.mjs`'s
`distributionCustomization` deliberately rejects `q.numeric` and `q.summary`.
`workbench/ViewDeclaration` and `ui/view-declaration.mjs` already support checked
numeric and duration measures, and the ordinary supplied dataset already carries
their exact values and support. The feature would extend an existing shortcut,
not introduce another calculator or authoring surface.

Keep the first cut to ordinary single-field Bar/Table over the existing numeric
binding or raw-span duration summary. Reuse the investigation trail, flat display
configuration, store and explicit Run. Preserve field kind/key, numeric type,
scale, literal unit and parse policy; duration summaries remain span-only. Do not
silently turn a selected mean, sum or percentile into count, population share or
an unweighted aggregate of displayed cells. Missing/empty/invalid support and
unavailable values must remain distinct from zero. Status ratios, numeric
histograms and grouped timelines are separate profiles.

Review checks: exact destination compatibility and unchanged destination question;
source-free stage/local edit/Inspect/Apply/Back; immutable old configuration;
quantitative labels and support in desktop/narrow views; rejecting incompatible
unit/type/parse and current-owner controls. A generated cross-language corpus
should cover the newly admitted arm, not repeat all unrelated native families.

### 2. Depth: reuse a grouped raw-span duration timeline

User task: “Keep a service or route latency-over-time view and compare the same
presentation on another fixed population.”

This is a real profile gap rather than missing execution.
`ui/grouped-time-series.mjs` already renders per-cell span duration summaries, but
395's `GroupedCountTimelineQuery` and grouped timeline declaration expressly admit
count only. `groupedSeriesAgree` already checks additive duration sums/counts
across independent reads; scalar percentiles cannot supply those checks.

Start with a finite spans-plus-`span-duration` mean/p95 profile using the same
supplied geometry, timeline/records composition and existing source/evaluator.
Carry count and duration support explicitly. A zero-record bin has no duration
value; it is not zero latency. A recorded zero duration can be real data. A
changed read must not acquire completeness or action authority. Each p95 comes
from that cell's raw durations; no combining or averaging quantiles across groups
or intervals. Keep the existing group/observed-cell/derived-cell limits and
heatmap width refusal. Do not add numeric parsing, status ratios, rates or a new
normalization/template arm to this cut.

The semantic gate must precede coding: pin exact field aggregation metadata,
units, support states and declaration compatibility; independently check cell
means/quantiles from raw spans, including known-empty and zero-duration cases.
If source summaries cannot attest a proposed meaning, narrow the profile. Existing
count-only artifacts need no compatibility bridge in this prerelease product.

### 3. Consolidate: remove passive presentation dependence on the app owner

User benefit: adding a chart profile should not require importing an application
controller to format or size a value, and alternative displays should agree on
exact labels and unavailable states.

There is an actual module cycle:
`ui/distributions.mjs` imports `ui/grouped-time-series.mjs`, which imports
`Distributions` for `groupMetric`, `metricWidth`, `valueLabel`, `percent` and
`durationMS`. Those pure presentation operations can live in an existing low-level
value module or one small explicitly named presentation module. Their application
ownership is incidental; it is not a source authority boundary to replicate.

Keep this extraction narrow. Cohort snapshots, source witnesses, queries,
selection effects and navigation stay with their current owners. Do not move
`Distributions.snapshot` into a passive formatting module simply because it shares
the file. Use the supplied-value contract where it already expresses the meaning;
avoid another math/formatting registry. Remove duplicate count/duration rendering
only when the new profiles cover the actual callers; remaining numeric/status
execution is useful functionality, not something to delete for a smaller diff.

Review checks: importing the passive renderer no longer creates the app-owner
cycle; exact count/duration/support labels stay equivalent; standalone HTML and
nested SPA property updates remain supported; source/store calls stay zero for
local rendering. Profile representative 8×240-cell input and Find/kind interaction
before making performance changes. Node counts or campaign wall time are not
application latency measurements. A no-optimization result is acceptable when
profiling finds no meaningful bottleneck.

## Reliability, Go and process conclusions

Go 1.27.3 is already current for this project. Existing use of typed structs,
`slices`/`maps`, `sha256.Sum256`, strict JSON decoding and `t.Context` is appropriate.
Use current standard-library features when a touched path becomes simpler or
safer. There is no evidence for a compiler-upgrade task, public generic algebra,
concurrency rewrite, pooling or index service as a prerequisite to these outcomes.

The transport audit is closed. Preserve family-specific readonly settings, row,
byte and time budgets, framing/error policy and cancellation. Repeated validation
at a typed boundary and after detaching is justified correctness work; removing it
requires evidence, not a branch-count target. Native profile dimensions remain
reported evidence, and a native HTTP-client histogram must not grant Client-span
population authority.

Retain one serial heavy-test lane. Parallel read-only reviews and independent
bounded edits help; overlapping browsers, property runners, model campaigns and
builds make failures harder to attribute. Tests should traverse actual producer,
wrapper and host callbacks: 391's second gate and 396's visible/save disagreement
were missed by inspecting only the final handler. Preserve genuine counterexamples
and the small fixtures needed for later regression work, without retaining every
failed build or duplicate screenshot indefinitely.

## Reconcile earlier proposals

- Count distribution customization, reusable grouped counts and the 396 local
  draft/lease distinction must not reappear as new categories if acceptance closes.
- Saved/service gauge observation context is still a useful retained-evidence
  adoption, but it should not automatically displace ordinary quantitative reuse.
  Stocks, broader resource contexts and exemplar citation need their own admission.
- A contextual review-task chooser and verified kept-occurrence review are real
  usability/target gaps. Captured-record review, labels, pinned export and keeping
  live observations already exist; do not propose them as missing foundations.
- Current capability and roadmap documents should supersede historical design
  status banners. For example, training-design text still claiming reviewed
  editing/export is unimplemented needs reconciliation with the delivered
  captured-record review contract; that is documentation hygiene, not a new store.
- Exponential histogram execution keeps its mathematical/source gate. General
  WYSIWYG/contributed executable kinds, multiplayer presence, tenancy, alert
  delivery, warehouse lowering and new bitemporal stores remain separate wider
  initiatives. None is an implicit prerequisite for the proposed finite queue.

The planner should combine this input with the independent product/design review
and final 396 acceptance, then publish one adopted roadmap rather than appending
another parallel backlog.
