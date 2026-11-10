# Go and library review after396

> Root acceptance closure: 396 actual main/actor, property and required observed
> model checks are accepted; see [validation](../../VALIDATION.md). The review
> findings below retain their original read-only context. The adopted queue is
> the [396 roadmap](../ROADMAP_2026_11_10_CHECKPOINT_396.md).

2026-11-10. Read-only review of production source, recent393–396 changes and
prior312 review; no new benchmark, dependency upgrade or production change.
The project already targets Go1.27 and the validation toolchain is Go1.27.3.

## 1. Consolidate declaration admission contracts before adding more display profiles

Priority: next consolidation slice.395 exposed a concrete parity defect: Go's
ordinary string decoding accepted `query:null`/`order:null`, while JavaScript and
the generated API schema refused them. The bounded Timeline decoder now rejects
these values. Similar validation is distributed across `workbench/view_timeline.go`,
`workbench/view_templates.go`, JavaScript declaration modules and manual refinements
in `scripts/build_openapi.py`.

Keep typed, ordinary Go declarations as the public API. Build a small shared
wire-contract vector corpus for each declaration family, executed through actual
Go decoding/validation, JavaScript validation and API validation. Cover omission,
explicit null, malformed strings, unknown keys and supported profile combinations.
Use the existing395 vectors as the starting point. Generate static enum/schema
metadata only where it removes proven duplication; do not introduce a universal
runtime schema interpreter or generic public builder machinery.

Acceptance: one deliberate contract mutation is rejected by every consumer;
existing valid saved-view, reuse and dataset cases remain admitted. This is a
reliability recommendation supported by a reproduced defect, not a performance
claim.

## 2. Make ownership-copy policy explicit at typed library boundaries

Priority: bounded library cleanup, after contract parity.312 fixed caller-aliasing
by freezing whole queries;393 showed that JSON cloning can also repair invalid
UTF-8 before validation. `exploration.Source.fetchProfile` and
`measurements.Question.detached` now validate before their JSON round trip.
`workbench.cloneTemplateView` still uses a JSON round trip, while native stored
view adaptation uses explicit `slices.Clone`/`maps.Clone` plus nested copies.
These mechanisms have different normalization and allocation behavior.

Inventory the small set of public freeze/detach boundaries and state whether each
one preserves invalid-input refusal, optional-field presence and nested ownership.
Retain validate-before-clone. Prefer explicit typed private clone functions where
existing types are small; `maps.Clone` and `slices.Clone` are shallow and require
intentional nested copies. Avoid a generic `Clone[T]` JSON helper: it hides the
exact concern that393 uncovered. Do not replace the currently correct JSON paths
without an allocation benchmark and caller-mutation regression demonstrating a
benefit. No caller-facing type tricks are needed.

Acceptance: mutation during/after a source callback cannot change question,
result basis or saved configuration; malformed literals remain refused before IO.
A benchmark of representative large predicates/stream maps decides whether a
clone implementation change is worthwhile. No speedup has been measured here.

## 3. Keep source mechanics shared; resist collapsing source semantics

Priority: maintenance follow-up, not prerequisite for the next user feature.
`internal/sourcehttp` and `internal/sourcejson` already share bounded HTTP/framing
mechanics. `sourcejson.ReadObjects` uses modern `bytes.Lines`; production code
already uses `slices` and `maps`. Existing393 live evidence established literal
transport parity across record/native adapters. Family-specific byte/row limits,
partial-result refusal, failure classification and custody effects remain useful
explicit boundaries.

If either adapter's literal transport is changed again, extract only the immutable
ClickHouse parameter escaping primitive and common settings construction, retaining
caller-selected limits. Keep the393 original-byte/membership tests as the gate.
There is no evidence supporting a common generic source engine, speculative pools,
concurrent six-query fanout, experimental JSON replacement or additional generic
public APIs. In particular, parallel source reads would change resource pressure
and consistency timing, and requires measured latency/concurrency budgets first.

No blanket Go feature migration is recommended. The useful next work is enforcing
existing contracts and reducing demonstrated duplication, with ordinary functions
and concrete types remaining the application-facing API.
