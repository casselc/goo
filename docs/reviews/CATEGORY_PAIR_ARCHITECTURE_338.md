# Category populations across fixed operands — 338

Status: source, generated semantics, native logs/spans, observed model and actor
acceptance pass. Actual frames preserve the stated limits.

## Reuse the existing source boundary

[337](PAIRED_FACET_ARCHITECTURE_337.md) admits a categorical facet on a fresh
fixed-query count comparison and defines a checked common-filter delta. Keep
that grammar. Baseline-only values need a second independent counted-values
read, not a second query language, a loaded-row aggregation or a new Go route.

`exploration/counted_values.go` already reads a full count and a categorical
group partition separately. `exploration/compiler.go` requests `LIMIT 1001`;
the source rejects more than 1,000 groups rather than returning a truncated
partition. It checks canonical uint64 totals, positive unique group counts,
nullable field semantics and exact conservation. Unequal count/group sums yield
`changed_between_reads`, whose groups are diagnostic only. The existing
`ui/filter-values.mjs` rejects that state and admits only `observed`/`no_data`.

A ready operand therefore supplies a bounded complete grouping of **observed
matching source records**, with independent count/group read receipts. It is
not a transactional snapshot and does not attest upstream telemetry completeness.
Its total can differ from a previously acquired comparison total; new facet
observations must not overwrite or silently rebase the held comparison facts.
No Go dependency/runtime change is required by this cut; the checked baseline
remains Go 1.27.3, pgx 5.12.0 and x/text 0.43.0.

## Passive model and alignment

The existing categorical facet supplied model keeps selected `values` and adds
optional `baseline` observation state. The small pure `categorical-pair.mjs`
projection validates and aligns those operands. Keep each side's state (`unloaded`, `loading`, `error`, `ready`),
field, total, exact items and independent read receipts. Do not convert an error
or a changing/incomplete source result into an empty ready list.

The projection unions exact literal identities from ready operands. Use the
existing `categoryKey` semantics: null is missing, while every string, including
empty, `"null"`, Unicode and key-like text, has a distinct identity. At most 1,000
items belong to either side and at most 2,000 to their union. Use bigint for exact
counts, conservation and signed differences; bounded pixel ratios alone may use
Number. Local search/order/pagination never affect measurement totals or read.

| Operand state | Category membership | Cell semantics |
| --- | --- | --- |
| ready, reported | present | Positive observed count |
| ready, not reported | absent from this supplied partition | Derived zero for observed matching records |
| ready, total zero | no reported categories | Derived zero opposite a reported category |
| unloaded/loading/error | unknown | Unavailable, never zero |

A row exists only if at least one ready operand reports it. If neither side is
ready, show acquisition/error states rather than inventing rows. The projection
marks `complete` only when both partitions are ready; that internal flag does not
attest upstream telemetry completeness. No numeric delta, percentage or rate is
supplied in 338. An unavailable side has no numeric bar and cannot be ranked as
a zero. Known cells can retain their bars on the shared available-count scale. A healthy ready
side can remain visible when its sibling fails. Even exact aligned zeros retain
the overall **upstream coverage unknown** qualifier.

Reuse `comparison-visual.mjs` shared zero-based count cells and existing bounded
category preview helpers. Do not manufacture `goo.group-count-comparison/1`
held snapshots to enter `goo-group-count-comparison`: that model carries its own
snapshot/profile/accounting contract, separate from counted-value receipts.
A common visual primitive is appropriate; a forged source-authority conversion
is not. The facet remains a passive property/HTML renderer with no fetch or
subscription and one Include/Exclude intent for each exact union value.

## Application ownership

The population owner retains both original questions. Build selected and baseline
requests through `FilterValues.query`; each preserves its own predicate, exact
window and checked environment policy, changing only the existing categorical
projection. One explicit Read both windows operation initiates two endpoint
requests. Each endpoint owns two SQL reads: report HTTP effects separately from
physical statements. Adding/reopening a facet and rendering/searching it read
nothing.

Use two existing `FilterValues.mount` owners under one pair visit. A pair key
must include both full projected question identities and the qualified field;
using the selected key alone would accept old baseline observations after a
baseline recipe edit. Current-owner checks also include generation, element,
visibility, current composition and child/active state. Retire before abort;
abort-insensitive late successes/errors cannot attach after departure, field or
operand changes, or an ABA revisit. Independent errors preserve sibling facts
without silently making a new pair claim. Retry remains explicit and read counts
must state which operand it reacquires.

Staging accepts a value only from a current ready operand's exact union, applies
the existing 337 `narrowFixedComparison` planner to both whole original scopes,
and reads nothing. Baseline-only values are as actionable as selected-only ones.
Other-field gestures still refuse; pending choices still disable parent Run and
its retained callback. Both operand definitions must admit the delta before a
child appears. A child starts with unloaded values and independent explicit Run.
Back restores the retained parent observations/local Find and renews listeners;
old reattached controls remain inert. Saved/YAML definitions contain neither
loaded partitions, pending choices, state, receipts nor source requests.

## Acceptance gates

