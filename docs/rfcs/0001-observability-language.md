# RFC 0001 Observability language and resources

Status: Proposed. Date: 2026-10-16. The runnable prototype does not implement this RFC.

## Problem

Teams need a common way to define observability meaning beside application code and explore that meaning interactively. Independent chart SQL, alert SQL, and SLO code drift. A chart builder alone cannot explain subject grain, missing evidence, asynchronous completion, or whether a proposed transformation is valid.

The initial backend is ClickHouse with ClickStack-shaped OTel logs, metrics, and traces. Initial workloads are web requests, internal gRPC calls, and event consumers. The initial interface supplements HyperDX and may replace selected workflows as coverage grows.

## Proposal

Define an expression grammar and a resource model separately. Expressions describe populations, transformations, measurements, relations, and visual encodings. Resources give expressions names, parameters, ownership, revision history, and operational behavior.

| Resource | Role |
| --- | --- |
| Source profile | Physical schema bindings, field types, units, conventions, capabilities |
| Population | Source selection and counting unit |
| Measure | Transformation and aggregation over a population |
| SLI | Eligible population and success measurement, including evidence policy |
| SLO | SLI, target, accounting window, definition revision |
| Alarm | Expression, schedule, condition, state policy, route references |
| Recording rule | Persisted evaluation with cadence, dimensions, and retention |
| View | Encodings, layers, scales, coordinates, facets, and allowed interactions |
| Dashboard | View arrangement and parameter bindings |
| Investigation | Definition revision plus user selections, branches, and pinned evidence |

Authoring syntax should lower to a single normalized representation. JSON is the first interoperable transport; EDN can provide native Clojure authoring. Syntax selection must not split the semantic implementations. Templates are versioned expansions producing ordinary resources; template expansion should be inspectable.

## Expression sketch

The following is proposed syntax. It is not accepted by the current /api/query endpoint.

```clojure
{:id :reservations/confirmation
 :population
 {:op :deadline-outcome
  :starts {:op :source :ref :reservations/accepted}
  :terminals {:op :source :ref :reservations/terminal}
  :key [:environment :tenant/id :reservation/id]
  :deadline {:value 60 :unit :second}
  :success [:= [:field :outcome] "confirmed"]}
 :measures
 {:eligible {:op :count-subjects :population :mature}
  :good {:op :count-subjects
         :population {:op :filter :input :mature
                      :predicate [:= [:field :classification] :good]}}}
 :sli {:op :ratio :numerator :good :denominator :eligible}}
```

The first implementation may use fewer resource kinds and different concrete AST fields. Preserve the distinctions between expression, named reference, literal value, and field access. Require an API version and reject unknown operators rather than interpreting arbitrary executable code.

## Types and inference

An expression type includes subject identity and grain; field names and types; numeric units; available dimensions; temporal anchor and interval semantics; aggregation representation; provenance; and evidence quality. A distribution, a scalar quantile, and a cumulative counter are different types even when their payloads contain numbers.

Type checking should explain why an operation is unavailable. A ratio requires compatible windows and grouping and a compatible denominator. Its numerator may be a proven subset of the denominator population; requiring identical filters would incorrectly reject ordinary success SLIs. Preserve population lineage and make subset relationships explicit where they cannot be inferred.

Scope changes are explicit. Relating a request to its spans yields span evidence; it does not mutate a request-count measure into a span-count measure. An aggregation removes dimensions unless their information is retained. A renderer cannot recover source records from an aggregate without an available evidence relation or source query.

## Composition and presentation

Filtering, derivation, grouping, aggregation, windowing, and relation traversal are ordered transformations. Do not assume commutativity. View composition has its own operations: layer, facet, align, and arrange. Multiple views can share a population and receive the same selection without becoming the same query.

The UI inspects the current type to offer controls. Users can still discover raw attributes not named in a package. Data changes that affect semantic assumptions should produce a validation or evidence state, not a quiet change to a chart.

## Execution and publication

Use the same normalized expression for the reference evaluator, SQL compiler, scheduled evaluator, and interactive UI. Capture evaluation time once. Parameterize values and validate identifiers through source profiles. An explain result includes inferred type, normalized expression, source/profile revisions, selected bounds, SQL, and evidence assumptions.

Publishing creates an immutable resource revision. Interactive changes create a session overlay. Proposed code changes can be exported for review; a dashboard filter change must not silently change the deployed objective. Package ownership determines reconciliation scope. Secrets are external route/source bindings.

HyperDX export reports capability per resource. Unsupported constructs remain unsupported unless an approximation is explicit and reviewed. Import/export need not be lossless across different semantic models.

## First acceptance example

Define an HTTP or gRPC success ratio by adding data only. Both evaluators must execute it without a new view-name branch. Reuse the measure in a time-series view and a scheduled-evaluation test. The UI changes group or filter by changing the expression, and explain shows the same definition used by the evaluator.

## Open decisions

- EDN authoring and Clojure implementation timing versus extending the Python core first.
- Minimal type representation and population-lineage proof rules.
- Whether deadline-outcome is a core operator or a standard-library expansion with an explicit typed temporal join.
- Precise package reference and revision syntax.
- Renderer choices for general charts versus waterfalls and evidence browsing.

The first slice should settle only the decisions needed for its example. Keep syntax refinements compatible through normalization or a documented migration.
