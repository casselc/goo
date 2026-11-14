# Check a condition without querying telemetry

This is an executed Go/HTTP planning contract, not a new visible editor.

1. Start with an ordinary fixed logs or spans question, including filters that
   the CEL vocabulary cannot print, such as an attribute condition.
2. Prepare an additional condition and binding to the original base, slot, owner
   epoch and draft revision. Run `goo plan-cel`, or submit that request to the
   in-process `/api/cel/plan` endpoint.
3. Inspect the returned predicate, complete refined question and parameterized
   target statements. The original parent, environment, clocks and page policy
   remain intact. Checking performs no source read or storage write.
4. Submit malformed, ill-typed or unsupported text. Inspect its typed diagnostic
   and Unicode position. No partial question is returned, including for an
   unsupported dead branch.
5. Change a base or contract field while retaining its old binding. Planning
   refuses. The future editor must also retire responses when its local owner
   changes; that interaction is not implemented in this milestone.

The shared Go planner, CLI tests, HTTP boundary tests and generated native
properties execute these cases. There are no new UI screenshots because the
visible Activity controls have not changed. Existing held-return screenshots
remain in [436.1](refinement-consolidation.md).

See [contract](../CEL_CHECKED_PLANNING.md) and [validation](../../VALIDATION.md)
for actual checks and limitations.
