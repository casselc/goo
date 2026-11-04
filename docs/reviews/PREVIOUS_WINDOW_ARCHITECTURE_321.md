# Previous-window count investigation — architecture 321

Status: source review accepted; generated execution recorded below.

The general-purpose outcome is to ask whether an ordinary logs/spans count view
changed against its adjacent preceding equal-duration window, then save/reopen the
same recipe. This alternates from recent native metric/policy depth to a reusable
investigation workflow. It does not require a new chart, source executor or query
algebra.

## Existing ownership and bounded change

`exploration/captured_count.go` already admits independently supplied ordinary count
cohorts with different windows; the existing comparison endpoint and grouped result
retain exact operand questions. `ui/population-workspace.mjs` owns fixed-query
baseline reads and postflight comparison, including the aggregate-only baseline
capability and full selected input. Preserve that owner and its cancellation,
readiness, retained-snapshot and independent-clock contracts.

The current query-count-comparison declaration stores the baselineQuery already.
Extend only its Go/JS relation admission: existing identical windows, or baseline
[from-width,from) for selected [from,to), width=to-from. Reject all other differing
windows: future, overlapping, gapped and unequal durations. Keep page-zero,
checked environment, group/profile admission and the existing independently
validated predicate allowance. Captured comparisons and held operands remain
unchanged. No relative-time execution expression or persistent temporal mode is
needed; the two fixed queries encode the recipe.

The explicit Compare previous window action clones the complete current query,
including typed environment, namespace clauses, other predicates and group/profile.
It stages configuration without telemetry reads and retires old comparison
results. The existing explicit Run independently acquires selected and baseline.
Before-epoch windows refuse rather than clamp/truncate. Existing unsupported
numeric/status/edge/grouped-time profiles remain refused with a reason.

Rebase preserves relation, not an arbitrary fixed offset: same-window pairs remain
aligned; preceding pairs recompute an adjacent baseline of the selected window's
new duration. This handles both resize and shift without stale baseline bounds.
Each query still independently satisfies bounds and bucket budgets. Source query
and source-returned scope must not be silently repaired.

## Presentation, reuse and acceptance

Both selected and baseline UTC windows must be visible with the comparison; read
clocks alone no longer distinguish the operands. Count differences remain captured
record counts rather than health, unique request throughput or causal significance.
Save, portable artifacts, CLI and question opening retain the existing fixed pair;
ordinary presentation Apply/template reuse stays refused because it would rebind
source intent. Do not change current facet/comparison coexistence rules to make
this milestone larger; existing explicit promotion/refusal remains applicable.

The new generated verification lane covers typed policy and namespace variants,
exact initial cloning, resize/shift adjacency, pre-epoch refusal, profile/policy/
group changes, input ownership and Go/JS/artifact admission. Actual browser tests
must exercise staging with zero reads, explicit Run with expected independent
source acquisitions, save/reopen and local return, including a genuinely different
previous-window population. Source budgets and uncertainty remain visible.


## Final source and side-action review

Go declaration validation first validates the existing cohort profile, then admits
only same-window or exact adjacent previous-window operands with page zero. The
JS union follows the same relation. The previous-query builder clones the whole
query, resets inspection offset and refuses a baseline before epoch. Rebase first
validates the old pair and constructs/validates both new questions before returning
any change, preserving independent predicates and all presentation fields.

The UI action uses the existing Count comparison disclosure and its current owner,
generation, draft, busy, selection and pending-read guards. It stages through the
existing root query owner and discards old result authority; it does not fetch or
transfer a held witness. Facet coexistence remains explicitly refused rather than
being silently removed. Visible fixed operand windows now accompany the comparison
before preparation and remain available with group read clocks. Generic labels
claim a fixed baseline, not a logical remainder or causal explanation.

Side drilldowns continue through `stagedGroupQuestion`, which clones
`pair[intent.side].scope.selector` rather than the selected root query. Thus a
baseline group stages the actual baseline time window, predicate and policy; each
side is checked against its own supplied result/basis before opening the existing
child. No new executor, snapshot authority or temporal query algebra was added.

Generated verification initially ran before the Go gate landed and correctly
refused the new pair; it was not treated as a completed implementation failure.
With the gate present, the first completed run passed 30 cases in 3.601 seconds.
The final run, including Go negative window/policy/group mutations alongside
the JS counterexamples, passes all 30 generated cases in 15.474 seconds, zero skips. Browser/source-budget
results are maintained separately in VALIDATION.
