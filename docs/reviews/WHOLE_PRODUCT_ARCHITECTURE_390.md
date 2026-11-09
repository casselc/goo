# Whole-product architecture checkpoint — 390

Read-only source/document review on November 9, 2026, after 390 source completion
and while its actual runtime acceptance is pending. This is input to one revised
roadmap, not an independent execution queue. No builds, benchmarks, browser runs,
production edits or data changes were performed by this reviewer.

## Product and layers to keep

The product loop remains **current telemetry → isolate a population → follow
evidence → preserve or reuse intent**. Ordinary OTel is sufficient; manifests can
add declared semantics without becoming an ingestion prerequisite. Existing
Activity, traces, comparisons, log patterns, native measurements, service/resource
observations and annotations are real capabilities, not categories to recreate
from a vendor checklist.

The current boundaries are appropriate:

| Layer | Keep its responsibility |
| --- | --- |
| `identity`, annotations and instrumentation | Qualified identity/presence, declared code contracts and caller-owned OTel providers/shutdown |
| `metrics`, `measurements`, `subjectfacts` | Native arithmetic, typed measurement planning and prepared subject facts; explicit lifecycle observation rather than hidden numeric-evaluation storage effects |
| `exploration` | Typed record/source questions, bounded population reads, trace search, log-pattern identity and explicit discovered service profiles |
| `scopedview`, supplied-data adapters and display modules | Checked immutable datasets, units/support/quality, passive chart/table/selection presentation |
| Application UI owners | Requests, drafts, cancellation, current capability, navigation and retained return |
| `workbench` and family stores | Immutable configuration/evidence publication, verified exact-ID loading and bounded metadata discovery |
| Internal HTTP/source packages | Admission/effect/deadline mechanics and bounded transport; family policy remains with the caller |

No new query engine, universal validator, renderer registry, navigation manager or
global generation-counter abstraction is justified by this checkpoint. The future
warehouse/compiler direction does not require replacing the present evaluator or
ClickHouse source planner now.

## Lessons from 387–390

387 separated the finite capture envelope from its six evidence validators and
receipt-only references. Failed request progress must not replace the identity of
accepted evidence still painted. This is a capability boundary, not a schema
versioning exercise.

388 showed why supplied-evidence staging and source acquisition cannot share one
retirement flag. An unchanged held gauge can derive another nearby question after
Back without reacquiring telemetry; source/save capabilities retain their own
checks. The projection is a clipped time association with retained sample
membership, not a metric-to-record causal join.

389 extended one NativeView owner to counter intervals, preserving destination
policy and exact unit/domain compatibility. Catalog domain metadata cannot grant
loaded definition or execution authority. Two concrete lessons matter for later
families: Go's omitted/empty integer default must agree with JSON/OpenAPI and JS
admission, and unavailable derived values may still have recorded source exports.
Neither a generic null label nor a mathematically valid plot proves honest paint.

390's source changes keep consolidation at the correct layer: owned native
layout, component-owned compact chart chrome, progressive identity/configuration,
and shared return styling. They do not merge the distinct NativeRecordNavigation,
ResourceReturn and Saved/service retained owners. Natural first-data geometry,
failure/unknown states and no-read return still require actual acceptance before
the roadmap is adopted; source inspection cannot certify those outcomes.

## Library, reliability and performance review

The older library proposal is substantially delivered: native ESM display and
application graphs, family HTTP extraction, caller-owned instrumentation runtime,
pure numeric evaluation and explicit lifecycle observation already exist. Do not
repeat those cuts under new names. Broader metadata/hash simplification remains
consumer-led work, not a prerequisite to ordinary calculations or new views.

Go already targets 1.27 with a validated 1.27.3 toolchain. Typed request structs,
`maps.Clone`, `slices` helpers, `reflect.TypeFor`, `t.Context` and current HTTP/context
facilities are appropriate where they simplify a touched path. There is no missing
runtime feature blocking the next user outcome. Internal helpers may use modern
Go without making callers manipulate a generic algebra. No blanket API renaming,
unchecked-hash rewrite or compiler-upgrade milestone is recommended.

Source HTTP is already bounded: exploration and measurements provide 15-second
default clients, close response bodies, and share `sourcehttp`/`sourcejson` without
erasing family policies. Exploration admits bounded record pages and skips blank
framing lines; measurements reject partial measurements and blank framing, with
different row limits. Do not merge those policies to reduce a few branches.
Browser cancellation through response consumption and current-owner rejection
must remain separate from publication/ambiguous-commit and streaming lifetimes.

The remaining literal-parameter encoders differ in source: exploration escapes
all ASCII controls and DEL, while measurements replaces backslash, newline, tab,
carriage return and NUL. This establishes duplication, not a reproduced transport
defect or latency bottleneck. A bounded actual round-trip audit can determine
whether one pure encoder is useful. It must not become another source client,
SQL compiler, URL-policy abstraction or generic request router.

