# Whole-product architecture checkpoint after the 429 cycle

Status: independent source/document assessment and proposed priorities. The final
429 matched browser reports pass; its strict native/model verdict remains separate
at this review's writing.
This review ran no new source-load benchmark, deployment audit or vendor campaign.
The [429 boundary review](MEASURED_CONSOLIDATION_429_ARCH.md) records independent
focused checks; [validation](../../VALIDATION.md) owns executed integration claims.

## Assessment

Goo has a complete on-demand investigation loop: ordinary OTLP discovery and live
logs/spans, facets and comparisons, traces, service/resource dashboards, native
measurements, reusable declarative views, checked CEL refinement, saved definitions,
captures, annotations and investigation documents. The gap is not another chart
engine, saved-object family or replacement query planner.

427–429 strengthen finding and preserving an investigation. Their lessons are
broader than preservation: disclosed controls need real action guards; supplied
projection provenance must not misdescribe stored evidence; and an authorized
mutation may outlive its permission to navigate. Keep these distinct local owners
rather than replacing them with a universal identity or permission cache.

One immediate documentation defect was identified and corrected without consuming
another feature milestone: the [capability guide](../CAPABILITY_GUIDE.md) described
interactive CEL/436 as unimplemented and reconciliation only through 435. The
updated guide now includes contextual CEL, exact saved recall and paired capture,
and removes the historical milestone rollcall. Operator-facing capability claims
should follow actual delivery rather than historical “next” paragraphs. This was
source-of-truth maintenance, not evidence that those features were missing.

## Three bounded outcomes to consider

These are ranked proposals cross-reviewed with the planner, not an execution
claim. The finite proposal is 437 breadth, 438 native depth and 439 consolidation.

### 1. Reach live logs without first knowing an environment name (437)

Offer an explicit, bounded **All environments** scope within the configured source,
initially for ordinary log Activity. Generic startup currently requires discovery,
scope choice, staging and Start before seeing records. The source already supports
bounded all-environment presence checks, but the record scope model only represents
named, present-empty and unreported environments. A first-use operator cannot yet
ask “show me what is arriving here” directly.

This is a new typed scope arm, not an empty string, unknown-name fallback or an
inferred tenant. Selection stages locally; an explicit Start/Run owns the read.
Retain environment attributes/pills on individual observations and exact narrowing
from the mixed population. Keep fixed bounds, qualified fields and all parent
predicates when continuing into an ordinary query.

Acceptance gates:

- Define server-side workspace/source authorization semantics before exposing the
  broader scope. The loopback single-source process is not multi-user tenancy;
  selecting All must never widen the configured source's permissions.
- Carry explicit All through typed admission, SQL adapters, query identity,
  filters, links, saved views and captures. Never silently reinterpret empty or
  unreported as All, or claim native/resource profiles support it without proof.
- Exercise two named environments, missing/present-empty labels and current/legacy
  key disagreement with real OTLP. Verify full source budgets, typed refusals,
  explicit reads and source-free staging/held return at narrow and desktop widths.

**Why before mixed Logs + Spans:** both are useful entry improvements, but a mixed
tail additionally needs independent watermarks, per-family faults/limits, ordering
and retention. Treat signal composition and environment scope as separate policies;
do not combine both into one difficult-to-review stream rewrite.

### 2. Inspect bounded raw gauge observations when evaluation is unavailable (438)

Keep the native evaluator's instance/stream identity guards, while providing an
explicit read-only route from a discovered gauge to its raw exported points.
Start with gauges only and existing fixed service/environment/instrument scope.
This helps an operator understand valid OTLP that Goo currently cannot evaluate
because `service.instance.id` is absent, without manufacturing an instance ID.

This is a real gap rather than another discovery table: `DiscoveryObservation`
explicitly retains identity labels, not values or interval validity; the discovery
UI renders metadata, counts and dimension presence. Existing native sample tables
and timelines are reached after a successful kernel Run, which the missing-instance
preflight withholds. Existing host/pod/container identity profiles remain separate.

Acceptance gates: bound source rows/bytes/time and disclose truncation; preserve
exact resource/scope/instrument attributes, unit, timestamps, duplicate occurrences
and reported scalar domain. Reuse supplied-data tables and timeline geometry where
their profile is compatible. Raw points do not establish continuity, freshness,
rates, resets, aggregation, health or coverage. Do not broaden Build/Preview
eligibility or introduce a capture family in the first inspection slice. Keep the
raw-observation data profile explicit rather than wrapping it in a falsely
evaluated gauge result. A local row occurrence key is presentation state, not a
manufactured producer/instance identity or proof that duplicate rows are unique
events.

### 3. Measure and consolidate first-use acquisition work (439)

