# Continuation task for Codex

Continue this repository into a composable Grammar of Observability. Read AGENTS.md, START_HERE.md, docs/ARCHITECTURE.md, docs/KNOWN_ISSUES.md, both RFCs, and docs/IMPLEMENTATION_PLAN.md. The source currently implements fixed query modes, not the proposed general language.

Work through the following first slice autonomously, resolving routine implementation choices and documenting substantive semantic choices:

1. Establish the baseline. Run the Python tests. On a Docker-capable host run Compose configuration validation, start the stack, run scripts/smoke.py, inspect actual table schemas, and verify the UI. Fix observed integration defects without expanding scope. Do not reset existing volumes without checking that they contain only disposable demo data.
2. Add deterministic fixtures with a frozen evaluation clock. Cover more than 1000 subjects, duplicate events, exact deadline boundaries, late arrival, absent completion, conflicting acceptance/terminal evidence, missing identity/timestamps, cross-tenant IDs, and zero traffic. Establish expected outcomes independently of generated SQL.
3. Implement a small typed expression model and an in-memory reference evaluator for sources, predicates, population selection, grouping, count, and compatible ratio. Add an explicit deadline classification operator or a documented lowering to equivalent primitives. Keep the richer resource model small and versioned.
4. Compile the same normalized expressions into ClickHouse. Fix completion totals so pagination does not affect them, select the correct mature acceptance window, look up terminal evidence outside that acceptance interval, and propagate quality states. Keep the original UI usable through a compatibility mapping while queries migrate.
5. Demonstrate a new declarative measurement without adding a new hardcoded view branch. Run the reference and SQL evaluators against identical fixtures and compare counts, identities, bounds, and evidence states. Persist the validation record and document remaining work.

Prioritize correctness and clear composition rules over a broad operator catalog. Keep alarm previews labelled until a stateful scheduler exists. Do not claim evidence completeness from an unsampled declaration alone. Do not build a full notification system, new collector, large frontend framework, or Kubernetes deployment in this first slice.

If Docker is unavailable, complete the evaluator and fixture work, retain the pending container acceptance tasks, and state the exact environment blocker. The original hosted environment allowed installation of Docker/Compose binaries but blocked daemon mounts, sockets, and rootless namespace setup. This may not apply to your new environment.

Use docs/IMPLEMENTATION_PLAN.md for follow-on stages. Finish with code changes, passing/failed/skipped checks, important design decisions, and a concise continuation note. This file is the task request; RFC examples are proposed syntax that you may refine with a recorded rationale.
