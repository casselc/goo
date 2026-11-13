# Roadmap after the 426 investigation checkpoint

Status: **426 is complete**. The 427–429 sequence below is the proposed next
finite cycle, reconciled with architecture and UX review. These next items are
not implemented. Actual 426 execution and bounded acceptance are recorded in
[validation](../VALIDATION.md) and its
[integration review](reviews/INVESTIGATION_CONSOLIDATION_426.md).

426 acceptance includes 1,648 passing frontend tests, the complete 17-fact
source/store journey and fresh exploratory actor, portable-view and attachment
regressions, selected/rest and row-copy checks, and native Quint/Hegel custody
checks without skips. Source/frame reviews and executed tests retain their
separate limits; this planning pass did not rerun them.

The product remains **current telemetry → isolate → inspect/compare →
preserve/reuse**. Ordinary OTel data is the entry; manifests and packages add
shared definitions. The next cycle should make existing useful work easier to
find and reuse, then consolidate the owners behind those interactions.

## What this cycle established

424 connects an inspected field to its exact applied conditions without removing
an arbitrary OR branch. 425 turns a counted categorical selection into existing
fixed selected/rest questions; counts supply choices, never comparison operands.
426 follows that work through explicit preservation and configuration reopen.
Its concrete fixes concern native preservation controls, selected-question versus
whole-view meaning, late completion, and active permission after a pause. The same
saved-view library is now reachable in the fixed-query builder before its first
source read, as well as in an active population's View settings. Portable import
therefore has a natural metadata-only entrance without another catalog or a
telemetry prerequisite.

Three distinctions are durable product rules:

- A saved view definition retains query and presentation intent. It is not a
  frozen pair of comparison operands. Opening stages; Run reads explicitly.
  Preserving a selected question is narrower than saving the whole comparison
  view; capture remains a separate explicitly acquired evidence artifact.
- A completed held result can survive a child investigation, while hidden or
  retired controls have no authority to acquire, publish or mutate it. Return
  should restore the actual opener and selection, not a rebuilt approximation.
- A generated selected/rest partition is relative to its parent predicates and
  fixed window. Independent source reads can disagree; logical complement does
  not establish a simultaneous or unsampled census.

## 427 — breadth: exact saved-view recall at the main entry

Outcome: “I saved this useful comparison; let me find it from Recent / Pinned
without rebuilding it.” Add an exact saved exploration-view reference to the
existing recall entry, reusing the current saved-view reader and validation.
Do not flatten it to the selected query: the second fixed question and the
composition must survive. Stage remains metadata-only; Use fresh and Run stay
explicit. A held visit still returns to retained data through its existing owner.

The gap is concrete: [scope history](../ui/scope-history.mjs) accepts Activity
and fixed-question configuration. The existing saved-view library is reachable
from the fixed-query builder and an active population's View settings, including
source-free portable import after 426. It still has no exact saved-view reference
in Activity's main Recent / Pinned entry. [Recent / Pinned](RECENT_PINNED_SCOPES.md)
explicitly excludes arbitrary composed views. Reuse the existing immutable view
catalog and reader; do not introduce another library or a saved-data format.

Acceptance: save and recall the 425 comparison through the main picker; preserve
both exact questions, grouping, typed environment policy, fixed bounds and kinds.
Pin/filter/remove stays local. Explicit Stage may read definition metadata but
must not read telemetry or execute a comparison. Missing, corrupt or incompatible
references show a refusal and retain the current held investigation. Cover
browser persistence opt-in, capacity, exact ID rather than title identity,
keyboard/narrow layout, withdrawn actions and late metadata responses.

## 428 — depth: preserve both fixed comparison questions as paired evidence

Outcome: “Keep evidence for both sides of this comparison, then reopen and share
it.” Adopt the existing pair-capture workflow from the fixed population
comparison. Keep selected-question capture as a separate choice; do not relabel
that single-query action as preserving the whole comparison.

