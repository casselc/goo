# CEL in the observability product: domain review

2026-11-14. Independent source/document review; recommendations below are proposed unless explicitly marked implemented. No source execution or new acceptance test run is claimed. This review is about general OTel investigation, not expansion of the reservation example.

## Three decisions to make first

1. **Ship the draft bridge against the existing fixed-field contract.** Checked CEL is an authoring projection of an ordinary record predicate, never another acquisition mode. Preserve the entire applied question and its counting grain. Unsupported visual-to-text conversion must leave the draft intact and explain its unsupported construct. Maps need a separate presence decision before they enter that bridge.
2. **Make source/profile capabilities and producer identity visible before extending the language.** Logs/spans and native metrics already have materially different bindings. A discovered metric must either run under a declared resource-grain contract or explain why it cannot; valid OTel without `service.instance.id` must not be presented as invalid telemetry. Do not remove native identity guards until a bounded generic-stream contract replaces their service policy.
3. **Give every continuation a declared population and evidence relation.** Log records, spans, reported trace IDs, server-call spans, metric streams and business subjects are different counting units. A trace search is currently independent existence tests, not ancestry. Full-window source counts and held-page Find remain distinct. CEL must not make these distinctions less visible.

These are small semantic contracts, not a prerequisite to implement every future helper or whole-query language. 436's aggregate/held-return consolidation remains the numbered operator priority; the CEL bridge can reuse its closed staging intents.

## What exists and what it means

| Domain | Implemented evidence | Consequence for CEL |
| --- | --- | --- |
| Fixed record selection | `querycel/compile.go`, `program.go`, `library.go`; `docs/CEL_PREDICATE_PILOT.md`: logs/spans, fixed nonnullable strings/uints, checked overloads, numeric severity, bounded reference evaluation, ordinary AST refinement | Useful first authoring slice. It is neither native metric selection nor a whole-query builder. No CEL UI exists. |
| Physical record binding | `exploration/compiler.go`: logical columns map to exporter columns; attribute/resource access has separate `mapContains` presence; baseline timestamps and duration have explicit units | Bind capabilities in the adapter. Never make presentation code infer physical keys or coalesce missing attributes to empty strings. |
| Native metrics | `measurements/source.go`, `metrics/gauge.go`, `counter.go`, `counter_decimal.go`, `histogram.go`, `stock.go`; `docs/MEASUREMENT_PACKAGES.md` | Kernels already preserve identity, temporality, resets, gaps and distribution representation. An exported numeric Value is not a general record measure. |
| Quality | `scopedview/model.go`, `metrics/state.go`; native and lifecycle contracts in `docs/MEASUREMENT_PACKAGES.md` | Reuse typed quality/support. Freshly observed domain label `Unknown` is different from unknown observation quality. Empty results are not proof of healthy operation. |
| Relations | `exploration/trace_search.go`, trace page/waterfall modules, reported-exemplar context | Existing independent EXISTS conditions select reported trace IDs. Witnesses and bounded pages do not establish complete traces or causal joins. |
| Discovery/defaults | `docs/SOURCE_CAPABILITIES.md`, `METRIC_SCOPE_DISCOVERY.md`, `INSTRUMENT_PREFLIGHT.md`; `exploration/dashboard.go` and profile tests | Partial installations and metric-only entry are real supported journeys. Defaults must follow discovered capability, not require demo instrumentation. |
| Reuse/custody | Measurement packages, portable view artifacts, immutable captures and exact revision annotations | A definition is intent; a captured read is evidence; a note is an assertion. CEL source belongs with authoring provenance, not as a substitute for these owners. |

The current roadmap explicitly defers generic resource/scope/instrument evaluation. The known issue on mandatory instance identity is significant: `metrics/resource_identity.go` requires nonempty service name and instance ID for service observations, whereas host/pod profiles have other identity policies. This is a Goo evaluator restriction, not an OTel validity rule.

## Bounded domain contract

Use existing typed questions and family evaluators as the semantic core. Add a small shared planning description at authoring/discovery boundaries, rather than a universal row evaluator:

- **Population:** signal/family, counting unit, exact scope, predicate and optional declared operation profile. A span count remains a span count even when its kind is server.
- **Fields:** qualified namespace, physical/source availability, logical type, unit, presence semantics and semantic-profile revision. Resource, instrumentation scope and record/point attributes are distinct namespaces.
- **Time:** event/observation anchor, half-open fixed query bounds, fixed evaluation time, and separately read/custody time. Family-specific start/end intervals and lifecycle knowledge cutoff remain explicit.
- **Evidence:** source/read basis, held support versus source population, sampling/completeness knowledge, quality diagnostics and result limits. Successful execution is not a completeness attestation.
- **Relations:** a finite relation kind and explicit source/destination grain, selector bounds and evidence basis. Selection of a related population changes the question explicitly.

This is a proposed contract for composition and admission, not a claim these fields already exist in one DTO. Keep native, lifecycle and record results typed. Avoid a new generic `unknown` bit that erases pending, invalid, stale, refused, no-data and unavailable distinctions.

### Presence and attribute types

