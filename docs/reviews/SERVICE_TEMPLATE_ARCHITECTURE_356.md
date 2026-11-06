# Service-context authored templates — architecture contract 356

Breadth cut selected by [355 reconciliation](ROADMAP_RECONCILIATION_355.md).
The initial contract below precedes implementation; subsequent source and runtime
acceptance is recorded at the end. Reuse the existing view store, metadata catalog,
`/api/view-templates/expand`, `ViewTemplate` checks and service return path. The explicit action stages the exact question into the
existing population workspace with `recent:false, templates:true`, reveals its
existing catalog and starts its ordinary metadata GET. No separate service chooser
is introduced.

## Destination eligibility and authority

The initial action belongs to a supported record widget after its aggregate
response is checked and is `observed` or `no_data`. Its frozen question must be
valid, page zero and have exactly one existing grouping. Exclude native widgets,
numeric/histogram questions, reported-edge populations and grouped-time series.
Do not merely remove those guards. Both spans and logs can use an ordinary
compatible template; existing HTTP/DB shipped Customize remains independently
available where eligible.

The destination is the complete frozen widget question: signal, exact environment
representation/policy, service/namespace predicates, fixed bounds, original
additional predicates, summary/status, interval width and grouping. The grouping
is already supplied by the widget planner (for example span name, log severity or
an admitted dimension). There is no new default grouping, automatic regrouping,
service-wide denominator or destination profile inferred from the specimen.
Offset remains zero. Do not use loaded row/group values as query authority.

The saved template contributes only its title, checked presentation/composition
and explicit binding requirements. Its specimen service, environment, window,
query, captured values and baseline cannot leak into the destination. The action
is a contributed saved configuration, not an assertion of verified team ownership
or authorization.

## Existing API and compatibility checks

Use the existing paged `GET /api/exploration-views` catalog, then the selected
`GET /api/exploration-views/:id` definition. Confirm the returned immutable ID is
the selected one and run `ViewTemplate.templateView` before expansion. Keep the
selected complete definition detached and stable while inspecting compatibility.
The existing authored request arm is `{source,query}` with an optional explicit
`baselineCapture`; it is exclusive of shipped `{template,version,widget}`.

`ViewTemplate.authoredRequest` and the Go `expandAuthoredTemplate` already check
same signal, exact profile requirements, composition/measure support and explicit
bucket width. The Go path verifies the source content ID before substituting the
destination. `authoredResponse` verifies the selected title, binding, composition,
reference and exact destination question before staging. Do not add a second
expansion implementation or accept only the claimed revision while trusting a
changed definition.

Initial requirements are deliberately bounded:

- Ordinary templates accept only their existing supported count/duration
  aggregate-and-records composition. Required measures must be present on the
  destination; do not enable them implicitly or strip an incompatible child.
- HTTP/DB profile requirements use the existing exact service, span kind and
  nonempty semantic attribute predicates. A name/family label alone is not proof.
- A timeline requires the same explicit existing destination bucket width as the
  specimen and the existing timeline reuse checks. Missing or different intervals
  are visible incompatibilities; no clock/width edit occurs during inspection.
- Comparison templates require an explicit destination baseline capture. This
  initial staged destination has no such selected baseline, so inspection naturally
  reports that requirement until the user explicitly configures it through the
  existing population controls. Existing explicit destination bindings remain
  supported. Do not inherit the specimen baseline, fetch one automatically or add
  a second chooser/default. This is an initial missing binding, not a permanent
  comparison-template capability ban.
- Wrong signal, missing grouping, specialized measures, changed selected
  definition and unrepresentable display requirements fail before adoption.

## Lifecycle and source boundary

The explicit Use action stages the fixed question and reveals the existing
population catalog, including its ordinary automatic metadata GET. Catalog paging,
selecting a definition, inspecting expansion
and staging read configuration only. They read zero telemetry and do not alter
any widget receipt, clock or article. Stage follows the established service
Customize return path into an unapplied fixed question; explicit Run acquires that
population. Back restores the original overview and its independent healthy
results without rereading them.

Capture an immutable service/overview/widget/definition selection token for each
metadata operation. Selection changes, replacement overview, hidden/retired
chooser, owner abort or service departure must prevent late metadata from
publishing or staging. Saved old Inspect/Stage callbacks remain inert after
selection replacement or Back. Abort superseded metadata when possible, using
existing local owner/controller patterns, without a global registry or dashboard
source retry. A metadata error is not an unavailable widget and must not replace
a healthy result slot.

## Acceptance

Use a saved authored template whose specimen scope differs from the destination.
Exercise typed current-only/effective named and an empty or unreported policy,
qualified namespace and literal predicates. Assert exact full destination query,
existing grouping/measures and unchanged fixed bounds; all inspection/staging
source counters stay flat. Explicit Run reads that question, then retained Back
preserves service widget clocks and receipts. Include unsupported baseline,
interval, profile and signal examples plus foreign selected-definition response,
late metadata and retired-action controls. Inspect desktop/narrow chooser and
staged scope. Native/model checks must describe observed actions only.

No new grammar, catalog, permission system, native composition, WYSIWYG editor,
subscription or user-defined JavaScript is part of this cut.

## Candidate source review

The candidate preserves this contract through `templateQuestion` (checked detached
question) and `templateDestination` (supported completed widget plus actual
aggregate-envelope admission). The dashboard only delegates a current visible
secondary action; it does not own another catalog or source acquisition. The
service browser shares its existing return path, fences generation/revision/signal
and blocks template departure while any overview read is pending, preserving the
354 pending-work retirement contract. Investigation/Explore force Stage for this
intent, then use the existing population catalog and authored expansion checks.

One first-entry presentation issue was reported for closure: the library is
nested in the existing View settings disclosure. Opening only the library and
calling `revealFocus` does not open that closed ancestor. The entry must reveal the
existing disclosure path before focusing the library summary; otherwise metadata
loads while the intended chooser remains hidden. This is a bounded reveal fix,
not a new chooser or layout. Source acceptance awaits that correction and the
parent's runtime gates remain separate.

The final candidate opens only existing `details` ancestors on the library-to-panel
path before focusing its summary and issuing the catalog metadata GET. The
first-entry reveal finding is closed. Source acceptance is complete; existing
catalog previews and explicit baseline/configuration workflows retain their own
lifecycle rather than acquiring an additional service chooser. Native verification
must still confirm visibility/focus and zero telemetry reads in the real entry.

The subsequent canonical native, focused actor and observed-model gates pass,
including the visible focused catalogue, zero-source configuration operations,
exact destination reads, Save/reopen and retained return. See the current
[validation record](../../VALIDATION.md) and
[storyboard](../journeys/356-service-templates.md).
