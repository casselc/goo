# Repository guidance

## Purpose and scope

Build a Grammar of Observability over existing OTel/ClickStack-shaped data. The same measurement definition should support code-managed resources and interactive exploration. Read START_HERE.md, VALIDATION.md, and docs/KNOWN_ISSUES.md before implementation.

## Working conventions

- Keep implemented behavior and proposed behavior explicit. Update validation evidence after meaningful checks.
- Implement milestones in docs/IMPLEMENTATION_PLAN.md as small reviewable changes. Finish the current milestone's acceptance criteria before adding unrelated features.
- Preserve the runnable Python lab while evolving the semantic core. Record a rationale before a language/runtime migration; JSON is current transport and EDN is a proposed authoring option.
- Use source adapters for physical columns. Do not scatter new map-key conventions through presentation code.
- Prefer pure AST validation and evaluation functions with a fixed evaluation time. Use typed query parameters for values and allowlisted or validated identifiers.
- Presentation limits must not define measurement totals. Distinguish exact relationships, declared business identity, and time-based contextual search.
- Missing telemetry is not automatically healthy or failed. Preserve pending, no-data, invalid, and unknown states.
- Native metric streams retain temporality, reset, unit, and histogram semantics. Never average scalar percentiles into a fleet percentile.
- Keep collector ingestion and app telemetry interoperable with standard OTLP. Keep trace links across async boundaries.
- Reuse the framework-neutral display elements in ui/components.mjs and the display-only comparison module before adding local metric widgets. See docs/UI_COMPONENTS.md and the proposed adoption contracts in docs/UI_COMPONENT_AUDIT.md. Keep source reads, navigation, query scope and evidence state in application adapters; custom-element lifecycle/disclosure must not subscribe, fetch or publish.
- Keep reusable component styles separate from page layout. Scope application table/viewport rules to their owning children and verify both standalone and nested composition. Server HTML attributes and detached SPA property updates must remain supported.
- Keep architecture/process diagrams, README and design/planning docs, executable storyboards, and captured screenshots synchronized with each feature. Regenerate affected frames from the running app and record actual validation; label planned behavior explicitly.
- As functionality develops, maintain real user journey storyboards and executable browser tests with captured documentation frames. Mark proposed/unimplemented steps explicitly; do not substitute mocked UI data for the live integration journey.
- Use hegel-go for semantic properties and state/swarm tests where appropriate. Keep native verification dependencies in the separate verification module. Use the sibling observability/quintgo tools for model-based replay and observed-trace validation; include negative controls that prove incorrect traces are rejected.
- Preserve business-event time, immutable custody time and frozen evaluation time separately. Historical corrections advance evidence basis while retaining the original question and receipt.
- Retain local loopback port bindings and disposable credentials in this development lab. Keep secrets out of packages and source.

## Commands

Use `docker compose config --quiet`, then `docker compose up --build -d` and `python3 scripts/smoke.py` on a capable host. Install requirements-dev.txt in a virtual environment for `python -m unittest discover -s tests -v` and `python tests/local_flow.py`.

Embedded ClickHouse checks complement the real container smoke test. Report skipped checks as skipped. An unavailable Docker daemon is an environment limitation, not evidence the Compose stack passed. Inspect browser behavior after UI changes when a browser is available.

The Go lifecycle/aggregate pilots have independent local checks documented in
docs/LIFECYCLE_PREPARATION.md and verification/README.md. Run `npm ci` before
`npm run test:journeys`; the wrapper pauses and restores only this demo's traffic
generator and retains existing stream entries/volumes. Browser screenshots under
docs/journeys/screenshots are generated evidence, not hand-edited mockups.

- Commit validated milestones promptly; do not let completed project work accumulate untracked. Review staged files, keep generated caches and secrets excluded, and report the commit ID.

## Deliverables

For each milestone, report concrete changes, checks, unresolved issues, and the next task. Keep README.md, VALIDATION.md, and the issue list consistent with code. Update RFC status when decisions change; do not rewrite the historical design snapshot to imply earlier implementation.
