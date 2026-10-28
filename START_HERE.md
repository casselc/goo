# Grammar of Observability Codex handoff

> Historical imported handoff. Current implemented behavior and startup are in
> [README.md](README.md), with actual checks in [VALIDATION.md](VALIDATION.md).
> [PROJECT_DIRECTION.md](PROJECT_DIRECTION.md) records the accepted Go/Rust and
> YAML+CEL/KCL direction. [Shared measurement packages](docs/MEASUREMENT_PACKAGES.md)
> now provide live typed gauges, counters, histograms and lifecycle investigations.
> The original design and continuation assignment below are background, not the
> current implementation status or a new user request.

This project explores a composable grammar for observability inspired by the Grammar of Graphics. Its first backend is ClickHouse containing ClickStack-shaped logs, metrics, and traces exported from OTel collectors. Its intended users run web services, internal gRPC services, and event consumers.

The intended product lets application teams keep metrics, SLIs, SLOs, alarms, default dashboards, and reusable views beside their code. Interactive investigations use the same semantic definitions. A UI presents and edits valid expressions while keeping arbitrary available attributes discoverable.

## What you have

Goo runs as one Go server with embedded UI/docs and typed API handlers. The local
Compose lab provides independent sample applications and an OTel/ClickHouse
pipeline. Python is used only by sample workloads and offline tooling/reference
tests. See [single runtime](docs/SINGLE_GO_RUNTIME.md), [current roadmap](docs/ROADMAP_2026_10_28.md)
and [generic startup](docs/GENERIC_STARTUP.md).

The sample path is HTTP → gRPC → Redis Streams → consumer, with OTel collection into baseline exporter tables. The default lab uses its own query interface. HyperDX can be added as an optional profile. The repository includes CI and executable browser/model journeys; production deployment hardening remains separate.

## Read in this order

1. README.md for startup and the actual prototype behavior.
2. VALIDATION.md and docs/KNOWN_ISSUES.md for evidence and gaps.
3. docs/ARCHITECTURE.md for code boundaries.
4. docs/rfcs/0001-observability-language.md and 0002-measurement-semantics.md for the target model.
5. docs/IMPLEMENTATION_PLAN.md for ordered work.
6. CODEX_HANDOFF.md for the first implementation assignment.

docs/design-draft.md preserves the earlier worked design with its larger EDN package. It is background, not a claim that the resources exist in code. The RFCs and issue list identify the differences that matter for implementation. Resolve contradictions by examining the source and making a documented design choice; do not silently treat illustrative syntax as a working API.

## Immediate outcome

The next milestone is a verified Compose run and a reliable measurement core. The first language milestone is one measurement executed by both an in-memory reference evaluator and ClickHouse, with matching results for deterministic inputs. Publishing a new measurement should require declarative data rather than another view-name branch in grammar.py.

Open the folder in Codex and use CODEX_HANDOFF.md as the continuation prompt. A Docker-capable local environment is expected. If unavailable, continue the pure evaluator and fixture work and record the blocked full-stack checks.
