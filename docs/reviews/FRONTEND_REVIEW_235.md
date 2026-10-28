# Independent frontend review after 235

Reviewed 2026-10-28 against `5d12ce9`. This is an independent frontend/source and
bounded interaction review, not an endorsement of the refreshed roadmap, a full
accessibility audit, a production performance benchmark or customer research.
No runtime code, telemetry, captures or investigations were written. Only this
review file is a repository change owned by this reviewer.

## Assessment

The framework-neutral/passive-component direction is sound. The current source
already has useful distinctions between display, selection and source owners;
shared records, exact-value plots, request cancellation, retained resource returns
and versioned questions are substantive foundations. A framework rewrite would
not resolve the main observed problems.

The refreshed roadmap underweights **interaction correctness before density**.
Make applied-scope safety and the inspect handoff the first acceptance gate. Mobile
button compression is worthwhile, but compacting an inspector entry that fails to
bring the inspector into view would optimize the wrong part of the task. Capture
search/recency remains the next broadly useful product improvement. Comparison
support and recurring trace reuse remain valid, later bounded proposals.

## Method and evidence limits

Read AGENTS, START_HERE, current validation/issues, the 235 roadmap/backlog,
component/source-owner code and recent journey contracts. Inspected live Activity
at 390×844, paused it, opened a real observed record, and examined its inspector,
fields, accessible-control metadata and focus. Requested one existing capture
catalog page read-only; that page contained no comparison-kind receipt, so no
fresh saved-comparison rendering claim is made. Saved comparison review below is
source-based. No uncaught page errors, duplicate document IDs or blank visible
form controls were found in the initial bounded probe; these are not whole-app
certifications.

Temporary probe data and frames used `/tmp/review235-frontend` and were removed
after review; the measured observations are recorded here. The inspector
geometry probe used keyboard activation and ordinary local OTel observations. A
separate two-size probe stopped at a missing exact Pause-stream label; it is not
reported as a successful desktop check. The successful mobile geometry result is
independent of that incomplete probe. Historical screenshots/journeys are supporting
context, not freshly reexecuted acceptance.

## Findings in priority order

### F1 — High: Inspect can reveal nothing in the current viewport

**Verified live and in source.** At 390×844, keyboard Enter on an Activity row's
Inspect button pauses the stream and creates the selected-record inspector, but
the inspector begins at y=901.09 and ends at y=1354.39. Focus remains on Inspect,
outside the inspector. The opening control has neither `aria-controls` nor
`aria-expanded`. The screenshot remains a record list; the user must discover the
new region by scrolling or keyboard traversal.

Evidence: `ui/components.mjs:12–17` implements a retained inspector but show does
not focus, reveal or announce it. `ui/stream-entry.mjs:46–59` appends the inspector
after the tail and calls show without an application handoff. The same shared
inspector is also consumed by catalogs and inline comparison detail, so its default
must not be globally changed to modal behavior without checking those contexts.

**Bounded response:** define an application-owned inspect/open/close handoff.
Activity should reveal the selected details and deliberately focus either its
heading or a suitable first control on keyboard activation, with an explicit
relationship from the opener and reliable close/return. Pointer behavior should
be deliberate too. Keep the passive presenter free of fetch/navigation and avoid
requiring every render to steal focus. A contextual option/callback can express
activation intent without a universal navigation system.

Acceptance: source-free keyboard and pointer opening at 320/390/700px and desktop;
visible selected details; meaningful active focus/announcement; close restores a
still-valid opener or a useful fallback after eviction. Append, pause, resume,
filter changes and withdrawn owners cannot expose mismatched record details.

### F2 — High: signal changes discard applied filters and immediately read

**Source-verified; independent live reproduction supplied by the product reviewer.**
`ui/stream-entry.mjs:92` resets filters and calls start on a normal signal change.
The other reviewer reproduced an applied Logs/service predicate disappearing when
Spans immediately issued a broader tail request. This is more urgent than cosmetic
row density because a user can lose the population they were investigating.

