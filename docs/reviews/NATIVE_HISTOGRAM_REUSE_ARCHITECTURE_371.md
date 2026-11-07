# Native histogram view reuse — architecture 371

Status: accepted and implemented as a separate bounded arm. Final Go/UI/Hegel,
actual native journey, positive actor and observed-model checks pass; see
[VALIDATION](../../VALIDATION.md). Ordinary record-template guards remain intact.

## Existing reusable substrate

`measurements.HistogramView` and `ui/native-histogram-view.mjs` already project one
actual retained evaluated export interval into `scopedview.Dataset` shape
histogram. They validate the complete measurement question, exact stream and
interval, result/source basis, read clock, unit, temporality and all explicit bins.
`view-schema.mjs` checks native histogram fields, counts, closure and unit identity.
`scope-view.mjs` renders instrument observation distributions and tables.

`Measurements.renderNativeHistogram` already exposes Histogram/Table display
choices on that actual interval snapshot. This is the smallest existing display
reuse seam. Timeline-versus-heatmap preferences are another useful but distinct
contract: they select representations over multiple intervals/layouts, with
latest-32/200 limits, rather than one distribution. Do not combine both cuts.

Record `ViewDeclaration`, `ExplorationView`, compositions and template bindings
are explicitly checked against `exploration.Query`. Native `measurements.Question`
is a different authority. Relaxing ordinary record validation would erase unit,
stream and temporal semantics. Those guards must remain unchanged.

## Proposed minimal closed arm

A separate workbench `NativeHistogramView` stores only configuration:

```
{version: 1, id?: contentID, title: boundedText,
 question: complete measurements.Question,
 declaration: {version: "goo.view/1",
   scope: {kind: "measurement-question", signal: "metrics"},
   selection: {kind: "export-interval"},
   kind: {name: "histogram", version: 1},
   display: {category: "bucket", value: "observation-count",
             requiresComplete: true}}}
```

The second supported display is the existing table kind, not an arbitrary Plotly
escape hatch. No donor result, observed temporality, stream/interval index, selected
export bounds, bin layout, count, basis, read clock or completeness witness belongs
in this declaration. Full query source bindings, including an explicitly chosen
exact stream, remain in the question because they are intentional query scope.

Use ordinary typed methods such as Validate, Reuse(declaration,source,destination)
and existing immutable-file publication/load patterns. No registry/list platform,
executor, dataset transformation language or evaluator is required. The production
API/UI surface should follow one selected workflow, not add a second editor.

## Storage identity and bounded implementation extent

The existing seam is `workbench.Store.publish/read`, not the record view envelope.
Use a distinct content-addressed `native-histogram-views` family with
CreateNativeHistogramView/LoadNativeHistogramView and a pure
ReuseNativeHistogramView(source,destinationQuestion). Reuse need not publish.
The HTTP surface is POST `/native-histogram-views` (configuration without ID),
GET `/native-histogram-views/{id}` and POST `/native-histogram-views/reuse`
`{id,question}`. Reuse reloads and verifies the immutable source by ID, then
returns an unsaved configuration with the complete destination question. All are
metadata only. There is no enumeration index, namespace registry, mutable latest pointer
or new persistence backend. Existing publication bounds/integrity checks apply.

The configuration artifact ID includes its stored question and declaration;
measurement Definition/Question receives **no presentation fields**. Toggling the
view can change a saved view ID but cannot change the measurement compiler plan,
definition revision, acquisition question or evidence basis. Runtime choice stays
outside the measurement object and is checked against the actual interval dataset.

A finite UI flow can add Save current display and exact-ID Load/Apply controls to
the existing selected-distribution/workbench owner. Apply uses the current checked
destination preview question and preserves it; no new route, catalog or editor is
required. Its owner must fence late metadata responses and retire old controls.
Exact-ID reuse is intentionally narrower than a discoverable library; do not
silently expand this cut into a catalog. If that UX is not acceptable, defer the
cut instead of smuggling presentation into measurement identity.

## Reuse and observed authority

Both source and destination questions must pass existing native histogram question
validation/compilation. Initially retain current per-stream service histogram
support without resource or histogramBucketMs bindings and exact `s` or `ms` unit; require equal units without conversion.
Unsupported native kinds, resource/grain variants or lifecycle arms visibly refuse.
Different metric/service/namespace/policy/window or exact-stream bindings may be
reused when the display shape remains compatible: the entire destination question
is cloned unchanged. Never copy donor scope or telemetry into destination evidence.

Stage validates configuration but reads no telemetry. Preserve fixed destination
fromMs/toMs and optional fromNS/toNS exactly; do not resolve another lookback while
applying display intent. Explicit Run acquires destination data. The user then
selects an actual destination retained export interval, and only the existing
HistogramView projector can supply the displayed dataset. No first/latest stream
or interval is selected implicitly by template reuse.

