# 181 — Plot values from an observed field

A reported string is a starting point for a question. **Plot values** lets the operator declare its numeric meaning, stage the original fixed scope, and explicitly read the full matching population. It does not infer type or unit from a sample.

The same `work.amount` key appears in attribute and resource namespaces with different values. Fixtures report ordinary OTel spans and logs, including empty and invalid strings, signed extrema, and coefficients beyond JavaScript's exact integer range. The declared interpretation in this journey is decimal64, scale 2, USD. See the [feature contract](../docs/CONTEXTUAL_NUMERIC_PLOTS.md).

## Journey

1. **Inspect Activity.** Start a saved Activity view and inspect a held record. Plot its attribute value. The stream pauses, and the declaration uses that observation's original completed read window and applied filter. Type and unit are blank; choosing decimal64 also requires an explicit scale. Cancel restores focus without an additional read.
2. **Declare and stage.** Enter scale 2 and USD. **Open value distribution** stages a fixed histogram question. It keeps signal, environment, exact window and original predicates, clears grouping and other measures, and does not read telemetry. The named Run reads all matching records. A page of 100 inspected rows does not supply the aggregate.
3. **Follow another entry point.** Plot the same qualified field from the completed fixed read: its explicit declaration is retained. Plot an observed empty resource namesake: type, scale and unit must be declared again. The observed-field browser supplies the same action after a completed discovery read.
4. **Investigate values.** Brush locally, explicitly run the selected population, return to the parent, and stage a numeric comparison with the rest. The pure range planner preserves exact human decimal thresholds and coefficient bin bounds; the comparison reads both full cohorts only on Run.
5. **Preserve and reuse evidence.** Capture the parent plot and reopen its immutable evidence. Plot a captured resource field to stage a fresh read of the captured scope. The original receipt remains unchanged. Return to paused Activity restores the original Plot opener's focus.
6. **Explore linked signals.** In a two-signal saved population, plot a supplied span attribute from its record inspector. Only that span question enters the child. Run explicitly, then Back restores the same logs companion node and the original two queries. Repeat from the logs observed-field browser with the resource namesake.
7. **Reject stale ownership.** Navigation, dirty scope, Apply and Resume prevent retained obsolete confirmations from staging or reading. An explicitly synthetic lifecycle case also verifies permanent cancellation after hidden-tab notification and rejection while the host is detached. It supplies no telemetry responses.

## Actual frames

| Frame | Evidence and purpose |
| --- | --- |
| [Spans declaration](../docs/journeys/screenshots/181a-spans-declare-value-plot.png) | Native compact declaration over an actual Activity observation |
| [Logs declaration](../docs/journeys/screenshots/181a-logs-declare-value-plot.png) | Explicit type, decimal scale and unit over the original log window |
| [Mobile spans resource plot](../docs/journeys/screenshots/181b-spans-full-value-distribution.png) | Actual fresh resource distribution; its internal chart scroll is intentional |
| [Mobile logs resource plot](../docs/journeys/screenshots/181b-logs-full-value-distribution.png) | Local mobile inspection without page overflow |
| [Desktop saved spans attribute plot](../docs/journeys/screenshots/181c-spans-saved-value-distribution.png) | Complete supplied domain from the immutable parent capture, reopened without a source read |
| [Desktop saved logs attribute plot](../docs/journeys/screenshots/181c-logs-saved-value-distribution.png) | Full readable bounds and legend from saved evidence |
| [Mobile logs resource right pan](../docs/journeys/screenshots/181d-logs-resource-distribution-right-pan.png) | Positive outlier and exact exclusive right bound for the same resource scope as the mobile logs left view |

The desktop frames show the captured **attribute** parent; the mobile frames show the subsequent fresh **resource** plot. They are different declared fields, not a baseline/comparison pair. Desktop screenshot metadata records the exact capture and query. Both complete domains fit their 1,430-pixel chart viewport without internal horizontal cropping.

The logs mobile left view (181b) and right pan (181d) show the same **resource** field and exact fixed scope. Frame 181d comes from one explicitly requested fresh read, not the saved attribute capture. Its full 128-bin payload and numeric coverage match the archived resource result exactly. At 390px, the internal chart is 520px wide in a 260px viewport; scrolling to 260px reveals the populated final bin and `92233720368547758.08 USD (exclusive)`. Opening **All exact numeric bins** exposes all 128 exact half-open bounds and counts; every displayed row was checked against the returned coefficient bin with exact decimal formatting. The internal pan does not create page overflow.

