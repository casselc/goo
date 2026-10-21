# Shared evidence summaries

Service widgets, captured measurements and their investigation links use the same passive `goo-evidence-summary` element. It groups the population, scope, fixed window, source-record count, declared grain and binding into labelled facts. The result state and evidence mode are separate. A visible notice preserves source-completeness and independent-read limits; exact identities and clocks stay in a native disclosure.

The application presenter in `ui/evidence-scope.js` interprets typed measurement captures and validated service-widget results. The custom element in `ui/evidence-summary.js` only renders its supplied display snapshot. It never queries telemetry or metadata, chooses a namespace, infers health, calculates counts, refreshes evidence, navigates, copies or captures. Existing adapter buttons remain outside it.

```js
summary.data = {
  mode: 'frozen',
  state: 'unknown',
  facts: [{label: 'Population', value: 'Native duration histogram'}],
  notice: 'Source completeness remains unverified.',
  details: [
    {label: 'Source read', value: '2026-10-21T00:00:01.123Z'},
    {label: 'Stored custody time', value: '2026-10-21T00:00:02.456Z'},
  ],
};
```

## Display contract

- Supply `.data`, including on detached or pre-definition elements, or an HTML-escaped JSON `data` attribute for server-rendered HTML. Getters and setters detach snapshots.
- Modes are authored, fixed, frozen and live. Result states are supplied explicitly: observed, unknown, unavailable, partial, no_data, invalid, pending, stale, comparable, incompatible or ok. Mode never determines state or connection continuity.
- Accept one to eight visible facts and up to 32 detail rows. Labels are nonempty strings up to 120 characters, values up to 8,192 characters, and the notice up to 2,048. JSON attributes are capped at 512 KiB. Invalid input shows unavailable and clears the old mode/facts; rows are never silently truncated.
- All strings render as text. The component owns its children; place adapter links/buttons beside it. It has no action/event API and starts no subscriptions. Disclosure is the only built-in interaction.
- Native definition lists preserve reading order and wrap full IDs/values without widening the document. Mode/result labels are neutral; no healthy status is inferred. The host is a labelled group, without automatic live announcements.
- A valid data update preserves whether the exact-detail disclosure is open and restores focus to its summary if that summary owned focus. Other invalid/structural updates may remove the disclosure; this is not keyed interactive-child reconciliation.

## Adapter semantics

The shared measurement presenter distinguishes wildcard, missing, present-empty and literal namespaces using the existing service-identity helper. It retains exact counts as strings and exact nanosecond question bounds in details. A pinned complete stream binding is distinguished from independent matching streams; lifecycle bindings retain subject/tenant terminology.

Source read and stored custody are separately labelled from their explicit fields. A missing source-read timestamp stays unavailable and never borrows storage time. Full lifecycle results may additionally supply business evaluation and knowledge-cutoff clocks. Missing receipt-only metadata is left unavailable: the existing metadata endpoint has no evidence basis or full evaluated diagnostics, and the presenter does not fetch another result to fill them in. The full capture and its native attribute viewers remain accessible through adapter-owned buttons.

Linked question evidence retains visible scope-difference guidance, including a matching question's independent-read caveat. Matching windows never imply a shared source snapshot. The same summary also presents a separate native capture card in investigation notes. Non-measurement evidence retains its existing presenters in this milestone.

The capture workspace suppresses its old repeated scope paragraph and the measurement renderer's duplicate count/read paragraph when the shared summary owns those facts. The operational status remains available for refresh, attachment and error feedback. Reusable result rendering elsewhere retains its existing metadata by default.

```mermaid
flowchart LR
  C[Validated full measurement capture] --> A[Application evidence presenter]
  M[Existing stored receipt metadata] --> A
  A --> D[Detached display facts and labelled clocks]
  D --> E[Passive goo-evidence-summary]
  E --> F[Capture workspace]
  E --> I[Linked investigation evidence]
  S[Validated record and native service results] --> A
  E --> W[Compact fixed service widgets]
  B[Adapter-owned open, refresh and capture buttons] --> F
  B --> I
```

## Executable storyboard

