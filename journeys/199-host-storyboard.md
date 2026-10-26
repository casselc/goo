# 199 · Exact host resources and native observations

Generated ordinary schema-1.37 OTLP fixture telemetry, not a live operating-system
scraper. No positive source or UI responses were mocked. The fixture deliberately
omits service name and instance from metric resources. Equal host IDs and names
across different full resource maps remain separate observed identities.

The accepted functional runtime is source `e5f490e`; the later CSS-only display
revision is `6a95d50`. Deployment hashes and actual HTTP/MIME
checks are in [deployment.json](../artifacts/hosts/deployment.json). Traffic stayed
running. No existing data volumes or application source were changed by acceptance.

## Actual journey

1. Open **Services & resources → Hosts**. Editing environment/lookback performs no
   telemetry read. Explicit **Find hosts** reads sixteen metric observations and
   exposes eight exact resource identities, including missing versus present-empty
   service fields and equal host IDs in different cloud accounts.
2. Open the resource with absent service fields. A pure host plan chooses two
   exact streams. Two independent reads show CPU user mode/logical CPU 0 at 25%
   and reported used memory at 2,048 By. This is not a per-host CPU average,
   fleet total, capacity/health judgment, or simultaneous snapshot.
3. **Matching resource logs/spans** stages an ordinary fixed question without a
   read. Explicit Run finds one matching app record per signal, with a real
   `service.name=host-context-app`. All selected reported resource attributes are
   preserved; absent service fields do not become `service.name=''`. This is
   matching attributes, not exact map/schema equality or a host-to-service join.
   **Back to host** retains the original rendered observations with no reread.
4. **Inspect measurement** stages the existing native owner with no read. Explicit
   Run and individual capture each read that exact source question. The ordinary
   immutable measurement capture reopens without a telemetry read. No host
   dashboard/composite capture format is introduced.

## Actual frames

| Frame | Origin |
| --- | --- |
| [199a desktop observations](../docs/journeys/screenshots/199a-host-observations.png) | Fresh display fixture after the CSS-only compact-plot fix; two independent exact native streams |
| [199b mobile observations](../docs/journeys/screenshots/199b-mobile-host-observations.png) | Same fresh display resource/read; both cards, plots, controls and context links visible |
| [199c workspace context](../docs/journeys/screenshots/199c-hosts-in-context.png) | Actual fresh display inventory and selected resource in the application |

Visible important content is checked on both axes within its card; mobile has no
widget horizontal overflow. Closed evidence descendants are excluded from visible
geometry checks. These frames do not claim to solve overall catalog density.

## Receipts and reproduction

- Independent actual verification **0.225s** (main 0.18s, owner 0.03s),
  **thirteen changed controls rejected**:
  [native log](../artifacts/hosts/native-validation.txt). Initial observer
  diagnostics assumed one memory sample instead of the actual two epochs, then
  compared network RawMessage indentation byte-for-byte. Those observer guards
  were corrected; full stored capture byte/hash checks remain strict. No app or
  fixture changed.
- Main **11.9s** / **14.0s invocation**:
  [ledger](../artifacts/hosts/host-observed-fbac46c3b3/ledger.json),
  [log](../artifacts/hosts/main-second.log).
- Actor **30.6s** / **32.5s invocation**, eight phases and 400 snapshots:
  [acceptance summary](../artifacts/hosts/host-observed-fa0ffde660/bombadil/acceptance-summary.json).
  Explicit inventory, two defaults, local disclosure, source-free native preview,
  explicit native Run and retained host return end at four telemetry reads and one
  host plan.
- Ownership **4.5s**, accepted in the first invocation:
  [ledger](../artifacts/hosts/host-observed-e69e4254ba/ownership.json).
  An actual CPU source response is held after `route.fetch`; leaving Hosts for
  Services withdraws authority. Releasing that real response cannot add widgets,
  change the route, or issue another source read.
- Main logical telemetry effects are `0 → 1 → 3 → 4 → 5 → 6 → 7`:
  inventory, two default measurements, logs, spans, native Run, individual capture.
  Staging, resize, retained Back and frozen reopen add zero. The pure host plan
  and existing native preview planner are separate from telemetry reads.
