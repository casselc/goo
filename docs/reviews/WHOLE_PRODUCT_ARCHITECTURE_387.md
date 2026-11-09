# Whole-product architecture checkpoint — 387

Read-only source review on November 9, 2026, while 387 runtime acceptance is
pending. This is input to one revised roadmap, not a second implementation queue.
No runtime, heavy tests or production changes were performed by this reviewer.

## Keep the product loop and existing layers

The core loop is current telemetry → scoped population → explicit follow-through
→ reusable intent or retained evidence. Discovery does not require Goo manifests.
The capability guide already documents comparisons, explanations, review labels,
log patterns, service/resource observations, display grammar and live Activity.
Vendor comparison should refine that loop rather than recreate these categories.

Retain the current layers: physical source adapters; pure typed query/measurement
planning; family evaluation; checked datasets and supplied-data displays;
application-owned reads/navigation; immutable definitions and captures. A new
DataFusion-style executor, universal family registry, generalized event bus or
second navigation framework is not justified by this review.

385 demonstrates a good breadth pattern: a small explicit HTTP-client lens emits
ordinary record questions, with exact Client/current-method/resource predicates.
It does not require another query engine. 386 demonstrates the corresponding
reuse boundary: one NativeView owner with closed gauge/histogram declarations,
while sample and distribution semantics remain separate. 387 should complete the
finite capture envelope without erasing its six family validators.

## Highest-value structural opportunities

### 1. Preserve capability separation at every reusable boundary

`exploration/capture_envelope.go` now separates family shape from evidence
semantics. `ui/saved-capture-protocol.mjs` separates full pages from receipt-only
document references. Keep that distinction when adopting more display kinds:
metadata can offer a candidate, but cannot become loaded evidence, source
membership, execution permission or compatible destination authority.

`ui/source-snapshots.mjs` is the right boundary for admitted supplied record data.
`exploration/view_dataset.go` preserves the selector; it should not gain UI
navigation or source execution. Native observations should retain their own
closed scope/selection semantics rather than being forced into record or
histogram intervals. Full UI expressibility should grow by admitted finite kinds,
not by weakening one universal dataset into an untyped bag.

Before adding another family, audit its path through question → response → dataset
→ display → save/template → capture. This is a checklist applied to the new slice,
not another runtime abstraction. Keep metadata-only summaries explicitly labelled.

### 2. Consolidate display geometry and chrome, not application authority

386 exposed a correctness issue even with a valid model: unavailable marks at the
numeric maximum implied magnitude. The fixed labelled rail lives in the shared
value timeline, preserving exact X positions and selection without inventing Y.
Other scales must similarly keep absent/unknown, zero and stale values distinct.
A physically misleading chart is a correctness blocker, not optional polish.

`ui/product.css` still applies broad `section,.card` rules; reusable components
need explicit chrome boundaries, as existing comments in `ui/components.css`
acknowledge. Avoid more local overrides that depend on incidental ancestor depth.
A bounded consolidation cut should identify reusable component roots, scope their
owned children and check standalone plus nested composition. Do not rewrite the
whole stylesheet or switch frontend frameworks.

The 390-pixel native view still spends substantial space on stream identity,
diagnostics and save controls before observations. Consolidate that surface into
a compact identity summary and one disclosure for full provenance/configuration,
while leaving unknown/stale/error and acquisition state visible. Save controls
should remain available without becoming the principal content. Measure natural
initial geometry in the actual scroll container, rather than claiming success
from a forcibly scrolled screenshot.

Keep application ownership of reads, scopes, generation leases and return visits.
`ui/native-view-reuse.mjs` is one such owner, while `ui/measurements.mjs` projects
supplied samples. Local kind switching must remain source-free. Do not extract a
universal owner merely because several modules contain generation counters:
retained return, streaming subscriptions and immutable paging have different
lifetimes. Extract only an actually repeated bounded operation with matching
retirement behavior and regression coverage.

### 3. Finish source mechanics consolidation only where policies agree

`internal/sourcehttp`, `internal/sourcejson` and `internal/sourcefailure` already
share transport, bounded JSONEachRow framing and failure classes. Keep source
limits and mathematical evaluation in their families. Exploration uses a 1,002-row
HTTP bound while measurements use 10,001 and reject partial measurements; blindly
merging those policies would change results.

