# Review reconciliation — 408

Inputs: [architecture/Go](ARCHITECTURE_408.md), [design/usability](DESIGN_408.md),
[product](PRODUCT_408.md), the [current capability guide](../CAPABILITY_GUIDE.md)
and the completed 406–408 cadence. The accepted next queue is
[409 → 410 → 411](../ROADMAP_2026_11_11_CHECKPOINT_408.md).

## Adopt

- Start with Services → existing preceding-window **captured-record** comparison.
  This is a general observability continuation, not another dashboard surface.
  Existing widgets may carry span-status measures excluded from count-comparison
  admission; use an explicit checked count projection, preserving population scope.
- Deepen current OTel incoming HTTP context with literal response-code groups.
  Reuse the outgoing context pattern while keeping response status, span Error
  and inferred health distinct.
- Consolidate shared quantitative chrome and bounded asynchronous handoffs.
  Actual 407 work caught wrong resolver dispatch, pending-button completion and
  retained-role redraw issues. Local presentation and original-read source
  authority must remain separate; “one generic view manager” is not the answer.

This prioritizes the current → scope → compare loop over the annotation/training
loop for the next breadth/depth pair. The latter is still desired and its exact
review-task ID entry is a real discovery gap, not completed functionality.

## Keep explicit, but defer

Architecture and design prefer a review-task chooser and captured review
navigation; architecture also proposes fixed-summary rate comparisons. These
remain finite, useful candidates, but do not displace the selected ordinary
service workflow or broaden 410 into a general outcome/SLI interpreter.

Verification is now roughly 92,000 lines in one nested test package, with private
numbered helper dependencies. This is a concrete maintenance/compile-cost concern,
not evidence that production Go needs a rewrite. Stop adding new numbered helper
coupling now. Plan a separately bounded semantic-family/support extraction with
measured compile behavior; do not bury that migration inside UI polish or add a
new large campaign for every local-control change. Existing independent oracles
must survive any extraction.

407 synthetic positive-aggregate/empty-page fixtures now correctly retain unknown
coverage. Tests must reflect source coherence; never manufacture completeness to
make examples or imports pass. 408's fixtures are explicitly supplied synthetic
data and claim no source observation or completeness witness.

Broader widget inventories, native exponential execution, grouped percentiles,
platform scope and arbitrary executable customization remain deferred. None of
this checkpoint establishes vendor parity, authenticated collaboration or a new
performance claim.

The stale XTDB setup command identified by the architecture review is corrected
to the current single Go service in this checkpoint. The preserved lab is unchanged.
