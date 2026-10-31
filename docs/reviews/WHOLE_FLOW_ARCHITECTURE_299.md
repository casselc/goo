# Whole-flow architecture checkpoint — 299

Source review of checkpoint `0550718`. This is a finite continuation decision,
not a claim that new template support or a browser walkthrough has shipped.
Existing executed behavior is recorded in [VALIDATION.md](../../VALIDATION.md).

## Decision

Make an existing timeline composition reusable as an authored template. This
closes a natural investigation task: find when a scoped population changed, keep
that presentation, and apply it to another compatible service without losing its
time view. It reuses shipped renderers, declarations, source aspects and catalog.
It does not introduce another query owner, service preset, store or editor.

Architecture and design arrived at this recommendation independently. A new
waterfall kind is technically possible, but it should not displace this concrete
break in the current save/reuse workflow.

## Reconciled flow

| User step | Existing owner and behavior | Current architectural limit |
| --- | --- | --- |
| See Activity now | `stream-entry.mjs`, `tail.mjs`; explicit live stream and field scoping | Live windows do not become fixed captured evidence automatically |
| Open a fixed population | `population-workspace.mjs`; one checked query, explicit Run, retained parent | Fixed signal/environment/time and predicates must remain explicit |
| Read Table or List | `record-results.mjs`, `record-interactions.mjs`; one page-position selection/Keep/Inspect owner | Page occurrences are neither stable event IDs nor totals |
| Inspect trace or log context | Existing record context owner performs explicit fresh reads | Cross-service trace expansion is not the parent records page |
| Compare populations | Canonical captured-count pair and group-side staging | Strict count profile; no scalar-percentile or incompatible-profile conversion |
| Customize/reuse HTTP or database service views | Shipped descriptors plus `goo.template/1`, ordinary saved-view catalog | Already delivered in 296; do not reimplement the same two-service journey as new product work |
| Save a timeline composition | Existing timeline declaration, parent `bucketMs`, save/reopen in 281/287 | It works as a fixed saved view |
| Make that composition reusable | Existing authored template validator/expander | Both Go and JS refuse every timeline child, forcing users to remove a useful existing view |

298's real journeys already cover shared actions and Back, and 296 covers a
Database client activity template applied to another service/grouping. The next
slice should connect those delivered capabilities with timing rather than repeat
the same preset/template adoption under a new milestone number.

## Proposed contract

Retain one required aggregate and one required records child, with one optional
existing timeline. Preserve declaration order, visibility, titles, selected
measure, complete-only setting, records kind and qualified fields. Continue
refusing captured baseline references, numeric/native/specialized query families,
tails, grouped time series and nonempty aggregate Find. Those need their own
explicit portability contracts; permitting a timeline does not permit all children.

A timeline consumes the destination query's already configured interval aspect.
Use the existing ordinary-composition reuse rule: supported positive source and
destination `bucketMs` must match, including when the timeline is hidden. The
source query is still a validation specimen, never a destination default. Missing
or mismatched destination width should explain that the user must configure the
interval explicitly with existing controls. Applying a template does not add or
change `bucketMs`, bounds, environment policy, predicates, grouping, summary,
paging root or source data. A subsequent Run remains explicit.

Count, mean duration and p95 duration are already timeline declarations. Permit
only what the destination query already supports. Duration presentation requires
its existing span duration summary. Preserve exact interval/support/no-data/zero
and completeness semantics through the existing timeline validator and renderer;
never merge stored percentiles or infer upstream completeness.

## Implementation boundaries

- `workbench/view_templates.go`: allow the bounded optional timeline in
  `ViewTemplateBinding.Validate`; keep comparison refusal. In authored expansion,
  compare source/destination timeline interval requirements before dropping the
  specimen query. Retain Go compilation, content-ID and detached-copy checks.
- `ui/view-template-handoff.mjs`: mirror the permitted role profile in
  `templateView`; check temporal compatibility in `authoredRequest` before its
  destination clone replaces the specimen query. Reuse
  `timeline-declaration.mjs`/`Timeline.reuse` rather than inventing another width
  validator. `authoredResponse` keeps exact question and presentation checks.
- `ui/population-workspace.mjs`: reuse current native interval controls, template
  authoring and metadata request fencing. Template preview should identify the
  timeline measure and interval width, rather than showing only the word
  “timeline”. Do not add another per-template query editor.
- `workbench/view_timeline.go`, `ui/timeline-declaration.mjs`,
  `ui/timeline-view.mjs` and composition validators already provide the typed
  aspect and presentation. Do not migrate them wholesale or add a new kind.
- Existing API arms and saved declaration fields suffice. Regenerate documentation
  examples/checks if needed; there is no new source endpoint or wire-field need.

The important cross-query check is easy to miss: validating the source and then
validating the same composition against a destination independently does not
prove equal widths. The timeline child intentionally has no second bucket-width
field, so compare the two parent queries before replacing either.

## Acceptance

Start with real OTel spans in an existing service default or ordinary fixed
population, configure a count/duration timeline through native controls, select
its presentation and enable Reusable template without removing it. Save, inspect
against another service with a different ordinary grouping but an explicitly
matching interval width, Apply locally, Run, save and reopen. Verify the target
question remains its own and exact declaration order/visibility/measure survive.

Use missing width, mismatched width, absent duration support, hidden incompatible
timeline and captured-comparison child as negative controls. Refusals and template
inspection/application must read no telemetry or silently mutate the query. An
explicit width configuration stages a query and requires Run as it does today.
Paging must retain the original timeline/aggregate bases, and pending record
selections retain 298's refusal behavior. Extend semantic/property checks and an
observed model/browser lifecycle around this actual outcome; do not duplicate the
validator as the test oracle.

## Waterfall boundary retained for later

