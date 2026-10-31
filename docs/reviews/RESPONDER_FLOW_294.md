# Responder entry and continuation review — 294

2026-10-31. Source review at `a357964` of Activity, fixed-analysis and composed
population owners, informed by the executed293 walkthrough and its explicit gaps.
This review did not run a browser, measure a layout, or establish a completed
end-to-end journey. Actual 294 browser results belong in VALIDATION and its storyboard.

The strongest next improvement is to **extend the existing Customize spans/logs
handoff to applied noncomposition populations**, including the fixed-window linked
population. This bridges the natural Activity analysis route into a single-query
view workspace without computing a new clock or adding another editor. A responder
should not need to know the internal recent/fixed owner distinction to continue.

Keep **Customize spans** and **Customize logs**: the labels identify which signal
is selected from a two-signal parent. Calling the action Compare would be too narrow;
the workspace also supports records, charts and preservation. The fixed analysis →
Explore linked population step still makes explicit independent reads. This change
must not imply those reads disappear, or that a child's staged configuration is
new evidence. The live walk must establish the actual complete path.

## Source-backed findings

| Priority | Finding | Consequence and bounded response |
| --- | --- | --- |
| 1 | Activity's **Analyze this scope/window** stages the completed observation window and exact applied selector (`stream-entry.mjs`, `analyze.onclick`). Fixed analysis does not offer the existing population **Customize spans/logs** action: `mountFocusControl` only renders it when `recent` is true. | The natural entry protects the right scope but does not expose the shortest continuation into reusable count/group comparison. Extend the existing checked Customize handoff to the fixed linked population, selecting the same signal question and view owner. Keep Run explicit. |
| 2 | **Show recent activity** calls `recentQuestion(...Date.now(), minutes)`; it adds default service grouping and opens the recent population, whose Customize action can create aggregate-plus-records composition. This button is also hidden in collapsed Activity analysis until Edit analysis. | This is a fresh-window workflow, not an equivalent continuation of the selected observation window. A successful journey through it must record that clock change. Do not use it as proof that original Activity scope is retained. |
| 3 | **Explore linked population** sets `recent=false`, calls the population planner and automatically executes its planned queries. Composing a new view requires exactly one query, an eligible ordinary grouping, and a reusable display. | “Explore linked population” is a real independent-read operation, not a passive view editor. It can introduce a companion signal and does not render the recent-only Customize action. Do not relabel it as customization or silently discard its companion to make comparison work. |
| 4 | Captured count comparison is configured under **View settings → Composition**, with title search, a capture picker and an exact-ID fallback. Choosing a capture stages configuration; preparation is separate. | Preserve that explicit stored-evidence boundary. Once the main entry bridge is solved, make discovery task-oriented in context; a second comparison wizard or new saved-object family is unnecessary. |
| 5 | The trace inspector sets its summary to **Reading trace context…**, then renders a waterfall and linked-log read. The293 screenshot caught only loading. |294 acceptance must wait for settled captured-span count and actual waterfall rows, then close and return without confusing that fresh context with the held population. Opening the dialog alone is insufficient. |

## Pit-of-success contract for the first change

- Act only from a current applied fixed question with no pending query/filter or
  record-selection state. A retained or detached old control must refuse.
- Preserve environment policy, original Boolean filters, fixed half-open bounds,
  chosen signal and typed grouping. Reset delivery offset only under the existing
  focused-question rule, and state that behavior.
- Reuse the existing population/view staging owner. Opening customization reads
  no telemetry, creates no capture and does not save a definition.
- Reuse the existing per-signal Customize label, placement, private lease and
  depth guard. Make it available only after that parent signal has applied evidence,
  including fixed linked populations; no additional top-level action is needed.
- Ordinary eligible grouping may immediately use the existing aggregate/records
  composition. An ungrouped or specialized profile should expose the existing
  explicit grouping path or explain its limit; do not invent a grouping silently.
- Keep the parent's held evidence and its return destination. Explicit child Run
  is independent; Return must not rearm retired source handlers wholesale.
- Preserve a view only after the responder can inspect it. Reopen must stage the
  exact saved definition with no implicit source read.

The 294 live walk may reveal a shorter existing path or a different blocker. Treat
this as a ranked source hypothesis, not evidence of a failed UI journey. Do not
expand this milestone into new chart types, a general navigation framework, or a
combined evidence/configuration save operation.

## Live-journey feedback: optional query measures

The 294 journey agent subsequently reported executing Activity service narrowing →
observed-field grouping → fixed Run → linked population → Customize → child Run,
including Back and fresh/retired Customize controls. It then captured a baseline
through the UI and read a changed current population. This is the executing agent's
report, not a second browser run by this reviewer; final results and frames belong
in its storyboard.

The next concrete blocker was measure configuration: the linked planner adds
`spanStatus:true`, and the authored captured-count profile intentionally excludes
it. The exact capture appears in the picker, but **Set baseline capture** is
disabled and its status explains the unsupported profile. There is no population
measure editor to make the ordinary-count choice explicitly. The fixed analysis
already has **Span duration summary** and **Span status and errors** controls,
but changing the earlier fixed question is insufficient if the planner adds status
again during the linked transition.

The coherent bounded fix is a reusable **Measures · spans** section under the
existing configured/applied population scope context, using those same two labels
with native checkbox changes acting as explicit source-free staging. This is a query-measure choice, not a
comparison conversion or display-only setting. Stage it in the already focused
single-query workspace before capturing the baseline, without consuming another
child depth. Preserve filters, environment policy, grouping and fixed bounds;
use the established zero-offset query staging rule. Run then acquires evidence for
that changed definition independently.

