# UX assessment: persistent facts and optional guidance

Status: design recommendation after the 459 shared-component polish. This document
proposes guidance behavior; it does not claim a tutorial system has been built.
The immediate Activity disclosure extraction was implemented during this review;
the broader guidance panel and walkthroughs remain proposals. Executed checks are
recorded in [validation](../../VALIDATION.md).
Reviewed the fresh 459 Activity, inspector, fixed-record, Services and Saved frames,
the 458 chronology/inspection frames, recent raw-metric frames, and their owners.

The interface should let an experienced operator read data immediately. Teach the
mental model once, keep the actual state visible every time, and make detailed
explanations available where they matter. A walkthrough cannot compensate for an
ambiguous action label or conceal a current limitation.

## Immediate finding

The Activity inspector still opens with a boxed explanation before the log message
and fields: “Stages a fixed question for a separate source read; held rows are not
saved evidence.” In the fresh 390px frame, the message and that paragraph consume
most of the 24rem inspector body before the first field appears.

Implemented during the review: that paragraph is now inside a collapsed native
disclosure labelled **Local observation · not saved**, retaining the complete
explanation inside and keeping the custody state visible in its summary. The
existing **Original observation window** disclosure retains exact current facts.
Future guidance should group related explanations rather than add several Help
controls. Do not move the log message into a tutorial: that is the operator’s data.
Increasing the scroll area alone would mask the text hierarchy problem; a responsive
inspector-height review can follow independently.

## Concrete inventory

| Surface and source | Keep visible | Move, consolidate or teach | Visual treatment |
| --- | --- | --- | --- |
| Activity inspection, `ui/stream-entry.mjs` | Local/not-saved state; selected record; Analyze action | The fixed-question/separate-read explanation belongs in the local-observation disclosure and a follow-along guide | Small text-bearing custody token or labelled native disclosure; no warning-card styling for ordinary operation |
| Activity tail, `ui/tail.mjs` | Live/paused/unavailable, held count, gaps, evictions; current source scope | Rolling-window mechanics, arrival versus event order, and multiplicity definitions belong in existing Tail scope and limits plus a short guide | Read window and last-read time as compact labelled facts; gap/eviction counts stay adjacent to state |
| Held chronology, `ui/held-chronology-view.mjs` | Separate Log/Span counts and states; newest-first order; Find boundary when lane Find is ignored | Stable tie ordering, lack of causality, and why original windows differ should be on-demand | Existing source disclosures are a good model. Add a labelled order affordance; do not replace state labels with color |
| Scope/filter editing, `ui/filter-editor.mjs`, `ui/cel-expression-editor.mjs` | One authoritative applied scope; unapplied/draft indicator; Apply action and validation errors | Repeated Parent/Applied predicate sentences can collapse into an Editor context disclosure; visual-to-CEL equivalence belongs in an authoring guide | One applied predicate chip set, a draft marker and explicit Apply; exact expression remains available |
| Services catalog, `ui/inventory.mjs` | Actual signal counts with grain-specific labels; Not observed; namespace; window | Repeated “counts retain signal grain…not metric values…not health” explanation belongs in a nearby About these counts disclosure | Keep three signal count columns and explicit missing states; a shared signal legend explains their units once |
| Distributed services, `ui/distributed-services.mjs` | Reported parent/child identity, child-span record count, unresolved counts | The long topology/manifest/non-causality qualification belongs in graph interpretation help | Label links **Reported parent**, use a distinct unresolved state; never suggest request volume or causal certainty through arrow styling |
| Raw gauge/histogram display, `ui/raw-gauge-points-view.mjs`, `ui/raw-histogram-buckets-view.mjs` | Recording uncertainty, unplotted counts, recorded-empty distinction, exact values | Plot geometry, overlapping coordinates, categorical widths and keyboard instructions belong in accessible on-demand chart help | Persistent state text plus marker, an unplotted-count affordance, compact order/unit labels; zero is a visible value, not an empty-state icon |
| Measurement timeline, `ui/measurements.mjs` | Exact stream identity, unit, interval/window, unknown/stale gaps, sample/interval count | “One exact stream”, stock/rate interpretation and retained-versus-total explanation should be a reusable chart interpretation guide | Unit and grain badges; actual gaps remain visually distinct. Do not fill a gap with a zero line |
| Saved comparison, `ui/retained-page-comparison.mjs` | Baseline/Selected labels, each side’s window and source clock, known partial/changed-read states, page denominators | Receipt mechanics, custody versus source time, raw basis and exact JSON fit existing Scope and receipt / Exact question disclosures | Align the two sides in the same labelled fact order; receipt ID and detailed lineage are secondary |
| Preservation controls, `ui/stream-entry.mjs` and existing Keep/Capture owners | Which operation saves configuration versus records; confirmation of the completed write | Teach Save question versus Keep/Capture once and provide contextual help at the preservation menu | A small **Question → View → Stored observations** diagram can explain categories. A success label must still name what was saved |
| Empty and error states, `ui/inventory.mjs`, source-refusal owners | Actual no-data/failed/unknown state, current scope, next applicable action | General observability theory belongs in help; the cause and recovery for this failure do not | Concise state + reason + Retry/change-scope action. Never use a green empty result to imply health |
| Definitions/authoring, `ui/definition-library.mjs`, `ui/telemetry-packages.mjs` | Contract identity, revision and whether an editor change is staged | Schema/basis/compiler internals and lengthy semantic references belong in authoring help/reference, not routine investigation | Advanced panels remain explicit. Validation messages point to the relevant field/expression rather than a generic tutorial |

