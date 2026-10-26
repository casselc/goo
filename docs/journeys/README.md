# User journeys and storyboards

This index preserves executable storyboards and their historical acceptance
records. Current product entry is Explore's Activity stream, with explicit staged
analysis and preservation; use the [roadmap](../ROADMAP_2026_10_24.md) for current
status and [README](../../README.md) for startup. A dated journey is not a request
to reimplement its feature. Synthetic component conformance fixtures remain
separate from real OTLP/application stories, as labeled below. Screenshots are
generated evidence; no frames were fabricated for this documentation repair.

## Recent source-backed entries

| Journey | Executable source | Contract / scope |
| --- | --- | --- |
| 44, M1 | [Numeric measures](../../journeys/numeric-measures.spec.js) | [Qualified numeric binding guidance](../NUMERIC_ATTRIBUTE_PLANNING.md); explicit type/unit and existing source bounds. |
| 87, M3 | [Trace finding](../../journeys/trace-finding.spec.js) | [Retained-page anchor continuity](../TRACE_FINDINGS.md); independent capture reads, no complete-capture membership claim. |
| 164 | [Operation status](../../journeys/operation-status.spec.js) | [Independent declared outcome and reported status](../OPERATION_STATUS.md). |
| 165 | [OTel field hints](../../journeys/otel-field-hints.spec.js) | [Exact qualified field discovery](../OTEL_FIELD_HINTS.md#executable-storyboard-165). |
| 166 | [Observed grouping](../../journeys/observed-grouping.spec.js) | [Source-free grouping before Run](../OBSERVED_GROUPING.md#executable-storyboard-166). |
| 167 | [Composed grouping](../../journeys/composed-grouping.spec.js) | [Two-field composition](../COMPOSED_GROUPING.md#executable-storyboard-167). |
| 168 | [Grouping control](../../journeys/grouping-control.spec.js) | [Passive picker and real composition](../GROUPING_CONTROL.md); the standalone component fixture is synthetic. |
| 169, M4 | [Explicit draft recovery](../../journeys/investigation-draft.spec.js) | [Implemented bounded same-tab recovery](../INVESTIGATION_DRAFT_RECOVERY.md), independently accepted in `6aba406`; Restore reads nothing and never retries publication. |
| 171, TF3 | [Typed numeric populations](../../journeys/typed-numeric-filters.spec.js) | [Real 512-record OTLP filtering, chart/capture/YAML/reopen](../TYPED_NUMERIC_FILTERS.md#real-data-storyboard-171); independent row/count oracle, Quint and Bombadil. |
| 172–173, TF2 | [Editor and authored-question conformance](../../journeys/typed-numeric-editor.spec.js) | Synthetic intent in the running UI with real Go document save; no telemetry execution. |
| 170, M4 | [Save/storage races](../../journeys/investigation-draft-races.spec.js) | Real late save acknowledgment, acknowledged-save/local-cleanup failure, and denied storage without losing authored work. |

The individual feature contracts link their actual existing frames. Current-round
executed checks and review decisions belong in [validation](../../VALIDATION.md)
and the [review-gated backlog](../BACKLOG_2026_10_24.md).

[Journey 111](../../journeys/fixed-library.spec.js) inspects same-title reusable
revisions without execution, holds real compiler responses across selection/close/
workspace/fixed-owner changes, refuses a mismatched compiler revision, then explicitly runs one fresh question.
[Mobile reusable revision detail](screenshots/111a-fixed-reusable-revisions.png).
Journey 33 additionally exercises real population measures/reuse and existing
Bombadil/Quint ledgers through the shared rows.

[Journey 110](../../journeys/definition-library.spec.js) explores compact record/native
revision libraries with unrelated same-title definitions. It preserves selected
annotation/editor drafts, rejects a delayed older real catalog response, and checks
source-free selection and mobile containment. [Selected revisions](screenshots/110a-selected-definition-revisions.png) · [Mobile rows](screenshots/110b-mobile-definition-rows.png).
See [the library contract](../DEFINITION_LIBRARY.md) for cache/search bounds and remaining B1 work.

[Journey 109](../../journeys/lazy-authoring.spec.js) checks that fresh Activity has
no fixed-editor DOM. Opening authoring creates one without a source query, returning
to paused Activity detaches it, and reopening restores the same draft/DOM identity.
Record deep links stage an editor without execution. [Explicit fixed authoring](screenshots/109a-explicit-fixed-authoring.png).
Journeys 102 and 105 additionally check the lazy mount boundary during real stream
inspection and fixed-read return, with existing Quint ledgers and rejecting controls.

Journey 107 now checks contextual lookback labels across routes, hides the default on Saved, and preserves its value without source reads. [Mobile Saved without a misleading lookback](screenshots/107d-saved-scope-without-lookback.png). The [acceptance audit](../S2_A3_ACCEPTANCE.md) records bounded S2/A3 delivery.

[Shared native values and coverage](../NATIVE_PRESENTATION.md): Journey 59 opens
service coverage without a source request, including [desktop](screenshots/59g-shared-native-coverage.png)
and [mobile](screenshots/59h-mobile-native-coverage.png). Journey 54 opens
[comparison window coverage](screenshots/54g-shared-window-coverage.png), checking
UTC bounds and exact nanosecond tooltips. Native navigation/timeline and live-counter
storyboards use the same presenter while retaining their own fixed/source/custody
contracts. Screenshot links are populated by the executable real-source journeys.

[Standalone tail and native identity inspection](../RECORD_PRESENTATION.md#standalone-tails-and-native-identity)
now share the selected-item shell. Journey 66 pauses on a retained log, searches
its typed fields, closes to the opener and reopens with cached search without a new
source request. Journey 92 preserves a held span population during trace inspection
and disconnect/reconnect. Journey 54 inspects native identity across exact-window
comparisons, preserving search and focus without reading telemetry. Journeys 55/59
inspect the same identity from supported-window previews and service summaries.
Native navigation/timeline journeys retain independent value/coverage semantics.
Tail and native-window Bombadil actors now inspect the new entry points and reject
legacy inline attribute disclosures in these surfaces. Updated [paused tail](screenshots/66b-bounded-tail-keyboard-pause.png), [mobile tail](screenshots/66c-mobile-live-tail.png) and [mobile native identity](screenshots/54f-mobile-stream-identity.png) frames show the real inspection flows.

The [shared record inspection cut](../RECORD_PRESENTATION.md#compact-record-inspection)
removes duplicate inline disclosures from fixed/saved/context tables. Journey 29
now opens the single inspector, searches qualified typed fields, copies multiline
and long values exactly, runs a breakdown/filter and inspects linked logs on desktop
and mobile. Journeys 32/45 exercise saved inspection and fixed refinement through
that same entry point; 97 retains multi-record shared-field actions. Updated
inspection/capture Bombadil actors require no duplicate disclosure and at most one
visible inspector/detail slot per table. These actors do not add source actions to
display components. Cross-app report and generated retry/reject/missing-terminal
campaign helpers now inspect each supplied record's raw evidence through the same
UI instead of reading the removed per-row JSON disclosure. Fixed-query campaigns
enter the explicit Questions route; fresh Explore retains its live-entry behavior.

[Journey 107](../../journeys/saved-owner.spec.js): open real stored spans without
query controls → inspect locally, close to the opener and reopen with retained search → explicitly expand cross-service trace/log context
→ close to the opener → open a filtered saved log → read surrounding logs → hold
a real response, switch captures and reject its late publication. Frames:
[explicit trace](screenshots/107a-saved-explicit-trace.png),
[log context](screenshots/107b-saved-log-context.png),
[mobile saved owner](screenshots/107c-mobile-saved-owner.png).

Journeys 32/41/45/50 now exercise [shared saved-record scope](../SAVED_RECORD_PRESENTATION.md)
in captures and document references. Journey 32 verifies source/custody/count facts,
page offsets and source-free fixed staging before an explicit Run. Existing frames
32a–32f, 41a–41b and 45d document the updated surfaces.

Journey 46 also opens the existing held service span page, inspects shared UTC/fields
and reported context, closes back to its opener, and verifies no local source reads.
Frames: [held service spans](screenshots/46g-held-service-spans.png) and
[mobile inspection](screenshots/46h-mobile-service-spans.png).

[Journey 106](../../journeys/record-presentation.spec.js): inspect a real log in
Live → stage/run a fixed query → compare the same presentation and local detail →
return → repeat for a span with a held parent and an unrelated same-ID span →
inspect mobile → verify standalone server-rendered/SPA components without APIs.
Frames: [shared log detail](screenshots/106a-shared-log-detail.png),
[reported span context](screenshots/106b-shared-span-context.png),
[mobile record](screenshots/106c-mobile-shared-record.png).
[Display contract and diagram](../RECORD_PRESENTATION.md).

[Journey 105](../../journeys/paused-return.spec.js): inspect a held generic log →
stage a field filter → open a fixed child → return while a real response is held →
release that response without replacing the parent → reopen/run the retained fixed
editor → return on mobile → explicitly Apply/Resume. It checks source counts,
selected identity, draft scope, viewport/inspector scroll and focus. Frames:
[pending child](screenshots/105a-pending-fixed-return.png),
[returned selection](screenshots/105b-returned-paused-selection.png),
[mobile return](screenshots/105c-mobile-paused-return.png).
The deliberate transport delay holds an actual source response; data is not mocked.

[Journey 104](../../journeys/live-filters.spec.js): observe four ordinary logs →
inspect west → stage Include without fetching → Apply one matching row → stage
Exclude and Apply east/empty (missing stays excluded) → remove the chip and Apply →
Resume explicitly → inspect empty → stage a fixed question without fetching → Run
one fixed read. Frames: [draft](screenshots/104a-staged-live-filter.png),
[filtered mobile](screenshots/104b-filtered-mobile-stream.png),
[fixed staging](screenshots/104c-staged-fixed-investigation.png).
[Contract](../LIVE_ACTIVITY.md). These frames are generated by the running lab.

[Journeys 102–103](../../journeys/stream-entry.spec.js) cover fresh live entry, scoped logs/spans, source-free inspection, explicit resume gaps, inert saved links, scroll anchoring, empty/missing scope and reconnect failure. Entry checks require the first desktop record above 300px and the first mobile record above 340px; the mobile Menu disclosure must not start another read. [Contract and flow diagram](../LIVE_ACTIVITY.md). Frames: [desktop](screenshots/102a-live-entry.png), [inspection](screenshots/102b-paused-inspection.png), [mobile](screenshots/102c-mobile-live-entry.png), [scroll anchor](screenshots/103a-paused-scroll-anchor.png). The hidden-document event is injected; the OTLP records and normal source reads are real.

[Journey 90](../../journeys/population-workspace.spec.js) exercises generic cross-service OTel populations, excluded resource/attribute groups, missing-versus-empty selectors, paired range filtering, explicit companion filters, column/layout changes, configuration save/reopen/run, partial failure, cancellation and fresh-window staging. [Storyboard, diagram and frames](../POPULATION_WORKSPACE.md).

[Journeys 88–89](../../journeys/linked-timeline.spec.js) cover a generic OTel service timeline, source-free shared range preview, Apply/reset across record and native widgets, mobile layout, Quint trace validation, Bombadil exploration and standalone range-component conformance. [Storyboard and frames](../LINKED_TIMELINE.md).

Journey [87](../../journeys/trace-finding.spec.js) follows a generic API population into a cross-service trace, exercises a failed second capture with retained first receipt, then stages and explicitly saves two capture-linked questions plus an authored finding. It reopens stored evidence without recapture, checks mobile overflow, validates the observed ledger with Quint/quint-go and supplies a Bombadil navigation/composer actor. [Storyboard and frames](../TRACE_FINDINGS.md).

These stories explain the current prototype and guide the intended product.
Screenshots are generated by real Playwright interactions with the running
sample app and investigation UI; backend responses are not mocked.

Run the default Compose stack, then `npm ci` and `npm run test:journeys`.
The wrapper pauses only the demo traffic generator, drains queued work without
deleting data, runs Chromium, and restores the generator if it was running.
Existing stale pending entries from prior crashes are retained and reported.
The sample does not reclaim those entries; the wrapper does not pretend to fix it.
The cloud environment uses the installed `/usr/bin/chromium`; set `CHROMIUM_PATH`
for another local binary. Test endpoints can be changed with `GOO_SAMPLE_URL` and
`GOO_QUERY_URL`, but the default wrapper manages the local Compose demo only.

The live app labels evidence `demo-unverified`, SLO/alarm results as previews,
and native metrics as points. These are visible limitations, not hidden test
exceptions. Examples uses the original bounded Python SQL templates; Explore and
operation packages use Go.
The new Go lifecycle accounting has separate fixtures and is not wired into it.

For the source/build, telemetry and investigation paths behind these journeys, see
[architecture and process diagrams](../ARCHITECTURE_DIAGRAMS.md). Production patterns
and unimplemented portability steps are labelled separately.

## 1. A request retries, succeeds, and crosses an async trace boundary

**Question:** Did a transient attempt fail the reservation, or did the eventual
business operation succeed?

| Frame | User action and evidence | Executable check |
| --- | --- | --- |
| [Create retry](screenshots/01a-create-retry.png) | Submit retry scenario through the actual sample page; retain the returned ID | Real HTTP → gRPC acceptance returns an ID |
| [Completion](screenshots/01b-retry-completion.png) | Open the investigation UI for that reservation | One subject matures to good; evidence remains unverified |
| [Async spans](screenshots/01c-linked-consumer-spans.png) | Click the subject's spans link | Consumer work is present, with a nonempty cross-trace link |
| [Attempt evidence](screenshots/01d-retry-log-evidence.png) | Inspect related logs | Retry and terminal events exist for the same reservation |

This teaches the distinction between attempts, subjects and trace identities.
A later UI should expose explicit scoped relations rather than correlating on
reservation ID alone; that is a planned improvement, not current coverage.

## 2. Fast acceptance does not establish eventual completion

**Question:** Why can a synchronous endpoint look healthy while its asynchronous
outcome is late, absent, or rejected?

The test creates all four scenarios through the sample page, waits for mature
outcomes, then inspects each selected subject. Frames: [good](screenshots/02-good.png),
[late](screenshots/02-late.png), [missing](screenshots/02-missing.png),
[rejected](screenshots/02-reject.png), and [synchronous latency](screenshots/02-synchronous-latency.png).
Good is asserted good; the others are asserted bad under the **demo's** missing
completion assumption. The latency view contains real synchronous observations,
but the story does not equate a low latency number with business success.

Under background traffic, even nominal good/retry requests can miss the demo
deadline while queued behind slow work. That observed behavior motivated isolating
the documentation workload; it was not fixed by weakening the assertions or
clearing the stream. Capacity/queue latency remains a useful future investigation.

## 3. Understand what the definition and query can claim

**Question:** Which definition did I use, what SQL ran, and are these numbers rates
or exported points?

Frames: [owned package](screenshots/03a-owned-definition.png),
[invalid expression](screenshots/03b-rejected-expression.png), and
[metric points/explain](screenshots/03c-metric-points-and-explain.png).
The test opens the actual package, submits an unsupported query and sees an
error, then successfully queries metric points and opens generated SQL. The UI
must retain its warning that the points are not derived rates.

## 4. From a latency distribution to linked asynchronous work

**Question:** Which calls make up this latency band, and what happened after
synchronous acceptance?

| Frame | User action and evidence | Executable check |
| --- | --- | --- |
| [Raw-duration heatmap](screenshots/04a-duration-heatmap.png) | Create a retry reservation, then inspect its gRPC duration distribution | Full matching span count, no percentile reconstruction |
| [Selected cell](screenshots/04b-cell-span-details.png) | Click a populated cell and select a matching span | Returned count matches the cell; exact timestamp lies inside its half-open bounds |
| [Reservation waterfall](screenshots/04c-reservation-waterfall.png) | Open the span's reservation to widen the investigation | Consumer work appears under its own trace; parent edges stay within each trace |
| [Follow async link](screenshots/04d-follow-async-link.png) | Select the consumer and follow its producer reference | The referenced span's exact ID appears in the details |

All four steps use real collector ingestion and ClickHouse queries. Cell queries
preserve fixed time bounds, but are not historical evidence snapshots. Trace
inspection remains capped at 200 spans and explicitly labels unknown completeness.
See [visualization contracts](../VISUALIZATIONS.md).

## 5. Discover the next question without losing the evidence scope

**Question:** How do I choose a useful next step, and what measurements are missing?

| Frame | User action and evidence | Executable check |
| --- | --- | --- |
| [Retry guide](screenshots/05a-guided-retry-evidence.png) | Select the retry-recovery question and inspect lifecycle evidence | The selected reservation survives the guided step into logs |
| [USE evidence inventory](screenshots/05b-use-coverage-gaps.png) | Choose resource pressure and inspect queue/worker evidence | Queue/worker telemetry is available; CPU/memory/pool measurements remain absent; resource scope clears unsupported subject filters |
| [RED exploration](screenshots/05c-red-volume-and-duration.png) | Choose service latency and open volume/duration | The same live raw-span heatmap opens; missing error-rate support is stated |

These are navigable query templates with explanations, not automated root-cause
diagnoses or a complete USE/RED implementation. See [workflow contracts and
follow-on instrumentation](../INVESTIGATION_WORKFLOWS.md).

## 6. Exercise a concrete adapter and discover reusable semantic packages

| Frame | User action and evidence | Executable check |
| --- | --- | --- |
| [Live worker/queue evidence](screenshots/06a-live-queue-and-workers.png) | Create a slow reservation and inspect USE resource evidence | Actual active-worker observation, all six metric rows, explicit instance identity and four plotted panels |
| [Semantic package library](screenshots/06b-semantic-package-library.png) | Open the prebuilt library and database-pool profile | Requirements include pool wait telemetry; outputs remain labelled proposed |

The Go `goo discover` command independently recognizes contract shapes from an
explicit inventory. The UI library is a catalog, not an automatic live discovery
scan. Hegel checks that wrong units, unknown versions, absent requirements or
separate scopes cannot activate the incomplete scope's profile. See
[semantic packages](../SEMANTIC_PACKAGES.md) and [queue contracts](../QUEUE_TELEMETRY.md).

## 7. Evaluate a bound package over live OTel spans

| Frame | User action and evidence | Executable check |
| --- | --- | --- |
| [Shared HTTP view](screenshots/07a-live-http-package-view.png) | Create a reservation and request the controlled failing HTTP endpoint; evaluate the Go package | Real ingested success/failure routes, columns from the expansion descriptor, visible sampling/completeness warnings |
| [Source and policy](screenshots/07b-http-source-and-policy.png) | Open the source/revision panel | Actual library binding, query parameters, content revision and explicit failure-status policy |

This executes HTTP → collector → ClickHouse → Go preparation/evaluation → shared
view rendering. It does not create a custody snapshot, materialize prepared export
tables or deploy Databricks objects. See [source contracts](../LIVE_HTTP_PACKAGE.md).

## 8. What did we know then, and what changed? — partly implemented

This is a historical investigation-UI design storyboard, not a fabricated browser
journey or the current execution queue:

1. Select a fixed cohort, definition revision, custody clock and evidence basis.
2. Inspect an original evaluation: a mature subject has no visible completion.
3. Advance only the evidence basis; an on-time terminal delivered later appears.
4. Show a new generation linked to the original, with changed outcome and counts.
5. Reopen the original receipt and reproduce its original answer.
6. Export the same definition/facts to the longer-term platform without combining
   generations or erasing evidence quality.

The Go CLI fixtures, Hegel properties and Quint replay execute the semantic
parts of this story, including strict cutoff boundaries and old/new snapshot
coexistence. The live UI has no basis selector, generation comparison or Databricks
deployment yet; no screenshots or browser-test claims are made for those steps.

See [lifecycle preparation](../LIFECYCLE_PREPARATION.md),
[bitemporal design](../BITEMPORALITY.md), and
[verification](../../verification/README.md) for the executable lower layers.

## Package-binding CLI journey

The [HTTP binding storyboard](../PACKAGE_EXPANSION.md) executes a separate
Go package workflow: recognize/bind → expand shared artifacts → prepare normalized
server requests → evaluate/reference-SQL parity. It uses fixed fixtures and
actual CLI/embedded ClickHouse execution. The separate live HTTP journey now renders its aggregate view descriptor. The
fixed CLI fixture remains distinct from that real-ingestion evidence.

## Annotation/weaving developer journey

The [annotation storyboard](../GOO_ANNOTATIONS.md) executes real Go weaving and
captures actual OTel SDK spans/events/metrics, checks application error/panic
behavior and compares emitted signal contracts with the source manifest. It is
a developer conformance journey rather than a new browser/collector journey;
no UI mockups or additional screenshots are claimed.

## Captured artifacts

Named documentation PNGs live in `screenshots/`. Playwright HTML reports,
failure screenshots and traces are generated under `artifacts/`. Screenshots
contain synthetic demo IDs and timestamps and change on recapture. They document
asserted behavior rather than serving as fragile pixel-perfect golden tests.

## Live woven operation journey

Journey 08 runs real Go success/error/panic paths through the collector and
ClickHouse, then inspects the operation overview and evidence binding. See
[the storyboard and captured frames](../LIVE_OPERATIONS.md).

## Operation duration and trace drilldown

Journey 09 follows a failed woven Go checkout through its operation error count,
raw duration heatmap cell, exact contributing spans and parent/child trace
waterfall. It verifies retained resource/window/basis without an additional
operation query. See [the storyboard](../OPERATION_DRILLDOWN.md).

## Scoped subject journey

Journey 10 creates related Go invocations plus an identical ID in another tenant,
then navigates trace → subject events → related traces without a new query.
[Storyboard and captured frames](../SUBJECT_NAVIGATION.md).

## Reusable package onboarding

Journey 11 generates a real report in a second Go app and explores its YAML-backed
operation/subject views. Journey 12 holds a real response to verify package
switching discards stale evidence. [Storyboards](../OPERATION_PACKAGES.md).

## Native counter rate journey

Journey 13 uses actual exported cumulative counter points, inspects supported
interval rates and their separate basis, then opens contextual spans.
[Storyboard](../NATIVE_COUNTERS.md).

## 14. Native duration histograms

**Question:** What can collected metric distributions tell us without pretending
that a percentile is an exact duration or that a metric bucket identifies spans?

| Frame | User action and evidence | Executable check |
| --- | --- | --- |
| [Native heatmap](screenshots/14a-native-histogram-heatmap.png) | Invoke the real woven app; inspect checkout over two minutes and expand native histograms | OTLP → collector → ClickHouse → Go returns exact bucket totals; unknown intervals have no distribution |
| [Bucket bounds](screenshots/14b-histogram-bucket-bounds.png) | Select a populated duration bucket | Exact population and whole-interval P95 bucket bounds; no interpolation or contributor claim |
| [Context](screenshots/14c-histogram-contextual-spans.png) | Inspect histogram context | Operation/time selection uses captured spans, retaining separate metric evidence |

Implemented by test `14 native histogram…` in `journeys/investigation.spec.js`.
The live source is cumulative; delta, resets, late repair and explicit aligned
merging have separate Go/Hegel/Quint checks. Fleet discovery and merged UI views
remain planned. See [native histogram semantics](../NATIVE_HISTOGRAMS.md).

## 15. From operation error ratios to traces

**Question:** Which observed error interval deserves inspection, and what can its
captured spans explain without claiming exact metric attribution?

| Frame | User action and evidence | Executable check |
| --- | --- | --- |
| [Calls/errors/ratio](screenshots/15a-operation-error-ratios.png) | Run real successful and failing checkout calls; expand operation ratios | Calls/errors share one counter basis; supported percentages stay within 0–100; unknown/idle intervals have no percentage |
| [Guided investigation](screenshots/15b-guided-error-investigation.png) | Choose “Why did errors rise?” on a populated error interval | Explicit counts and outcome choices, with no automatic cause/change claim |
| [Contextual spans](screenshots/15c-contextual-error-spans.png) | Inspect error context | Captured checkout/server/error spans selected by interval, retaining separate evidence |
| [Trace waterfall](screenshots/15d-error-trace-waterfall.png) | Open an available matching trace | Existing scoped waterfall contains checkout operation; switching packages clears ratio evidence |

Implemented by test `15 operation error ratios…`. The live source is cumulative;
missing/reset/gap/idle/delta semantics have independent Go/Hegel/Quint checks.
See [error-ratio semantics](../OPERATION_ERROR_RATIOS.md). This original journey
does not depict fleet views or change detection; the following journeys record
the later delivered bounded fleet coverage and comparison work.

## 16. Declared fleet membership and interval coverage

**Question:** Does observing both instances establish a common fleet population?

| Frame | User action and evidence | Executable check |
| --- | --- | --- |
| [Declared members](screenshots/16a-declared-fleet-members.png) | Invoke both live checkout instances; expand fleet coverage | Both declared instances observed through actual OTLP/source queries |
| [Unknown exact total](screenshots/16b-unsupported-fleet-total.png) | Inspect an unsupported interval | Differing SDK windows leave exact fleet calls/errors/percentage unknown |
| [Coverage explanation](screenshots/16c-fleet-coverage-explanation.png) | Explain that interval | Per-member states distinguish observation from exact alignment |

Test `16 declared fleet…` checks package clearing and separate member evidence.
Missing-member and failed-source cases retain Go/Hegel/Quint fixtures. See
[fleet coverage](../FLEET_COVERAGE.md) and the bounded comparison journey below.

## 17. Two live instances, imbalance, loss and restart

**Question:** How can an operator compare independently timed telemetry and still
recognize loss and reset without constructing an unsupported fleet percentage?

| Frame | User action and evidence | Executable check |
| --- | --- | --- |
| [Actual windows and imbalance](screenshots/17a-actual-member-windows-and-imbalance.png) | Send more successful traffic to woven-demo and errors to woven-peer; explain a supported comparison | Actual windows/counts retained, endpoint spreads within policy, no combined percentage |
| [Requests during metric loss](screenshots/17b-peer-serves-requests-with-metric-loss.png) | Pause peer metric exports through the sample app, then make a successful checkout | Real HTTP200 while explicit lab metric gate is off |
| [Unknown during loss](screenshots/17c-peer-loss-stays-unknown.png) | Investigate after more than the maximum counter gap | Main has a supported interval; unmatched peer has no invented counts/percentage |
| [Restart epoch](screenshots/17d-restart-epoch-is-not-a-comparison.png) | Resume metrics, observe the excessive gap, then restart only the local peer | Gap and new counter epoch remain unknown; raw counter cause is explained |

Test `17 live peer…` restores metric export in `finally`. Restart control manages
only the owned local Compose peer; volumes, pending entries and other services
remain. The default stack now has twelve services and loopback8092 for the peer.
No automatic clock correction or common-window aggregation is depicted.
See [comparison semantics](../FLEET_COMPARISONS.md) and diagram8.

## 18. A late initial response preserves an edited lookback

**Question:** Can startup work silently replace an operator's chosen query window?

| Frame | User action and evidence | Executable check |
| --- | --- | --- |
| [Edited lookback retained](screenshots/18a-edited-lookback-survives-late-response.png) | Delay the actual initial query response; edit lookback to two minutes, release it, then inspect operations | Edited input stays2 and the operation request/result spans exactly120 seconds |

Test `18 late initial…` delays delivery of a real upstream response, not its data.
This guards the shared control race found by the loss journey; source limits stay
unchanged and excess rows still reject rather than becoming partial success.

## 19. Save, reopen, refresh and compare

An investigator selects checkout errors and a trace, captures a fixed window,
and retains the receipt for handoff. The packages process restarts to demonstrate
volume persistence. The journey ingests a labelled late fixture span via real
OTLP in the old event window; reopening still shows the original capture/trace.
Refreshing makes a successor with the same question and selection, includes the
late observation and compares separate bases. The old receipt stays unchanged.
No browser operation/metric fetches are permitted during reopen. Saved evidence has its own presenter: the live Operations package/evidence/selection survives saving and return navigation unchanged. A full reload verifies the saved receipt path.

- [Saved receipt](screenshots/19a-saved-fixed-evidence-receipt.png)
- [Reopened trace context](screenshots/19b-reopened-trace-context.png)
- [Late-evidence revision comparison](screenshots/19c-late-evidence-successor-comparison.png)

This test controls only the local Compose packages service and collector. It
adds explicitly labelled test telemetry; no real business completion is invented.
See [capture contract and limitations](../SAVED_INVESTIGATIONS.md).

## 20. General exploration and population-safe paging

A labelled collector fixture emits 123 spans across two synthetic resource names.
The investigator discovers arbitrary keys, filters a cohort and groups a key with
61 present-empty values and 62 missing values. A 100-row page and its 23-row
successor keep the same full captured-record count/group results. The fixture
also exercises an SQL-looking key as a literal, AND/OR, consistent no-data and a
delayed genuine response discarded after a signal change. These are controlled
telemetry records, not real service executions.

- [Independent counts and missing/empty grouping](screenshots/20a-full-counts-and-missing-key-groups.png)
- [Bounded next page](screenshots/20b-bounded-next-page.png)
- [Trace page with parents outside the page](screenshots/20c-trace-paging-keeps-partial-context-visible.png)

## 21. Real distributed trace and exact linked logs

A real reservation request passes through web and gRPC. Starting with a gRPC-only
attribute search, the investigator expands its exact trace to include the web
parent in the same environment/time bounds. Original filters/basis remain
separate. The UI shows the cross-service waterfall and trace-linked acceptance
logs; clicking a span narrows logs to exact trace AND span IDs. Async consumer
trace links are retained as references, not silently joined as parent edges.

- [Cross-service waterfall and trace logs](screenshots/21a-cross-service-waterfall-and-linked-logs.png)
- [Exact selected-span log scope](screenshots/21b-exact-span-log-scope.png)

See [the query and expansion contract](../EXPLORATION_DESIGN.md).

## 22. A saved operation expands independent live context

The investigator reopens an immutable Go checkout capture, then explicitly
expands its trace to live cross-service context and exact log lookup. The original
fixed environment/time bounds and origin basis remain visible. The saved receipt
is checked before/after and stays equal; expanded evidence has separate questions
and bases and is not silently archived with the operation capture.

- [Saved origin and independent live context](screenshots/22a-saved-question-and-independent-live-context.png)

## 23. One investigation workspace from distribution to trace

A labelled search-index refresh OTLP fixture supplies 12 captured records with region and outcome
attributes, including missing and present-empty regions. The investigator groups
by region, pins the baseline, filters to four error records and sees east rise
from 3/12 (25%) to 3/4 (75%), a +50 percentage-point difference. Both questions
retain identical time bounds. This is a controlled documentation population,
not a claim about real production errors or causality.

| Frame | User action and evidence |
| --- | --- |
| [Distribution](screenshots/23a-distribution-workspace.png) | Explore complete captured groups inside the shared product shell. |
| [Comparison](screenshots/23b-pinned-baseline-comparison.png) | Compare two explicitly sized populations within a fixed window. |
| [Definitions](screenshots/23c-definition-workspace.png) | Inspect contracts, then return without losing comparison state. |
| [Mobile comparison](screenshots/23d-mobile-comparison.png) | Use the same navigation/result tabs on a 390px viewport without document overflow. |
| [Mobile inspector](screenshots/23e-mobile-trace-inspector.png) | Open a selected trace and close with Escape. |

The executable journey also checks keyboard tab navigation, distribution-value
refinement within the pinned window, an empty selection
without invented percentages, baseline clearing after a signal change and no
uncaught browser errors. Earlier journeys now navigate actual workspace destinations;
reference workflows live under Examples. See [workspace contract](../PRODUCT_WORKSPACE.md).

Journey 11 also expands the real report application into the generic trace inspector,
with report-specific attributes and no reservation-specific action.

## 24. Build a package over existing Python telemetry

A small real Python OTel SDK process emits six `index.refresh` spans through the
collector. Set `GOO_PYTHON` to a Python interpreter with the project requirements
installed; the managed-cloud default is `/workspace/.venvs/goo/bin/python`. It has no goo attributes, annotation or source manifest. This is a
controlled SDK documentation workload, not a production business outcome. The user
filters its cohort, groups region (including one missing key), creates a package
and previews the same six records in the original fixed window. The saved package
has a reusable lookback and no page offset or absolute evaluation timestamp.

| Frame | Action and evidence |
| --- | --- |
| [Author and preview](screenshots/24a-author-existing-python-telemetry.png) | Build a definition from captured data; preview counts/groups with separate basis. |
| [Instrumentation handoff](screenshots/24b-optional-instrumentation-handoff.png) | Choose an explicit Go operation name/role and review the optional template. |
| [Reuse in place](screenshots/24c-reuse-package-without-source-manifest.png) | Restart the Go packages process, reopen the persisted package and explore the unchanged Python telemetry in a fresh window. |
| [Import and revise](screenshots/24d-import-and-revise-definition.png) | Export/import YAML and save an edited revision while retaining the original. |
| [Mobile authoring](screenshots/24e-mobile-package-authoring.png) | Author the same definition on a 390px viewport without document overflow. |

The executable journey also verifies that local query refinements preserve the
base selector without mutating the saved revision, and that delayed genuine draft/validation
responses cannot replace edited metadata or enable saving newer, unvalidated YAML.
The Create action reveals the editor and focuses its name field. Generated Go source is
separately passed through the actual annotation scanner in Go tests; the browser
does not modify the application. See [package contract](../TELEMETRY_PACKAGES.md).

## 25. Import a nested log package without a source manifest

A labelled OTLP log fixture represents a Node-style job logger with four records,
ERROR/INFO severities and present/missing queue attributes. These are controlled
collector records, not a claim that a Node SDK/service was executed. The user
imports nested AND/OR YAML directly in Definitions, applies it to logs, and adds
a severity refinement. Both imported breakdowns and the original selector remain
intact; records without valid trace IDs remain unlinked. A preview of an absent
selection returns `no_data`, without invented successes. A delayed genuine package
apply response cannot replace a newer environment selection after leaving Definitions.

- [Nested log package and two breakdowns](screenshots/25a-nested-log-package-and-two-breakdowns.png)

## 26. Correlate a real application with goo during generated UI exploration

Run `npm run test:cross-app`, optionally with `GOO_CROSS_APP_QUINT=1` to check the
actual event ledger against a Quint model using the sibling quint-go tools.

| Frame | User action and independently checked evidence |
| --- | --- |
| [Generate a real report](screenshots/26a-correlated-real-report.png) | A user submits the report app while a separate Bombadil browser explores goo navigation. The returned report ID is recorded; this is a reported outcome, not a durable job guarantee. |
| [Find the captured trace](screenshots/26b-correlated-goo-trace.png) | A direct ClickHouse read finds its emitted subject/trace. The observer filters Explore by that trace and compares rendered IDs/counts with an independent source read over exactly the UI's frozen window, then opens the waterfall. |

The coordinator retains run identity, actual source SQL/results, question/basis,
causally ordered checkpoints and Bombadil artifacts under ignored Playwright output.
Quint optionally checks matching identity/scope/window/count/membership and checkpoint
order; its protocol fixture tests reject twelve mutations and incomplete runs.
Bombadil's bounded action vocabulary currently covers navigation only. It does not
generate application requests or check the selected trace itself; the Playwright
observer does those checks. This exercises concurrent independent browsers, not a
claim of deterministic distributed scheduling or complete telemetry capture.

## 27. Follow a generated workload and revisit captured evidence

Run `npm run test:generated-campaign` with the lab and Go/quint-go available.
`GOO_CAMPAIGN_SEED=99` chooses another recorded schedule;
`GOO_CROSS_APP_QUINT=1` additionally checks every evidence ledger with qobs.
The ordinary browser suite skips this opt-in journey.

The model schedules three actual report submissions, an inspection for each,
two revisits of already inspected results, and completion. Bombadil submits through
the app button in fresh browser profiles while another Bombadil actor explores goo.
Returned IDs come from actual captured DOM states. The Playwright observer then
selects the corresponding trace and compares rendered members/counts against an
independent ClickHouse query over exactly the UI's frozen time bounds.

| Frame | Action and evidence |
| --- | --- |
| [Bombadil's application frame](screenshots/27c-bombadil-application.jpeg) | Unchanged screenshot captured by the app actor after its real request. No second request was staged for documentation. |
| [Generated inspection](screenshots/27a-generated-inspection.png) | A model-selected report index resolves to its actual returned identity and captured trace. |
| [Generated revisit](screenshots/27b-generated-revisit.png) | Close the prior inspector, clear filters, and inspect a previously checked report in a fresh read. The generated seed determines ordering and target. |

The driver records effects only after real handlers succeed, then quint-go
`connect.Driver` compares all nine observed checkpoints with the generated ITF
poststates. A complete schedule and positive workload/source/UI exercise are
required. The artifact reader rejects missing/duplicate submissions and unstable
identities; projection tests reject changed counts/targets, reordered and truncated
campaigns. Plans, seed, partial observations, actor traces and evidence checks stay
in ignored runtime artifacts. The model does not infer business outcomes, verify
durability or claim deterministic scheduling. This report campaign remains
success-only; the separate Python fault campaign follows below.

## 28. Investigate retry, rejection and stalled work

Run `npm run test:fault-campaign` with the local Compose lab and Go/quint-go.
The wrapper pauses/restores owned background traffic and drains active work without
deleting pre-existing stale stream entries. This opt-in journey uses the same seeded
Quint schedule as journey 27, with three fixed fault coverage cases and two revisits.

Bombadil selects a scenario and clicks the real Python app button in a fresh
browser. The returned ID scopes independent ClickHouse reads and ordinary Explore
span/log queries; no source annotations, backend responses or telemetry are mocked.
Another Bombadil actor explores goo navigation concurrently. Counts, record
identities and attributes are checked against each UI query's exact fixed window.

| Frame | Question and captured evidence |
| --- | --- |
| [Retry then confirmation](screenshots/28-retry-captured-evidence.png) | An error attempt and retry are visible alongside the later confirmed terminal; an attempt error is not itself terminal failure. |
| [Business rejection](screenshots/28-reject-captured-evidence.png) | A terminal log carries `rejected` even though the spans are non-error. Transport/attempt status and business outcome remain distinct. |
| [Stalled work](screenshots/28-missing-captured-evidence.png) | A positive stalled event is present, while no terminal is captured in the bounded question. The test records `unknown-no-terminal`, without inventing success or rejection. |

These labels are the test observer's interpretation of raw displayed records,
not new goo health badges or durability claims. The Quint fault model checks each
inspection, including explicit terminal-presence flags and question/record
fingerprints. Hegel state/swarm checks over the actual test projection cover
duplicates, ordering, conflicting outcomes and missing terminals. A present empty
terminal is distinguished from absence. Missing required source evidence fails
readiness instead of satisfying a fault case. Seed, partial observations, actor
traces, raw source reads, UI evidence and model reports remain archived under
`artifacts/bombadil/fault-campaigns/<run-id>/`.

## 29. Structured inspection and filter composition

`journeys/ui-polish.spec.js` publishes a controlled documentation fixture through
the real OTel collector and reads it back through ClickHouse and the Go query API.
It does not claim SDK instrumentation coverage. Six spans and a correlated log
exercise duplicate resource/span keys, large numeric-looking strings, exact
whitespace, Unicode, long values and HTML-like captured text.

| Frame | User journey |
| --- | --- |
| [Structured attributes](screenshots/29a-structured-attribute-inspection.png) | Open a record, search by key/value and scope, copy exact values, expand long values and retain raw evidence as a secondary disclosure. |
| [Facets](screenshots/29b-facets-and-explicit-empty-values.png) | Break down from an attribute, search categories without changing denominators, distinguish empty from missing and remove a filter chip. |
| [Trace context](screenshots/29c-selected-trace-and-structured-context.png) | Expand the six captured spans across services, select a span and inspect structured fields; detailed evidence limitations remain available. |
| [Mobile](screenshots/29d-mobile-attribute-inspection.png) | Read and search an attribute on a narrow screen without horizontal page overflow. |

Assertions cover exact clipboard content, escaped text, local-search query counts,
ALL/ANY populations, missing-value controls, duration-only ordering, selected-span
state, Escape focus restoration and mobile layout. Pure helper tests additionally
cover nested value types and literal searches. Counts always describe captured
records; the fixture does not establish production completeness or sampling guarantees.

## 30. From log search to reusable span distributions

`journeys/discovery.spec.js` publishes a labelled OTLP documentation fixture to the
real collector: 13 spans with exact bucket-boundary durations, 108 same-service
logs, plus other-service/environment controls. This proves the captured query/UI
path, not SDK instrumentation or production completeness.

| Frame | User journey |
| --- | --- |
| [Literal log search](screenshots/30a-literal-log-search.png) | Find exact service/severity and a literal message containing regex punctuation; inspect readable messages and visible filter chips. |
| [Surrounding logs](screenshots/30b-independent-surrounding-logs.png) | Read nearby same-service logs with different messages/severities/trace IDs in a separate bounded time context. |
| [Matching traces](screenshots/30c-matching-traces-and-expanded-evidence.png) | Four matching error spans lead to seven captured spans across services; original filters do not silently constrain expanded context. |
| [Duration distribution](screenshots/30d-full-population-span-duration-buckets.png) | Inspect complete matching-span counts in fixed half-open duration intervals. |
| [Reusable package](screenshots/30e-reusable-discovery-package.png) | Build, preview, save and reuse the selected filters/duration grouping without a source manifest. |
| [Mobile discovery](screenshots/30f-mobile-log-discovery.png) | Search logs with narrow-screen controls and no horizontal page overflow. |

Assertions cover literal search, severity, full context count versus its 100-row
page, next-page count stability, service/environment exclusion, exact anchor fields,
escaped multiline/long messages, focus return and late-read isolation. An independent
fixture expectation checks all eight duration intervals, including exact boundaries.
The package retains filters/grouping across fresh reuse; it saves configuration,
not evidence. Time context is not causal correlation, and matching-span counts
are not complete trace/root-latency measures.

With `GOO_DISCOVERY_QUINT=1`, the journey writes a permanent ignored ledger under
`artifacts/discovery/` and checks its actual context, trace and package questions
against Go references through qobs/quint-go. Three negative controls reject widened
context, a different trace environment and lost package grouping. Hegel state/swarm
separately explores context anchors/radii/pages; partition properties and Go unit
checks cover half-open buckets and UInt64 limits.

`npm run test:discovery-campaign` also adds 100 worker spans in this isolated
fixture after the deterministic assertions, then runs the dedicated Bombadil
actor. It must reach every prefix stage and observe 113 matching duration-group
members while the row page displays 100. Its actions include quick search from
an empty OR composer, trace/context paging, radius changes, re-anchoring, duration
breakdowns and package draft/edit/validate/preview/use. Data invariants and a
second actor-derived Quint ledger complement the deterministic journey. The actor
publishes no definitions; its fixture driver retains the earlier saved package.

## 31. From observed services to a reusable question

`journeys/inventory.spec.js` sends labelled OTLP fixture spans through the real
collector: 113 edge spans, three worker spans and an excluded environment control.
An independent ClickHouse query checks source counts and Error statuses. This
validates the captured discovery path, not SDK instrumentation completeness.

| Frame | User journey |
| --- | --- |
| [Observed services](screenshots/31a-observed-service-inventory.png) | Discover two services, full captured counts, error statuses and span-duration buckets without a manifest. |
| [Fixed-window exploration](screenshots/31b-inventory-fixed-window-exploration.png) | Open one exact service/name pair; 113 matching spans remain distinct from the 100 displayed records. |
| [Reusable preview](screenshots/31c-inventory-reusable-package-preview.png) | Build and independently preview a reusable package from the selected question. |
| [Mobile inventory](screenshots/31d-mobile-service-inventory.png) | Use the same discovery surface on a narrow screen with readable stacked cards. |

Local filtering retains capture totals/basis. Scope edits invalidate pending reads;
no-data results retain unknown health. The actual capture/search/open/package
ledger is always checked with quint-go and three rejecting controls (changed local
total, widened open window, changed package environment). Hegel state/swarm checks
observe/reset accounting. `npm run test:inventory-campaign` adds a read-only
Bombadil actor with a required discovery/filter/clear/open prefix and independent
accounting/rendered-count/scope/stale-evidence invariants; it saves no definitions.
Permanent ignored ledgers, coverage and actor traces live under `artifacts/inventory/`.

## 32. Preserve general evidence and page a stored record bank

`journeys/captures.spec.js` sends a labelled OTLP fixture through the real collector
and ClickHouse: 1,205 edge spans, five worker spans and 107 ordinary log records.
An independent source query checks the edge count. The test then inserts a late
span before the existing sort position. This is a captured-path fixture, not a
claim about SDK completeness, unique invocations or production custody.

| Frame | User journey |
| --- | --- |
| [Saved full population](screenshots/32a-saved-question-full-population.png) | Capture a named question with 1,205 matching spans, complete groups and 1,000 stored inspection records. |
| [Stable record page](screenshots/32b-stable-stored-record-page.png) | Page within immutable stored evidence while preserving count and basis. |
| [Successor revision](screenshots/32c-fixed-question-successor.png) | Refresh the same fixed question to see the late span without altering the predecessor. |
| [Saved inventory](screenshots/32d-saved-inventory-local-search.png) | Preserve the bounded inventory and locally search without changing its totals. |
| [Saved logs](screenshots/32e-saved-log-inspection.png) | Inspect structured log attributes on the second stored page. |
| [Mobile capture](screenshots/32f-mobile-saved-investigation.png) | Read the receipt and captured evidence on a narrow screen. |

Assertions cover exact row replay after late insertion, cross-revision cursor
rejection, actual volume persistence across a packages-service restart, a fresh
browser reopening/paging without live-query requests, copied deep link, explicit
fresh live branch, span/log/inventory component isolation, stale open guards,
full-count versus retained/page accounting and mobile containment. Go HTTP tests
also close the source entirely before reopening and paging, and reject attempts
to submit evidence or change refresh scope. Source-plan changes block refresh.

The actual save/page/refresh/reopen ledger is always checked through quint-go,
with four rejecting controls: changed old counts, changed page basis, changed
refresh scope and acceptance of a cross-revision cursor. Hegel state/swarm explores
arrivals, source changes, captures, branching refreshes and stable row banks.
`npm run test:captures-campaign` adds a read-only Bombadil actor with a required
page-forward/back, distribution and attribute-inspection prefix. It checks count,
retention, scope, basis, replay, read-only controls and unique component identities;
it never saves, refreshes or emits workload. Evidence lives in ignored
`artifacts/captures/<fixture environment>/` directories.

## 33. Core measures, reuse and service discovery

Executable: `journeys/core-basics.spec.js`. Real OTLP spans/logs verify 205 matching
spans versus 100 records displayed, exact upper p95 195 ms, saved-page stability,
empty intervals/no-data, reusable YAML question with a fresh window, log-only
services, reported `api → worker`, ambiguous/unresolved parents, source cancellation
request and mobile containment. Native cancellation tests separately establish Go
context propagation/admission. Optional `GOO_CORE_BOMBADIL=1` runs saved-measure/time-bar and live reuse/discovery/cancellation actors without changing source data. Quint validates the live ledger
and rejects five corrupted variants; Hegel exercises duration state/swarm.

| Frame | User action / asserted meaning |
| --- | --- |
| [33a](screenshots/33a-full-population-measures-and-time.png) | Inspect full-population duration measures and fixed count intervals. |
| [33b](screenshots/33b-reusable-question-fresh-window.png) | Run a saved package question in a fresh source window. |
| [33c](screenshots/33c-service-signals-and-reported-dependencies.png) | Discover span/log services and parent-link diagnostics; event age is not ingestion lag. |
| [33d](screenshots/33d-mobile-saved-measures.png) | Reopen archived measures/chart on a narrow screen with stable values. |

## 34. Duration breakdowns and saved measures

`journeys/grouped-measures.spec.js` uses 124 real OTLP spans with fast/slow/missing/
empty route groups. It verifies full-population p95 versus 100 displayed records,
one/two-field group summaries, exact ranking/search/refinement, cohort measure
accounting, retained pages, package preview and mobile containment. The live
ledger is checked by Quint with three rejecting controls; the optional
`GOO_GROUPS_BOMBADIL=1` campaign explores saved-only ranks/search/pages.

| Frame | User journey |
| --- | --- |
| [34a](screenshots/34a-duration-breakdown-ranking.png) | Rank complete route groups by p95 while keeping the overall measure separate. |
| [34b](screenshots/34b-saved-mean-breakdowns.png) | Compare exact means from an immutable capture. |
| [34c](screenshots/34c-mobile-duration-breakdown.png) | Inspect the same stored breakdown on a narrow screen. |
| [34d](screenshots/34d-grouped-package-preview.png) | Preview total and grouped duration measures from reusable YAML. |


<a id="error-and-latency-comparison"></a>
## 35. Error and latency comparison

`journeys/error-comparison.spec.js` ingests 80 prior and 124 current OTel spans.
The user ranks recorded errors, compares prior-window errors and latency,
saves both populations, reopens the pair and follows a changed route into a
cross-service trace and linked logs. Missing/empty groups remain distinct;
only 100 current inspection rows are displayed. Quint checks the actual ledger
with four rejecting controls. `GOO_ERROR_COMPARISON_BOMBADIL=1` explores saved
measure changes with immutable evidence and exact known-result invariants.

| Frame | User journey |
| --- | --- |
| [35a](screenshots/35a-error-rate-change.png) | Compare recorded errors with the previous fixed window. |
| [35b](screenshots/35b-latency-change-and-missing-groups.png) | Rank p95 changes; absent baseline groups have unavailable differences. |
| [35c](screenshots/35c-saved-population-comparison.png) | Reopen the server-reread pair with links to both record captures. |
| [35d](screenshots/35d-mobile-comparison.png) | Inspect the same comparison on a narrow screen. |
| [35e](screenshots/35e-changed-route-to-trace-and-logs.png) | Follow a changed route to explicit trace and log context. |


## 36. Deployment context and ordered notes

`journeys/context-notes.spec.js` generates a real report in the Go sample, finds
its OTel trace through independent ClickHouse evidence, and opens a fixed Explore
question. The user records a manual deployment claim, pins it with frozen evidence,
edits/reorders hypotheses, saves immutable successor/branch revisions, retains an
unapplied draft and opens exact stored evidence without telemetry execution.
The deployment marker is authored context; the test does not deploy a new sample
version. Quint checks actual observations with six rejecting controls.
`GOO_CONTEXT_BOMBADIL=1` explores local draft operations without publishing them.

| Frame | User journey |
| --- | --- |
| [36a](screenshots/36a-deployment-context-on-captured-time.png) | Read separate event/custody times alongside the captured time chart. |
| [36b](screenshots/36b-ordered-hypothesis-and-frozen-references.png) | Preserve ordered hypotheses with exact capture and deployment references. |
| [36c](screenshots/36c-narrative-revision-and-preserved-parent.png) | Save findings as a new narrative revision with an exact parent. |
| [36d](screenshots/36d-follow-frozen-evidence-without-rerunning.png) | Open a cited capture without rerunning telemetry. |
| [36e](screenshots/36e-mobile-investigation-notes.png) | Continue the same ordered investigation on a narrow screen. |

## 37. Shared examples and scoped investigation recipes

`journeys/recipes.spec.js` generates real report work and independently locates its
scoped ready event in ClickHouse. The primary example gallery runs ordinary
packages through Explore. A trace-scoped recipe preserves fixed bounds/selectors;
preview and narrative actions do not read telemetry. The journey freezes evidence,
saves an exact narrative reference, then derives/reuses a package; wrong-signal and delayed
preview guards and mobile containment are checked. Quint has six rejecting
controls. `GOO_RECIPES_BOMBADIL=1` exercises source-free preview guards.

| Frame | User journey |
| --- | --- |
| [37a](screenshots/37a-shared-example-packages.png) | Choose a sample application package in the shared product workflow. |
| [37b](screenshots/37b-fixed-scope-recipe-preview.png) | Review prerequisites, retained population and limits before a source read. |
| [37c](screenshots/37c-recipe-operation-distribution.png) | Inspect complete operation duration groups from the recipe. |
| [37d](screenshots/37d-recipe-hypothesis-and-evidence.png) | Keep the hypothesis separate from an exact frozen evidence citation. |
| [37e](screenshots/37e-mobile-log-recipe.png) | Preview the same generic log recipe on a narrow screen. |

## 38. Shared resource and lifecycle measurements

`journeys/measurements.spec.js` generates ordinary good/retry/missing/rejected/late
reservation work, a Go checkout and a Go report. Independent ClickHouse queries
locate actual trace-correlated Go log records and check the gauge source count.
The user opens a declarative USE step without reading telemetry, then executes
fixed questions across all four measurement families. Mature lifecycle subjects
preserve missing→unknown; typed subject drilldown retains exact tenant/event scope.
A gauge capture refreshes without changing the question, reopens without source
execution, is cited in notes, and becomes a reusable saved package. Independent
Go lifecycle definitions exercise a different app/event vocabulary. Quint validates
the actual live ledger and rejects seven altered ledgers; Hegel runs measurement
state/swarm sequences. Optional `GOO_MEASUREMENTS_BOMBADIL=1` explores preview/edit
invariants without creating evidence. This is bounded local verification, not a
production delivery or completeness guarantee.

| Frame | User journey |
| --- | --- |
| [38a](screenshots/38a-guided-resource-investigation.png) | Choose a data-only USE flow and inspect its requirements and shared window. |
| [38b](screenshots/38b-resource-gauge-observations.png) | Inspect independent gauge samples with full stream identity. |
| [38c](screenshots/38c-native-duration-intervals.png) | Inspect native histogram intervals, bucket ranges and exact counts. |
| [38d](screenshots/38d-scoped-lifecycle-evidence.png) | Distinguish mature observed outcomes, pending work and missing evidence. |
| [38e](screenshots/38e-measurement-capture-in-notes.png) | Cite a frozen typed capture in the common investigation narrative. |
| [38f](screenshots/38f-mobile-measurement-editor.png) | Edit and validate a reusable definition on a narrow screen. |
| [38g](screenshots/38g-fixed-record-package-workflow.png) | Preview a record package without a read, then explicitly run its fixed workflow window. |


### 39 · Ordinary OTel metric discovery

An engineer starts with an independently instrumented application with no goo
manifest. They discover its `jobs.completed` instrument and dimensions, build and
save an editable package, run it, distinguish outcomes in stream labels, capture
it, reopen the retained evidence and cite it in notes.

- [Discover observed instrument metadata](screenshots/39a-metric-discovery.png)
- [Evaluate independent metric streams](screenshots/39b-discovered-measurement.png)
- Executable: `journeys/metric-discovery.spec.js`; optional exploration invariants:
  `GOO_METRIC_DISCOVERY_BOMBADIL=1`.

### 40 · One investigation loop

An engineer finds a reusable measurement in Definitions, previews it without a
source read, selects a service in inventory and carries its scope into metric
discovery. They abandon a delayed actual discovery response by switching modes,
confirm that it cannot publish stale results, then explicitly discover and run an
ordinary OTel metric package. The authoring surface shows canonical YAML and a
three-stage flow that fits the mobile viewport.

- [Shared searchable definitions](screenshots/40a-shared-definitions.png)
- [Question to evidence](screenshots/40b-question-to-evidence.png)
- [Mobile measurement flow](screenshots/40c-mobile-measurement-flow.png)
- Executable: `journeys/coherence.spec.js`.

### 41 · A cross-service, cross-signal investigation handoff

An engineer starts with a question before evidence exists. They gather checkout
and report spans, report logs and an independent OTel job counter. Capture-dialog
destinations attach evidence to the current draft or an exact saved revision.
The handoff preserves findings, uncertainty, time/population scope and reusable
metric identity. It reopens without telemetry reads; fresh metric execution creates
new evidence and a child investigation while the earlier handoff remains unchanged.

- [Question, findings and evidence](screenshots/41a-cross-signal-investigation.png)
- [Fresh reuse with preserved citations](screenshots/41b-reuse-with-preserved-evidence.png)
- [Mobile investigation handoff](screenshots/41c-mobile-investigation-handoff.png)
- Executable: `journeys/investigation-workspace.spec.js`; optional draft exploration:
  `GOO_INVESTIGATION_BOMBADIL=1`.

### 42 · From a slow operation to a supported finding

A real delayed operation raises competing error-path, delayed-work and resource
hypotheses. An engineer previews scope changes without a source read, compares
recorded duration/status in the original population, follows a reported trace and
its exact logs, then discovers worker instruments in the same fixed window. They
capture supporting evidence, record a missing-instrumentation question and save
a finding that distinguishes reproduced delay from production causal attribution.

- [Explicit preserved and changed scope](screenshots/42a-explicit-follow-up-scope.png)
- [Trace and exact log inspection](screenshots/42b-trace-and-exact-log-context.png)
- [Evidence gap and supported finding](screenshots/42c-evidence-gap-and-supported-finding.png)
- [Mobile investigation handoff](screenshots/42d-mobile-guided-investigation.png)
- Executable: `journeys/followups.spec.js`; optional source-free browser exploration:
  `GOO_FOLLOWUPS_BOMBADIL=1`.

Scope consolidation: journeys 37 and 42 now assert the shared checked transform
relationship and render computed population/grain changes. Their refreshed frames
include progressive disclosure for source binding and unbound OTel semantics.
The source reads remain explicit; this does not demonstrate automatic convention
profile resolution. See [the shared contract](../QUESTION_SCOPE.md).


### 43 · Compare HTTP route measures over time

Implemented in `journeys/grouped-time-series.spec.js`. A controlled ordinary OTLP
producer sends 124 server-shaped spans through the real collector/exporter into
ClickHouse. The engineer compares route-specific duration/status over fixed time
cells, distinguishes missing and empty routes, and searches locally without
changing chart scale or querying again. A 100-row page does not limit the measures.
The engineer captures the chart, derives/saves/reuses its package in a fresh window,
then reopens the original evidence without source reads. Typed planning identifies
span-record grain and raw-value percentile requirements. This does not certify
HTTP semantic-convention coverage, unique requests or production health.

![Fixed route/time comparison](screenshots/43a-http-route-time-comparison.png)
![Frozen grouped evidence](screenshots/43b-frozen-grouped-evidence.png)
![Mobile grouped evidence](screenshots/43c-mobile-grouped-evidence.png)

### 44 · Measure numeric attributes with explicit coverage

A database engineer selects a numeric attribute from ordinary OTel spans, declares
integer type and unit in Explore, and sees exact arithmetic beside missing, empty
and malformed-value counts. A controlled producer sends 128 spans and 128 logs
through OTLP and the real exporter; a 100-row inspection page cannot define totals.
The engineer explores route/time breakdowns locally, captures evidence, saves and
reuses its YAML in a fresh window, then reopens the original immutable capture.
Unsigned, signed and resource bindings, including overflow strings, are compared
to the Go reference for every total/group/cell. No manifests or convention/type
inference are required. Controlled data does not certify completeness or health.

- [Declared measure and coverage](screenshots/44a-numeric-coverage-and-series.png)
- [Frozen numeric evidence](screenshots/44b-frozen-numeric-evidence.png)
- [Mobile numeric evidence](screenshots/44c-mobile-numeric-evidence.png)
- Executable: `journeys/numeric-measures.spec.js`; optional local exploration:
  `GOO_NUMERIC_BOMBADIL=1`. The live Quint ledger reuses the existing six-stage
  custody model with numeric reference projections and eleven rejecting mutations.

### 45 · One question from entry to frozen evidence

An engineer arrives at a compact Explore entry, composes service/duration shortcuts
and an advanced filter without a source read, then runs once. A route attribute
narrows the fixed question in one gesture. Trace context and return preserve the
main result; returning issues no query. An edit supersedes a delayed real response,
which cannot publish stale results. From Services the engineer explicitly discovers
a native worker instrument in the inventory's incident window, then captures the
main record question and reopens frozen evidence without rereading telemetry.
The controlled producer uses ordinary OTLP spans and a gauge, not mocked evidence.

- [One primary entry](screenshots/45a-one-question-entry.png)
- [Applied scope and contextual narrowing](screenshots/45b-applied-scope-and-drilldown.png)
- [Metric discovery retains incident bounds](screenshots/45c-service-metric-window.png)
- [Mobile frozen evidence](screenshots/45d-mobile-frozen-scope.png)
- Executable: `journeys/question-workbench.spec.js`; optional source-free draft
  exploration: `GOO_QUESTION_BOMBADIL=1`. The observed Quint lifecycle validates
  generations/phases and rejects mutated publication, scope, read and custody facts.

### 46 · Browse a service and follow observed conventions

Run `journeys/service-browser.spec.js` through the normal journey wrapper. An
ordinary catalog service emits HTTP/database spans, a log, a process utilization
gauge and an unsupported nonmonotonic sum; another service emits only logs. Open
service detail, choose an observed HTTP breakdown, return locally, preview a native
measurement without execution, then inspect the log-only service. The journey also
delays a real metadata response and changes environment to check stale publication.

- [Service signals, questions and instruments](screenshots/46a-service-signals-questions-instruments.png)
- [Fixed service route question](screenshots/46b-fixed-service-route-question.png)
- [Native instrument preview](screenshots/46c-fixed-native-instrument-preview.png)
- [Mobile service detail](screenshots/46d-mobile-service-detail.png)

`GOO_SERVICE_BROWSER_BOMBADIL=1` enables the matching 30-second exploration campaign.
The actual ledger is validated by quint-go with six rejecting mutations. See
[service browsing](../SERVICE_BROWSING.md) for scope and semantic boundaries.

### 47 · Discover metric-only services

`journeys/metric-services.spec.js` ingests an ordinary metric-only worker with
gauge, counter and histogram points, plus an unsupported-only worker. Discover
services directly from native points, open the original window, preview a gauge
without reading telemetry, run it explicitly and return locally. Boundary fixtures
prove inclusion at the start and exclusion at the end using independent SQL.
Counts are records, not the histogram's 5000 observations or metric values.

- [All-signal catalog](screenshots/47a-metric-only-service-discovery.png)
- [Metric-only instrument detail](screenshots/47b-metric-service-instruments.png)
- [Fixed native execution](screenshots/47c-fixed-native-measurement.png)
- [Mobile metric service](screenshots/47d-mobile-metric-service.png)

`GOO_METRIC_SERVICES_BOMBADIL=1` enables its 30-second actor. The observed ledger
is validated by quint-go with seven rejecting mutations.

### 48 · One Services catalog

Executed by `journeys/service-browser.spec.js` alongside storyboard 46. Discover
once, see one row per service with distinct signal counts, search locally, follow
a service question, return to the retained detail and then the filtered catalog.
The journey checks no implicit return reads, empty-filter keyboard focus, mobile
containment and rejection of an intentionally tampered real response scope.

- [One catalog entry](screenshots/48a-one-services-catalog.png)
- [Retained local filter and return](screenshots/48b-local-filtered-return.png)
- [Mobile catalog](screenshots/48c-mobile-services-catalog.png)

See [the catalog contract](../UNIFIED_SERVICE_CATALOG.md) for capture scope and
remaining boundaries. Existing storyboard 31 now reaches operation inventory by
opening a service; saved inventory 32 retains its original environment-wide scope.

## 49 — Resume a fixed question

`question-links.spec.js`: reopen a question without evidence, explicitly Run question,
share with a second browser, traverse history without reads, reject malformed input
and inspect mobile layout. Frames: `49a-reopened-question.png`,
`49b-share-question.png`, `49c-mobile-resume.png`.

## 50 — Mixed investigation steps

`investigation-steps.spec.js` builds a named investigation with record and native
questions plus a UI-captured evidence reference, displays same/different windows,
saves a revision, reopens without reads and explicitly reruns. A second browser
opens the shared native question. Frames `50a-mixed-question-investigation.png`,
`50b-resume-native-question.png` and `50c-mobile-native-question.png` document it.

## 51 — Linked question evidence

`linked-evidence.spec.js` runs and captures a saved question twice, explains a
changed-filter capture, recovers from document publication failure without another
source read and reopens stored history on mobile. A companion case captures native
metrics with their stored read time and typed definition. Frames 51a–e show the
history, drift preview, recovery, mobile layout and native result summary.

### 52 — Compare stored evidence

`evidence-comparison.spec.js` exercises real OTLP HTTP spans and worker gauges:
52a shows latency/error changes between stored windows, 52b refuses a changed
selector, 52c checks mobile containment, and 52d preserves native stream identity
and side-by-side saved results. It also checks empty results, failed GET recovery,
no source reads and immutable receipts, with Quint and optional Bombadil coverage.

### 53 — Save and share a comparison

`saved-comparison.spec.js` saves a comparison and finding, shares the exact revision
link with a fresh browser, preserves older findings, recovers failed publication
and a changed-context receipt by ID, and saves native unit incompatibility.
53a shows the saved finding, 53b the recipient view, 53c mobile containment and 53d
the frozen incompatible decision. Ledger replay uses Quint with six rejecting
controls; the optional Bombadil actor reopens stored comparison cards.

Journey 53 also covers the shared stored comparison action group: inspect a
baseline, return keyboard focus, inject a saved-receipt outage and recover through
Comparison options without comparison-engine or telemetry execution. Saved
Bombadil actions open the actual options disclosure and assert hidden fixed-pair
selectors and no source reads.

### 54 — Compare complete native windows

`native-windows.spec.js` ingests standard OTLP sums and explicit histograms, then
compares complete windows and shares a versioned frozen receipt. Frames 54a–e show
counter rates, whole histogram populations, mobile containment, counter decrease
refusal and incompatible bucket layouts. Incomplete coverage is also refused.
Quint validates source-free transitions with six negative controls; optional
Bombadil exploration reopens the saved native result.

### 55 — Find and capture observed SDK windows

`window-discovery.spec.js` ingests real standard OTLP cumulative counters and delta
explicit histograms with 123 ns offsets. Broad captured questions cannot assert
complete coverage. Discovery reads stored intervals only, previews exclusions and
requires both selections before two explicit source reads create exact captures. Shared native previews show UTC bounds with exact nanosecond tooltips, readable durations, validated export coverage and exact-rate/histogram values. Exclusions explain missing coverage without claiming ingestion completeness.
It compares and shares frozen evidence with a fresh browser and checks mobile
containment. Original captures remain byte-for-byte unchanged. A forced second-capture failure
retains the first ID and leaves the original comparison selected.

- [Original support and exclusions](screenshots/55a-observed-support-and-exclusions.png)
- [Explicit exact sub-scope selection](screenshots/55b-explicit-exact-window-selection.png)
- [New captures compared](screenshots/55c-selected-scopes-captured-and-compared.png)
- [Mobile histogram support](screenshots/55d-mobile-histogram-support.png)
- [Completed first capture retained after second capture failure](screenshots/55e-partial-capture-recovery.png)

Quint observed-trace replay validates source-free discovery/selection and explicit
two-read capture, with six rejecting mutations. Optional Bombadil exploration
compares original captures, discovers support and toggles selections while checking supplied coverage timestamps.

### 56 — Reuse comparison questions

`reusable-questions.spec.js` captures real standard OTLP request counters with
123 ns offsets, then adds each saved comparison side as a separate reusable
question. It forces a publication failure and retries without recapture, shares
the revision with a fresh browser, restores exact scope without telemetry reads,
runs and captures a child revision and previews/runs a fresh window with the same
definition. Earlier revisions and evidence remain unchanged.

- [Questions with linked captured evidence](screenshots/56a-comparison-to-reusable-questions.png)
- [Restored exact question](screenshots/56b-restored-exact-question.png)
- [Rerun appends evidence history](screenshots/56c-rerun-linked-evidence-history.png)
- [Fresh window on mobile](screenshots/56d-mobile-fresh-window.png)

Quint replay includes six rejecting mutations. Optional Bombadil exploration adds
questions and reverts drafts, checking bounded cards and one linked capture per
new question.

### 57 — Resume with fresh evidence

`fresh-comparison.spec.js` captures real HTTP server spans and OTel memory gauges.
It checks baseline focus on entry and opener focus after closing through Preview
options, then previews fresh scope without telemetry reads, rejects changed-definition prior
evidence, selects a frozen baseline and performs one new capture. A forced
comparison failure is retried without recapture. The finding is saved and shared
with a fresh browser; original questions, evidence links and captures are unchanged.

- [Explicit fresh scope and prior evidence](screenshots/57a-explicit-fresh-scope-preview.png)
- [Fresh evidence against the frozen baseline](screenshots/57b-current-evidence-and-stored-baseline.png)
- [Shared follow-up finding](screenshots/57c-shared-followup-finding.png)
- [Native follow-up on mobile](screenshots/57d-mobile-native-followup.png)

Quint replays observed reads and saved evidence relationships with six rejecting
mutations. Optional Bombadil exploration opens/closes previews and selects prior
evidence without capturing, while checking one visible fixed capture/retry action.

## 58 — Locate a change by captured attribute

Executable: `journeys/comparison-breakdown.spec.js`. Real standard OTLP spans
create before/after populations of 1200/1300 records. Select HTTP route without
a source read, explicitly capture both windows, inspect missing/empty/new groups,
open a fixed-side question, save a finding and reopen it source-free on mobile.
Frames: 58a explicit recapture, 58b full population breakdown, 58c shared mobile.

- [58a — explicit breakdown recapture](screenshots/58a-explicit-breakdown-recapture.png)
- [58b — complete route populations](screenshots/58b-complete-route-change.png)
- [58c — frozen mobile finding](screenshots/58c-shared-mobile-breakdown.png)

## 59 — Automatic service overview and distributed relationships

Executable: `journeys/service-dashboards.spec.js`. Ordinary OTel spans, logs and
native gauges produce manifest-free service widgets and a resolved parent edge.
Switch to the distributed view without a read, open a node in the same scope,
stage its widget question, capture it in an immutable revision and reopen in a
fresh browser. An independent log-source failure retains span results. Registry
declarations remain separate from observed data. Frames 59a–d show relationships,
automatic widgets, mobile layout and partial-source availability.

- [59a — reported distributed relationships](screenshots/59a-distributed-reported-relationships.png)
- [59b — automatic service widgets](screenshots/59b-automatic-service-widgets.png)
- [59c — mobile service overview](screenshots/59c-mobile-service-overview.png)
- [59d — independent partial availability](screenshots/59d-partial-source-availability.png)

## 60 Namespace service identity

[Executable journey](../../journeys/service-namespaces.spec.js) injects four
same-named OTel services: missing, empty, payments and orders namespaces. It
checks independent inventory/native discovery isolation, a reported parent edge
between same-named services in different namespaces, and a fixed payments
question. A registered operation package produces a question matching its
instance, instrumentation scope/version and operation name/role; a foreign role
is excluded. Schema URL remains declared metadata. The observed ledger is replayed
through Quint with namespace, count and window mutations required to fail.
Optional `GOO_NAMESPACE_BOMBADIL=1` explores catalog/distributed/service transitions
and checks namespace-preserving widget questions and uncaught exceptions.

![Separate service identities](screenshots/60a-namespace-service-identities.png)
![Fixed namespace overview](screenshots/60b-namespace-service-dashboard.png)
![Declared operation question](screenshots/60c-declared-operation-question.png)
![Mobile contract card](screenshots/60d-mobile-declared-operation-question.png)

## 61 Declared services without observations

[Executable journey](../../journeys/declared-services.spec.js) opens a real stored
historical inventory as a fixed-window Services question. It discovers registered
application bindings with no telemetry, sees them separately in the distributed
view, opens the unified service overview, and explicitly explores/captures an
empty span question. An explicitly selected namespace remains in the next captured
question. A recipient reopens that receipt without a source read.
Independent registry failure remains visible, and a late registry response is
rejected after a scope edit. The observed ledger is replayed through Quint with
five mutations; optional `GOO_DECLARED_BOMBADIL=1` explores declaration disclosures
and catalog/service transitions without inventing observed widgets or edges.

![Declared services and explicit absence](screenshots/61a-declared-services-without-observations.png)
![Unified empty service overview](screenshots/61b-declared-service-empty-overview.png)
![Mobile declared service](screenshots/61c-mobile-declared-service.png)
![Stored empty service question](screenshots/61d-empty-service-evidence.png)

## 62 Investigate an exact reported edge

[Executable journey](../../journeys/reported-edges.spec.js) injects 130 child span
records plus duplicate, ambiguous, internal, wrong-trace, wrong-environment and
out-of-window parent cases. It stages the edge without a read, runs full latency
and recorded-error summaries beyond the 100-record page, narrows to errors while
retaining the parent universe, expands broader trace context, and saves/reopens
an investigation without rereading telemetry. A separate namespace-filtered case
proves that parent ownership keeps the original discovery namespace scope.
Quint replays the observed ledger with five rejecting mutations; optional
`GOO_EDGE_BOMBADIL=1` explores reruns, error filters and evidence disclosures.

![Staged edge question](screenshots/62a-staged-reported-edge.png)
![Exact child population and full summaries](screenshots/62b-exact-edge-latency-and-errors.png)
![Stored mobile edge evidence](screenshots/62c-mobile-stored-edge.png)

Comparison storyboards 35 and the stored/native comparison journeys now show
[aligned visual comparisons](../VISUAL_COMPARISONS.md), with full-population values
and explicit unavailable states rather than a wall of prose.

## 63 Framework-neutral display components

[Executable component conformance tests](../../journeys/components.spec.js) open
`/ui-components`, a standalone reference containing explicitly synthetic fixtures.
No app runtime or telemetry API is loaded. Server HTML upgrades to native custom
elements; SPA properties remain reactive even when assigned before definition.
Tests cover detached inputs, exact large values, zero vs absence, malformed JSON,
text safety, keyboard disclosures, shared geometry and mobile containment.
These fixture frames explain the library; real OTLP-backed journeys 35/52/54/58/59/62
verify its adoption in application workflows.

![Standalone component reference: synthetic fixtures](screenshots/63a-framework-neutral-display-components.png)
![Mobile component reference: synthetic fixtures](screenshots/63b-mobile-display-components.png)

Journey 64 (`components.spec.js`) checks caller-owned table rows, focus retention
during appends, reconnects, explicit state transitions and keyboard overflow on
mobile without API reads. These are synthetic component conformance fixtures.
Journey 59 covers adoption in real service observations and partial source failure.

Journey 65 (`components.spec.js`) exercises synthetic SSR/SPA facet and typed-value
contracts without source reads. Journey 29 now verifies shared elements across
real attribute/facet narrowing and trace/log inspection; optional
`GOO_INSPECTION_BOMBADIL=1` explores that isolated fixed question. Journey 34
checks full-population/saved facet scales with an updated Bombadil invariant.

Journey 66 follows an applied log question into real OTel live observations,
late duplicates, bounded eviction, focus pause, resumed/reconnected gaps and scope
cancellation, with Quint ledger mutations and an opt-in Bombadil campaign.
Journeys 67 and 68 are clearly synthetic standalone waterfall/time-bar contracts;
actual app adoption is exercised by journeys 29, 30, 33, 45 and 59.

## 69. A frozen baseline beside a rolling record population

`journeys/live-comparison.spec.js` sends real OTel spans for a generic search
service. Select linked baseline evidence, explicitly start a rolling comparison,
pause, add observations, and rescale when the population exceeds the fixed axis.
A failed refresh retains the old window as stale. Another source record arrives
before freezing: the capture retains the seven displayed records, not the eight
now in the source. Reopen evidence and verify the baseline is unchanged. The
optional Bombadil actor exercises the live controls; the observed ledger runs
through Quint with negative controls.

![Frozen baseline and rolling count](screenshots/69a-frozen-baseline-rolling-count.png)
![Stale read retains its axis and original window](screenshots/69b-stale-read-stable-axis.png)
![Displayed evidence frozen without another source read](screenshots/69c-exact-displayed-evidence-frozen.png)

![Frozen result linked to an immutable investigation revision](screenshots/69d-frozen-result-linked-to-investigation.png)
![Mobile rolling comparison](screenshots/69e-mobile-live-comparison.png)

## 70. Passive chart frame in server HTML and SPA updates

`journeys/chart-frame-components.spec.js` demonstrates safe metadata, caller-owned
native actions and disclosures, pre-upgrade properties, contained mobile layout,
and focus preservation when state/window metadata changes. Synthetic standalone
HTML tests the reusable display contract; journey 69 supplies real-source product
coverage.

![Passive chart frame](screenshots/70a-passive-chart-frame.png)
![Mobile stale chart frame](screenshots/70b-mobile-stale-chart-frame.png)

## 71. A service regression across count, latency and recorded error share

`journeys/live-span-measures.spec.js` ingests real OTel spans and captures a
question requesting duration and recorded status. Two baseline spans have 1 ms
latency and one recorded error. Four current spans have 10 ms latency and one
recorded error. The shared frame shows both compact percentile rails on the same
fixed domain, plus Error/total fractions on a fixed 0–100% axis.

Add three 100 ms errors: count grows to seven, p95/p99 grows to 100 ms, and Error
share becomes 4/7 (57.14%). Explicitly rescale count and latency. A source failure
retains all summaries, window and basis. After another 1,000 ms error arrives,
freeze the displayed seven spans and their 340 ms duration sum/four errors into
an investigation revision. Old source responses cannot populate a reopened panel.
Quint rejects ten corrupted ledger observations; Bombadil checks exact percentile
positions, per-measure domains and required start/freeze/link interactions.

![One live window for count, latency and recorded error share](screenshots/71a-frozen-baseline-rolling-count.png)
![Stale evidence retains all measures together](screenshots/71b-stale-read-stable-axis.png)
![The complete displayed result is frozen](screenshots/71c-exact-displayed-evidence-frozen.png)
![Frozen service regression evidence in the investigation](screenshots/71d-frozen-result-linked-to-investigation.png)
![Mobile shared-domain comparison](screenshots/71e-mobile-live-comparison.png)

A separate real-source check freezes an empty one-millisecond rolling window.
No latency markers or error-share bars appear, and freezing keeps the result
unavailable rather than inventing zero latency or zero percent errors.

![Empty rolling window withholds numeric comparison](screenshots/71f-empty-live-window.png)

## 72. Follow a native counter stream without hiding export gaps

Run `journeys/live-counter.spec.js`. From a saved generic HTTP counter question,
choose one frozen OTel stream baseline (3/s), then start live comparison (5/s).
A newer incomplete export window withholds rates. Late coverage restores 15/s;
the original domain remains fixed until Rescale. A failed read preserves the
same display and basis. A later export arrives before Freeze, but the saved and
reopened evidence retains the displayed window. Explicitly add that capture to
the investigation. The original question and baseline remain unchanged.

![Exact native rates on one domain](screenshots/72a-native-rate-shared-domain.png)
![Latest export gap withholds rates](screenshots/72b-latest-export-gap-withheld.png)
![Late coverage repaired and explicitly rescaled](screenshots/72c-late-coverage-restored.png)
![Displayed native window frozen](screenshots/72d-displayed-native-rate-frozen.png)
![Native rate comparison on mobile](screenshots/72e-mobile-native-rate.png)

## 73. Focus one native stream from a shared instrument

Run `journeys/exact-stream.spec.js`. A generic HTTP instrument has two baseline
instances. Inspect their independent streams, focus instance one, and preview
without a source read. Run the exact question, share it, save the package revision
and capture the baseline. Seven current identities include near matches with
missing/added/empty attributes and changed resource schema. Live comparison shows
only the pinned stream's 5/s against its frozen 3/s baseline. Freeze and reopen
retain the exact binding. Explore all matching streams returns to the original
broad question with an explicit preview and Run; prior evidence remains unchanged.

![Independent native streams](screenshots/73a-discover-independent-streams.png)
![Read-free exact-stream preview](screenshots/73b-exact-stream-preview.png)
![Captured selected identity](screenshots/73c-selected-stream-evidence.png)
![Pinned live native rate](screenshots/73d-pinned-live-rate.png)

![Selected stream on mobile with horizontally scrollable observations](screenshots/73e-mobile-selected-stream.png)

## 74. Find a native stream and understand unavailable coverage

Run `journeys/native-navigation.spec.js`. Open saved evidence containing 28 HTTP
counter streams. Page past the first 25, select Needs review, see an explicit
empty search and find a stream by a resource value. Its retained export crosses
the fixed boundary: readable guidance explains why it cannot be split into an
increment, with exact diagnostic codes preserved. Clear/search retains the open
stream disclosure and original question/basis without reading telemetry. Focus
previews its exact typed identity. Reopening the capture is unchanged.

![Bounded retained stream inventory](screenshots/74a-bounded-stream-inventory.png)
![Boundary coverage explained after identity search](screenshots/74b-explained-boundary-coverage.png)
![Native review controls and scrollable observations on mobile](screenshots/74c-mobile-stream-review.png)

## 75. From a native service overview to exact-stream evidence

Run `journeys/service-native.spec.js`. Discover an ordinary OTel HTTP service in
namespace payments without a Goo manifest. Two native widgets load automatically:
six independent request-counter streams and a native request-duration histogram.
The newest absent interval withholds a value; another stream shows a valid latest
5/s interval alongside incomplete whole-window coverage. Histogram p95 stays the
0.5 … +∞ s bucket range. A same-named service in another namespace is excluded.
Expand/search the retained six streams without a telemetry read, then open one
exact stream question. Preview preserves scope/identity; Run and Capture are
explicit reads. Reopen retains immutable evidence. A runtime service separately
shows a fresh gauge observation. The mobile compact table scrolls within its owner.

![Native service values and separate window coverage](screenshots/75a-native-service-overview.png)
![Retained service stream browsing without a read](screenshots/75b-retained-service-streams.png)
![Compact native service card on mobile](screenshots/75c-mobile-native-service-card.png)

## 76. Choose the native measurements in a service overview

An operator opens a discovered payments API with multiple OTel instruments,
including same-name version/schema variants. They stage non-default instruments
without reading telemetry, see the two-card limit, then explicitly reread the
same fixed service window. They capture one selected question and reopen that
evidence without source reads, clear native cards, and restore defaults. A later
empty Apply supersedes a held real source response. Selections last for this visit.

Executable: `journeys/service-native-selection.spec.js`. Its real OTLP fixtures,
Bombadil actions and observed Quint replay cover selection and evidence custody;
Hegel generates independent planner selection sequences.

![Staged choices and the native budget](screenshots/76a-staged-instrument-selection.png)
![Selected native service measurements](screenshots/76b-selected-service-overview.png)
![Mobile instrument picker](screenshots/76c-mobile-instrument-selection.png)

## 77. Inspect native observations over time

An operator opens ordinary OTel gauge, counter and histogram questions for a
payments API. They inspect signed/stale/zero gauge samples, a counter reset and
unknown intervals, then histogram intervals with an uncovered gap, zero
observations and changed native bucket bounds. Selection is local and read-free.
They capture the unchanged histogram question and reopen its stored timeline
without source reads. Unknown selection never reuses a prior distribution.

Executable: `journeys/native-timeline.spec.js`; actual OTLP, observed Quint replay
with nine rejecting controls, and saved-evidence Bombadil inspection.

![Independent gauge samples and latest unknown](screenshots/77a-gauge-native-timeline.png)
![Counter interval values and typed unknowns](screenshots/77b-counter-unknown-intervals.png)
![Selected native histogram interval and its bounds](screenshots/77c-histogram-interval-bounds.png)
![Mobile native interval and bucket inspection](screenshots/77d-mobile-native-timeline.png)

## 78. Reuse the passive timeline in server HTML or an SPA

The labelled synthetic component catalog demonstrates supplied observations and
controlled selection. Executable `journeys/value-timeline-components.spec.js`
checks keyboard focus, exact zero/unknown values, detached property upgrade,
getter detachment, reactive empty results and mobile composition without API reads.
This component contract complements, rather than replaces, journey77's real data.

![Passive supplied timeline at mobile width](screenshots/78a-passive-scalar-timeline.png)

## Native bucket heatmap journeys

Journey 77 now covers layout switching, recorded zero, unknown selection and mobile overflow over real OTLP evidence; its 77c/77d screenshots show the grid in context. Journey 54 shows the compatible-window count grid and rejects a grid for incompatible layouts. Journeys 79/80 document the passive component using a clearly synthetic fixture (`79-bucket-component-contract.png`).

## 81 · Explicit histogram time buckets

`journeys/histogram-time.spec.js` injects real OTLP exports, validates an explicit width, merges complete intervals, withholds crossings/layout changes, distinguishes zero/unknown, switches views without reads and reopens captured evidence. [Storyboard and frames](../HISTOGRAM_TIME_BUCKETS.md). Screenshots: `81a-go-time-bucket-aggregation.png` and `81b-mobile-time-bucket-unknown.png`.

## 82/83 · Shared evidence summaries

Journey82 exercises real OTLP capture → exact evidence summary → linked investigation, preserving unsaved findings and retained receipt identity without source reads. Journey83 covers synthetic standalone SSR/properties/focus and mobile layout. [Storyboard and frames](../EVIDENCE_SUMMARIES.md).

Journeys59/75 now document compact shared service summaries and readable UTC time bounds with exact-value tooltips; journey83 covers formatter precision and keyboard inspection. Journeys82/77 refresh the affected captured-evidence and selected-interval frames. Details: [evidence-summary storyboard](../EVIDENCE_SUMMARIES.md#compact-service-cards-and-readable-time).

Journey63 now checks the collapsed latency readout stays under 100px in its standalone desktop fixture, with the positional rail revealed through keyboard disclosure. Its desktop/mobile reference frames and journey59’s real service frames show the inline Mean/p50/P95/P99 layout.

See [comparison presentation](../COMPARISON_PRESENTATION.md) for aligned scalar pairs, compact metric matrices, frozen-side summaries and UTC tooltip adoption.

Journey84 documents the standalone selected-item shell; journey35 exercises real log/span inspection and journey77 exercises native selection replay. [Inspector storyboard](../SELECTED_ITEM_INSPECTORS.md).

## Core experience refinements

See [the implemented workflow](../CORE_EXPERIENCE_REFINEMENTS.md) for retained
application visits, typed population selection, compact diagnosis, trace
breakdowns and independently loaded context. These compose passive display
elements; application adapters own navigation, read lifecycle and scope. Journey
85 exercises real OTLP data and refreshed screenshots; the core-refinements
Bombadil actor covers selection and returning to retained evidence. Hegel
selection properties and the question-session state swarm complement observed
quint-go validation against the existing question-links model.

Journey **91** (`population-workspace.spec.js`) explores discovered attributes,
compares two service values using a pinned baseline and explicit filter reset,
saves/reopens/replays the configuration, checks an empty preceding window and
missing versus empty attributes, and captures desktop/mobile frames. Enable
`GOO_BREAKDOWN_BOMBADIL=1` for ranking and preceding-window actor actions.

Journey **92**, `population-live.spec.js`, watches ordinary spans/logs, investigates
reported trace context while paused, prepares full-population baseline comparison,
saves/reopens without subscriptions, recovers with a continuity gap, changes signal
and checks mobile composition. Enable `GOO_POPULATION_LIVE_BOMBADIL=1` for actor
start/pause actions. The observed ledger counts subscription/read initiations,
not the Go tail's internal polling reads.

## 93. Enter through recent activity

[Executable journey](../../journeys/recent-entry.spec.js) and
[contract/storyboard](../RECENT_ACTIVITY.md): ordinary OTel multi-service data,
explicit direct recent reads, local loaded-signal switching, service facets,
source-free range preview and saved configuration reopening, mobile composition,
log-only/no-data/partial sources and stale-response rejection. Quint checks the
observed request ledger with rejecting controls; optional Bombadil actions run
recent activity and switch loaded record pages without starting a tail.

## 94–95. Brush a population and retain the parent

[Executable journeys](../../journeys/range-drilldown.spec.js) and
[contract/sequence/storyboards](../RANGE_DRILLDOWNS.md): real generic OTLP clusters,
reverse brush and Escape, source-free child staging, independent fixed reads,
retained parent baseline/record inspector/tail rows/focus/scroll, nested returns,
late-response rejection, keyboard handles and deliberate mobile touch selection.
Quint/quint-go replays observed depth/read transitions with seven rejecting
controls. Bombadil runs/stages/returns without an implicit subscription. Journey
95 separately checks the synthetic server-rendered component reference, detached
property updates and native point/keyboard compatibility without API calls.

## 96. Stage categorical populations

[Contract, sequence diagram and generated frames](../FACET_SELECTION.md).
`journeys/facet-selection.spec.js` ingests real OTLP spans with missing, empty,
west and east region attributes. Include/Exclude, Undo/Clear and child staging
perform no source reads; explicit Apply/Run yields the expected populations.
Back preserves the original baseline and draft; Compare uses Go kernels; saving
and reopening restores typed configuration without reading. The journey also
checks mobile layout and combining fields across an explicit breakdown read.
Hegel checks construction against Go matching/compilation, Quint replays the
observed depth/read ledger with five rejecting controls, and optional
`GOO_FACET_BOMBADIL=1` exercises the categorical actor.

## 97. Derive a population from captured records

[Contract, process diagram and generated storyboards](../RECORD_POPULATIONS.md).
`journeys/record-population.spec.js` verifies inspector Include/Exclude, shared
fields across two selected observations, exact session provenance, retained
parent navigation, baseline comparison, 20-of-22 selection bounds, refresh reset,
stale-action rejection and mobile layout over real OTLP data. Optional
`GOO_RECORD_BOMBADIL=1` runs selection/shared-field/open/return/clear actions.
The related [training-label proposal](../TRAINING_ANNOTATIONS_DESIGN.md) has
conceptual training journeys only; shared tags/notes and optional saved-evidence retention are covered by journeys 98–99. Training export remains proposed.

## 98. Authored context across the workbench

Implemented: [executable journey](../../journeys/annotations.spec.js),
[contract](../WORKBENCH_ANNOTATIONS.md). Discover a real OTel-emitting scoped service,
add a tag/note, cancel another draft, and use the same annotation panel on a saved
exploration, narrative revision and telemetry package. Find all four by exact tag.
Correct a service assertion in parallel, withdraw one branch, and inspect immutable
history. Annotation interactions perform no additional source reads. The observed
ledger is checked with quint-go and injected negative controls; Bombadil explores
local draft/cancel/history actions. Rich training labels remain proposed.

![Scoped service annotation](screenshots/98a-scoped-service-annotation.png)
![Cross-product exact-tag catalog](screenshots/98b-cross-product-annotations.png)
![Corrections, parallel branches and withdrawal history](screenshots/98c-annotation-branches-history.png)
![Compact mobile annotation history](screenshots/98d-mobile-annotation-history.png)

## 99. Retain selected evidence in XTDB

Optional experiment: [executable journey](../../journeys/retained-evidence.spec.js),
[setup and limitations](../WORKBENCH_ANNOTATIONS.md#optional-xtdb-retained-evidence-experiment).
Emit ordinary OTLP records and explicitly save a fresh fixed-query capture. Select
one stored observation, annotate it and copy it to XTDB. Lose the successful save
response and retry with the same request key; the original annotation is recovered.
Open attributes, correct the review and reload the retained evidence with browser
source API requests blocked. This verifies source-free UI reopening, not physical
shutdown of ClickHouse or a completed native external-source connector.

![Retain a selected saved observation](screenshots/99a-retain-selected-observation.png)
![Evidence retained in XTDB](screenshots/99b-xtdb-retained-evidence.png)
![Review correction history](screenshots/99c-retained-evidence-review-history.png)
![Mobile retained evidence](screenshots/99d-mobile-retained-evidence.png)


## 100. Retain the observation selected from live activity

Implemented: [executable journey](../../journeys/live-retention.spec.js), optional
XTDB overlay required. Emit 102 ordinary OTel spans and open recent activity.
Select one displayed row, then ingest a later span: the selected page stays fixed.
Explicitly retain the selected observation with its original token/question/basis;
no save-capture step or source reread occurs. Lose the successful native response,
retry the same intent, and recover the original annotation. Repeat that response
loss for a correction and verify exactly two review revisions.

Reload with a committed await-token lower bound while browser telemetry APIs are
blocked. Inspect the exact retained attributes and compact mobile view. Returning
to activity shows the late span on a new read; pagination and refresh clear row
selection. ClickHouse remains running and no source records are deleted. Root Go
clock tests separately check expired pages/capacity; native tests recover a committed
request without its ephemeral page. The observed Quint ledger verifies original
membership and rejects token/scope/member mutations in addition to the existing
hidden-read/publication/history controls and a duplicated retry publication. Bombadil covers both review history and
live selection → unpublished draft → clear, observing native POST count.

![Select an observation from live activity](screenshots/100a-live-page-retention.png)
![Retained live observation and explicit provenance](screenshots/100b-frozen-live-observation.png)
![Idempotent review revision history](screenshots/100c-idempotent-review-history.png)
![Compact mobile retained observation](screenshots/100d-mobile-live-evidence.png)


## 101. Find retained evidence, recover interrupted saves and pin a review

Implemented: [executable journey](../../journeys/evidence-library.spec.js), optional
XTDB required. Emit three ordinary OTel spans, retain a selected observation and
lose the actual committed response. Reload; the pending browser intent survives.
Retry explicitly and recover the original annotation. Start an investigation with
that exact review, save it, then correct the review and repeat response-loss/reload
recovery. The document keeps its original interpretation.

Search the retained catalog by current tag, environment, exact recorded service
name and capture-time bounds. Superseded tags no longer match. Reopen the document
and inspect its pinned evidence/review on mobile. Browser telemetry requests remain
blocked throughout retention recovery and citation; ClickHouse stays running.
Hegel exercises the actual journal; the observed Quint ledger rejects eight
mutations. Bombadil explores readonly catalog, pending and citation controls.

![Recover a pending native request after reload](screenshots/101a-recovered-pending-save.png)
![Pin the original review in an investigation](screenshots/101b-pinned-review-citation.png)
![Search current retained evidence metadata](screenshots/101c-retained-evidence-catalog.png)
![Reopen the original review on mobile](screenshots/101d-mobile-pinned-review.png)
![Browse the retained catalog on mobile](screenshots/101e-mobile-retained-catalog.png)


The recent-activity journey 93 also checks compact desktop/mobile first-record
positions, absence of duplicate facet-workspace summaries, and source-free facet
disclosures. Its timeline opens range controls when a chart preview is selected.

### 112 — Inspect investigation revisions before opening

[Executable storyboard](../../journeys/document-library.spec.js): create same-title
unrelated roots and a preserved child revision through the real API; follow server
cursors and search within each page; inspect the exact child and parent without
replacing a pending editor note. Explicit Open refuses the pending draft. Filtering
retains the selected detail. Revert then Open loads the exact child without telemetry
execution. [Mobile frame](screenshots/112a-investigation-revision-inspection.png).
Journey 36 retains the deployed-app evidence path and the Quint ledger with rejecting
controls; its optional Bombadil actor now explores revision browse/inspect/close as
well as draft edits, with bounded rows and exact document selection invariants.


Journey [101](../../journeys/evidence-library.spec.js) now selects compact retained
catalog metadata before explicit opening, rejects an older real response after a
filter change and preserves the selected snapshot when the optional store fails.
Selection issues no manifest/review GET or source/publication request. It retains
the real pending-save/reload/retry and pinned-old-review narrative path, including
Quint rejecting controls and the expanded Bombadil catalog select/close actions.
[Catalog](screenshots/101c-retained-evidence-catalog.png) ·
[Mobile catalog](screenshots/101e-mobile-retained-catalog.png).


[Journey 113](../../journeys/field-presentation.spec.js) stages omitted-empty,
missing and literal-sentinel filters without execution, checks a real Go package's
qualified field handoff and preserves the fixed draft across navigation.
[Mobile qualified filter chips](screenshots/113a-qualified-empty-presence-filters.png).
Journey 104 continues to exercise real ingested OTel data, source-free staged
selectors, fixed authoring and paused return with its Quint/Bombadil campaigns.


Journey [60](../../journeys/service-namespaces.spec.js) now covers six namespace
populations using real OTel spans and gauges, including literal `missing` and
`present empty` namespaces alongside missing-key and empty-string states. It
checks distinct catalog labels, scoped native values, record question links and
declared operation bindings, plus the existing observed Quint/rejecting controls
and expanded Bombadil label invariant. [Namespace catalog](screenshots/60a-namespace-service-identities.png).


[Journey 114](../../journeys/question-protocol.spec.js) captures a real ordinary
OTel gauge with exact native bounds, links its Go-canonical receipt to an authored
draft with omitted native defaults and verifies no false definition drift. Saving
preserves exact strings; fresh preview accepts the baseline without source reads.
[Mobile native wire/default scope](screenshots/114a-native-wire-defaults-and-exact-scope.png).
Its Bombadil actor permits only source-free preview/close; existing mixed-step,
linked-evidence and fresh-comparison journeys retain their observed model checks.

## 115. Continue from Activity into fixed analysis

[Executable journey](../../journeys/activity-analysis.spec.js) ingests ordinary
OTel logs, stages a filter, proves scope Analyze requires Apply, and then stages
the exact last completed window. Inspecting/editing the staged question performs
no telemetry read or publication. Run reads the fixed selector once; Return
restores the paused buffer and focus without resubscribing. An empty span window
also stages without inventing bounds from rows.

[Compact staged analysis](screenshots/115a-staged-activity-analysis.png) ·
[Fixed analysis result](screenshots/115b-fixed-activity-analysis.png).
The [contract](../ACTIVITY_ANALYSIS.md) distinguishes this applied-scope action
from the inspector's original-observation-window action. The bounded
[Bombadil actor](../../verification/bombadil/activity-analysis.ts) stages,
discloses editing and returns; it never executes or publishes a question.

## 116. Compare a staged population to its current population

[Executable journey](../../journeys/population-context.spec.js) emits four ordinary
OTel spans with missing, empty, west and east region values. It inspects each
signal's applied selector without reads, stages missing-or-empty values, and
compares the two-record selection with the full four-record baseline despite a
two-record inspection page. The companion's predicate is preserved; each configured signal is read once,
with no extra baseline read. The comparison opens
visibly on mobile. Restore starting population explicitly rereads the exact
original fixed questions.

[Selection actions](screenshots/116a-population-selection-actions.png) ·
[Comparison](screenshots/116b-selected-population-comparison.png).
The [population contract](../POPULATION_WORKSPACE.md#contextual-selection-comparison-and-restoration)
keeps this original selection-versus-current workflow separate from the later
[bounded categorical selection-versus-rest](../POPULATION_PARTITIONS.md); general
Boolean complements remain outside that contract.

## 117. Browse saved work and recover context without rerunning telemetry

[Executable journey](../../journeys/saved-workspace.spec.js) ingests a real log,
preserves it in a capture and links that capture from a narrative revision.
Browse investigations previews catalog metadata, then the exact narrative opens.
An unsaved finding survives opening the captured evidence and returning to notes.
Raw-ID fields remain secondary; Tools discloses annotation search and recovery.
A valid browser-local retry fixture produces a notice without a write, and an
unreadable journal remains discoverable. A malformed operation ID reports its
error inside Tools instead of hiding it with the selected-case area.

[Selected saved evidence](screenshots/117a-selected-saved-evidence.png) ·
[Contextual recovery](screenshots/117b-contextual-pending-recovery.png).
The [Saved composition contract](../SAVED_WORKSPACE.md) and
[bounded Bombadil actor](../../verification/bombadil/saved-workspace.ts) keep
browsing and disclosure separate from execution/publication.

Journey 118 (`journeys/record-columns.spec.js`) covers live observed field
selection, fixed-window layout handoff, staged field predicates and saved
population layouts. Frames 118a–c show live, saved and mobile records. Optional
`GOO_RECORD_COLUMNS_BOMBADIL=1` runs bounded local layout exploration.

Frame 118d and journey 118 also cover reopened saved record-field exclusion,
missing/empty preservation, selected-child staging/Run/return and observed
Quint trace validation. The field Bombadil campaign additionally stages an
exclusion after local column changes and checks that no source read follows.

Journey 119 (`journeys/log-comparison.spec.js`) covers ordinary log comparisons,
full groups versus bounded row pages, missing/empty groups, source-free stored
comparison and mobile focused-question access. Frames 119a–c document live,
stored and mobile surfaces. Set `GOO_LOG_COMPARISON_BOMBADIL=1` for its bounded
stored-comparison campaign.

Journey 120 (`comparison-ranking.spec.js`) ingests ordinary spans and demonstrates
low-share error/latency regression discovery, saved error/count ranking reopening,
full-population support independent of row limits, source-free stored ranking and
mobile access. Frames 120a/b show live error and stored mean ranking. Enable
`GOO_RANKING_BOMBADIL=1` for local rank-choice exploration.

Journey 121: [raw-span time-duration selection](../HEATMAP_SELECTION.md) stages
by pointer and keyboard, explicitly executes exact bounds and captures/reopens
selected evidence. 121b covers the passive component contract.

Journey 122: [explicit population time expansion](../POPULATION_WORKSPACE.md#expanding-the-time-window), source-free preview/open/return, independent selectors and explicit full-population reread.

Journey 123: [selected versus rest](../POPULATION_PARTITIONS.md), missing/empty and qualified selectors, full counts with one-row pages, parent return, saved views and captures, and an empty remainder.

Journey 124 checks the mobile selection summary is unobscured and its comparison action is keyboard-focusable below the wrapped investigation trail.

Journey 125: [whole-condition population broadening](../POPULATION_BROADENING.md), missing/empty OR groups, draft blocking, mobile/keyboard staging, full membership beyond one-row pages, source-free parent return and saved-view reopen.

Journey 126: [passive comparison context](../COMPARISON_CONTEXT.md), server HTML and early/detached SPA properties, exact clocks/counts, unknown transitions, source-free disclosure, keyboard focus and mobile layout. Existing journeys 120, 52, 53, 57, 69 and 72 exercise its real fixed/stored/fresh/live/native adoption.

Journey 127: [contextual comparison actions](../COMPARISON_ACTIONS.md), single-signal inspection of either side, retained comparison/ranking/focus, missing versus empty, shared stored record inspection and obsolete-response rejection.


Journey 128: [retained activity comparison](../COMPARISON_SERIES.md), keyboard/brush selection, exact side window, explicit child read, retained return and stored/mobile continuation.

Journey 69 now covers live bucket refresh before Freeze, retained alignment/selection, explicit child Run and source-free Back. Frames 129a/129b show the live series and retained selection; see [contract](../LIVE_COMPARISON.md#retained-bucket-series-during-refresh).

## D1 comparison acceptance audit

The [acceptance map](../D1_ACCEPTANCE.md) reconciles the comparison storyboards
with the completed bounded milestone. Journeys 69 and 71 now close previews through
Preview options and check focus before reopening, then reject held obsolete source
responses. Native-window Bombadil Reopen now uses Comparison options and requires
another observed immutable receipt read with no telemetry/comparison-engine call.
The native-question actor also uses the actual preview menu.

## 130 — Discover native instrument capabilities

`instrument-capabilities.spec.js` starts with manifest-free, language-neutral OTLP
HTTP histograms and custom sums. Discover is explicit; metadata disclosures remain
local; the exact seconds binding previews without evaluation, then Run reads once
in the fixed service window. Mismatched units and non-monotonic sums stay honest.
`instrument-catalog-components.spec.js` checks local 25-row paging and exact Build
intent without API requests. Quint replays actual discovery/result bindings with
rejecting controls; the configured Bombadil actor opens metadata without evaluation.
See [contract and ownership diagram](../NATIVE_INSTRUMENT_DISCOVERY.md).

- [Observed HTTP and generic capabilities](screenshots/130a-observed-instrument-capabilities.png)
- [Selected HTTP metadata hints](screenshots/130c-http-metadata-hints.png)
- [Contained mobile catalog](screenshots/130b-instrument-catalog-mobile.png)

## 131 — Inspect observed stream presence

`instrument-coverage.spec.js` emits three native HTTP streams with present, empty
and missing method/status keys, duplicate observations and a missing instance ID.
Discovery supplies the typed Go plan; local disclosures read or publish nothing.
An independent ClickHouse identity query supplies the Quint oracle, with seven
rejecting mutations. Missing metadata and truncated stream lists do not alter
summary totals or become verified coverage.

- [Observed stream presence](screenshots/131a-observed-stream-presence.png)
- [Mobile stream inspection](screenshots/131b-stream-presence-mobile.png)

132. **Manifest-free messaging processing discovery** —
`messaging-profile.spec.js` emits real OTLP and inspects an experimental pinned
processing-duration candidate, conditional label presence, missing/empty identity,
and desktop/mobile composition. Observed Quint replay uses independent source
identities and rejecting controls. The shared Bombadil discovery actor exercises
this consumer without implicit execution/publication. See
[the contract and frames](../NATIVE_INSTRUMENT_DISCOVERY.md#messaging-processing-candidate).


133. **Candidate versus evaluated evidence** — `instrument-acceptance.spec.js`
emits manifest-free OTLP across ordinary consumers and a custom counter. Explicit
Preview/Run keeps complete processing windows observed, incomplete coverage
unknown, and refuses missing identity or unsupported precision. Refusals clear the
result and disable capture; metadata remains not-run. The existing Quint discovery
replay validates observed/unknown results and rejecting controls; Bombadil
inspects an unidentified candidate without executing it. Four real frames and the
[acceptance map](../E1_ACCEPTANCE.md) document the boundaries.


134. **Time/latency region versus rest** — `region-partition.spec.js` checks a real
OTLP partition at exact time/duration boundaries with a one-row inspection page,
source-free mobile staging and parent return, reusable configuration and immutable
comparison evidence. Hegel/Quint and Bombadil guard membership and ownership.
See [the contract and frames](../REGION_PARTITIONS.md).


## 138. Inspect and cite retained Activity observations

Implemented: [executable journey](../../journeys/tail-evidence.spec.js), optional
local XTDB/ClickStack lab. This foundation journey predates browser Activity
selection/publication; journeys 139–145 below record its subsequent delivered
adoption and bounded acceptance.

1. Preparation ingests two duplicate span occurrences through OTLP into ClickHouse.
   The [development fixture helper](../../scripts/tail-evidence-fixture/main.go)
   polls actual source data through the Go adapter, freezes separate read windows
   and publishes through the typed backend into XTDB. No UI data is mocked and no
   client-record publication endpoint is added.
2. Open that retained selection: see Activity origin and recorded custody time,
   then inspect an occurrence's compact service/event-time/status/duration summary.
   Read time, poll window, basis and occurrence ordinal remain distinct. Attributes
   disclose locally; UTC event timestamps retain exact nanoseconds in tooltips.
3. Start and save an investigation with the original review pinned. Revise the
   review on the retained selection and reopen the document: the citation keeps
   its original review and read lower bound.
4. Browse retained metadata scoped to the environment and Activity origin. Selecting
   metadata does not fetch records/reviews; opening is explicit. Inspect the mobile
   citation without horizontal overflow or telemetry reads.
5. Observed Quint replay rejects extra source reads/publications, changed pinned
   review, custody clock or exact manifest. Bombadil opens/folds provenance and
   attributes, searches locally and checks occurrence count, origin, no writes/
   reads and layout. Existing journeys 99–101 cover the v1 page/citation paths.

![Compact retained Activity provenance](screenshots/138a-retained-activity-provenance.png)
![Pin an Activity review in a saved investigation](screenshots/138b-pinned-activity-review.png)
![Browse the origin-aware retained catalog](screenshots/138c-activity-catalog.png)
![Read the pinned Activity citation on mobile](screenshots/138d-mobile-activity-citation.png)

### 139 — Preserve exact Activity occurrences and recover a lost response

[journey](../../journeys/tail-selection.spec.js) injects two identical real OTLP
logs, independently checks ClickHouse multiplicity, receives server-issued SSE
references, selects both and pauses. Tags/note appear only after selection.
A real XTDB commit is allowed to finish before its browser response is dropped.
Reloading Saved retries the unchanged journal intent and opens the same manifest,
review and custody clock, without any telemetry source reads. Observed Quint
replay includes rejecting controls. Bombadil selects, clears and resumes without
publishing or generating workload. These handles describe observed multiplicity,
not physical event identity or a complete trace/population.

![Selected Activity occurrences](screenshots/139a-activity-selection.png)
![Recovered retained Activity](screenshots/139b-retained-activity.png)

![Compact mobile selection](screenshots/139c-mobile-selection.png)

### 140–145 — Activity retention under failure

[Executable matrix](../../journeys/tail-acceptance.spec.js) and
[bounded acceptance](../ACTIVITY_RETENTION_ACCEPTANCE.md): actual isolated Go/Python
owners over real OTLP/ClickHouse and optional XTDB exercise local expiry,
unconfigured/failed configuration checks, actual database transport failure,
browser quota refusal, genuine 64-batch capacity exhaustion and proof loss on a
real Go restart. Submitted retries preserve intent; inspection survives failure.
Quint replays the observed outage/retry ledger with rejecting controls. Bombadil
explores capacity-refused inspection. Keyboard/mobile focus and no hidden source
reads are checked. Only the failed configuration GET and local journal quota are
browser fault injections; successful streams/publications are real.

![Expired local selection on mobile](screenshots/140a-expired-selection-mobile.png)
![Unconfigured retention](screenshots/141a-unconfigured-retention.png)
![Save awaiting acknowledgment](screenshots/142a-pending-save.png)
![Capacity-refused stream](screenshots/144a-capacity-unavailable.png)
![Definitely uncommitted proof](screenshots/145a-proof-unavailable.png)
![Mobile original-request recovery](screenshots/145b-mobile-recovery.png)

## Subsequent Activity and component storyboards

`POST /activity-views`, `GET /activity-views/{id}` and `GET /activity-views` expose
immutable version-1 rolling selector and column configuration. Publication has a
16 KiB Go/proxy bound and rejects execution/custody fields. List supports exact
optional environment and cursor parameters, returning up to 20 items. These
source-free file routes are independent of fixed `/exploration-views` and XTDB.
Opening configuration never opens a tail connection. See [Activity views](../ACTIVITY_VIEWS.md).

### 146 — Reusable rolling Activity views

[Executable journey](../../journeys/activity-views.spec.js) uses ordinary real
OTLP logs, actual field filtering/column changes, immutable API publication and
clipboard sharing. Reload and Saved/Services reopen stage without source reads;
Start explicitly reads a new recent population. Draft keep/discard and mobile
layout are exercised, followed by observed Quint controls and Bombadil's paused
configuration inspection. [Contract and diagram](../ACTIVITY_VIEWS.md).

![Staged rolling Activity configuration](screenshots/146a-staged-activity-view.png)
![Mobile configuration before Start](screenshots/146b-mobile-staged-view.png)

Journey [147](../../journeys/activity-view-race.spec.js) delays an actual committed
configuration-save response while another view opens. Releasing it preserves the
newer draft's keep/discard controls and title, with no telemetry reads.

## 148 — Write a filter before a matching observation exists

Open a saved Activity configuration paused. In Filters, edit OR conditions for
missing versus empty release, exercise a nested group and correct an invalid
column. No telemetry is read. Apply with the keyboard, inspect the two matching
ordinary OTel records, and save/reopen the exact selector. In a fixed saved view,
edit the same tree and open a child; only explicit Run reads it. Desktop/mobile
frames: [148a](screenshots/148a-filter-editor.png),
[148b](screenshots/148b-mobile-filters.png). Executable journey:
`journeys/filter-editor.spec.js`; [semantics and boundaries](../FILTER_EDITOR.md).

### 149 — One filter editor across record questions

[Executable journey](../../journeys/single-filter-editor.spec.js) starts from a
paused shared fixed question, edits a nested missing/empty selector, corrects an
invalid field without losing it during environment changes, shares/reloads without
reads, and runs via keyboard. It verifies exact ordinary OTel record populations
and fixed-window refinement, then validates observed staging with rejecting Quint
controls and runs a source-free Bombadil editor campaign.

![Shared single-question editor](screenshots/149a-single-question-filters.png)
![Mobile typed editing](screenshots/149b-mobile-single-filters.png)

### 150 — Discover observed fields before refining a scope

[Executable journey](../../journeys/filter-suggestions.spec.js) reopens a saved Activity view without reads, explicitly starts ordinary OTel logs, browses a bounded field sample, and stages present-empty from the keyboard. A single fixed question selects missing; a saved population stages a canary child without reading it. Held actual responses and a negative transport fault verify stale suggestions stay withdrawn. Quint checks observed staging; Bombadil reads once explicitly, then explores values without rerunning.

![Observed fields](screenshots/150a-observed-activity-fields.png)
![Mobile values](screenshots/150b-mobile-observed-fields.png)

### 151–152 — Counted values beyond the shown page

[Executable journeys](../../journeys/filter-values.spec.js) compare a 100-row sample with all groups from 103 ordinary OTel logs, then exercise staged missing/canary selection, keyboard/mobile/local search, delayed response refusal, negative transport retry, saved child scopes and paused Activity. A separate 1,001-value source fixture verifies refusal without partial totals. [Contract and process diagram](../FILTER_VALUES.md).

![Scoped counted values](screenshots/151a-counted-field-values.png)
![Mobile counted list](screenshots/151b-mobile-counted-values.png)
![Cardinality refusal](screenshots/152-counted-value-refusal.png)

153: [Latency histogram selection](../LATENCY_SELECTION.md), real OTLP/ClickHouse, reverse drag, keyboard/cancel, explicit child/rest comparison, source-free return, desktop/mobile frames and Bombadil. 154: synthetic passive count-bar range contract, SSR/detached upgrade, out-of-plot release and behavior removal.

155: [Optional Activity latency](../ACTIVITY_LATENCY.md), real OTLP/ClickHouse, full counts versus one-row aggregate inspection, actual cadence and pause interval, brush/Apply, fixed remainder comparison and return, stale-response/empty/failure/retry checks, desktop/mobile frames, Hegel state sequences, Quint replay and Bombadil.

156: [Captured Activity latency baseline](../ACTIVITY_BASELINE.md), real OTLP/ClickHouse, frozen 4 versus moving 5→6 counts despite a late baseline arrival, stored-only selection/clear, actual cadence, independent clocks, selector/window differences, no-data, two stale-response paths, negative comparison fault/retry, capture preview cancellation, desktop/mobile frames, Hegel, Quint and Bombadil.

157: Synthetic passive paired-histogram row contract: aligned readable bins, exact supplied percentages, common scale, semantic headers, source-free rendering and mobile composition.

158: [Saved Activity baseline reference](../SAVED_ACTIVITY_BASELINE.md), real OTLP/ClickHouse, exact reference save, source/receipt/comparison-free reload, explicit Start/Show, preserved old views, clear-versus-delayed-save guard, unavailable receipt/retry, actual desktop/mobile frames and staged-only Bombadil/Quint checks.

159: [Activity saved-view discovery](../ACTIVITY_VIEWS.md#activity-entry-storyboard-159), real OTLP/ClickHouse, source-free inline browsing, applied/draft environment isolation, existing keep/discard navigation, actual delayed catalog refusal, cleared inspector, loaded-page search, mobile frames, explicit Start and Bombadil/Quint controls.

109 now exercises the [consolidated Activity navigation](../ACTIVITY_NAVIGATION.md): folded Investigate, keyboard authoring, source-free paused return/opener focus, retained editor identity/draft, mobile frames, observed Quint controls and Bombadil reopen/edit/return exploration. Existing 102/104–106, 111, scope-analysis and record-column journeys retain their distinct read/selection contracts through the new entry.

160–161: [Compact mobile records](../COMPACT_MOBILE_RECORDS.md). 160 uses ordinary OTLP/ClickHouse, five viewport widths, keyboard occurrence selection/inspection, exact timestamps, source-free paused return/field edits and explicit resume, with observed Quint controls and a real mobile Bombadil viewport. 161 is a separate synthetic SSR/SPA/legacy composition contract, including reversed stylesheet order and zero API reads.

162: [Fixed/saved compact record adoption](../COMPACT_MOBILE_RECORDS.md#fixedsaved-adoption-storyboard-162), separate real OTLP log/span fixtures, independent counts, five widths, full messages, context actions, original receipts/basis, selected fields, source-free saved inspection and two mobile Bombadil campaigns. The synthetic 161 contract now includes spanning inspectors and responsive full/short dates under population styles.

46 now checks [compact service observations](../COMPACT_MOBILE_RECORDS.md#service-observation-adoption-46) at five widths, keyboard inspection, exact time/projection, source-free resize/handoff/return, retained inspector focus and a 390×844 Bombadil workflow. Updated desktop/mobile frames retain the supporting-read limitation and source/evidence controls.

163: [Operation and renderer audit](../RECORD_RENDERER_AUDIT_2026_10_24.md), fresh woven subject fixture, readable UTC/durations, exact trace metadata, mobile containment, keyboard time focus and source-free local navigation with Bombadil. 162 adds real namespace/timestamp checks in waterfall detail; 139 verifies retained timestamp attributes without source access; 67 covers the passive adapter's missing-status behavior.

174: [Exact integer distributions](../NUMERIC_HISTOGRAMS.md#storyboard-174--ordinary-telemetry-to-reusable-numeric-populations), ordinary OTel spans with integer extrema and nonvalid values, exact full bins beyond the inspection page, keyboard range intent, selected/rest count comparison, capture/YAML reuse, frozen mobile display and delayed real planner navigation. A focused time-bars regression also checks keyboard and pointer ranges after retained-child restoration. Model replay and Bombadil evidence are recorded in VALIDATION.md.

175: [Numeric measure comparisons](../NUMERIC_COMPARISONS.md#storyboard-175--compare-populations-without-losing-numeric-meaning), ordinary OTel spans/logs, signed and uint64 values, full numeric totals/groups beyond inspection pages, fractional deltas and coverage, pin/previous-window/selected-rest flows, immutable capture and reusable-view reopen, mobile paired visibility, delayed real comparison refusal, independent native oracle with 22 controls and a finite Bombadil actor.

176: [Aligned numeric distribution comparisons](../DISTRIBUTION_COMPARISONS.md#storyboard-176--isolate-a-changed-shape), ordinary OTel shape changes, exact counts/all-record shares, local scale/brush intent, each side's original scope, source-free child staging then explicit Run, capture/view and v5/v4 reuse, delayed plan ownership, mobile common geometry, synthetic SSR/SPA/retained gestures, independent model oracles and a finite Bombadil campaign.


177 — **browser, native/model, Bombadil and legacy regression acceptance pass**:
[Explain a selected population](../POPULATION_EXPLANATIONS.md#storyboard-177--explain-and-preserve-a-selected-population),
ordinary OTel selected/remainder comparisons, source-free side filters, exact
observed promotion after late data/lost acknowledgment, immutable reuse,
cardinality refusal and delayed-owner cancellation. Actual frames:
[selected population](screenshots/177a-explain-selected-ordinary-work.png),
[mobile frozen observation](screenshots/177b-mobile-frozen-observation.png), and
[partial field with complete leads](screenshots/177c-partial-field-with-complete-leads.png).
Native replay has 21 main and 11 ownership rejecting controls; the 60-second
Bombadil campaign completes ten phases with two source runs, one observed
promotion and one pure plan. Legacy journeys 174/176 pass. Final artifact review
and push status are recorded in VALIDATION.md.


178 — **AE2 browser, visual, native/model and actor acceptance complete**:
[Explain a completed Activity window](../ACTIVITY_EXPLANATIONS.md), implemented
in `cdd5a88` with retained-visit ownership fixed in `a6533a6`, reuses the shared
Go explanation engine from compact Activity. Corrected real-source tests pass
spans (39.0s), logs (31.0s) and browser ownership (26.1s); the bounded actor passes
60 seconds through 12 phases. Independent visual review found no blocker in these actual frames:

| Frame | Observed flow |
| --- | --- |
| [Spans Activity explanation](screenshots/178a-spans-activity-explanation.png) | Completed applied window, held-occurrence guide and explicit explanation |
| [Logs Activity explanation](screenshots/178a-logs-activity-explanation.png) | The same fixed-scope flow over ordinary logs |
| [Mobile spans](screenshots/178b-spans-mobile-explanation.png) | Compact paired explanation at mobile width |
| [Mobile logs](screenshots/178b-logs-mobile-explanation.png) | Logs comparison with the same compact ownership and scope cues |

The journey exercises explicit pause/time brush, previous-window or saved-capture
scope baselines, fresh reads of both populations, exact observed Save, side
drilldown and source-free paused return. Ownership checks cover Apply/resume/
navigation cancellation and late responses. The held guide is not a full-population
chart; a saved capture supplies scope, not frozen baseline counts. Its originating
link is UI context and is not an ID retained in the saved explanation payload.

Native observed-ledger checks pass: spans 200.18s with 25 rejecting controls,
logs 189.83s with 25, ownership 79.54s with 16, and baseline 4.61s with eight:
74 rejecting controls in total. Seventy Hegel examples pass. The baseline late-arrival check reads 240 then 241
records against selection 248; stored/public capture bytes remain unchanged.

Journey 155 browser assertions passed, followed by a 60-second combined-wrapper
timeout. Direct exact native validation (14.34s) and a fresh actor (25s) passed
separately. These are separate successful checks, not a successful combined
invocation. Final acceptance commit, review and push status are recorded in
VALIDATION.md; no additional commit ID is inferred here.


179 — **ET1–ET3 acceptance complete**:
[Recorded error/exception triage](../ERROR_TRIAGE.md), ET1 `982b9e9` and ET2
`0c4ab38`, uses ordinary OTel named span events, reported status, numeric/canonical
log severity and structured attributes. The existing fixed workflow counts records
once by service/type, stages exact group or service-scoped single-type questions,
and reuses captures, packages and investigation notes without inventing issues.

Real OTel spans (41.2s), logs (35.8s), ownership (10.8s) and a 55-second/11-phase
actor pass. Passive SSR/SPA component cases pass (2.4s/6.8s) on supplied synthetic
data; they are not OTLP acceptance. Actual frames:

| Frame | Evidence |
| --- | --- |
| [Span groups](screenshots/179a-spans-error-evidence-groups.png) | Ordinary span error/event record-grain groups |
| [Log groups](screenshots/179a-logs-error-evidence-groups.png) | Numeric/text/structured log evidence rules |
| [Span exception details](screenshots/179b-spans-reported-exception-detail.png) | Supplied named events and distinct reporting states |
| [Log exception details](screenshots/179b-logs-reported-exception-detail.png) | Supplied structured exception attributes |
| [Passive component contracts](screenshots/179c-passive-exception-details-components.png) | Synthetic SSR/SPA data supply and local disclosure |

Native checks pass spans (23.51s, 33 rejecting controls), corrected logs (20.09s,
33) and ownership (10.85s, 14): 80 controls in total. Seventy Hegel examples pass
(latest 4.95s). Existing 178 Activity explanation logs/spans regressions both pass
(59.5s). OpenAPI 69 Go/75 total operations and frozen 222 responses, 56 requests
and 14 refusals pass.

Supplemental log probes were corrected through actual API requests after a
harness flattening/index mistake. The original rejected ledger remains retained;
all other observations are unchanged. VALIDATION.md records these outcomes and
the final acceptance commit/review/push status.


180 — **real-source, passive component, actor and native/model acceptance pass**:
[Declare and reuse exact decimal quantities](../../journeys/180-decimal-numerics-storyboard.md)
and the [fixed-scale contract](../FIXED_SCALE_DECIMALS.md) cover ordinary span/log
attributes with explicit `decimal64`, scale and unit declarations. Typed predicates
use human fixed-scale literals; source summaries, histogram geometry and arithmetic
retain exact integer coefficients. No binary-float or silent rounding conversion
is introduced, and absent, empty and invalid records retain separate coverage.

Actual real-source main journeys pass spans (53.2s) and logs (approximately 1.2m),
2.2m together. They exercise local declaration/repair and range intent, full
aggregates beyond 100-row pages, exact selected/rest comparisons, immutable capture
reopen, reusable YAML, scale1/9 and resource bindings, integer compatibility,
sentinel handling and explicit refusals. Final mobile comparison figures verify
all four metric rows, both values and complete human range captions without
clipping or sticky-header overlap.

| Frame | Evidence |
| --- | --- |
| [Spans decimal range](screenshots/180a-spans-decimal-range.png) | Exact source coverage and human range intent |
| [Logs decimal range](screenshots/180a-logs-decimal-range.png) | The same typed grammar over ordinary log attributes |
| [Mobile spans comparison](screenshots/180b-spans-decimal-comparison.png) | Selected/rest values with exact human-rational tooltips |
| [Mobile logs comparison](screenshots/180b-logs-decimal-comparison.png) | Shared signed geometry and legible finite range captions |
| [Passive decimal components](screenshots/180c-passive-scaled-decimal-components.png) | Explicitly synthetic SSR/detached values, tiny scale-nine means and unchanged shares |

The 55-second Bombadil campaign passes all 10 phases with four source reads, two
pure planners and one comparison protocol. Source-free component acceptance
passes in 4.9s. Independent native verification passes logs in 33.20s and spans in 36.68s,
rejecting 38 changed controls each (76 total), including raw source/page uniqueness
and full stored capture custody hashes. The first spans fresh-package attempt
had a separately retained transient pre-execution timeout; the final mains passed
with unchanged source budgets. VALIDATION.md records review and final publication
status; no acceptance commit or push is inferred here.


## 181 — Plot values from inspected and discovered fields

[Storyboard](../../journeys/181-contextual-plots-storyboard.md) and
[contract](../CONTEXTUAL_NUMERIC_PLOTS.md). Choose an attribute/resource key,
explicitly declare numeric meaning in the compact dialog, stage its original
fixed scope, then Run. Invalid/empty supplied values remain eligible entry points;
the clicked value does not become an implicit filter.

| Frame | Actual provenance |
| --- | --- |
| [Spans declaration](screenshots/181a-spans-declare-value-plot.png) | Original observation window and explicit field declaration |
| [Logs declaration](screenshots/181a-logs-declare-value-plot.png) | The same contextual flow over ordinary OTel logs |
| [Mobile spans resource plot](screenshots/181b-spans-full-value-distribution.png) | Fresh resource plot; chart scrolls internally |
| [Mobile logs resource plot](screenshots/181b-logs-full-value-distribution.png) | Left edge of the pannable resource distribution |
| [Desktop saved spans attribute plot](screenshots/181c-spans-saved-value-distribution.png) | Complete supplied domain from immutable captured attribute evidence |
| [Desktop saved logs attribute plot](screenshots/181c-logs-saved-value-distribution.png) | Source-free saved rendering with readable full bounds |
| [Mobile logs right pan](screenshots/181d-logs-resource-distribution-right-pan.png) | Same resource scope as 181b: visible positive outlier and exact right bound |

The mobile resource plots and desktop captured attribute plots are different
qualified fields, not baseline/comparison sides. Exact scopes and image metadata
are retained beside the actual ledgers. Each fixture emits 400 records; the
original observation windows match 387 spans and 382 logs, not an assumed 400.

Both full signal journeys pass in 2.4 minutes. Ownership (12.1s), explicitly
synthetic hidden/detached lifecycle (4.1s), and two-signal population origins (19.1s)
pass. The 50-second Bombadil campaign executes all 12 phases and requires an
actual visible full-scope histogram. Independent raw/source/capture and finite
model replay rejects all 96 changed controls across main, ownership and population
ledgers; 70 generated client-to-Go examples pass. Saved captures and parent
companions stay unchanged; only explicit Run reads the plotted source.

The source milestone is `b4d88db`. VALIDATION.md records observer corrections,
actual custody/review results and publication status; screenshots alone do not
establish complete source membership or immutable byte identity.

The representative 181d resource replay performs one explicit logs read, verifies
all 128 exact displayed bins against the archived result, and pans the actual
520px chart to its right edge in a 260px viewport. The 382-record scope and
original captured bytes remain unchanged. The storyboard pairs this with 181b
and distinguishes it from source-free 181c saved attribute evidence.


## 182 — Full-window log patterns

[Storyboard](../../journeys/182-log-patterns-storyboard.md) and
[profile contract](../LOG_PATTERNS.md). Investigate → Log patterns pauses and
stages the completed applied log scope. Run measures inferred ASCII digit-run
shapes, exact counts/shares and overall volume; opening one shape stages its
records/timeline. Service/rest comparisons preserve the full parent and include
unclassified bodies.

| Frame | Actual evidence |
| --- | --- |
| [Full log-pattern population](screenshots/182a-full-population-log-patterns.png) | 382 matching logs and 27 complete groups, with safe labels and exact totals |
| [Saved mobile pattern view](screenshots/182b-saved-log-patterns-mobile.png) | Retained counts, explicit unclassified buckets and fresh-read links |
| [Cardinality refusal](screenshots/182c-pattern-cardinality-refusal.png) | Actual 1,100-pattern source refuses without partial totals |

Main/ownership/cardinality pass on their first browser run (28.9s/4.5s/2.4s).
The 55-second actor completes all 13 phases. Independent native main/ownership/
cardinality checks reject all 65 changed controls; 220 generated cases and
1,400 state steps pass. Full raw capture bytes/custody are retained separately
from the 100-record page. Source-read time, original window and captured evidence
remain distinct; invalid UTF-8 has raw Go/ClickHouse tests, not a fabricated
OTLP JSON browser case. Final review and post-deployment equivalence are recorded
in VALIDATION.md.

## 183 — Log patterns across time

[Storyboard and replay commands](../../journeys/183-log-pattern-time-comparisons-storyboard.md),
[executable journey](../../journeys/log-pattern-time-comparisons.spec.js) and
[comparison contract](../LOG_PATTERN_TIME_COMPARISONS.md). Stage a previous or
saved scope, Run both independently, inspect exact count/rate/share changes,
drill into either side and reopen fresh-read comparison evidence from storage.

| Frame | Actual evidence |
| --- | --- |
| [Unequal-window rates](screenshots/183a-unequal-window-pattern-rates.png) | Baseline 320 logs over 120 seconds and current 160 over 60 seconds have equal 8/3 rates and equal bars |
| [Frozen mobile comparison](screenshots/183b-mobile-pattern-time-comparison.png) | Source-free replay, unobscured title/selector, full pattern union and both sides |
| [Empty baseline shares](screenshots/183c-empty-baseline-share-unavailable.png) | Known zero counts with unavailable shares/differences and honest stable-order copy |

All three browser cases and the 55-second, 13-phase Bombadil campaign pass.
The saved baseline reread advances from 400 to 412 logs while original stored
bytes stay unchanged. The durable visual replay distinguishes source-free
mobile mode from an explicit three-operation empty-window reread. Initial
framing diagnostics are retained. Independent native/finite-model acceptance
and final publication status are recorded in VALIDATION.md.

## 184 — Observed trace-page latency

[Storyboard](../../journeys/184-trace-latency-breakdown-storyboard.md),
[executable ordinary OTel journey](../../journeys/trace-page-breakdown.spec.js),
[synthetic passive-component cases](../../journeys/trace-page-breakdown-components.spec.js)
and [interval contract](../TRACE_PAGE_TIMING.md).

| Frame | Actual evidence |
| --- | --- |
| [Desktop page breakdown](screenshots/184a-observed-page-latency.png) | Readable service/operation groups, inclusive/outside sums, known coverage and equal-width tracks |
| [Mobile page breakdown](screenshots/184b-mobile-observed-page-latency.png) | Compact responsive cards, full labels and physically comparable tracks |

Real overlap/nesting, duplicate IDs, cycles, missing parents, outside-child timing
and 126-span paging pass. Main shows 100ms inclusive, a 70ms direct-child union
and 30ms outside observed children, without double-subtracting overlap or
grandchildren. Group/disclosure actions are local; selecting a row retains the
existing linked-log read. Raw capture reopening preserves all 8651 stored bytes.
The 55-second actor reaches all nine phases; independent properties and all 80
native rejecting controls pass. Initial guard/layout defects and harness
diagnostics remain recorded separately from accepted evidence in VALIDATION.md.

## 185 — Compare two observed trace pages

[Storyboard and commands](../../journeys/185-trace-comparison-storyboard.md),
[ordinary OTel journeys](../../journeys/trace-comparison.spec.js),
[synthetic passive component](../../journeys/trace-comparison-components.spec.js),
[Bombadil actor](../../verification/bombadil/trace-comparison.ts) and
[comparison contract](../TRACE_COMPARISON.md).

| Frame | Actual evidence |
| --- | --- |
| [Two returned trace pages](screenshots/185a-two-observed-trace-pages.png) | Seven supplied label groups, exact signed changes, repeated ambiguity, kind mismatch and readable unavailable rails |
| [Mobile unique-label pair](screenshots/185b-mobile-unique-label-pair.png) | Actual root durations 100/130 ms, inclusive change +30 ms and outside-observed-child change +25 ms |
| [Qualified attribute differences](screenshots/185c-qualified-pair-attribute-differences.png) | Actual changed/added/removed values, present empty versus missing, and separate resource/attribute namespaces |

Pin an already returned reference page, explicitly read another trace in its own
fixed window, and compare locally. Unique service/operation labels with matching
reported SpanKind are candidates, not invocation identities. Repeated labels
remain unpaired; omitted page rows are not zero work. Both rails share relative
elapsed time and equal physical width. Pin, Compare and attribute inspection add
no API request; existing trace expansion and linked logs retain explicit reads.

Main (18.5s), partial/outside-parent (12.9s), stale-owner (20.8s) and the final
synthetic lifecycle case (2.9s) pass. All 80 independent native/model controls are
rejected, and 220 generated cases plus eight boundaries pass. The 55-second actor
completes all 16 phases across 871 snapshots, validates its actual source selector
and exact deltas, then clears the local reference at six source operations.
Original raw capture bytes remain unchanged; no paired-capture wire format is
introduced. The affected 184 main/graph regressions pass (10.5s/10.4s).

The first owner test used an incorrect Back-link assumption for a direct question
URL, and the first actor batched editor mutations into a stale declaration.
Corrected natural draft recovery and separate actor actions preserve the source
and completion guards. Initial contrast frames and failed traces remain retained
and distinguished from final passes in the storyboard and VALIDATION.md.

## 186 — Find traces containing independent span conditions

[Storyboard and flow diagram](../../journeys/186-trace-search-storyboard.md),
[ordinary OTel journeys](../../journeys/trace-search.spec.js),
[Bombadil actor](../../verification/bombadil/trace-search.ts) and
[trace-search contract](../TRACE_SEARCH.md).

| Frame | Actual evidence |
| --- | --- |
| [Cross-service witnesses](screenshots/186a-cross-service-condition-witnesses.png) | Slow database and worker-error records in the same reported trace, with qualified predicate evidence |
| [Same-span witnesses on mobile](screenshots/186b-mobile-same-span-witnesses.png) | One reported span satisfies both conditions; no distinct-span requirement or horizontal overflow |
| [Original-window trace comparison](screenshots/186c-matching-trace-page-comparison.png) | Fresh full-trace reads preserve bounds; existing held-page comparison adds no source read |
| [Controls through first result](screenshots/186d-controls-to-first-result.png) | Concise actual predicates and closed advanced editors; first result 464.046875 px below the panel start |

The applied parent matches 135 records versus a 100-row inspection page. Explicit
trace search finds 15 distinct reported IDs across the full source. Repeated
120/120 matching records remain one trace; a two-item list still reports total
15. A-only and B-only records in different traces do not join. Shared environment
and fixed window constrain all conditions; the original filters constrain only
condition 1, permitting other services in added conditions. Representative
witnesses are not every span, a causal path or completeness proof.

Main (13.9s), ownership including an actual closed-editor rerun (10.8s), and the
55-second actor pass. The actor reaches all 16 phases across 1,089 snapshots,
proves cross-service/same-span/repeated cases and closes at four source reads.
Main uses ten explicit browser source operations plus two separately labeled API
probes; setup independently proves 291 records/19 IDs plus the excluded
environment's three records/one ID. Existing raw capture bytes remain unchanged;
there is no new saved trace-search or paired-capture format. Independent native
main/ownership verification rejects all 80 changed controls, including actual
Quint lifecycle failures; 300 generated cases pass. Final publication is recorded
in VALIDATION.md.

Initial strict-selector/framing failures and the real verbose-editor/mobile-basis
layout fixes remain documented separately from final evidence. No positive
source response was mocked. Malformed/all-zero identity exclusions use backend
raw-source checks, not invented successful OTLP ingestion.

## 187 — Automatic incoming HTTP service overview

[Storyboard and flow diagram](../../journeys/187-http-server-dashboard-storyboard.md),
[real OTel journey](../../journeys/http-server-dashboards.spec.js),
[Bombadil actor](../../verification/bombadil/http-server-dashboard.ts) and
[automatic dashboard scope](../AUTOMATIC_DASHBOARDS.md).

| Frame | Actual evidence |
| --- | --- |
| [Incoming HTTP overview](screenshots/187a-incoming-http-service-dashboard.png) | Compact timeline, full 137-record server population, nullable routes and separate native intervals |
| [Mobile server summary](screenshots/187b-mobile-http-server-records.png) | Recorded count, span Error and duration without conflating client calls or native observations |
| [Desktop in context](screenshots/187c-service-overview-in-context.png) | Real navigation, service scope and profile controls; no isolated demo shell |
| [Mobile in context](screenshots/187d-mobile-service-overview-in-context.png) | Actual catalog/header overhead remains visible; long native lane names cannot overlap plots |
| [Route to trace](screenshots/187e-route-to-original-window-trace.png) | Actual original-window span waterfall and honestly empty linked logs after explicit route Run |

Main passes in 17.6s over 225 ordinary spans. Incoming HTTP matches 137 records
versus 198 in All spans and a 100-row held page. The current-method Server rule
excludes clients, legacy-only, missing and empty methods; reported custom values
are not called verified conformance. The eligible native HTTP duration instrument
is prioritized without combining its four observations with span-record counts.

Explicit profile switches preserve scope and make exactly the bounded widget
reads. A positive one-record window outside all held supporting rows retains the
accepted HTTP profile despite a false hint. Staging/retained-control rejection is
source-free; named Run executes the exact drilldown. Generic worker fallback and
client-only All spans → explicit HTTP zero → All spans seven remain honest.
Equal 20 ms route values use equal 98 px mean/p95 tracks at both widths.

The 45-second Bombadil campaign passes all nine phases over 974 snapshots,
reaching its exact 17-source-operation/three-plan terminal proof. Independent Go
source verification passes in 0.055s over 19 actual responses and rejects all 19
changed-count controls. A separate route-to-trace supplement passes in 9.1s:
staging is source-free, explicit Run returns 120 matching records, two fresh
context reads preserve original bounds, and browser Back restores the same held
overview without reading. The fixtures are generated ordinary OTLP telemetry;
positive source/UI responses are not mocked. Three harness failures and actual compactness/label fixes are retained separately from
the final accepted run. This slice adds no persistence or capture format.

## 188 — Incoming legacy RPC service overview

[Storyboard and flow diagram](../../journeys/188-rpc-server-dashboard-storyboard.md),
[real OTel journey and bounded actor wrapper](../../journeys/rpc-server-dashboards.spec.js),
[unit-label/frame replay](../../journeys/rpc-dashboard-labels.spec.js),
[Bombadil actor](../../verification/bombadil/rpc-server-dashboard.ts) and
[automatic dashboard scope](../AUTOMATIC_DASHBOARDS.md).

| Frame | Actual evidence |
| --- | --- |
| [Incoming RPC overview](screenshots/188a-incoming-rpc-dashboard.png) | Full 132-record population, explicit reported-method grouping and separate native `[ms]`/`[s]` identities |
| [Mobile summary](screenshots/188b-mobile-rpc-summary.png) | Recorded span count, Error and duration without protocol-outcome inference |
| [Desktop in context](screenshots/188c-rpc-service-in-context.png) | Real discovery scope, profile controls and application navigation |
| [Mobile in context](screenshots/188d-mobile-rpc-in-context.png) | Honest entry overhead and native label containment |
| [Method to trace](screenshots/188e-rpc-method-to-trace.png) | Original-window fresh waterfall and honestly empty linked logs, followed by source-free held-service return |

Main passes first attempt in 24.4s: 132 Server records with nonempty legacy
`rpc.system` and `rpc.method`, versus 203 mixed records and a 100-row held page.
Modern keys alone, missing/empty keys, wrong kinds and namespace collisions do not
silently join the profile. One reported method groups 120 records across two
systems/services; both occur in the actual first page. Its exact staged query,
explicit Run, trace/linked-log reads and retained browser Back are verified.
Client-only RPC defaults to generic, with explicit incoming zero and generic
seven-record recovery. The accepted profile survives an unheld one-record range
whose new observation hint is false.

The 30-second actor completes nine phases over 498 snapshots with exact 17-source/
three-plan effects. Independent raw-source verification passes in 0.091s across
18 responses and rejects 18 changed-count controls. A 9.8s focused display replay
qualifies same-name native units while keeping histogram count axes unchanged;
its initial overview uses nine source operations/one plan, then display changes
are source-free. Frames 188a–d use that fresh display fixture; 188e retains the
accepted main. Original semantic main/actor were not rerun for display-only
`ccdd2d1`. Full receipts and the corrected oracle diagnostic are in the storyboard.
Generated ordinary OTLP telemetry is used; positive source/UI responses are not
mocked. No new persistence format or convention-conformance claim is introduced.

## 189 — Database client overview

[Storyboard and flow diagram](../../journeys/189-db-client-dashboard-storyboard.md),
[ordinary OTel journey and actor wrapper](../../journeys/db-client-dashboards.spec.js),
[Bombadil actor](../../verification/bombadil/db-client-dashboard.ts) and
[automatic dashboard scope](../AUTOMATIC_DASHBOARDS.md).

| Frame | Actual evidence |
| --- | --- |
| [DB client overview](screenshots/189a-db-client-dashboard.png) | Full Client population, literal missing/empty operation groups and separate native units/identities |
| [Mobile summary](screenshots/189b-mobile-db-summary.png) | Client-record count, recorded span Error and duration |
| [Desktop in context](screenshots/189c-db-service-in-context.png) | Real discovery/service scope and explicit role choices |
| [Mobile in context](screenshots/189d-mobile-db-in-context.png) | Honest entry overhead and contained native labels |
| [Database population to trace](screenshots/189e-db-population-to-trace.png) | Fresh original-window waterfall and honestly empty linked logs from the isolated 108-record population |

Main passes in 28.6s over 233 generated ordinary OTLP spans: 137 current-key Client
matches versus 189 mixed records and a 100-row held page. Legacy-only system keys,
wrong kinds and namespace collisions do not silently join. `SELECT` combines 120
records across two systems/database namespaces; both occur in the actual first
page. Explicit filters narrow to 108 before trace expansion. Edits, Records-tab
selection and retained service return are source-free; named Runs and trace/log
reads remain explicit. Client observations are not database-server or pool health.

Two pure metadata API probes demonstrate diagnosis fallback from operation to
database namespace to system without automatic read fanout or raw SQL/summary
selection. The full browser task uses 49 logical source operations/11 plans,
with those two probes separately labeled. The 30-second actor completes nine
phases/461 snapshots and its exact 17-source/three-plan proof. First-run independent
Go verification passes in 0.145s across 25 actual responses and rejects 25
changed-count controls. Five actual frames use the accepted main fixture.

The initial harness failed to select Records after a filter-editor Run; the
source-free tab correction and preserved 34.4s diagnostic are documented separately
from the accepted main. No positive source/UI response was mocked and no new
persistence format is introduced. Full timings, receipts and limits are in the
storyboard and VALIDATION.md.

## 190 — Messaging consumer and producer overviews

[Storyboard and flow diagram](../../journeys/190-messaging-dashboard-storyboard.md),
[ordinary OTel main and actor wrapper](../../journeys/messaging-dashboards.spec.js),
[responsive display replay](../../journeys/messaging-dashboard-display.spec.js),
[Bombadil actor](../../verification/bombadil/messaging-dashboard.ts) and
[automatic dashboard scope](../AUTOMATIC_DASHBOARDS.md).

| Frame | Actual evidence |
| --- | --- |
| [Messaging overview](screenshots/190a-messaging-consumer-dashboard.png) | Full Consumer population, literal destination groups, explicit roles and separate native identities |
| [Mobile summary](screenshots/190b-mobile-consumer-summary.png) | Captured Consumer-record count, recorded Error and duration |
| [Desktop in context](screenshots/190c-messaging-service-in-context.png) | Real navigation, service scope and readable profile names |
| [Mobile in context](screenshots/190d-mobile-messaging-in-context.png) | Two-column full labels without overflow; existing entry overhead remains visible |
| [Consumer population to trace](screenshots/190e-consumer-population-to-trace.png) | Original-window supplied trace and honestly empty linked logs, without inferred asynchronous pairing |

Main passes in 33.3s over 353 generated ordinary OTLP spans: Consumer 137,
Producer 120 and All spans 309 in the same scoped service. Each role requires its
exact kind and nonempty current `messaging.system`; records are not unique
messages, processing completions or delivery confirmations. Destination `orders`
combines two systems/groups and 120 records. Explicit system/group filters return
108 before trace expansion. A genuine reported link is inspected in the existing
Raw record UI without reading; its Producer target is not automatically joined,
fetched or drawn as a parent. Trace origin and source-free service return remain
exact. Missing/empty groups, false-hint range retention and explicit zero fallback
are verified.

The 30-second actor completes 11 phases/511 snapshots, including both role
switches, at exactly 25 source operations/five plans. Independent raw-source checks
pass first run in 0.225s across 27 responses and reject 27 changed-count controls. Main
uses 53 browser source operations / 12 plans plus two separate pure fallback probes.
A 10.0s display replay verifies six full labels, selected state, mobile two-column
layout and source-free resizing after nine initial source operations/one plan.
Frames 190a–d use that fresh display fixture; 190e retains the accepted main.

The initial ambiguous waterfall selector and actual mobile letter-stack layout
issue are preserved separately from final evidence. The latter was fixed by
CSS-only `8ae7573` without repeating the semantic main/actor. No positive source/UI
response was mocked, no queue/broker health or inferred pairing is claimed, and
no persistence format is added. Exact receipts and limits are in the storyboard.


## 191 — Generic startup with ordinary Python-tagged OTel

[Executable storyboard](../../journeys/191-generic-startup-storyboard.md) · [Activity](screenshots/191a-generic-python-activity.png) · [Service](screenshots/191b-generic-python-service.png) · [Mobile frozen evidence](screenshots/191c-generic-saved-python-evidence.png).

The isolated generic runtime contains no grammar, sample definitions or operation bindings. Blank idle startup makes no population reads; explicit Activity → discovered service → staged fixed question → Run returns 140 Python-tagged server-span records, then capture/reopen preserves the same fixed evidence. Main 20.0s, bounded actor 30.9s/six phases, 465 snapshots, offline API supplement 2.0s and refreshed lab compatibility 6.0s pass. Two initial real startup defects (automatic local tail and residual example catalog fetch) are fixed and retained as diagnostics; no failed attempt is claimed passed. See the storyboard for source effects, capture custody, exact receipts and reproduction commands.


## 192 — Native interval to coincident records

[Executable storyboard](../../journeys/192-native-record-context-storyboard.md) · [Selected native interval](screenshots/192c-selected-native-interval.png) · [Exact span context](screenshots/192a-native-interval-record-context.png) · [Mobile logs](screenshots/192b-mobile-exact-interval-logs.png).

A live counter/histogram interval stages an exact nanosecond log/span question using only shared service/resource identity. Explicit Run finds 115 records beyond the first 100-row page; existing trace expansion intentionally reads a disclosed broader coarse window. Back restores retained native evidence and focus without reading. Main 9.2s, four bounded cases 14.7s, ownership 5.6s, eight-phase actor 25.8s/418 snapshots and one-read native-frame replay 3.6s pass. Unknown exports, missing/empty namespaces, empty context, held responses and hidden controls are covered. The initial obsolete trace-button selector failure is retained separately. See the storyboard for exact receipts and effect accounting.


## 193 — Counted-values read projection

[Executable storyboard](../../journeys/193-counted-values-projection-storyboard.md) · [Full-population values](screenshots/193a-counted-values-full-population.png) · [Mobile](screenshots/193b-mobile-counted-values.png).

Explicit field lookup uses the narrow counted-values endpoint, retaining exact numeric predicates, missing/empty states and independent count/group read receipts. Main 21.7s verifies 143 full records beyond the first 100 shown, a separate five-record case, source-free staging and held-response scope rejection. The bounded actor 31.6s passes seven phases/511 snapshots with one new-endpoint lookup and no legacy group request. A separate controlled audit observes four→two source reads for each fixed cardinality case; see its receipts and limitations in the storyboard. Two setup-policy omissions and the initial actor's unscoped hidden-editor selector remain preserved diagnostics. Independent raw-ledger validation passes in 0.177s with eight changed controls rejected. Prior 151/152 compatibility expectations were updated, not rerun.


## 194 — Rich service record projections

[Executable storyboard](../../journeys/194-service-dashboard-projection-storyboard.md) · [Desktop aggregates](screenshots/194a-rich-service-aggregates.png) · [Mobile](screenshots/194b-mobile-service-aggregates.png) · [Product context](screenshots/194c-service-overview-in-context.png).

Service widgets use rich aggregate-only reads while retaining exact count/group/duration/status/time-series meaning. Main 19.9s proves 137 HTTP spans, 155 logs, missing/empty diagnosis, full 100-row admission/drilldown data, fresh immutable capture 137 and source-free retained return; range continuity and late-owner 8.0s checks pass. The actor 31.3s covers nine phases/462 snapshots. Four controlled widget pairs preserve aggregate fingerprints while main reads fall 5→3 and diagnosis reads 4→2; see the audit receipts and limits. Capture-history/quoted-label and canceled-response collector harness diagnostics are retained. Independent native gates reject eight main and four owner controls (combined 1.008s; final stricter owner check 0.218s). Prior 187–190 transport expectations are updated, not rerun; historical frames/receipts remain unchanged.

## 195 — Cumulative stock observations

[Executable storyboard](../../journeys/195-cumulative-stocks-storyboard.md) · [Desktop](screenshots/195a-cumulative-stock-observations.png) · [Mobile epoch](screenshots/195b-mobile-stock-epoch.png).

Actual cumulative nonmonotonic exports retain seven exact pool/state/instance streams, descending values, signed fractions, accumulation epochs, stale observations and unknown missing/conflicting reports. Main first pass 9.5s and six-phase actor 30.3s/509 snapshots pass; independent validation 0.037s rejects six changed controls. A fresh full measurement capture reopens source-free with identical 12,964-byte stored evidence. Generic stock eligibility is separate from semantic-convention compliance, physical pool counts or automatic pool health; no stock interval context is claimed. See the storyboard for source effects, the separate pure planner scope correction and full receipts.

## 196 — Named database pool observations

[Executable storyboard](../../journeys/196-pool-connections-storyboard.md) · [Desktop](screenshots/196a-named-pool-counts.png) · [Mobile states](screenshots/196b-mobile-pool-states.png).

Pinned 1.37 connection stocks assemble five exact pool identities with separate instances and reported states, including a genuine zero, custom state, epochs, stale data and unknown missing/conflicting exports. HTTP native priority and the two-widget budget remain unchanged. Main 23.9s and eight-phase actor 30.6s/478 snapshots pass; independent validation 0.362s rejects six changed controls; native staging and local filters are source-free, and fresh capture preserves its domain through frozen reopening with identical 16,904-byte stored evidence. Generic signed stocks remain separate from the certified count domain. The initial negative-metadata request-size harness failure is retained; no pool totals, utilization or health are inferred.

## 197 — Independent pool capacity and acquisition wait

[Executable storyboard](../../journeys/197-pool-signals-storyboard.md) · [Desktop options](screenshots/197a-pool-signal-options.png) · [Mobile](screenshots/197b-mobile-pool-options.png) · [Actual wait interval](screenshots/197c-exact-wait-interval.png).

Contextual pure planning matches complete pool identity after removing only count state; existing default HTTP/count reads remain unchanged. First-pass main 14.4s and late-response ownership 4.5s pass. Explicit capacity zero/epoch and wait bucket-bound p95 reads each capture and reopen immutable bytes; preview and retained Back/navigation are source-free. The nine-phase actor 31.0s/405 snapshots passes; independent validation 0.373s rejects eight changed controls. Wait-only and unconfirmed plans stay useful/honest, and mismatched identities are excluded. No utilization ratio or pool-health inference is introduced; see the storyboard for exact requests, source effects and both capture hashes.


## Gotogether cross-app completion pilot

[Experiment and executable source](../GOTOGETHER_FOLLOWUP.md) · [Ready report](screenshots/gotogether-report-ready.png) · [Correlated trace](screenshots/gotogether-correlated-trace.png) · [Frozen reopening](screenshots/gotogether-immutable-reopened.png).

The real sample-app report → ordinary OTLP/ClickHouse → Goo fixed records/waterfall
→ fresh capture → source-free reopen journey passes in 13.074s. Ten persisted
invocation/completion checkpoints feed a single bounded coordinator model at
pinned gotogether `65093ec`; validation and four controls pass in 24.498s, followed
by a strengthened conforming-prefix check in 3.296s. Full stored evidence stays
6,080 bytes with the same content ID/SHA256. The first JS hashing-harness failure
is retained separately; the accepted run uses independent typed Go hashing.
Settling concerns only finite driver calls/bodies, not global collector drain or
telemetry completeness. This is a Playwright integration pilot; existing Bombadil
coverage and Goo dependencies remain unchanged.


## 198 — Compact aligned pool evidence

[Executable storyboard](../../journeys/198-pool-view-storyboard.md) · [Desktop](screenshots/198a-aligned-pool-evidence.png) · [Mobile](screenshots/198b-mobile-aligned-pool.png) · [Whole viewport](screenshots/198c-pool-view-in-context.png).

Explicit optional reads compose held pool evidence without changing the default read budget. Exact latest observation/epoch alignment certifies reported used 3 / maximum 10; six real boundary probes withhold or flag invalid ratios. Main 21.1s, ownership 12.0s, eleven-phase actor 31.5s/285 snapshots and independent validation 0.549s/twelve rejecting controls pass. Cached reopening is local, partial signals remain useful, and individual capture bytes remain unchanged. Final same-fixture display replay 12.5s proves complete mobile values, square controls, full timestamp tooltips and no X/Y cropping; earlier geometry/harness/CSS diagnostics are retained. No atomic-snapshot, health or composite-custody claim is made.
