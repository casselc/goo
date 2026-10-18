# Core primitives review

Reviewed 2026-10-18 against the implemented Go models and current workbench.
This is a proposed consolidation plan, not a claim of implemented capabilities.

The most useful next work is to make the same question retain its meaning across
Explore, measurement packages, recipes, captures and backend compilation. More
individual views will otherwise keep adding special cases to those handoffs.

## What already works

Definitions, immutable evidence captures and investigation revisions are useful
product objects. Keep them. Go already owns query validation, exact count/duration
semantics, native metric kernels and bounded source execution. The capture store
already dispatches typed questions; unification should build on that tagged model,
not replace typed signal semantics with an untyped query map.

The seams are concrete:

- `exploration.Query` mixes population/time, aggregate choices and inspection paging.
  Measures are switches (`summary`, `spanStatus`, `bucketMs`), not composable measures.
- `exploration.TelemetryPackage` and `measurements.Definition` have different
  authoring, source selectors and execution models. The `measurement` package's
  prepared-fact compiler is another bounded semantic model, including warehouse output.
- `exploration.CaptureQuestion` unifies storage dispatch, but is not yet a common
  planning/capability contract. Legacy operation receipts remain separate.
- `exploration.Recipe`, `measurements.Workflow` and `exploration.Followup` each
  describe next steps differently. Requirements are largely explanatory text.
- `workbench.Document` validates capture references on evidence cards, but findings
  and evidence-gap citations remain text rather than structured evidence links.

## Priorities

| Order | Missing/shared primitive | Simplifies existing use cases | First useful increment |
| --- | --- | --- | --- |
| 1 | Resolved question scope and explicit scope changes | Filters, reusable lookbacks, fixed captures, recipe previews, trace/log navigation, resource discovery | Shared Go scope descriptor and checked change description over existing typed questions |
| 2 | Typed measures, dimensions and result shape | Duration/status switches, arbitrary numeric attributes, grouped time charts, comparisons, reusable definitions and warehouse lowering | Record count plus duration measures represented explicitly; preserve existing outputs through an adapter |
| 3 | Machine-readable source/backend capabilities and requirements | Metric discovery, semantic-convention presets, recipes, annotation gaps and portability diagnostics | Source-free planning reports supported/unsupported/needs-binding per requested operation |
| 4 | Declared identities and relationships | Trace links, business subjects, async work, log correlation, fleet membership and sample-independent navigation | Exact trace/span and service/time relations share one checked relation vocabulary; scoped subjects follow separately |
| 5 | Shared execution/evidence metadata | Captures across families, partial reads, query cancellation, refresh and bitemporal corrections | Common metadata around typed payloads; retain each kernel's detailed states and receipts |
| 6 | Structured investigation steps and citations | Recipes, multistep workflows, contextual follow-ups, competing explanations and handoffs | Typed planned step referencing its origin question/capture; findings cite specific evidence explicitly |

These are incremental contracts, not six prerequisites for shipping another
feature. Scope plus one end-to-end measure path should land first.

## 1. Question scope is more than time and service

A scope descriptor should identify the source binding/revision, environment,
half-open event-time window, population selector and declared grain. Grain must
say whether we count span records, selected server spans, metric streams or
scoped subjects. Root requests and unique operations cannot be inferred by simply
renaming span counts. Inspection page size, local search and chart presentation
must not become population identity.

Keep relative lookback as reusable intent; resolve it once into fixed bounds for
execution. Keep evaluation time and evidence cutoff separate where supported.
Do not invent a custody timestamp for adapters that have no custody history.

A question transform should return the next typed question plus a checked change:
what is identical, replaced or widened, its origin, and the relationship used.
Descriptions should come from the actual question difference. Exact-ID trace
navigation replaces predicates; service/time discovery is contextual. Neither
should claim to preserve the complete population merely because time is unchanged.

First integrate existing recipes and follow-ups. Keep old wire formats and hashes
stable; derive the descriptor without rewriting stored captures. Typed native
measurements can join this contract through an adapter rather than sharing record
predicates that have different meaning.

## 2. Measures need a small typed algebra

Represent a measure's input field/type/unit, aggregation, grouping and time
bucketing explicitly. Start with record count and current duration sum/min/max/
exact percentiles; add numeric attribute conversion and grouped time series next.
Casting needs an explicit policy for absent, empty, malformed, nonfinite and
out-of-range values, with invalid-input accounting. It must never coerce to zero.

Derived ratios need compatible population/basis and explicit denominator and
no-data rules. Counter rates still require reset/temporality handling; native
histograms retain bucket/schema semantics. An average or scalar percentile is not
a mergeable state. Keep the existing specialized kernels behind typed operators.

