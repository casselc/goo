# Prose and guidance: designer review after the 459 polish

Status: the Activity custody disclosure was implemented during this review; its
final browser validation is recorded in [validation](../../VALIDATION.md). All
other content migrations and walkthroughs below are proposals.
Reviewed the polished Activity, fixed-record, Services and Saved 459 desktop/mobile
frames, the 458 chronology and 455/456 metric frames, plus their owning UI sources.
The designer changed this assessment only; the agreed Activity disclosure was
implemented by the primary agent.

## Decision

Keep **facts about this result** and **the consequence of the next action** visible.
Move **definitions of recurring concepts** and **how to perform a workflow** into
contextual help. The current interface often repeats all three as paragraphs.

The first agreed change is now a native disclosure titled **Local observation ·
not saved** in the Activity inspector, containing the existing staging paragraph.
The custody state stays visible, and the full explanation remains one action away.
Do not move the actual log message, selected scope, error, or recording validity into
help. Increasing every inspector's height would conceal the copy problem rather
than resolve it; consider viewport sizing separately after this extraction is measured.

## Concrete content inventory

| Surface and source | Current example or information | Keep on screen | Move / replace |
| --- | --- | --- | --- |
| Activity inspector, `ui/stream-entry.mjs` | “Stages a fixed question for a separate source read; held rows are not saved evidence.” | **Local observation · not saved** and the Analyze action | **Implemented:** original explanation in a closed native disclosure. **Proposed:** a short first-use Observe → Analyze → Run guide teaches the transition. |
| Activity, `ui/tail.mjs` | Long explanation of the rolling 60-second window, 2-second polling, late records, multiplicity and saved evidence | Live/Paused, loaded/evicted counts, actual continuity gap, active scope | Existing **Tail scope and limits** is the right home for mechanics. A small `60s window` fact can identify current bounds; no extra paragraph beside every record. |
| Activity Find, `ui/tail.mjs` | “Record text, IDs and chosen fields” / Find does not change query scope | Search label and matching/loaded counts; **Unapplied scope** when filters change | Find-versus-source-filter explanation in local Help and the first-use refine guide. Do not make the distinction depend on a title tooltip. |
| Chronology, `ui/held-chronology-view.mjs` | Independent source facts and “All held rows · lane Find not applied” | Both lane counts/states; the Find-boundary warning whenever a lane Find exists; unplaced/evicted/gap facts | Explain why windows differ and how ties are ordered inside the already collapsed source facts / **How chronology works**. Never replace two source facts with one “healthy” light. |
| Services, `ui/inventory.mjs` | “Counts retain signal grain. Metric counts are exporter data-point records…” | Column units (**Span records**, **Log records**, **Metric data points**), actual **Not observed**, discovery window | Move the general explanation beneath **About signal counts**. The existing units already do most of this work visually. Keep missing signals distinct from numeric zero. |
| Services declarations, `ui/inventory.mjs` | “Bindings are current declarations, not verified emission, deployment or conformance…” | **Declared · Not observed in this window** and number of bindings | Keep detailed limitations inside the existing declared-bindings disclosure. One-time definition of declared versus observed belongs in service Help. |
| Fixed questions, `ui/exploration.mjs` | Draft/applied/running/cancelled messages and resolved scope | Current state, requested environment/window, Run/Rerun label, source disagreement | Keep actual state transitions inline. Move general explanations of independent rereads and lookback resolution to a small Read semantics help panel. |
| CEL/filter editing, `ui/filter-editor.mjs`, `ui/cel-expression-editor.mjs` | Parent predicate, applied predicate and chip can repeat the same condition | One authoritative applied-condition representation; edited slot/draft state and compiler diagnostics | Put the exact parent context in a disclosure when it duplicates the visible applied condition. Keep parent/slot differences visible when they genuinely differ. A CEL mini-guide is optional advanced help. |
| Charts, `ui/chart-frame.mjs`, gauge/histogram views | Repeated coverage and plotting-mechanics sentences | State label, units, actual window, plotted/unplotted counts, incompatible/invalid reasons | Retain the new text-plus-shape state markers. Definitions and keyboard/selection instructions belong in an accessible **Chart help** disclosure, not only `title`. |
| Comparisons, comparison/preservation owners | Educational “save both questions / independent reads” copy above data | Baseline/current labels, both windows/read clocks, comparable/incompatible state, denominator/unit | Use an aligned two-side header and a compact read-lineage diagram. General save semantics belongs beside Save in contextual help, not above the comparison chart. |
| Saved, `ui/index.html`, capture and retention views | Stored selection versus live question, refresh creates a revision | **Stored evidence**, capture title/time, current revision and explicit Refresh action | Put the revision model and “opening does not run a query” lesson into Saved Help. Keep the consequence of Refresh available at the action. |
| Retention failures, `ui/retention-status.mjs` | Commit outcome unknown, proof expired, unchanged retry request | Actual outcome, expiry/conflict/unconfigured reason and safe next action | Do **not** hide these in a tutorial. A details panel may explain recovery mechanics, while the uncertain outcome and Retry consequence remain visible. |
| Operations, `ui/index.html` | Package introduction, retry semantics, local sample :8090/:8091 links | Real package prerequisite/empty state and **Invocations (includes retries)** beside counts | Move sample-app instructions to Examples or an on-demand Operations walkthrough. Do not let lab-specific examples define the general product. |
| Workspace footer, `ui/index.html` | Visually hidden “Captured records are evidence, not a guarantee of coverage…” remains available to assistive technology | Per-result coverage state already shown in the result | Put generic semantics in one equally accessible **Data semantics** Help entry. The hidden global paragraph is not a visual-density issue, but is a poor substitute for contextual result facts. |

