# Whole-product reconciliation — 363

Adopted after the completed 359 sequence: 360 captured span pages, 361 retained
page comparison and 362 current RPC capability parity. This review selects the
next finite outcomes; none is implemented by this document. Read with the
[architecture review](WHOLE_PRODUCT_ARCHITECTURE_363.md),
[design review](WHOLE_PRODUCT_DESIGN_363.md) and
[active roadmap](../ROADMAP_2026_11_05.md).

Goo's loop remains **see telemetry → isolate a population → follow supporting
signals → preserve and reuse**. The next breadth cut connects retained metrics
to that investigation loop; the depth cut addresses the user's longstanding
reviewed-example goal if its new contract closes; consolidation makes an actual
comparison readable before exhaustive provenance. This is not a new maintenance
queue or a claim of broad observability-suite parity.

## Reconcile delivered, unfinished and outdated work

| Inventory | Decision |
| --- | --- |
| Stored waterfall, retained-page comparison, current RPC capability parity | Delivered in 360–362. Remove the corresponding absent/generic claims from active work; retain page limits, independent receipts and unverified metadata meaning. |
| Live native interval → records | Implemented. The saved-origin/return path is still absent: captured rendering does not supply `onRecordContext`. Extend that entry, not the planner or metric engine. |
| Tags/notes and optional XTDB evidence | Implemented, but not reviewed task labels. `MeaningTarget` has no capture-member target and file `CreateMeaning` rejects optional-store evidence. A new bounded target/judgment contract is a prerequisite. |
| 361 paired provenance; 358 result placement | Real unresolved usability cost. Stacked 361 facts precede timing by roughly a desktop viewport; narrow frames do not establish narrow chart appearance. 358's improved plots still remain below the fold. |
| Basic service/host/Kubernetes/native defaults | Substantial delivered coverage, including pod/container resources/state/history and host network/disk/filesystem. Do not reopen these as wholly missing dashboard families. |
| Broker backlog, exemplars, large facets | Relevant but not ready defaults. Broker/exporter identity and exemplar fields need research; the 32k-element facet concern needs a fresh interactive cost baseline before choosing machinery. |

## 364 — breadth: saved native interval to supporting records

**Task:** “I saved this metric interval; let me inspect related records and return
to the same original evidence.”

`ui/measurements.mjs::recordContextPlans`/`appendRecordContext`, the pure
`exploration.PlanNativeRecordContext` and live navigation already supply the
numerical and contextual semantics. `ui/captures.mjs` renders saved measurements
without that handoff. Start with that one saved-capture origin; defer service
cards and every resource origin until a real later task needs them.

Use an already-supported counter/histogram **export interval**, original validated
question and exact selected stream. Preserve half-open nanosecond membership,
enclosing millisecond bounds, typed environment policy and established resource/
service/namespace/instance identity. Keep the planner's existing admission limits.
No arbitrary lookaround, gauge/stock interval, aggregate time-bucket attribution,
exemplar or new metric-to-record join is inferred.

The saved origin carries an immutable capture citation, stream, interval, basis,
source-read and custody clocks. Stage reads no telemetry; explicit Run acquires
fresh contextual records with their own receipt/clock. Shared time/resource
identity does not establish contribution or cause. Return restores the original
saved evidence and current opener without evaluating the native metric again.

Adapt the existing origin/return owner deliberately: `native-record-navigation`
currently assumes the live Measurements host/workspace. Passing that callback
unchanged from a saved capture would be incorrect. Reuse the existing capture
visit and fixed-question owners, not a global navigation/cache framework.

**Acceptance:** genuine saved interval → Stage → exact logs/spans Run → failed/
cancelled child recovery → Return to the same capture. Include exact submillisecond
boundaries, typed policy, namespace/instance differences, retained unknown native
values where already admitted, stale/hidden actions and late replies. Preserve
full capture bytes and independent original clocks. Count store GETs separately
from telemetry, proving no metric reread on Stage/Back. Reuse existing context
arithmetic properties and drilldown models; native entry/return checks establish
the newly changed ownership. No new source endpoint or capture format is assumed.

## 365 — depth: one captured-record review task, contract first

**Task:** “Give this exact retained example a reviewed task-specific meaning that
I can later export, without making ordinary notes into training truth.”

This is new metadata semantics, not just an annotation button. The readiness gate
must close before implementation:

1. **Exact target:** demonstrate server resolution of immutable capture ID plus
   absolute retained occurrence. Bind the receipt and original stored member;
   visual position, trace/span ID, message text or browser-authored JSON cannot
   substitute. Duplicate payloads at different positions remain distinct.