## Acceptance evidence

Both real-signal mains pass in 2.4 minutes against the frozen deployment. Each fixture emits 400 records. The actual observed windows contain 387 spans and 382 logs; every total, coverage value, exact bin and selected population is independently derived from the raw emitted facts and the fixed query. Each main ends with nine named source operations, two pure range plans, one comparison protocol and one separately tracked Activity stream. Declaration, editing, cancellation, staging, Back and saved reopening add no source reads.

- Spans: `artifacts/contextual-plots/contextual-plot-073c7ceb02/ledger.json`.
- Logs: `artifacts/contextual-plots/contextual-plot-f339d8d5ac/ledger.json`.
- Ownership: `artifacts/contextual-plots/contextual-plot-334d9637c4/ownership.json`, passed in 12.1s with zero source operations.
- Hidden/detached synthetic lifecycle: passed in 4.1s.
- Population origins: `artifacts/contextual-plots/contextual-plot-06a76e3f3d/population.json`, passed in 19.1s. Eight recorded transitions end at four explicit source reads; both children contain only the requested signal.
- Bombadil: `artifacts/contextual-plots/contextual-plot-653481246f/bombadil/acceptance-summary.json`. The 50-second campaign passes all 12 phases, including cancellation and restaging, with a two-source-read ceiling and an actual full-scope numeric result and visible histogram.

Each main has adjacent full `scope-capture.json` and `capture-custody.json` files copied read-only from the actual store. Public pages alone do not establish full retained-row or byte custody. `desktop-screenshot.json` records the later source-free frame capture. Earlier rejected harness runs and their traces remain separately retained; the successful final mains use the same production source budgets.

Independent native arithmetic, source/custody and lifecycle checks pass, rejecting all 96 changed controls: 34 for each signal, 14 ownership controls and 14 population controls. Final receipts are recorded in VALIDATION.md and verification/README.md.

The later right-pan replay is recorded in `artifacts/contextual-plots/contextual-plot-f339d8d5ac/mobile-pan-replay.json` and `artifacts/contextual-plots-mobile-pan-final.log`. It records the exact query, one `/api/explore` request, all 128 displayed rows, actual pan/bound geometry, and a fresh read of the original stored capture bytes whose SHA-256 remains unchanged. The raw request is preserved; only absent versus empty `groupBy` is normalized when comparing the editor request with Go’s omitted-empty wire form. A preceding spans diagnostic stopped at that request-format assertion after validating the chart and is retained separately; it is not the published right-pan evidence.

## Run

```sh
npx playwright test journeys/contextual-plots.spec.js --grep '181 (spans|logs|contextual plot dialogs|synthetic|population record)'
GOO_CONTEXTUAL_PLOT_BOMBADIL=1 npx playwright test journeys/contextual-plots.spec.js --grep '181 Bombadil'
```

Reproduce the saved desktop frame and the mobile resource right pan from an explicit retained main ledger:

```sh
node journeys/181-visual-replay.js --ledger artifacts/contextual-plots/contextual-plot-f339d8d5ac/ledger.json --mode both
```

The checked-in replay script also accepts `--mode saved-desktop` (capture GET only, zero telemetry reads) or `--mode resource-right-pan` (exactly one explicit Explore read). `both` runs those modes sequentially in separate browser contexts. The resource mode checks the unchanged exact scope, full numeric summary and bins, all 128 displayed bounds/counts, internal pan geometry, right-bound visibility, page overflow, and the original capture's actual stored-byte SHA-256. It requires the ledger's adjacent `capture-custody.json` and read-only access to the local query container; `GOO_QUERY_CONTAINER` can name that container. `GOO_QUERY_URL` and `CHROMIUM_PATH` follow the existing journey settings.

The representative logs replay above passes from the checked-in script; its log is `artifacts/contextual-plots-reproducible-visual-replay.log`. It regenerates 181c/181d and their adjacent metadata with zero source reads for the saved frame and one for the resource frame, preserving the original capture hash.

The tests require the deployed lab and real OTel ingestion. Browser campaigns run serially; source-free synthetic lifecycle checks are explicitly named.
