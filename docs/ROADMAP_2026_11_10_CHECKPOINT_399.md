# Roadmap after the 399 checkpoint

Completed by 402 and superseded by the [402 checkpoint](ROADMAP_2026_11_10_CHECKPOINT_402.md).

Adopted November 10, 2026 after the 399 source/frontend/Go regressions and
independent whole-product architecture and design reviews. This replaces the
completed [396 checkpoint](ROADMAP_2026_11_10_CHECKPOINT_396.md).
[Validation](../VALIDATION.md) owns executed results; the
[review reconciliation](reviews/ROADMAP_RECONCILIATION_399.md) records the adopted decisions.

One finite cadence is active: **400 breadth → 401 depth → 402 consolidation**, then
whole-product review. Deferred ideas are not additional execution queues.

## Direction and lessons

Keep **current ordinary OTel → isolate a population → inspect or compare evidence
→ preserve or reuse the question and presentation**. Existing Live Keep, typed
filters, comparisons, traces, service/resource views, captures and reviewed export
are delivered capabilities. A manifest remains optional.

397 made existing numeric and duration Bar/Table displays directly customizable.
398 made grouped mean duration reusable after independently proving raw-cell
sum/count support. Empty duration is unavailable; a reported zero duration remains
valid zero. 399 extracts pure display arithmetic/formatting from application owners. Its
acceptance establishes unchanged values and effects, not an assumed speedup.

The next missing breadth is authoring an existing population summary through the
same grammar. The next depth is following a retained gauge observation from the
surfaces where operators already discover and preserve native evidence. The next
consolidation makes dense tables easier to read without another table framework.

398 exposed an important acceptance gap: a leaf declaration can be valid while
its enclosing saved-view schema still rejects the query/profile. For every new
role or adoption arm, execute finite Go/JavaScript/API vectors through both the
leaf and real carrier, with refusal controls. Keep this inside feature acceptance;
do not create a universal runtime schema interpreter.

Use Go 1.27.3 and modern standard-library features where they improve concrete
code. This prerelease has no legacy-format compatibility obligation. Keep public
APIs ordinary and typed; no language-feature quota, generic clone machinery,
unmeasured performance claim or speculative source concurrency belongs here.

## 400 — breadth: author a compact population summary

Delivered through [authored population summaries](AUTHORED_POPULATION_SUMMARIES.md).
[Validation](../VALIDATION.md) records actual source, actor and contract checks.
The remaining active cadence is 401 → 402, then whole-product review.

User task: “Keep the few relevant facts beside these records or this chart, then
reuse that presentation over another compatible population.”

`summaryDataset` and `goo-measure-summary` already render supplied population
facts. The composition has aggregate, records, timeline, comparison and facet
roles, but no summary role; current automatic summaries sit outside authored
composition. Add one optional typed summary child under the existing fixed record
query owner, store and reuse flow. Do not add summary SQL or another evaluator.

Before implementation, enumerate the finite ordinary log/span query combinations
and summary fields already supplied by the source. Pin supported field families,
kind/version, a bounded distinct field list and compact defaults. Count, supported
numeric and span-duration facts are the initial target; status joins only when its
existing count/fraction support is explicitly admitted. Keep records required;
summary admission must not require an unrelated categorical grouping. Preserve optional aggregate/timeline compatibility through the existing
composition rules; specialized histograms, native scopes, reported-edge and other
unsupported profiles do not gain admission merely by supplying numeric values.
Grouped/native/comparison-summary and arbitrary calculated fields remain outside
the initial summary arm.
Do not silently broaden a query or manufacture a summary aspect to satisfy a view.

The initial visible summary should show a few chosen facts, not all available
fields or a new prose wall. Use the existing passive component and disclosure for
support/method/read facts. Do not render a duplicate automatic summary beside an
authored one. Selection and field order are presentation, not queries.
Count zero, empty input, no valid numeric values, unavailable measures and changed
reads remain distinct. A summary read does not prove transactionality with a row
page or complete upstream telemetry.

Acceptance:

- Actual ordinary log/span workflows configure selected supplied fields, reorder
  locally, save/reopen and Inspect/Apply in a compatible destination. Preserve the
  complete destination predicate, environment policy, exact bounds and paging.
- Independent source oracles check selected values and their count/support,
  method and units. Numeric fields require the same qualified binding, type,
  scale, literal unit and parse policy; duration/status retain their own meaning.
- Stage, field/order changes, Inspect/Apply, reopen and held Back read no telemetry.
  Run is explicit; current configuration publication is counted separately.
  Real wrappers refuse hidden/inert, replaced, pending and stale owners.
- Validate leaf declarations and enclosing composition/store/API carriers with
  omission/null/unknown-field/text and unsupported-combination controls. A donor
  contributes presentation only, never snapshots, completeness or source leases.
