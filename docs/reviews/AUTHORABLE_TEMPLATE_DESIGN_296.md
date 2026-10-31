# Authorable reusable templates — design review 296

2026-10-31. Source review of the shipped HTTP descriptor, population composition
editor and saved-view catalog. This is a proposed interaction contract, not an
executed browser or accessibility review. Validation belongs in VALIDATION and
the eventual executable storyboard.

## Decision

Use the current composition as the visual authoring canvas and the current saved
exploration-view resource/catalog for persistence. A template is an explicit,
portable presentation binding over a destination question, not a second dashboard
editor or a way to transfer another service's query or captured evidence.

The first binding should be **Destination grouping**: the canonical category slot
uses the destination's already-selected ordinary single grouping field. This is
meaningfully different from existing exact-field Apply composition. It keeps the
entire target query, including service, environment policy, predicates, window,
requested measures and delivery settings. It must not silently choose a grouping,
copy source filters, request missing measures or widen the destination scope.

## Author and save

- Keep kind, measure, columns and component order editing in the existing controls.
  Users see their actual chart/records rather than starting from source JSON.
- In Preserve → View definition, add a compact **Reusable template** checkbox near
  the existing title/Save controls. The view title is the template name; do not
  introduce a duplicate name field or another prominent toolbar.
- Add one **Data profile** choice: **Ordinary records**, **OTel database client
  spans**, or **OTel HTTP server spans**. Profiles resolve through common checked
  descriptors; ordinary authors do not need a predicate/expression editor. Validate
  the actual source question against the chosen profile before saving.
- Explain the checked mode briefly: **Use the destination's current grouping;
  keep its query and scope.** Show the actual source grouping as context, not as
  a requirement that is silently substituted later.
- Saving persists the materialized source view and checked template declaration
  in the same resource. A contextual **Save template** label can identify that
  intent while regular views retain their current Save action.
- First-profile limitations should be visible refusals. Capture-comparison
  references, live controllers, unsupported families and source-specific Find
  literals must not disappear as a side effect of enabling template mode. Tell
  the user what to clear or keep it as a regular saved view.

## Discover, preview and apply

The existing Saved exploration views list should mark template entries and offer
**Inspect template**. Its existing compatibility preview is the right place to
show the concrete destination grouping, component order, kinds, measures, columns
and the chosen profile’s required selector conditions. **Open original view** remains distinct: that
restores the saved source question, while applying a template keeps the current
question. Avoid parallel Apply composition and Apply template actions on one
entry when they have different binding meanings.

**Apply template** becomes available only after checking the actual destination
query. Missing duration support, incompatible grouping/profile or absent mandatory
selectors gives a specific reason; application must not add those conditions or
change requested measures. Preview/application are configuration operations, with
no implicit telemetry read, capture or save. Explicit Run acquires data when
required, and Save preserves the resulting destination view independently.

Reuse the established draft/selection/paging/live guards and current-owner fencing.
A delayed preview cannot apply after navigation or a newer query. The target query
must be compared in full in the real journey, not only service/window labels.

## OTel proof and limits

Use a broad **Database client activity** default over current DB semantic
conventions, requiring Client kind and nonempty `db.system.name`. The common
profile descriptor/required-conjunct checker should serve HTTP and DB defaults;
do not add another hardcoded DB-only expansion mechanism. A DB operation grouping
can bind explicitly to another compatible destination grouping, with that exact
field shown in preview. Counts remain captured client-span records, not unique
queries or database-server health.

Prove ordinary user creation without source JSON, save/reopen, application to a
second compatible service, and refusal of an incompatible destination. This cut
need not add arbitrary expressions, uploaded code, free-form layout, new renderers
or a separate catalog. Those are not prerequisites for useful team reuse.

## Implemented source review

The native **Reusable template** checkbox and **Data profile** dropdown reuse the
existing title and composition canvas. Profile choices use the agreed bounded
enum. A rejected change restores the checked/profile state and presents its reason;
unsupported Find, comparison and specialized content is not silently removed.
The retained **Save exploration view** label is acceptable: the explicit template
mode saves that same materialized view resource, and the catalog identifies it as
**Template · name**. No duplicate name or top-level authoring toolbar was added.

The catalog distinguishes **Open original question** from **Inspect template**.
The checked preview names the actual destination grouping, profile requirements,
UTC window and full selector, component kinds/measures and record columns.
**Apply template** is separate from source reading and uses the existing owner;
template entries do not also offer the differently bound ordinary Apply composition.
The database Customize entry uses the shared handoff rather than another editor.

The source review caught a root-offset edge: a nonzero-offset destination could
otherwise preview/apply and then fail template Save. The application now refuses
that root explicitly before Apply; ordinary record-page inspection with a page-zero
root remains supported. The pure API's detached expansion may normalize delivery
offset independently. This avoids silently changing a held source question or
leaving the current view unexpectedly unsaveable. A stale Template unavailable
message also clears after the current configuration becomes valid again.

With those corrections, no source UX blocker remains.

Actual browser frames and interaction results remain pending; this source review
does not claim keyboard, screen-reader or mobile checks.

## Actual frame review

Inspected the generated 296 authoring, compatibility-preview, applied and reopened
frames. The native checkbox/profile controls are compact within Preserve, reuse
its existing title, and clearly name destination grouping. The preview explicitly
shows `attribute:db.namespace`, the destination service/namespace predicate and
UTC window, DB profile requirements, table/p95 choice and record columns before
Apply. This makes the new binding meaning visible without reading source JSON.

The applied and reopened frames show lowercase `ledger`/`archive` destination
groups, millisecond durations and UTC record timestamps; no source-group values
appear to have been relabeled as destination data. The surrounding cards and
controls retain existing styling. No visual blocker to the template contract was
found. Source-free application/reopening, negative-profile refusal and stale
ownership are interaction assertions owned by the executed journey, not facts
proved by these images alone.

Two frame/readability limits are not template-binding defects: a sticky trail
crosses the applied screenshot, and the existing exact table presents all duration
columns, leaving the selected p95 column off the initial horizontal viewport.
The former can be cleaned as capture framing; the latter belongs to a deliberate
shared-table usability pass rather than a new template-specific renderer.
