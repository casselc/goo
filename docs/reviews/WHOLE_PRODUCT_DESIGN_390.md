# Whole-product design checkpoint — 390

Read-only review, November 9, 2026. Milestones 388–389 are delivered. The 390
implementation and component checks are available, but final application geometry,
actor and model acceptance remain root-owned gates. This document proposes inputs
to the next finite roadmap; it does not activate a second backlog or claim those
gates have passed.

Inputs include the [capability index](../CAPABILITY_GUIDE.md),
[known limits](../KNOWN_ISSUES.md), [adopted queue](../ROADMAP_2026_11_09_CHECKPOINT_387.md),
the [387 design review](WHOLE_PRODUCT_DESIGN_387.md), current application and
declaration code, 388/389 actual frames, and the 390 component/initial-layout work.
No telemetry queries, new browser campaign or heavy tests were run for this
whole-product review. The separate 390 component test is described below only
within its actual supplied-fixture boundary.

## Product center

Keep the center **current logs/spans → isolate a population → follow or compare
evidence → preserve/reuse the question**. Live Activity already supplies the
firehose entry. Fields, typed filters, facets, time selection, record inspection,
trace ancestry/waterfalls, selected/rest comparisons, time comparisons and ranked
explanations already exist. A new home dashboard or universal investigation
canvas would dilute that center rather than fill a missing primitive.

Services and native measurements are useful adjacent entry points. They should
lead back into the same ordinary record questions, not require another grammar
or a manifest. Manifests can improve naming and authored intent without becoming
an admission requirement for ordinary OTel. Service, namespace, instance and
environment distinctions remain meaningful; reported associations are not proof
of physical topology or application health.

The product is broader than the demonstration applications. Existing automatic
service/resource observations, RED-style span measures, USE-style native resource
views, log-pattern analysis and cross-service search should remain discoverable
from actual held data. Do not recreate them under vendor feature names. Their
remaining boundaries are narrow source/profile or display contracts, not evidence
that entire categories are absent.

## What recent work changes

388 makes an unavailable or stale gauge timestamp useful: a user can stage nearby
logs/spans, explicitly read them, then return to unchanged evidence. Its repeated
exploration fix also establishes a product rule: deriving another question from
accepted held data is different from reacquiring telemetry. A disabled source
lease should not silently disable a safe supplied-data interaction. Keep that
distinction explicit in each owner; it does not justify a universal session engine.

389 closes a real native display gap with counter Timeline/Table reuse. Unit and
integer/decimal applicability are visible and checked, while the destination's
full question and evaluation policy win. It does not create another counter
evaluator. Its image review found a more general presentation rule: unavailable
calculation and absent optional output are different from unrecorded source data.
Those states must also remain distinct from numeric zero.

390 addresses the space between these capabilities and a usable product: primary
data precedes full identity, diagnostics and save/configuration controls. It
retains visible state, unit, fixed window and frozen receipt identity while
progressively disclosing detail. The public compact chart-frame profile uses an
ordinary HTML attribute and remains display-only. Its separate SSR/nested-SPA
fixture passed at desktop/narrow widths with child identity, focus and open
disclosures preserved and zero API reads. That proves component composition,
not ordinary/service/saved telemetry entry acceptance.

The initial actual narrow application run still found a clipped unavailable rail.
The current correction targets redundant chrome and the disclosure marker's own
line, rather than shrinking the data area or scrolling the test to make it pass.
Final application frames and measured outcomes must close that gate before the
planner treats 390 as delivered. Successful component models or percentage
coordinates alone do not establish human-visible correctness.

The subsequent run also exposed the Saved embedding boundary: catalog/navigation
and provenance chrome preceded the active native capture and pushed its
unavailable interval offscreen. The current correction gives accepted evidence
priority in DOM/tab order and moves optional context below its visual, while
retaining frozen receipt/custody, state and window information. This is why the
same component must be checked in actual parent surfaces, not only in isolation.

## Recommended finite cadence

### 391 — breadth: reusable log-pattern count displays

**User task:** “This pattern population is useful; show it as a comparable bar or
table view and reuse that display on another log-pattern question.”

Log-pattern inference, pattern records, local sorting, selected/rest comparison,
time comparison, capture and saved questions already ship. The concrete gap is
that the shared saved display grammar explicitly excludes
`column:body_pattern_v1` in both
[`ui/view-declaration.mjs`](../../ui/view-declaration.mjs) and
[`workbench/view_declarations.go`](../../workbench/view_declarations.go).
Closing that gap advances the log-first product and the user's shared-model goal,
instead of scheduling another native-metric adoption immediately after 390.