No benchmark was run and no source-read, scan, allocation or latency improvement
is claimed. Source statement counts are not physical scans; browser campaign
duration includes the actor and is not product response latency. Profile a real
slow or allocation-heavy path before introducing pooling, parallel source reads,
extra indexing or buffering changes. Native density improvements are usability
claims that need geometry, not backend performance claims.

## Recommended next finite cadence

Subject to final 390 acceptance and product/planner reconciliation:

1. **Breadth — checked log-pattern display reuse.** Existing log-pattern analysis
   already has full populations, fixed timelines, comparisons, capture and exact
   `lp1` identity. Both `workbench/view_declarations.go` and
   `ui/view-declaration.mjs` explicitly exclude `column:body_pattern_v1` from saved
   Bar/Table display admission. Close that finite portability gap for count-only
   logs grouped by this exact versioned field. Do not merely remove the ban:
   readable decoded labels must retain canonical IDs as selection/scope keys;
   empty, oversize and invalid-body buckets keep full denominators without a
   false common-message-shape claim. Reuse the existing store, question and
   supplied-data owner. New normalization, grouped-time profiles and general
   pattern templates stay separate. Check valid/refused Go/JS/API boundaries,
   immutable configuration, explicit Run and source-free local presentation.
2. **Depth — discovered native HTTP-client duration.** The 385 Client-span lens is
   delivered; its test explicitly prevents a generic
   `http.client.request.duration` instrument from acquiring profile authority.
   Current instrument-capability and default-dashboard code has an HTTP-server
   duration profile, so this is a real finite depth opportunity. Pin the current
   semantic-convention shape/schema and exact standard unit rather than changing
   the server metric name in a copy. Generic histogram evaluation already exists;
   add justified eligibility/defaults and reuse it. Conditional absent dimensions
   must not invent response/error rates, attempts, callee health, complete trace
   relationships or Client-record-lens authority. Metric-only observations remain
   distinct from positive span evidence. Validate raw buckets/epochs, exact
   identity, compatibility refusals and the actual discovery→Inspect→Run path.
3. **Consolidation — prove literal parameter transport before sharing it.** Use
   actual ClickHouse `SELECT {value:String}` round trips across canonical Unicode,
   backslash and ASCII controls for both existing source adapters. Preserve exact
   original values and typed parameters, including empty values. Share only a
   pure Escaped-format encoder if the proof supports one; retain deadlines,
   row/byte budgets, readonly settings and error mapping in their families. If
   no useful difference is established, close the investigation without a forced
   refactor or claimed speedup. This is a bounded reliability cut, not a platform.

The defaults/null/presence drift found in 389 should add focused protocol vectors
to each touched feature. A universal schema engine or a campaign of paperwork is
not required before another capability can ship.

## Reconciled deferrals and process

Saved/service gauge continuation remains a valid follow-up requiring retained
sample admission; 388 ordinary support and 390 shared rendering do not confer it
automatically. Resource/stock contexts and counter reuse remain governed by their
own authority and unit/domain limits. Broader native histograms, including
exponential execution, retain their separate representation/source gate.

Contextual review-task discovery and verified kept-occurrence review are still
coherent deferred annotation work; captured-record labels/review/export already
exist. Do not silently remove that pair or let it become a prerequisite to the
core observability sequence. Broad contributed executable kinds/WYSIWYG,
authenticated multiplayer/tenancy, alert delivery, new bitemporal storage engines
and warehouse lowering remain wider design efforts, not parallel implementation
queues. Earlier runtime/ESM/HTTP-layer plans and completed 388/389 work are history,
not unfinished tasks to repeat.

Keep heavy builds/browser/generated runs serial on the constrained development
host, while source/design/planning reviews can run in parallel. Preserve retained
stores and lab data during cleanup; disposable diagnostic and build caches have
different lifetimes. Retain only useful current acceptance artifacts and concise
validation summaries. Raw-source checks, finite models, visible actor completion,
schema conformance and image inspection establish different facts; do not expand
one into a claim of exhaustive schedules or complete ingestion.

Use one adopted roadmap and a current capability index. Historical reviews can
retain their context, but new acceptance paragraphs need not be copied verbatim
into every status document. Reassess after the next finite cadence or when actual
usage invalidates its priorities.

Inputs: [390 source review](NATIVE_CONSOLIDATION_REVIEW_390.md),
[387 architecture review](WHOLE_PRODUCT_ARCHITECTURE_387.md),
[capability index](../CAPABILITY_GUIDE.md),
[library refactor proposal](../LIBRARY_REFACTOR_PLAN_2026_10_28.md),
[log patterns](../LOG_PATTERNS.md), [outgoing HTTP](../OUTGOING_HTTP.md),
and [validation](../../VALIDATION.md).
