# Roadmap reconciliation — checkpoint 372

The [design](WHOLE_PRODUCT_DESIGN_372.md) and
[architecture](WHOLE_PRODUCT_ARCHITECTURE_372.md) reviews agree that Goo already
has substantial general exploration, service/resource views and reusable record
compositions. The useful next work joins those capabilities rather than starting
another platform. Finish 372 acceptance before beginning this sequence.

## Selected next sequence

1. **373, breadth: discover native display configurations in context.** Reuse the
   existing definition-library presenter and immutable workbench store. Browse a
   bounded metadata page, Inspect a verified configuration, and Apply to the
   unchanged current question before explicit Run/export selection. Show exact
   unit and compatibility. Keep deterministic ordering honest: IDs are not a
   creation clock. Bound artifact reads and response size; make corrupt or partial
   metadata explicit. No new registry, mutable latest pointer or donor evidence.
2. **374, depth: close exponential-histogram readiness.** Pin one real OTLP and
   exporter/table contract with independent expected facts. Decide whether scale,
   offsets, positive/negative/zero buckets, units, temporality/reset and exact
   boundaries fit the existing model. Implement only a genuinely bounded arm if
   it closes; otherwise publish a precise defer decision. Do not silently convert
   exponential buckets to guessed explicit bounds or create a new evaluator.
3. **375, consolidate: keep group identity beside its action.** Improve the shared
   grouped-record presentation so a user at 390 px can see service/type identity,
   count and its primary continuation together. Cover a representative ordinary
   grouping and typed triage, long/missing/empty values, keyboard focus, held Back
   and current action leases. Measure actual task geometry. Presentation changes
   must not alter full counts, query scope or source-read budgets.

The architecture proposal for a separate native-admission truth-table milestone
is useful acceptance work, but does not need a separate maintenance release.
Fold the finite Go/JS/schema omission/null/empty/Unicode/nanosecond matrix into
373's touched native boundaries. Existing independent numeric oracles remain;
there is no universal validator project.

## Lessons and retired assumptions

371's native reuse pilot is delivered, with exact-ID discovery friction; native
views are no longer wholly absent. 372 closes a real typed route barrier and an
implicit-read mismatch. A checked question alone is insufficient: current producer,
visible action, exact planner intent and destination staging must agree. Pausing
proof must also pause visible controls synchronously, before asynchronous work can
restore an enabled button. Include actual frames alongside I/O assertions.

Keep scope admission separate from evaluation. Go owns numeric membership and
arithmetic; the UI checks the requested wire intent using existing exact helpers.
Back keeps held evidence, while this bounded selection path explicitly requires a
parent Run before selecting again. Do not imply all held action proofs are renewed.

Broker health, exemplars, alert delivery, Databricks lowering and review assignment
retain their own source/identity prerequisites. Go 1.27.3 is already pinned;
modern stdlib use should solve a concrete problem, not create an upgrade queue or
user-facing type machinery. Review again after this finite sequence, or earlier
if exponential histogram readiness exposes a material model boundary.
