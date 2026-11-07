# Group row identity and actions — architecture 375

Status: delivered; source, matched browser frames, accessibility diagnostics and observed replay accepted. See [validation](../../VALIDATION.md).

## Existing owner and problem

`RecordResults.groups` in `ui/record-results.mjs` owns the fallback native grouped table, including scalar/typed error-triage results with explicit actions. Its early branch delegates **any** supplied `viewDataset` without extra actions to `ScopedBreakdown`, including the actual ordinary two-field journey. That separate path is unchanged.

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

Use the existing typed-triage 370 journey and actual grouped-row owner fixtures as the starting point, with one ordinary multi-field grouped result as an unchanged-presenter regression. The genuine ordinary result uses ScopedBreakdown, not the changed fallback. Check:

- At 390px, one actual row simultaneously shows service/type, exact count, Inspect and Compare within the viewport width; no separate horizontal crops or page scrolling to unrelated controls.
- Literal long labels, missing and present-empty values remain distinguishable; optional numeric/duration labels remain associated with their values. Check desktop columns too.
- Keyboard reaches both original anchors with visible focus, activates the staged child with zero source reads, then explicit Run reads the child. Back retains original row/anchor DOM, evidence and focus with zero reads.
- Captured stale, hidden, inert, replaced and reattached anchor callbacks still refuse navigation. Style-only reflow does not alter the established lease contract.
- The browser accessibility representation retains table, caption, column headings and labeled values; automated markup assertions alone are insufficient evidence for a screen-reader claim.

Keep screenshots and storyboard text current with the accepted layout. No source/query/compiler/backend changes are needed for this proposal.


## Independent candidate source verdict

The actual renderer now opts the existing `goo-table` into stacked mobile layout.
One headings array produces native column headers and every cell's data-label;
sequential appends preserve numeric mean/coverage, duration and action associations.
Literal, missing and present-empty values retain their prior formatting. The
ScopedBreakdown early path is unchanged. No source/query/row-count transformation
or duplicate action tree was introduced.

The actual action-anchor creation, captured leases, current-group admission and
suspend/resume code remain unchanged. Reflow adds no authority and does not rebuild
the held owner on Back. Independent source review finds no material blocker.
Actual narrow geometry, keyboard focus and browser accessibility/header
associations still require the planned runtime checks; markup alone does not
establish assistive-technology support.

Final candidate additions are also source-accepted: only literal strings equal to
existing missing/empty/exception marker labels are visibly quoted; raw group values
and href predicates remain unchanged. Narrow action CSS wraps the existing links,
provides 28-pixel minimum target height and hides only decorative line breaks.
There is no duplicate action tree, tab-order change or lease change. Actual visual
and accessibility acceptance remains separate from this source verdict.


## Initial presenter correction (superseded by extended scope below)

The genuine baseline confirmed ordinary two-field results contain `viewDataset`
and no extra actions, so the early ScopedBreakdown branch wins. Its `multi` default
is the shared table kind; `scope-view.mjs:table` creates a compact goo-table without
the stacked opt-in. 375 does not change that renderer or claim its ordinary mobile
presentation improved. The primary scalar/typed triage results supply actions and
therefore exercise the changed fallback. Supplied fallback measure/label tests are
component coverage, not evidence that the genuine ordinary route used it.

Keep the actual presenter/AX evidence for each case in validation. Ordinary scope,
selection and source budgets are regression checks; visual improvement and native
header/action association claims belong to the actual changed triage path. No
additional production scope is required by this correction.

## Extended scope after genuine baseline

The actual ordinary presenter did not fit at 390 px: its canonical table measured
724 px scroll width in a 262 px client region. Triage fallback tables also overflowed
(605/322 and 595/322); desktop cases fit. The ordinary case therefore cannot be
reported as an accepted fit regression or removed from acceptance.

Root approved a narrow extension: the passive goo-chart host can opt into the
existing stacked table layout and forward that presentation attribute to its
inner goo-table; cell labels derive from actual native headers. ScopedBreakdown
opts its grouped table into that presentation. This adds no dataset grammar,
query, source ownership or second row tree. Final source review and actual
ordinary/triage narrow geometry remain required before acceptance.

The frozen canonical extension is source-accepted. RecordResults opts only its
existing no-extra-actions ScopedBreakdown path into `mobileLayout:'stacked'`.
That sets a passive goo-chart attribute, whose existing custom-element attribute
lifecycle rerenders and forwards only the literal stacked value to the inner
goo-table. Other callers retain default layout. Every cell label comes from the
actual rendered native header, including Share, interval state and Selection;
there is no second label catalogue. Dataset/display grammar, selection intent,
source proof and callback handlers are unchanged. No MutationObserver/resize
observer or new owner is introduced. Actual ordinary narrow fit is now a changed-
path acceptance target alongside triage, not merely an unchanged regression.