Start count-only, logs-only, one exact derived grouping, Bar/Table. Reuse the
existing declaration/store/source owner, not a new log-pattern view engine. A
readable pattern label is not its identity: retain canonical versioned LP1 IDs,
escape literal tokens and controls, and keep empty/oversized/invalid-body groups
visibly unclassified. Every matching log remains in the captured denominator.
Do not broaden templates, introduce another normalizer, infer events from message
text, or silently turn bounded top-K rows into a whole-population share.

**Acceptance:** one ordinary Activity → pattern analysis → local Bar/Table → save
display → different compatible log scope → Inspect/Apply → explicit Run journey;
whole destination predicates/window/environment preserved; exact IDs and counts
with unclassified coverage; no-read local switches; changed/failed reads withhold
unsupported shares/actions. Invalid UTF-8 belongs in raw Go/source tests, not
fabricated positive OTLP JSON. Go/JavaScript/OpenAPI default/presence cases should
be part of this admission's normal verification.

### 392 — depth: native outgoing HTTP duration

**User task:** “The outgoing HTTP observations are interesting; inspect the
reported client-duration histogram for this service without confusing it with
the captured Client-span population.”

385 already supplies an explicit Client-span lens. Current
[`OUTGOING_HTTP.md`](../OUTGOING_HTTP.md) and
[`SERVICE_DASHBOARDS.md`](../SERVICE_DASHBOARDS.md) explicitly defer execution-profile
adoption of `http.client.request.duration`. This is a bounded source/profile gap
that complements the existing outgoing workflow and OTel semantics.

Adopt only a supported exact instrument/unit/temporality/schema contract through
the current native planner and histogram evaluator. Generic discovery metadata
does not grant execution. Preserve native stream identity, method/destination
dimensions, unknowns, empty data, units, resets and independent read clocks. Do
not infer native histogram counts from spans, merge the two populations, call
response codes universal failures, or infer a callee service from a hostname.
Legacy metric names, exponential representations and fleet percentiles remain
separate gates, not fallback aliases.

**Acceptance:** actual supported native exports and partial/no-match cases;
raw bucket/record checks; an outgoing service lens with clearly separate native
evidence; exact staged record context, explicit Run and held return where
supported; saved histogram display compatibility unchanged. Exercise unsupported
units/source representations as refusals and retain the current source-read
budgets. No new query engine, store or dashboard framework is needed.

### 393 — consolidation: literal source-parameter round trips

The previous architecture review identified different ClickHouse Escaped-format
parameter encoders in record and measurement source paths. From a product
perspective this is worth a bounded correctness audit: a literal field value
should identify the same data when an operator moves between sources.

Prove actual round trips with quoted/backslash/control/Unicode values under the
existing typed-parameter interface before changing code. Share only the small
encoder if a real mismatch is established; preserve source limits, readonly
settings, cancellation and family refusal behavior. A documented no-change result
is an acceptable closure when parity holds. Do not turn a source-only difference
into a universal transport/validator project or an indefinite performance claim.

The 389 empty-domain schema mismatch is real evidence for a shared finite
Go/JavaScript/OpenAPI admission corpus. That work should travel with each new
display arm's acceptance, rather than displacing the log breadth step with another
native-policy-only milestone. The planner/architecture review can choose the
smallest evidenced consolidation cut if the parameter audit is already resolved.

## Reconcile older work

| Older item or temptation | Current treatment |
| --- | --- |
| Capture-family consolidation, counter display reuse and ordinary gauge nearby context | Delivered in 387–389; retain regressions rather than recreate them. |
| Native data-first density | 390 acceptance gate, not a new next milestone under another name. |
| Saved/service gauge continuation | A useful retained-owner follow-up; keep explicit, but do not let it crowd out log breadth after this native sequence. |
| Log-pattern analysis, comparisons or saved questions | Already delivered. The proposed work is checked display admission/reuse only. |
| Contextual review-task chooser and verified kept-occurrence review | Still coherent, but existing annotation/review/export already works. Do not promote this pair automatically over core investigation gaps. |
| Arbitrary custom executable kinds, WYSIWYG and multiplayer whiteboards | Wider design initiatives; continue through finite checked profiles, not another platform or scripting escape hatch. |
| Exponential histogram evaluation | Preserve the source/mathematical gate. Discovery visibility is not execution authority. |
| Broad broker/DB health, alerts, tenancy, warehouse adapters and bitemporal stores | Separate product/platform decisions, not missing prerequisites for this queue. |
| Framework rewrite or universal generation/session abstraction | Reject without concrete matching ownership semantics; 388/390 show why local presentation and acquisition lifetimes differ. |

The next whole-product checkpoint should follow the adopted finite cadence, or
arrive earlier if actual users invalidate it. Keep one current queue and retire
historical “next” claims through explicit reconciliation. The goal is a smaller,
more coherent investigation loop, not a longer feature catalog.
