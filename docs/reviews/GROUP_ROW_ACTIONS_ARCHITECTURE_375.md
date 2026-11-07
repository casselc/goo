# Group row identity and actions — architecture 375

Status: source inspection and bounded proposal only. No production change or visual acceptance is claimed.

## Existing owner and problem

`RecordResults.groups` in `ui/record-results.mjs` owns the native grouped table used by ordinary multi-field grouped results and typed error-triage follow-through. Single-field canonical datasets without extra actions already use `ScopedBreakdown`; that path should remain unchanged.

The fallback table already contains a caption, native `table`/`thead`/`tbody`/`tr`/`th`/`td` elements and column headers with `scope=col`. It renders dimensions in declared order, captured count, optional numeric/duration measures, then action anchors. Error-triage supplies **Inspect group** and **Compare within service** from `Exploration.groupActions`. No separate triage table is needed.

370 established exact typed continuation and retained Back ownership, but its 390px evidence needed separate horizontally scrolled crops for identity versus actions. 375 should show each group's service/type identity, count and actions together without horizontal scrolling. It must not change grouping, scope, reads, or the meaning of a captured count.

## Preferred existing-component reuse

Opt this existing `goo-table` into its existing `mobile-layout=stacked` presentation. `ui/components.css` already defines a two-column narrow layout at 700px, wraps exact values, keeps the native header in the DOM and gives the last cell a full-width action area. Add `data-label` to each cell from the same header specification that produces its actual column header; do not maintain a second list of mobile labels.

The initial implementation can remain in the existing `groups` renderer with a small shared cell/header description, rather than extracting a new row framework. Keep existing desktop columns and existing row/anchor DOM. Scope any needed action spacing to the grouped table; use the shared focus treatment and compact density. Do not add a generic card component, duplicate mobile markup, second set of anchors, or an alternate selection manager.

At 390px, the two dimension values occupy a readable identity area, followed by the labeled count and primary actions within that same card. Inspect remains first, Compare second. If optional measure cells push the actions too far from identity/count, a narrowly scoped grid ordering rule may place the existing count/action cells before auxiliary measures. That is a presentation choice for the design review, not a reason to move or rebuild their owners. For the primary service/type case without optional measures, the existing stacked mechanism should suffice without a new layout API.

Alternatives considered:

- Sticky first/last desktop columns preserve a table but leave too little space for both values and long action labels at 390px.
- A second mobile card tree duplicates links, keyboard stops and leases and complicates retained focus. Avoid it.
- Collapsing all dimensions into a single desktop identity column reduces comparison across fields and is unnecessary when the shared responsive opt-in already exists.

## Labels and accessibility

Keep literal text via `textContent`, including quotes, punctuation and long service/type values. Preserve `FieldPresentation.missingLabel(field)` and the distinct **Present empty string** label; do not substitute a dash, truncate, or use tooltip-only identity. Exception-type columns retain **No unique reported exception type** where that is the existing meaning. Preserve exact field qualification (`column:service`, `attribute:exception.type`, and so on) in cell labels and headers.

Native anchors remain links with their checked question href, not role-changing buttons or row-wide click targets. Their staged-read explanation remains available. The visible action labels already communicate different continuations; optional accessible names may add exact group context if a demonstrated screen-reader ambiguity requires it, without inventing a single-string identity or conflating missing/empty values.

Keep the native caption, column headers and associations when visually stacking. CSS `display` changes can affect accessibility trees in some browser/assistive-technology combinations; inspect the actual browser accessibility representation and native header relationships before claiming accessibility acceptance. If explicit `headers` references or roles are required, add only those supported by the observed result. Do not replace the table with an ARIA grid, add arrow-key navigation, or remove headers with `display:none`.

Keyboard order remains the existing DOM order, with Inspect before Compare. Focus must remain visibly contained in the card without horizontal scrolling. Any optional visual reordering must not create a conflicting sequence of focusable controls. The `goo-table` region's existing keyboard behavior remains intact.

## Ownership and retained navigation

375 must preserve the existing `groups` anchor lease unchanged. A click requires its captured lease, current parent result/group membership, connected and contained DOM, visible/non-inert anchor and current caller authority. Old callbacks prevent default and stop immediate propagation before the trail delegate, including reattached old anchors. Styling cannot confer authority.

`RecordResults` already retires the previous grouped owner on replace/clear/suspend. Typed triage Back deliberately rebinds the same retained anchors only after its current question, evidence basis, generation and held-result checks. Reflow must not replace those anchors, rerender the retained parent on return, issue a source read, or broaden resume to distribution proofs. Keep modified-click, copied links and explicit child Run semantics.

## Bounded verification proposal

Use the existing typed-triage 370 journey and actual grouped-row owner fixtures as the starting point, with one ordinary multi-field grouped result as shared-component coverage. Check:

- At 390px, one actual row simultaneously shows service/type, exact count, Inspect and Compare within the viewport width; no separate horizontal crops or page scrolling to unrelated controls.
- Literal long labels, missing and present-empty values remain distinguishable; optional numeric/duration labels remain associated with their values. Check desktop columns too.
- Keyboard reaches both original anchors with visible focus, activates the staged child with zero source reads, then explicit Run reads the child. Back retains original row/anchor DOM, evidence and focus with zero reads.
- Captured stale, hidden, inert, replaced and reattached anchor callbacks still refuse navigation. Style-only reflow does not alter the established lease contract.
- The browser accessibility representation retains table, caption, column headings and labeled values; automated markup assertions alone are insufficient evidence for a screen-reader claim.

Keep screenshots and storyboard text current with the accepted layout. No source/query/compiler/backend changes are needed for this proposal.