Composition validation must visibly refuse incompatible changes: disabling duration
summary cannot silently replace a duration display; enabling status cannot silently
remove an existing count comparison. Keep numeric, histogram and reported-edge
profiles unchanged, and constrain the editor to its supported profile rather than
normalizing them. Never strip status in Customize or weaken baseline compatibility.
The planner's status default remains useful for initial investigation.

The executing agent also reported nested comparison setup as discovery friction.
Keep that as a later contextual-entry candidate after this query primitive and the
full journey are validated; no extra comparison wizard is justified in this cut.


## Implemented source review

The implemented `mountSpanMeasures` uses immediate native checkbox staging instead
of an extra Stage button. That is coherent: one completed toggle validates a whole
next question, clears held results and requires Run. There is no intermediate
unapplied measure draft to manage. The concise note **Changes stage the query;
Run reads telemetry** explains the effect before interaction.

The controls live under the existing scope disclosure in **Measures · spans**, use
the same fixed-analysis duration/status labels, and are restricted to supported
single-query span compositions. After staging, the scope and measure disclosures
remain open and focus returns to the corresponding fresh checkbox. Refusal restores
the original checked state and presents **Measures unchanged** with the validation
reason. This avoids suggesting that a rejected query mutation succeeded.

No usability source blocker found. This approval assumes the mount's enabled-state
refresh occurs after its owning context connects, as the implementation review
identified; a disconnected-control bug is not a native-disclosure limitation.
Final actual 294 frames and executed interaction checks remain necessary. This is
not a claim of new keyboard, screen-reader or mobile validation.

## Proposed next outcomes after the unseeded flow gate

1. **Make baseline comparison discoverable at the applied aggregate.** Offer one
   contextual entry into the existing capture picker and comparison declaration,
   beside the data rather than requiring knowledge of View settings → Composition.
   Reuse its exact compatibility, preparation and saved-reference semantics. Explain
   unsupported profiles and point to ordinary query measures; never change measures
   or capture evidence as a side effect of opening the entry. Success is a responder
   finding and completing this path without source JSON or knowledge of editor roles.
2. **Turn shipped view templates into a small authorable package contract.** Start
   from the existing HTTP descriptor, supported bar/table/timeline/record composition
   and checked scope owner. Let a team name a reusable template, bind supported
   fields and choose its existing kinds, preview compatibility, then apply/save/reopen
   it in another matching service scope. Add one useful OTel database or messaging
   default through that same path to prove it is not an HTTP-only special case.
   Avoid arbitrary code, a new catalog or a general expression language in this cut.
3. **Add one further useful display shape through the same contract.** Choose a
   record-detail/list or trace-waterfall journey with real demand, explicit identity,
   support and completeness requirements, and source/property parity. This advances
   the user's whole-interface grammar goal. It should not require converting every
   specialized presenter or adopting Plotly before any authoring can ship.

These outcomes advance discovery, reuse and expressiveness in that order. New
source guards should follow observed ownership defects in those journeys; they are
not a parallel roadmap of isolated defensive patches. Final 294 execution may still
change this order if the completed flow exposes a more serious functional blocker.

## Actual 294 frame review

Inspected the executing journey's generated Activity, explicit-measures, compared-
groups, completed-trace and reopened-view frames. The compared groups visibly align
`/checkout` baseline 1/current 2 and `/health` 1/1 on a shared 0–2 count scale; the
comparison is quickly readable without interpreting prose. The completed trace
frame now shows a settled **2 captured spans** summary and both `api · checkout
followup` and `worker · persist checkout` waterfall rows. Its linked-log result is
explicitly zero with no health inference. This closes the 293 screenshot's loading-
only visual gap; the journey owns the independent DOM/read assertions.

The final measures frame shows the applied population's optional duration/status
choices, with compact inline checkbox labels and the note **Changes stage the
query; Run reads telemetry**. Staging, held-result clearing and incompatible-change
refusal were independently asserted by the executing journey; this frame shows the
controls before that conversion, not a later staged state.

The first frame exposed two small style issues corrected in this cut: checkboxes
inherited grid-label layout and sat above their text, while compared route values
inherited uppercase table-header styling. The final regenerated frames show the
existing `span-summary-control` row layout and original lowercase `/checkout` and
`/health` values. The journey also asserts computed `text-transform:none`; exact
case-sensitive values now remain exact visually as well as in query intent.

The reopened-view frame shows the fixed scope, chosen route and two matching spans
after Run. It does not by itself prove source-free reopen; that is a journey
assertion. Its expanded live controls plus older Pin as baseline/Compare previous
window controls also support the first proposed next outcome: consolidate comparison
entry around the applied data, rather than adding another disconnected wizard.
The visual hierarchy of unrelated live controls can be handled in that same future
interaction pass, not as a separate defensive micro-milestone.


## Final verdict

Re-inspected the regenerated measures and compared-groups frames after the final
42.5-second passing journey. Inline checkbox labels are compact and consistent
with fixed analysis. Original route casing and shared-scale comparison are clear.
The final journey additionally asserts two waterfall rows, actual bars and exact
span IDs; the settled trace frame was already visually reviewed above. No design
blocker remains for 294. This review does not extend the executed coverage to
screen readers, mobile layout or unsupported query profiles.
