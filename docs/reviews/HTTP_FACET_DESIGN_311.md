# HTTP service view with a declared method facet — design review 311

Status: implemented source and all three actual browser frames accepted.
Read with [HTTP service Customize](../HTTP_SERVICE_VIEW_TEMPLATE.md) and
[declarative facets](DECLARATIVE_FACET_DESIGN_310.md).

## User outcome

A responder discovers an HTTP service, customizes its useful existing route view
and narrows it by reported method without building a control from scratch. This
is adoption of the new grammar control in an existing default, not another
service dashboard, profile, catalog or editor.

Use an **HTTP methods** facet over qualified
`attribute:http.request.method`, beside the existing route count chart and
records. Method complements the chart's `attribute:http.route` grouping instead
of duplicating the same route selection twice. The facet comes before long
records so it is discoverable. Existing authored/YAML ordering remains respected.

## Existing discovery and template boundary

Keep the current HTTP server routes widget's Customize view entry and pure
checked template expansion. Preserve its complete checked question, including
exact service/namespace/environment, fixed bounds, required reported Server kind,
nonempty current method and any additional predicates. Do not reconstruct a
wider scope from the service name, union legacy aliases or claim method presence
certifies OTel conformance. Native metrics and unrelated RPC/database/messaging
profiles do not acquire a method facet through this cut.

Customize stages configuration only. Existing Run acquires the ordinary route
chart and bounded records; **Read values** separately acquires method values in
that current applied scope. Neither template expansion nor facet mounting loads
values. A no-data widget remains usable configuration without implying that the
service is healthy or fully observed.

## Minimal interaction and interpretation

Use the existing facet's Read values, Include/Exclude, Clear staged facet and
shared selection tray. No new per-service filter toolbar is needed. Display
reported method strings as data, including supported custom nonempty methods;
never reinterpret arbitrary methods as verified request types or outcomes.
Counts describe captured server-span records, not unique requests or rates.

The chart remains grouped by route while method selection stages a checked
predicate. Existing filters, environment and fixed window remain. Opening a
selected child stages it; Run reads the new scope explicitly. The facet then
needs an explicit values read for that applied scope. Do not show the parent's
method counts as current child values or reset unrelated filters when clearing
this facet's pending choices.

Record fields, Table/List/Waterfall choice where compatible, Inspect/Keep and
parent return use the existing controls. Save/export/import preserve one ordinary
view definition. Reuse applies the admitted presentation/field control to a
compatible destination through existing category/scope checks; no source counts,
read clocks or selection token travel with it.

## Acceptance as one real service task

Discover an actual ordinary-OTel HTTP service and open its current route widget.
Customize through the existing entry, verify the staged facet/chart/records
composition and no telemetry on expansion. Run; then read methods explicitly.
Show multiple methods across routes so the control adds meaningful scoping rather
than a single-value demo. Select one method and use the existing staged child/Run
path; retain the exact server-span population and additional filters.

Verify route counts and bounded record counts separately, with UTC/read/support
facts. Return to the service/parent without rereading or resuming hidden activity.
Save and reopen/export the composed view without implicit values reads, then use
its ordinary compatible reuse path. Include no-data, stale template response and
wrong-profile controls. Review actual compact frames for discoverability,
method-versus-route meaning, staged scope and retained records. Lifecycle/model
checks separately establish effects and completion.

## Next substantive checkpoint

After this first operational adoption, revisit the whole product before adding
facets to every service family mechanically. The 304 review preceded persisted
query comparisons, shared-window editing and declarative scope controls. Those
changes alter what users can compose and how they navigate, execute and preserve
an investigation; another one-widget pass will not expose their combined costs.

Use one realistic incident task across Activity → discovered service → useful
team view → isolated cohort → supporting trace → saved explanation/configuration
and reopen. Assess whether a responder can find the next action without knowing
implementation terms, distinguish query intent from retained observations, and
recover their original scope/focus. Architecture should examine duplicated
source/control ownership, unnecessary reads and declaration/host boundaries in
that same task, rather than demand that every typed role become one generic
snapshot.

Reconcile the roadmap from observed failure to complete or reuse the task. Select
one high-impact next outcome, remove obsolete competing entry points where that
improves the flow, and keep further family/profile expansion conditional on that
need. This is a product/architecture checkpoint with a real deliverable, not a
standing visual-polish queue or an automatic next comparison flag.


## Source and actual 311 acceptance

The shipped HTTP template now contributes the declared current-method facet
through the existing checked expansion. Template composition may retain an
ordinary literal facet field when reused with another eligible grouping; the
facet is not rebound to that grouping or treated as HTTP-profile proof. Existing
unsupported numeric/specialized shapes do not silently acquire the facet. This
keeps generic facet compatibility separate from the stricter initial HTTP service
widget boundary. No additional component or above-Activity controls were added.

The actual [HTTP scope](../journeys/screenshots/311-http-route-methods.png) shows
107 matching server-span records: GET 100, PATCH four and POST three. The method
facet precedes the route chart, whose missing/empty groups remain distinct.
The [selected scope](../journeys/screenshots/311-selected-http-methods.png) shows
GET plus POST 103, with matching /read 100 and /write three route counts. The
[destination reuse](../journeys/screenshots/311-typed-template-destination.png)
shows its own seven records (PUT five, POST two) and retained service grouping,
while the facet remains literal `attribute:http.request.method`. The UI does not
show source method counts as destination observations.

All three frames are accepted with no visual blocker. The verification owner
reports main native green in 1.7 minutes, including real discovery, retained
template/List and typed destination reuse. Native/model reporting separately
establishes explicit read effects, record retention and lifecycle completion.
The actual crops emphasize facet/chart interaction; they alone do not prove the
record assertions. No source changes were requested in this final review.

One nonblocking observation belongs in checkpoint 312: a copied authored chart
title such as HTTP server routes can remain while compatible destination grouping
is service. Titles are presentation, not checked semantic assertions. Revisit how
default titles and actual scope/grouping stay understandable during reuse as part
of the whole-product task, rather than expanding 311 with another title API.
