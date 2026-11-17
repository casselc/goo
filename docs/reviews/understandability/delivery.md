# Delivery and verification understandability review

Reviewed 2026-11-16, source inspection only. No implementation or new test campaign
was performed. [Walkthrough](../../walkthroughs/delivery.md) describes current
behavior; findings below are proposals, not delivered simplifications.

## Assessment

The production boundary is clear in code: one Go executable; optional instrumentation
owns no providers; model/browser tooling stays in nested modules. The harder part
is discovering the right development path among historical samples, many commands
and progressively appended validation narratives. An afternoon is sufficient for
the principal flow if readers start from a bounded map, not the full verification
README or validation ledger.

## Prioritized changes

| Priority | Evidence and problem | Bounded change | Acceptance |
| --- | --- | --- | --- |
| High | `Makefile:test` runs only root Go and UI tests. `.github/workflows/api-contract.yml` also tests `instrumentation`, `examples/operation-app` and `examples/report-app`. `CONTRIBUTING.md` warns about nested modules, but the command called `test` still omits shipped sample/instrumentation code. | Provide one explicit fast/default check entry listing these application modules; keep optional native campaigns separate. Reuse the same module list in CI rather than copy shell loops. | A deliberate failing nested-module test fails both local default checks and the corresponding CI entry; no optional tool is silently installed or skipped. |
| High | `scripts/run_journeys.sh` always talks to Redis `reservations`/`inventory`, pauses `traffic` and waits for reservation drain before any requested spec. `journeys/191-generic-runtime.js` and modern fixtures can isolate their own runtime/data independently. | Separate generic journey execution from the reservation lab lifecycle. Keep a named reservation wrapper only for scenarios that need it. | A generic-runtime journey runs without Redis/reservation services; the reservation scenario still pauses/restores only its traffic and retains pre-existing data. |
| Medium | `.github/workflows/api-contract.yml` runs all root Go tests, then reruns wire, embedded-doc and saved-workflow subsets. Some repeats collect unique observations or check generated API drift, while others appear to rerun unchanged coverage. | Name the test tiers and retain repeats only for a changed artifact, observation output or otherwise distinct condition. Measure CI durations before removing checks. | Same rejecting/wire/drift assertions remain covered; explicit observation-producing runs remain; CI reports time saved rather than assuming duplicates are expensive. |
| Medium | `scripts/verify_optional.py` validates strict JSON test events; `journeys/support/cross-app.js` imports equivalent event validation from `journeys/support/go-verdict.js`. Multiple language runners must agree on truncation, skips and missing verdicts. | First document one small verdict contract and feed both implementations the same fixtures. Consider a single native verification entry after consumers are mapped; do not port all tools merely to remove Python. | Both runners reject identical malformed/torn/skipped event streams and accept the same complete runs. No new framework or second product server. |
| Medium | `.go-version`, `Dockerfile`, example Dockerfiles and `.ko.yaml` separately carry compiler/base/tool pins. The API CI checks application modules but does not prove the ko image path or woven builds on every change. | Add a bounded consistency check for compiler/module/tool pins; define an appropriate image/woven smoke cadence. Preserve the useful distinction between generic nonroot image and source-configured lab image. | A stale compiler pin is caught. An actual woven sample and minimal image start without adding tooling to the production image. |
| Medium | `VALIDATION.md` is roughly 25,000 lines and `verification/README.md` roughly 5,000 lines at review; many feature-numbered campaigns have separate actors/specs/tests. `docs/SINGLE_GO_RUNTIME.md` still contains present-tense migration prose saying the Python compiler is live before later implementation notes. | Keep a concise current check matrix and delivery walkthrough; archive the historical narrative without deleting useful models, accepting receipts or regression behavior. Mark old architecture paragraphs as historical at their entry. | A new contributor can find root/unit, nested-module, native, browser and container checks from one page, and cannot mistake the removed Python gateway for current code. |
| Medium | `cmd/goo/weave.go` is explicitly test-only, requires `orchestrion.tool.go`, writes a manifest then runs Orchestrion tests. `scripts/build_operation_app.sh`/`build_report_app.sh` separately scan then build woven binaries. | Give the scan-before-weave contract one documented owner and share only the repeated build validation if it simplifies callers. Do not create a general build DSL. | Invalid directives fail before either test or build; both paths use the module-pinned Orchestrion version and never edit dependency declarations. |
| Low | Native `verification/go.mod` replaces the sibling `observability/quintgo`; `verify_optional.py` checks an exact sibling commit. Gotogether has its own pinned module and strict control runner. This is reliable when prepared, but not a standalone checkout experience. | Keep prerequisites explicit; evaluate packaging the sibling tool as a pinned consumable dependency only when it preserves the same native trace guarantees. | Fresh-checkout setup reports one actionable prerequisite, and required checks cannot turn missing tools into passing skips. |

## What should not be called dead code

- `grammar.py` labels itself an offline historical reservation SQL reference and is
  still imported by `tests/test_grammar.py` and `tests/test_clickhouse.py`. It is not
  in the product startup path. Retirement may be desirable in this prerelease,
  but first identify which reference assertions remain uniquely useful, move
  those to maintained Go scenarios, then delete the reference and its inputs
  together. It is active test code, not proven unreachable product code.
- `sample.py` and the Dockerfile `sample` target supply independent OTLP workloads.
  Removing them solely because they use Python would remove a useful ingestion
  fixture. Prefer clearly named optional lab profiles or one smaller generic
  workload if real coverage is preserved.
- `prebuilt` is a Go package with catalog/HTTP/source-contract tests. Its name alone
  is not evidence that it belongs to the retired reservation UI.
- Annotation manifests describe declarations; instrumentation and observed
  conformance remain separate. Collapsing them would invent deployed coverage.

## Reliability and performance to retain

`instrumentation.Runtime` uses application-owned providers, single binding and
idempotent completion. Its finite outcome initialization is bounded at 1,024
operation keys; observations continue beyond that limit without seeding synthetic
zero baselines. Preserve that behavior when simplifying its bookkeeping.

`retain_diagnostics.py` is dry-run by default and bounds raw recordings by outcome/
age/budget while protecting pins and recent failures. `run_journeys.sh` restores
traffic via a trap. CI's XTDB job checks actual native verdicts and stops its owned
fixture even on failure. These are purposeful reliability boundaries, not merely
incidental scripting to delete. The user does not require indefinite evidence
retention; retain only reproductions/contracts/documentation with a clear owner.

The highest-value simplification is a small, truthful command/check matrix and
module map—not unifying all tests behind a new orchestration platform. No numeric
performance improvement is claimed by this source review.