## Visual language, without hiding meaning

Use a small consistent vocabulary rather than converting every sentence to an icon:

- **Text plus redundant shape:** neutral ring for unknown, neutral filled marker
  for observed, warning diamond for partial/stale/unavailable. Observed never means
  healthy. Keep labels, including in forced-colors modes.
- **Compact factual pairs:** `Window [from, to)` and `Read at …`; `Stored at …`
  only for custody. Show the actual values, not abstract provenance essays.
- **Scoped chips:** Environment, service/namespace and applied predicates. Chips
  are facts or active controls, not decorative explanatory badges.
- **Read lineage:** `Held observation → staged question → explicit Run → result`
  and `result → Save → stored revision`. Use labelled arrows and real current
  state; a tutorial animation must not execute any transition.
- **Examples next to the concept:** equal-count bars for unequal histogram bins
  can demonstrate “counts, not density” in Help. Do not alter the actual chart to
  illustrate a lesson or mix synthetic values into its domain.

## Guidance structure

Prefer one optional **Help / Learn this view** entry with 3–5-step mini-guides:

1. **Find something happening now:** choose source scope, Start, Pause/Inspect.
2. **Narrow records:** local Find versus field Include/Exclude and explicit Apply.
3. **Follow context:** reported trace IDs/parent links versus nearby time context.
4. **Analyze a fixed window:** stage from an observation, inspect bounds, Run.
5. **Keep and reuse:** save a question versus retain observations versus reopen a
   stored revision. Explain acknowledgment/expiry only when that workflow needs it.

Offer the relevant guide from an empty state or local Help, and make it replayable.
Do not launch it automatically over incoming data, obscure operator controls, force
completion, or require an acknowledgment checkbox. First use should not need a
Goo manifest or the sample application. Use safe supplied examples in a separate
lesson, never silent reads or writes against the operator's current scope.

## Minimal reusable pieces and order

1. **Implemented during this review:** native custody disclosure in Activity,
   using `UI.disclosure` and `ui-context-help`; custody state stays in its summary.
   Final keyboard, focus and source-free browser checks are pending at the time
   of this design note; see the validation record for the executed result.
2. **Next:** one shared contextual-help disclosure pattern, with a stable visible
   trigger, native keyboard behavior, concise subject-specific content and a link
   to a relevant mini-guide. Reuse existing `UI.disclosure`; no tour SDK is needed.
3. **Then:** a replayable guide surface whose steps are declarative text/targets,
   not executable source actions. Preserve focus/scroll, support Escape/Close and
   never auto-advance when new telemetry arrives.
4. **Content migration:** consolidate duplicate applied predicates and generic
   Services/comparison explanations; retain actual dynamic constraints and errors.

Acceptance should measure space to the first useful field/record and show that
operators still identify current scope, source/custody state and next-action
consequences. Keyboard users must reach, open and close Help without losing their
place; touch users cannot be required to hover. Existing source-free interaction,
state-transition, focus and uncertainty tests must continue passing.