The highest-value language extension after fixed fields is guarded qualified attributes. Before implementation, specify a truth table for absent, present-empty, present-value and malformed values. Goo's current predicate SQL already carries presence separately; CEL missing-map access errors. Naively lowering `attrs["key"] != "x"` or applying Boolean negation can change membership.

Recommend admitting only a reviewed guarded form first: existence/missing tests and presence-guarded exact comparison or literal contains, lowered to existing AST leaves. Decide explicitly whether a missing arm outside a guard refuses or evaluates false; preserve the same behavior in CEL reference evaluation and SQL. Do not infer it from short-circuit behavior, because today's reference evaluator validates every referenced fixed field before evaluation. This changes admission and needs its own properties.

Preserve the actual source type. String-exported attributes need explicit supported numeric interpretation; a numeric-looking string is not automatically an OTel number. Malformed numeric values need diagnostics or refusal, not silent zero. Do not promise complete OTLP AnyValue, arrays or scope attributes until the selected exporter/profile exposes them faithfully. The fixed `record.scope` currently denotes scope name, not arbitrary scope traversal.

### Resource, scope and schema evolution

Native stream identity already includes resource schema URL, full resource map, scope name/version/schema URL, metric name/unit and point attributes. Preserve separate streams across deployment/resource changes even if a service tile combines their display. Empty optional scope version/schema are actual exact identities. A schema URL does not authorize automatic field renaming or unit conversion.

Record CEL exposes scope name/version but not corresponding schema URLs. The record SQL projection is narrower than native source binding. Document that asymmetry; add schema fields only with adapter readiness/projection/reference parity. Distinguish three independent revisions: physical source binding, OTel semantic-convention profile and Goo authoring/helper library. Current/legacy HTTP, RPC, database and messaging profiles should remain exact selectable interpretations, never a catch-all heuristic that silently combines incompatible conventions.

A generic metric evaluation proposal should begin with a single exact discovered stream: full resource/scope/instrument/attribute identity, no synthetic instance ID, no cross-stream merge, existing interval/quality kernels. That would unlock non-service and otherwise valid metrics without first building arbitrary fleet SQL. Acceptance must include a metric-only installation without service identity and two streams differing only by schema/version/resource attribute. Existing service templates can retain stronger declared resource requirements.

### Measures, units and clocks

Do not add CEL `rate`, `p95`, `sum` or arithmetic over raw metric exports as ordinary helpers. Counter rates require start time, temporality, resets and boundary coverage; histograms require bucket compatibility and retain quantile ranges. Existing histogram support is explicit nonnegative duration buckets; discovery of exponential histograms does not make them executable. No average of exported scalar percentiles gives fleet p95.

A duration literal helper is useful only after exact lowering to existing uint nanoseconds is specified, including overflow, unit spelling and equality boundaries. It can be bounded constant conversion, not general unit arithmetic. Preserve uint64 values and decimal strings beyond JavaScript's exact integer range. Unit conversions must belong to explicit measurement/profile interpretation, not a chart label.

For logs, baseline event Timestamp is not observed/ingestion time. For spans, start plus duration is not a request completion cohort. Native start/end points identify accumulation intervals. Lifecycle first Goo observation is local custody time, not collector ingress; its correction mode advances knowledge while retaining the original business question. Add ingestion/observed-clock helpers only when adapters actually bind those clocks. A wider quiet-window query cannot diagnose retention or clock skew by itself.

### Quality, uncertainty and completeness

Keep measured zero, no matching records, unsupported signal, refused source, pending completion and unknown outcome separate. A source-wide aggregate can be exact over stored rows and still have unknown upstream coverage or sampling. It should say what was counted, not claim an unsampled census.

Represent sampling/completeness as supplied knowledge, not a deduction from trace flags, successful reads or pagination completion. Counter gaps and stale gauges remain uncertainty in their own measurement domains. No generic Boolean `healthy()` or `complete()` helper should precede a versioned evidence policy. A future SLI must declare subject grain, numerator/eligible relationship, maturity/unknown policy and evidence cutoff; native chart ratios and observed span error shares do not automatically establish that SLI.

### Trace relationships and joins

`exploration/trace_search.go` explicitly warns that conditions are independent EXISTS tests, one span can satisfy several, and no order/topology/causality follows. Keep that semantics. Raw span counts, valid reported trace IDs and service server-call counts must have different names; neither distinct trace ID nor root span establishes a unique end-user request across retries, missing roots or linked async traces.

Before any ancestry predicate, define exact parent edges by trace/span identity, handling duplicate/conflicting IDs, missing parents, cycles, truncation and bounded traversal. Span links are distinct reported relations across async boundaries, not implicit parents. Time-nearby logs remain contextual search. A reported metric exemplar is evidence of a reported association, not proof the span caused a bucket's entire population.

Prefer finite commands such as open reported trace, show exact span logs, find traces satisfying these conditions, or inspect a declared subject's events. A future ancestry query should be a new checked relation plan with its own bounded source implementation; do not disguise it as an arbitrary CEL function or SQL join.

## Helpers to prioritize