- Refresh compact desktop/narrow frames, storyboard, relevant Bombadil actions
  and generated/observed checks. Adopt no new summary arithmetic in this cut.

## 401 — depth: records around retained gauges in Services and Saved

Delivered through [retained gauge context](RETAINED_GAUGE_RECORD_CONTEXT.md).
[Validation](../VALIDATION.md) records actual source, actor and contract checks.
The remaining active milestone is 402, then whole-product review.

User task: “This reported gauge observation is unusual or unavailable. Show logs
or spans around it, then return to exactly the native evidence I was inspecting.”

Ordinary Measurements already offers this gesture. Services and Saved have useful
native evidence and existing held-return owners, but their current context
admission is counter/histogram oriented. Extend those owners with the existing
service-gauge observation planner, not a fabricated native export interval.
Stocks and resource-bound gauges remain outside this cut.

First prove retained-evidence admission separately for both surfaces. Services
must bind the accepted successful result to its actual requested question,
widget, current visit and exact retained stream/sample. Saved must first admit
its immutable receipt/result, then match the selected stream and uniquely retained
reported timestamp. Preserve capture custody time, original metric read/evaluation
window and fresh record read time independently. A caller-supplied timestamp or
current clock cannot establish membership.

Reuse the existing ±60-second plan clipped to the original half-open nanosecond
bounds, with enclosing millisecond execution bounds and exact timestamp predicates.
Retain environment policy, service instance and missing/empty/named namespace.
Metric attributes, scope/schema identity and unrelated resources do not become
record filters. A recorded timestamp can supply time association even when its
gauge magnitude is unavailable; it grants no causality, contributing-request,
complete-trace or coverage claim.

Acceptance: actual service and immutable Saved gauge → staged logs/spans → explicit
Run → held return, including real zero and unavailable sample values. Return keeps
selection, evidence/receipt, focus and scroll with no native reread or capture-store
GET. Configuration edits, Apply, replacement, departure and late responses retire
old staging; returning renews only the unchanged current evidence capability.
Forged stream/timestamp/question/basis, wrong receipt and out-of-window controls
must fail. Stage grants no metric capture, source execution or display-save lease.
Run finite Go/JS/API/carrier vectors where touched and actual/model/property/actor
checks; update the existing custody/process diagrams and storyboards.

## 402 — consolidation: discoverable table overflow and retained context

Operator task: “Read the later quantitative fields or intervals without losing
which group, stream or record I am looking at.”

Current 397/398 narrow frames and native retained tables show a concrete issue:
wide content is internally scrollable but overflow is easy to miss, and identity
can leave the viewport while values are compared. `goo-table` already supplies
keyboard focus and region semantics; these are not missing features.

Use that shared passive boundary and scoped styles to make horizontal overflow
apparent only when present and retain sufficient row/header context while reading.
Choose the smallest checked context treatment after inspecting actual ordinary
numeric, grouped-interval and native-table markup. Keep literal case, UTC/exact
interval tooltips, units, support and state visible. Do not add a second grid,
query sorting, virtualization, permanent controls, a table-authoring language or
page-level overflow. Preserve SSR/detached/native HTML and nested SPA use.

Acceptance: standalone and actual nested tables at desktop/390px, keyboard and
pointer access to later columns, stable identity/header context and accessible
names without duplicated announcements. Nonoverflowing tables gain no redundant
chrome. Local scrolling/context changes perform zero IO and do not modify query,
selection, full-population scales or saved definitions. Preserve existing typed
custom-column behavior. Refresh only affected genuine frames/storyboards and
meaningful shared-component regressions. Claim no performance benefit without a
measured path.

## Reconciled deferrals

| Candidate | Treatment |
| --- | --- |
| Grouped P95 / percentile comparison reuse | Still requires independent raw-cell/population support gates; 398 mean support does not make scalar percentiles composable. |
| Passive display entry/gallery expansion | Useful caller discoverability work, deferred after the concrete 399 dependency closure; no framework migration prerequisite. |
| Clone/copy optimization or broad source refactor | Preserve validate-before-detach and ownership. Require mutation regressions and representative allocation/latency benchmarks before changing correct implementations. |
| Broader dashboard/profile coverage | Existing service/resource and OTel-derived dashboards are shipped. Additional DB/broker/native families need explicit source/semantic gates, not a vendor-widget checklist. |
| WYSIWYG/executable kinds, tenancy, alerts, warehouse/new stores | Separate design/platform gates; do not bundle with a summary role or gauge continuation. |
| Task discovery / training-data review expansion | Existing Keep, captured-record review, vocabularies and export are delivered. Reassess concrete operator gaps before appending a new review surface. |

Keep heavy jobs serial, retain useful current acceptance artifacts, and preserve
lab telemetry and stores. Review the whole product after 402 or sooner if 400's
source-field or 401's custody gate invalidates the order. Do not auto-append deferrals.
