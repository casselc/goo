# Architecture and Go review — 408

Read-only review of the product after 406–407, while 408's passive exports and
embedding reference were being prepared. No tests, benchmarks, browser campaign,
dependency upgrades or production edits were performed for this review. Actual
408 acceptance belongs in [validation](../../VALIDATION.md). Recommendations are
inputs to the reconciled roadmap, not a second execution queue.

Reconciliation: the [408 roadmap](../ROADMAP_2026_11_11_CHECKPOINT_408.md)
prioritizes service captured-record comparison, incoming HTTP response context
and quantitative chrome consolidation. The review-task, rate-comparison and
verification-family proposals below remain deferred candidates. Final static
review of the passive exports/reference and that queue found no architectural
blocker; actual embedding acceptance remains a separate validation result.

## Keep the architecture

The current product loop is useful: ordinary OTel Activity → scope a fixed
population → inspect or compare evidence → preserve/reuse its question. Services,
Measurements, Saved, annotations and packages contribute to that loop; they are
not separate products needing their own query engines.

The one Go runtime and present layers remain appropriate:

| Layer | Responsibility to preserve |
| --- | --- |
| Discovery and identity | Exact service/resource identity, qualified attributes, environment policy and supported instrument metadata. |
| `exploration`, `measurements`, `metrics` | Source planning and budgets, checked questions, exact arithmetic, support/reset/state semantics. |
| `scopedview` and source adapters | Supplied values, geometry, per-read clocks and explicit completeness. |
| Passive components | Local display and typed intents; no subscriptions, telemetry reads or publication. |
| Application owners | Draft versus accepted question, execution epoch, current/held capabilities, navigation and explicit Run. |
| `workbench` | Immutable configurations and evidence; metadata judgments remain distinct from telemetry facts. |

The existing bounded HTTP/framing helpers and route policy composition already
avoid a shared generic source engine. See [source mechanics](../../internal/sourcehttp/read.go),
[framing](../../internal/sourcejson/read.go) and [query budgeting](../../cmd/goo/query_budget.go).
No evidence in this review justifies DataFusion, a new database, concurrent source
fanout, a global renderer registry or another application lifecycle manager.

## Recent lessons worth applying

406 correctly introduced a distinct rate family and exact count/exposure facts,
rather than stretching bounded percentage metadata. 407 reused existing numeric
bins and the range planner, rather than interpreting display declarations as
queries. Their exact scope, typed-unit and full-carrier checks should remain the
pattern for future display arms.

407 also exposed adapter mistakes that static declaration admission could not
catch: using the generic resolver for a histogram; rejecting an internally
pending-disabled opener after successful planning; routing held redraw to the
wrong renderer; and accidentally treating all passive geometry as retired source
geometry. The fixes in [the population owner](../../ui/population-workspace.mjs)
and [numeric adapter](../../ui/numeric-distribution.mjs) now distinguish:

- Local presentation controls, which may remain useful on held data.
- Source-range authority, which belongs to the original accepted read and cannot
  be recreated merely by redrawing the same result object.
- Source-free Saved highlighting, which is not permission to stage/read telemetry.

The existing [retained definition controls](../../ui/retained-definition-controls.mjs)
explicitly cover synchronous definition edits only. Keep that limit. A later
owner cleanup should name local presentation, source suspension and held redraw
separately, using concrete role adapters. Do not replace these distinctions with
one generic `resume()` or infer source authority from DOM connectivity.

## Ranked proposals for the next finite cadence

### 1. Breadth: make existing review tasks discoverable

The annotations/review model already covers immutable captured occurrences,
explicit vocabularies, correction branches and manifest export. The practical gap
is task selection: [the review UI](../../ui/capture-record-review.mjs) exposes an
exact task-ID input and task declaration, while [the API](../../cmd/goo/capture_reviews_http.go)
provides creation and exact-ID loading but no bounded task catalogue.

