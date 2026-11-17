# Semantic core understandability review

Reviewed 2026-11-16 at `9119a967`. This is code inspection and focused checks,
not a performance benchmark or proof that every profile is correct. No production
code was changed. [Current walkthrough](../../walkthroughs/semantics.md).

## Assessment

An afternoon is enough to understand the contract boundaries and follow one record
query or native instrument. It is not enough to learn every allowed combination
in `measurements.Definition.Validate` and `Evaluate`. Those cross-products are the
largest local reading cost. The smaller `identity`, `querycel`, `lifecycle` and
`subjectfacts` packages already have useful pure entry points and explicit failures.

Do not combine the core into one universal evaluator. The record predicate,
native interval and prepared-subject aggregate paths carry different authority.
`scopedview` is wire data for supplied display, not another query language.

## Findings and proposed changes

| Priority | Evidence in current code | Bounded proposal / acceptance |
| --- | --- | --- |
| P1: split measurement admission/orchestration by semantic phase | [model.go](../../../measurements/model.go) `Definition.Validate` interleaves kind, grain, resource and domain/profile checks; [source.go](../../../measurements/source.go) `Evaluate` performs profile-dependent decoding and dispatch in one row loop. `source.go` also owns SQL compilation and result preparation. | Keep the public question/result types. Introduce small internal steps for validated profile admission, raw source decoding and kernel dispatch. Compute profile facts once before iterating rows. Preserve raw-field presence checks: some profiles intentionally inspect original bytes, not canonicalized rows. Acceptance: the existing strict decoder, source/profile, exact-bounds, no-partial-evaluation and custody boundary tests remain authoritative. No runtime plugin registry or new public generic type hierarchy. |
| P1: put field capabilities behind one checked Go description | [exploration/model.go](../../../exploration/model.go) `columns` / `Field.Type`, [compiler.go](../../../exploration/compiler.go) `compiler.field`, and [querycel/program.go](../../../querycel/program.go) `declarations` describe overlapping record fields separately. | Consolidate name/type/signal/physical-column metadata for ordinary fields, with explicit handlers for derived fields. Keep the narrower CEL profile as an allowlist, not automatic exposure of all source fields. Add a parity check that every admitted ordinary field lowers with the intended signal/type, and every exposed CEL declaration belongs to that set. Review JavaScript field projection separately rather than silently sharing server SQL syntax with the browser. |
| P2: make repeat validation/copy cost visible before optimizing | [source.go](../../../measurements/source.go) `Source.Run` detaches and compiles; `Evaluate` calls `prepareResult`, which detaches and compiles again; `Compile` validates the question and `Definition.Revision` validates the definition again. [querycel/planning.go](../../../querycel/planning.go) also deliberately validates the base and complete result independently. | Benchmark representative maximum-size definitions/row sets. If material, use a private admitted-question/plan value within one call, while all public entry points retain admission and detached ownership. Do not remove independent owner/binding validation as apparent duplication. Acceptance includes unchanged plans/revisions/bases and mutation tests, plus measured allocation/latency results. No present latency regression is established by this review. |
| P2: finish checked identity boundaries instead of keeping a generic unchecked digest convenient | [measurements/model.go](../../../measurements/model.go) exported `Hash(any)` ignores `json.Marshal` errors. In contrast, [revision.go](../../../measurements/revision.go) and [question_key.go](../../../measurements/question_key.go) return errors and validate their identity family. Production references remain in discovery, lifecycle custody, instrument capabilities, pool signals, replay and result basis. | Audit each remaining identity family and replace the generic entry point with a typed checked helper. Preserve intentional distinctions between authored revision, fixed question, observed result and custody record; do not normalize them all into one hash. Invalid serializable input must return no identity. Existing `Hash` is **not dead code**, and no invalid-current-caller exploit is asserted here. |
| P2: document profile gates in one reader-facing matrix | [identity/environment_selection.go](../../../identity/environment_selection.go) distinguishes current-only, effective and All policies; [measurements/model.go](../../../measurements/model.go), [measurements/environment.go](../../../measurements/environment.go) and [exploration/query_environment.go](../../../exploration/query_environment.go) admit different subsets. | Maintain a compact input/profile/carrier table beside the semantic entry points. Use it for the proposed All-environments span work. Removing `RequireNarrow` globally would broaden unreviewed carriers. This is purposeful source authority, not legacy compatibility to erase blindly. |

## Things to keep simple by keeping them separate

- `metrics` is a pure supplied-observation kernel library; source HTTP does not
  belong there. The existing typed requests and explicit state/reason results are
  easier to audit than an open reducer interface.
- `subjectfacts.Fact` uses `map[string]any`, but `Definition.Evaluate` validates
  declared string/bool fields, evaluation identity and duplicate subject keys
  before population filtering. Replacing it with arbitrary CEL evaluation would
  weaken the compact aggregate contract, not simplify it.
- `lifecycle.Prepared.Evaluate` intentionally validates receipt-wide impairment
  before returning ratios. This is not redundant with validating selected fact rows.
- `querycel` runtime helpers and compiler overload admission are separate checks.
  Adding an evaluator helper does not automatically grant SQL lowering authority.
- Profile-specific host/pod/container/runtime helpers describe concrete OTel
  semantic conventions. Their file count is not evidence that they are unused.

## Dead or unnecessary work: evidence boundary

No production package in this scope was proven unused. `Hash` has active callers;
Databricks metric-view compilation is an implemented portable-output path even
without a Databricks runtime; native and subject-fact compilers operate on different
inputs. Do not delete any of these based on their similarity or age. Candidates
above require focused call-graph and behavior checks before code removal. The
most credible removable *work* is repeated in-call validation/compilation, subject
to measurement and ownership proof.

## Executed checks

`go test -p 2 ./identity ./querycel ./metrics ./subjectfacts ./lifecycle ./scopedview -count=1`
passed on the workspace Go 1.27.3 toolchain. Five packages have passing tests;
`scopedview` reports no test files. Source-backed measurement/exploration/workbench
suites were not rerun by this component review; historical results remain in
`VALIDATION.md`. No claim is made about new end-to-end, warehouse or source checks.