This is the missing bridge to portable compilation. Lower supported operators to
ClickHouse and, later, the Databricks metric-view profile. Report unsupported grain,
precision, percentile or temporal semantics explicitly. The existing prepared-fact
Databricks compiler does not establish that live OTel questions translate today.
Choose CEL/KCL authoring after these types and lowering rules are explicit; neither
a new syntax nor DataFusion supplies the observability semantics. A query engine
could execute/optimize a lowered plan without owning evidence policy.

## 3. Capabilities should drive discovery and planning

Separate static adapter/backend support from observed field/instrument metadata
and from verified coverage. An observed instrument is not a completeness proof.
A source-free plan can check declarations and types; discovering data is a separate
explicit read with its own window and evidence basis.

Use structured requirements: field/type/unit, identity, outcome policy, metric
kind/temporality, supported operators and budgets. Plans should return typed
reason codes plus human explanations, including a proposed instrumentation gap
when a declared requirement is unmet. This lets USE/RED and OTel conventions
provide templates through the same planner rather than new hardcoded screens.
An absent observation remains unknown; an unsupported operator is a planning error,
not a no-data result. Version source bindings before offering source onboarding.

## 4. Relations need identity, cardinality and scope

Declare exact trace/span matches, parent edges, scoped subject identity, declared
membership and service/time context as distinct kinds. Include tenant/namespace
keys, permitted direction, time expansion and ambiguity policy. Cross-signal
attribute translation requires a declared mapping; matching attribute names alone
is insufficient. Async span links are distinct from parentage.

The same relation descriptor should drive navigation, capture provenance and
follow-up previews. It must retain missing/ambiguous/invalid outcomes. A selected
trace can still be partial. Dynamic membership requires activation windows and a
coverage policy before it can define a fleet denominator.

## 5. Evidence needs common metadata, not one universal health flag

Share question/definition/adapter revisions, resolved scope, read/evaluation times,
source bases, budgets/truncation and relation provenance around typed results.
Keep execution failure/cancellation, valid no-data, invalid observations, disagreement
between reads and unknown coverage separate. Successful SQL does not certify data
quality. Retain per-component evidence when separate reads disagree; do not imply
a transaction when counts happen to agree.

Add actual collector/exporter lag, drops, queues and sampling evidence as an
independent pipeline-health source. Event age alone is insufficient. Extend custody
and correction lineage only where the source supports them; immutable capture time
is not equivalent to source ingestion time or historical replay.

## 6. Investigations need explicit evidence links

Use one planned-step contract for a recipe, workflow step or contextual follow-up:
origin, typed question transform, requirements, scope change and explicit execution.
Keep a planned question, execution and captured evidence as different objects.
Do not build an automatic agent or workflow scheduler just to connect them.

Add structured links from an authored claim/open question to exact captures and,
where supported, a selected row/group/measure. Links identify evidence; they do not
prove the claim. This supports competing hypotheses and a later successor comparison
without inventing automatic causality or silently moving old references. A spatial
board or real-time collaboration can reuse these objects later.

## Next implementation slice and acceptance

Implement a source-free scope/transform contract for existing record recipes and
follow-ups, with adapters for the current capture question types. Expose source,
grain, fixed bounds and structured changes in their existing previews. No new
navigation destination or saved-object family is needed.

Acceptance:

1. The same applied question yields the same normalized semantic scope through
   recipe preview, package round-trip and capture/reopen; page changes do not alter it.
2. Duration/status transforms preserve the complete selector; trace navigation
   reports its replacement; resource discovery reports service/time context.
3. Plans are source-free, explicit execution remains required, and old capture
   JSON/hashes and reopen behavior remain valid.
4. Hegel state/swarm checks compose transforms and mutate returned plans without
   changing origins. Quint validates observed handoffs and rejects scope drift;
   Bombadil explores preview/execution/draft guards in existing real journeys.
5. Existing screenshots/diagrams are regenerated only when their visible behavior
   changes. Do not add a new parallel demo to demonstrate this consolidation.

Then add the first explicit typed record-measure path and grouped time series,
using the shared scope and capability contract. Migrate only equivalent semantics;
keep adapters for old definitions and reject lossy conversions.

## Separate deployment gate

For a shared deployment, authentication/tenant isolation, scoped source credentials,
retention/quotas, backups/audit and per-tenant query budgets are core requirements.
They should precede exposing this lab to teams. Durable SLO evaluation, alert state
and notification delivery require a separate execution/storage milestone. These
remain important, but do not replace the semantic consolidation above.
