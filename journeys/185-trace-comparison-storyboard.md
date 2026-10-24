# 185 — Compare two observed trace pages

The real-telemetry main, partial-page and ownership journeys, explicitly synthetic
component acceptance, 55-second Bombadil campaign and focused trace-page
regressions pass. Independent native/model checks reject all 80 changed controls.
No comparison asserts that two spans represent the same invocation.

A developer pins a returned trace page, explicitly reads another trace, and asks
which reported operations changed duration or supplied attributes. The held
reference is local to the record-results session; it is not a new stored evidence
kind, a causal alignment or a complete-trace comparison.

## Follow two real traces

[Fixture](185-trace-comparison-fixture.js) emits ordinary OTLP spans without
manifests. The reference and current traces each contain seven observations in
the same unique environment. Additional traces exercise outside-parent timing
and a 126-span fan-out. [Executable journey](trace-comparison.spec.js) and
[source/observation helpers](185-trace-comparison-support.js) preserve exact raw
rows, API questions, fixed windows, read clocks and basis metadata.

1. Run the six-second fixed reference scope, then Expand trace. The existing
   trace read and exact linked-log read remain independent source operations.
   Pin the returned page locally; pinning adds no API request.
2. Close the inspector, change the trace filter, and explicitly choose a fresh
   one-minute window in the same question owner. Run and Expand the current trace.
   Its source question has different bounds; the reference's rows and provenance
   remain unchanged rather than being silently reread into the new window.
3. Compare with reference locally. Unique service/operation labels are only
   candidates: exact nonempty matching SpanKind and usable unique identity are
   required for row pairing. The two traces have different trace/span IDs.
4. Inspect exact signed timing differences and a shared 130 ms relative elapsed
   scale. The root grows from 100 to 130 ms inclusive (**+30 ms**) and from 50 to
   75 ms outside observed direct children (**+25 ms**). The catalog lookup shrinks
   from 20 to 10 ms (**−10 ms**). These are supplied-page observations, not CPU
   time, complete self time, critical path or whole-request attribution.
5. Keep ambiguity visible. Two prepare rows on each side remain unpaired even
   though their known inclusive subtotal changes by +10 ms. Cache/enqueue groups
   appear on only one supplied page; absence is not a zero observation. The
   dispatch label has differing SpanKinds and no row pairing. A zero marker has
   exactly zero painted bar width, not a minimum positive mark.
6. Inspect the root's qualified attribute differences without a source read.
   Changed, added and removed keys remain distinct; present empty string versus
   missing uses the shared typed-field component. `work.namespace` remains two
   separate fields in span attributes and resource attributes. An unchanged key
   is omitted from the difference table; no unsupplied map completeness is inferred.
7. Reopen existing reference evidence without querying telemetry. Comparison
   controls do not add capture fields or mutate its stored bytes.

The final main passes in **18.5s** after the reviewed contrast polish:
`artifacts/trace-comparison/trace-comparison-618c4bebb6/ledger.json`.

There are **six logical browser source operations**: reference search, reference
expansion and linked logs, current search, current expansion and linked logs.
Pin, compare, pair inspection and stored-capture reopening add no source read.
Pin, compare and attribute inspection also assert no API request of any kind.
All API requests are retained, including metadata GETs, and unexpected POST
endpoints count against the source budget. These are browser operations, not
ClickHouse statement counts.

The existing reference capture is created as a separately labeled setup source
operation before browser instrumentation. Full files `scope-capture.json` and
`scope-capture.after.json` are byte-identical:

- ID: `a3fe17e9079d11802fe0b62847e1a5b86126e52d70e35df5bfc64b6aaec35b0d`
- Size: 12,477 bytes.
- SHA-256: `c5d4a6ab3be045df0ed6338f171b3da8fc5127d0c2df5e06094ee5055ce16c0c`.

[Desktop: all seven supplied label groups](../docs/journeys/screenshots/185a-two-observed-trace-pages.png)
shows exact positive/negative changes, repeated labels, unmatched sides and the
kind refusal. Shared rails have equal physical widths; zero and unavailable
alignment are not painted as positive work.