| Priority | Proposed vocabulary | Admission/evidence condition |
| --- | --- | --- |
| P0 | Existing fixed fields plus `otel.severityBand` | Already delivered; preserve numeric/text independence and unsupported-operation refusals in app bridge. |
| P1 | Qualified attribute presence and guarded exact values | Truth table, source type fidelity, CEL/AST/source parity and whole-parent retention first. |
| P1 | Exact duration literals and profile-aware span kind/status choices | Lower only into existing typed comparisons; no synthetic failure from HTTP text, log severity or arbitrary status strings. |
| P2 | Profile-specific HTTP/RPC/messaging/database predicates | Reuse dashboard profile meanings; show required fields/profile and refuse unavailable interpretation. Avoid an ambiguous universal `isError`. |
| Separate decision | Metric stream selection, ancestry, lifecycle policies, health/SLO functions | Typed family/relation contracts; these are not ordinary record helper functions. |

Do not block the fixed-field text bridge on P1/P2. No regex, arbitrary maps, UDF registry, joins or whole-query CEL is necessary for the next coherent release.

## Natural operator journeys and acceptance

1. **Unfamiliar logs:** discover available source and scope → inspect/hold Activity → Find locally → explicitly stage full-source Body search or numeric severity filter → run fixed question → compare source population → inspect trace and return to held evidence. Acceptance: >100 matching source rows; Find still refers only to held rows; a producer `FATAL` label with numeric Info stays Info; an unavailable source retains held evidence and explicit Retry.
2. **Slow calls:** choose spans and declared server/client/consumer operation profile → inspect latency distribution → select range → stage ordinary refinement → open representative reported trace → inspect exact logs or contextual nearby logs with distinct labels. Acceptance: multiple spans per trace and retries do not inflate a request claim; incomplete ancestry is explicit; Return performs no hidden requery.
3. **Metric-only deployment:** discover actual resource/scope/instrument identity → inspect unsupported metadata or preview admitted native definition → evaluate gauge/counter/histogram under existing quality rules → inspect interval/exemplar → retain definition or capture. Acceptance: missing optional tables do not block discovery; absent instance identity explains evaluator restriction; no fabricated instance, merged percentile or inferred healthy zero. Generic-stream execution remains proposed until its separate decision is accepted.
4. **Team reuse:** refine visually or with supported CEL → preview the same normalized question → export package/view intent for code review → explicitly publish/import → run and capture evidence → attach an exact-target note. Acceptance: text changes neither window nor grouping/measure/limit; failed compile leaves draft/result intact; equivalent source forms retain normalized semantic identity; deployed definition and local draft remain distinct.

The visual/text round trip should guarantee semantic equivalence for a declared subset, not byte-for-byte preservation of formatting. Keep original source/library revision as authoring metadata and canonical AST as execution authority. Unsupported existing visual predicates must remain editable visually without lossy text conversion. Every new helper needs deterministic reference/lowering tests, negative controls and parity over absent/empty/invalid boundaries when applicable; reuse existing source/UI admission tests rather than inventing a second query endpoint.

## Annotations, training and infrastructure as code

`docs/WORKBENCH_ANNOTATIONS.md` implements exact-target tags/notes and revision branches; it explicitly excludes classifier labels and health verdicts. Training vocabulary, reviewed labels and export are broader proposals in `docs/PRODUCT_ANNOTATIONS_DESIGN.md`. Keep assertions separate from telemetry fields and CEL predicates unless a future reviewed annotation relation is explicitly selected. Training exports need label vocabulary/revision, author/review status, target evidence identity, applicability time and sampling/support; current local unverified attribution is not authenticated ground truth.

`docs/GOO_ANNOTATIONS.md` describes a different annotation domain: source weaving emits standard OTel operation spans/events and explicit subject mappings. Its operation role supplies span kind, not protocol semantic conventions. Package templates should accept ordinary third-party OTel; instrumenting the demo or adopting weaving cannot be an entry requirement.

For IaC, export normalized typed question/measurement/view declarations and their profile/library references through existing package and artifact owners. CEL is convenient authoring metadata; do not require runtime CEL evaluation to execute an already admitted AST. Explicit preview/plan and publication preserve the difference between editing a local view, deploying an objective and capturing historical evidence. Alarm schedules, routes and recording rules in RFC 0001 remain proposed rather than implied by package portability.

## Alignment with earlier proposals

This follows RFC 0001's grain/unit/provenance types and one normalized representation, RFC 0002's evidence/maturity/temporality rules, and proposal 260's scope → typed dataset → compatible view model. It also follows 260's prerelease policy: change the canonical contract and callers together when necessary; do not add permanent compatibility fallbacks to development formats.

The useful correction to the historical broad language vision is sequencing: the delivered fixed-field compiler already shares the ordinary AST. Extend that authoring path and close discovery/held-return rough edges before proposing whole-query CEL or a universal expression engine. Preserve native kernel and evidence distinctions while sharing planning descriptions and user-facing staging. This yields a coherent product without making arbitrary SQL, reservation-specific subjects, or an exhaustive helper catalog its organizing model.
