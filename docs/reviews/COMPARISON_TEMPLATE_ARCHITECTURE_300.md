# Explicit destination baselines in reusable comparison templates — 300

2026-11-01. Architecture contract and source review for milestone 300.
Executed browser/model outcomes belong in VALIDATION.md; the acceptance below is
not itself an execution claim.

## Decision and typed contract

Reuse the existing authored-template, captured-count comparison, baseline picker
and preparation owners. A comparison template declares
`template.baseline: "destination"`. That field is required exactly when the
composition contains a comparison, including hidden comparisons; it is forbidden
on templates without that role. The source composition's immutable capture ID is
a validation specimen, never a reusable default.

The authored expansion arm gains an explicit `baselineCapture` input. It is
required exactly when the source contains a comparison and must name one valid
immutable capture ID. Omitted, null, empty and malformed values refuse; an input
on a noncomparison template or the shipped `{template, version, widget}` arm also
refuses. No previous value, source ID or “latest” is a fallback. Choosing a
source-equal ID is permitted only as the explicit input and remains unverified.

Go and JavaScript must agree on wire presence and role-dependent admission.
For an optional Go string field, `omitempty` alone is insufficient: a custom wire
check must distinguish an absent field from explicitly supplied empty/null data.
Update the existing generated Go/OpenAPI schema and strict request-arm tests,
rather than rewriting field names in serialized JSON.

## Pure expansion and exact response checks

Validate the immutable source and its content digest before replacing anything.
Validate the destination query through the existing ordinary template and
captured-count eligibility functions. Clone the composition, replace only its
comparison `baselineCapture` with the explicit input, and validate that detached
composition against the destination. Preserve all other child order, hidden flags,
kind, title, display/view/order, record layout and completeness settings.

Do not call ordinary source/destination captured-comparison reuse compatibility:
that would require the specimen's grouping/filter root and defeat the authored
template's declared destination-group binding. The target grouping may differ
from the specimen. Its chosen actual baseline must match the target root later.
Optional timelines retain 299's explicit equal positive source/destination width,
measure support and interval-count admission, even when hidden.

The comparison remains count-only by its existing display contract. Preserve the
existing eligibility predicate: it refuses numeric, status, split-time and
reported-edge profiles. It does not independently forbid a supported
`summary: "span-duration"` directive; a count comparison and an existing duration
timeline may coexist when all current validators pass. Prepare must then also
match the requested summary directive exactly.

Expansion performs no telemetry read, capture fetch or publication. Its syntactic
reference is **unverified**. Response validation constructs the exact expected
composition using the chosen replacement ID, then compares the complete remaining
presentation. Never weaken whole-composition equality into a loose kind check or
accept the source specimen ID merely because it has the same syntax.

## Application ownership and preparation

The application obtains the destination baseline from its current existing
comparison picker/configuration, not from template source metadata. Preview names
the selected destination reference and says that Prepare still verifies it.
Opening the picker, configuring its ID, previewing and applying presentation do
not fetch a capture or query telemetry. Preserve documented existing Run effects;
a visible comparison can be prepared by the current Run path.

Changing the configured baseline retires prepared pairs and template previews.
Capture the exact chosen ID in each expansion request/response owner and recheck
it on Apply; do not resolve a mutable global after an asynchronous reply. Existing
query/draft/generation/serial guards remain necessary. An old preview must not
apply after a different picker selection or after that selection returns to an
older ID. Native add/remove comparison gestures should keep an already-enabled
template's binding marker consistent without silently changing its data profile.
Imported/source declarations still undergo the strict explicit binding checks.

Existing explicit Prepare loads the chosen stored evidence, verifies receipt ID
and admitted content, checks captured/current query compatibility and performs
the canonical count comparison. Environment policy, service/filters, grouping,
interval width and requested measures must match the **destination**; only fixed
windows and documented inspection paging may differ. Invalid, unavailable or
wrong-root references display unavailable/unverified state. No syntactic input
confers a coverage or completeness witness.

## Acceptance and negative controls

Author a real service-A comparison through the existing UI, enable its reusable
template and Save. In service B, change grouping, choose a separately captured B
baseline, inspect/apply the template and Prepare. Verify the output and saved/reopened
B view retain exactly that explicit reference, and that no A capture is fetched or
retained through implicit copying. Apply itself must remain source-free.

Exercise missing/malformed/empty/null input, wrong request arms, baseline marker
without a comparison, missing marker with a hidden comparison, absent and
mismatched timeline widths, unsupported status/numeric profiles, unavailable and
valid-ID wrong-root captures, changed destination baseline while preview is in
flight, retired Apply handles and a second expansion with no new baseline input.
Prove generated Go/JS parity, immutable/detached output, observed lifecycle and
external-driver completion rather than treating a click as a successful comparison.

## Implementation source review

Reviewed the Go expansion/binding changes and JavaScript handoff/application
changes after implementation. No source architecture blocker was found.

- Go validates the original source/content digest, validates input presence, clones
  the source, substitutes only the comparison capture ID and validates the complete
  result against the destination. The source object is not mutated. Existing
  comparison eligibility and timeline width/measure/budget validation remain in
  their existing owners. No capture client or source reader was added to expansion.
- Binding JSON admission rejects explicit null/empty/unsupported baseline values;
  request JSON distinguishes missing input from present malformed/null input.
  Shipped/noncomparison expansion cannot accept a stray baseline input.
- The browser reconstructs the exact expected composition with the chosen ID and
  retains equality checks on every other presentation field. It captures that ID
  in the request and inspected response; currentness includes the configured ID,
  draft revision and request serial. Native composition edits derive only the
  conditional baseline binding while preserving the selected profile.
