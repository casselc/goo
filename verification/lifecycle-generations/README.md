# Scoped lifecycle generations

An independent, bounded extension of `../model/lifecycle.qnt`: two tenants share
subject and immutable event IDs, and two preparation generations answer the same
question at T=100 with custody cutoffs 96 and 100. The cohort stays [57,87).
The original snapshot remains immutable. Late evidence can turn missing completion
into good, or good into unknown when immutable terminal payloads conflict.

The Quint contract specifies classifications, impairment, counts, ratio state,
generation selection and refusals independently of the Go implementation. The
adapter calls real `lifecycle.Policy.Prepare`, `ValidateCorrection`,
`Prepared.Evaluate` and `subjectfacts.Definition.Evaluate`. It records observations
after each operation; it does not substitute model answers for kernel results.

## Run

Use Go 1.27.2, Node 22, Quint 0.33.0 (Rust evaluator 0.7.0), and Java 21.
Install Quint and its evaluator with the pinned toolkit helper:

```sh
cd verification/lifecycle-generations
go mod download
core_dir="$(go list -m -f '{{.Dir}}' github.com/chucklehead-dev/gotogether)"
QUINT_VERSION=0.33.0 bash "$core_dir/ci/install-quint.sh"
GENERATIONS_OUT=/tmp/lifecycle-generations ./check.sh
```

Quint downloads Apalache on first `verify`; the recorded run used Apalache 0.62.1.
`GOTOOLCHAIN=go1.27.2` can select the Go toolchain explicitly. This standalone module
pins gotogether/qhegel at `1cb37c492401` and Hegel 0.9.27. It changes neither the
root dependencies nor the older verification module's sibling replacements.
Root package tests also need `examples/lifecycle` and `examples/portable` in sparse
checkouts. Tests require their tools when `GOTOGETHER_STRICT=1`; missing tools must
not silently turn CI green.

## What is checked

- Tenant identity survives identical subject/event IDs, including redelivery.
- Original and corrected snapshots retain separate classifications, receipt
  identities, clocks and cohort bounds. Replaying the original cutoff over newly
  appended evidence preserves its subjects, facts, visible-evidence digest and
  snapshot digest. The input-manifest digest may change.
- Low-level aggregation explicitly selects one evaluation from coexisting facts.
  `Prepared.Evaluate` instead refuses foreign-generation facts. These are distinct
  API contracts.
- Corrections preserve the question and scope identities. A changed tenant scope
  refuses before preparation.
- Completed traces carry an explicit settle marker after `finish`. Generated
  prefixes are checked for conformance without asserting eventual completion.
  A deliberately premature settle fails the completion obligation.

Three source overlays remove tenant grouping, the prepared-artifact mixing guard,
and correction scope validation. Original-source SHA256 guards prevent stale
whole-file overlays from hiding future source changes. Review the contract and
regenerate overlays when those files change; do not merely update the hashes.
Hegel shrinks real observed violations within a generator varying duplicate noise
and selected generation. The mixing mutant proves missing refusal, not actual
numerical double counting: the lower-level evaluator still filters generations.

Observation corruption checks distinguish semantic kills from malformed logs:
deleting a required field is refused before a verdict and earns no kill credit.

## Recorded result (2026-10-29)

See [evidence/result.json](evidence/result.json), model gate output and the three
smallest failing observation traces. The faithful baseline is Goo
`d102aade4dec645740d5000e1d2cb35eef3b1fa3`. The full suite passed 12 generated traces / 300 steps (139 non-idle), five
directed scenarios, and source mutants shrunk to 2/6/3 steps. Eight valid
observation corruptions were killed; four malformed logs were refused. Targeted
faithful scenarios also passed under Go race instrumentation. No production defect was found in this
bounded exercise. The suite demonstrates an additional consumer of model-driven
traces and source mutation/shrinking beyond the earlier cutoff pilot.

This is a pure synchronous kernel model with synthetic complete coverage. It does
not establish storage isolation, concurrent delivery, real telemetry completeness,
or lifecycle liveness under an external scheduler. Time boundaries are fixed here;
existing lifecycle tests/model retain responsibility for the wider cutoff matrix.
The stored aggregate is the last explicit selection result; the selected snapshot
view can change independently during correction.
