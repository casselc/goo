# Explain an Activity window

Implemented in `cdd5a88`, with fixed-visit ownership repaired in `a6533a6`.
Real-source, native/model, Bombadil and regression acceptance is delivered after
independent review. Activity stages the existing
population explanation from its last completed stream window and applied selector. Opening, brushing,
editing fields, loading saved metadata and returning to paused Activity do not
query telemetry. **Run explanation** explicitly reads both fixed populations using
the existing bounded Go engine; no new endpoint, evaluator or subscription is added.

## Use the workflow

1. Start an applied log or span Activity scope and wait for a completed read.
2. Open **Investigate → Explain this window**. Activity pauses; the new panel
   stages a fixed pair without reading either population.
3. Brush the held-observation time chart to select a subwindow, or keep the entire
   completed window. The previous equal-width window is the default comparator.
4. Optionally select **Saved capture scope**, enter its exact capture ID and click
   **Load baseline scope**. Inspect its original clocks and query before Run.
5. Choose the explanation fields and click **Run explanation**. The existing
   compact paired table shows full-population differences and per-field failures.
6. Stage a value filter, open its side question and Run to inspect records. Return
   to paused Activity restores the held window and visible Investigate trigger.
   **Save observed explanation** preserves the exact result; reopen/reuse uses the
   existing saved explanation flow.

## Scope and interaction

**Investigate → Explain this window** is available for clean applied logs or spans
with a completed 60-second window, including a zero-row window. Entry freezes that
exact window and pauses the stream and optional latency updates. It does not use
wall-clock time or the timestamp extent of the held rows to define membership.

A small time chart shows held observations within that window. It describes the
bounded local buffer, not full population counts, rates or completeness. Brushing
stages a fixed subwindow. The default baseline is the immediately preceding window
of the same width, with the same signal, environment and complete selector. Exact
UTC bounds remain inspectable; no statistical or causal claim is inferred.

A saved baseline scope can instead be loaded explicitly from an immutable raw
record capture in the same signal/environment. The baseline retains that capture's
original query window and selector, even if widths differ or populations overlap.
This is **scope reuse**: Run rereads both windows, and late data may change baseline
counts. It does not use frozen saved field counts. The original capture link and
source/custody clocks are shown separately from the explanation's new read clocks.
The capture reference is UI context; the saved explanation retains its request and
observed aggregates, without claiming durable storage of that originating link.

One to six qualified fields, exact ranking, missing/empty distinctions, partial
field limits, staged side filters, observed Save and reusable captured requests use
[the existing explanation contract](POPULATION_EXPLANATIONS.md). Saved explanation
publication promotes the exact held result without rereading telemetry. Scope and
baseline edits withdraw obsolete results/tokens. Scope Apply, resume and view
replacement invalidate staged work. Leaving or hiding cancels pending work; return
to the same paused Activity window does not subscribe or query automatically.

```mermaid
sequenceDiagram
    actor User
    participant Activity
    participant Editor as Shared explanation owner
    participant Store as Capture store
    participant Go as Go / ClickHouse
    User->>Activity: Explain completed window
    Activity->>Activity: Freeze applied scope; pause stream/latency
    Activity->>Editor: Adjacent equal-width baseline + selected window
    User->>Activity: Brush held-observation time chart
    Activity->>Editor: Stage exact subwindow and preceding window
    opt Saved baseline scope
        User->>Activity: Load explicit capture link
        Activity->>Store: GET immutable capture
        Store-->>Activity: Original query and clocks
        Activity->>Editor: Stage original baseline query
    end
    User->>Editor: Run explanation
    Editor->>Go: Read both fixed full populations
    Go-->>Editor: Bounded exact differences + held result token
    User->>Editor: Save observed explanation
    Editor->>Store: Promote exact result token; no telemetry read
    User->>Activity: Return to paused Activity
    Activity->>Activity: Restore held rows and focus; no subscription
```

## Acceptance gates

AE1 delivered in `cdd5a88`: source-free query/window helpers, compact Activity entry and shared explanation
ownership. Refuse dirty/no-window/mixed scope and stale responses; preserve legacy
Activity/latency behavior. Pure Node and independent source/design review.

AE2 delivered after actual acceptance: real ordinary OTel logs/spans, full-population counts beyond the held page,
time brush, previous/saved-scope baselines, late-arrival observed Save, side drilldown
and paused return, scope/resume/navigation cancellation, desktop/mobile screenshots,
Hegel properties, bounded Quint replay/rejecting controls and Bombadil actions.
Update README, diagrams, planning, limitations and validation; review, commit, push.

## Storyboard 178

The ordinary OTel [executable journey](../journeys/activity-explanation.spec.js)
passes for spans and logs, including full counts beyond the held page, checked
brushing, original saved scopes, exact observed Save/retry, fresh reuse, side
inspection and source-free paused return with focus restoration. A separate real
response-holding journey covers resume, navigation, dirty scope, saved-scope GET
retry and delayed filter plans. These are actual captured application frames:

| Frame | Workflow |
| --- | --- |
| [Span explanation](journeys/screenshots/178a-spans-activity-explanation.png) | Applied completed window, sampled chart, saved scope and full paired differences |
| [Log explanation](journeys/screenshots/178a-logs-activity-explanation.png) | The same workflow over ordinary log records and severity |
| [Mobile span explanation](journeys/screenshots/178b-spans-mobile-explanation.png) | Field/value context and both labeled population shares |
| [Mobile log explanation](journeys/screenshots/178b-logs-mobile-explanation.png) | Shared compact layout without clipped side values |

A supplemental real-source check inserts one late baseline record: the new
explanation baseline advances from 240 to 241, selection stays 248, and the full
stored capture and public response remain unchanged. This demonstrates saved
scope rereading rather than frozen-count comparison. Native checking validates
complete stored capture hashes/membership separately from the 100-row public page;
JSON transport formatting is not mistaken for record-byte identity. See
[VALIDATION.md](../VALIDATION.md) for exact executed results, finite model limits
and the final regression/review status.
