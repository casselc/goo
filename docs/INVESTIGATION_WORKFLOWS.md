# Guided investigations

The investigation page offers **Guided investigations → Investigation question**.
Each recipe explains a question, lists evidence limits, and provides ordered
query actions. These are optional paths through the same expression API: users
can change filters, select a heatmap cell, inspect spans, or edit the expression
between steps. The currently displayed view is highlighted in the guide.

| Question | Path | What it teaches |
| --- | --- | --- |
| Why is completion delayed? | Outcome cohorts → acceptance heatmap → async waterfall → lifecycle logs | A fast synchronous request does not establish a successful business outcome |
| Did a retry recover? | Retry outcomes → attempt spans → retry/terminal logs | An error attempt and a failed subject have different denominators |
| Is the service slow? (RED) | Volume/duration heatmap → span status → exported counters | Raw request counts, durations, errors and counter rates have distinct contracts |
| Could resource pressure explain it? (USE) | Queue/worker evidence → work/error spans → supporting events | Check utilization, saturation and errors for a resource; explicitly identify missing measurements |

Query steps retain the selected reservation where supported. The retry recipe
sets its scenario filter and clears an incompatible prior reservation selection.
The metric inventory explicitly switches to service scope and clears unsupported
reservation/scenario filters. Moving from a heatmap selection into a guided query
removes that cell's interval; the reservation selection remains. The expression
and visible selection always expose the actual scope. No recipe edits a managed
measurement definition or sends notifications.

## Evidence availability in the reference adapter

USE now opens **Queue and workers**: Redis group lag (undelivered), pending
(delivered/unacknowledged), active worker count, dispatch capacity, worker
utilization, and observed first-attempt queue wait. These are custom instruments
with standard messaging identity attributes, not invented OTel metric names.
See [queue contracts](QUEUE_TELEMETRY.md).

CPU, memory and database-pool utilization remain uninstrumented in this sample.
Pending includes active and stale work and is not equivalent to backlog. Broker
failures and unknown lag emit no observations; stale/no-data states remain
unknown. A ratio for active workers cannot establish host CPU or memory health.

RED is partial: the heatmap uses actual gRPC server spans, with counts per UTC
minute and duration bands. Sampling and ingestion completeness are unverified.
It does not produce a reset-aware counter rate or full service error ratio.

The [semantic-package layer](SEMANTIC_PACKAGES.md) generalizes these workflows
beyond this application: it recognizes versioned contracts and proposes reusable
capabilities. Most catalog outputs are proposed; the live adapter is one concrete
integration, not the product's package resolver.

## Design influences and reuse

- Brendan Gregg's [USE method](https://www.brendangregg.com/usemethod.html), already
  recorded in SOURCES.md: resource checklist and explicit measurement gaps.
- RED: service request volume, errors and duration; the guide marks the subset
  supported by this sample instead of presenting an unsupported scorecard.
- The core analysis loop recorded in SOURCES.md: ask a question, inspect evidence,
  refine the selection, then follow a relation. These guides suggest actions;
  they do not diagnose causation automatically.
- `oscope/docs/demo/README.md`: follow a queue signal and carry query selection
  into the next visualization. Oscope has its own queue-depth gauge fixture;
  goo now has a separate Redis-group adapter with its own explicit gauge semantics.

Definitions live as data in `ui/workflows.js`. They call the same validated query
compiler as manual controls. A later Go/Rust integration should lift these into
versioned recipe descriptors with typed requirements, supported scopes and
result-dependent suggestions. Avoid making each methodology a separate runtime
or reimplementing measurement semantics in browser code.

## Next useful extensions

1. Expand the Go contract proposals into typed measures/views/recipes after
   explicit source and policy binding. Add broker, framework and library adapters
   without adding application-specific runtime branches to the package model.
2. Add a typed span-status error measure and traffic rate with declared
   sampling/completeness assumptions for a proper RED overview.
3. Add persisted hypothesis/selection history and cohort comparison so users can
   branch an investigation and return to the same evidence basis.
4. Connect the Go lifecycle receipts to a historical correction guide. The CLI
   and Quint/Hegel checks already cover portions of its semantics, but the live
   UI must first acquire actual basis/receipt support.

The [browser storyboards](journeys/README.md) execute a selected retry through the
log step, check USE capabilities/gaps, switch to resource telemetry, and open RED volume/duration. They use real services and no UI mocks.

## Shared package workflows

The primary Examples UI now reads five strict, versioned descriptors from
[config/workflows.json](../config/workflows.json): delay, retry, RED, USE and
reports. Each declares requirements and ordered measurement/telemetry package
steps, with a shared fixed two-minute window. Opening a step validates/plans
without source execution. Measurement Run or the workflow package's Run fixed
question in Explore explicitly reads that scope. Capture and notes use common
product tools. Subject/event inspection carries explicit tenant and subject
identity; resource steps disclose scope changes rather than inventing continuity.
The earlier JavaScript recipes described above remain in the labelled compatibility
lab. See [MEASUREMENT_PACKAGES.md](MEASUREMENT_PACKAGES.md) for current boundaries.