Journey 82 injects real OTLP histogram exports, creates an immutable measurement capture and links it into a question plus a separate capture card. It opens the full capture, inspects exact read/storage/basis values, follows the saved investigation, verifies receipt-only limitations and matching scope, and opens the same disclosure while preserving an unsaved finding. Mobile checks verify internal wrapping, with no source reads during inspection and no change to the retained capture.

[Full capture summary](journeys/screenshots/82a-capture-evidence-summary.png) · [Linked native evidence](journeys/screenshots/82b-linked-evidence-summary.png) · [Mobile evidence summary](journeys/screenshots/82c-mobile-evidence-summary.png)

Journey 83 covers the clearly synthetic standalone gallery, SSR attributes, pre-definition properties, detached getters, safe text, invalid data, keyboard disclosure, update focus and mobile layout. Pure tests cover bounds, large counts, namespace distinctions, exact timestamps and absent-clock handling. Observed Quint replay checks the actual capture/linked summaries and rejects eight corruptions; the saved-evidence Bombadil actor explores the two real adoption surfaces.

[Standalone synthetic reference](journeys/screenshots/83-synthetic-evidence-summary.png)

Service widgets now use the same element through their own typed display adapter. Comparison sides remain a later adoption candidate. See the [whole-product sweep](WHOLE_PRODUCT_COMPONENT_SWEEP.md) for the later selection/detail and complex-composition work.

## Compact service cards and readable time

Service widgets lead with their values and charts. A compact “Fixed question · result · Scope and evidence” disclosure opens the shared summary, with one population fact beside the fixed-question/result labels. Scope, fixed window, exact count, grain, source read and evidence basis share one disclosure, rather than repeating source metadata paragraphs. The enclosing service keeps its fixed scope; each widget retains its independent read. A fixed widget is not a stored capture or a continuous subscription. Capturing explicitly performs another read. Record results use their `capturedAt` source-read field; native results use `readAt`. Neither becomes a custody time.

Default summary spacing and card padding are tighter. The overview keeps independent-read, unverified-coverage and capture-reread guidance visible; per-widget detail retains the same limits. Native cards still distinguish the latest interval/endpoint from complete-window coverage and never imply a fleet aggregate. Full stream identities, source warnings and typed question JSON remain inspectable.

Timestamp rows can supply `format: 'unix-ns'`, `'unix-ns-range'` or `'utc'` (RFC3339 UTC clocks). Their raw string `value` remains unchanged in the display snapshot. The element shows millisecond UTC dates and exposes exact Unix nanoseconds in a native tooltip, with a keyboard-focusable value. Scalar `<time>` elements retain a nine-digit fractional `datetime`. Invalid formats or out-of-range dates fail closed. These formats apply only to absolute timestamps: durations, units and counts retain their existing semantics. Native stream tables and selected intervals use the same timestamp formatter. This is a targeted adoption; other legacy diagnostic JSON and raw exporters retain exact source fields.

Journeys 59 and 75 exercise real record/native service cards, disclosure without source reads, namespace and fixed-mode semantics, source failures, exact-stream navigation and mobile containment. Their screenshots now show the compact summaries. Journey 83 additionally checks UTC text, exact tooltip/`datetime`, keyboard focus and invalid formatting. Journey 82 refreshes the capture/investigation summary frames.

[Service overview](journeys/screenshots/59b-automatic-service-widgets.png) · [Mobile service overview](journeys/screenshots/59c-mobile-service-overview.png) · [Native service overview](journeys/screenshots/75a-native-service-overview.png) · [Mobile native service card](journeys/screenshots/75c-mobile-native-service-card.png)

The service summary sits inside the single widget-level scope/evidence disclosure, after values, charts and actions. Source warnings and typed question JSON nest there too. This keeps the default card focused on observations while retaining the complete read context. Cards no longer stretch to fill the height of a longer neighboring card.

Latency readouts now use inline Mean/p50/P95/P99 label/value pairs with compact spacing and wrapping. The percentile-position rail and exact counts/values stay in the keyboard-accessible Latency details disclosure. No duration calculations or query behavior changed.

See [comparison presentation](COMPARISON_PRESENTATION.md) for aligned scalar pairs, compact metric matrices, frozen-side summaries and UTC tooltip adoption.