Add a small source-free task catalogue and contextual chooser over existing task
revisions. Reuse file-store admission, bounded scans and explicit catalogue-change
handling; choosing a task must retain the exact occurrence target and do no
telemetry read. Keep uncertainty/abstention, unverified attribution and explicit
Save. Do not infer labels from tags, silently choose a task or require XTDB.
This adds useful reach across inspection/Saved review without introducing another
telemetry primitive or storage engine. Product/design should settle the first
surface; support for kept-only occurrences remains a separate admission decision.

### 2. Depth candidate: fixed-summary captured-rate comparisons

406 supplies exact rates in ordinary fixed summaries, but
[the held comparison construction](../../exploration/view_comparison.go) strips
them and [the shared comparison validator](../../ui/view-comparison.mjs) explicitly
refuses rate fields. A finite comparison arm could answer whether captured
activity changed when the two fixed windows have different widths.

Use each operand's independently proved count and elapsed duration; derive an
exact records/s difference in Go. Retain both scopes and clocks, unknown sampling,
known zero and unavailable support. Do not calculate from a displayed page,
normalize by another window's width or call this application throughput. Start
with ordinary fixed summaries only; grouped rates, rolling/live comparisons,
relative-percentage claims and rate timelines remain outside this proposal.
Existing counts already support the arithmetic, so any extra physical source SQL
would need a separate justification. This is a product-priority candidate, not a
reason to relax the current explicit comparison refusal ahead of implementation.

### 3. Consolidation: split one verification family from the milestone package

Static inspection counted about 92,000 lines across `verification/*_test.go` in a
single package. The 407 ledger uses helpers named for 384, 394, 400 and 404; see
[the ledger](../../verification/histogram_view_407_ledger_test.go) and
[property corpus](../../verification/histogram_view_407_property_test.go).
Reusing independent membership/bin oracles is good, but file chronology has become
an implicit support API. Editing one feature compiles unrelated test families,
and old fixture assumptions can become accidental new-feature expectations.
This is a structural observation, not a new compilation-speed benchmark.

Take one bounded cut: extract the independent ordinary-record membership,
exact-number JSON comparison and histogram support oracle into named internal
verification support, then move 406/407 carrier checks into one semantic-family
package. Keep model/observed tests opt-in and report missing evidence as skipped,
not success. Preserve negative controls and oracle independence; do not import
production evaluators to calculate expected answers. Measure cold and warm
compile time/memory before claiming a speedup. Do not move every historical test
or build a generalized campaign interpreter in the same change.

## Go, process and deferred work

Go 1.27.3 is current in validation; production already uses concrete typed APIs,
`math/big`, `slices`/`maps`, contexts and bounded source readers. Continue ordinary
private functions and concrete domain types. `maps.Clone`/`slices.Clone` remain
shallow; they do not replace validate-before-detach or nested ownership checks.
No broad generic public builder or experimental JSON migration is warranted.
A small neutral wire-field helper may later replace the summary-named decoder
used by [histogram declarations](../../workbench/view_histogram.go), but this is
lower value than the scoped tasks above and needs no separate milestone now.

Keep heavy validation serial and freeze participating files before the root
acceptance run. Distinguish production defects from test DOM/gesture mistakes;
repair test support to match native HTML rather than weakening product guards.
The full-carrier properties and actual visible-control journeys caught different
classes of defects in 407; keep both, but avoid duplicating another complete
source oracle for each numbered feature.

The XTDB documentation mismatch identified during review is fixed in this
checkpoint: [annotations](../WORKBENCH_ANNOTATIONS.md) now uses `xtdb packages`,
matching [the current overlay](../../compose.xtdb.yaml). No runtime compatibility
alias for the removed `query` service was introduced.

Service discovery/automatic overviews, resource investigations, native instrument
selection, trace continuation, comparisons and review/export are delivered; do
not relist them as missing vendor parity. Fleet quantiles, exponential histogram
execution, auth/tenancy, alert delivery and external warehouse deployment remain
explicit separate design/source gates. No new measurements here establish runtime
bottlenecks or production readiness.
