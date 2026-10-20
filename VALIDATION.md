# Validation record

Current additions: see the [cloud baseline](ENVIRONMENT_VALIDATION.md) for the
successful default Compose run, real cross-app Bombadil journeys and generated
Quint MBT success/fault campaigns in this environment. The historical hosted
environment limitations below describe the original handoff session.

## Live span latency and recorded error share (2026-10-20 UTC)

Saved span questions requesting duration/status now compare count, compact
mean/p50/p95/p99 and recorded Error/total share under one rolling window and
current evidence basis. The existing count aggregate already returns these
measures in one SQL statement. Current refresh responses expose current/baseline
basis metadata without changing historical comparison receipt formats. Freeze
preserves the complete original result without another source read.

- Root Go suite/vet passed; exploration/HTTP checks passed under the race
  detector. A new source check retains full measures over 1,200 records while
  keeping only 1,000 inspection rows, verifies one combined aggregate per read,
  and freezes/reopens summaries after the source is closed.
- The frontend suite passed 66 tests. Shared latency elements accept optional
  server `ceiling-ns` or detached `.ceilingNs` properties. Exact geometry retains
  nanosecond integers; invalid/undersized domains withhold markers. Existing
  automatic scales remain the default for other callers.
- Hegel passed 100 generated large-population comparisons, immutable summaries,
  Unknown status preservation and descriptive-change reversal checks.
- The main campaign passed five browser checks (3.5 minutes): both chart-frame
  contracts, existing record/native fresh comparisons, count-only live comparison
  and real OTel live span regression. Both live actors executed required
  preview/choose/start/freeze/link actions. The span actor additionally checks
  per-measure common domains, exact percentile positions and snapshot basis.
- The span journey observed baseline 1 ms/1 recorded error out of 2, current
  10 ms/1 out of 4, then 100 ms p95/p99 and 4 errors out of 7. Count and latency
  required explicit rescale; recorded error share retained its 0–100% axis.
  Failed refresh preserved all displayed summaries, basis and bounds. After an
  eighth 1,000 ms error arrived, freeze still retained the displayed 7 spans,
  340 ms duration sum and 4 errors. Quint accepted the observed ledger and
  rejected all ten corruptions. The count-only ledger retained six controls.
- A separate real-source check passed empty-current-window capture/freeze:
  no percentile or error-share geometry, no zero-latency/zero-error claim, and
  unavailable comparison state retained after freeze. The supplied-domain HTML
  and pre-upgrade SPA checks also passed. Eight distinct browser tests passed
  across the completed campaigns; source-backed 71 and affected 57/63/69 frames
  were regenerated and desktop/mobile 71 frames inspected.
- The first supplemental wrapper attempt stopped before Playwright: Docker
  briefly reported the owned traffic generator as a zombie. It subsequently
  exited and was restarted; the retry drained without deleting data and restored
  traffic. One added standalone assertion incorrectly assumed the unit fixture's
  p50 rather than the gallery's 40 ms p50; the corrected assertion passed. These
  unsuccessful attempts are not counted as passing checks.

Final runtime check matched image and served source/binary bytes, all 12 services
running, health OK and historical pending IDs retained. The traffic process was
also checked alive (not zombie). Evidence: `artifacts/live-span-journeys.log`,
`live-span-component-empty-retry.log`, `live-span-components-final.log`,
`live-span-hegel.log`, `live-span-final-health.json`, and
`artifacts/live-span-measures/0cb74c185d/` including its actual ledger/actor trace.

