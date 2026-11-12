# Result presentation review

Status: accepted for the bounded interface refinement after 415. No remaining
blocking design, ownership or inspected accessibility findings. Milestone 416
remains the separately planned exception-grouping work.

The reviewer inspected the source changes and actual generated desktop, 390px
and 320px screenshots for service-version comparison, service overview and scope
recall. This is a focused review of the affected surfaces, not a whole-product
accessibility certification.

## Findings addressed

1. **Comparison hierarchy and comparability.** The result now has one title and
   prominent paired values on a shared scale. Baseline blue and comparison teal
   supplement explicit labels and numbers. The 132/168 captured counts and 36
   change are readily comparable at all three widths. Exact scopes, clocks,
   quality and acquisition facts remain in the existing native disclosure rather
   than repeating transport metadata above the chart. The visible notice retains
   independent-population and unknown-coverage meaning. Active failures retain
   their refusal/status paths.
2. **Shared composition.** `goo-measure-comparison` uses explicit `data-first`
   and `external-context` presentation attributes. Standalone data-first use keeps
   its own read-facts disclosure; the validated composing adapter owns those
   facts when supplying external context. Display components gain no source or
   navigation authority. The application removes its duplicate title and moves
   the explicit baseline action into the same facts disclosure.
3. **Service overview and timeline.** Compact statistics, latency and the action
   footer have a consistent hierarchy. Dense question-mark glyphs are replaced
   by quiet unavailable markers. Each supplied marker retains its identity,
   state, accessible label, exact tooltip and selection behavior. The all-unknown
   lane has a shorter brush canvas, with explicit unavailability and exact UTC
   endpoint information. Unknown observations are not shown as numeric zero.
4. **Recall and held-visit readability.** Rows distinguish a primary filter from
   secondary kind, signal, environment policy and time. Held views retain visit
   identity and state. Actions align consistently on desktop and wrap at narrow
   widths. Exact timestamps remain available through titles and accessible
   descriptions. The ordinary visit bar now uses a compact UTC range instead of
   repeating the full environment/predicate paragraph; exact title information
   remains. Configuration staging, held-result return and persistence remain
   separate operations.

## Ownership and interaction checks

Moving **Read fixed baseline** into a disclosure requires checking the button's
current native visibility and enabled state at invocation. The inspected handler
uses `nativeControlReady(prepare)` alongside the existing comparison owner gate.
The actual journey invokes a retained handler while facts are closed and records
zero additional requests or SQL reads. Keyboard opening of comparison/read facts
also remains source-free. Group-action failures explicitly open the containing
facts disclosure so the relevant status can be seen.

The narrow comparison label slots keep both numeric values on the same baseline,
including the wrapped baseline label at 320px. The inspected actual report gives
comparison-table offsets of **99.7px / 97.7px / 132.4px** at 1440/390/320px; paired
value positions are equal at every width. Facts are initially closed and routine
coverage text does not appear as an amber alert in the result. The primary
comparison requires no horizontal scrolling on the inspected narrow frames.

## Evidence inspected

- Final frontend log: **1,561 passed**, no failures or skips.
- Three real-data browser journeys pass in the serial run: existing service
  previous-window comparison, Recent/Pinned scope recall and service-version
  comparison. Their source/count/scope and held-return checks remain active.
- Version journey: **18 recorded facts**, zero page errors, completed actor with
  **488 snapshots**, no forbidden publication and preserved held parent evidence.
  Discovery remains two SQL statements; each explicit comparison remains six;
  local actions and closed-disclosure callbacks add none.
- Scope journey: **18 facts**, zero page errors. The saved actor report records
  **582 snapshots**, completion and unchanged POST/SQL counters of **3/7**. The
  final compact-visit-bar build was then checked by a separate successful recall
  rerun, with regenerated 1440/390/320px frames personally inspected.

Actual reports are ignored under `artifacts/service-versions415/` and
`artifacts/recent-pinned-scopes/` (including `actor-report.json`). Tracked frames
are under `docs/journeys/screenshots/`. Source/model gates, the embedded-runtime
check and final application refresh are recorded in [validation](../../VALIDATION.md).

## Limits

The review covers Chromium source behavior and generated frames. It does not
claim physical screen-reader testing, cross-engine conformance or complete WCAG
coverage. Long authored filters and large histories may still require the
bounded native list's internal scrolling. The pass changes presentation and
explicit control visibility, not query populations, statistical interpretation,
read clocks or retained-evidence meaning.
