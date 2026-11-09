# Whole-product design checkpoint — 384

2026-11-08. Read-only product/design review during the bounded fixed-result layout
consolidation. These recommendations are proposals, not a second execution queue.
The root checkpoint roadmap will choose the next work after 384 acceptance.
Runtime outcomes belong in [validation](../../VALIDATION.md); no source queries or
heavy tests were run for this review.

## Product center

Goo helps an investigator see current telemetry, scope a population, compare it,
follow a useful record or trace, and preserve or reuse the question and evidence.
A useful default or reusable view should shorten that loop. More widgets or another
permanent builder do not automatically improve it.

The present product already covers much of that loop:

| User task | Implemented capability | Boundary that still matters |
| --- | --- | --- |
| See and scope current activity | Log/span Activity, qualified field gestures, columns, charts, paused inspection and retained continuations | Held rows and selected occurrences are not full-population measurements; acquisition and resume remain explicit outside the fresh-entry policy. |
| Isolate a distinct population | Fixed filters, grouping, selected/rest comparisons, fresh-window comparisons and ranked field/value explanations | Explanations use explicitly chosen fields and independent full aggregates; descriptive change is not significance or cause. |
| Follow a trace across services | Checked trace search, qualified witnesses, waterfalls, page timing and captured-search follow-through | Reported IDs do not prove unique invocations, complete traces or causal order. |
| Keep and reuse useful work | Exact kept occurrences, immutable captures, question cards, reusable Activity views, packages and view templates | Intent, source-read basis and custody remain separate; configuration supplies no execution or completeness authority. |
| Inspect service and resource signals | HTTP/RPC server, DB client and messaging views; supported native measurements; host/pod/container observations | These are reported observations, not invented health, verified topology or complete resource coverage. |
| Annotate examples | Tags/notes plus captured-record categorical tasks, immutable decisions, historical manifests and pinned JSONL export | Existing reviews use capture occurrences and exact task IDs; they are local judgments, not authenticated consensus or trained classifiers. |
| Customize a representation | Checked record display declarations, record tables/lists, timelines, facets, authored templates and reusable native histogram displays | Contributed executable kinds and a WYSIWYG editor are not delivered; specialized execution is not missing merely because one display profile is not reusable. |

Evidence: [current capability index](../CAPABILITY_GUIDE.md),
[population explanations](../POPULATION_EXPLANATIONS.md),
[kept observations](../KEPT_OBSERVATIONS.md),
[automatic dashboards](../AUTOMATIC_DASHBOARDS.md),
[view templates](../AUTHORABLE_VIEW_TEMPLATES.md),
[saved display declarations](../SAVED_DISPLAY_DECLARATIONS.md) and
[captured-record review](../CAPTURE_RECORD_REVIEW.md).

## Ranked next task cuts — proposed

### 1. Breadth: identify a slow outgoing HTTP dependency

User task: “My service is slow. Which outgoing web destination or method should I
inspect?” Add one current-OTel HTTP client observation lens through the existing
service view planner and entry controls. Preserve the exact service, namespace,
environment policy and fixed window; group reported destinations/methods, show
full-population duration and recorded span status, then use existing explicit
trace continuation and retained Back.

The present planner's profile allowlist in
[dashboard.go](../../exploration/dashboard.go) includes HTTP and RPC server
profiles, DB client and messaging consumer/producer profiles, but no HTTP client
profile. This is a concrete everyday observability gap beyond the earlier deferred
candidate list. The existing DB client slice demonstrates the dependency workflow;
HTTP client coverage can reuse that structure rather than introduce another
navigation surface or query engine.

Before implementation, pin and check the applicable official OTel convention.
Metadata hints must remain distinct from positive same-record evidence. Retain
missing/empty destination populations. A client-span count is not a count of
unique remote requests, and client observations do not establish remote server
health or causality. Native HTTP client duration binding is a separate capability
unless independently admitted; do not infer it from span keys or convert units.

### 2. Depth: reuse a native gauge display in another compatible scope

User task: “I have a useful gauge presentation; apply its display choices to this
other selected stream without copying the data or changing its question.” Start
with a bounded service-bound, per-stream gauge configuration and the existing
native presentation/metadata-discovery path. Inspect and Apply should remain
configuration-only. Keep the destination's complete source binding, exact unit,
stream, environment and bounds; Run reads independently. Unknown intervals and
gaps remain unknown rather than zero or connected observations.

Native gauge execution and presentation already exist. Reusable native display
configuration currently has an explicit histogram-only contract:
[native histogram reuse](../NATIVE_HISTOGRAM_VIEW_REUSE.md) supports service-bound
per-stream explicit histograms in seconds or milliseconds, and
[native display discovery](../NATIVE_VIEW_DISCOVERY.md) lists that family. Extending
one adjacent family would deepen the composition promise without pretending
native execution itself is absent.

Keep counter/rate policy, resource-native variants and arbitrary unit conversion
outside the first gauge cut. Configuration must not carry donor observations,
read clocks, selected exports or a transferable completeness witness. Reuse the
existing passive kinds where compatible; a plotting-vendor escape hatch is not
needed to establish this task.