Count-only captures retain their definition; native metric comparisons remain
fixed-window. Measure refresh is bounded polling, not SSE. Status coverage and
sampling/completeness remain explicit unknowns. See [the contract](docs/LIVE_COMPARISON.md),
[the component domain API](docs/UI_COMPONENTS.md#supplied-percentile-domain), and
[journey 71](docs/journeys/README.md#71-a-service-regression-across-count-latency-and-recorded-error-share).

## Rolling record baseline and passive chart frame (2026-10-20 UTC)

Saved span/log questions now support explicit rolling full-population count
refresh beside frozen linked evidence. The app owns cancellation, generations,
source-window labels, stale retention and explicit axis rescaling. Freeze
publishes the displayed server snapshot without another telemetry read; linking
it stages an investigation draft and saving creates an immutable child revision.
The parallel component slice supplies `goo-chart-frame` with SSR attributes,
detached SPA metadata, native child/focus retention and no source lifecycle.

- Root Go suite and vet passed. Source/capture and HTTP checks passed under the
  race detector, including source-free exact freeze, repeated-freeze idempotence,
  strict request validation, 16-draft admission, budget replacement and obsolete
  capability rejection. Python proxy syntax and Compose configuration passed.
- The frontend suite passed 64 tests, including unavailable/incompatible frozen
  display state and unstarted-panel pause semantics. Hegel passed 100 generated
  shifted equal-duration windows and incompatible environment cases. Its first
  run generated durations beyond the existing two-hour record-query limit;
  the corrected generator matches that supported domain rather than widening it.
- Four browser checks passed together: both passive frame contracts, the existing
  fixed-window fresh-comparison record/native journey, and the real OTel rolling
  comparison journey. Frame tests use explicit synthetic standalone HTML; the
  product journey reads actual ClickHouse ingestion through standard OTLP.
- The rolling journey observed 2 baseline/4 current records, grew current to 7,
  required explicit rescale, injected a 503, retained the stale result/window,
  then ingested an eighth source record before freezing exactly the displayed 7.
  It reopened the stored result, verified the baseline unchanged, and rejected a
  completed response held across preview disposal. Quint accepted the observed ledger and rejected all six
  corrupted observations.
- Initial source failure is labelled unavailable with no invented retained result.
  Mobile frame bounds are asserted independently of page overflow: screenshot
  inspection exposed nested paired-bar clipping, fixed in the shared component
  stylesheet. Baseline/source provenance is disclosed without repeating it inline.
- A subsequent full rolling run passed explicit investigation linking and saved
  revision preservation. Its 30-second Bombadil actor executed preview, baseline
  choice, start, freeze and link; the test requires all five observed actions.
  The final state mapping also has focused regression coverage: unavailable
  frozen evidence remains unavailable, and pausing an empty panel stays unknown.

Final release: all four selected browser tests passed together in 1.4 minutes
(`artifacts/live-measures-release.log`), including the required Bombadil actions
and live Quint replay with six rejecting mutations. Runtime validation matched
the final image and served source bytes, all 12 services running, health OK and
all five historical pending stream IDs preserved. The wrapper restored traffic.
`artifacts/live-measures-final-health.json` records the final checks.

Live native metrics, duration/status charts and SSE measure refresh remain
planned. Record refresh uses bounded polling and a two-minute process-local
snapshot capability until explicitly frozen. Serialized draft budgets do not
establish a production authorization model or telemetry completeness.
See [the contract and sequence diagram](docs/LIVE_COMPARISON.md),
[journeys 69–70](docs/journeys/README.md#69-a-frozen-baseline-beside-a-rolling-record-population)
and [the passive component contract](docs/UI_CHART_FRAME.md).

## Live logs, passive waterfalls and count-time plots (2026-10-20 UTC)

Explore now tails an applied log selector over a separate rolling event-time
window. The Go source/compiler adapter emits bounded SSE batches through the
Python streaming proxy; the app owns generations, reconnect gaps, buffer
retention, follow-scroll and inspection pause. Counts label local observations,
not the full question population. Two parallel agents extracted passive
`goo-trace-waterfall` and `goo-time-bars`, adopted by existing trace inspection,
Explore and Services, and prepared a chart/detail reuse review.

- Root Go tests and vet passed; the frontend suite passed 60 tests. The tail
  HTTP admission/cancellation checks passed under the race detector.
- Hegel passed 100 generated overlapping-multiset/late-arrival campaigns. An
  initial property assumption incorrectly classified a first observation as late;
  the corrected property compares against a previously observed maximum.
- The release campaign passed nine tests: core basics, discovery, workbench,
  Services, UI inspection and both standalone waterfall/time-bar contracts.
  Core/Services Bombadil campaigns require actual chart interaction and check
  exact count/state geometry and focus preservation. The inspection campaign
  completed deterministic waterfall selection/detail checks before exploration.
- Tail browser assertions passed duplicates, same reported IDs, late data,
  foreign-scope exclusion, 100-row eviction, focus pause, resumed gaps, injected
  transport failure recovery, mobile containment and scope clearing. Its first
  integrated model check found a Quint built-in-name collision; after renaming
  that action, the observed ledger and all six rejecting mutations passed.
- Earlier live checks exposed newest-first batch eviction discarding the newest
  observations; the source now presents each bounded batch in event-time order,
  with a regression check. A chart fixture retained application/json from /health
  and did not upgrade custom elements; its explicit synthetic HTML fixture passed.
  An interrupted image deployment was not counted as validation; subsequent
  deployment checks matched running image IDs and served asset bytes before tests.

- The final tail rerun passed end to end, including the 30-second Bombadil
  start/pause/resume/focus/close sequence with required completed-stage evidence,
  Quint live replay and all six negative controls. Across the corrected runs all
  ten selected browser tests passed. New standalone and real app frames were
  regenerated and inspected; the chart/detail review is linked from the library.
- Runtime verification passed: all 12 services running, health OK, served tail,
  waterfall, time-bar and application hashes plus Go binary match the workspace
  and build context. Existing five stale Redis pending IDs were preserved; one
  current active entry brought the final pending count to six. Traffic restored.
- Durable ingestion cursors/replay, trace tails, arbitrary chart algebra and
  baseline-versus-live measures remain planned. Count plots retain explicit
  zero/unknown distinctions; native interval/histogram semantics remain separate.

## Shared facets and typed attributes (2026-10-20 UTC)

Explore live/frozen distributions now compose a caller-owned `goo-facet-list`.
Local search hides rows without changing supplied exact counts, scales, ranking
or action state. No local matches is distinct from no captured source records.
Attribute inspectors now reuse `goo-field-value` and a passive
`goo-attribute-list`, keeping qualified mapping, search, clipboard feedback and
query actions in the existing app adapter. Values preserve explicit presence,
null/empty/false, exact integer strings, safe text, lazy nested batches and a raw
value escape. Standalone SSR and detached SPA use are documented.

- Frontend suite passed all 54 tests. New pure checks cover exact values and
  absence/type distinctions, JSON rejection, cyclic/unsupported input and
  accessor/inherited serialization hooks without executing hooks.
- The initial live campaign passed eight browser tests: component conformance,
  log/trace discovery, grouped measures, question workbench and UI inspection.
  Discovery/group/workbench observed Quint ledgers and rejecting controls passed;
  30-second grouped and attribute-inspection Bombadil campaigns passed with new
  facet visibility/presence invariants.
- Standalone checks cover SSR/pre-upgrade properties, detached values,
  reconnects, safe text and native facet row/state retention without API reads.
  Live inspection verifies scoped duplicate keys, literal strings, clipboard,
  no-match status, original-window narrowing and trace/log/mobile adoption.
- Source reads, query ASTs and Go semantics were unchanged. A parallel agent
  extracted and reviewed typed fields; review caught an inherited array
  serialization hook and a page/component CSS precedence issue, both corrected.
- Final post-review campaign passed all seven browser tests, including new
  invalid-property and 50-to-75-item lazy-expansion assertions, corrected
  trace-facet layout, updated grouped/inspection Bombadil campaigns and Quint
  discovery/group trace checks. Real storyboards and synthetic component frames
  were regenerated; standalone and live attribute screenshots were inspected.
- Runtime checks passed: all 12 Compose services running; health OK; served
  component/attribute/application hashes and binary match workspace/build context.
  All five pre-existing Redis pending entries remain and traffic was restored.
- Field snapshot updates reset disclosures. Facet row patches require explicit
  `refresh()`; there is no mutation observer or live transport. Broader
  screen-reader validation remains outstanding as documented for the library.

## Passive tables and explicit result states (2026-10-20 UTC)

`goo-table` and `goo-result-state` now provide native caller-owned table viewports
and supplied state displays in Explore and Services. Captions, scoped headers,
keyboard overflow and compact density share one implementation. Source reads,
full totals, paging and action callbacks remain in the app. Native row patches,
attribute changes and reconnects retain existing row/action identity; the component
has no subscription or query behavior. Draft/cancel paths clear loading state.
The design subagent implemented Services adoption and reviewed the shared contract.

- Frontend suite passed all 52 tests. The first live campaign passed all seven
  browser tests (component contracts, grouped measures, workbench, reported edges
  and Services), including observed Quint trace validation/negative controls and
  the 30-second Services Bombadil campaign. New Bombadil invariants check table
  captions/scoped headers and explicit polite result regions.
- Browser contracts check server HTML, pre-upgrade state assignment, caller-owned
  row/action identity through append/reconnect, retained focus, supplied state
  transitions, no component API reads and mobile keyboard horizontal scrolling.
  Services checks local-search misses without new reads and independent partial
  source availability. Workbench checks loading-to-draft state reset.
- Parallel review caught stale loading styling on cancelled/edited questions,
  a lost widget viewport bound and redundant outer scroll containers; corrected.
  Screenshot inspection also corrected mid-word table headings. Shared styles
  preserve readable headers with contained horizontal overflow.
- Final post-review campaign passed all six tests (components, grouped measures,
  workbench and Services), including the updated state/table assertions, Quint
  checks and a second Services Bombadil campaign. Affected real journey frames
  and synthetic reference frames were regenerated and inspected.
- Runtime verification passed: all 12 services running, health OK, query/packages
  image matches, served component/application asset hashes and Go binary match
  the workspace. All five pre-existing Redis pending entries retained; demo
  traffic restored. Build cache/dangling images were pruned without touching volumes.
- No Go semantic core changes were made. Full screen-reader validation of the
  existing mobile table-to-card pattern remains outstanding; the browser checks
  establish markup, keyboard behavior and containment. SSE remains planned.

## Framework-neutral display components and compact exploration (2026-10-20 UTC)

The public library now registers `goo-latency`, `goo-stat-list` and
`goo-metric-comparison`, using ordinary server HTML attributes and detached SPA
properties. Its standalone `/ui-components` reference has synthetic fixtures and
loads no app runtime/query client. Services and live/frozen Explore reuse the
latency/stat primitives; stored comparisons group mean/P95/P99 on one matrix.
Only supplied percentiles are shown; no P90 evaluator was introduced. Query
construction now occupies one compact workbench, and applied measures do not
force advanced settings open. Widget secondary actions use native disclosure.

The [whole-app design audit](docs/UI_COMPONENT_AUDIT.md) prioritizes passive tables,
result states, facets and typed attributes next. It also specifies eventual SSE
and baseline-vs-live boundaries. These additional elements and streaming features
remain proposed. Current summary updates replace inner DOM/reset disclosures;
they are not high-rate focus/scroll-preserving tailers.

- Frontend suite passed 52 tests, including exact large-value formatting,
  zero/absence, missing quantiles and input immutability.
- `TestHegelDashboardPlanning` passed 100 generated cases for the unchanged typed
  query/adaptor contract.
- Standalone browser tests passed server HTML upgrade, pre-definition property
  assignment/reactivity, detached values, malformed JSON, safe text, keyboard
  disclosure and mobile containment with no API requests.
- The corrected live campaign passed all 11 browser tests (component conformance,
  error/latency and stored comparisons, grouped measures, native windows, question
  workbench, full breakdowns, exact edges and service dashboards). Their observed
  Quint checks passed. Thirty-second Bombadil campaigns passed for comparisons,
  exact edges and service dashboards, including the new quantile display invariant.
- The standalone fixture also checks configurable heading levels, reconnects and
  detached current-side comparison updates that leave supplied baseline values
  unchanged. This validates display updates, not SSE delivery or a live evaluator.
- Direct screenshot inspection exposed an ancestor CSS minimum width leaking
  from an outer breakdown table into a nested component. App selectors are now
  scoped to their owning table; a composition check requires the Change column
  to fit on desktop. Grid placement belongs to the app, not the reusable element;
  a bounded heading-level setting preserves the embedding outline.
- After component chrome isolation, the six focused browser checks passed again,
  including standalone/reconnect updates, nested comparison geometry, mobile
  layouts, source-free edge evidence and the service-dashboard Bombadil actor.
  Affected live frames and clearly labeled synthetic component frames were
  regenerated and inspected. All 12 Compose services are running; health and
  served component/app asset hashes match the built workspace. Existing Redis
  pending IDs and volumes were retained and traffic restored. Unused build cache
  and dangling images were pruned after a full-disk build condition.
- The first live campaign passed eight tests and exposed two older fixture
  assumptions. A demo gauge legitimately became unknown while traffic was paused;
  its test now injects an isolated ordinary OTel gauge with a fixed fresh scope.
  A metric-handoff assertion now expects the explicit missing-namespace selector
  introduced by the earlier identity work. Neither fix relaxes source semantics.

## Reported-edge investigation and visual comparisons (2026-10-20 UTC)

Distributed relationships now stage typed exact child-population questions, keeping
parent ownership in the original environment/window/namespace universe. Filters,
packages, captures and investigation steps retain that selector; explicitly wider
child-service and whole-trace actions remain labeled. Real fixtures include 130
children beyond the 100-record page, duplicate/ambiguous parents and namespace-
scoped ownership. See [reported edges](docs/REPORTED_EDGES.md).

Comparison views now use paired zero-based bars, common duration ceilings,
0–100% error/share axes, equal-width breakdown tracks, explicit unavailable
hatching and aligned native histogram bins. Detailed scope/provenance and exact
tables remain in disclosures. See [visual comparisons](docs/VISUAL_COMPARISONS.md).

- Root Go tests and vet passed; the frontend suite passed 51 tests.
- Hegel edge planning passed 100 generated cases, including missing/empty/literal
  namespaces, parameterized service names and detached question custody.
- Live journeys 35, 52 (record and native cases), 54, 58, 59 and 62 passed,
  including real OTLP ingestion, saved evidence, source-free reopening, mobile
  containment and independent source failures. Their Quint observed-trace checks
  passed; edge replay rejected five deliberately corrupted ledgers.
- Thirty-second Bombadil campaigns passed for latency/error comparisons and
  reported edges. Comparison invariants check common scale metadata and bounded
  geometry; the breakdown journey additionally measures equal track pixel widths.
- A final native-layout refinement promotes the actual rate/distribution ahead of
  coverage clocks and retains metric-record counts in the exact-values disclosure.
  Its focused journey passed, checking visual order as well as reset/gap/layout
  withholding; the refreshed native frames were inspected directly.
- Final runtime checks: all 12 Compose services running, health OK, served UI
  hashes and Go binary match the built workspace. Existing Redis pending IDs were
  preserved; traffic was restored after each journey wrapper.
- Initial journey runs exposed fixture/test issues: the UI's two-minute scope
  legitimately included an extra parent, the trace action was attempted while
  Distribution was selected, and native-window tests still expected receipt v2
  after the earlier v3 rollout. The journeys now select the intended one-minute
  window, open Records before expanding a trace, and check the current version.
  No comparison receipt algorithm or historical receipt was rewritten here.

## Declared services alongside observed telemetry (2026-10-20 UTC)

Services now combines independent fixed-window discovery and current operation-
package registry metadata. Declared-only rows retain explicit not-observed states,
namespace uncertainty and all declared bindings without creating counts, event
timestamps or parent edges. Registry failure remains distinct from no declarations;
late registry results are discarded after scope edits. Current declarations do not
establish deployment or emission during a historical telemetry window.

- Frontend suite: 50 tests passed.
- `TestHegelDeclaredCatalogIsolation`: 100 generated cases passed, checking
  namespace presence/value isolation, declaration order, retained bindings and
  unchanged source evidence.
- Journey 61 passed its initial live historical-inventory → service → empty capture
  → source-free recipient reopening path, registry failure and stale-result checks.
  Its Quint replay and 30-second Bombadil campaign passed. The final focused run
  additionally checks explicit namespace retention and unavailable-registry wording.
- That added check exposed a real namespace-inventory persistence defect (HTTP
  502, `invalid saved inventory`): the receipt validator compared selector pointers
  after serialization. It now compares semantic values. Missing/empty/named
  namespace save/reopen/refresh and mismatched-selector regressions pass in
  `go test ./...`; `go vet ./...` also passes. Existing receipt formats are unchanged.
- The first post-fix run passed the live namespace capture and all five Quint
  negative controls, then exposed a Bombadil startup race: repeated start actions
  canceled their own in-flight discovery. The actor now starts once after the
  capture is enabled and waits for discovery; its liveness deadline is unchanged.
- Initial browser campaign: journeys 46, 47, 59, 60 and 61 all passed (4.5 minutes).
  Final focused journey 61 passed (52.1 seconds), including the namespace capture,
  `no_data` custody, five rejecting Quint mutations and the corrected 30-second
  Bombadil campaign. Ledger/actor output: `artifacts/declared-services/1792462359930/`.
- Refreshed affected service screenshots and added storyboard 61 with desktop/mobile
  frames and a registry → discovery → question → capture diagram.
- Final health/source checks: all 12 services running; query/packages image
  `sha256:5c88917f0a65ffd2ae2cab406ffdb88d275d37a1e4f74667de7f9bdad7b62251`; served assets and binary
  match the tested source/build. All five pre-existing stale pending IDs remain;
  traffic is restored (6 total pending entries at the health read).
  Evidence: `artifacts/declared-final-health.json`. Unused build cache/images were
  removed; application volumes and stream entries were preserved.

## Namespace-aware service identity and declared operation questions (2026-10-19, America/Los_Angeles)

[Service identity](docs/SERVICE_IDENTITY.md) now separates missing, present-empty
and literal OTel namespaces across discovery, reported parent ownership, fixed
record questions, native bindings and operation-package readers. Omitted
selectors retain legacy question/definition serialization and SQL. Registered
operation bindings generate ordinary span-record dashboard questions; schema URL
remains declared metadata because actual `DESCRIBE TABLE otel_traces` exposes
scope name/version but no schema URL column.

- `go test ./...` and `go vet ./...`: passed.
- Frontend suite: 49 tests passed.
- Hegel dashboard, namespace predicate, service suggestions and catalog properties:
  four checks passed (100 generated cases each). Inventory state/swarm regression:
  100 sequences of up to 20 steps passed.
- The initial browser campaign passed four journeys; inventory's final Quint
  comparison rejected its old two-filter expectation after namespace was added.
  The expected ordinary question now explicitly includes the namespace predicate.
- The subsequent campaign passed all four affected regression journeys. Journey
  60 completed its product and desktop/mobile assertions, then exposed an empty
  parameter-list import generated by the Quint harness. Its model now uses the
  existing explicit `INITIAL` constant convention.
- Final affected regressions: journeys 31, 46, 47 and 59 passed. Focused journey
  60 then passed, including missing/empty/named namespace dashboards, actual
  ClickHouse cross-namespace parent SQL, exact native discovery, declared-binding
  role exclusion, desktop/mobile containment, Quint observed replay and four
  rejecting mutations. The existing dashboard replay rejected eight mutations.
- Namespace Bombadil actor: 30-second campaign completed without violations.
  Ledger and actor output: `artifacts/service-namespaces/1792461286328/`.
- Refreshed affected screenshots and added journey 60 storyboard/frames, service
  identity diagram, dashboard design and product-plan updates.
- Final health/source checks: all 12 services running, query/packages image
  `sha256:10c97f0e3fda42c9d458c66fbd15b83d93f3bb79e063e3033da9cce08d2d1817`,
  served UI/binary hashes match source, all five pre-existing stale pending IDs
  preserved. Traffic restored; six total pending entries at the health read
  include active demo work. Evidence: `artifacts/namespace-final-health.json`.
  No Databricks instance was used or claimed.

## Automatic service overviews and distributed relationships (2026-10-20)

[Service dashboards](docs/SERVICE_DASHBOARDS.md) now use a pure Go planner over
supplied observed metadata plus matching server-registry declarations. Existing
record/native evaluators execute bounded widgets with two concurrent requests.
The Services distributed table reuses resolved parent evidence without a new
source read and labels child-span drilldown as a wider population. Widget questions,
explicit captures and shared revisions reuse existing investigation contracts.

Local checks passed: full root Go suite and vet (`dashboard-final-go.log`,
`dashboard-final-vet.log`), pure planning/registry HTTP checks, **48 frontend
tests** (`dashboard-final-node.log`), **100 requested Hegel dashboard planning
cases** and **60 saved-comparison state/swarm scenarios of 15 steps**
(`dashboard-hegel-final.log`), verification vet, Python server syntax compilation
and Quint model typecheck. Planning properties check deterministic order,
capability gating, exact scope/service retention and input detachment. Concrete
native tests reject mismatched units/bindings; unsupported temporality/kinds fail.

The first live campaign passed metric-only browsing but exposed mobile activity
bar overflow and a new test fixture without the kernel-required service instance
identity. The chart now uses at most 60 default intervals with shrinking bars and
no bar for zero-count intervals; the supported-native fixture supplies identity.
The older missing-identity fixture still exercises an unavailable widget rather
than weakening kernel validation. Instrument preview selectors now target the
instrument inventory rather than matching a similarly titled automatic widget.
The deployment ran out of disk during recreation; pruning unused Docker build
cache with the workspace Docker configuration restored space, without deleting
volumes or stream entries. Recreation completed with all services running.

The follow-up regression campaign passed inventory/package handoff, metric-only
browsing and full service-browser scope/cancellation checks (**3 passed**); journey
59's product checks and Quint replay passed, but its actor startup failed by
reading the fixture environment after navigation changed the URL hash
(`dashboard-final-journeys.log`). Correcting the actor to retain that environment
also allows it to inspect actual widget questions and native detail disclosures.
A filtered-list focus regression was fixed by focusing a real service-row action
or the search field, rather than a view-switch button. The new capture assertion
now checks the existing linked-revision status rather than a generic save label.

Final focused journey: **1 passed, 1.4 minutes**
(`artifacts/dashboard-final-focused.log`), with the three existing regressions
already passed on the same build. Real OTLP fixtures cover manifest-free spans,
logs and gauges, resolved and unresolved parents, exact widget question handoff,
immutable Run and capture, source-free fresh-browser reopening, partial log-source
failure and rejection of a late widget response after environment change. Registry
planning also preserves declared-only bindings without inventing observed data.
Frames 31, 46–48 were refreshed; new 59a–d show distributed relationships,
automatic widgets, native values/mobile layout and partial availability.

Observed ledger: `artifacts/service-dashboards/1792459008491/ledger.json`.
Quint replay passed with **eight rejecting mutations**, including invented edges,
changed unit/scope/selector, extra reads and an invalid action sequence. The final
**30-second Bombadil campaign** passed with no violations: 21 view switches,
17 service opens, 28 widget-question inspections, 18 native disclosure toggles
and 16 local returns. The actor validates scope, containment, reported edges and
absence of fabricated totals in failed widgets.

Final image:
`sha256:badca4da5a34205ba719c5db679119052ad578ba6cdac4ca060901e847f37bef`
(`dashboard-native-image.log`, `dashboard-native-up.log`). Final health, served
asset hashes and Go binary hash passed, with all **12 services running** and all
five pre-existing stale Redis pending entries preserved. There were six pending
entries at the health observation with normal traffic restored
(`artifacts/dashboard-final-health.json`); no stream entries or volumes were reset.

Limits: no whole-dashboard custody artifact, graphical/async/declared topology,
automatic SLOs, authoritative fleet membership or declared-only catalog route.
Application contracts retain declared revision IDs rather than claiming runtime
conformance. Record-service identity remains environment/service-name scoped;
native evaluations retain full OTel stream identities. Unsupported native source
identity/intervals stay unavailable/unknown. See the updated whole-product review
and known issues for the remaining roadmap.

## Captured attribute breakdowns and whole-product review (2026-10-20)

[Captured breakdowns](docs/COMPARISON_BREAKDOWNS.md) compare complete one-field
span/log group aggregates, preserve missing/empty/one-sided groups, and open
fixed-side focused questions without source reads. A new dimension requires two
explicit recaptures of the original fixed windows. Frozen receipts use algorithm
v3; v1/v2 receipts reopen without reevaluation. Results precede provenance,
with readable duration/error displays and internally scrolling mobile tables.

Local checks passed: full root Go suite and vet (`breakdown-go.log`,
`breakdown-vet.log`), explicit v1/v2 frozen-receipt compatibility
(`breakdown-legacy-go.log`), **45 frontend tests** (`breakdown-node-final.log`),
**100 requested Hegel partition cases** and **60 saved-comparison state/swarm
scenarios of 15 steps** (`breakdown-hegel-final.log`), verification vet, Python
server syntax compilation and Quint typecheck. The first Hegel invocation used
the default read-only home cache; setting the documented workspace XDG cache fixed
that environment error without changing the property.

The live OTLP route journey exercises 1200/1300 matching records and complete
aggregates beyond the 1000-row inspection cap, explicit recapture, source-free
focused questions, frozen sharing and mobile containment. Its first model replay
exposed the adapter's requirement for a parameterized module import; supplying the
initial-state constant fixed the harness. Replay then passed with six rejecting
mutations (implicit read, changed scope, altered group count, evidence-link drift,
wrong baseline, wrong selected group). An initial regression draft was stopped
to update selectors for the intentional provenance disclosure. A later visual
inspection found that the mobile table squeezed headers into vertical text despite
page containment; minimum readable table widths inside scroll areas fix it, and
the live test now requires an actual group value in view. Interrupted drafts are
retained as `breakdown-journeys.log` and `breakdown-mobile-draft-journeys.log`.

Final live campaign: **8 passed, no skips, 2.5 minutes**
(`artifacts/breakdown-final-journeys.log`). It covers the new route/log breakdowns,
existing record and native comparisons, fresh follow-ups, reusable exact questions,
saved findings, publication recovery and fresh-browser frozen reopening. Frames
52, 53, 56 and 57 were regenerated alongside new 58a–c; the corrected mobile frame
shows readable columns in an internal scroll area, with an actual group in view.

Observed ledger: `artifacts/comparison-breakdown/1792456617714/ledger.json`.
Quint replay passed with **six rejecting mutations**. The **30-second Bombadil
campaign** passed, recording 319 provenance-toggle and 275 focused-question
inspection actions without invariant failures. These inspect saved links and
receipt disclosure; the actor does not execute new telemetry reads.

Final image:
`sha256:476e83d5d4b7a7e537a30a5b4e4d55b803db6358496ea55b3e5ebfd95340445d`
(`breakdown-mobile-image.log`, `breakdown-mobile-up.log`). Final source/binary
hashes, HTTP health and all **12 running services** passed, with exactly the five
pre-existing stale pending Redis entries preserved
(`artifacts/breakdown-final-health.json`). Traffic was restored by the wrapper;
no volumes or stream entries were reset.

The [whole-product review](docs/WHOLE_PRODUCT_REVIEW.md) covers navigation,
discovery, service/distributed overviews, manifests, measurements and investigation
flows. Automatic service/distributed dashboards are **designed, not implemented**;
the next sequence is a bounded capability-driven planner, service defaults,
distributed relationships and investigation integration. The narrower
[comparison review](docs/COMPARISON_PRODUCT_REVIEW.md) records this feature's
usability rationale and remaining proposals.

## Fresh follow-up comparison (2026-10-20)

[Saved questions compare with fresh evidence](docs/FRESH_COMPARISON.md) through a
fixed same-duration preview, explicit compatible prior selection and one new
capture. The baseline is not refreshed. Existing stored comparison and finding
publication preserve the original question and evidence links. Native scopes
retain exact duration without inferring an SDK export phase.

Local checks passed: full root Go suite and vet (`fresh-comparison-go.log`,
`fresh-comparison-vet.log`), all **44 frontend tests**
(`fresh-comparison-node.log`), **100 requested cross-language fresh-scope cases**,
**100 immutable-revision cases**, **100 observed-support cases** and **60 saved
comparison state/swarm scenarios of 15 steps** (`fresh-comparison-hegel-final.log`),
verification vet and Quint typecheck. The fresh-scope property invokes the actual
browser projection in Node and validates it through Go's typed-step contract,
checking exact duration, definition retention and input immutability.

Early browser drafts corrected disabled-option and textarea-value assertions
after inspecting the DOM. The native mobile case then exposed a real overflow:
a long selected baseline label expanded its inline containing label. Block labels
and a width-constrained selector fix the layout without hiding content. Prior
scope, source-read time and custody time are now explicit in the preview.

Final live campaign: **12 passed, no skips, 5.4 minutes**
(`artifacts/fresh-comparison-final-journeys.log`). The new journey uses real
standard OTLP HTTP spans and system-memory gauges. It proves preview and baseline
selection cause no telemetry read; capture creates one new evidence request; a
forced comparison GET failure retries without recapture; changed-definition prior
evidence is disabled; original questions, evidence links and baselines remain
unchanged; and findings reopen source-free in a fresh browser. Native mobile
containment passes with the selected long baseline label.

Observed ledger: `artifacts/fresh-comparison/1792455384426/ledger.json`. Quint replay
passed with **six rejecting mutations**. The **30-second Bombadil campaign** passed
with 177 preview, 169 prior-selection and 181 close-preview actions. Frames 57a–d
show scope/clock preview, actual comparison, recipient finding and mobile native
follow-up. Affected frames 38 and 50–56 were regenerated. Existing shared
measurements, fixed reruns, linked history, saved comparisons and native support
discovery all passed in the same campaign.

Final image:
`sha256:f8cd5f6c377bed933fa5c336216ce7ad09999af32b79fbbc982afffb4a221161`
(`fresh-comparison-fixed-image.log`, `fresh-comparison-fixed-up.log`).
Final checks confirm all **12 services running**, HTTP health, served source asset
and Go binary hashes matching the tested build, and all five pre-existing stale
pending Redis IDs preserved (`artifacts/fresh-comparison-final-health.json`). The
wrapper restored the owned traffic generator; no volumes or stream entries were
reset.

## Reusable comparison questions (2026-10-20)

[Comparison sides become reusable questions](docs/REUSABLE_COMPARISON_QUESTIONS.md)
with a detached typed step, unique question key and exactly that capture linked.
Staging and shared reopening do not read telemetry. Existing immutable document
publication and fixed Run and capture append evidence in a child revision. Fresh
window preview now sits beside the time controls and retains the definition while
proposing a new scope, leaving the saved question unchanged.

Passed local checks: full root Go suite and vet (`reusable-questions-go.log`,
`reusable-questions-vet.log`), all **43 frontend tests**
(`reusable-questions-node.log`), **100 requested Hegel exact-scope/revision cases**
and **60 saved-comparison state/swarm scenarios of 15 steps**
(`reusable-questions-hegel.log`), verification vet and Quint model typecheck.
Frontend checks cover detached evidence/definition objects, exact native scope,
record questions and malformed/unsupported receipt rejection.

The first Quint draft used reserved action names; the renamed model typechecks.
The initial live run reached successful fixed rerun, then exposed that fresh-window
preview was hidden inside the package editor. That owned run was interrupted and
its wrapper restored traffic. Moving the existing preview button beside the time
controls makes the intended workflow directly accessible without opening YAML.

Final live campaign: **11 passed, no skips, 4.0 minutes**
(`artifacts/reusable-questions-final-journeys.log`). It covers record/gauge/native
comparisons, mixed question steps, linked evidence, shared measurements, reusable
questions, saved findings and observed-window discovery. Storyboard 56 uses real
standard OTLP counters with **123 ns** offsets. Adding both sides, retrying a
failed publication and reopening a shared exact question perform no telemetry
read; fixed Run and capture appends one evidence ID in a child revision; fresh
preview retains the definition and requires explicit execution for its new scope.
Earlier revisions and captures remain unchanged; mobile containment passed.

Observed ledger: `artifacts/reusable-questions/1792454583766/ledger.json`. Quint
replay passed and all **six deliberate mutations were rejected**. The **30-second
Bombadil campaign** passed with 120 baseline-question actions and 114 draft
revert actions, checking linked evidence and the card bound. Storyboard 56 frames and
affected frames 38 and 50–55 were regenerated from the running app. Architecture,
workflow docs and the implementation plan describe the implemented contract.

Final image:
`sha256:1bc22f7438d40cba68f4468177c3f6d9a5b518a234e1972bdc4284dd888f7f86`
(`reusable-questions-final-image.log`, `reusable-questions-final-up.log`).
Final health checks passed for all **12 running services**, HTTP health, Go binary
and served HTML/JS/CSS hashes matching tested sources, and preservation of all
five pre-existing stale pending Redis IDs
(`artifacts/reusable-questions-final-health.json`). The wrapper restored the owned
traffic generator; no stream entries or volumes were reset.

## Observed native window discovery (2026-10-19)

[Find comparable windows](docs/OBSERVED_WINDOW_DISCOVERY.md) reads validated
retained kernel intervals without a telemetry query. It intersects boundaries
across all retained streams, previews exact support and exclusions, requires
explicit selection on both sides and rereads two new exact questions. New results
are compared through the existing engine; original captures remain unchanged.

Local checks passed: full root Go suite and vet (`window-discovery-go-final.log`,
`window-discovery-vet-final.log`); all **42 frontend tests**
(`window-discovery-node-final.log`); **100 requested Hegel support properties**,
**100 native exact arithmetic/histogram properties** and **60 saved-comparison
state/swarm scenarios of 15 steps** (`window-discovery-hegel-final.log`);
verification vet and Quint model typecheck. Unit coverage includes source-offline
HTTP discovery, strict request validation, reset taint, all-member coverage,
asynchronous boundaries, layout changes, empty histograms, aggregate overflow,
10,000 intervals and explicit 32-window/64-exclusion output bounds.

The first Hegel invocation omitted the configured writable XDG cache and failed
before exercising properties; the configured-cache rerun passed. The first Quint
model used a built-in action name; renaming it fixed typecheck. The initial live
browser workflow and six rejecting Quint controls passed, but the new Bombadil
invariant rejected the pre-result null state. Its guard now allows that initial
state while continuing to reject an implicit comparable result.

Final live regression campaign: **6 passed, no skips, 3.7 minutes**
(`artifacts/window-discovery-final-journeys.log`). Real standard OTLP cumulative
counters and delta explicit histograms carry **123 ns** SDK offsets. Browser
checks prove discovery/selection do not read telemetry, both selections are
required, capture performs two explicit reads, exact shared questions restore
without telemetry reads, original captures remain unchanged and saved receipts
reopen in a fresh browser. A forced second-capture failure retains the first ID
and leaves the original comparison selected. Mobile containment passed.

The new observed ledger is
`artifacts/window-discovery/1792453916512/ledger.json`. Quint replay passed with
all six rejecting mutations; the **30-second Bombadil campaign** passed, recording
1 compare, 141 discovery and 170 selection/clearing actions. Existing record/gauge,
complete-native-window and saved-comparison journeys also passed. Frames 52–55
were regenerated as affected; storyboard 55 includes exclusions, exact selection,
actual recapture/comparison, mobile histogram support and partial recovery.

Final offline-built image:
`sha256:db18045a876d23a228443a9a6420d60c24fd1065988a244771523abf7c89f2e6`
(`window-discovery-final-image.log`, `window-discovery-final-up.log`). Health
checks confirm all **12 services running**, HTTP health, served JS/CSS/HTML and
Go binary hashes matching tested sources, and preservation of all five pre-existing
stale pending Redis IDs (`window-discovery-final-health.json`). The wrapper
restored the owned traffic generator; no volumes or stream entries were reset.
Only unused Docker build cache and dangling images were pruned to recover disk.

## Exact native scope transport (2026-10-19)

Implemented optional native nanosecond question bounds, parameterized Nano source
predicates, exact kernel evaluation, capture replay and shared-step round trips.
Passed full root `go test ./...` and all 41 frontend tests, including exact-bound
plan/evaluation/replay and shared-link/capture round trips with rejecting controls.
Logs: `artifacts/exact-native-scopes-go.log` and `exact-native-scopes-node.log`.
The discovery picker is not implemented in this milestone. No live deployment
or screenshot regeneration was performed; no new interactive control is added.

## Complete native window comparisons (2026-10-19)

[Native window comparison](docs/NATIVE_WINDOW_COMPARISON.md) uses complete validated
coverage for exact OTel streams. Counter rates use exact window increments and
actual durations; histogram comparisons combine compatible bucket populations
and retain p95 ranges. Missing members, resets, gaps, incomplete boundaries,
temporality changes, incompatible layouts and histogram count overflow withhold
changes. One unsupported member withholds all deltas. New receipts use algorithm
v2; v1 snapshots remain readable without reevaluation.

Passed local checks: full root Go suite including counter exact totals beyond
uint64, histogram overflow refusal, whole-population p95s, multi-stream withholding
and v1 receipt compatibility (`artifacts/native-windows-go-final.log`); 40 frontend
checks (`native-windows-node.log`); 100 requested Hegel exact-rate,
share-conservation/layout-refusal, reversal and immutability cases plus 60 saved
comparison state/swarm scenarios of 15 steps (`native-windows-hegel-complete.log`);
verification vet and Quint model typecheck (`native-windows-vet-final.log`,
`native-windows-model.log`). An initial property filename used the Go Windows-only
suffix and was renamed before accepting results. The first actual run then exposed
an incomplete test stream fixture; adding its required instrumentation scope fixed
that fixture. Both named Hegel tests executed and passed afterward.

Focused live journey: **1 passed, no skips, 1.3 minutes**
(`artifacts/native-windows-focused.log`). It ingests real standard OTLP delta and
cumulative sums and explicit histograms, compares complete windows, shares a v2
snapshot in a fresh browser, checks mobile layout and refuses counter decrease,
coverage gaps and bucket mismatch. Live Quint replay passed with six rejecting
mutations; 30 seconds of Bombadil frozen-result exploration passed. Initial ledger:
`artifacts/native-windows/1792448783045/ledger.json`.

Compatibility campaign: **5 passed, no skips, 3.6 minutes**
(`artifacts/native-windows-release.log`), covering record/gauge comparison 52,
native windows 54 and saved sharing/recovery 53. All enabled Bombadil exploration
and live Quint checks passed. Visual review then improved native table scrolling
on mobile and reused the existing captured-attribute viewer for stream identity.
The final affected journey passed again (**1 passed, no skips, 1.1 minutes**,
`native-windows-polish-focused.log`), including attribute disclosure and actual
horizontal scrolling, plus six rejecting trace controls and Bombadil. Its ledger
is `artifacts/native-windows/1792449184577/ledger.json`. Five new focused frames and
affected 52/53 screenshots are updated; all new frames were visually reviewed.

Final health confirms 12 running services, functional query health, restored owned
traffic, all five pre-existing stale Redis pending IDs retained, and served assets
matching source (`native-windows-final-health.json`, `native-windows-served-assets.json`).
Final query image:
`sha256:f234be3fd02517e080c77f1e94c7bb4c5f15b8e3c525c1f052763cf48bbcac29`.
Exact native coverage remains deliberately strict; millisecond question bounds
can reject nanosecond-offset interval edges, as documented in known issues.

## Saved comparison evidence (2026-10-19)

[Saved comparison cards](docs/SAVED_COMPARISONS.md) freeze the server-verified
ordered capture IDs, displayed decision/result and algorithm version. Findings
revise in immutable documents. Shared investigation links read the saved snapshot
without querying telemetry or reevaluating the comparison. Retained receipt IDs
support explicit recovery after context changes; failed document publication
retains the draft and finding.

Passed local checks: full root Go suite including source-free HTTP publication,
forged-decision rejection, missing references, integrity checks and preserved
finding history (`artifacts/saved-pair-go-http.log`); 40 frontend checks
(`saved-pair-node.log`); 60 requested Hegel state/swarm scenarios of 15 steps
(`saved-pair-state.log`); verification vet (`saved-pair-vet-final.log`), Quint
model typecheck (`saved-pair-model.log`) and Python server compilation.

The first live campaign completed all sharing, recovery, mobile and native UI
assertions, and Quint trace validation with six rejecting controls, but its new
Bombadil actor used an overly broad last-child selector that clicked Remove.
Targeting the saved comparison's direct reopen button and adding a card-count
invariant corrects the test actor. That campaign reports one failed/one passed;
its passing ledger replay is
`artifacts/saved-comparison/1792447998043/TestQuintSavedComparisonLive.log`.

Final live campaign: **5 passed, no skips, 3.1 minutes**
(`artifacts/saved-pair-release.log`): notes/context 36, both comparison journeys 52,
and both new sharing journeys 53. All enabled Bombadil campaigns and live Quint
validations passed. The final ledger is
`artifacts/saved-comparison/1792448210686/ledger.json`; its six rejecting controls
cover implicit reads, changed receipt references, forged decisions, lost findings,
false publication transitions and reversed capture identity. Both native gauge
snapshots and frozen unit incompatibility reopen in a fresh browser without engine
requests. Four focused storyboard screenshots were captured and visually reviewed;
affected 36/52 frames were regenerated.

All 12 services are running, query health passed, owned traffic was restored, and
all five pre-existing stale Redis pending IDs were retained
(`artifacts/saved-pair-final-health.json`). Served document/comparison scripts and
styles match source (`saved-pair-served-assets.json`). Final query image:
`sha256:e3dc20bac4df747e4230b10be5747cc6dbbf4d19b19118f669740d7b01823bc7`.

## Stored capture comparison (2026-10-19)

[Stored comparison](docs/CAPTURE_COMPARISON.md) loads immutable capture IDs and
compares full record counts, span duration/error summaries and exact-stream native
gauge endpoints without telemetry reads. Incompatible definitions, environments
or units withhold deltas. Empty/unknown results retain their uncertainty; native
counter/histogram/lifecycle results remain inspectable side by side.

Passed local checks: full root Go suite, 40 frontend checks, 100 requested Hegel
precision/reversal/immutability/compatibility cases, verification vet, Quint
model typecheck and Python server compilation. Logs:
`artifacts/evidence-pair-{go,node,hegel-ready,vet,model}.log`.
The live ledger replay passed with four rejecting controls
(`artifacts/evidence-pair-quint-replay-ready.log`). The first trace-validation
attempt exposed the checker's generated empty-argument import for a model without
constants; declaring the explicit initial read count fixed the model adapter.
The first Hegel attempt used an unwritable default cache; setting the documented
workspace `XDG_CACHE_HOME` fixed the setup failure. The first live browser campaign
found a missing entry for the new script in the Python server's static asset
allowlist; the corrected build serves source-matching assets.

Final live campaign: **5 passed, no skips, 3.4 minutes**
(`artifacts/evidence-pair-release.log`). This covers saved captures 32, both new
comparison journeys 52, and both linked-history journeys 51. Enabled Bombadil
campaigns and live Quint validations passed. The new observed ledger is
`artifacts/evidence-comparison/1792447023081/ledger.json`; its four negative
controls reject hidden reads, incompatible deltas, forged arithmetic and changed
receipts. Final native expansion retains explicit Baseline/Comparison labels.
Storyboards 32/51 were refreshed with the changed UI, and four new 52 screenshots
were captured and visually reviewed (record, incompatible, mobile, native).

Final health checks confirm 12 running services, a functional query health request,
restored owned traffic, all five pre-existing stale Redis pending IDs preserved,
and source-matching served HTML/JS/CSS (`evidence-pair-final-health.json`,
`evidence-pair-served-assets.json`). Final query image:
`sha256:8e031dc91f396851ce4e9865975b048d9da9df1d60cabf10a07842d86116ed03`.
Build disk pressure was resolved by clearing unused build cache and dangling
images; telemetry volumes and pending stream entries were retained.

## Linked question evidence (2026-10-19)

[Linked evidence](docs/LINKED_QUESTION_EVIDENCE.md) gives saved questions an
explicit Run and capture action, immutable capture history, stored result counts
and source read times, and visible question/scope differences. Partial publication
preserves the capture ID for explicit recovery without another source read.

Passed: full root Go suite, 40 frontend checks, verification vet, 100 requested
Hegel cases comparing the actual JavaScript scope/definition logic, and 60
requested state/swarm scenarios of 15 steps preserving immutable revisions.
Logs: `artifacts/linked-go-summary.log`, `linked-node-ready.log`,
`linked-verification-vet.log`, `linked-hegel.log`, `linked-state-final.log`.

Both new browser journeys passed (no skips, 1.1 minutes): record capture history,
refined-question drift, failed-publication recovery, stored reopening, mobile
layout, and native capture with one source execution. The live ledger passed Quint
trace validation with seven rejecting mutations; 30 seconds of Bombadil exploration
passed without implicit source reads (`artifacts/linked-release-focused.log`,
`artifacts/linked-evidence/1792445673965/ledger.json`). The first focused run exposed
mobile overflow from an unwrapped capture ID; using the existing reference-ID
style fixed it, and the final mobile assertion passed. Storyboard 51 and its five
screenshots document the final behavior.

Final compatibility campaign: **4 passed, no skips, 5.2 minutes**, covering
saved captures 32, investigation notes 36, mixed steps 50 and measurements 38,
with their enabled Bombadil campaigns and live trace validations
(`artifacts/linked-regression.log`). Refreshed affected screenshots accompany the
new storyboard. All 12 services are running, owned traffic was restored, all five
pre-existing stale Redis pending IDs were preserved, and the three served UI assets
match the tested source (`linked-final-health.json`, `linked-served-assets.json`).
Final query image: `sha256:495e49a385806d103d9c398b3dda90f1c245b9c0490b2217bfde524478ffac3e`.

## Typed investigation steps (2026-10-19)

[Question steps](docs/INVESTIGATION_STEPS.md) persist record questions and complete
measurement definitions in named immutable investigation revisions. The common
versioned scope/view envelope rejects mismatches between a record/definition and
its declared environment/window. Question intent stays separate from frozen
capture references. Both record and native question URLs stage explicit Run again;
native reopening performs definition planning without telemetry reads. Legacy
record URLs and existing stored document IDs remain compatible.

Passed local checks: full root Go suite (including mixed-card validation and
immutable predecessor checks), 39 frontend checks, verification vet, and 100
requested Hegel cases through the actual JavaScript envelope codec. Logs:
`artifacts/steps-{go,node-final,vet-final,hegel}.log`.

The first focused browser run completed all mixed-question UI assertions, source
reads, capture attachment, second-browser reopening and screenshots, but its new
Quint model used a reserved action name. Renaming the modeled explicit-read action
fixed it. Replaying that real ledger passes, with seven rejecting controls for
scope drift, native unit/identity changes, hidden reopens/recipient reads, false
window equivalence and wrong observed read transitions (`steps-quint.log`). Both
existing question-link checks passed in that initial focused run
(`steps-focused.log`). The model checks observed save → reopen → explicit-read
transitions, preserving source read counts across saving and reopening.

Final live campaign: **11 passed, no skips, 11.7 minutes**
(`artifacts/steps-release.log`). This covers coherence 40, notes/context 36, core 33,
inventory 31, new mixed questions 50, measurements 38, metric-only services 47,
both question-link cases 49, workbench 45 and service browser 46/48. All enabled
Bombadil campaigns and associated live Quint validations passed. The new actor
returns to the active investigation and reopens each question type without running
it. The mixed-step ledger is
`artifacts/investigation-steps/1792444529620/ledger.json`.

The existing Hegel investigation state/swarm model now also creates record/native
question cards and revises detached question scopes after saving/forking. Its
60 requested scenarios of 15 steps passed, preserving all prior revisions and
context semantics (`steps-state-swarm.log`); verification vet passed again afterward.
New frames 50a–c and affected existing frames were regenerated from real telemetry.
The final desktop investigation and mobile native-question frames were visually
reviewed. README, architecture/process flow, contracts, plans and known boundaries
are synchronized.

Deployed image:
`sha256:5bae60c988431e0ec7edcbc9dab734428fe303e16960f1a3270f9d0eb06f4fd2`.
All 12 services run, traffic is restored, all five stale pending entries remain,
and served navigation assets match the checkout (`steps-final-health.json`).
No telemetry volumes were removed.

## Resumable record-question links (2026-10-19)

[Question links](docs/QUESTION_LINKS.md) preserve fixed record scope, selectors,
grouping, measures and Records/Distribution view. Reopening or traversing history
stages Run again and clears live evidence without a source read. Capture receipt
links remain a separate frozen-evidence workflow. Native drafts, service-detail
panels and unsaved keystrokes are outside this first link format.

Passed: full root Go suite, 38 frontend checks, verification vet, and 100 requested
Hegel cases through the actual URL codec and session state. The observed browser
read ledger passes quint-go validation and rejects three mutations (restore read,
foreign fixed window, history read). Logs: `artifacts/navigation-{go,node-final,
vet,hegel-final,quint}.log`.

The first browser run completed its UI assertions and screenshots but failed on a
new Quint model import with no constants. Adding its explicit initial-state constant
fixed that setup. The first Hegel invocation used the read-only default cache;
rerunning with the established XDG cache path passed. Initial logs remain in
`navigation-journey.log` and `navigation-hegel.log`.

Final browser evidence: seven existing journeys passed across
`artifacts/navigation-release.log` (captures, workbench, service browser) and
`artifacts/navigation-final.log` (coherence, core, inventory, metric-only services).
Their seven enabled Bombadil campaigns and live Quint checks passed. The new
link journey initially exposed an actor trying Back before creating in-app history;
its actions now stay within their own entries. A retry hit Bombadil's existing-output
safeguard; runs now retain separate trace directories rather than overwrite evidence.
Those earlier failures remain in the two campaign logs.

The final image adds sender-side draft-window pinning. Both link browser tests
passed on it, including the corrected eighth Bombadil campaign and live Quint check
with all three rejecting mutations (`artifacts/navigation-links-final.log`,
52.5 seconds, no skips). The final frontend rerun again passed all 38 checks
(`navigation-node-shared.log`). Storyboard 49 uses actual populated `lab-web`
telemetry, a second browser, fixed-window history, a duration refinement, malformed
links, view retention and mobile containment. Existing affected frames 31–33,
40 and 45–48 were refreshed. Desktop and mobile frames were visually inspected.

Final deployed image:
`sha256:9d761f84bbe610e2a9cb4fee88fbfbc7be5e7009551ad7b232030071c5e00c97`.
All 12 services run, traffic is restored, and all five stale pending entries remain
(`artifacts/navigation-final-health.json`). Unused Docker build cache was pruned;
telemetry volumes were retained. Served navigation assets match checkout hashes.

## Unified Services catalog (2026-10-19)

[Services](docs/UNIFIED_SERVICE_CATALOG.md) now has one discovery action and local
search, with one row per observed service. Span/log/native point counts retain
separate grains and exact decimal strings; absent observations are not zeros.
Service detail independently reads signals, span attributes, native instruments
and environment-wide operation inventory in the original catalog window. Operation
inspection and capture actions moved inside detail. The inventory capture label
explicitly describes its original whole-environment scope. Saved inventory rendering
is source-free; capture formats, backend kernels and query semantics are unchanged.

Passed: root Go suite, 36 frontend checks, Python compilation and verification vet.
Hegel ran 100 requested catalog-grain cases plus 100 suggestion-scope cases through
the actual JavaScript planners. The catalog property includes UInt64 values beyond
JS number precision and service/signal grouping. Logs:
`artifacts/catalog-{go,node,hegel,vet}.log`.

Final live integration: all seven journeys passed in 11.3 minutes, without
skips (`artifacts/catalog-release.log`): captures 32, coherence 40, core basics 33,
inventory 31, native-only services 47, question workbench 45 and service/catalog
46/48. All seven enabled Bombadil campaigns and associated live Quint checks
passed. The updated service ledger rejects eight mutations, including summed/mixed
signal counts and a foreign catalog window. The browser also deliberately tampers
a real catalog response's window and confirms rejection before publication.
Ledger: `artifacts/service-browser/browse-aafd713e/ledger.json`.

Storyboards 31/32/33/40/45/46/47 were migrated and affected frames regenerated.
New frames 48a–c show one entry, retained filtered return and mobile catalog;
desktop/mobile catalog frames were visually reviewed. The list-return path checks
read-free navigation and search focus when no row matches. Architecture/process
diagrams, README, current contracts, planning and known boundaries are updated.
Final deployed image: `sha256:2ce745c0a8e180bd06aba310c7f086b5cc34880fcc557376cf7963c538b0a469`.
All 12 services are running, traffic is restored, and all five pre-existing stale
Redis pending entries remain (`artifacts/catalog-final-health.json`). No telemetry
volumes were removed.

The first focused run passed the native-only journey, then found an old browser
test trying to reopen a service while its catalog was hidden behind detail. The
journey now returns to the catalog explicitly before opening again. That initial
attempt is retained in `artifacts/catalog-focused.log`.

## Metric-only service discovery (2026-10-19)

[All-signal service discovery](docs/METRIC_SERVICE_DISCOVERY.md) now includes native
gauge, sum and explicit-histogram data-point records. It groups presence by service
without combining metric values, units or histogram observations. A direct Services
entry opens the catalog; metric rows open the original fixed window in service
browser/native measurement flow, never as a record query. Dependency inference
remains trace/parent-span only. Summary/exponential-histogram families remain outside
the current adapter, and no complete fleet or health claim is introduced.

Passed: full root Go suite, discovery race checks, root and verification vet,
35 frontend tests, Python compilation and Compose configuration validation.
Hegel generated 100 requested valid scope/count/bound cases and verified rejection
of the exclusive upper event bound. Unit cases cover malformed/duplicate rows,
zero counts, event bounds, UInt64 decimal counts, parameter isolation and the
1000-pair global cardinality budget. Quint model typechecking passed.
Logs: `artifacts/metric-services-{go,race,root-vet,vet,node,hegel}.log`.

Final integration passed: core basics 33, native metric discovery 39, new
metric-only services 47 and service browsing 46 in 4.8 minutes
(`artifacts/metric-services-release.log`), plus inventory 31 in 1.1 minutes
(`artifacts/metric-services-inventory.log`). All six enabled Bombadil campaigns
and associated live Quint validations passed. New storyboard 47 validates the
actual ledger with seven rejecting mutations: point count, window, service identity,
preview read, local-return read, forged dependency and exclusive event bound.
The histogram with 5000 observations contributes one of four point records;
gauge execution returns its two selected points. Local search remains applied
after an explicit catalog refresh, without altering source totals.
Ledger: `artifacts/metric-services/metric-services-195c85b9/ledger.json`.

New frames 47a–d and affected inventory/core/native/service-browser frames were
regenerated. The new desktop catalog and mobile service detail were visually
reviewed. README, source/process diagram, implementation plan, known boundaries
and executable storyboards are synchronized. Final image:
`sha256:fa69ff90cb9bfb46bac4197878b845c0df03817efca3fedbcb27e197cf6f6f18`.
All 12 services are running, traffic is restored and all five pre-existing stale
Redis pending entries remain (`artifacts/metric-services-final-health.json`).

The first live attempt confirmed cross-family record counts and exact boundary
inclusion using independent ClickHouse SQL, then failed native execution because
the new fixture omitted `service.instance.id`. The fixture now supplies that
required stream identity; the native validator remains intact. The initial attempt
is retained in `artifacts/metric-services-journey.log`. Disk pressure during image
recreation was resolved by removing unused Docker build cache, without removing
telemetry volumes or running images (`metric-services-cache-prune.log`).

## Browse-first service investigations (2026-10-19)

[Service browsing](docs/SERVICE_BROWSING.md) now connects fixed-window service
inventory and span/log discovery with observed-attribute questions and native
instrument previews. Opening a service explicitly reads three independent bounded
sources; unavailable sources remain separate. HTTP/database/messaging suggestions
use exact observed OTel attribute keys (including documented aliases), add an
explicit exists predicate, and enter the shared Explore executor. Native previews
retain instrument identity and fixed bounds; unsupported kernels cannot execute.
Back to service is local. Scope changes abort reads and reject late publication.
Metric-only fleet enumeration and durable navigation history remain unimplemented.

Passed: 35 frontend tests; Python compilation; verification vet; Quint typecheck;
Hegel (100 requested generated cases) invoking the actual JavaScript suggestion
planner with generated windows and preferred/legacy/absent keys. Storyboard 46
passed with real OTLP spans/logs/metrics through ClickHouse, a 30-second Bombadil
campaign, and observed quint-go ledger validation with six rejecting mutations.
The journey checks narrowed HTTP count 4 of 6 spans, preserved service/window,
source-free native preview, read-free return, unsupported nonmonotonic sums,
log-only service browsing and cancellation of a delayed real metadata response.
Evidence: `artifacts/service-{node,hegel,vet,final-journey}.log` and
`artifacts/service-browser/browse-9939f5b3/ledger.json` (release rerun).

Existing core basics 33, inventory 31, metric discovery 39 and question workbench
45 all passed in 4.5 minutes (`artifacts/service-regression.log`), including five
enabled Bombadil campaigns and their live Quint validations. These ran before the
final two presentation fixes for local service-button filtering and unread
operation inventory wording; storyboard 46 checks both on the release image.

The release rerun passed (`artifacts/service-release-journey.log`), including
local service-button filtering without reads and explicit unread operation inventory
for log-only service detail. Its live Quint validation and Bombadil campaign passed;
all six distinct enabled campaigns across this slice and regressions succeeded.
Final image: `sha256:f51dff13fb845ad45b648608bbadeee36af1cea13dab41b6aa1f692e029b50f1`.
All 12 services are running, traffic is restored and all five pre-existing stale
Redis pending entries remain (`artifacts/service-final-health.json`). Only dangling
Docker images were pruned for space; no telemetry volumes were removed.

Frames 46a–d document service detail, fixed-window record exploration, native
measurement preview and mobile detail. Final desktop/mobile service frames were
visually reviewed. The detail replaces the operation table while open, with
question/instrument cards arranged in desktop grids and mobile columns. README,
architecture/process diagram, implementation plan and known boundaries are updated.

Early integration attempts found two fixture/assertion mistakes: monotonic delta
counters are supported (the unsupported fixture now uses a nonmonotonic sum), and
measurement environment is nested in the definition rather than the question root.
Those attempts remain in `service-journey.log` and `service-journey-rerun.log`;
the final successful run is distinct. No backend kernel or persisted schema changed.

## One scoped question workbench (2026-10-19)

The first [entry-point design slice](docs/QUESTION_WORKBENCH.md) is implemented:
shortcuts, advanced filters and measures share one Run question executor. Draft,
requested and applied scopes have generation ownership; editing cancels pending
reads and clears stale evidence actions. Ordinary edits and explicit drilldowns
retain fixed bounds. Fresh-window staging performs no read. Evidence capture and
reusable definitions have separate labels, and Services → native metric discovery
preserves inventory service/environment/bounds without asserting the same population.

Passed: full root Go suite, 34 frontend checks, Python compilation, verification
vet, and Hegel state/swarm (100 requested sequences × 25 steps) replaying the actual
JavaScript session against an independent Go oracle. Quint typechecking and the
new live session/evidence ledger validation passed, including eight rejecting
controls for generation, scope, bounds, local navigation reads and frozen custody.
Storyboard 45 passed through the real OTLP/ClickHouse pipeline, delaying an actual
response to check stale completion rejection; its 30-second Bombadil actor passed.
The structured-attribute desktop/mobile polish journey 29 also passed. See
`artifacts/entry-{go,node,hegel,vet,ledger,focused}.log`; the latest workbench ledger
is `artifacts/question-workbench/entry-6f69f0e8/ledger.json`.

The final regression results cover nine distinct browser journeys: new workbench
45 and attribute polish 29 (`entry-focused.log`), core basics 33, follow-ups 42,
grouped measures 34, grouped time series 43 and recipes 37 (`entry-regression.log`),
plus corrected discovery 30 and numeric measures 44 (`entry-rerun.log`, two passed
in 2.8 minutes). All nine enabled Bombadil campaigns and their associated live
Quint checks passed across those successful journeys. Discovery includes both
its Playwright and Bombadil-observed ledger validations. Existing affected frames
were regenerated. The numeric rerun opens the relocated measure disclosure rather
than waiting on its old direct-child selector.

The final deployed image is
`sha256:628c8735182a85cb1b25556b4a8cfd2f66bc82fd6df8601c7926a4a509e359c3`.
All 12 services are running, traffic is restored, and all five pre-existing stale
Redis pending entries remain (`artifacts/entry-final-health.json`). No telemetry
volumes were removed; only dangling Docker images were pruned for disk space.

Desktop/mobile frames 45a–d document initial entry, applied scope, metric handoff
and frozen evidence. The initial and applied desktop layouts were visually
reviewed; the frozen mobile layout was also visually reviewed. The broader
navigation/Library migration, native browse-first flow and durable navigation/draft
recovery remain proposed, rather than claimed complete.

Early validation found and corrected a Quint action-name collision, Go omitted
empty predicate defaults in returned question identity, service-shortcut loss
when switching signals, and older tests acting on hidden result/disclosure controls.
The first focused discovery attempt also assumed the first log had a trace link;
the mobile check now explicitly selects a linked log. The exploratory discovery
actor also retained a reuse expectation after intentionally changing the breakdown;
its explicit breakdown actions now clear that old intent. These failed attempts are
retained in the local artifact logs.

## Live numeric attribute measures (2026-10-19)

One optional explicit int64/uint64 string-map binding now executes through the Go
ClickHouse adapter, Explore summaries/breakdowns/time charts, immutable captures
and reusable YAML. Missing/empty/invalid values retain separate counts; only valid
values enter exact sum/extrema/rational means. Signed chart rankings use labelled
absolute magnitude with signed values retained. Preflight now reports supported
bindings executable. Multiple numeric bindings, float/decimal parsing, conversion
and numeric cohort comparisons remain unsupported. See [contract and process diagram](docs/NUMERIC_ATTRIBUTE_PLANNING.md).

Passed: full root Go suite; exploration/HTTP race tests; root vet; 29 frontend
checks; Python compilation; Hegel partition conservation and numeric state/swarm
(100 requested cases/sequences, 25 state steps), plus existing package state/swarm.
Native checks cover parameterized keys/units, exact type bounds, impossible summary
states, missing source summaries, same-count sum disagreement, capture rejection
and YAML/typed-plan round trips. Three retained historical source-free plans still
match SQL, canonical YAML and package revisions exactly on the final image.

Controlled integration sends 128 ordinary OTLP spans and 128 logs through the real
collector/exporter. Signed values produce 116 valid, one missing, one empty and ten
invalid records; exact sum is 9223372036854776141. Unsigned and resource bindings,
all groups and every interval are checked against the Go reference, independent of
the 100-row page. A live six-stage custody ledger passed quint-go/Quint validation
and rejected eleven mutations, including coverage, units/parsing, sum, cell bounds,
local reads and reopened/reused evidence. All six final browser journeys passed
without skips in 7.5 minutes: core basics 33, follow-ups 42, grouped measures 34,
grouped time series 43, numeric measures 44 and recipes 37. All seven enabled
Bombadil campaigns and associated live Quint checks passed. Final numeric ledger:
`artifacts/numeric-measures/numeric-deeb5719/ledger.json`. Desktop/mobile numeric
frames and affected existing frames were regenerated; both numeric layouts were
visually reviewed. All 12 services are running, traffic is restored, and inspection
confirmed all five pre-existing stale pending entries remain.

An early browser attempt was interrupted after a time-control blur moved the
editor click target; the journey now waits for the layout change. A subsequent
attempt completed the UI path but launched an unrelated system `go`; the final
campaign uses the pinned Go toolchain. An immediate post-deploy probe also hit a
startup connection reset; the ready-service preflight passed. Visual review led
to wider exact-value tiles and a clear unsupported numeric-comparison message;
enabling a numeric measure clears the earlier count baseline.

Final image: `15d9fcb7d278`. Artifacts: `artifacts/numeric-live-{go,race,vet,node,hegel,compatibility,preflight,ledger,regression}.log` and
`artifacts/numeric-measures/*/ledger.json`. Existing volumes and the five stale
pending entries are retained. Screenshot/storyboard updates are in journey 44
and affected regression frames. No environment configuration changed.

## Numeric attribute preflight and reference semantics (2026-10-19)

Added explicit int64/uint64 attribute/resource bindings, canonical-decimal parsing,
declared units, exact reference sum/extrema/rational-mean states and separate
valid/missing/empty/invalid counts. A source-free HTTP preflight reports structured
capability diagnostics and always marks execution unavailable. Live numeric SQL,
UI controls and package/capture persistence remain future work; existing executable
question/package/capture formats are unchanged. See [contract and process diagram](docs/NUMERIC_ATTRIBUTE_PLANNING.md).

Passed: full root Go suite, exploration/HTTP race tests, root vet, Python compilation
and Hegel partition conservation (100 requested cases). Tests include signed and
unsigned limits, sums above uint64, zero versus unavailable extrema, malformed
syntax, bad unit labels, detached scope and strict HTTP decoding. Reproducible live
proxy checks passed with `python3 scripts/check_numeric_preflight.py`, including
supported declarations, multiple unsupported diagnostics and unknown-field rejection.
No UI behavior changed, so browser journeys/screenshots and Bombadil/Quint live
ledgers were not regenerated for this source-free milestone.

Artifacts: `artifacts/numeric-plan-{go,race,vet,hegel,image,up,live,services,pending}.log`.
Deployed image `bb4ee24d75af`; all 12 default services remain running, including
traffic. Volumes were preserved. Inspection confirmed all five pre-existing stale
pending entries remain; an additional ordinary in-flight entry was present with
traffic active. No telemetry reset or traffic pause was performed.

## Typed record measures and grouped time charts (2026-10-18)

Existing record controls now expand into typed count, duration and recorded-status
operators used for SQL lowering. Input/unit/operator/percentile-profile checks fail
closed; arbitrary attribute casts and general expressions remain unsupported.
Optional `groupTimeSeries` adds bounded group/interval evidence, common-scale local
chart controls, capture/reopen and definition reuse. See [the contract](docs/RECORD_MEASURES.md).

Passed: full root Go suite, exploration/HTTP race tests, root vet, 32 frontend
checks, Python compilation and expanded Hegel state/swarm (100 requested sequences
of 25 steps each for grouped-series, recipes and follow-ups). Native tests cover
missing/empty dimensions, no-data cells, duplicate/invalid bounds, cardinality/cell
budgets, same-count duration disagreement, signal-specific lowering and unsupported
measure bindings. Three source-free plans retained from the prior deployed image
match the final image's SQL, package revisions and canonical YAML exactly.

The initial real-stack storyboard passed in one minute with seven rejecting Quint
controls and its Bombadil campaign (`artifacts/grouped-time-series/series-d6d2af76/`).
It sends 124 ordinary OTLP server-shaped spans through the real collector/exporter,
compares four route-value series with a 100-row inspection page, checks local scale
and no-read guards, and captures/reuses/reopens the evidence. This is controlled
integration telemetry, not production traffic or a convention-coverage certificate.
Screenshot inspection exposed a checkbox style issue, corrected before regression.

The build filled the workspace disk. Regenerable Go compilation cache, dangling
images and unused Docker build cache were removed; telemetry volumes and running
application data were preserved, restoring approximately 14 GiB free before caches
were warmed. An intermediate test-file editing error was caught by compilation and
corrected; the final full native suite passed. No environment configuration changed.

Artifacts: `artifacts/series-{go,race,vet,node,hegel,compatibility}.log`,
`artifacts/series-journey.log`, `artifacts/series-regression.log` and the live
storyboard ledgers. Final image: `bfb62c257374`. All five affected browser journeys passed in 5.9
minutes, without skips: core basics 33, follow-ups 42, grouped measures 34,
grouped time series 43 and recipes 37. All six enabled Bombadil campaigns and
associated live Quint checks passed. The final grouped-series ledger rejected all
seven mutated controls: `artifacts/grouped-time-series/series-68ffb67e/`. Final
ledger replay also passed after strengthening exact integer accounting and total
series validation (`artifacts/series-ledger-final.log`). Affected
desktop/mobile frames were regenerated. Traffic was restored; existing telemetry
volumes and all five stale pending entries were retained.

## Shared question scope and checked transforms (2026-10-18)

Recipes and capture-derived follow-ups now use shared detached Go planning metadata.
Same-population/same-question and exact-ID relationship claims are checked against
actual typed questions; grouping/measure/chart/page changes are computed. Record
scope excludes presentation paging and retains complete typed OTel/custom selectors.
Native measurement bindings retain declared schema/library/instrument identity;
unknown record schema or convention profile remains explicit. Inventory and
comparison adapters retain underlying grain and separate member windows.

Passed: full root Go suite; exploration/HTTP race tests; root vet; 32 frontend
checks; Python compilation; expanded Hegel recipe and follow-up state/swarm
(100 requested sequences each, 25 steps). Checks cover detached ownership,
OTel/custom namespaces, record/package/wire scope roundtrips, native binding
identity, false-claim rejection and unchanged capture serialization. The first
Hegel invocation used the unwritable default cache; rerunning with the existing
workspace XDG cache passed. No environment configuration change was required.

Artifacts: `artifacts/scope-{go,race,vet,node,hegel}.log`. The deployed offline
image is `15330826a024`. Initial image `b501ce099b8c` passed all three affected
browser journeys in 3.7 minutes, with all enabled Bombadil and live Quint checks.
Final review separated metric signal from gauge/counter/histogram measurement kind;
root/race/vet/Hegel checks passed again after that correction. Final image passed both planning journeys again (2.4 minutes), with no skips,
including both Bombadil campaigns and live Quint checks. Follow-up replay rejected
all nine mutated ledgers; recipe replay checked the full shared transform plan.
Final evidence: `artifacts/scope-final-journeys.log`,
`artifacts/followups/47900071/`, `artifacts/recipes/c1bc9301/`. The earlier broader
run is `artifacts/scope-journeys.log`. Desktop/mobile screenshots were regenerated;
the desktop follow-up preview was inspected. Traffic was restored after both
wrappers; existing telemetry volumes and all five stale pending entries remain.
The [contract](docs/QUESTION_SCOPE.md), diagrams, planning docs and existing
storyboards describe this bounded slice. General typed measures, convention-profile
resolution and full backend capability negotiation remain future work.

## Core primitives review (2026-10-18)

Reviewed the implemented question/package/capture, workflow/recipe/follow-up and
investigation models. [The consolidation plan](docs/CORE_PRIMITIVES.md) records
missing shared contracts, migration boundaries and next-slice acceptance criteria.
The plan also explicitly requires OTel convention profiles across all six primitives,
with versioned bindings, optional-field degradation and conformance cases.
Documentation only: changed-file local links and `git diff --check` passed; no
runtime changes, new browser claims or application test reruns in this review.

## Contextual investigation follow-ups (2026-10-18)

Evidence cards now propose source-free questions from verified captures, with
explicit preserved and changed scope. Go compiles generated record/measurement
questions; exact trace/span IDs remain distinct from service/time resource
context. Discovery retains the original fixed window through package preview.
Authored evidence gaps cite the originating capture and remain unsaved drafts.
See [the contract](docs/INVESTIGATION_FOLLOWUPS.md) and storyboard 42.

Passed: full root Go suite, exploration/HTTP race tests, root vet, 32 frontend
checks, Python compilation and Hegel state/swarm (100 requested sequences, 25
steps). Hegel exposed shared predicate trees between returned actions; each
query now owns an independent deep copy, with a unit regression and passing
property rerun. Stored capture immutability is checked separately.

The focused real-stack journey passed in 1.4 minutes: reproduced delayed sample
work, compared duration/status over the same matching population, followed exact
trace/log IDs, explicitly discovered instruments in the original window and
saved a finding with an origin-cited instrumentation question. It does not infer
resource causality. Its quint-go ledger passed and rejected seven altered ledgers;
the new Bombadil source-free preview campaign passed. The initial attempt failed
only a textarea assertion using text content instead of value; corrected assertions
passed. Initial artifacts remain in `artifacts/followups-initial-failure/`.

Evidence: `artifacts/followups-{go,race,vet,node,hegel}.log`,
`artifacts/followups-journey-final.log`, and
`artifacts/followups/7cc4cfd1/`. Final image `77be6b447ade`: all six affected browser journeys passed in 5.5
minutes, with no skips: coherence 40, context notes 36, follow-ups 42,
investigation workspace 41, measurements 38 and metric discovery 39. All five
enabled Bombadil campaigns and associated live Quint/negative controls passed.
The expanded mobile preview passed containment and source-free checks. Affected
screenshots were regenerated. Final ledger: `artifacts/followups/aec68658/`;
regression log: `artifacts/followups-regression.log`. The wrapper restored owned
traffic and preserved existing telemetry volumes and five stale pending entries.

## Question-first investigation workspace (2026-10-18)

Investigations now retain a question, current finding, open questions and ordered
cross-signal evidence. An active title/question context follows navigation.
Explore's capture dialog can preserve only the capture, append to the current
investigation draft, or publish a child of an explicitly selected saved revision.
Unrelated dirty drafts remain intact; capture and attachment publication are
separate, and attachment failure leaves the capture available. Cards show stored
receipt scope and native definition identity through a receipt-only API, without
returning record banks or querying telemetry.

Passed: full root Go suite; affected workbench/HTTP race suites; root vet; all
32 frontend checks; expanded Hegel context state/swarm (60 requested sequences,
15 steps) covering complete brief mutation, fork and preserved history. Explicit
storage tests preserve the pre-brief version-1 hash/JSON shape, normalize omitted
question-only card lists and reject oversized summaries. The receipt HTTP test
runs with its telemetry source closed and verifies that no source read occurs.

Initial storyboard 41 passed (1.2 minutes) with real checkout/report spans, report
logs and an independent ordinary OTel counter. Its dedicated quint-go ledger
passed and rejected six mutated ledgers. Final deployed image `d13fcbdbce4a`:
all four browser journeys passed, with no skips, including captures 32, context/
notes 36, investigation workspace 41 and measurements 38. All four enabled
Bombadil campaigns and the associated live Quint/negative-control checks passed.
The final new journey additionally creates a saved branch while retaining an
unrelated dirty question draft, with no telemetry read. Fresh metric definition
reuse adds a capture/child revision while the prior capture and handoff reopen
unchanged. Desktop/mobile containment and screenshots cover the connected flow.

Artifacts: `artifacts/investigation-{core-go,core-node,race,vet,hegel}.log`,
`artifacts/investigation-workspace-journey.log`,
`artifacts/investigation-final-journeys.log`, and
`artifacts/investigation-workspace/3d3506c8/`. The diagrams, contract, planning docs,
README and affected storyboards were updated. No telemetry volumes were reset;
the five pre-existing stale pending entries were retained and traffic restored.

## Product essence and connected investigation flow (2026-10-18)

[Product brief](docs/PRODUCT_ESSENCE.md): definitions describe meaning, captures
preserve fixed evidence, and investigations explain it. The README now leads with
this loop. Definitions exposes the same measurement revisions as Explore, with
local search; saved record packages also have local search. Service inventory
carries explicit service/environment into metric discovery without automatically
reading telemetry. Measurement stages, canonical YAML, named time scopes and
consistent cards make authoring/execution boundaries visible. Discovery supports
explicit cancellation and suppresses publication after navigation or mode changes.

Passed: 32 frontend checks and all five affected browser journeys across the
initial run and focused rerun: coherence 40, inventory 31, measurements 38, metric
discovery 39 and attribute/facet polish 29. The initial run had three passes and
two harness failures: an inventory selector assumed one button per row, and
Bombadil refused an existing output path. The selector now uses the intended
Explore action's accessible name; discovery campaigns use unique output paths.
The corrected inventory/discovery rerun passed both cases (1.7 minutes).
Original failure artifacts are retained under `artifacts/coherence-initial-failures`.

The new journey passed its delayed **actual source response** abandonment check,
source-free definition preview, service handoff, canonical authoring and mobile
containment. Measurement, inventory and discovery Quint ledgers/negative controls
and all three enabled Bombadil campaigns passed. Discovery exploration includes
abandon/retry and scope-change invariants. Twenty storyboard images were generated
or refreshed, including three new frames for journey 40; desktop/mobile frames
were inspected. No measurement kernel or source query semantics changed.

Evidence: `artifacts/coherence-node.log`, `artifacts/coherence-journeys.log`,
`artifacts/coherence-final-journeys.log`, and `artifacts/metric-discovery/`.
The query/packages image is `11a6f58b5951`. Traffic was restored and the five
pre-existing stale pending entries and telemetry volumes were retained.

## Ordinary OTel metric discovery (2026-10-18)

Bounded service/environment discovery now proposes editable typed packages over
ordinary gauge/sum/explicit-histogram observations. Optional scope version/schema
remain exact identity values, including empty values. No invented identity or
implicit fleet aggregation. Stream summaries expose sorted dimensions.

Passed: root Go suite; race checks for measurements/metrics/HTTP; root vet;
32 frontend checks; Hegel optional-identity properties and existing measurement
state/swarm checks. The independent Python OTel SDK producer has no goo manifest
and omits scope version/schema. The initial live discovery/package/save/run/
capture/reopen storyboard passed (21.1 s). The expanded run passed both affected
journeys (3.1 minutes): measurement storyboard 38 and discovery storyboard 39,
including notes, package saving, capture reopening and native dimension labels.
Bombadil recorded 440 measurement and 248 discovery observations with zero
violations. The shared measurement Quint ledger passed with seven rejected
mutations. Nine affected storyboard images were refreshed.
The final discovery-only rerun passed (22.5 s), including its dedicated Quint
ledger with three rejected mutations: wrong service, missing proposed definition
and tampered evidence basis. Traffic was restored and stale pending data retained.

Artifacts: `artifacts/metric-discovery-{go,race,vet,node,hegel}.log`,
`artifacts/metric-discovery-final-journeys.log`,
`artifacts/metric-discovery-model-journey.log`, and
`artifacts/metric-discovery/ledger.json`.

## Typed measurement packages and standardized examples (2026-10-18)

[Contract](docs/MEASUREMENT_PACKAGES.md): twelve YAML packages bind actual queue
and application telemetry to Go gauges, reset-aware counters, native histograms
and scoped lifecycle preparation. Five data-only ordered workflows preserve a
fixed window and preview without source execution. General captures, immutable
package revisions and ordered notes are shared across all families. Both Go
examples now emit real trace-correlated OTLP logs; Python lifecycle records carry
exact business nanoseconds and stable per-emission event IDs.

Passed native checks: root Go suite, root race suite/vet, instrumentation module,
19 Python tests, 32 frontend checks, portable compiler/lifecycle/package CLI parity
storyboards and Compose configuration. The full verification module passed
uncached (240.18 s), including Hegel state/swarm and existing Quint MBT/trace
checks. Live-ledger tests in that module intentionally skip without supplied
artifacts; the browser journey supplies its actual measurement ledger.
Additional race coverage checks atomic custody publication by 16 concurrent
readers. Binding tests reject a consistently recomputed receipt that falsely
claims complete coverage, not merely an inconsistent receipt.

Focused storyboard 38 passed (1.4 minutes): ordinary reservation scenarios,
Go checkout/report work, independent ClickHouse gauge counts and actual Go logs,
all four measurement families, mature good/retry/missing/rejected/late outcomes,
scoped subject drilldown, fixed question capture/refresh, source-free reopening,
exact notes reference, immutable package save and fresh reuse. Both record and
measurement workflow steps preview without telemetry reads. Quint accepted the
real ledger and rejected seven altered ledgers. Bombadil recorded 462 observations
through initial/preview/invalid/complete stages without invented evidence.

The journey caught an incorrect checkout UI assertion and a real transport bug:
ClickHouse's escaped JSON spellings changed a raw-input evidence hash after a
browser round trip. Source inputs are now canonicalized before hashing, with a
transport regression test. A source metadata mismatch also exposed the Python
meter's missing version/schema; those are now emitted and bound explicitly.
Stored source plans remain reopenable as evidence after compiler changes;
refresh requires an identical current plan.

Broad browser regression: 37 passed, one failed (19.1 minutes), with all
opt-in generated/fault/Bombadil campaigns enabled. The failure was the fault
campaign's six-second actor budget expiring before any browser observation, not
an application or UI invariant failure. The harness now permits browser startup
headroom while preserving its five-second application-response property, and
reports empty traces explicitly instead of a JSON parse error.

Final targeted run on the deployed final image: **2 passed, no skips (4.0
minutes)**. The corrected fault campaign completed all three scenarios and five
inspections/revisits with its model checks (2.5 minutes). Measurement storyboard
38 passed again (1.4 minutes), including valid empty-selector capture
normalization, rejection of unsupported lifecycle scope filters, and a saved
same-name custom package that cannot shadow a builtin workflow's environment.
Thus all 38 browser journeys were verified across the broad run and corrected
focused rerun. All 132 storyboard images (131 PNGs and one JPEG) were refreshed; seven new frames and
architecture flow/sequence diagrams document this slice.

Artifacts: `artifacts/standardization-regression.log`,
`artifacts/standardization-final-journeys.log`,
`artifacts/standardization-race.log`, `artifacts/standardization-verification.log`,
and `artifacts/measurements/360b30f8/`. The first failed fault trace is preserved
under `artifacts/standardization-regression-failure/`; the successful fault run is
`artifacts/bombadil/fault-campaigns/40131596-b9ce-4cb2-9834-9270c7047939/`.
Final scope/normalization/default-resolution checks also have native regressions.
The final measurement Bombadil run recorded 565 observations through all four
required stages; its eight model reports contain one pass and seven intended
conformance rejections.

The broad run used query/packages image
`sha256:7494aa650a6be69c8d06e7d58a1fc88e93149d88b057600bf0beb2da0eea0010`.
The final targeted image, including selector normalization, strict lifecycle scope
validation and builtin workflow resolution, is
`sha256:4df78f6a14a53320181c8f0934f32b73fe1648c22938ab77f215dfb2d6f49b48`.
Runtime frontend hashes match the repository. Telemetry volumes and all five
pre-existing stale Redis pending entries were preserved. The local custody clock
is not warehouse ingestion or a completeness guarantee; missing completion stays
unknown. Databricks YAML was structurally checked, not warehouse-executed.

## Scoped recipes and shared sample investigations (2026-10-18)

[Contract](docs/INVESTIGATION_RECIPES.md): three Go recipe transforms preserve the
applied population and fixed bounds, with source-free preview and explicit
execution. Six ordinary YAML packages move sample investigations onto shared
Explore/package/capture/notes workflows. The specialized lifecycle/resource lab
starts only when opened. Opening a derived recipe question now clears the previous package banner.

Passed native checks: root Go race suite and vet, 19 Python tests and 27 frontend
checks. Native cases cover OR/filter preservation, immutable inputs, parameterized
selectors, wrong-signal rejection, strict HTTP bodies and all six example packages.
Hegel state/swarm passed 100 sequences of 25 recipe/signal/scope/predicate steps
(0.19 s reviewed).

Initial browser harness failures were corrected for the actual save message and
retained disclosure state on same-document navigation. A reserved Quint action
name was renamed. The conformance check then correctly rejected a capture taken
after package reuse changed the time bounds; the journey now verifies the fresh
reuse separately and captures the original fixed question before deriving
a reusable definition from that stored capture. This distinction is part of the documented contract, not a change to
package semantics.

Final browser evidence: the broad 37-journey run completed with 36 passes and
one legacy-sidebar harness timeout (14.8 minutes); that missing disclosure action
was corrected. A final focused run passed both the corrected definition/invalid-
query journey and the strengthened recipe journey (2 passes, 1.2 minutes). The
latter uses ordinary UI filters, captures/cites the fixed question first, then
creates and freshly runs a package from its stored capture. No skips occurred in
the broad run; generated/fault campaigns and cross-app Quint were enabled.

The final recipe ledger passed Quint with six rejecting controls (13.04 s).
Bombadil recorded 467 observations through all four required stages and all three
recipe choices. Wrong-signal rejection, delayed-preview isolation, source-free
notes/preview, literal query preservation and mobile containment were checked.
All 125 storyboard images were refreshed. Five new frames and architecture flow
24 document the slice; new desktop/mobile frames were visually reviewed.
See [environment evidence](ENVIRONMENT_VALIDATION.md) for image and artifact paths.

## Deployment context and ordered notes (2026-10-18)

[Contract](docs/WORKBENCH_CONTEXT.md): authored deployment receipts preserve event
and server custody times; explicit fresh local context accompanies fixed-time
Explore charts without modifying telemetry receipts. Ordered plain-text narrative
cards cite exact general-capture or deployment IDs. Saving creates immutable
parent/root revision branches, preserving older evidence and interpretations.
Dirty/unapplied drafts are retained; no mutable latest head or shared-author claim
is introduced.

Passed: root Go race suite/vet, 19 Python tests and 26 frontend checks. Native
checks cover half-open environment scope, separate custody time, unsafe URLs,
integrity, marker/card/catalog limits, reference validation, pagination and
concurrent branch preservation. Hegel state/swarm passed 60 sequences of 15 steps
(0.38 s reviewed), including Unicode note content, save/fork/order/history,
deployment arrivals and scope/window changes.

The initial browser flow completed its UI assertions but failed verification due
to a test type name colliding with an existing discovery model. Renaming it fixed
the verification build; the captured ledger replay then passed. Journey 36 next
passed with observed evidence/narrative transitions and a seven-stage Bombadil draft prefix
(383 records), including real report-app work, independent ClickHouse trace
discovery, marker boundary/environment exclusions, literal-text safety, ordered
reference editing, immutable revisions/branches, unapplied-draft retention,
source-free capture opening and mobile containment. Quint originally used five
rejecting controls; the strengthened adapter also checks the report's scoped
ready event and rejects a sixth incorrect-subject control (19.97 s replay).
Five new storyboard frames and architecture flow 23 document this slice.

All 36 journeys passed in one broad run (12.6 minutes), with both generated
campaigns and cross-app Quint enabled and no skips. All 120 storyboard images
(119 PNG and one JPEG) were refreshed. Final keyboard/capture-navigation guards
passed three focused journeys (2.0 minutes), including keyboard reorder/remove
focus, cited-evidence focus and delayed-revision isolation. Screenshot inspection
then caught an obsolete loading message after a canceled read; a final status
fix passed journey 36 again on the reviewed image (1.2 minutes), with six rejecting
Quint controls (17.15 s) and 310 Bombadil records through all seven stages.
The final narrow-screen frame was visually inspected. Evidence and image hashes
are in [the environment record](ENVIRONMENT_VALIDATION.md).

## Error and latency comparisons (2026-10-18)

[Contract](docs/ERROR_COMPARISONS.md): exhaustive captured span-status categories,
full-population error-percentage ranking, independently computed mean/p95/p99
changes across fixed windows, retained-baseline group drilldown and explicit
server-reread saved comparison pairs with links to both record captures.
Unset and unrecognized status remain visible; no-data/disagreement suppress
normalized differences. Existing count-only captures/SQL remain compatible.

Passed: root Go race suite/vet; focused capture-refresh race checks; 19 Python
tests; 24 frontend checks; Hegel comparison state/swarm (100 sequences of 25
steps, 0.54 s initial / 0.67 s model regression), plus existing cohort state/swarm
and Quint cohort regression (7.55 s). Native checks cover invalid/missing status
counts, category drift with identical total counts, missing groups, package YAML,
interval validation, source-free saved pair reopening, refresh child lineage and
retaining the predecessor on source outage.

Journey 35 passed initially (54.8 s) and on the polished image (47.0 s). It checks
80 previous/124 current real OTLP spans with only 100 current displayed rows,
2.50% → 16.13% recorded errors (+13.63 pp), +90 ms p95, missing/empty routes,
saved comparison reopening, changed-route-to-trace/log navigation and mobile
containment. The final new-feature campaign passed Quint with four rejecting
controls (7.20 s) and 342 Bombadil records through all six required stages.
All 11 affected regression journeys passed (3.9 minutes). A final trace-action
guard was checked by both discovery and the strengthened comparison journey
(2 passes, 1.3 minutes), including live adapter status-class groups and saved
trace-ID pair behavior. The final comparison campaign recorded 501 Bombadil
records through six stages. Five new frames and affected existing storyboards
were regenerated; see [environment evidence](ENVIRONMENT_VALIDATION.md).

## Grouped duration measures and polish (2026-10-18)

[Contract](docs/GROUPED_MEASURES.md): complete per-group duration aggregates,
normalized semantic results, exact rational ranking, fixed local-search scales,
retained capture pages and reusable package previews. Count-cohort comparison
projects out duration metadata and remains a record-proportion comparison.

Passed: root Go race suite/vet, 19 Python tests, 23 frontend checks and Hegel group
count/sum conservation over 100 generated populations (0.254 s), including uint64
extremes. Native source tests reject missing summaries/duplicate dimensions and
preserve same-count duration disagreement. A live fixture verifies 124 spans versus
100 displayed rows, missing/empty routes, zero duration, p95 ranking, one/two-field
groups, count-cohort projection, package preview and mobile containment.

The initial extended campaign passed five of six journeys, including journey 34
(52.1 s), Quint with three rejecting controls (7.32 s) and 454 Bombadil records
through all nine saved-ranking prefix stages. The discovery regression exposed a
removed accessible trace-button label, now restored. Browser checks also found a
blur-linked control shift, fixed by input-time scope invalidation. Typed model
comparisons preserve exact values without treating JSON indentation as evidence.
The final reviewed-image run passed all 10 selected regression journeys (3.5 minutes), including the original discovery assertion. The final group campaign passed three rejecting Quint controls (6.46 s) and 383 Bombadil records through nine stages. Four new and affected existing frames were refreshed; evidence paths are in [the environment record](ENVIRONMENT_VALIDATION.md).

## Core workbench basics (2026-10-18)

[Contract](docs/CORE_WORKBENCH.md): opt-in full-population span-duration summaries
(count/exact sum/min/max/upper p50/p95/p99), fixed count-over-time intervals with
per-interval durations, immutable capture retention, reusable versioned questions,
span/log service activity and reported parent links. General query admission and
explicit cancellation reach the Go context through the Python proxy. This remains
a local lab; event age is not ingestion lag, and query IDs are not authorization.

Passed: root Go race suite and vet, 19 Python tests, 22 frontend checks, Compose
configuration, Hegel duration state/swarm (100 sequences up to 30 steps). Native
HTTP tests exercise saturated admission, source-free local access, cancellation
propagation and delayed-request tombstones. Source tests reject malformed/missing
summary fields, invalid partitions/bounds and ambiguous dimensions; exact sums
cover values beyond uint64 and JavaScript integer precision.

The first deterministic journey 33 passed in 18.9 s. The strengthened focused
capture/core/inventory run passed all three journeys in 2.3 minutes, including
187 Bombadil records through all six core prefix stages and actual-ledger Quint
validation (13.26 s) with four rejecting controls. Five controls now check plausible-but-wrong p95/sum, ambiguous-link classification, cancellation and reused environment scope. All 33 journeys passed in one comprehensive run (11.9 minutes), with both generated campaigns enabled and no skips. All 106 storyboard images were refreshed. The expanded final campaign passed again with both Bombadil actors (216 saved-measure records plus 181 live-workflow records), all required stages and five strengthened rejecting controls. Artifact paths are in [the environment record](ENVIRONMENT_VALIDATION.md). Architecture flow 20 and four
new screenshot storyboards document this increment.

## General evidence captures and stable saved paging (2026-10-18)

[General captures](docs/GENERAL_CAPTURES.md) now persist manifest-free span/log
questions and service inventory in the existing investigations volume. Reopen and
paging read verified local evidence; refresh creates a successor with the same
normalized fixed question and compiled source plan. Explore captures retain up to
1,000 records, shown in pages of 100, while full supported aggregates retain their
own counts and source states. These bounded records do not claim source completeness.

Passed: root Go race suite and vet; focused race checks after final normalization;
19 Python tests and 21 Node checks. Native HTTP tests reopen with the source closed,
reject client-supplied evidence and cross-revision cursors, and preserve predecessors
on failed refresh. Hegel state/swarm passed 60 sequences of up to 15 steps against
final code (12.939 s), including arrivals, source resets, captures and branching
refreshes. Quint validates actual browser capture/page/refresh/reopen observations,
with four rejecting controls for old-count, page-basis, question and cursor drift.

Journey 32 covers 1,205 matching spans, a 1,000-record retained bank, stable pages
after late arrival, a 1,206-span successor, source-free browser reopening after
process restart, explicit live handoff, inventory, logs, narrow screens and stale
response isolation. Six new storyboard frames accompany diagram 19. Expanded trace
and log context remains a separate fresh read; saved baseline pairs, team access
control and collaborative editing remain future work.

The initial default suite had 29 passes, two opt-in skips and one outdated wording
assertion in journey 22. That assertion now matches the general “saved captures”
disclaimer; its semantic checks remain and focused journeys 19 and 22 passed.
Final capture/Bombadil and both opt-in campaigns passed (three tests, 3.5 minutes); all 32 distinct journeys passed across these runs. The capture actor recorded 367 trace records through all eight stages. All 102 storyboard images were refreshed; detailed evidence paths are in the cloud record.
An initial final-test selector matched no tests, and one retry lacked workspace Go
cache variables; both invocation issues were corrected before final verification.

The agent's [collaboration research](docs/COLLABORATIVE_INVESTIGATIONS_RESEARCH.md)
distinguishes verified Honeycomb Boards and Datadog multiplayer Notebooks from an
unverified Honeycomb whiteboard identification. It proposes ordered investigation
cards before authorized collaboration/presence; no multiplayer feature is claimed.

## Observed service inventory

Implemented: single-read full-population service/span-name counts, OTel Error
statuses and fixed span-duration buckets; local filtering preserves captured
accounting; drilldown opens a separate fixed-window Explore read and package
preview. [Contract](docs/SERVICE_INVENTORY.md), [diagram 18](docs/ARCHITECTURE_DIAGRAMS.md#18-discover-services-and-promote-an-investigation-question)
and [journey 31](docs/journeys/README.md#31-from-observed-services-to-a-reusable-question)
record semantics and executable frames. Unknown fleet membership/health and
span-versus-request grain remain explicit.

Passed: root Go race suite/vet, 19 Python tests, 20 Node checks, Compose config,
Hegel observe/reset state/swarm (100 sequences, up to 20 steps), and actual-browser
Quint capture/search/open/package conformance with three rejecting controls.
Source tests retain counts above JavaScript integer precision and reject invalid
partitions, errors exceeding counts, missing/null dimensions, duplicate groups,
cardinality overflow and unavailable-to-zero conversion. Query values remain
parameters; HTTP rejects unknown fields and multiple objects.

The live 116-span fixture checks independent ClickHouse counts, 113 matching
spans versus 100 displayed rows, package preview, local filter stability, no data
and pending-read invalidation. The dedicated read-only Bombadil campaign passed
609 records across all six prefix stages plus refresh/search/scope edits. Evidence:
`artifacts/inventory/inventory-a777c2549e/`. Narrow-screen screenshot review exposed
compressed table headings; stacked mobile cards and a width assertion corrected
that issue, and the updated journey passed. See the cloud validation record for
the subsequent broader browser run and final-image verification.

## Manifest-free ad hoc telemetry packages

Implemented: captured Explore question → named telemetry package YAML → shared Go
query plan → separate live preview → immutable local configuration revision → fresh
Explore reuse. Import/export, nested predicates/two-field breakdown preservation,
local refinements and optional downloadable Go annotation templates are included.
No target application manifest is loaded. See [contract](docs/TELEMETRY_PACKAGES.md),
[diagram 12](docs/ARCHITECTURE_DIAGRAMS.md#12-ad-hoc-telemetry-packages-and-optional-instrumentation)
and [journeys 24–25](docs/journeys/README.md#24-build-a-package-over-existing-python-telemetry).

Passed: root Go race suite/vet, 19 Python tests, 14 Node checks and Compose config.
New Go fixtures check strict bounded YAML, unknown fields/aliases/multiple documents,
unsupported grains/adapters/roles, time/page separation, stable YAML round trips,
concurrent idempotent publication, immutable old revisions, path validation and
corruption detection. HTTP tests pass generated Go code through the actual source
annotation scanner. No generated template is patched or deployed into an app.

Hegel state/swarm passes 100 sequences of up to 30 edit/save/window steps; Quint
observes draft/save/edit/reopen/apply and rejects mutated-original and nonreusable
negative controls. Focused verification passed (5.1 seconds); the complete nested
verification suite passed (95.5 seconds), with ignored log
`artifacts/verification/telemetry-packages-full-suite.log`.

The first actual Python SDK authoring journey passed (9.4 seconds). After preview
and handoff presentation improvements, journeys 24/25 passed together (14.0 seconds):
ordinary SDK spans without goo attributes/manifests, collector ingestion, exact
preview counts, save/export/import, service-restart persistence, fresh reuse,
local refinements, preserved nested log/two-field selectors, no-data preview and
stale-validation isolation. The Node-style log workload is a labelled OTLP fixture,
not a claim that a Node SDK service ran. The generated Go template is scanner-tested,
not source-equivalence, weaving, deployment or coverage proof.

VFS build history exhausted workspace bytes during recreation; unused dangling
images/build cache were reclaimed, retaining running containers, all telemetry
volumes and existing stale pending entries. Recreation then succeeded. No daemon,
proxy, credentials or storage-driver settings were changed.

This package kind supports captured-record counts and existing views, not arbitrary
measures, inferred SLOs/business semantics, automatic source patches, source authenticity,
production collaboration or Databricks lowering. Saved definitions do not archive
preview evidence. Mermaid source/fences and local links are checked without claiming
rendered Mermaid validation.

All 25 real browser journeys passed together (4.1 minutes), including the new
Python SDK and imported-log paths, native metric/fleet regressions, saved receipts
and baseline comparison. All 74 documentation frames were regenerated. Preview
tables, readable handoff code and mobile authoring were visually inspected. Final
editor-focus and metadata/navigation generation guards are covered by a focused
rerun of the two affected authoring journeys, which passed together (16.7 seconds).
Their six new frames were regenerated again after those refinements. All 12 services
are running, traffic is restored, and the five pre-existing stale pending stream
entries and all data volumes remain intact.

## Cohesive general-purpose workspace and cohort distributions

Implemented locally: Explore/Operations/Saved investigations/Definitions navigation,
secondary reservation Examples, shared responsive styles, result tabs and trace/log
inspector. Complete captured distributions support a browser-held fixed-window
baseline and Go rational percentage-point comparison. The shared waterfall accepts
caller-provided related actions; reservation navigation is supplied only by Examples.
See [workspace and scope contract](docs/PRODUCT_WORKSPACE.md),
[diagram 11](docs/ARCHITECTURE_DIAGRAMS.md#11-workspace-navigation-and-captured-distribution-comparison)
and [journey 23](docs/journeys/README.md#23-one-investigation-workspace-from-distribution-to-trace).

Passed: root Go race tests/vet, 19 Python tests, 14 Node checks, strict comparison
HTTP validation and Compose configuration. Fixtures cover exact uint64 counts,
missing/empty identities, category unions, scope drift, duplicate/partial groups,
empty populations and changed-read unknowns. Hegel cohort state/swarm runs 100
sequences of up to 30 steps. Quint validates observed pin/selection/empty/changed/
scope-change results with three rejecting negative controls. Full nested verification
suite passed (91.9 seconds); ignored log: `artifacts/verification/workspace-full-suite.log`.

The initial targeted seven browser journeys passed (37.6 seconds), covering actual
workspace navigation, old and new query paths, saved/live expansion, and a labelled
12-record collector fixture compared with a four-record filtered population. The
source fixture is documentation telemetry, not a production error attribution.
Mobile screenshot review prompted labelled comparison rows and visible navigation.
The full run then passed 22/23 journeys (4.8 minutes). Journey 07 exposed a
pre-existing readiness assumption: its five-minute API poll could see an older
failure while the UI used the new two-minute default. Its fixture now aligns
the window and waits for a newly captured failure identity. The first focused
rerun also encountered an actual transient source error during baseline acquisition;
readiness now checks HTTP success before accessing facts. Source failures are
not converted into zero-valued evidence. Expanded journey 23 passed again (6.7
seconds), including distribution-value refinement in the baseline's fixed window.
The corrected journey 07 then passed (4.0 seconds; focused command 6.3 seconds).
Together with the full run and focused journey 23, all 23 journeys have passing
results against this round's runtime. All 68 documentation frames were regenerated;
desktop comparison/definitions and mobile comparison/inspector frames were visually
reviewed. The inspector frame uses the actual viewport, avoiding full-page fixed-overlay
screenshot distortion. All 12 services are running and traffic is restored. Existing
five stale pending entries and all volumes were retained.
Cohort observations were strengthened to check expected positive/empty/unknown
results explicitly; focused Hegel/Quint verification passed again (4.7 seconds).
Baselines remain volatile browser state; caller-held comparison evidence is not
source-authenticated. Optional HyperDX, a Databricks instance, production deployment
and a complete accessibility audit were not validated. Mermaid source/fences and
relative documentation links were checked; no rendered-Mermaid validation is claimed.

## Bounded general exploration and expanded trace/log context

Implemented: shared Go query model/reference predicates and parameterized
ClickHouse compiler; UI/API/CLI source planning; discovered attribute keys,
string and UInt64-duration predicates with bounded AND/OR; independent captured
record counts/groups and bounded pages; missing-vs-empty grouping; detected
cross-read disagreement; cross-service trace pages and exact trace/span log scopes.
Operation/saved waterfalls explicitly expand into separate live evidence. See
[design](docs/EXPLORATION_DESIGN.md),
[diagram 10](docs/ARCHITECTURE_DIAGRAMS.md#10-general-exploration-and-independent-expanded-evidence)
and [journeys 20–22](docs/journeys/README.md#20-general-exploration-and-population-safe-paging).

Passed: root Go race tests/vet, 19 Python checks and 12 Node checks. Core/HTTP
fixtures cover type/operator/budget validation, literal parameter safety, exact
UInt64 comparisons, missing values, page-independent count/group SQL, fixed
trace scope, source/group overflow failures, detected read disagreement, CLI
planning and invalid requests rejected before source execution. Hegel state/swarm
passes 100 sequences of up to 30 steps; Quint observed search/page/trace/log steps
pass with three negative population/scope/identity controls. All three Dockerfile first-stage COPY contexts compile the root CLI with
VCS stamping disabled for temporary contexts. Full nested
verification suite passed (89.4 seconds), log in ignored
`artifacts/verification/exploration-full-suite.log`.

Live targeted journeys 20/21 passed together (13.8 seconds); expanded 20–22
passed together (37.6 seconds), including trace paging with missing parents and
saved/live evidence separation. An initial journey read collapsed diagnostics
using rendered text and failed; it now reads the actual stored text content.
Direct read-only ClickHouse checks found that raw HTTP query parameters reinterpret
backslashes and reject literal tabs. The adapter now escapes ClickHouse parameter
text as well as URL encoding. Exact backslash/tab/newline/Unicode/DEL round trips
were verified, and journey 20 now includes literal transport characters in both
keys and values. This is a transport correction, not an SQL interpolation path.

All 22 live browser journeys passed together (4.2 minutes), including the literal
transport regression in the 123-record fixture. The 63 documentation frames were
regenerated. After a UI origin-metadata refinement, focused journeys 20–22 passed again
(22.7 seconds), including the assertion that log→trace navigation carries the log
response question/basis as its origin. Context diagnostics also retain source
state/warnings; actual read disagreement is covered by source fixtures. No Go
query kernels changed after the full browser pass. The six new documentation
frames were regenerated, and the exploration, distributed waterfall/log and
saved/live-context frames were visually inspected. All 12 services are running;
the owned traffic generator is restored and no queues/volumes were cleared.
Mermaid source and relative documentation links were checked without claiming a
local Mermaid render.

General query/cursor persistence, whole-trace completeness/accumulation, live tail,
value distributions, cohort comparison, production authorization and warehouse
lowering remain unimplemented. No production deployment, Databricks instance or
full networked Docker build is claimed. General context is not added to saved
operation receipts.

## Saved fixed-window investigations

Implemented: bounded server-captured immutable operation evidence; local named
volume; package revision/fixed-window successor refresh; reopen without telemetry
queries; stored operation/filter/trace/subject and ratio/member explanation state;
component error preservation and comparison of independent evidence bases.
See [contract](docs/SAVED_INVESTIGATIONS.md),
[architecture diagram 9](docs/ARCHITECTURE_DIAGRAMS.md#9-captured-investigations-and-explicit-refresh)
and [journey 19](docs/journeys/README.md#19-save-reopen-refresh-and-compare).
The capture is an explicit fresh read when saving, not a silent claim to freeze
the prior UI response. It archives normalized results, not all raw metric inputs
or renderer/evaluator binaries. No production collaboration is claimed.

Passed: root Go race tests/vet, 19 Python checks and 11 Node checks. HTTP fixtures
verify trusted source capture, optional metric failures, offline reopening,
package/question conflicts, invalid selections and original survival after a
failed refresh. Store tests cover process-independent reload, exact JSON integer
preservation, concurrent publication, historical retention and corruption rejection.
Hegel state/swarm passes 100 sequences of up to 25 steps; Quint validates five
observed save/reopen/refresh/reject/reopen transitions and rejects two controls
mutating original integrity or the fixed question. Complete nested verification
suite passes (75.3 seconds); log in ignored
`artifacts/verification/investigations-full-suite.log`. An initial model adapter
parse failure with a parameterless module was corrected by declaring its initial
state constant; the final observed checks pass.

Root and both sample Dockerfile first-stage COPY contexts compile the root CLI
with VCS stamping disabled for temporary sandbox contexts; sample manifests pass.
Full networked multistage image builds are still unverified.

Targeted live journey 19 passed twice (21.2 and 24.5 seconds), including actual
packages-service restart, unchanged receipt reopening, real collector ingestion
of a labelled late test span, and successor retention of the fixed question and
trace selection. The second pass also verifies ratio/member explanation state.
All 19 browser journeys passed in two full runs (3.9 minutes each). The final
run follows the heatmap selection-state correction and includes saved cell-bound
restoration, ratio/member explanations, real late arrival, restart persistence
and a delayed real saved response discarded after a newer package selection.
All 57 documentation frames were regenerated; the three new saved-investigation
frames were visually inspected. A final test-only refinement scopes saved ratio
and comparison explanations to checkout (rather than whichever operation appears
first). Focused journey 19 then passed (16.2 seconds), regenerated its three
frames and asserted both selected stream identities. No application code changed
after the final full 19-journey pass. All 12 default services are running, metric
export is enabled, and the owned traffic generator was restored. No queues or
volumes were cleared. Documentation source links resolve; Mermaid is documented
without a claim of local rendering. General observability gaps and sequencing are recorded in
[PRODUCT_GAPS.md](docs/PRODUCT_GAPS.md), with reviewed public research sources and
explicit limits.

## Live peer and bounded-skew comparisons

Implemented: two actual woven checkout instances with distinct resource identity
(default ports 8090/8092); optional explicit bounded-skew policy; actual per-member
window/count preservation; pairwise endpoint spreads; deterministic conservative
matching without reusing intervals in successful rows. Comparison rows have no
combined percentage; exact fleet totals remain strict. The lab-only export gate
can drop metrics while requests and the independent trace exporter remain enabled.
See [semantics](docs/FLEET_COMPARISONS.md),
[diagram 8](docs/ARCHITECTURE_DIAGRAMS.md#8-two-live-instances-and-bounded-comparisons)
and [journey 17](docs/journeys/README.md#17-two-live-instances-imbalance-loss-and-restart).

Passed: root Go race tests/vet, 19 Python checks, ten Node checks; Hegel comparison
state/swarm: 100 sequences of up to 30 steps; six observed Quint outcomes and three
negative controls rejecting invented common populations. Targeted live journey 17
passed through imbalance, successful requests during metric loss, excessive gap,
and controlled peer restart/new epoch. Four screenshots were generated. The first
attempt exposed a real startup-query/control race (a late 30-minute response
replaced the edited two-minute lookback); controls now preserve edits and stale query
responses are generation-guarded. Source limits were retained. Targeted journeys 17/18 passed together (41 seconds); journey 18 guards
that race using a delayed real response. The full nested verification module passed (60 seconds).
All 18 browser journeys passed in one full run (3.2 minutes), refreshing 54
documentation screenshots. New member-window and loss frames were visually
inspected; updated documentation file links resolve. All twelve services are
running and the peer metric-export gate is enabled. Mermaid source is documented
without claiming a local Mermaid render or production deployment validation.
A focused post-run storyboard check scoped same-timestamp rows to checkout
and asserted the displayed reset cause. Gap/restart checks also require timestamps
from the current injected fault, avoiding matches against older retained telemetry.
The final focused journey17 rerun passed (36.7 seconds), including fresh-fault
timestamp guards and operation-scoped screenshot/assertion selection. No
application code changed after the successful full18-journey run.

## Historical declared fleet membership milestone


Implemented: fixed-window YAML membership for 2–4 instances; bounded concurrent
member queries; separate observed/unobserved/unavailable source states; exact
aligned population totals only when every member supports an interval. Missing,
reset, skewed or incompatible members leave totals unknown. Valid selected-instance
rates survive peer source failure. The default lab intentionally declares an
undeployed peer, not a live second replica. See [semantics](docs/FLEET_COVERAGE.md),
[diagram 7](docs/ARCHITECTURE_DIAGRAMS.md#7-fleet-membership-and-coverage-before-totals)
and [journey 16](docs/journeys/README.md#16-declared-fleet-membership-and-missing-coverage).

Passed: root Go race tests/vet, 19 Python checks, nine Node checks; Hegel fleet
state/swarm: 100 sequences up to 30 steps; Quint six observed aligned/missing/
skew/reset/idle/recovered steps and four rejected invented totals. Race-tested HTTP
fixture verifies peer failure preserves selected-instance rates. Targeted live
journey 16 passed with three generated frames. Dynamic membership, automatic
alignment and live multi-replica aggregation remain unvalidated/unimplemented.
The complete nested verification module also passed (70.8 seconds).
All 16 live browser journeys passed in one full run (2.8 minutes), refreshing
49 documentation screenshots. Fleet coverage frames were visually inspected and
updated documentation file links resolve. The default lab retains eleven services;
no second replica or automatic alignment is claimed.

## Operation error ratios and guided contextual investigation

Implemented: strict YAML ratio policy references an exhaustive calls-rate contract;
exact aligned outcome deltas produce calls/errors/fraction and a display percentage.
Missing/skewed/unsupported intervals remain unknown, supported zero calls are no
data, and full resource identity prevents implicit instance/operation merging.
The runtime explicitly observes zero outcome counters on first operation use.
The UI shows interval charts and guides error-context selection into captured
spans and the existing scoped trace waterfall. It does not detect increases or
causes automatically. See [semantics](docs/OPERATION_ERROR_RATIOS.md),
[diagram 6](docs/ARCHITECTURE_DIAGRAMS.md#6-from-aligned-counters-to-an-error-investigation)
and [journey 15](docs/journeys/README.md#15-from-operation-error-ratios-to-traces).

Passed: root Go race tests/vet, 19 Python checks and eight Node checks. Go fixtures
cover exact sums exceeding uint64, unknown/idle/alignment, invalid members/policies,
bases and a mixed idle/traffic window. Hegel ratio state/swarm checks pass: 100 sequences of up to 30 steps.
Runtime outcome-baseline state/swarm checks also pass 100 sequences. Quint validates
six observed traffic/missing/idle/reset/gap/late-repair steps and rejects four
invented ratios. Targeted live journey 15 passed, generating four readable frames.
Its first attempt correctly found no contextual spans before span ingestion;
the journey now waits for both kinds of real evidence before demonstrating traces.
Counter source cap increased to 3,000 raw points per contract for initialized outcome
streams; 8 MiB and Float64 precision guards remain, histogram caps unchanged.
The full nested verification module also passed (47.8 seconds). All 15 browser
journeys passed in one full run (2.0 minutes), refreshing 46 documentation frames.
Updated documentation file links resolve; ratio and waterfall frames were visually
inspected. All eleven default services, including traffic, are running. Mermaid source/link checks
are performed without claiming a local Mermaid render or production validation.

## Native histogram intervals and synchronized documentation

Implemented: exact uint64 cumulative/delta histogram evaluation, explicit
unit/layout/temporality/resource contracts, unknown reset/decrease/gap/flag/
boundary intervals, late repair and explicit aligned merging. Live checkout
histograms now feed a per-stream bucket heatmap with bucket-bounded P95 and
context-only span navigation. Automatic fleet discovery is not implemented.
See [semantics](docs/NATIVE_HISTOGRAMS.md),
[updated diagram 5](docs/ARCHITECTURE_DIAGRAMS.md#5-native-histogram-intervals-and-bounded-investigation)
and [journey 14](docs/journeys/README.md#14-native-duration-histograms).

Passed: root Go race tests/vet; 19 Python checks; seven Node visualization checks;
CLI fixture count4 and explicit merged count8; Hegel histogram state/swarm
100 sequences and merge properties150 cases; Quint observed restart/gap/repair/
merge validation with reset and incompatible-merge negative controls. All 14 live browser journeys passed in one full run (2.0 minutes), refreshing
42 documentation frames, including three cropped histogram frames. Targeted
journey14 also passed separately. File links resolve; Mermaid source was inspected,
not locally rendered. Documentation, diagrams, storyboard and generated frames
are required alongside feature changes in AGENTS.md.

## Native counter semantics, live rates and architecture diagrams

Implemented: pure Go cumulative/delta monotonic counter evaluation with exact
uint64 deltas, explicit stream identity, reset/epoch/decrease handling, maximum
gaps, missing flags, clipped unknown boundaries, duplicate/conflict checks and
canonical late-input revisions. CLI evaluate-counters runs without ClickHouse.
YAML rates and counter-rates views bind supported operation calls/error outcomes.
The checkout app now exports actual OTel native metrics every two seconds through
the collector; a bounded sum-table adapter validates resource/schema/unit/flags/
temporality and rejects Float64 precision risk. Charts show per-stream supported
intervals and separate counter bases. Linked spans are contextual start-time
matches, not exact contributors or exemplars.

Passed: root Go race tests/vet, exact CLI fixture deltas 3/4, 19 Python checks and
six Node checks. New counter/source fixtures cover uint64 precision, zeros,
resets/decreases, missing points, delta overlaps, conflicting copies/temporality,
identity, edges, late repair and rejected exporter precision loss. Hegel passes
100 state/swarm sequences (up to 30 steps). Quint validates five observed evaluator
steps through restart/gap/late repair, and rejects a rate invented across restart.
Artifacts are in artifacts/verification/counter-*.

The first full browser run had 12 passing journeys and one failure: an old broad
summary selector became ambiguous after adding the counter panel. That selector
was scoped to its intended operation explanation label. Affected journeys 08/11
and native-rate journey 13 then all passed in a targeted rerun (17 seconds).
Thus all thirteen journeys are verified across the two runs; this is not a claim
of a subsequent full-suite rerun. Two new counter frames bring the total to 39.
Eleven services run, traffic is restored and four stale pending entries remain.
Environment restoration initially left ClickHouse/Redis stopped; their existing
containers were started without resetting volumes.

The user-requested subagent prepared four Mermaid architecture/process diagrams
covering build/planning, illustrative production OTel/ClickStack topology, captured
evidence investigation and the separate Databricks compiler boundary. Links and
fences were checked; Mermaid syntax was inspected but not locally rendered.
These diagrams are documentation, not production deployment evidence.

Running images use offline-wheel builds plus host-built static binaries. Cumulative
metrics are verified live; delta semantics are fixture-tested. Native histograms,
fleet aggregation, error ratios, metric exemplars, native metric Databricks lowering,
immutable custody and production automation remain outside this slice. Updated
startup instructions are saved as a draft, not published. See
[native counters](docs/NATIVE_COUNTERS.md) and
[architecture diagrams](docs/ARCHITECTURE_DIAGRAMS.md).

## Declarative YAML operation packages and second application

Implemented: strict versioned YAML operation packages with source adapter/resource/
instrumentation bindings, exact source operation/event contracts, full subject
identity, enabled built-in views and explicit unverified evidence assumptions.
`goo plan-operations` and the investigation service share the same planner. Plans
carry content revisions that contribute to captured evidence bases; mutation is
rejected before fetching. Repeatable operation-package flags replace the former
source-directory flag and sample-specific Go binding. The UI selects packages,
honors enabled views, clears previous evidence and discards stale responses.

A separate woven Go report generator now uses its own service/instance, source
manifest, report.generate operation and report namespace. Its YAML omits the
heatmap and reuses the same query/UI code. Both packages run in default Compose;
there is no application-name branch in the operation renderer or source adapter.

Passed: all twelve live Playwright journeys (1.2 minutes), including report
creation → YAML-backed view → report subject and deliberate delayed delivery of
a real checkout response during a switch to reports. No telemetry response data
was mocked. Five new documentation frames bring the total to 37. Eleven services
run, demo traffic was restored and three stale pending entries retained.

Root Go race tests/vet, 19 Python checks and six Node checks pass. CLI tests plan
both checked-in packages. Source/plan tests cover stale roles/versions, weakened
identity, unknown views/fields, duplicate keys, aliases and multi-doc YAML, revision
determinism/nonmutation, resource provenance and changed-plan rejection. Hegel
passes 150 new planner cases for typed resource values, revision sensitivity and
nonmutation. Existing 100-case heatmap and 100-case subject properties plus the
five-action Quint navigation checks/negative controls remain passing.

The report helper actually validates and builds a static woven executable. Running
images use the established offline build with host-built Go binaries; the new
networked report Dockerfile was not executed here. Install/start additions are
saved in the cloud environment draft, not published. No root dependency was added.

This is a local-source, fixed-adapter/view YAML package slice. General package
resolution, source-free manifests, arbitrary expressions/CEL/KCL, IaC apply,
native metric reset/histogram policy and operation Databricks lowering remain
unimplemented. See [package specification and executable onboarding](docs/OPERATION_PACKAGES.md).

## Scoped subject event navigation

Implemented: optional runtime subject namespace/tenant context, declared-event
normalization from actual ClickHouse nested event arrays, and trace → subject
events → related event-bearing invocations across traces in the captured evidence.
The key retains environment/service/namespace/tenant/ID; missing or mismatched
identity stays unlinked with a diagnostic. The sample now generates fresh default
IDs and can explicitly reuse an ID in either demo tenant, fixing its former
constant-subject collision. No subject dimensions were added to metrics.

Passed: all ten real browser journeys (1.2 minutes). The new live journey submits
two invocations with one scoped subject, plus an identical ID under a second
tenant. It confirms four events across two related traces, excludes the other
tenant, retains the original basis and performs no additional operation request.
Three new frames bring the storyboard total to 32. All ten services run; demo
traffic is restored and three pre-existing stale pending entries were retained.

Go race tests/vet, 19 Python checks and six Node checks pass. The actual OTel SDK
scope test passes, as do 100 new Hegel scope-isolation cases, the existing 100
heatmap cases and 200 observation-copy cases. The five-action Quint kernel fixture
passes conformance/safety; both wrong-cell and cross-tenant membership mutations
are rejected. Observed steps and ITF are retained under artifacts/verification/
subject-navigation*. Existing actual woven SDK/HTTP preparation checks and Quint
negative control remain passing. Root runtime dependencies remain unchanged.

Returned event arrays/timestamps are validated against their parent spans. The
relation joins only event-bearing invocations, not failed calls inferred from
absence. Reported terminal outcomes do not establish retries, transaction commit,
durability or complete subject lifecycles. Query scope is still one explicit local
resource/library/span-start window with mutable late-data reruns. Updated startup
instructions are saved as a draft, not published. See [subject storyboard](docs/SUBJECT_NAVIGATION.md).

## Operation duration heatmaps and trace drilldown

Implemented: the Go source response retains deduplicated logical span evidence
with exact string start/duration nanoseconds and validated parent IDs. Operation,
calls, error and panic selections feed a raw-duration heatmap and span inspector.
Cells inspect exact contributors; opening a trace expands to the other operation
spans in that trace within the same resource/window/basis. The existing waterfall
renderer is reused with a separate details target. The Go sample now includes a
nested stock-check operation with bounded simulated success/error work.

Passed: nine real Playwright journeys (about one minute), including live failed
checkout → reserve error selection → duration cell → contributor → parent/child
waterfall. The new journey independently checks membership and contributor count,
changes the lookback control during exploration, and verifies a retained basis,
original bounds and no additional operation request. Three new frames bring the
storyboard total to 29. Existing flows remain passing; traffic was restored and
three stale pending entries preserved.

Go race tests/vet, 19 Python checks and five Node checks pass. Hegel passes 100
new generated cases exercising the actual browser kernel against a Go reference,
plus the existing 200-case operation-copy property. A four-action observed
navigation fixture passes quintgo/Quint conformance and safety; an incorrect cell
membership mutation is rejected. This models the actual selection kernel over a
bounded fixture, not browser DOM behavior or source completeness. Reports are in
artifacts/verification/operation-navigation*.txt. Source tests additionally check
logical evidence precision, parent validation and deduplicated inspector records.

The updated Go binaries and UI run in the established offline-built images.
Trace scope stays limited to observed operation spans in one selected resource,
library and span-start window. Sampling/completeness remain unknown, late-data
refreshes can change results, and no immutable custody or full distributed trace
fetch is claimed. See [the executable storyboard](docs/OPERATION_DRILLDOWN.md).

## Live woven Go operation integration

Implemented: separate Go checkout sample built with actual Orchestrion weaving,
standard OTLP/HTTP trace export through the collector into ClickHouse, explicit
resource/library binding, and a Go operation endpoint behind the existing query
proxy. The investigation panel lists declared operations, evidence status,
completed invocations, errors, panics and duration sample counts, with source,
window, manifest revision, evidence basis, typed SQL and duplicate explanation.
Unobserved declarations display unknown counts.

Passed: all eight real Playwright journeys (53.8 seconds), including actual sample
HTTP 200/503/500 paths and observed compatible operation rows. Three new captured
frames bring the storyboard total to 26. Ten default Compose services run; traffic
was restored and three pre-existing stale pending entries retained. Go race tests
and vet pass; 19 Python checks and three Node geometry tests pass. Hegel passes
200 generated cases checking invocation totals under duplicate copies and reversed
input order. Source-adapter tests cover typed values, exact bounds, resource/kind/
scope rejection, conflicting copies, deduplication, raw overflow and HTTP decoding.
The lint-before-build sample helper was executed successfully. Existing actual
woven SDK conformance and HTTP preparation/Quint replay plus negative control pass.

The running images use offline Python wheels and host-built static binaries.
The new networked sample Dockerfile was not executed. A read-only buildx cache
path was corrected with BUILDX_CONFIG=/workspace/.cache/buildx. A Compose name
conflict during replacement was recovered by recreating only query/packages and
the new sample; ClickHouse/Redis volumes were preserved. Updated reusable install
and start instructions were saved to the environment draft, not published.

Schema compatibility is declared because the exporter drops scope schema URLs.
This sample exports traces; native metric export/rate/reset/histogram semantics
remain separate. Coverage, sampling, completeness, immutable custody and durable
business subject semantics remain unknown. See [live operation storyboard](docs/LIVE_OPERATIONS.md).

## Observed operation conformance

Implemented source-manifest matching against normalized completed operation
spans, an invocation overview and `goo observe-operations`. Root race tests pass.
The actual woven SDK test also passes and checks three completed spans against
the overview (two reserve calls including one error, one panic). It exports
artifacts/annotations/operations.json for the executable CLI journey. Negative
checks reject cross-resource evidence, mismatched scope/version/schema,
undeclared roles, invalid outcomes/IDs and conflicting duplicates; identical
copies count once. Exact duration precision and unobserved declarations are tested.
Existing woven HTTP/Quint replay and negative control still pass.

This is SDK evidence, not collector/ClickHouse/browser integration. Native metric
rates and histogram merging, live operation discovery and business aggregation
remain unimplemented. See [the operation journey](docs/OPERATION_CONFORMANCE.md).

## Goo operation/event annotation slice

Implemented: bounded operation/event directives, deterministic source-contract
manifests, isolated OTel runtime hooks and test-only lint-before-weave CLI.
The normal runtime compiler module has no new OTel dependency. The default
Dockerfile now includes the annotation parser in its Go build stage; no running
service or UI was changed by this opt-in developer slice.

Passed: root race tests/vet; actual Orchestrion weaving of the example with real
OTel SDK export. Checks cover success/error/panic application behavior, three
completed-invocation/duration observations, zero leaked active invocations,
two subject events, three diagnosed event drops, and manifest scope/version/unit
agreement. The existing woven HTTP preparer/Quint test and dropped-fact negative
control remain passing. Telemetry is retained in artifacts/annotations/otel.json;
manifest in artifacts/annotation-manifest.json.

The native Hegel state/swarm check passes with 100 requested sequences of up to
30 steps, checking actual SDK active counts after start, successful/error/panic
completion and repeated finalization. The developer journey uses in-memory OTel
export, not live collector ingestion or a new browser screen. Full build manifests,
automatic inventory activation, operation aggregate lowering and production
weaving deployment remain next steps. See [the executable annotation storyboard](docs/GOO_ANNOTATIONS.md).

## Prior live HTTP source, shared view and opt-in weaving

Passed in the current workspace:

- Go race tests/vet and 19 Python checks, plus the shared package CLI/reference/
  embedded ClickHouse parity storyboard and three Node geometry checks.
- Seven real browser journeys (47.8 seconds), capturing 23 frames. The new
  journey creates actual successful and failing HTTP responses, waits for OTel
  ingestion, renders the Go view descriptor and inspects policy/source/revision.
- Source adapter unit checks for typed parameters, exact half-open bounds,
  valid OTel IDs, deduplication, conflicting copies and partial-scope rejection.
- Hegel: 600 requested cases across discovery, HTTP policy and source dedup.
- Real Orchestrion weaving of `Expansion.PrepareHTTP` into quintgo recording;
  its retained-count trace passes Quint and a deliberate dropped-fact trace is
  rejected. Positive and negative artifacts are kept separately.

The isolated weaving module pins x/tools 0.52.0: the initially selected 0.49.0
could not read Go 1.27 export data. The updated tooling passes; the root runtime
module is unaffected. The initial module download denial was resolved through
the official GitHub source with Go checksum verification. No network policy was
changed to complete the check.

The actual running images use the established offline-wheel build plus a
host-built static Go executable. The new networked multi-stage Dockerfile was
not executed here. Default Compose now includes the Go packages service; all
nine services run and traffic is restored after journeys. Reusable setup draft
instructions were saved, not published.

Schema compatibility is declared by library binding because this exporter drops
scope schema URLs. Sampling/completeness, immutable custody snapshots, generic
instrumentation auto-discovery, prepared-table materialization and Databricks
execution remain unverified/unimplemented. Woven model observations are tested
in the optional module, not enabled in the running service. See
[LIVE_HTTP_PACKAGE.md](docs/LIVE_HTTP_PACKAGE.md) and
[INSTRUMENTATION.md](docs/INSTRUMENTATION.md).

## Prior HTTP binding and expansion slice

Implemented: explicit HTTP server-span binding into the existing shared Go
measurement model, aggregate view/recipe descriptors, both backend plans and
normalized server-request preparation. Content revisions cover policy and emitted
artifacts. No current UI or raw-exporter adapter integration is claimed.

Passed: root Go race tests and vet; package CLI storyboard with two independently
specified route groups matching both reference evaluation and generated embedded
ClickHouse SQL; generated Databricks YAML structural checks. Unit checks reject
unsupported profiles, missing policy/source contracts, resource mismatches,
client spans, duplicate identities and mutated expansion content. Hegel adds
250 requested policy/status cases alongside the existing 200 discovery cases.

Databricks execution/deployment, raw-source SQL preparation, duration operators,
objective/alert configuration and browser rendering of these descriptors remain
unimplemented. Existing Quint lifecycle models were not extended or rerun for
this stateless binding/classification slice. See [the executable CLI storyboard](docs/PACKAGE_EXPANSION.md).

## Prior semantic packages and worker/queue adapter

Current additions on 2026-10-17:

- Ten versioned Go contract profiles and per-resource-scope capability proposals.
  CLI catalog/discover examples run; unknown versions/units and co-observed
  dimension requirements are tested. Generated UI catalog matches the Go source.
- Root Go race tests and vet pass. New Hegel property passes with 200 requested
  generated cases across profiles: absent/mis-versioned/mis-unit evidence cannot
  be borrowed from another scope; inventory ordering and duplication do not
  change proposals. Native verification stays in its separate module.
- Nineteen Python API/SQL/adapter tests pass, including exception-safe bounded
  worker counts, unknown broker lag, clock violations, per-instance gauge
  snapshots and raw-observation wait percentiles. Three Node geometry tests pass.
- Six real Chromium journeys pass (45.1 seconds), capturing 21 documentation
  frames. New coverage observes active work through actual Redis → OTel →
  ClickHouse ingestion, checks all queue/worker metrics and opens the package
  catalog. The final catalog-only update is also checked with the queue/library
  journey. Demo traffic is restored by the wrapper.

The new catalog is metadata recognition/proposal generation, not automatic live
inventory extraction, source/value-level binding, executable expansion of all
profiles, or Databricks deployment. The reference queue view is implemented;
most other profile outputs remain explicitly proposed. Versioned examples are
illustrative declarations, not live-scan evidence. No new scheduler or custody
model was introduced; existing Quint lifecycle models were not extended to claim
queue-adapter or package-expansion verification.

Queue observations distinguish undelivered lag, pending (including stale work),
local active/capacity and first-start waiting time. No stream reset was used;
the current wrapper reports three stale pending entries, up from two before the
consumer update. The sample still cannot reclaim crash/interrupted work.
See [semantic packages](docs/SEMANTIC_PACKAGES.md),
[queue contracts](docs/QUEUE_TELEMETRY.md), and [journeys](docs/journeys/README.md).

## Prior visualization and guided-investigation slice

Passed against the current local stack on 2026-10-17:

- Fifteen Python API/validation/embedded ClickHouse tests. New checks cover
  duration-band boundaries, zero and UInt64-maximum durations, exclusive time
  bounds, invalid selections, and 1101-span heatmap totals versus 200-row inspection.
- Three Node geometry checks cover exact nanosecond offsets, scoped parent
  identities, missing/cyclic parents, and zero-duration/empty spans.
- Five live Chromium journeys pass (41.7 seconds), capturing nineteen storyboard
  frames. Added paths exercise heatmap → exact raw spans → reservation waterfall
  → async producer link, and guided retry/USE/RED navigation with scope checks.
- Inspected actual waterfall and USE screenshots; Compose configuration and
  JavaScript/Python syntax checks pass. Demo traffic is restored after the run.

The first async navigation run correctly exposed an omitted producer span: its
sample instrumentation lacked reservation identity. New producer spans now carry
that identity and related attributes; the passing journey follows the exact
linked span. Older data is not backfilled and external links remain labelled.
The heatmap's minute conversion was also checked against both embedded and live
ClickHouse rather than relying on assumed timestamp function return types.

These visualizations remain in the Python reference lab, not a Go/Rust API
migration. USE utilization/queue measurements and a complete RED error-rate
measure are explicitly absent. Historical custody bases and Go lifecycle receipt
integration remain unimplemented in the UI. Existing Hegel/Quint semantics were
not changed or rerun for this presentation/raw-span slice.
See [visualization contracts](docs/VISUALIZATIONS.md),
[guided investigations](docs/INVESTIGATION_WORKFLOWS.md), and
[captured journeys](docs/journeys/README.md).

## Prior lifecycle, journeys and verification

Passed in the current cloud workspace:

- Go race tests and vet, plus all twelve existing Python tests.
- Thirteen independent lifecycle scenarios and 26 total/grouped aggregate SQL
  comparisons in embedded ClickHouse, including old/new snapshot coexistence.
- Three Hegel properties: generated deadline/dedup cases across epoch ranges,
  historical evidence identity under invisible appends, and state/swarm sequences
  with invariants after every enabled step (300/200/150 requested cases).
- Quint 0.32.0 bounded simulation: 100 traces, maximum 30 steps, seed 42.
- Eight Quint-generated MBT traces replayed through the actual Go preparer:
  160 steps with state comparisons. An eight-step observed Go trace conforms
  and passes `safety`; an impossible good count is correctly rejected.
- Three real Chromium/Playwright journeys pass after workload isolation,
  capturing twelve documentation frames. The sample's good, retry, late, missing
  and rejected scenarios traverse the real services and investigation UI.

Observed and addressed: restored ClickHouse/Redis processes needed startup;
recreated application containers needed local-service proxy bypasses; an initial
live journey under background load missed the deadline; the documentation runner
now pauses/drains/restores the owned generator without deleting stream data.
Pre-existing stale pending messages remain unreclaimed. Screenshot review also
found and fixed stale view controls after expression edits and an inappropriate
metric-point disclaimer on trace/log/latency results.

The normalized-event kernel is tested in Go; generated SQL **preparation** from
raw exporter data is not implemented. The Go semantics are not yet connected to
the live UI. Browser coverage does not resolve the Python accounting defects.
Databricks deployment/query/ingestion and optional HyperDX remain unverified.
See [journey frames](docs/journeys/README.md), [lifecycle contracts](docs/LIFECYCLE_PREPARATION.md),
[verification commands](verification/README.md), and [the bitemporal spike](docs/BITEMPORALITY.md).

## Portable Go aggregate pilot

The follow-on pilot is documented in [PORTABLE_COMPILER.md](docs/PORTABLE_COMPILER.md).
Passed in this workspace:

- `go test -race ./...` and `go vet ./...` using Go 1.27.2.
- `python scripts/check_portable.py`: ten total/group fixture comparisons across
  independent expected results, the Go evaluator, and generated embedded
  ClickHouse SQL; an additional declarative-count/escaped-literal case passes.
- Structural parsing of generated Databricks YAML and CLI output contracts.
- The runnable CLI example returns two eligible subjects and success 0.5.
- All original twelve Python unit/API/embedded ClickHouse tests still pass.

The docs URL was retried and remains blocked by proxy 403; the pinned official
Databricks creation reference was re-read from accessible GitHub content.
Databricks DDL/query execution, Unity Catalog deployment and physical ingestion
remain unrun. Normalized lifecycle preparation is now tested as described above;
real exporter adaptation and SQL preparation remain pending. The Go pilots do
not resolve the Python prototype's existing measurement defects.

## Original handoff record

The current lifecycle and journey validation is recorded separately in
[LIFECYCLE_PREPARATION.md](docs/LIFECYCLE_PREPARATION.md),
[the journey storyboard](docs/journeys/README.md), and
[verification/README.md](verification/README.md). These additions supersede the
original environment's unverified browser/default-stack status, while preserving
the distinction between the Python prototype and new Go semantics.

Created 2026-10-16.

Passed:

- Docker Compose 2.35.1 `config --quiet` against the complete manifest.
- Twelve Python unit/API tests, including all five query modes executed in embedded ClickHouse 24.8 through chdb 3.3.0.
- Application flow test: real HTTP and gRPC processes, fake Redis TCP/Streams, real consumer processing, and a local protobuf OTLP receiver. Verified accepted/retry/terminal log records, gRPC server spans, asynchronous span links, and exported metrics.
- Python compile checks and JavaScript syntax validation.
- OTel SDK/exporter construction with pinned Python dependencies.

Not verified in this workspace:

- Full Compose startup, collector configuration acceptance and ingestion into the pinned ClickHouse 25.3 container, image pulls, optional HyperDX startup, and `scripts/smoke.py` against that stack.
- Browser rendering. The Playwright browser download was unavailable; only JavaScript syntax was checked.

Docker 28.0.4 and Compose 2.35.1 binaries were installed for validation. The hosted workspace denies mount, Unix-socket, and user-namespace operations needed by the Docker daemon; rootless setup is also unavailable. Installing binaries cannot supply those host permissions. Run the provided full-stack smoke test on a Docker-capable machine before treating the container integration as verified.

The first application-flow run exposed inherited trace sampling that discarded demo spans. The SDK now explicitly retains all sample traces. A subsequent integration run verified the emitted spans and links.

## Codex handoff packaging

The handoff adds repository guidance, a continuation task, architecture, two proposed RFCs, an implementation plan, known issues, source references, and a portable snapshot of the original design Page. No application behavior changed during this packaging step.

The twelve unit/API/embedded ClickHouse tests and Compose configuration validation were rerun successfully. Local Markdown links and fenced code blocks were checked. The archive excludes caches, environments, downloaded runtime binaries, and database volumes, and includes SHA256SUMS for its source and documentation files. Full container startup and browser rendering remain unverified as noted above.

## Log and trace discovery

Implemented readable literal log search, independently scoped surrounding logs,
matching reported trace-ID groups, fixed full-population span-duration buckets and
available-field suggestions feeding reusable packages. See the
[discovery contract](docs/EXPLORATION_DESIGN.md#log-discovery-time-context-and-span-duration-buckets),
[journey 30](docs/journeys/README.md#30-from-log-search-to-reusable-span-distributions)
and the current detailed results in [ENVIRONMENT_VALIDATION.md](ENVIRONMENT_VALIDATION.md).
This is a captured-record investigation workflow; it does not establish complete
traces, root latency, live tail or production readiness.

Final inventory milestone: default browser run passed 29 journeys (two opt-in
skips); explicit fault, generated-success and final-image inventory campaigns
passed together (3.3 minutes). All 31 distinct journeys passed across these runs.
The final inventory actor recorded 685 steps with all required stages and rejecting
Quint controls. All 95 PNG documentation frames and the generated-campaign JPEG frame were regenerated; see
[the cloud record](ENVIRONMENT_VALIDATION.md#2026-10-18--observed-service-inventory).

## 2026-10-20 — exact-stream live counter comparison

The native live pilot uses the existing typed evaluator and exact OTel identity.
Focused source tests and race checks cover nanosecond bounds, newest-window gaps,
cumulative resets, changed instances, stale endpoints and source-free freeze.
`go test ./...`, `go vet ./...`, the three affected packages under `-race`, 68
frontend tests and `TestHegelCounterWindowRateAndCoverageState` pass. See the
[counter contract](docs/LIVE_COMPARISON.md#exact-stream-counter-rates) and
[journey 72](docs/journeys/README.md#72-follow-a-native-counter-stream-without-hiding-export-gaps).
Five real-source browser checks and the final-image native rerun (including
Bombadil and observed Quint replay with seven rejecting controls) passed.
All 12 services are running with matching served-source and binary hashes.
Deployed image evidence is recorded in ENVIRONMENT_VALIDATION.md.
