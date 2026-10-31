# Authorable view-template architecture — 296

2026-10-31. Source/design proposal against the 295 checkpoint. This is not an
implemented API or executed validation record. Read with the
[design review](AUTHORABLE_TEMPLATE_DESIGN_296.md).

## Decision

Reuse `ExplorationView`, its immutable save/catalog, and its existing composition
as the authoring canvas. Add one optional declaration:

```json
{"template":{"version":"goo.template/1","group":"destination","profile":"otel.db-client"}}
```

The view title is the template name; composition remains the sole owner of kinds,
child order, visibility and display options. Ordinary saved views retain exact-field
Apply composition behavior. A reusable template explicitly opts into binding its
canonical category to the destination's existing qualified one-field grouping.
It does not change that grouping, add a predicate or compute a new window.

First profiles: `ordinary`, `otel.db-client`, `otel.http-server`. Ordinary templates
retain the source signal. Database/HTTP profiles require spans and exact mandatory
profile constraints. The same profile descriptor/validator must check both the
source at save time and the destination at expansion time. Profile labels never
establish semantic-convention conformance or rewrite a query to make it eligible.

## Bounded template shape

Accept one ordinary count/duration composition with existing aggregate/record kinds.
Reject numeric/histogram/reported-edge/grouped-time queries, multi-query views,
tails, capture-comparison children and source-category-specific nonempty Find.
Do not silently discard them when enabling template mode. Fixed qualified record
field columns can remain explicit; missing observed values are not proof a field
exists upstream. Preserve complete-only display requirements as requirements,
never as transferable runtime witnesses.

A timeline can remain supported only under its existing explicit bucket-width
compatibility; omitting it from this first cut is preferable to implicit rebucketing.
No arbitrary code, expression evaluator, recursive layout or new kind registry is
needed for this outcome.

## One pure expansion boundary

Extend the existing `POST /api/view-templates/expand` with an authored arm, proposed
as `{source: ExplorationView, query: Query}`, exclusive with the shipped
`{template, version, widget}` arm. Expansion is pure configuration work:

1. Validate the complete source view and its explicit template declaration.
2. If a supplied source ID is echoed as provenance, verify its content digest with
   ID omitted, matching existing saved-view integrity. Do not trust a client ID.
3. Compile and validate the complete destination query. Check signal, ordinary
   one-field grouping, requested measure support and mandatory profile conditions.
4. Validate the detached composition against the destination. Count/duration-only
   validation permits the opted-in category binding without constructing a fake
   source query or weakening ordinary `ReuseViewComposition` compatibility.
5. Return the existing expansion envelope containing detached target configuration.
   Copy the destination query exactly, resetting only delivery offset to zero.

The source service, predicates, environment policy, namespace and fixed window do
not become target defaults. No records, source-read clocks, custody, completeness
witnesses or action leases enter the result. Preview, Apply and Save do not read
telemetry; Run acquires independent target data. Saving the applied result uses the
same immutable view catalog and does not mutate the source revision.

## Shared shipped profiles

Export/reuse the current HTTP descriptor's closed profile checking and All-only
mandatory-conjunct logic. HTTP retains exact nonempty service, Server kind, current
nonempty `http.request.method`, and its shipped route-group/widget restrictions.
An occurrence under a weakening OR must still refuse. Do not replace the HTTP path
with a looser generic span label.

Add a database preset through the same machinery: `otel.db-client-operations`,
broad title **Database client activity**, `attribute:db.operation.name` grouping,
exact Client kind and nonempty current `db.system.name`; service handoff also
requires the existing exact nonempty service selector. Database client records do
not establish server/pool health or unique operations. Authored expansion can bind
a different existing ordinary destination grouping only through the explicit
template declaration and checked profile.

## Acceptance and ownership

Native author controls belong beside Save and the existing composition editor.
Saved-library Inspect template previews concrete target grouping, profile,
measures, kinds/order, record fields and unchanged scope before explicit Apply.
A stale preview or expansion cannot apply after root/query/draft/owner changes;
reuse the existing library/application ownership and retained-navigation rules.
No automatically loaded source query and no second catalog are warranted.

Meaningful checks are pure Go validation/aliasing tests, Hegel properties over
actual expansion preserving randomized target policies/Boolean filters/qualified
fields/windows, rejected profile constraints under OR, and actual UI save→inspect→
apply→Run→save/reopen with a database service distinct from the source. Observed
Quint/Bombadil checks should assert real completed publication/read outcomes,
including pending/stale refusals, rather than mirror the implementation in a second
planner. Regenerate OpenAPI and affected storyboards/diagrams from actual results.

## Implemented backend review

Reviewed the actual Go request, saved-view, expansion and descriptor changes.
No backend architecture blockers found. `ViewTemplateBinding` uses the agreed
version/group/profile contract on the existing immutable view resource. The first
implementation deliberately admits aggregate+records only; timelines are refused.

The two JSON request arms are exclusive even when mixed fields are empty/null.
Source validation and optional source content-ID verification precede expansion.
The original destination compiles before offset is reset, then the same binding and
composition validators check its profile, signal, field grain and measure support.
JSON detachment prevents aliasing through predicates, record fields, composition or
binding. Neither arm performs source/storage work inside the pure function.

HTTP and DB share required-conjunct descriptor checking. Existing shipped HTTP
queries with numeric bindings retain their complete materialized query, while the
unsupported reusable-template binding is omitted; they are not falsely advertised
as portable authored templates. OpenAPI's exclusive arms and object constraints
match the decoder's non-null wire contract. Focused tests and API validation are
reported separately by their executing owners; this review did not rerun them.

Frontend preview/application ownership and live lifecycle validation remain a
separate review gate before completing this milestone.

## Implemented frontend ownership review

Reviewed the native author controls, shared handoff validator, DB service action
and saved-library template preview/application path. No frontend architecture
blockers remain in the inspected source. Author controls capture their mount,
query reference and generation, reject disabled/unchanged events and pending
work, and remount after retained Back. Invalid profile/nonportable choices restore
the existing binding rather than dropping query or presentation configuration.

Inspect uses the existing library/root/draft/serial context before and after the
pure expansion request. Responses must retain the selected source identity,
profile, composition and exact target query (with the declared delivery reset).
Apply repeats current-context and compatibility checks before replacing only the
current composition/binding; held target data and inspection paging remain local.
No source inputs or execution guarantees arrive with the template.

The application refuses a nonzero configured root offset before Apply rather than
advertising a compatible template that cannot be saved. This does not prohibit
normal record inspection paging, whose configured root remains page zero. The
pure expansion API still validates a paged input and stages offset zero as designed.
Successful author revalidation clears an obsolete unavailable message after an
invalid local Find is corrected. Shipped DB Customize uses the existing service
owner and response/query handoff checks. Actual lifecycle, delayed-response,
publication and no-read outcomes remain the browser/model verifier's responsibility.

The configuration handoff now compares validated questions using the existing
`QuestionSession.identity` codec identity. Go omits empty predicate values and
false optional directives; raw JSON equality incorrectly refused those unchanged
questions. This fix preserves predicate operators/Boolean structure, qualified
fields, grouping, policy and fixed scope. Both questions validate before matching.
Composition, template binding and source content identity retain their strict
checks; dataset/witness/action identity is not relaxed. Focused codec-equivalence
and changed-scope negative controls are reported by the implementation owner.