### 3. Consolidate: make capture-family admission coherent

User task: “Reopen this evidence and trust that every view interprets the same
family, scope and result.” Close the current capture question/page union
consistently across Go and JavaScript, using the existing generated admission
corpus and measured journeys. Remove family shortcuts that bypass necessary
exclusive-arm checks. Keep this a bounded current-wire cleanup rather than a
universal validator, new registry or compatibility layer.

The recent 383 generated tests found a genuine hole: relabeling search evidence
as an ordinary capture reached a JavaScript record-family fast path while Go
refused it. That new arm is now closed in both directions; the wider current
family shortcuts remain visible in
[saved-capture-protocol.mjs](../../ui/saved-capture-protocol.mjs).
The [383 review](CAPTURED_TRACE_SEARCH_REVIEW_383.md) and validation record provide
fresh evidence for this reliability cut. Preserve each family's distinct capture
meaning, clocks, bounded coverage and navigation authority while simplifying the
shared boundary.

## Candidates considered but not automatically promoted

- **Natural example review:** a contextual task chooser and verified kept-occurrence
  target would join the existing Keep and captured-record review workflows. The
  missing bridge is real: current review requires `{captureId,index}` and an exact
  task ID. It deserves a coherent later cut; do not create a fresh Explore capture
  to pretend it is the original kept observation.
- **Log-pattern display grammar:** log patterns, counts, timelines and comparisons
  already exist. The remaining gap is checked `body_pattern_v1` display admission,
  not a new clustering system. This is a useful alternative if reuse friction is
  observed more often than outgoing dependency investigation.
- **Native counter display reuse and service exemplar adoption:** valid later
  depth, with separate rate/interval/stream or citation authority. Existing Saved
  histogram authority does not grant service-widget authority.
- **Wider initiatives:** alert delivery, broad database/broker health, tenancy,
  contributed executable kinds/WYSIWYG, warehouse lowering and new stores remain
  independent decisions. None is a prerequisite for the task cuts above.

## 384 layout acceptance review — accepted

Final frames reviewed on 2026-11-09: desktop ordinary fixed records, narrow
Activity and ordinary fixed records, the narrow Saved-search continuation, and
narrow Saved records in both Table and Span waterfall modes. The actual report
contains ten desktop/narrow frames across these four contexts. All retain a full
first actual record or span inside the natural scroll viewport, with scope and
actions visible and no horizontal page pan.

The matched baseline uses the same fixed source fixture. First-record vertical
bounds show the change directly (desktop viewport 1000 px high; narrow 844 px):

| Context | Desktop before → after | Narrow before → after |
| --- | --- | --- |
| Ordinary fixed table | 1199–1322 → 500–624 | 1433–1521 → 550–638 |
| Saved-search child table | 1202–1325 → 564–688 | 1419–1507 → 643–730 |
| Saved record table | 1263–1386 → 517–641 | 1791–1878 → 603–691 |
| Saved record waterfall | Not part of matched baseline | Final first full span: 608–814 |

The compact applied-scope/count and UTC window remain readable. Rerun, Edit and
Preserve form one small action row; the Saved child retains its origin/Back bar.
Optional related-population controls, examples and raw receipts no longer precede
the data. A single-visit trail no longer spends space on disabled Back/Forward.
The Saved view picker is inline. The waterfall retains its visible relative-time
scale and unknown-coverage qualification, exact UTC context, span content and
fresh-read action; a supplied-summary option removes duplicated counts without
changing standalone component semantics.

The first actual candidate was correctly rejected: its narrow first row extended
past the viewport, and the direct Saved waterfall needed further structural
consolidation. The final improvement comes from DOM ordering and meaningful
context disclosures, not tiny text or a scripted scroll to the records.

A nonblocking future measurement opportunity remains: desktop record actions
stack vertically and make each row approximately 123 px high. This does not block
384—the full first item and all four fixture rows are visible—and it is not an
additional task in this checkpoint queue.

Reviewed examples:
[ordinary desktop](../journeys/screenshots/384-ordinary-desktop.png),
[ordinary narrow](../journeys/screenshots/384-ordinary-390.png),
[Saved-search narrow](../journeys/screenshots/384-saved-search-390.png),
[Saved record narrow](../journeys/screenshots/384-saved-record-390.png) and
[Saved waterfall narrow](../journeys/screenshots/384-saved-record-waterfall-390.png).
Functional ownership, no-read actions, dirty/invalid editor behavior, actor/model
outcomes and rejecting controls are recorded in validation; screenshot appearance
alone does not establish those guarantees.

The root planner adopted the finite 385→386→387 sequence in the
[November 9 roadmap](../ROADMAP_2026_11_09_CHECKPOINT_384.md) after the final cold-entry
acceptance. The final report is `b8bba76d-200d-459a-8dc1-3a061a7843f5`; prior passing
geometry used the same question but retained an Activity navigation utility. Cold
entry is now independently asserted, with the same first-item bounds.
