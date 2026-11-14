# Choose unfamiliar telemetry from Activity

Activity's **Find data** control sits beside Environment. You can discover
what was reported before knowing any labels, without a manifest or sample package.
In generic mode the chooser is highlighted until a scope is chosen; opening it
performs no source read.

1. Open Find data. Read available signals if you do not know which signal exists;
   Choose logs/spans stages the signal without reading environments or restarting Activity.
2. Read a bounded window of environment labels across that signal's services.
3. Preview a named, empty or unreported environment.
4. Stage in Activity, then Start stream.

The preview uses the effective current/legacy environment policy. Empty labels,
missing labels and conflicting reported keys remain distinct. Stage preserves
existing Activity filters; it neither drops them nor starts a stream. The preview
may therefore contain records excluded by those filters when streaming starts.

The existing discovery and preview endpoints are reused. No all-environments
selector or direct-stage shortcut is introduced.

```mermaid
flowchart LR
  Entry[Blank generic Activity] -->|local| Chooser[Find data]
  Chooser -->|explicit source read| Labels[Observed label groups]
  Labels -->|explicit source read| Preview[Record preview]
  Preview -->|local stage| Ready[Typed scope ready]
  Ready -->|Start stream| Live[Rolling event-time records]
  Live -->|Pause| Held[Held records]
  Held -->|Resume| Live
```

Named preview buttons expose the complete environment value in their accessible
names. Scope cells are row headers. Explicit Pause/Resume keeps keyboard focus on
the available counterpart; automatic pauses leave the user's focus alone.
Unchanged stream statuses no longer rewrite the live region, and idle local Find
counts do not announce background updates. Changing Find results can still
announce; this is not a complete assistive-technology audit.

Native controls must remain available in the active disclosure. Hidden, inert,
disabled or stale discovery actions cannot start reads or stage a scope. Returning
from a withdrawn request releases its private busy state without publishing data.

## Metrics-only data

When Read available signals finds a usable metric family, **Find metric scopes**
opens the existing Measurements workspace locally, retaining that read's exact
window. Read metric scopes explicitly, choose an environment/service/namespace,
then Discover instruments. The direct Measurements entry also supports this flow
without knowing any labels. Selection and package preview do not read telemetry;
Run keeps its existing execution requirements.
See [metric scope discovery](METRIC_SCOPE_DISCOVERY.md) and its
[executable storyboard](journeys/metric-scopes.md).

## Remaining work

[Source capability checks](SOURCE_CAPABILITIES.md) now identify available
signal families and allow partial-family service/instrument inventories.
Record environment discovery still uses the selected Logs/Spans signal; metrics
use their separate qualified scope discovery above. Streaming still uses the
latest 60 seconds of event time; sparse and
delayed-event recovery is separate work. See the
[ad-hoc review](reviews/ADHOC_OTLP_REVIEW_2026_11_14.md) and
[current roadmap](ROADMAP_2026_11_14_CHECKPOINT_430.md).

[Executable storyboard](journeys/unknown-data-entry.md) ·
[Independent acceptance](reviews/UNKNOWN_DATA_ENTRY_430.md).
