# Record-list architecture review — 297

2026-11-01. Source review of the 296 checkpoint, `record-page`, source admission,
record presentation/results, composition validation and scoped tables. This is a
reviewed implementation contract; executed results belong in VALIDATION and the
297 storyboard.

## Decision

Add `record-list` version 1 as a second kind of the existing **records** child.
Do not add a selected-record role, another inspector owner, a new query endpoint,
or a rich-record-to-ranked-dataset conversion. The existing page and query owner
continues to own pagination, count, scope, read lifecycle and source actions.

The first list is display-only. A compact item shows configured UTC/service,
summary, qualified fields and optional supplied trace context. Native Details can
reuse the existing passive structured attribute/exception presentation. It cannot
fetch a trace, refine a query, Keep a row, or infer a primary selected record.
A quick Table choice retains those established application-owned workflows.
Switching is refused while existing selections, submissions, reads or drafts make
replacement ambiguous; it must not silently discard a pending Keep workflow.

## Data and identity

- Consume the same deeply immutable `goo.record-page/1` as record-table. The page
  carries exact delivery question, basis, UTC read clock, rows, offset, limit,
  returned length, reported/unknown continuation and bounded-page quality.
- A public page adopter may validate serialized input for passive display, but
  that does not create source authority. Application integration must still use
  `SourceSnapshots.getRecordPage`, whose private WeakMap binds page and wrapper.
- Validate plain data descriptors and existing JSON budgets before property reads;
  reject accessors, cycles, sparse arrays, unknown roots and forged tokens or
  completeness fields. Metadata must agree with the question and supplied rows.
  Reusing constructor-validated immutable pages can avoid repeated large copies.
- Page positions identify observations within this read only. Equal records may
  occur twice, and reported trace/span IDs do not establish unique events. Do not
  persist an item index as an event ID or carry Details state onto a new page.
- Returned rows are not a population denominator. Unknown continuation remains
  unknown, and no-data/changed-between-reads stays distinct from healthy/failed.
  Reported ancestry uses only supplied page rows, with existing ambiguity/missing
  parent handling; it cannot infer completeness or fetch more context.

## Declaration and ownership

Reuse `RecordViewDeclaration` scope/version and the existing `columns` plus at
most eight qualified `fieldColumns`. UTC and summary remain required. Record-list
uses the same field lookup and human UTC/duration formatting, preserving field
namespace/case and missing, null, empty and literal values. `context` means passive
reported trace context in List, not the table's source-action buttons.

Go and JS validators admit only pinned `record-table` or `record-list` names.
Save/reopen, exact composition reuse and opted-in template expansion must retain
the kind. Existing display-layout updates must not reset it. Preview should derive
the actual records kind rather than hardcode record-table. No persisted row
selection, source rows, token or witness is added to the declaration.

The native Table/List selector is an application command: capture current root,
generation, record response/page owner and visible mount. Recheck pending-state
and ownership at invocation. Changing kind is local configuration, with no read,
save, capture, offset change or aggregate replacement. Old controls after page,
rerun, child navigation or redraw remain retired. Back may renew only the current
selector through the existing retained owner.

## Selected-measure table readability

The scoped table currently chooses all source fields in source order, putting p95
off the initial viewport. Fix presentation order only: declared dimensions (or
explicit category), then the selected value, then every remaining supplied field
in original order without duplicates. Keep fallback behavior for unmapped shapes;
histogram bounds, time bounds, interval state, support, exact values and action
context remain available. Do not create a projected dataset, lose fields, alter
row order/counts, or transfer completeness to a new subset. All existing runtime
eligibility and private action bindings remain attached to the original snapshot.

## Acceptance

Use actual log and span pages, including same-named resource/attribute fields,
missing/empty/null values and duplicate records. Check passive mount/disclosure
and Table/List switching with no IO; pending selection refusal; explicit paging
retaining aggregates; saved/template kind round trips; and retired selector refusal.
Standalone serialized and detached-property rendering must work without granting
source actions. Property tests should exercise the actual page adopter and kind
validators, not a duplicate planner. The browser must verify the selected measure's
header **and cell** are in the initial viewport, beyond merely finding their text.

## Backend implementation review

The implemented Go change admits only the second pinned name in the existing
record declaration validator. Columns, qualified fields, query validation and
composition roles are unchanged. Generated OpenAPI uses the same two-name enum;
there is no new endpoint, schema family or source operation. Focused tests cover
immutable save/load, strict composition reuse, authored template rebinding,
detachment, invalid fields/kinds and HTTP configuration with source transport
forbidden. No backend architecture blocker found. Application lifecycle and
passive-renderer verification remain separate gates.

## Frontend and passive-input implementation review

The implemented public page adopter validates serialized plain data with the same
budgets, verifies derived page metadata, and refuses unknown continuation coercion,
forged completeness/tokens and no-data pages containing rows. Its WeakSet permits
sharing only already immutable constructor outputs; this is passive structural
integrity, not the private SourceSnapshots admission binding. No new source
capability is created. The list uses actual record-list declaration validation,
withdraws stale observed/signal metadata on invalid input, and lazily renders
Details only for the current connected page. It emits no query/action events.
Qualified fields share the table formatter and distinguish missing/null/empty,
number/boolean versus literal strings, namespaces, UTC clocks and durations.

The application uses one Records view selector with current root/epoch/mount and
pending-state guards. A first review found that retained Use Table/field-picker
closures would become dead after Back; the implementation now replaces those
metadata controls with fresh current-page owners. Their older nodes remain
retired. The supplied list and native Details can stay retained, while older
source pager/table controls are deliberately not blanket-reactivated. Changing
kind creates no new selection owner, source read or data authority.

The selected-measure table helper reorders field references only, retains every
supplied field, and anchors declared time/histogram bounds before the measure.
Rows, snapshots, proofs and action intents are unchanged. Backend, passive-input
and application source reviews have no remaining architecture blocker; actual
browser, standalone lifecycle and viewport checks belong to the executing owner.

A subsequent structural review found two real retirement/memory hazards before
release. Record-row actions now also require the exact current
`entry.recordTableOwner`, so reattaching an old table cannot revive its actions
when query generation and held data happen to be unchanged. That owner is assigned
before any renderer callback can invoke it, avoiding a temporal-dead-zone failure.
The list also computes bounded ID-only ancestry metadata once, rather than passing
the entire rich page into each context component's cloning setter. Per-item
context and lazy Details use that compact ancestry input; original records remain
available to Details. This avoids quadratic rich-payload copying without changing
page data, duplicate handling or source guarantees. Structural tests cover large
attributes/malformed identifiers rather than relying on flaky timing assertions.

The shared ancestry zero-parent check now requires a string, so malformed arrays
cannot coerce to a genuine all-zero parent and object conversion cannot throw.
The compact metadata projection and full supplied-record presentation therefore
retain the same invalid/unreported/no-parent distinctions, including malformed
inputs. Valid reported IDs are unchanged. Final source review has no architecture
blockers; the executing owner still supplies live browser and viewport evidence.