Exercise the adopted source-wide log entry and raw gauge inspection together with
ordinary Activity/fixed/held-return workflows on realistic label cardinality and
source latency. Record per-family readiness/metadata probes, query counts,
bytes/time, cancellation and local buffer/render work in cold and retained states.
Use the measurements to remove demonstrated repeated work or normalize identical
supplied projections; do not preselect a cache or controller rewrite.

Acceptance gates: no new reads on local disclosure, staging or Return; retain row,
byte, time and concurrency budgets, typed refusal state and exact configured scope.
If metadata reuse is justified, start request-local, with failure and cancellation
semantics explicit. A source/schema change between independent operations must
still refuse honestly. Compare measured work before/after without converting a
layout reduction or one lab run into a throughput claim.

The planner's proposed **437 → 438 → 439** cycle is architecturally sound. Raw
native inspection takes precedence over another language extension because visual
attribute filtering already works, while blocked native instruments have no value
inspection path. This is a prioritization choice, not a claim that attribute CEL
is unnecessary. The broader language option remains gated below.

The exact [reconciled roadmap](../ROADMAP_2026_11_14_CHECKPOINT_429.md) was
independently cross-reviewed and accepted: explicit All never becomes a fallback,
raw inspection remains distinct from evaluated results, and measured acquisition
work precedes caching or structural changes. No architectural blocker is identified.
The parent's full 429 frontend suite (1,816 passing, zero skips) and root Go checks
are separate from this proposal assessment. Actual matched browser reports retain
applied positions/actions and main source work while reducing edit/dialog overhead;
the [429 review](MEASURED_CONSOLIDATION_429_ARCH.md) records the precise contract.
The final strict native/model result remains parent-owned at this point.

## Cross-cutting engineering gates

Keep the current Go 1.27 single-server architecture, typed handler policies,
query cancellation, parameterized source adapters and framework-neutral display
components. The existing generic HTTP handler boundaries already use generics
without imposing type machinery on callers; no runtime/library rewrite is needed
merely because the application is prerelease.

For whichever cycle is adopted, include realistic high-cardinality and large-label
source fixtures. Record query/metadata request counts, source bytes/time, bounded
buffer work and cancellation/departure behavior. Measure cold and retained paths
separately. Do not introduce durable readiness caches or merge source controllers
unless actual repeated work and invalidation semantics justify it. Request-local
reuse is the smaller first optimization when measurements support it.

Local UI layout gains are not source latency or throughput gains. Internal request
diagnostics must not turn sensitive query parameters into unbounded metric labels.
Continue deterministic/property checks, actual OTLP journeys, scoped exploratory
actions and observed-model negative controls for meaning and authority.

## Explicitly deferred decisions

### Presence-safe qualified CEL attributes

Visual filtering already supports resource/record attributes and distinguishes
missing from present-empty. Contextual CEL is delivered, but its bounded vocabulary
still excludes attribute maps. An operator following ordinary OTel/library-specific
dimensions should be able to express the same supported attribute condition in the
existing additional-condition slot.

Start with literal qualified string keys and explicit presence semantics over the
current ordinary predicate AST. Preserve resource versus record/scope identity;
do not infer numeric types from strings or coerce missing to empty. Unsupported
dynamic access, types or expressions must fail before execution. This does not
require a second evaluator, general plugin registry or whole-query CEL language.

Acceptance gates: specify the presence contract first; reuse parameterized source
adapters; verify CEL/reference/ClickHouse parity for missing, empty and nonempty
values, punctuation and same-name qualified fields. Retain exact parent scope,
owned slot replacement, positioned diagnostics, saved authoring identity and stale
check refusal. Use current CEL-Go extension facilities behind a small ordinary API.


Production identity/source authorization, shared workspaces and proactive monitor
scheduling/delivery need their own operational design before team-facing claims.
Missing instance IDs remain an evaluator-policy decision, not an OTLP defect.
General metric-stream evaluation, exponential histograms, arbitrary executable
kinds/WYSIWYG, classifier task expansion and warehouse lowering retain their
existing semantic/operational gates. None should be smuggled into another small
preservation or formatting milestone.

Sources: [current startup](../GENERIC_STARTUP.md),
[source capabilities](../SOURCE_CAPABILITIES.md),
[instrument preflight](../INSTRUMENT_PREFLIGHT.md),
[discovery](../NATIVE_INSTRUMENT_DISCOVERY.md),
[checked CEL planning](../CEL_CHECKED_PLANNING.md),
[vendor gap review](PRODUCT_GAP_ANALYSIS_2026_11_12.md),
[post-CEL cycle](ROADMAP_AFTER_CEL_436_PROPOSAL.md),
`ui/record-environment.mjs`, `ui/stream-entry.mjs`,
`ui/metric-discovery.mjs`, `measurements/discovery.go`,
`measurements/instrument_capabilities.go`, and `internal/queryhttp`.
