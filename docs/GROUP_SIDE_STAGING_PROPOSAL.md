# Explicit group-side question staging — proposal history

> Implemented in292. The [current contract](GROUP_SIDE_STAGING.md) and
> [validation](../VALIDATION.md) record delivered behavior; the reviewed proposal
> below is retained as design history.

**Status: proposed, unimplemented.** This bounded contract follows the delivered
[group-count comparison](GROUP_COUNT_COMPARISON.md). It lets a responder choose
one reported group on either the baseline or current side and stage that side's
question in the existing retained child. Staging performs no source, storage or
comparison request. Only the child's explicit Run acquires new telemetry.

## Reuse and authority boundary

The existing `groupIntent` applies to ranked datasets. The group-count profile
instead carries summary operands and an aligned group union. Do not manufacture
a ranked dataset or route its summary through that helper. Existing H1
`scope-actions` guards also deliberately refuse baseline-role refinement; do not
relax them globally. Baseline staging here is an explicitly labelled **question
intent**, not authority to refine captured evidence or expose original stored rows.

Reuse the context vocabulary already emitted by `view-schema`:

```json
{"type":"scope-group-side","side":"baseline","key":"1checkout",
 "datasetId":"<chosen side summary ID>","revision":"<chosen side revision>",
 "scopeId":"<chosen side scope ID>","generation":"<chosen side scope generation>"}
```

`side` is exactly `baseline` or `selection` (the latter labelled Current in the
interface). `key` is the canonical typed group key, `0` for null/missing or `1`
plus the literal string. The intent contains no free-form query, category value,
field selector, capture ID or source permission. `generation` is the existing
snapshot scope-generation string; the application owner's read/action epoch is
separate and private.

A passive component may pass this closed intent from a positive count only when
a host explicitly opts into staging through a nonserialized property binding. Standalone displays remain static by
default. Bars remain passive, and there is no row-wide default-side gesture.
The accessible button name identifies the chosen side and group; visible count
text can be the compact activation target. It never fetches, navigates or builds
an executable source query by itself.

## Eligibility and revalidation

The host accepts an intent only while all of the following remain current:

1. The exact admitted prepared comparison pair, baseline reference, held current
   result and chosen side summary still belong to this active population owner.
   Both summary contexts and the intent's chosen side tuple agree. The grouping
   field and group key resolve through that admitted full union, not the visible
   filtered/page subset or event-supplied values.
2. The whole pair is coherent and non-diagnostic. Both supplied side partitions
   have coherent accounting. The chosen row cell is `present`, `derived` and a
   positive canonical count. An aligned absent zero, unknown/unavailable cell or
   diagnostic pair has no staging action. A `no_data` envelope with a legitimately
   empty opposite side may still stage the positive present chosen side.
3. The action owner, element and containing comparison are connected, visible and
   active. No child is already open, read/preparation/page operation is pending,
   facet/query/source draft is unresolved, record selection is pending, navigation
   retired the owner, or nesting limit prevents another retained child. These
   guards belong to the application, not presentation eligibility.

Membership or coherent sums alone do not create source attestations. The host
binds the intent to the existing admitted response and its current lifecycle.
The public context tuple is descriptive, not event authentication. Do not install
an unrestricted DOM-event listener that grants staging authority merely because
a dispatched tuple currently matches. Use a property-only host callback/action
binding with an opaque private owner lease. Each rendered handler captures that
binding and epoch; it must not dynamically consult a newer `element.interaction`
when an old detached button is clicked. Redraw/rebind/disconnect retires those
handlers. Serialized JSON or a forged `dispatchEvent` cannot opt into authority.

Changing side while keeping another side's context, using an unknown key or
clicking a retired handle must refuse without staging, discarding drafts or
reading telemetry. A descriptive intent tuple may remain semantically valid
against the same held pair after Back; the pure planner does not pretend to know
whether an old UI event was authorized. Runtime callback/private-lease retirement
provides that temporal refusal, separately from pure semantic validation.

