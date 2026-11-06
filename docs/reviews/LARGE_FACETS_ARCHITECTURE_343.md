# Large counted facets — 343

Status: implemented source review accepted. The baseline is 342 (`b7adf8d`);
corrected styled measurements and actual runtime results are recorded in
VALIDATION.md and the executable storyboard.

The ordinary facet must remain usable when a legitimate grouping approaches its
existing admission limit. This is consolidation of the supplied-value surface,
not a new query, cardinality service, virtualizer, pagination API or completeness
claim. Measure first; make the smallest supported change afterward.

## Existing authority and limits

`FilterValues.query` derives an independent grouping question from the complete
applied selector and fixed window (at most 120 minutes); inspection-page limit
never defines the grouping population. `FilterValues` checks independent count
and grouping receipts and their accounting before publication. `countedState`
accepts at most 1,000 unique positive-count values per ready operand, validates
canonical uint64 counts with BigInt, and requires their exact sum to equal that
operand's supplied total. `projection` unions literal values and can therefore
produce 2,000 rows for disjoint ready operands. It sorts by exact maximum count,
then typed literal key. `null` and `""` remain different keys.

An absent value is zero only for a ready returned partition. An unread, loading or
failed operand remains unknown; comparison geometry is withheld until both are
ready. Retain the independent receipts and upstream-coverage qualification. Local
search must not relabel a filtered list as a smaller measured population.

The current component model has no per-value or cumulative text-byte limit.
The 8 MiB raw `data` attribute guard applies to that SSR parsing path, not to SPA
properties. Do not describe 1,000 rows as a memory/byte bound or invent a maximum
admitted label length. Selection admission remains separate from display: a value
that cannot form an allowed predicate must retain its identity and explain its
disabled action. A new source/model refusal requires demonstrated need and an
explicit contract, not silent truncation.

## Source findings to measure

- `CategoricalFacet.render` validates/clones supplied state, projects paired
  operands, replaces its DOM and constructs all rows. Each paired row includes
  two count-cell trees and include/exclude actions even when only a small portion
  fits in the existing 18rem scroll region.
- Full labels are repeated in searchable data, accessible names, titles and
  closures; only the visible value text is shortened at 160 code units. Long
  attributes can therefore cost more than the visible row suggests.
- `find()` assigns `goo-facet-list.query`, whose observed-attribute callback
  synchronously calls `refresh()`, then explicitly calls `refresh()` again to
  obtain match counts. This duplicates the full DOM scan/lowercase pass per input.
  Removing that redundant work is a candidate only after measuring its contribution.
- The shared list promises to retain children and action/focus identity; its
  `refresh` toggles `hidden` rather than replacing rows. Preserve that contract
  for ordinary shared-list users. Avoid a global virtualization rewrite to solve
  one component's measured cost.

## Baseline protocol

Use the genuine ordinary source path for a 1,000-value single facet and two
1,000-value operands with a disjoint 2,000-value union. Include missing, present
empty, baseline-only and distinct long values sharing the same first 160 code
units; insert a unique distinguishing suffix beyond that preview. Keep exact
expected membership/counts independently, including a synthetic passive
component case above 2^53 where a live record fixture cannot supply that count.

Record fixture text/UTF-8 bytes, source response size, admitted item count and
rendered DOM-node count. Measure source acquisition separately from model
adoption/mount, local search and next paint. Use real desktop and narrow viewport
sizes; distinguish a fresh entry from naturally scrolled interaction frames.
Run several warm independent samples and report median/p95 or the individual
samples if too few for a meaningful percentile. Baseline and final samples must
use the same fixture, viewport, browser and machine constraints; avoid parallel
heavywork. A 100ms local input-to-next-paint target is a usability diagnostic, not
a universal machine-speed assertion or a fragile unit-test timeout.

Search probes include no match, one suffix match, broad match and clear. Measure
focus and viewport overflow as well as elapsed time. Long-label fixtures are
explicitly sized examples, not proof over arbitrarily large admitted strings.

## Closed acceptance

1. Every admitted literal value remains reachable and searchable, including the
   distinguishing suffix. Missing/empty and ready absence/unknown retain their
   meaning. Counts, ordering, full-population scales, clocks and receipts do not
   change with local search. Never search only visible/preloaded DOM rows while
   claiming all loaded values were searched.
2. Find, local disclosure, scrolling and any local result navigation perform zero
   source, comparison, capture, save or metadata requests. Include/exclude stages
   exactly the chosen original literal value; it does not accidentally use the
   preview, normalized display string, row index or another operand's value.
3. Keep explicit Read/Retry ownership and full-query scope unchanged. Partial-side
   display and Retry remain useful; a failed side is not zero. A new receipt may
   reset local Find according to the existing read-identity contract; same-receipt
   rerenders preserve Find and read-detail disclosure.
4. Search keeps its focus and caret. Keyboard users reach actions and can return
   to Find without traversing thousands of irrelevant hidden controls. No narrow
   page-level overflow; values remain identifiable through accessible names and
   full-value detail. Detached SPA properties and SSR attributes still work.
