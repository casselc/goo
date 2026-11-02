# Whole-product incident checkpoint — verification 312

Status: final actual browser journey passed in 38.9 seconds (43.4 seconds with
runner). All selected-fix checks ran in that same finite journey; 50 generated
JavaScript→Go transition/lease cases also passed. Source-only pre-fix findings
are distinguished from executed post-fix acceptance below.

## Task and boundaries

One responder starts in Activity, notices a POST timeout, finds its HTTP service,
customizes the route/method view, narrows to POST, inspects a correlated cross-service
trace, saves useful configurations and reopens the exact trace question. The
fixture has 107 incoming HTTP spans (100 GET, seven POST), two child storage spans,
and three logs including one correlated error. It uses ordinary real OTel ingestion
and an isolated generic Go store; no saved definitions are seeded through APIs.

Maintained execution: [incident journey](../../journeys/312-incident-checkpoint.spec.js).
The run records actual endpoint counts, response completion times/body sizes,
source scope, sampling bounds, checkpoints and page errors. Streaming lifetimes
are distinguished from JSON response latency. This is a finite task, not another
model/property matrix for existing functionality.

## Confirmed observations and priorities

| Finding | Executed evidence | User consequence / recommendation |
|---|---|---|
| P1: Activity-to-service handoff is manual | The actual error row offered service/namespace filtering, Inspect, Open trace and Select, but no service dashboard entry. Services environment started blank. The responder re-entered it and rediscovered. | Add a checked service handoff carrying environment policy, namespace and observation window; do not reconstruct or silently widen scope from a service name. |
| P1: comparison availability is not explained by the actual constraint | Default HTTP route view had a genuine applied single-field breakdown, but Compare rest was disabled. Tooltip nevertheless said it required an applied single-field breakdown. GUI status-projection removal/Run enabled it. | Reconcile count capability with existing projections, or explain the real supported transition. Do not make responders infer internal projection flags from an unrelated tooltip. |
| P2: service entry repeats discovery | Two identical `/api/discovery` requests used the same resolved scope. The manual service-entry segment added eight source requests. | Review reuse of checked metadata within the existing visit/scope owner before adding another read/cache owner. This run does not establish a production bottleneck. |
| Explicit sampling and trace expansion work | HTTP result was 107 with 100 inspection rows; POST refinement was seven. Trace expansion explicitly dropped service/method filtering in favor of exact trace ID, returning three spans across two services plus one correlated log. | Keep counts/page extent/upstream uncertainty visible; preserve the explicit trace-scope expansion and its return context. |
| Persistence stores the question rather than silently freezing observations | Cohort and trace configurations saved successfully. Reopening staged the exact trace question/window without a read; explicit Run returned three again. | Keep configuration/evidence wording distinct. Joining these items into an investigation document was not exercised, so no missing-document capability is claimed. |

Activity's held row covered a 60-second read window. Re-discovery resolved a new
120-second window, including a new end time. The manual detour therefore changed
both context and time; it was not merely navigation to the same checked question.

## Selected-fix coverage

The completed primary run held one actual counted-values refresh while opening a child,
returned to the retained parent, released the old response and checked readiness with
no Back read. Clear/Undo recreated an identical selection (ABA); the four old commands were
reattached inside the current tray, so connection/container guards were satisfied.
They neither read, opened a child nor mutated the retained POST selection. A live
service-filter draft disabled all current tray commands; discarding it restored
a valid Apply, which explicitly read the seven-record POST cohort.

A completed secondary GUI-only lane explicitly removed the status projection, ran the
ordinary count profile and staged selected-versus-rest. It checked the new
**Include comparison without facet** action: source-free staging, preserved query
filters/records and an admitted fixed comparison, without claiming a pre-fix runtime failure. The contradiction was source-confirmed;
this execution proves the repaired action, not an unexecuted historical state.

## Executed results and limits

- Browser: one complete incident/fix flow passed on frozen source, **38.9s**
  (**43.4s** runner), with no page errors. The extra final run captured the new
  promotion wording and the Activity window; it was not another feature matrix.
- Fifty [generated transition cases](../../verification/incident_transition_property_test.go) passed in **16.36s** (package **16.393s**):
  actual partition helpers → atomic promotion → Go validation/YAML round-trip,
  exact retained query context/records, invalid context refusal and no mutation.
  Actual lease helper cases independently reject old owners, changed epochs/
  drafts/query/selection and every readiness blocker; renewed owners work.
- No new shadow model was introduced for existing functionality. Backend's actual
  transport/page matrix and focused frontend tests remain separate evidence.

| Checkpoint | Cumulative source requests | Meaning |
|---|---:|---|
| Activity error observed | 1 | One explicit tail connection; three logs held |
| Service rediscovered/opened | 9 | Manual discovery/service visit; eight additional source requests |
| Customized HTTP read + methods | 11 | Explicit root and counted-values reads; 107 total / 100 rows |
| Pending lookup / child / Back | 13 | One intentionally aborted refresh plus child read; Back added none |
| Valid POST scope preserved | 15 | Explicit Apply and values read; seven matching spans |
| Correlated trace inspected | 17 | Explicit spans and linked-log reads |
| Trace view Run/reopened Run | 19 | Two narrow page reads; three spans each |
| Ordinary comparison configured | 21 | Explicit count-profile Run and values read; promotion itself added none |

There were three successful configuration saves and one pure template expansion.
One count refresh was intentionally canceled by navigation. Optional declared
operations were unconfigured in generic mode (`503`); that availability response
was not treated as healthy absence or a page error.

### Single local-run response measurements

These are browser completion times on a disposable local fixture, not production
SLOs or a load benchmark. Streaming duration is excluded from JSON latency.

| Endpoint | Requests | Median / max completion | Largest JSON response |
|---|---:|---:|---:|
| discovery | 2 | 435.5 / 471ms | 7,334B |
| full explore | 7 | 361 / 636ms | 92,425B |
| counted values | 4 (one canceled) | 83 / 111ms, completed only | 1,102B |
| record page | 2 | 250.5 / 340ms | 3,514B |
| configuration save | 3 | 41 / 95ms | 2,374B |
| pure template expansion | 1 | 342 / 342ms | 1,553B |

The 100-row cap was exercised against 107 matching HTTP spans. This run did not
stress stream overflow, long retention, production-scale cardinality or concurrent
responders. No absence of those failures is inferred. Compact request summaries,
actual scope/checkpoint data and maintained screenshots are sufficient for this
checkpoint; raw temporary stores are not required for future product use.
