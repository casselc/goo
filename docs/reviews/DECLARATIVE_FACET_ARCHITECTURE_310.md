# Declarative categorical facet architecture — 310

Status: backend/schema and frontend source review accepted; executed checks
are tracked in `VALIDATION.md`. This is the first contributed scope-control primitive over the shared
view/query model, using existing acquisition and selection owners.

## Declaration and bounded shape

Add one `facet` child role with `goo.view/1`, record-query scope/signal,
`categorical-facet` kind/version 1, one qualified string `field`, and bounded display
options. The field is top-level; display contains `basis: applied` and optional title.
The wire field location is a binding, not a query; selection values,
source counts, source basis and execution state never belong in the declaration.
The single role supplies stable identity without introducing arbitrary control IDs.
Declare current-applied-scope count semantics explicitly in contract/UI; no
self-excluding/global-universe count interpretation is permitted.

Support ordinary single-query log/span views and one facet at most. Reject
specialized numeric/histogram/reported-edge/split profiles and both captured/fixed
comparison modes initially. Records remain required. Allow records plus facet
without mandatory aggregate, preserving raw-entry usefulness. `recordOnly`/read
strategy must recognize this no-aggregate shape rather than use child count one;
Run/paging continue through the existing count-plus-rows profile. Facet visibility
alone cannot convert hidden records into a meaningful raw display. Existing
aggregate forms and the four-child budget remain intact. Existing summary, bucket
and span-status projections are allowed: `FilterValues.query` already derives a
count-only demand from exact scope without inheriting measure projections. Optional
existing timelines remain usable; numeric/histogram/edge/split modes remain refused.

## Data and action ownership

Mounting, disclosure, changing presentation and ordinary Run do not secretly load
facet values. Offer an explicit **Read values** action backed by existing
`FilterValues.query/result` and `/api/explore/counted-values`. The application
adapter owns this demand, request controller and exact applied scope; the facet
display remains passive. The current qualified predicate/type whitelist determines
string eligibility; reject unsupported numeric fields instead of stringifying.

Counts describe the current applied fixed question, not rows currently visible in
an inspection page or a universe with this facet removed. Preserve the existing
independent count/group read receipts and changed/no-data refusal. Missing values
and present empty strings remain distinct; source field identity remains qualified.
No view declaration supplies an acquisition guarantee or executable renderer.

Include/Exclude gestures stage the existing selection tray and perform no read.
Existing Apply/Run owns the resulting ordinary query predicate. Do not present
checkboxes that imply replacing a privately owned applied filter. To clear pending
facet choices, attach an app-only origin marker and remove only those pending
items, with **Clear staged facet** wording. Never erase another same-field filter,
chart selection or arbitrary applied predicate by matching its field name. Once
applied, the clause is ordinary query intent and existing explicit filter-removal
controls own broadening. Persisted clause identity and self-excluded counts are
outside this first cut.

## Lifecycle, persistence and reuse

Bind value responses to exact applied query, field and current declaration/owner.
Scope changes, replacement reads, child suspension, declaration edit/removal and
visit retirement invalidate pending work and actions. Invalid/replaced controls
cannot stage values under another scope. Existing pending record/page/source/filter
and selection guards remain; no new source executor or component-owned scope state.

Save/export contain field/presentation intent and the ordinary question only.
Open/import uses existing catalog and validation. Reuse must preserve and validate
the facet explicitly; the old aggregate/records/timeline/comparison branches cannot
silently drop it or treat it as a timeline. Category templates refuse
the new control, including hidden facets, rather than imply parameter rebinding
support. Ordinary presentation reuse preserves field/title/basis exactly and
validates destination field, profile and signal; it copies no query or loaded values. No new
registry, package loader or evaluator is introduced.

## Acceptance and source review

Use a contributed ordinary service view with a useful attribute such as HTTP route
or a diagnostic cohort. Explicitly load scoped values, include/exclude multiple
values, clear only pending facet-owned intent, apply/read, inspect retained records,
and Save/export/import the control. Demonstrate raw records+facet and existing
count breakdown+facet, no mount/read/disclosure fetch, exact typed scope and read
receipts, missing/empty distinction and stale value/control refusal. Verify shared
field selections from other controls survive facet Clear. Keep actual browser,
property/model, API/schema and documentation frames synchronized.


## Backend source review

The checked facet declaration projects the exact applied scope through the existing
counted-values validator, preserving predicate/environment while omitting root
measure projections. Strict display decoding rejects unknown/null options. The
composition requires visible data (visible records for raw+facet), preserves the
four-child cap and refuses either comparison kind beside a facet. Templates refuse
facets even when hidden. The explicit reuse branch copies declared scalar
field/display intent and validates the destination rather than falling through to
timeline handling. OpenAPI describes the same bounded role/visibility/variant
rules. No backend architecture blocker remains; execution results belong in
validation.

## Frontend source review

The browser role validator preserves visible-data/raw-record requirements and the
raw read-strategy predicate now includes records+facet. The passive component
validates field/count partition/read-receipt input and emits Include/Exclude/Clear
intents; it owns no fetch. The app mounts the existing counted-values loader only
for explicit Read values, binding q/field/owner/generation and retaining the applied
scope through staged choices. Loaded values are not included in saved/portable
representations. Find is keyed by field/read basis rather than pending selections,
so staging does not discard local search.

Pending origin participates in staging identity; the query builder deduplicates
equivalent predicate values independently of origin. Thus a facet Include cannot
overwrite another control's identical pending value, and facet Clear removes only
its own marker. The existing tray/Run retains all applied clauses and external
choices. Child staging retains the declaration but transfers no loaded values;
Back renews the current adapter and explicit action leases.

Source inspection found no remaining architecture blocker. `git diff --check`
passed. Frontend freeze/native/property/model results remain recorded by their
owners and validation; this section does not substitute source review for those
executed checks.

## Model lesson and next outcome

The normalized counted-values model is an intentional typed control input, not an
excuse to manufacture a ranked snapshot. Proposal 260 explicitly distinguishes
several controlled model roles inside one shared scope context. Counted-values has
its own independent read receipts and applied-field demand; converting it to H0
ranked data merely to reuse H1 would risk inventing a guarantee. The present
app-owned q/epoch/field binding is acceptable for the first control, but should be
factored as a common typed control binding rather than copied for every new widget.

Prefer a substantive next task: one useful OTel HTTP-service contributed view,
combining its existing count chart, records and categorical control through the
current template/inventory context. This can prove useful defaults and team reuse
while standardizing the counted-values role/admission/binding in the existing
scope owner. It does not require another renderer or source executor. Category
binding must preserve qualified field identity and checked destination scope.

Self-excluding counts require stable applied-clause identity and explicit
broadening semantics; matching field names is insufficient. Defer that expansion
until real exploration shows the need. Likewise, a second search/control should
not introduce another private scope/session loop before common typed control
binding and intents are settled. This is a model-coherence rule, not a requirement
to turn the entire application into one universal row table first.

Final value-gesture guard review accepts `selectionReason`, which delegates to the
existing predicate validator. Values that are valid supplied display data but
cannot be represented as exact query predicates remain visible with counts while
Include/Exclude are disabled with a reason. The app handler independently repeats
that check after current-scope membership validation, so synthetic events cannot
bypass the control. This preserves display/selection distinction without a new
predicate policy. No remaining blocker; executed negative and packaged-component
checks are tracked by validation.