5. If measurements require a bounded rendered result slice, retain all admitted
   data locally, expose truthful shown/matched/loaded counts and a clear way to
   reach the rest. This is conditional, not a selected design before baseline.
   Do not silently hide arbitrary categories or alter predicate identity.
6. Preserve 342 old-owner fencing: retired handlers cannot stage after Back or
   replacement. Current local controls work, and reveal/search never rearms old
   source work. Reuse existing semantic/model assertions for observed local
   effects; add a model only for genuinely new lifecycle state.

The UI/design/native agents should review the measured baseline together before
production edits. Record the chosen minimum change, comparison measurements and
remaining limits here or in the validation record; current storyboards/screenshots
must reflect the real resulting interaction.

## Baseline finding: independent HTTP and cardinality budgets

The genuine baseline exposed a cross-layer constraint before UI adoption:
`internal/recordhttp/policy.go` bounds `/api/compare` request bodies to 1 MiB.
A pair of 1,000 groups with 512-byte labels, plus JSON and comparison metadata,
can exceed that aggregate body budget despite satisfying per-cohort cardinality
and the cohort validator's per-label 4,096-byte bound. These are independent
limits; the component's item-count admission does not promise successful transport
of every possible group combination. Do not raise the HTTP limit blindly.

The genuine four-case baseline uses explicit 384-UTF-8-byte labels with exact
serialized request bytes recorded. A separate passive SPA benchmark uses
2,048-byte labels and is not evidence that those values traversed `/api/compare`.
Facet counted-value reads use `/api/explore/counted-values`, whose request carries
a bounded query rather than the returned groups. The aggregate comparison service
and counted-value display therefore have distinct body/response contracts. A
future demonstrated task may require coordinated byte budgeting/refusal across
those layers; 343 changes neither route admission nor cohort/source semantics.

## Frozen-source acceptance

After the genuine baseline passed the 1,000/2,000-value desktop/narrow profiles,
the selected minimum change adds `goo-facet-list.filter(query)`. It suppresses
only its own synchronous observed-attribute refresh in a `try/finally`, then runs
one fresh pass over current children and returns exact local match counts. It
caches no rows/results and changes no row identity. Direct property, SSR attribute,
detached element and explicit refresh behavior remain available. Categorical
facets, the Distributions fallback and the component gallery use this atomic path.

Visible long-label previews now show 104 Unicode code points, an ellipsis and the
last 55 code points. This makes terminal identifiers distinguishable within the
existing 160-code-point display budget without splitting surrogate pairs. Full
search text, title, accessible names and emitted original values remain unchanged.
No source, count, receipt, partial-state or comparison-geometry code changes.

Architecture accepts that frozen source as the bounded response; whitespace
validation passes. Final source-versus-baseline timing, real interaction/focus
checks, tests and storyboards remain the scheduled acceptance gate, not claims
made by this source-only review. No universal virtualizer, arbitrary data loss or
new maximum-label promise is introduced.

## Follow-up candidate: operation-level publication, not a new virtualizer

Source review after the final gates were scheduled found additional ownership-path
work that 343 does not change. Both connected `.data` and `.interaction` setters
render the whole facet. `renderFacet` sets interaction once during owner setup,
then its final `refreshFacet` sets data. A retained element is also reinserted,
which can invoke `connectedCallback` over its held data. Back subsequently calls
`updateSelection` (which calls `syncCompositionControls` and `refreshFacet`) and
then another `syncCompositionControls`. Several full ready rebuilds can therefore
belong to one retained-return operation.

This is **not** evidence that each source publication renders twice. The ordinary
`setFacetValues` publication calls only `refreshFacet`/the data setter; it does
not reassign interaction. Native instrumentation records individual render sync
times, while a data-setter measurement includes its nested render: summing both
would double-count that work. A genuine ready render around 1.1 seconds versus a
supplied-component render around 0.37–0.66 seconds needs matched operation-level
measurement; the setter pattern alone does not establish that causal difference.
The ready paired tree remains roughly 32,000 DOM elements, so a single render
still has material work even after duplicate publications are removed.

A bounded next depth investigation may introduce one atomic passive update for
supplied data plus interaction and simplify the existing application owner to
publish once per logical transition. Preserve synchronous semantics: validate and
detach the supplied model, install the current interaction, invalidate the prior
render epoch and build once when connected. Existing independent properties and
SSR behavior remain meaningful; no deferred fetch, timer subscription or global
batch manager is needed. Atomicity must not mean reusing a claimed basis string
as proof that arbitrary caller data is unchanged, nor may it revive old handlers.
Measure owner initialization, one-side/second-side publication, local staging and
Back separately before choosing that change. Assert unchanged query/receipts,
known/unknown geometry, focus/Find/disclosure behavior and retired-handler refusal.
It does not select row virtualization or claim that DOM work is now bounded.

The independent over-limit comparison-request refusal/feedback task remains
separate. Atomic UI publication neither raises nor reconciles HTTP byte budgets.