- Missing-input recovery reveals the existing count editor or composition control
  and performs no source/capture read. Preview distinguishes presentation
  compatibility from an unverified baseline. Actual evidence admission remains
  the existing Prepare path, including receipt identity and target-root checks.

The response's template reference remains the source **view-definition** digest.
That metadata is expected and is not a transfer of the source **capture** ID into
the destination comparison. Browser acceptance should distinguish those two IDs
rather than forbid every source identifier in the expansion response.

Final schema admission, generated semantic properties and real browser/late-reply
outcomes are recorded by their respective test lanes and in VALIDATION.md.

## Proposed next 301 — supplied-span waterfall

After timeline and comparison reuse, one bounded supplied-span waterfall is a
more useful next grammar outcome than another template-profile extension. Add a
spans-only records presentation over the existing admitted `goo.record-page/1`.
It uses the existing record interaction owner and geometry/renderer; it adds no
query, acquisition role, store, trace-completeness guarantee or stable event ID.
Name it as supplied spans, not a complete trace. Explicit Expand trace remains the
independent cross-service context read.

Keep the same query, page/token/basis, configured fields and outer pager through
Table/List/waterfall changes. Save, authored template application and reopen retain
the pinned kind. Logs refuse it. Pending record workflows retain the existing
replacement guards; old mounts stay retired. The public component remains passive,
including serialized/detached use, until an application attaches its own controls.

Implementation hazards to resolve as part of that outcome:

- `waterfall.layout` currently builds a single-value map of trace/span IDs and
  therefore overwrites duplicate parent candidates. A parent edge requires valid
  same-trace IDs and exactly one candidate occurrence. Duplicate/ambiguous,
  missing, invalid/unreported and cyclic parent cases must stay explicit; never
  select a convenient candidate or join unknown trace IDs into an implied trace.
- Parent-first visual order differs from page order. Host handles must remain
  keyed by original page position; selection, Inspect and Keep must use that
  occurrence index, including two equal/repeated objects. The shared controller
  must not substitute visual row ordinals or `rows.indexOf(record)`.
- Existing waterfall rows are native buttons. Do not place selection checkboxes
  or action buttons inside those buttons. Provide adjacent passive action/detail
  mounts, attached by the same application owner, and preserve exact render leases.
- Current geometry uses one global earliest/latest extent across all traces. A
  service page with many short unrelated traces across minutes can become
  unreadable. Prefer trace-relative group origins with one common duration scale
  across groups, explicitly labelled as relative to earliest **supplied** span,
  plus exact absolute UTC tooltips. Independent per-trace autoscaling would make
  lengths incomparable. Settle this fixed kind meaning with design before coding.
- Reuse exact nanosecond subtraction before approximate coordinates, human duration
  formatting and raw/exact tooltips. Missing or invalid geometry must be visible
  as unavailable, never coerced to zero or silently removed. Zero-duration spans
  remain inspectable. A missing parent does not prove a missing upstream event.
- Preserve all admitted configured fields and missing/empty/null distinctions.
  The current waterfall's four-fact cap cannot silently truncate the records
  contract's eight qualified fields. Use existing field formatting and keep
  component CSS separate from application layout.

Acceptance should use real multi-service OTel spans with multiple traces,
out-of-order rows, duplicate parent IDs, missing parents and cycles. Verify local
kind changes read nothing, visually reordered selection Keeps the exact original
page occurrence, pagination preserves held aggregate/timeline bases, template
save/apply/reopen retains kind, and explicit Expand trace independently acquires
another service. Use generated projection/index/ancestry properties, current-owner
negative controls and a completed browser/Bombadil journey. This is a proposed
next outcome, not functionality delivered in 300.

## Native capture/return integration review

The native 300 journey exposed an existing retained-visit boundary: leaving for a
capture retired the population generation, while returning displayed enabled
metadata controls whose old callbacks still belonged to that retired generation.
A forced population Run would hide the defect by rebuilding controls and rereading
telemetry, so it is not the intended recovery path for choosing a captured baseline.

Reviewed the bounded implementation in Population, Exploration and Investigation
Trail. No source architecture blocker remains:

- Suspension retires pending page/comparison/template owners. Cancelled page flags
  are cleared while their page serial advances, retaining only the last completed
  held page; no interrupted operation is retried automatically.
- Exploration delegates a retained-return operation to Population. Trail invokes
  it after restoring saved custom-element children, so that restoration cannot
  overwrite newly mounted controls. The workspace fallback is queued and the
  operation is idempotent for a completed return.
- Return requires the active connected visit, a visible owning workspace/panel and
  no pending acquisition. It allocates a fresh controller and rebuilds metadata
  and record owners over held results. Cached measurement views remain paused;
  Run still explicitly reads/reactivates them. No old control epoch is changed or
  granted currentness again.
- Raw record-controller state is examined before replacing owners, even though its
  old `current()` predicate is false. Selected/submitted/busy workflows preserve
  their DOM and state with controls unavailable and an explicit paused reason.
  Existing Run is a deliberate page-replacement action; a submitted Keep may still
  finish. This cut does not invent a new Keep transfer or retry framework.
- Explicit execute/plan/load/stage/invalidate clears the retained-return flag.
  Retained child ownership delegates to the active child rather than renewing the
  hidden parent. Old asynchronous responses remain fenced by generation/serial
  and the existing exact-baseline checks.

The required real acceptance path is native Capture → return → Set baseline,
without a population rerun, plus a stale old metadata-handle negative control.
Browser outcomes and request accounting remain in VALIDATION.md rather than being
inferred from this source review.
