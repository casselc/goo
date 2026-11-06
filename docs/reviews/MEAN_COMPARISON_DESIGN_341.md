# Previous-window mean duration — design review 341

Status: source and actual-frame acceptance complete for 341. Native execution
results are recorded separately in the journey and validation record.

The useful task is “did latency change, and over how many contributing spans?”
Reuse the existing ordinary population workflow and passive measure comparison.
The first cut adds summary mean duration to fresh fixed-query comparisons. It
adds no dashboard, query engine, chart catalogue or permanent toolbar.

## Source findings

`ui/population-workspace.mjs` already stages an adjacent equal-duration window,
requires an explicit Run, reads the baseline independently and retains the parent.
Its previous-window admission currently requires Count.
`ui/captured-count-comparison.mjs` also pins declaration value to Count.
`ui/comparison-visual.mjs` already composes independently labelled sides, UTC
windows/read times, scope disclosure and `goo-measure-comparison`; its projection
currently chooses only `count`. `ui/measure-summary.mjs` already renders exact
supplied comparison cells, changes, shared scales and keyboard disclosures.
The backend supplies duration sums/counts and exact rational mean changes; this
work exposes that authority rather than estimating from inspected records.

## Interaction and compact presentation

Start with an ordinary span population using Mean duration. The existing Compare
previous window action stages the complete selected question and adjacent
baseline. Staging makes no source read. Explicit Run independently acquires both.
Keep Previous window / Selected window labels stable across numeric values,
support counts, windows, clocks and supporting-record continuation.

The main summary must say **Mean duration**, show readable **ms** values and the
signed supplied change, and expose each side's **contributing spans**. A whole
record count is useful context but is not the mean denominator. Do not label the
record count as contributors or infer contributors from the visible record page.
Use the existing compact comparison/table geometry, with aligned values and a
shared scale. A summary-only mean declaration needs no Summary/Groups selector:
count group tables must not appear to be a duration breakdown.

Keep the full original query definitions, typed environment policies, UTC bounds,
source basis and independent read clocks inspectable in the existing disclosure.
Do not replace the two read clocks with a common “updated” time. Avoid repeating
long explanations above the result. Existing supporting-record actions and Back
retain their scope and ownership; neither disclosure nor Back issues reads.

## Meaning and precision

- A measured duration sum of zero with a positive contributing count is **0 ms**.
- Zero contributors gives **No data**, not zero latency. A missing or untrusted
  value is **Unavailable/Unknown**, with its existing diagnostic reason.
- Each mean comes from its whole-population exact sum/count. Unequal counts do
  not justify averaging subgroup means; inspected pages are not a sample basis.
- The exact rational remains authoritative even beyond JavaScript integer range.
  Round only the visible human value. Exact disclosure may retain the labelled source rational in ns
  with contributing counts; raw nanoseconds must not dominate the primary view.
- Keep unit incompatibility and diagnostic comparisons explicit; no persuasive
  comparative geometry or invented change when the supplied change is unavailable.
- Upstream telemetry coverage remains unknown. A complete aggregate acquisition
  does not imply complete instrumentation or a complete distributed trace.

Captured-baseline mean, category duration comparisons, scalar percentile
averaging and percentile/status expansion are outside this cut. Count remains
usable without forcing duration fields onto log or count-only populations.

## Acceptance storyboard

The native journey should use genuine OTel spans with independently computed
unequal contributing counts and known means; stage the previous window, verify
zero reads, Run, and compare both human ms values plus their supplied change.
Inspect the two original definitions, typed policies and independent source read
clocks. Open supporting records and return without weakening either question or
adding replacement reads. Exercise measured zero, no-data and unavailable states
in focused semantic/mounted checks, including values beyond 2^53 and incompatible
units. Preserve the existing Count regression.

Review actual desktop and narrow Playwright frames after implementation, including
comparison, scope detail and retained return. Verify keyboard disclosure, no
horizontal page overflow, readable contributor labels and visible focus targets.
Actor JPEGs that repeat intermediate states are not terminal visual acceptance;
use independent current Playwright frames and the observed action/source ledger.

## Implementation source review

The frozen implementation now carries `duration.mean` through fresh query
admission and staging, projects only that mean in the existing passive summary,
and takes each contributor count from the supplied duration cell support. Its
Groups selector is hidden for mean; the entry explains that captured baselines
support counts only. Human primary values and changes use ms. The existing two
windows/read clocks and scope disclosure remain intact. Source review finds no
new interaction surface or numerical authority derived from loaded records.

The exact value disclosure deliberately retains the labelled source rational in
ns as numerical authority. This is supplemental precision detail, not the primary
comparison language; readable ms values satisfy the human-unit requirement.

## Next consolidation judgment — 342

A measurement-first review of reusable high-cardinality facets is an appropriate
next consolidation cut. Measure a 1,000-value single side and disjoint 2,000-value
pair, including maximum admitted labels, at desktop and narrow widths. Observe
time from typing to the rendered filtered list, focus/caret continuity, selection
and keyboard reachability, and retained return. Optimize only demonstrated costs
or usability failures in the existing component. Preserve both operand counts,
missing-versus-empty categories and source-free local filtering. A new virtualizer,
query engine or interaction mode is not justified by row count alone.

## Actual frame acceptance

Reviewed the genuine applied `341-mean-pair.png` and reopened narrow
`341-mean-narrow.png` after native Save/YAML acceptance. Desktop shows aligned
2 ms / 3 ms / 1 ms change, matched shared-scale bars and five versus three
contributing spans. Narrow retains directly comparable side-by-side mean values
and bars, stacking the change below; no clipping or horizontal overflow is visible
in the supplied frame. Both preserve labelled previous/selected windows, separate
UTC read clocks, unknown upstream coverage and scope disclosure. No irrelevant
Groups selector appears. Visual acceptance passes for these actual frames.

The passive source provenance repeats some window/read information on narrow
screens. That is a later consolidation opportunity, not a blocker or a claim that
the full page has completed its density work. These cropped interaction frames do
not establish fresh-entry above-fold placement, keyboard behavior or terminal
actor appearance; native and observed checks supply their separate evidence.

The final audited `341-zero-and-no-data.png` also passes visual acceptance:
Previous window shows zero contributors and **No data**, while Selected window
shows two contributors and **0 ms**. The change reads **Unavailable**, not zero,
and comparative bars are withheld. Per-cell 0/0 versus 2/2 support and separate
read clocks remain visible. The overall No data status does not erase the known
selected zero. This frame directly demonstrates the intended distinction.