- Full stored individual capture is **5,618 bytes**, ID
  `fdf520f858bf93d385cbd07147bd650bd46cad18e1c19b06c3ec87cf9d17004e`;
  SHA256 before/after
  `cea836acf2036872bb751cda541614c22b58faa0307a2bfeab0cf1d2a10fd668`.
  Both raw store copies are adjacent to the main ledger.

```sh
npx playwright test journeys/199-host.spec.js journeys/199-host-ownership.spec.js --workers=1
GOO_HOST_BOMBADIL=1 npx playwright test journeys/199-host-bombadil.spec.js --workers=1
```

The first main failed a harness geometry assertion because it measured hidden
children of closed evidence details as zero rectangles outside the widget. The
actual screenshot was inspected; visible desktop contents were contained. The
helper now explicitly checks displayed descendants while retaining visible X/Y
bounds. Full initial trace/frame/context remain under
`artifacts/hosts/first-failure-diagnostics`; no application change was made and the
accepted independent ownership case was not repeated.

The raw fixture also contains CPU-only, stale, missing-data, conflicting,
out-of-range fraction, unsupported counter and schema controls. Their presence in
source inventory is distinct from the two default streams actually rendered here;
this main does not claim every control received a separate browser native read.

Two initial actor attempts stopped at phase 2 because its unqualified
`[data-host-key]` selector matched the coordinator's own HTML data attribute rather
than the Open host button. A scroll-only correction did not resolve that collision.
The accepted actor uses `button[data-host-key]` and scrolls the real button into
view. Both partial traces and screenshots are retained separately; this was a
harness error, not evidence of an application offscreen-click restriction.

## Separate compact display replay

The first actual frames were readable but the host's declared 90px plot height
was ignored by the reusable component's hard-coded 150px CSS. The reviewed
`6a95d50` CSS change honors that variable while retaining a 150px default.
Only `value-timeline.css` was deployed; no Go rebuild, restart, query or arithmetic
change occurred.

A separate fresh ordinary fixture, `host-observed-eef8d1d896`, passed the bounded
display replay **6.5s / 9.0s invocation**:
[receipt](../artifacts/hosts/host-observed-eef8d1d896/display.json),
[CSS deployment](../artifacts/hosts/display-deployment.json). It performs one
actual host inventory read and two native measurement reads, plus one pure plan.
Both-width viewport resizing adds zero reads. All plots are computed 90px high;
visible card content passes X/Y containment, with readable UTC and exact
nanosecond tooltips. The final three frames above come from this fresh display
fixture and its new read clocks, not the earlier functional/capture fixture.
The actual mobile and whole-context PNGs were inspected after capture.

```sh
GOO_HOST_DISPLAY=1 npx playwright test journeys/199-host-display.spec.js --workers=1
```

No semantic main, accepted owner or actor campaign was repeated for this CSS fix.
Traffic remained running, temporary build/deploy helpers were removed, and all
compact ledgers, raw capture copies and initial diagnostics were retained.

## Narrow authoring deployment receipt

The later reviewed authoring-only fix `9649373` normalizes omitted host services
and exact attributes on fresh YAML reads. It does not change direct evaluation,
capture semantics or source execution. The updated runtime binary SHA256 is
`332cb5c25f1c201ed8b1338a6e2a176e86edbc1758266d90210815330bb4c4fa`.
Only the backend binary and two static authoring modules were updated; the query
container was not restarted. Actual module HTTP/MIME/byte hashes match both
runtime copies.

[Authoring deployment manifest](../artifacts/hosts/authoring-deployment-manifest.json)
records metadata readiness and a pure plan POST using the original accepted
request. Its complete JSON result equals the accepted plan. The original stored
capture remains byte-identical at 5,618 bytes and its original SHA256. This check
performed zero telemetry reads, zero custody writes and zero browser campaigns;
main, ownership, actor and display were not repeated. Exact temporary build/deploy
helpers were removed afterward.