[Mobile: the root's unique-label candidate](../docs/journeys/screenshots/185b-mobile-unique-label-pair.png)
focuses the same real pair, its 100/130 ms durations and +30/+25 ms changes.

[Qualified attribute differences](../docs/journeys/screenshots/185c-qualified-pair-attribute-differences.png)
shows the same real root pair. All frames use actual application styles and
ordinary scrolling. The checked-in capture helper requires the element to fit
inside the real fixed inspector's clipping bounds; it does not hide overlays or
alter the rendered component.

## Preserve incomplete timing

The partial/outside-parent case passes in **12.9s**:
`artifacts/trace-comparison/trace-comparison-ae88eb2a0c/partial.json`.

A child outside its current parent makes the parent's remainder difference
unavailable while its known inclusive difference stays zero. The child's own
reported inclusive duration can still differ by +10 ms. A missing parent remains
a local page diagnostic. No invalid edge is clipped into a plausible remainder.

The fan-out query counts 126 matching spans but supplies only 100 in this page.
The comparison explicitly marks partial context, uses those 100 observations
and keeps 99 fan-out rows unpaired. It does not expand full-query groups into
invented row pairs or claim that an omitted label proves new/removed work.

## Retain ownership through delayed reads

The final ownership journey passes in **20.8s**:
`artifacts/trace-comparison/trace-comparison-82a9feba5a/ownership.json`.

It checks closed comparison details, a connected but hidden tray and retained
pair actions, then explicitly re-expands before comparing again. The reference
survives unchanged. Hidden controls can remain enabled DOM elements but are inert;
this is distinct from an invalidated or pending accepted-page epoch.

For pagination, the test fetches a real successful second-page response and
holds only its delivery. Pin and Compare are disabled while the visible owner
waits; retained programmatic clicks add no comparison or source read. A natural
environment draft change invalidates the owner. Releasing the real old response
publishes neither a comparison nor a linked-log read. Restoring the environment,
then explicitly Run and Expand, recovers accepted authority with the original
held reference. The complete eight-action lifecycle ends at 15 source operations.
Passive per-action observations retain actual URLs, connection/visibility,
comparison/difference counts, disabled controls and reference provenance.

## Synthetic and exploratory checks

[Passive component acceptance](trace-comparison-components.spec.js) is explicitly
synthetic. Its final focused run passes in **2.9s** (see exact log below). It covers SSR attributes, predefinition properties, detached updates,
unknown versus missing maps, accessor rejection, exact zero pixel width, equal
reference/current rails and semantic object-key-order equality. Retained pair
buttons under closed details, while detached, and under `aria-hidden` emit no
new intent. It tracks all fetch/XHR attempts and requires zero. Invalid supplied timing is not represented
as an ordinary-OTLP ingestion claim.

[Bombadil actor](../verification/bombadil/trace-comparison.ts) passes its
**55-second** campaign in 56.5s test time. All **16 phases** (0–15) appear across
871 observed snapshots. The terminal proof retains +30,000,000 ns inclusive and
+25,000,000 ns outside-child differences, distinct trace IDs and repeated-label
ambiguity; it confirms the reference was cleared and Compare disabled, with
exactly six source operations.

Each filter transition is a separate action. The actor checks the visible
namespace/key/operator/value before Run and the actual intercepted source AST
after Run (`attribute:work.role = current`). It then compares locally, inspects
qualified attributes, rejects a closed pair action and clears the local
reference. The checked-in wrapper verifies phase coverage and terminal facts;
exit code alone is insufficient.

`artifacts/trace-comparison/trace-comparison-e1710ff6f2/bombadil/` retains the
actual trace, process log, run result and `acceptance-summary.json`.
`artifacts/trace-comparison-actor-final.log` records the final pass.

Independent native verification passes with 38 main controls (7.639s), 22 partial
controls (22.960s) and 20 ownership controls (9.414s). Each changed control is
rejected. Adjacent `native-validation.txt` files retain the source, exact
interval/delta/geometry/custody checks and finite observed-effect replay. The
separate generated campaign covers 220 cases and eight fixed boundaries,
including odd-coefficient values beyond 2^53; it passes in 6.166s.

The affected existing 184 main and graph/paging regressions pass in **10.5s** and
**10.4s** after the shared record-results owner changed. Their fresh artifacts
are `trace-breakdown-ffc6310687/ledger.json` and
`trace-breakdown-5c0cec8afb/graphs.json` under `artifacts/trace-page-breakdown/`.
They ran in the same serialized batch as the initial passive pass (2.2s) and the
failed first actor; that combined invocation is not reported as an all-pass run.
The final passive lifecycle supplement is separately retained in
`artifacts/trace-comparison-passive-final.log`.

## Reproduce

```sh
npx playwright test journeys/trace-comparison.spec.js \
  journeys/trace-comparison-components.spec.js --grep-invert 'optional real'
GOO_TRACE_COMPARISON_BOMBADIL=1 npx playwright test \
  journeys/trace-comparison.spec.js --grep 'optional real'
```

Use the repository wrapper for wider-suite traffic-generator isolation. These
focused checks use the already deployed local lab with an exclusive browser slot.
Custody helpers address the explicit local Docker socket, clear remote/TLS
selectors and preserve the inherited credential configuration.

The deployed asset/route hashes and actual MIME/status checks are retained in
`artifacts/trace-comparison/deployment.json`. TC1 `8b9c4d0`, TC2 `b6422c1` and the
display-only contrast/compactness follow-up `2200410` are the tested source basis.
No Go binary rebuild or API/capture wire change was needed.

## Preserve initial diagnostics

The first main (13.5s) and partial case (12.9s) passed their source assertions.
The first owner case failed a harness assumption: navigating directly to a
question URL restores a one-visit trail, so there was no Back link. The corrected
case uses an actual environment draft change and explicit recovery; no app
behavior was loosened.

The initial desktop frame exposed light default rails in the dark product;
unavailable labels were difficult to read and empty rails looked filled. The
reviewed scoped contrast/compactness patch keeps arithmetic and sources unchanged.
The old frame and first trace are preserved under
`artifacts/trace-comparison/first-diagnostics/`, with
`artifacts/trace-comparison-first.log`. Final main/owner both pass in 43.6s in
`artifacts/trace-comparison-second.log`. The already accepted partial case did not
need repeating for display-only changes.


The first actor (`trace-comparison-1a250bd67e`) also remains retained. Its batched
filter edits raced the editor's namespace rerender and queried
`attribute:service = current`, yielding no matching records and correctly failing
the completion requirement. The fix separated each edit into its own actor action
and added pre-Run declaration plus actual-request guards. No application behavior,
source expectation or time budget was relaxed. Its failed trace/screenshots are
under `artifacts/trace-comparison/actor-first-diagnostics/` and its original
campaign directory. The final actor alone was rerun; accepted source journeys,
native evidence and the already passing 184 regressions were not repeated.
