# Contextual captured-baseline comparison — architecture review 295

Source review, 2026-10-31, against `9e5b7e1`. Read with the
[design review](CONTEXTUAL_COMPARISON_DESIGN_295.md). No browser execution is claimed
here; actual results belong in VALIDATION and the executable storyboard.

## Decision

Reuse exactly one existing captured-count editor, placing it in a native
**Compare with captured baseline** disclosure beside the applied aggregate. If
that aggregate is hidden, use the visible comparison child; staged or fully hidden
compositions retain the advanced-settings fallback. Do not create a second picker,
comparison controller, endpoint, capture type or query profile.

Mount the editor outside the cached authored display host. `renderAuthored` can
retain that host across redraws, whereas configuration controls must belong to
the current root, execution and visible publication. Choose/remove should use one
reveal helper rather than hardcoded View settings → Composition expansion.

## Ownership and effects

- Opening or closing the disclosure is presentation only: no catalog GET, baseline
  GET, comparison POST, query mutation, capture or save. Loading the catalog remains
  explicit metadata retrieval. Selecting a capture stages its exact immutable ID.
- The command owner must check current root reference, source generation, connected
  visible mount, active workspace, no hidden parent/child, and the existing draft,
  facet, selected-record, paging, comparison-read and live/tail restrictions.
  Retained controls must not become usable merely because Back reveals their old DOM;
  renew the current editor controls, preserving intentional draft/disclosure state.
- Bind asynchronous catalog results to their request and mount/root owner. A retired
  request cannot replace newer options, choose an ID, reopen an old disclosure or
  clear another editor's input. Delayed responses after navigation must be inert.
- Keep `CapturedCount.eligible` and Go query validation unchanged. Unsupported status
  queries may reveal existing Measures · spans, but opening the entry must never
  disable status, normalize a query or rewrite a capture. Numeric/specialized limits
  remain explicit. Actual baseline compatibility is checked when preparing evidence.
- Reuse `prepareCountComparison`: exact stored receipt GET, verified admission,
  canonical pure comparison, and its existing serial/current-result checks. Preparation
  does not query current telemetry. Run remains an independent current read.
- Preserve visibility semantics. Hidden comparison children do not prepare on Run.
  Selecting a reference does not silently expose or prepare a hidden child; any reveal
  must be an explicit action. Do not relax `Composition.visible` checks.

A configured comparison child can sit after bounded records in component order.
After choosing, expose a clear route to its existing **Prepare baseline** action
without starting that action automatically. Reusing that action inline is also sound
if it remains one command under the existing owner, rather than a duplicate controller.

## Acceptance

The natural Activity journey should reach this entry without opening internal
Composition settings, retain its exact query and held evidence, select a UI-created
capture and explicitly prepare the same comparison. Assert no source or metadata
reads merely from opening, the exact preparation request budget, pending-state
refusal, hidden-child behavior, retired controls after child/Back and delayed catalog
responses after navigation. Saving/reopening keeps the declaration and immutable
reference; opening a saved definition does not fetch or prepare its baseline.

The pure query, compatibility and comparison algorithms are unchanged. Reuse their
existing semantic/property checks; actual browser ownership and HTTP accounting are
the important new coverage. No mirror model or generic action framework is needed.

## Implemented source review

The implemented single-disclosure placement follows the above contract. It is
outside cached display hosts, falls back to the visible comparison or settings,
and receives a fresh owner on render and retained Back. Its captured root reference,
source generation, held result and private mount token fence configuration commands.
Current pending, selected-record, draft, page-read and live/tail state is checked
again at invocation. Render-final refresh runs after controls connect.

Catalog results require the current request serial, unchanged title and current
visible owner. Search changes abort the earlier request. An already retired mount's
metadata request may finish, but is discarded; adding a cancellation registry is
not necessary for correctness. Opening the native disclosure changes local display
state only. Unsupported span status exposes the existing measure controls without
changing them. Choose/remove use the common reveal helper; choosing a new reference
never prepares it, and an existing hidden comparison stays hidden.

Source review found no architecture blocker. Actual browser checks, including
retained controls and delayed metadata responses, remain the executing agent's
responsibility. This is not a new source-query or evidence-admission review.
