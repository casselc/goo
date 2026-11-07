# Whole-product architecture checkpoint — 375

Status: proposal, not adoption. 373 discovery and 374's source/representation gate
are accepted; 375 responsive grouped-row acceptance remains pending here.

## What the completed sequence established

Native display configuration is now separate from measurement calculation,
view discovery is bounded and explicit, and typed record continuations preserve
fixed scope and current-owner authority. 375 reuses the shared stacked table so
identity/count/actions can remain together instead of maintaining a second mobile
UI. Its literal marker disambiguation changes presentation, never predicates.

374 established a material source-fidelity boundary, not a failed implementation:
real exponential points are retained in the deployed table, but zero threshold
and optional-statistic presence are lost. General execution is also blocked by
exact irrational bounds, negative closure and zero-region semantics in the current
explicit-bin representation. Do not reopen that decision as a rounded converter
or silently infer lost values. A source observation may still be discoverable
without being executable.

## 376 breadth: make unsupported observed instruments visible

Task: “My application exports a metric. Show that you observed it and why it
cannot be evaluated, rather than making it disappear.”

`measurements/discovery.go` currently unions only gauge, sum and explicit histogram
tables. Its instrument DTO already has `supported`, `unsupportedReason`, optional
Definition and capability metadata. The 374 real exponential observations are
therefore outside Goo inventory, even though the source contains them.

Add a bounded metadata-only exponential observation path with a distinct kind,
full reported stream identity and explicit unsupported/fidelity explanation.
Never synthesize a Definition, enable Preview/Run, convert bucket values or label
missing source fields as defaults. Use existing inventory/capability presenters.

Close source/table availability first: adding an unconditional UNION over an
optional absent table must not break existing supported discovery. An explicit
optional metadata action or a checked source-capability arm is preferable to a
silent fallback. Keep its read cost and partial capability clear. Retain existing
global row limits, typed environment/namespace membership and immutable source
identity; no broad source registry or per-vendor adapter framework.

Acceptance: actual 374-shaped instrument discovery and exact identity, no hidden
execution, unsupported reason and source-retention limitation visible, absent
optional table handled honestly, and unchanged supported explicit/gauge/counter
flows. This is not exponential kernel support or producer-conformance validation.

## 377 depth: one explicit histogram exemplar → trace readiness outcome

Task: from a selected supported metric export, follow a trace ID actually carried
by a sampled exemplar rather than guessing from coincident time.

374's pinned exporter source includes exemplar fields; the current measurement
source model/evaluator does not expose them. That is a useful research starting
point, not sufficient evidence for a working link. Begin with one existing
explicit histogram family and a real collector/table fixture. Pin retained trace/
span IDs, timestamps, value type/presence, filtered attributes, limits and exact
stream/export association. Distinguish invalid/unreported IDs and absent exemplars
from known empty data. Do not infer the original attribute types or optional
presence from flattened/defaulted source columns.

If the source contract closes, expose one explicitly sampled trace continuation
using existing trace planners/owners. Stage must preserve actual reported origin
and keep a broader trace query distinct from the selected metric interval. If
association or IDs cannot be retained honestly, publish a precise defer. No
exponential execution, automatic metric-span join, causality/contribution claim or
new query engine is required. Independent raw IDs/time/association checks and one
genuine source journey matter more than adding a catalogue of chart links.

## 378 consolidation: native refusal/recovery parity in Measurements

`internal/nativehttp` already runs `/measurements` through shared typed
`ReadHandler` refusals. Service widgets consume the shared source-refusal protocol,
but `ui/measurements.mjs:request` still flattens failures to `Error(data.error)` and
the workbench renders one generic error string. `measurements/source.go` also
lacks the explicit postcompile budget/protocol classification used by ordinary
record source boundaries. These are concrete consistency gaps in the primary
native workflow, not missing new backend infrastructure.

Use the existing five-kind refusal DTO/passive renderer and explicit same-fixed-
question retry. Distinguish pre-source invalid/admission errors from source
refusal; preserve unknown/no-data and never show healthy zeros. Apply typed native
classification only around source transport/decoding/evaluation after successful
validation/compile; optional custody publication errors must not be mislabeled
malformed telemetry. Start with the supported native measurement path, retaining
the existing read-with-optional-custody effect boundary. Keep fixed exact native
bounds, generation cancellation and unrelated metadata/display intent intact.
Retry remains user initiated and cannot advance a preview lookback silently.
No raw-message matching, scheduler or global recovery manager.

Acceptance should cover one success/refusal/retry and a retired pending read in
the actual Measurements entry, plus budget/malformed/timeout classification at
the bounded source boundary. Reuse existing independent kernel semantics and
owner schedules; do not rebuild a model for a passive reason formatter.

## Reconcile and keep the scope finite

373's bounded discovery closes native-view exact-ID-only friction. 370/372 close
the selected typed follow-through paths; do not resurrect their old scalar-only
barriers as a product requirement. Saved/service interval context, current RPC,
trace tails, captures and bounded review/export already exist. Keep held return,
shared question, durable evidence and reusable configuration distinct.

Broker backlog remains separately pinned-exporter research; vague “broker health”
is not the next implementation. Review enumeration/authenticated consensus,
Databricks lowering, broad WYSIWYG and durable alert delivery remain independent
contracts. None requires a new store or evaluator merely because annotations or
histogram views exist.

Go 1.27.3, typed ordinary APIs, maps/slices cloning and errors.AsType are already in
use. Apply them to actual ownership/classification work, not a runtime-upgrade or
advanced-generics queue. Continue measured full-model presentation budgets and
independent numeric/identity oracles; do not resume an automatic facet optimization
loop after 369. Review the product after this finite outcome sequence or explicit
depth deferral, and remove covered consolidation work rather than manufacture it.