**Bounded response:** preserve cross-signal-compatible predicates, make incompatible
ones explicit, and use a coherent staged/apply or clearly disclosed switch policy.
Do not silently broaden a selector. Check source effects and draft/applied state
rather than only chips/text. Coordinate with the product review for the detailed
reproduction and preferred policy; this reviewer did not repeat that live sequence.

Acceptance: compatible service/environment/namespace filters retain exact identity;
incompatible signal fields are surfaced before broadening; actual request payloads
and read counts prove the chosen gesture. Saved views, native continuations and
hidden-document/resume behavior retain their existing contracts.

### F3 — Medium/high: density must include keyboard workload and field actions

**Verified source structure and bounded live inspection.** Every record timestamp
is focusable (`ui/record-presentation.mjs:15`) and every tail row adds independent
trace/select/inspect buttons (`ui/tail.mjs:75`). In a short live buffer one DOM probe
counted 340 visible control/tabindex candidates. That probe included disabled
controls and is **not an exact tabbable count**. The structural multiplication is
nevertheless clear. Inspecting a 29-field record adds per-field Copy, Include,
Exclude and often Plot buttons; other contexts also add trace-search/filter/run/
breakdown actions (`ui/attributes.mjs:104–109`). The structured viewer is much
better than raw JSON, but its command repetition becomes another long workbench.

**Bounded response:** P1 should treat row and field command disclosure as a single
interaction design problem. Reuse the existing inspector/action primitives: one
obvious inspect action, compact explicit selection, and contextual secondary
commands with clear staged-versus-read effects. Keep literal field qualification
and complete value inspection. Do not simply remove keyboard access or hide every
operation behind an unlabeled icon.

Acceptance: same-data row/toolbar measurements; a realistic keyboard path to one
record and one attribute action; touch target/spacing checks; no menu-open read;
no conflation of Inspect, selection and trace loading. Exact UTC values remain
available without forcing every timestamp into the primary Tab traversal.

### F4 — Medium: plot keyboard navigation does not scale with dense evidence

**Source-verified accessibility/design risk, not a measured whole-app defect.**
`ui/value-timeline.mjs:38` creates a normal button for every mark, with no roving
Tab policy. The 209/234 container history can expose 256 observations per page.
The heatmap creates a button per supplied cell (`ui/bucket-heatmap.mjs:25`). Its
range keyboard logic helps extension, but ordinary Tab still visits individual
cells. Timeline marks have 14px minimum width/16px height
(`ui/value-timeline.css:6`); dense/overlapping values can be difficult to target.
Do not declare a WCAG failure from size alone: dense-visualization exceptions,
spacing and equivalent alternatives need a real audit.

**Bounded response:** add one shared plot-navigation convention: enter a plot,
move a keyed active observation/cell with arrows, explicitly select/anchor/range,
and leave the plot with Tab. Provide a usable exact-value alternative for overlapping
points. Preserve each independently reported point and unknown state; do not merge
observations to make interaction easier.

Acceptance: exact supplied-key selection, multi-series overlap, unknown intervals,
reverse ranges, touch scroll versus brush, Home/End/arrow/Escape behavior, and
property updates without accidental source reads. This is a follow-up to P1's
record accessibility, not a prerequisite for every display component.

### F5 — Medium: live property updates have inconsistent focus/disclosure contracts

**Source-verified behavior; current production disruption not reproduced.**
`goo-latency` replaces its children whenever summary/attributes render
(`ui/components.mjs:46–52`); `goo-value-timeline.render` also replaces its entire
subtree (`ui/value-timeline.mjs:34–38`). Focused children and disclosures cannot
survive such replacement automatically. By contrast the record-column picker
explicitly preserves disclosure/search/focus by identity
(`ui/record-columns.mjs:17–20`), and record summaries preserve a disclosure flag.
This inconsistency matters for the user's eventual SSE/live component direction.
It is not evidence that every current screen is already broken on every update.

**Bounded response:** document and test component update policy before more live
adoption: identical input is a no-op where practical; selected keys and meaningful
focus survive compatible supplied-data updates; invalid/new-scope input may reset
with an explicit application decision. Keep passive components free of source
subscriptions and application history. Apply this to a named live widget first,
not a sweeping base-class framework.