A concrete remaining duplication is ClickHouse parameter escaping:
`exploration/source.go` handles every ASCII control and backslash, whereas
`measurements/source.go` uses a smaller replacement set. A future bounded cut can
share one tested Escaped-format encoder, preserving literal Unicode and control
strings. That utility need not own URL policy, compilation or family errors.
Assess current differences before replacing either implementation.

Capture loading has a 24 MiB artifact bound; catalog search has a 16 MiB scan
budget. 387's RawMessage-only envelope handling is preferable to decoding a
second complete recursive graph. Avoid adding cross-family canonicalization or
repeated semantic recomputation merely to reduce a few validation branches.
Profile a real large capture before changing buffering or introducing pooling;
no performance improvement is claimed from source inspection alone.

## Reliability and operational lessons

Preserve retained application telemetry and captures during disk cleanup. The
known-issues record distinguishes bounded lab diagnostic cleanup from a production
retention guarantee. Future reproducibility work should make disposable logs,
build caches and retained user artifacts distinguishable, then bound diagnostic
growth. No automatic deletion of evidence, store migration or new retention
platform is authorized by this review. Disk pressure can prevent validation and
must not be mistaken for a product semantics failure.

Observed completion must mean the visible outcome, not a clicked action or a
model phase alone. 385 rejected a premature actor completion and added a visible
Services-return gate. Continue using the existing Playwright/Bombadil and
Hegel/Quint tools; do not add another framework. Keep exact boundaries in reports:
a browser pass, source ledger pass, generated admission check and image inspection
prove different things. Minimize retained artifacts to what documents behavior or
supports useful future verification.

## Modern Go without clever caller APIs

The project already targets Go 1.27 and the managed toolchain is 1.27.3. Existing
`maps.Clone`, `slices.Clone`, typed request structs and `reflect.TypeFor` use are
appropriate. New small helpers should accept explicit typed policies and return
ordinary errors; advanced type machinery would not improve these finite unions.
Use current standard-library collection, context and HTTP facilities when they
simplify a touched path. Do not create a runtime-upgrade milestone or speculative
stdlib rewrite: no missing Go feature blocks the next user outcome.

## Reconciled next finite cadence

The parallel product review proposes a more direct core-observability sequence.
Source inspection supports it, subject to 387 acceptance:

1. **388 breadth — logs/spans around a selected gauge observation.** This is a
   real gap: `recordContextPlans` explicitly suppresses gauges, the underlying
   Go/JavaScript planner admits only a one-nanosecond gauge point, and the held
   owner admits only counter/histogram intervals. Add a distinct bounded
   observation-context projection using a documented fixed radius clipped to the
   original exact source window. Preserve service, instance, namespace and
   environment authority; do not turn metric attributes into record filters or
   infer causality. Label time association explicitly. The held path must verify
   the selected sample really belongs to the retained exact stream. Stage an
   ordinary typed question without a source read; Run acquires records. Do not
   globally relax histogram/counter interval validation or broaden resource
   bindings in this cut.
2. **389 depth — native counter display reuse.** Counter evaluation and interval
   rate presentation exist, but NativeView only admits gauge/histogram arms. Add
   a closed timeline/table counter arm using evaluated intervals/rates, retaining
   resets, unknown states, declared decimal domain and exact unit semantics.
   Validate the supported destination policy and keep the entire destination
   question authoritative; donor freshness/reset configuration and selected
   intervals must not transfer. Do not coerce counters into gauge samples, infer
   rates from displayed totals or create another native-view owner.
3. **390 consolidation — data-first native entry and component boundaries.**
   Compact stream identity/diagnostics/configuration; scope reusable primitive
   chrome against global CSS; reconcile existing return affordances where a
   single bounded adjustment is justified. Preserve real retained-return
   authority and no-read behavior. Validate natural initial geometry at wide and
   390-pixel widths, not merely scrolled screenshots.

This supersedes this review's initial contextual-review-task/kept-occurrence
proposal. That annotation pair remains valid deferred work, not a prerequisite to
this core queue. The sequence reuses typed record questions, native evaluation,
existing displays, NativeView storage and source/session owners. It adds no
platform, profile engine, arbitrary extension registry or testing framework.
Log-pattern grammar admission and native HTTP-client duration remain later finite
alternatives; existing histogram comparisons should not be recreated.
