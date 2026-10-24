# 184 — Understand timing within a returned trace page

The live journeys use ordinary OTLP spans from three services, without a goo
manifest. A developer follows a trace, compares inclusive span durations with
time outside observed direct children, highlights the contributing page rows,
and preserves the original evidence. The source-free projection is **not** CPU
time, complete self/exclusive time, a critical path or proof of a complete trace.
Group duration sums overlap and can exceed trace wall time.

Browser acceptance and the 55-second Bombadil campaign pass. Independent native
verification also passes all 80 rejecting controls over the actual main, graph
and ownership ledgers. Exact results and finite-model limits are in
[verification](../verification/README.md) and [validation](../VALIDATION.md).

## Follow the overlapping trace

[Executable journey](trace-page-breakdown.spec.js) emits the deterministic
[ordinary OTLP fixture](184-trace-fixture.js): seven traces, 142 span observations
and three services. Its unique environment and fixed query window are retained
in each source ledger. It never supplies fabricated positive API responses.

1. Run the fixed trace question explicitly, then Expand trace. The existing
   inspector reads the trace page and exact linked logs separately. The new
   timing projection performs no additional query.
2. Open **Observed page latency** and **Supplied span diagnostics**. The root is
   100 ms, with direct children covering [10,60) and [40,80) ms. Their interval
   union is 70 ms and the time outside them is 30 ms. Summing the overlapping
   children would incorrectly give 90 ms. A nested grandchild affects its direct
   parent without being subtracted a second time from the root. A zero-duration
   marker stays a real supplied row with explicit zero measures.
3. Highlight and clear service/operation page groups, including keyboard Enter.
   These interactions and disclosures add no source reads. Every displayed group
   shares the exact 100,000,000 ns scale. Both measure tracks have equal physical
   widths: 145.640625 px on desktop and 144.609375 px on mobile.
4. Select the root through the diagnostic row. Its existing waterfall selection
   shows the 100/70/30 ms timing summary and performs the existing exact linked-log
   read. This read belongs to span inspection, not the latency evaluator.
5. Close the inspector normally and Capture evidence through the existing fresh
   source read. Reopen the stored capture without querying source telemetry. Its
   full stored representation stays byte-identical across reopening.

The final main passes in **11.2s**, with five logical browser source operations:
search, trace expansion, initial linked logs, selected-span linked logs and the
capture reread. These are not ClickHouse statement counts. Group selection,
clearing, disclosures and reopening add none.

- Main ledger:
  `artifacts/trace-page-breakdown/trace-breakdown-36054c3400/ledger.json`
- Full stored capture: adjacent `scope-capture.json` and
  `scope-capture.after.json`, both 8,651 bytes.
- Capture ID:
  `de613dc3648092b7bd4a1ad1e7379506a34255df66d75771fe190c7c3e5ad64a`
- SHA-256:
  `f5baae3da9a4d968dde7fcc2c7ab05c7be4e47fd42f981d78c5665edb82dc452`

[Desktop: readable supplied-page duration groups](../docs/journeys/screenshots/184a-observed-page-latency.png)
shows the actual returned overlap trace. The diagnostic disclosure is closed for
this compact documentation frame after its exact row values and readability
were checked. The label cells remain fully visible after focusing Highlight.

[Mobile: the same trace and common scale](../docs/journeys/screenshots/184b-mobile-observed-page-latency.png)
uses actual component styles and two-column cards. The page has no horizontal
overflow; labels and exact measure values remain unclipped. Diagnostic span cells
are 480 px wide in the desktop inspector and 219 px at the mobile layout, avoiding
character-by-character label wrapping.

## Inspect uncertainty rather than inventing missing time

The real graph/page case passes in **12.3s**:
`artifacts/trace-page-breakdown/trace-breakdown-382ef74648/graphs.json`.

- A missing reported parent is explicitly outside this page, not evidence that
  the span is a root or that the trace is complete.
- A child outside its parent makes that parent's child-union/remainder
  unavailable; the child is not clipped into a plausible zero.
- A grandchild outside its direct parent makes that parent's remainder
  unavailable, while its valid interval still contributes to its own parent's
  coverage. The 100 ms root retains a 20 ms child union and an 80 ms remainder.
- Duplicate SpanIds survive as distinct real source observations; ambiguous
  parent edges and duplicate identities stay unavailable. They are not silently
  deduplicated into invocation counts.