## Exact query derivation

Resolve the selected side's original checked `scope.selector`; do not use the
latest inspection page or unconditionally choose the current parent query.
Clone it, set only `offset` to zero, and retain its original `limit`, half-open
UTC window, signal, typed environment policy, predicates, grouping, bucket width
and every summary/measure directive.

Resolve the grouping field and category from the admitted canonical row. Build
one typed predicate:

- null/missing category: `{field, op: "missing"}`;
- present string, including empty string: `{field, op: "eq", value}`.

If there is no original predicate, use that leaf. Otherwise wrap the original
whole predicate and leaf in `all: [originalWhere, leaf]`. Never flatten an OR
branch, replace an existing condition, convert missing into empty, remove a
filter to make a group selectable or add the opposite side's window. Validate
the resulting full query. If the added predicate exceeds existing node/depth or
other semantic limits, refuse explicitly; do not silently simplify or widen it.
The original comparison data and query remain unchanged and detached.

No new endpoint, query-bearing chart child, saved declaration kind, snapshot token
or durable evidence object is required. A small pure Go `StageGroupSide` planner should accept the
original `ComparisonRequest` plus a `GroupSideIntent`, call existing comparison
validation to reconstruct its canonical pair, check the context and selected row,
and return the exact query. The browser counterpart accepts the admitted canonical
comparison and intent and performs equivalent checked derivation. Generated tests
compare these outputs. The Go function is a semantic reference/library helper;
this cut does not add a planner HTTP operation or make staging a server read.

## Retained child and Back

Use the existing retained-child population staging path. Opening the child pauses
the parent and stages exactly one chosen-side query; it does not Run it. The child
message identifies Baseline or Current and the group, and states that Run is an
independent fixed read. Exploring a baseline question is not filtering original
capture rows. Late arrival or retention may change its future result.

The child uses the existing ordinary query presentation. It does **not** inherit
the parent's captured comparison child or baseline reference: adding a group
predicate changes the root definition and would make that baseline incompatible.
The parent retains its complete comparison unchanged.

The parent pair, passive Summary/Groups setting, Find/order/display page and fixed
aggregate clocks remain held. Returning with Back makes no read. Because opening
an existing child increments the source generation, Back must establish a **fresh
specific group-action owner/controls** over the still-held admitted pair. Old
buttons, intents or queued callbacks remain retired permanently. Do not rearm the
old source epoch merely to make cached component controls work. Passive display
state can be copied/restored or kept independently of those retired action handles.

Run/rebaseline/navigation, an incompatible reused definition, baseline drafts,
hiding or replacing the comparison and source owner suspension retain their normal
retirement behavior. Config-only preservation after Back remains available through
its existing dynamic owner guard; this contract does not broaden generic H1 source
action authority.

## Proposed journey and verification

From prepared Groups, choose a known positive Baseline count. Observe a staged
retained child with the exact baseline window and old predicates plus its typed
group predicate, with zero new requests. Run reads independently. Back restores
the unchanged held pair/Find/page with no read; a retained old staging button does
nothing, while a new Current-side control stages the current question. Repeat for
null/missing and present empty-string values, proving different AST leaves.

Cover absent zeros, a known-empty opposite side, diagnostics, wrong side/context,
unknown keys, hidden/detached/retired controls, navigation, pending records/facets/
source drafts, depth limit and final-query predicate-budget refusal. A late
inspection page must never become the staged root or override its held scope.

Go/JavaScript properties should compare exact detached derived queries across typed
environment policies, windows, paging, where/OR structures and category identities.
Quint should distinguish prepared owner, retired old action owner, staged child,
explicit Run and returned fresh owner; negative traces reject implicit reads and
rearmed old controls. Bombadil must independently observe child staging, explicit
Run and successful Back before completing. Update real screenshots/storyboards and
source/visual review after implementation; this proposal claims none have run.
