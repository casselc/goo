# Live Activity entry

The S1 entry composition uses the existing bounded log/span tail. A fresh `/` or
`#explore` visit automatically connects one log monitor for the lab's explicit
`local` environment. The environment input and Apply scope start another exact
scope; Logs/Spans selects one active signal. Missing scope asks for an environment.
This does not discover or query every environment in a deployment.

The normal entry has one compact scope/action bar: environment, signal and Pause.
Apply appears only while editing scope; a draft shows the currently applied scope.
Coverage state and local counts share a compact metadata row. Repeated scope pills,
workspace/source banners and the visible table caption are removed from this mode
(the caption remains available to assistive technology). Mobile navigation is
collapsed behind an accessible Menu toggle; other product areas remain reachable.
Opening or closing that menu performs no source read.

Rows and local buffer counts appear before the advanced fixed-question workspace.
UTC time is compact, with the full UTC timestamp and exact nanoseconds in the
row tooltip. Logs show body/severity; spans show name/duration/reported status.
Inspect uses the held record and attributes without a source request. Open trace
is a separately named source continuation using that row's original read window.
No ancestry is inferred from arrival order or a shared trace ID; richer reported-
parent presentation remains S2 work.

## Intent and ownership

`GooStreamEntry` owns the fresh-entry intent, applied environment/signal, mode,
local inspector and lifecycle. `EntryGate` permits one automatic claim for a fresh,
visible, active records entry. Saved/fixed links, saved exploration views and the
explicit `#questions` builder never claim a monitor. Repeated route events,
returning from Services and returning from Measurements do not restart a paused
connection. Open live activity is an explicit mode change.

The monitor uses `GooTail` with a compact composition. It does not run the
population planner, a preliminary query or an aggregate before opening the SSE
stream. Display elements remain passive. Scope edits stay drafts until Apply;
applied scope remains visible while a replacement is pending. New scope and
signal generations abort the preceding reader and cannot accept its late batches.

```mermaid
flowchart LR
  URL[URL and visit intent] --> Gate[Fresh-entry gate]
  Gate -->|fresh and visible| Monitor[One active log or span monitor]
  Gate -->|saved or fixed| Fixed[Stored or staged viewer]
  Monitor -->|Inspect: local| Held[Paused buffer and inspector]
  Monitor -->|scroll / leave / hide| Held
  Held -->|explicit Resume with gap| Monitor
  Held -->|Include / Exclude: local draft| Draft[Qualified filter chips]
  Draft -->|Apply: new scoped read| Monitor
  Draft -->|Open fixed investigation: no read| Staged[Fixed question draft]
  Staged -->|Run explicitly| Context[Fixed source results]
  Held -->|Open trace: named read| Context
  Fixed -->|Open live activity| Monitor
```

## Pause, resume and failures

Scrolling away from the live edge pauses transport, not merely auto-scroll.
Focusing or inspecting a record, leaving Explore, selecting Measurements or a
hidden-document event also pauses. Resume explicitly starts a new recent window
and exposes a continuity gap. The inspector is cleared on explicit resume so a
cached connection-local observation key cannot masquerade as the new read.

Reconnects retain last-good rows until a successful replacement batch arrives.
HTTP/source failures remain stale/unavailable with a gap; retry does not present
an empty healthy view. Compact mode offers Pause reconnecting as well as explicit
resume. Scope replacement intentionally clears the old scope rather than mixing
its rows under new labels.

This is a bounded recent-event-time monitor: 60-second lookback, 2-second polls,
100-row/2 MiB display buffer, 1,000-row/512 KiB source-poll cap and four global
prototype stream slots. Counts are local observed/retained/evicted observations,
not population totals, request rates or evidence of telemetry completeness.
There is no ingestion replay guarantee. Last-good rows can refer to an earlier
completed poll while the next connection is unavailable.

## Deliberate limits and migration seam

The first S2 interaction cut adds Include/Exclude through the shared typed
attribute viewer, removable filter chips and explicit Apply. Predicates are exact
recorded strings, qualified by column/resource/attribute namespace, AND-combined
and bounded to 16. Empty string is a recorded value; absence is not coerced to it.
Exclude uses the existing `ne` predicate: the field must exist and differ.

Staging or removing filters pauses transport and retains the held rows/inspector.
Draft chips are marked; the previous applied environment/signal is shown and its
selector is available in the tooltip. Apply replaces the buffer with a new scoped
read. Resume refuses an unapplied scope; changing signals clears signal-specific
filters. Opening a fixed investigation from the inspector stages the drafted
predicate in the selected row's original 60-second read window. It does not fetch;
the existing Run action explicitly rereads that fixed question. It is not a
capture or an assertion of complete source coverage.

Supplied-row ancestry, a common record presenter across all surfaces and retained-
parent return after a fixed investigation remain S2/A3 work.
The legacy advanced editor is a temporary compatibility consumer; the
[whole-product reviews](STREAM_FIRST_WHOLE_PRODUCT_DESIGN.md) and
[active queue](IMPLEMENTATION_PLAN.md) name its replacement/removal gates.

The monitor does not publish captures or annotations. Exact retention of selected
tail observations requires server-owned membership/read-window capabilities.
Current tail handles are not native `resultToken` values. Existing fixed-page and
saved-capture retention remains available through its existing explicit paths.

Journeys 102/103 exercise real OTLP logs/spans, startup request counts, inert fixed
and saved links, scope/mode changes, local inspection, scroll anchoring, capacity
failure and explicit gap-bearing resume. The hidden-document lifecycle input is
injected in 103; it is not a claim of a native OS backgrounding campaign. Hegel
checks the actual JavaScript entry gate across state/swarm transitions. Quint
validates observed entry traces and rejecting controls. See [validation](../VALIDATION.md)
for actual completed runs and final image identity.

The HTTP proxy watches downstream disconnects and closes the upstream SSE socket
promptly, releasing its existing admission slot. `python3 scripts/check_tail_disconnect.py`
checks eight rapid aborts and the unchanged four-reader limit (fifth reader: HTTP 429).
Run this local-lab check without another active monitor.

Journey 104 follows four real generic OTel logs (west, east, empty and missing
attribute) through Include, Exclude, chip removal, explicit Resume and a staged
empty-string fixed query. It checks exact typed selectors and request counts;
[frames and storyboards](journeys/README.md) show draft, mobile filtered stream and
fixed-question staging. Hegel state/swarm campaigns run the actual JavaScript
filter model; observed browser transitions use the entry Quint protocol. The
Bombadil filter actor scopes the lab, inspects a record and stages a predicate.
