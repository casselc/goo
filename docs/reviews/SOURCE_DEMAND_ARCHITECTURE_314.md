# Source-demand correctness review — 314

Status: architecture source review accepted; actual measurement and executed
checks are tracked in `VALIDATION.md`. This review
identifies one acquisition saving while preserving query intent, keeping aggregate
certainty explicit and retaining the meaning of captured/held operands.

## Chosen target: fixed-query baseline only

Before 314, `prepareCountComparison` in `ui/population-workspace.mjs` acquired the
fixed baseline with `/api/explore`, then submits only its checked cohort fields
(question, basis, capturedAt, count, state, groups and supported measures) to the
existing comparison endpoint. Baseline rows, field discovery, inspection paging,
result token and H0 display witnesses are unused by this branch. The root selected
result still needs full acquisition because its records remain visible/inspectable.

Switch this baseline demand to existing `/api/record-aggregates`, preserving the
exact baseline query and every profile/limit/bucket/predicate/environment field.
`ValidateRecordAggregatesQuery` already supports the ordinary single-string-group
count profile used here, including admitted duration/time projections. The endpoint
returns independent count/groups and optional time receipts, its own aggregate
basis and state. Submit those admitted cohort values to `/compare`; do not mint a
full exploration snapshot, row membership, result token or completeness guarantee.

Without timeline this removes baseline rows and field discovery: four statements
become two. With an admitted timeline five become three, retaining the existing
query rather than stripping its bucket to force a lower count. These are source
path expectations; actual live/SQL measurements must establish the delivered
budget. No automatic request cache, parallel executor or matching-query deduplication
is needed.

## Other operands deliberately unchanged

- Selected/root acquisition remains full for its rendered records and actions.
- Captured baselines are loaded from immutable storage; replacing them with a
  fresh query would alter the question and offers no source-read saving.
- Runtime remainder comparisons currently require a private admitted full groups
  snapshot in `SelectedCount.request`. An aggregate DTO cannot be passed through
  that validator by fabricating `viewDataset`; leave that branch unchanged until a
  separate real need justifies typed aggregate operand admission.
- Held parent remains the original supplied observation and is never re-read as
  an optimization.

## Shared aggregate admission

`ui/service-dashboard.mjs` exports `recordAggregates`, which already checks exact
question, version/state/provenance, independent receipts, typed measures, group
uniqueness/accounting and time buckets. Extract this pure admission contract for
both real callers rather than importing the full dashboard renderer graph into
population or copying its validator. Refuse unsupported response slots through a closed DTO contract, so claimed
rows/H0 snapshots cannot hitchhike through aggregate admission. Preserve the
existing caller-owned transport identity; later cohort construction owns copying. Keep actual read receipts and warning/state semantics.

The endpoint's observed state establishes count/group/time accounting, not absence
of contradictions in an unacquired row page. Full exploration can discover an
additional row/count inconsistency; aggregate-only acquisition cannot claim it
performed that check. Preserve the aggregate endpoint's own diagnostics rather
than relabeling it as full exploration certainty. Changed states remain diagnostic
through the existing comparison admission/components.

## Measurement and acceptance

The existing `cmd/goo-read-audit` supports `explore` and `record-aggregates` modes
and an aggregate fingerprint independent of DTO/provenance shape. Use the same
fixed question on controlled stable source data to compare shared count/group/
measure meaning, while reporting independent clocks/bases and active-traffic limits.
Do not use missing rows/fields or different provenance hashes as semantic equality
criteria. Report per-statement response bytes and elapsed time separately from
logical read count; no speculative speedup claim.

Prove exact baseline request preservation, two/three source statements as applicable,
no baseline row/field query, unchanged selected-root records and comparison values,
unchanged captured/held branches, no invented H0 snapshot, malformed/mismatched
aggregate refusal, diagnostic/no-data behavior, cancellation and late-result
retirement. Keep the existing single count controller/executor and Save/portable
question-only contract.

## Profile boundary refinement

Ordinary captured-count compatibility also permits numeric fixed grouping columns,
while the aggregate endpoint requires a string group. Therefore the optimization
must select the aggregate profile only after existing query/field validation and
an explicit string-group capability check. Go's current field map has exactly
three uint64 columns: `duration_ns`, `timestamp_ns`, and `severity_number`; validated
resource/attribute and other available column groups are strings. Retain full
baseline acquisition for those three unsupported groups before any request. Do
not catch aggregate errors and retry full reads, and do not broaden endpoint
admission merely to claim a universal optimization. Include cross-language
capability/negative fixtures and describe the measured saving as string-group only.


## Shared admission source review

The extracted pure protocol retains exact scope, provenance, measure/accounting
and diagnostic checks, adding closed top/nested DTO fields and receipts. It returns
caller-owned JSON as the former service adapter did, explicitly granting no
snapshot capability; `CapturedCount.cohort` owns subsequent copying. This avoids a
gratuitous DTO-identity change while preventing unattested rows/dataset claims.
The service wrapper retains its signature and the new capability helper excludes
only validated unsupported group/profile shapes. Source review accepts this lane;
fixed-baseline integration and actual measured checks remain pending.


## Final fixed-baseline integration review

The fixed-baseline helper clones and validates the requested query before IO,
chooses the aggregate endpoint only by supported-profile preflight, forwards the
existing controller signal and refuses aborted responses before admission. It
clones the payload before strict raw aggregate admission; unsupported group types
use the existing full endpoint up front. There is no error-driven retry, fallback
or snapshot fabrication. Current owner/serial guards remain before comparison.

The selected, captured and runtime held/remainder branches remain unchanged. Scope
and evidence details identify aggregate-only baseline acquisition, its actual
per-read receipts/warnings and absence of row-page consistency checks. Source review
has no remaining architecture blockers. `git diff --check` passed. Delivered SQL
budgets, byte counts and timing claims must come from actual source measurements.