2. **Declared task:** pin one bounded task/vocabulary version and record target
   grain. A judgment on one captured record does not label its population or a
   selected set. Do not start a free-form ontology or task registry.
3. **Append-only judgment/review:** define corrections, explicit abstention and
   uncertainty, withdrawal and concurrent/disputed branches. Unlabeled is not
   negative; absence is not abstention. No last-write-wins truth or authenticated
   reviewer claim arises from the lab's unverified local identity.
4. **Publication and export:** demonstrate bounded file-backed publication using
   existing workbench store primitives, plus a pinned bounded manifest of exact
   task/judgment revisions and target occurrences. Preserve exact original record
   values and separate event/custody/review times. Reopen/export reads stored
   evidence, never silently reruns telemetry or selects a mutable latest head.

Architecture and design must review that concrete target/schema/publication
proposal and small resolver demonstration. Reuse the existing capture store and
append-only metadata infrastructure; neither optional XTDB nor a new retention
backend, ML service or general evidence framework is a prerequisite. Notes/tags
remain descriptive context and do not become labels by prefix convention.

**Gate outcome:** if this contract cannot close in a bounded spike, explicitly
defer 365 and advance 366. Do not hold the known comparison layout problem hostage
to a new registry or expand the spike into a second storage platform.

**Acceptance after the gate:** an ordinary captured log/span occurrence → a value
or abstention under the one declared task → explicit review/correction → pinned
bounded manifest/export → reopen with telemetry unavailable. Include duplicate
records, second-page absolute-index mapping, exact large numeric/timestamp values,
uncertain/withdrawn/disputed outcomes and immutable older exports. User intent and
review state remain distinct from observed OTel status and producer conformance.
Use the existing contextual inspector/annotation vocabulary; no separate training
workspace, active learner, automatic labeling or broad dataset/split pipeline.

This outcome is selected ahead of broker depth because it directly serves an
explicit user goal and can reuse actual immutable captures. Its prerequisites
remain real and are not represented as already delivered.

## 366 — consolidate: compare results before full provenance

**Task:** “Show me both operands and their timing difference before a page of
paperwork, while keeping the original evidence inspectable.”

Start with 361's actual retained-page comparison. Measure fresh-entry 1280×1000
and 390×844 first substantive timing position, independent role visibility and
keyboard access. Replace stacked exhaustive fact panels with a compact paired
header and existing disclosures for complete questions/receipts. Keep baseline
and selection labels, policy/window differences, supplied-page limits, unknown
state and distinct source-read/custody clocks available and correctly scoped.

Reuse shared fact formatting and passive display primitives. Extend to another
paired owner only where the same primitive is demonstrated useful; do not mandate
a blanket migration, new sidebar or source/session owner. Dirty/invalid editing,
applied pair immutability, exact opaque side opening and local inspection/Return
must retain their current semantics.

**Acceptance:** matched genuine before/after entry frames show the actual timing
comparison, not only a count, scope card or manually task-scrolled crop. Report
actual positions and remaining below-fold content honestly. At narrow width,
inspect the chart itself, not just readable provenance. Disclosures, selection,
side opening and Back retain exact facts/bytes and their original source/store
budgets; compacting layout must not acquire data or conceal uncertainty.

## Verification effort and deferred work

Scale validation to the changed authority. 364 changes origin/return ownership:
use independent context properties plus focused actual navigation/late-response
checks and existing observed models. 365 adds target and revision semantics:
use pure admission/deterministic-export properties and meaningful branch/review
state tests where that lifecycle changes. 366 is presentation: focused mounted/
keyboard and genuine responsive-frame/IO checks matter more than another model
or rerunning every historical campaign. Reuse mechanical fixture/transport
helpers while keeping independent domain oracles. A small metadata description
change does not require a new bespoke verification framework.

Keep broker backlog research separate until one real exporter/name/unit/type/
temporality/resource/queue identity is pinned; `lab.queue.depth` is not that proof.
Exemplars need actual source fields, not inferred trace membership. Large-facet
work needs a fresh styled mount/search/action/Back baseline separating build,
layout/paint and wall time; old repeated-build CPU is not current paint cost.
Retain blocked vulnerability assessment, comparison-body/label byte limits and
historical harness failures explicitly. Production tenancy, verified topology,
WYSIWYG, Databricks lowering and store migration remain separately scoped work.

After these finite outcomes—or explicit 365 deferral—review the whole product
again. No fourth delivery or universal queue is automatically appended.