### F6 — Lower/measure first: row projection repeats whole-buffer work

**Source-verified work pattern, unmeasured cost.** Tail reconciliation retains
existing row elements, which is good, but calls record-column update over the full
buffer (`ui/tail.mjs:75`). That update discovers fields, projects every row and
assigns a new picker model (`ui/record-columns.mjs:25–27`); picker render replaces
its controls. This can do repeated work for an unchanged schema/layout during
polling. No frame-time or memory regression is established by this review.

**Bounded response:** measure the same fixed buffer/layout and append schedule,
including an open field picker. If material, cache the supplied schema/layout
signature and update only affected rows. Do not add a global data cache, virtualize
every table or adopt a framework based on module/DOM counts alone.

## Comprehensive surface assessment

| Surface / concern | Assessment and action |
| --- | --- |
| Entry and mobile chrome | Stream-first entry is appropriate. Fix scope safety and inspect handoff before compressing controls. Global `details` borders/margins (`product.css:262–274`) can leak into toolbar disclosures; use scoped component chrome, not a broad all-details redesign. |
| Fixed/live/saved records | Shared supplied-row rendering and exact clocks are worth preserving. Apply one inspect/return contract across embeddings without making an inspector a modal everywhere. |
| Attributes/facets | Qualified keys, missing/empty distinctions and explicit action labels are strong. Consolidate repeated commands while preserving staged versus immediate effects. |
| Traces and charts | Existing exact ranges, keyboard range selection and ancestry limits are useful. Add scalable plot navigation; do not replace exact semantics with cosmetic merged percentiles or inferred causation. |
| Comparisons | P3's support/denominator work is justified. Also make side-inspection entry/return consistent (`comparison-inspection.mjs:7–18`). No fresh saved-comparison live rendering was reached in this bounded review. |
| Saved / Definitions | Capture search/recency is the right next product feature. Existing typed definition selection, paging and outside-filter snapshot notices (`definition-library.mjs:4–13`) should be reused, not replaced by title-based grouping. |
| Resource/native returns | The explicit application presenters and recent exact-question holds are a useful bounded pattern. Shared visible focus restoration should also inform inspect entry, while not making ancestry persistent/global. |
| ESM/component architecture | Retain framework-neutral web components and application source owners. App modules mounting once for page lifetime is not itself a memory leak. Add dispose symmetry where a real remount/dynamic-host task needs it; do not mandate it everywhere. |
| Async reliability | Existing cancellation/generation and source-family boundaries are substantial. Guard actual read payload/selector transitions and context-sensitive actions, not just late DOM text. |
| Styling and maintenance | `product.css` is large and chronological. Consolidate styles only along a changed component/owner boundary with composition checks. An arbitrary file split would add churn without addressing F1/F2. |
| Security / SSR | The examined display paths use supplied nodes/text and bounded attribute parsing. No new XSS defect was identified, but this is not a security audit. Keep server-rendered and detached-property upgrade tests. |

## Recommended execution order

1. **Interaction safety gate:** scope-changing signal controls plus visible,
   accessible record inspection and close/return, with observed source effects.
2. **P1 refined:** mobile/desktop row and field-action density, keyboard workload
   and scoped toolbar chrome, using fixed-data measurements.
3. **P2:** bounded capture search/recency and exact typed reopening. This is a
   broader everyday task improvement than adding another profile or editor.
4. **P3:** support/denominator context with exact totals and honest uncertainty.
5. **P4:** one recurring trace-question definition, after its versioned protocol
   and source-free reopen contract are concrete.
6. Named enabling follow-ups: shared plot navigation/update stability, measured
   row-projection optimization, and duplicate retention-status removal.

Do not add a framework migration, generic UI store, universal ancestry router,
mandatory bitemporal store, broad new telemetry profiles or a new component for
every screen. Do not reclassify all previous milestones as complete accessibility
or production-readiness acceptance. Preserve the rich semantic and negative-control
work while turning the next review gate into one complete human task.