Temporality is generally observed after Run, not guaranteed by the measurement
question. This one-distribution display can support both evaluated delta and
cumulative intervals, labeled with the actual observed temporality. It does not
mix or compare them and cannot claim producer conformance from metadata. A future
comparison operator would need stronger same-definition/layout/temporality rules.

## Exact interval, uncertainty and limits

Use only `measurement-interval`, not `measurement-time-bucket` or whole-window
selectors, for this initial arm. Existing aggregated-bucket/window projections
remain separate. Complete-bin proof means all bins of this evaluated interval are
present; it is not completeness of source ingestion, interval coverage or fleet
membership. Zero evaluated observations is no data, unknown/reset/gap intervals
withhold the distribution, and overflow/malformed input refuses. Original units,
left-open/right-closed bounds, unbounded tail and exact observation counts remain.
No interpolation, rebinning, rate conversion, fleet merge or causal record claim.

Acceptance should execute save/load/reuse with donor/destination differences,
whole-query equality and caller detachment; explicit Stage zero-read/Run; actual
interval selection; unknown/no-data/malformed refusal; and saved display reuse
without carrying donor stream indexes/layout/counts. Reuse independent Go/JS
histogram projection properties and one genuine native workflow. Add ownership
checks for retired display actions; no new lifecycle framework is justified.

## Decision

Root adopted this separate bounded arm and explicit-selection user flow. The
backend uses existing Store publication/loading, strict bounded JSON/Unicode
admission and content verification; pure reuse clones nested maps/slices/pointers
without donor observations. Focused tests and genuine frontend acceptance remain
pending, so this source preparation is not delivery proof. The already
working local display toggle alone is not reusable publication, and neither a
measurement package nor a captured result should be mislabeled as a saved native
view. No record-template alias or broad whole-UI migration is needed to close it.

## Backend candidate validation

The adopted implementation provides `NativeHistogramViewInput`, immutable
`NativeHistogramView`, strict declaration structs and an ID-based reuse request.
`ReuseNativeHistogramView(source,destination)` manually detaches nested source
maps/slices/pointers, preserving complete destination configuration without JSON
number repair or donor observation transfer. Create/load verify content identity;
the metadata HTTP layer reloads the saved source instead of trusting a supplied
source envelope. All three routes reject query parameters and use no-store.

Focused Go tests passed (workbench 0.014 seconds, HTTP 0.016 seconds; the HTTP
fixture rerun 0.014 seconds). Four actual success wire cases and 13 refusal cases
were emitted. Checks include distinct view IDs for Histogram/Table with unchanged
measurement compilation, whole typed destination equality, mutation detachment,
corrupt immutable winners, exact-unit refusal, invalid profile/grain/resource/
bucket arms, strict Unicode/unknown/null-bound input and zero telemetry access.
Frontend, full integration and genuine runtime acceptance remain pending.

## Final cross-layer source review

The separate declaration/store/HTTP arm, pure frontend guards and Measurements
integration are source-accepted subject to the final focused canonical-bounds
regression. Apply preserves the current preview question, clears old results and
waits for Run. Applied views suppress automatic latest-interval selection; only a
new explicit actual interval supplies bins. Runtime display remains separate from
measurement revision/basis, and complete-only means complete evaluated bins.

Review found and closed two wire-presence gaps: `source.resource:null` must not
enter the service-only arm, and present empty fromNS/toNS must not alias omitted
native bounds. Shared native question wire checks now reject both on create,
loaded artifacts and reuse; focused create/reuse regressions cover them. These
are prerelease strict-admission corrections, not compatibility aliases.

Root reports full Go and 1,285 UI checks passing before the last empty-bounds
correction; the API agent reports 111 operations with actual success/refusal
schema checks. Genuine runtime and final post-correction gates remain pending.

The final canonical-bounds focused rerun passes (workbench 0.021 seconds, HTTP
0.017 seconds), regenerating four success and 17 refusal wire cases. The complete
main Go suite also passes (cmd 10.558 seconds; workbench cached). Its first sandbox
attempt could not open httptest loopback sockets; the authorized loopback-enabled
rerun completed successfully. The strict source verdict is closed; genuine final
runtime/frontend acceptance remains with the integration owner.

## Final acceptance

Cross-layer review corrected resource:null presence and explicit empty native
bounds; Go, JavaScript and native-specific API schema admission now agree on
those refusals. Final direct Go tests capture four successes and 17 refusals.
Real donor/destination exports, immutable bytes, stale/canceled responses and
source-free metadata pass. The existing observed model checks acquisition only;
independent identity/provenance/order admission remains separate. No broader
native authoring, source attestation or registry is implied.