A supplied-span waterfall can be another presentation of `goo.record-page/1`,
provided multiple trace IDs are grouped and missing/duplicate/ambiguous parents,
exact UTC/duration values and unknown upstream coverage stay visible. It must not
advertise a complete trace from a bounded service-filtered page.

The trace inspector expands source scope across services/filters using independent
reads. Authoring that acquisition would require an explicit context contract and
existing context owner, not a records-kind toggle that secretly changes queries.
This remains a distinct future product decision after the reusable timeline cut.

## Implementation review

Reviewed the 299 changes to `workbench/view_templates.go`,
`workbench/timeline_view_templates_test.go`, the existing HTTP expansion test,
`ui/view-template-handoff.mjs` and the population authoring/preview integration.
No remaining architecture blocker was found in this source review. Actual
browser/model/property outcomes belong in the validation record.

The implementation keeps the bounded two-or-three-child template profile and
explicitly refuses captured comparisons. Existing composition validation still
requires exactly one aggregate and records child, admits at most one timeline,
and validates every child even when hidden. The Go expansion validates source
content and destination compilation, then compares the two parent bucket widths
before constructing the target view. No specimen scope is copied into the target.
The existing content digest and detached JSON-copy path remain intact.

JS checks `Timeline.reuse` before replacing the specimen query and repeats request
admission during response validation. This prevents a forged independently valid
response at a different width from bypassing the source/destination contract.
Source/target duration support, supported widths, 240-interval budget, optional
complete-only presentation and hidden-role semantics remain in existing typed
validators. The response keeps exact destination question and composition checks.

Preview now names the timeline measure, display kind, hidden state, required width
and current width. Its correction control reveals/focuses existing destination
interval configuration; it does not install the source width or issue a query.
The current metadata request/owner fencing and explicit Apply/Run path are reused.

The added Go tests cover all existing timeline measures and both kinds, visible and
hidden order-preserving declarations, List records, saved/reloaded source identity,
different target environment/service/grouping/bounds, page-zero delivery reset and
detached input preservation. Refusal cases cover missing/mismatched/unsupported
width, interval budget, missing duration summary, split intervals and invalid
source declarations. The existing HTTP arm test now exercises a hidden complete-
only p95 timeline with matching and mismatched widths while retaining its
source-read refusal setup. No new API fields, versions or source endpoints are
required for this implementation.

## Recommended next outcome — 300

After validating the timeline slice, support reusable captured-count comparison
presentation with an **explicit destination baseline binding**. This continues the
same investigation/save/reuse task and fills the remaining evidence-reference gap;
it is more useful than adding another renderer by default. This is a proposed
next contract, not implemented behavior in 299.

Do not implement it by merely allowing the comparison child in a template. A
comparison contains an exact immutable capture reference, unlike the timeline's
query-supported aspect. Copying the specimen's `baselineCapture` into another
service would transfer the wrong evidence while appearing to be a presentation
change.

The bounded contract should be:

1. An explicit template binding marks the comparison baseline as supplied by the
   destination. The source saved view can retain its existing concrete reference
   as a validation specimen, but it is never a default for expansion.
2. The destination chooses an exact baseline ID through the existing picker or
   its already configured comparison. Pass that target ID as explicit input to
   configuration-only expansion. Missing input refuses with a recovery path to
   the current picker; never fall back to the source ID or “latest”.
3. Expansion validates target count-profile eligibility and ID syntax, and emits
   an ordinary comparison declaration with exactly the supplied target ID. It
   preserves source comparison display/order/title/visibility and all other
   configured presentation, while preserving the complete target query.
4. Pure expansion does not load a catalog, capture, stored rows or telemetry.
   A syntactically admitted reference is **unverified**, not proof of baseline
   compatibility. Applying the declaration remains local. Existing explicit
   Prepare loads verified stored evidence, checks its root against the target,
   and runs the canonical comparison. Run retains its existing documented effects.
5. The target grouping may differ from the specimen grouping through the existing
   destination-category binding. The chosen actual baseline must match the
   **target** grouping, checked environment policy, service/filters, bucket width
   and requested measures; only fixed windows and delivery pages may differ.
   Do not call ordinary source/destination captured-comparison reuse compatibility
   to enforce specimen-root equality, because that would defeat template regrouping.
6. Keep the current strict count-only profile. Status/numeric/native/reported-edge
   queries stay unsupported; unequal windows remain raw count differences with
   visible durations, not automatic rates. Timeline plus comparison may coexist
   only when each existing aspect/width/profile requirement is met.
7. Reuse the current picker, prepared-pair owner and retirement rules. Choosing a
   new target reference invalidates old preview/prepared results and requires a
   fresh compatible expansion. Do not add a second baseline catalog or fetch from
   component lifecycle. Pending records, drafts and reads retain existing guards.

The response validator must permit exactly the intentional baseline-reference
substitution while retaining exact checks on the remaining source presentation;
do not weaken whole-composition equality generally. The new baseline parameter
belongs in the typed expansion contract, not a string replacement over JSON or an
implicit mutable application global. Update OpenAPI/Go/JS request admission and
wire-arm tests together when choosing the exact prerelease syntax.

Acceptance should reuse a real comparison template from service A against service
B with a separately chosen B baseline and a different grouping. Verify no A
capture is read or retained in B's output; Apply reads nothing; explicit Prepare
verifies B's actual capture; saved/reopened B reference stays exact. Negative
controls need missing target ID, malformed ID, unavailable capture, A-shaped
capture with valid ID, wrong target grouping/filter/environment/width, forbidden
status profile, late old-baseline response and a second expansion with no new
baseline input. A chosen source-equal ID may be allowed only as an explicit user
choice and still passes the same target verification; equality must not serve as
implicit authorization.
