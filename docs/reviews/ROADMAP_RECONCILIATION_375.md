# Roadmap reconciliation — checkpoint 375

The [architecture](WHOLE_PRODUCT_ARCHITECTURE_375.md) and
[design](WHOLE_PRODUCT_DESIGN_375.md) reviews propose the same concrete sequence.
373 made native presentation reusable and discoverable; 374 proved a source and
representation gap instead of publishing a misleading executor; 375 unifies
read-and-act grouped-row presentation through existing components.

## Adopted next sequence

1. **376 breadth: show observed unsupported exponential instruments.** Current
   discovery omits this table entirely despite 374's real persisted points. Add
   bounded metadata presence with an explicit unsupported reason and no executable
   definition. Preserve environment policy, namespace and fixed window. Missing
   optional tables must have explicit availability semantics without breaking
   supported discovery. Count exported data points, not observations or health.
   Inspect metadata stays distinct from Preview/Run and supports ordinary OTel
   services without a manifest.
2. **377 depth: pin explicit-histogram exemplar readiness, then one exact reported
   trace continuation if it closes.** Verify actual exporter fields, trace/span
   IDs, exemplar timestamp/value and identity before UI work. Follow only an
   actual reported reference with checked scope and explicit acquisition. It is
   not a proof of contribution, causal attribution or complete trace capture.
   If fidelity cannot close, record the specific defer decision promptly; do not
   infer links from temporal proximity or reopen exponential evaluation.
3. **378 consolidate: native Measurements failure/recovery parity.** Reuse current
   source-refusal presentation and typed failure/protocol helpers instead of
   flattening native errors to one string. Preserve the exact current fixed
   question for explicit retry, retirement on owner changes and honest unknown
   state. Separate validation errors, refused source reads and cancellation;
   retain safe optional detail. No rolling-window retry, automatic I/O or new
   backend failure platform.

Native-view compatibility work already closed in 373 is not another milestone.
Messaging-role reuse and broker-health research remain conditional candidates;
current source evidence supports the sequence above more directly. Review after
this finite sequence or earlier if exemplar readiness reveals another material
source/model boundary.

## Lessons incorporated

Inspect actual presenter routing before claiming shared coverage. The ordinary
multi-field result used ScopedBreakdown, while triage actions used the fallback;
matched live evidence exposed both overflows. One passive chart-table opt-in
closed the ordinary path without changing display grammar, source reads or
selection authority. Preserve native DOM and derive labels from actual headers.

Keep metadata presence, executable capability and source fidelity separate.
374's missing zero threshold and optional presence cannot become defaults, and
positive-scale irrational bounds cannot become exact binary64 aliases. Reuse
current owners and data models only after the new contract closes.

Use matched original data and explicit task geometry for usability claims, and
keep raw membership, clocks and source budgets independent. Browser AX evidence
is bounded evidence, not universal screen-reader certification. Go 1.27.3 remains
current in this workspace; choose newer stdlib/runtime features when a concrete
implementation benefits rather than creating a version or type-machinery project.