[General captures](GENERAL_CAPTURES.md) already provide the pair store and accept `kind: comparison` with baseline and selection questions;
[`CaptureStore.createComparison`](../exploration/capture_store.go) independently
captures both server-read children and publishes a pair with their exact IDs and
comparison. The [distribution workflow](ERROR_COMPARISONS.md) already reaches it
through Capture comparison. The current population Preserve entry captures only
the selected query. The ordinary single-group Count questions created by 425 fit
the existing capture-family admission: same signal, checked environment and
grouping, with independently valid predicates and bounds.

Opening the capture form is local. Confirmation explicitly rereads both fixed
questions; it does not archive the earlier browser pair verbatim. Preserve operand
order, original configured scope, fresh source clocks, immutable child/pair IDs,
unknown coverage, source-limit refusal and partial-publication recovery. If the
second read or pair publication fails, a successfully stored child can remain an
independently useful artifact. Retain the current evidence and report the failure;
an explicit full-pair retry may reread both questions. Do not imply rollback or
start another read automatically.

Acceptance: start with the 425 selected/rest view, preview both exact questions,
confirm pair capture, reopen the stored pair and each child through existing
Saved controls with zero telemetry reads, and attach the parent through its
existing capture-card/reference path. A new capture may differ from the held
comparison after late telemetry; both results keep their own clocks and meaning.
Test no-data/diagnostic sides, source refusal, pending departure, actual return
where the owner remains retained, fresh-browser reopening and keyboard/narrow
presentation. Unsupported profiles refuse before a capture request.

Keep the two existing artifact families explicit. A general comparison capture
stores a `SavedComparison`; its validation checks the captured cohort result.
A [saved comparison receipt](SAVED_COMPARISONS.md) separately freezes ordered
capture references and a versioned `EvidenceComparison` decision. The latter's
ordinary-record compatibility currently withholds numeric comparison when
selectors differ. Do not assume a selected/rest pair can become a comparable
receipt there, loosen that contract, or substitute one ID family for another.
Use the existing general pair/capture-card path first; audit any additional
finding/receipt handoff against its real profile before extending it.

Reusable categorical selection pieces remain a possible later depth item.
[Typed packages](TYPED_PACKAGE_SCOPES.md) already store complete questions and
[view templates](AUTHORABLE_VIEW_TEMPLATES.md) preserve destination query scope.
A new fragment carrier needs an actual second-parent reuse gap and a minimal
validated owner contract; it is not required to complete this investigation loop.

## 429 — consolidate: shared scope and action contracts

Use the exercised 427/428 flow to consolidate pair-facing preservation commands,
then duplicated normalized-question projections and native action/late-completion guards where their semantics are
actually identical. Keep authored clause-path identity, semantic query identity,
source profile, held custody and active permission distinct; one universal key
must not erase those differences or merge artifact families with different
comparison/reopen contracts. Reuse native HTML controls and the shared
component action rows, rather than add another shell or a generic permission
cache.

Measure the affected path, including repeated pause/return metadata work, before
claiming a performance benefit. Preserve the existing row-copy budgets and
source/store budgets. Validate pending-save withdrawal, pause reentrancy, scope
replacement/ABA, definition reopen and actual-opener return. Withdrawal prevents
late UI adoption; it does not undo an already completed explicit store write.
Refresh relevant storyboards and desktop/narrow frames, then use that experience
for the next
whole-product checkpoint instead of automatically extending this queue.

## What stays outside this cycle

Services/resource dashboards, trace search and waterfalls, population explanation,
native measurements, display kinds/templates, captures and investigation documents
already exist; historical proposals do not turn them into missing features.
Broader source authorization/tenancy, monitoring delivery, arbitrary executable
kinds, classifier training, warehouse lowering and native histogram expansion
retain their separate operational/design gates. Preserve the shared data/kind
contracts and general-purpose OTel semantics while adopting the existing product.

Evidence for this plan: [capability index](CAPABILITY_GUIDE.md),
[previous finite cycle](ROADMAP_2026_11_13_CHECKPOINT_423.md),
[425 integration/UX review](reviews/COUNTED_SELECTION_REVIEW_425.md),
[authorable templates](AUTHORABLE_VIEW_TEMPLATES.md), and the source links above.
This is a bounded source/product planning review, not a new vendor campaign or
independent execution of 426's browser/model checks.