The footer’s “Captured records are evidence, not a guarantee of coverage” is already
visually hidden. It should not become another repeated screen-reader-only tutorial.
Explain that principle in interpretation help and preserve actual coverage labels on
results; hidden verbosity still imposes a cost on assistive-technology users.

## Coherent help entry points

1. A consistent **Help** entry opens a route-aware, nonmodal guidance panel. It
   offers a short explanation of this surface, relevant workflows, and reference
   links. No inline copy is duplicated simply to populate the panel.
2. Reuse native **Read facts**, **Scope and receipt**, and **Tail scope and limits**
   disclosures for exact current facts. Guidance explains their meaning; it does
   not copy another snapshot of those facts.
3. Contextual **How this works** controls appear only beside concepts that need
   explanation. Use visible text or an accessible name; do not depend on hover,
   title-only instructions, unexplained glyphs or color.
4. Walkthroughs are opt-in and replayable. A quiet link in an initial empty state
   can advertise them. Do not auto-launch an overlay over arriving telemetry.

A native disclosure is enough for the immediate extraction. A shared guidance panel
can follow once the content inventory and routes are agreed; a third-party tour SDK
or a new tutorial DSL is unnecessary for the first increment.

## Suggested first workflow: investigate ordinary telemetry

| Step | Operator task | Guidance teaches | Completion evidence |
| --- | --- | --- | --- |
| 1. Choose a source scope | Choose logs/spans and environment; Start explicitly | The UI observes existing OTel data without requiring a manifest | The current scope and a completed read are visible |
| 2. Hold and inspect | Pause or Hold chronology; inspect one occurrence | A held view stops changing; it is not automatically saved evidence | A real inspector is open for a retained occurrence |
| 3. Narrow the question | Find locally, then stage a field condition and Apply if desired | Local presentation changes differ from a new source read | Visible Find/draft/applied state changes, with Run/Apply still explicit |
| 4. Follow relevant context | Open a reported trace or an original observation window | Reported identity differs from temporal proximity | A staged question names the original scope/window; only the user runs it |
| 5. Reuse the work | Save the question or preserve observations deliberately | Configuration and stored observations have different clocks and purposes | The actual operation’s success receipt is shown, or the user skips the step |

Use the operator’s current data. Offer an explicitly selected sample workspace only
when they request it. Steps may be skipped, revisited or closed without losing
Find, source scope, paused rows, selection or navigation history. Guidance must never
start a stream, execute a query, modify filters, create a capture or mark a step
complete merely because it highlighted a button.

Separate on-demand mini-guides should cover comparisons, interpreting metric
uncertainty, and authoring a reusable definition; they should not burden the initial
logs/spans investigation.

## Accessibility and acceptance

- Open guidance only on explicit request; move focus predictably and restore the
  trigger on Close. Keep it nonmodal while the operator works in the app.
- Use a heading and ordered steps, with **Next**, **Previous**, **Skip** and **Close**
  controls. Do not trap focus or require precision pointer gestures.
- On narrow screens, put instructions in document flow or a resizable panel;
  anchored speech bubbles must not cover the target, active editor or data rows.
- Announce a changed step once. Retain actual source/result announcements without
  making a guide repeat the entire page or current query.
- Handle unavailable targets and empty/error states honestly. A source failure
  must not become a compulsory tutorial step or a false success.
- Verify keyboard-only, zoom/320px, touch, reduced-motion and screen-reader order.
  Test that opening/closing/replaying guidance adds zero source or write actions.

## Priority

**Implemented in 459:** the ordinary Activity inspector explanation is collapsed
under **Local observation · not saved**, preserving the real message and facts.
Shared attribute controls and state markers also improve scanning. See validation
for final browser and keyboard checks.

**Next:** audit duplicate applied-scope copy, add a small shared context-help entry,
and deliver the five-step optional workflow above. Start with three or four short
explanations rather than migrating every caveat at once.

**Later:** interpretation mini-guides and authoring reference; consolidate repeated
receipt and completeness explanations across Services, charts and Saved. Keep
source-specific limitations and current errors at the point of use.