- A real reported parent cycle remains inspectable and explicitly diagnosed.
- A fan-out trace counts 126 matching spans but returns pages of 100 and 26.
  Their memberships cover 126 distinct SpanIds. Each timing projection uses its
  own supplied rows, and both pages disclose partial-page context. Full-query
  aggregate counts and supplied-page interval accounting remain separate.

The graph case ran against unchanged arithmetic core `c1c29ef`; the later
`94722e7` patch changes disclosure guards and layout, not the interval model.

## Keep retained controls within their owner

The ownership case first exposed a genuine application defect: Chromium retained
client rectangles for controls under closed details, so programmatic clicks
could commit group/span selection. The unchanged-selection assertion was kept.
The reviewed fix explicitly rejects closed disclosures and hidden ancestors.

Final acceptance checks closed group and diagnostic controls, retained new
controls and the old native waterfall row after navigation, and current controls
on return. Stale actions perform no selection, highlighting, navigation or
linked-log read. Returning permits current group highlighting with no source
read. The final ledger records actual before/after URLs rather than inferring
navigation from source counters alone. Its final focused rerun passes in **9.2s**:
`artifacts/trace-page-breakdown/trace-breakdown-96fed1dbe0/ownership.json`
(`artifacts/trace-page-breakdown-ownership-final.log`). The earlier 4.8s accepted
owner observation remains retained; only the future harness metadata getter was
made explicit. No source or application behavior changed for that rerun.

## Separate passive rendering from live telemetry

[Standalone component acceptance](trace-page-breakdown-components.spec.js)
passes in **1.7s**. Its route is explicitly labeled a synthetic supplied-page
fixture and checks actual production styles and scripts:

- SSR JSON attributes and a predefinition property upgrade;
- detached SPA property updates and reattachment;
- controlled native keyboard selection, with focus preserved on the same button;
- exact 80,000,000 ns remainder, no-data and malformed supplied-row states;
- mobile containment and zero non-GET requests.

These supplied records prove a component contract, not ingestion or source
measurement. They are separate from the ordinary-OTLP journeys above.

## Explore repeatedly with Bombadil

[Actor](../verification/bombadil/trace-page-breakdown.ts) passes a **55-second**
campaign (56.3s test time). The checked-in test reads the actor trace, requires all
nine phases 0 through 8 across 931 observed snapshots, and checks the terminal five-row trace, four source reads
and the root's 100/70/30 ms selected summary. It does not accept exit code alone.

`artifacts/trace-page-breakdown/trace-breakdown-200cd7ae7c/bombadil/`
contains the process log, run result, observed trace and
`acceptance-summary.json`. The actor exercises Run, Expand, disclosure, highlight,
clear, selected-span inspection and closed-diagnostic rejection. Its source-free
operations preserve read counts throughout.

## Reproduce

With the current lab deployed on 8088 and ordinary OTLP ingestion on 4318:

```sh
npx playwright test journeys/trace-page-breakdown.spec.js \
  journeys/trace-page-breakdown-components.spec.js --grep-invert 'optional actor'
GOO_TRACE_PAGE_BOMBADIL=1 npx playwright test \
  journeys/trace-page-breakdown.spec.js --grep 'optional actor'
```

Use the repository's journey wrapper for managed traffic-generator isolation
when running the wider suite. These focused checks used the already deployed,
exclusive browser slot. `GOO_QUERY_URL`, `GOO_QUERY_CONTAINER` and
`CHROMIUM_PATH` select the existing local endpoints/tooling. Custody reads use
an explicit local Docker socket and clear remote/TLS selectors.

## Retained initial failures

`artifacts/trace-page-breakdown-first.log` and
`artifacts/trace-page-breakdown/first-failure-diagnostics/` preserve the initial
failures and provisional clipped desktop/long mobile frames. That invocation was
not a complete pass. Besides the real closed-disclosure defect, two harness
issues were corrected: Capture was attempted behind the open mobile inspector,
and the standalone fixture used `setContent` on a JSON health document instead
of a declared HTML route. No forced clicks or weakened source assertions were
introduced.

The final focused main/ownership/component invocation passes all three cases in
20.7s (`artifacts/trace-page-breakdown-second.log`); the already accepted graph
case was not repeated because its arithmetic/source contracts did not change.