- Pure generated cases cover union symmetry, exact conservation, missing/empty,
  Unicode/key-like literals, baseline-only and selected-only values, empty sides,
  uint64 boundaries and a 2,000-row union. Mutating a missing failed-side cell to
  zero must fail; corrupt receipts/field/counts and invented union rows refuse.
- Owner schedules cover independently failed/late operands, pair-key changes,
  retry, hide/departure/Back and ABA. A matched total is not permission to reuse
  a different operand's old observation. Preserve the unchanged 337 common-filter
  semantic truth-table and budget refusal coverage.
- The live logs/spans journey needs a baseline-only category, exact independently
  observed counts, missing and present-empty, a common filter from the union,
  unchanged held parent facts, source-free staging/return, and save/reopen.
- Update Bombadil actions/invariants and observed Quint effects to require both
  value reads and a verified terminal state. Model effects do not prove predicate
  authority; source requests/returned records and generated semantic tests do.
- Inspect actual desktop/narrow viewports. A driver DOM completion is not a
  screenshot guarantee; retain the 337 capture limitation until independently
  resolved. Document pending work rather than presenting proposed gates as passed.

## Final source review

The application helper owns two existing `FilterValues.mount` instances, with a
whole-question pair key. Response admission and facet gestures both check the
live pair identity. Failed-side Retry preserves the healthy operand and rechecks
ownership after a potentially reentrant pause callback. Retained partial/error
state rehydrates a bound owner on Back, so Retry can issue only the failed read.

Review found a pending-cache retirement defect before native execution: local
rerender disposed the old owner after clearing its authority, causing invalidation
to leave cached `loading` states with no live request. The bounded fix normalizes
only the exact retired owner's loading sides to unloaded before disposal, retaining
ready/error witnesses. The subsequent native pending-read hide/reveal check proves that the public
control is usable and the retired response cannot publish, as recorded below.
Generated/model checks and actual frames provide separate acceptance evidence.

The source-grounded projection checks pass: forty Hegel-generated cases execute the
actual JavaScript module and compare each exact cell, typed union, maximum and
readiness flag against an independent Go membership oracle. Metamorphic checks
swap the operands and protect detached inputs; malformed states, field/count/
receipt mutations, duplicate values, invalid Unicode and a 1,001-value operand
refuse. A separate disjoint 1,000+1,000 fixture retains all 2,000 union rows.
The 21-step effect model checks nine counter/effect mutants through Quint;
preservation and derived endpoint-budget guards reject two additional ledgers,
and a shortened prefix is refused. These do not replace real request/record or
viewport acceptance.

## Actual-frame review and next consolidation

Architecture independently inspected the current desktop pair, partial-read and
narrow viewports. Baseline-only DELETE is aligned as selected zero/baseline two;
missing and present-empty remain distinct. In the planted baseline read failure,
the selected facet count stays 27 while baseline values remain unavailable. The
separate held comparison still shows 27/4 with its own facts; that is retained
prior evidence, not a replacement of the failed facet observation. Narrow operand
labels and counts remain readable without horizontal overflow; the lowest rows
require vertical scrolling. These frames do not establish pristine entry height.

The next consolidation candidate is presentation of repeated page-level fixed
scope/window/policy facts and task advice. Keep the count-pair/facet components
self-contained when independently embedded, and preserve separate observed
receipts. Reuse the existing UTC/window formatting and display primitives before
inventing a scope manager. Measure a natural initial viewport and keyboard/focus
behavior for any 339 density claim; the task-scrolled 338 frame and its prior
wrong-field advice are not a default-entry baseline.

A smaller optional shared-cell cleanup could represent a known count separately
from unavailable comparison geometry, rather than passing an unavailable numeric
value and restoring the visible count through wrapper metadata. Select it only
with a concrete reuse benefit. The two-sided 1,000-value limits are cardinality
bounds, not a proven combined byte/DOM budget; do not claim worst-case property/
HTML-attribute serialization parity without a separately measured budget check.

## Final observed acceptance

The native 338 journey passes in 4.2 minutes for logs and spans: 107/13, baseline-only
DELETE 0/2, normal exclusion 32/8, inclusion 27/4 and reopened 27/4. Supporting
previous-window API records return three POST rows. Independent raw review verifies
the 21-step ledger and terminal counters: eleven record reads, four value lookups,
five comparisons, one Save and one encode, with fifteen source requests. Go/Quint
replay passes in 29.934 seconds.

The supplemental real-response checks prove independent baseline failure, one-read
Retry with the healthy selected witness unchanged, local Find and Back without
reads, old-control refusal, and retirement of a held baseline counted response.
Hide/show preserves ready selected values while normalizing abandoned baseline
loading to unloaded; late completion cannot publish, and explicit Read reacquires
two operand lookups. These seven supplemental requests are outside the ledger.

The 35-second actor passes with 461 snapshots, phases 0–7, four verified transitions,
six source requests including two value lookups, two comparisons, zero writes and
no overflow; the parent pair remains immutable. Driver screenshots are separately
audited: four unique images across 461 captures, with 453 repeating a stale child
loading frame, including the terminal capture. They do not prove visual completion;
reviewed Playwright viewports supply that evidence. No upstream telemetry coverage,
transactional snapshot, arbitrary cohort or worst-case DOM budget is claimed.
