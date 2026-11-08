# Whole-product architecture checkpoint — 381

Status: read-only source review and proposals for root/planner reconciliation.
381's decoder candidate is underway; this review does not establish its runtime,
benchmark or source-regression success and does not adopt another execution queue.
No heavy checks were run by this reviewer.

## Preserve the product loop and completed foundations

Goo already supports current log/span Activity, exact scoped filters, typed numeric
measures and distributions, population explanation, recorded error triage, log
patterns, service/resource defaults, cross-service trace search, waterfalls and
trace-page comparisons, immutable captures/keeps, investigation findings, record
review, and reusable record-view compositions. These are implemented capabilities,
not missing vendor feature categories. `README.md`'s capability matrix and the
actual owners below are better evidence than historical milestone prose.

376–380 closed concrete connections: unsupported exponential metadata is visible;
reported histogram references can be followed in live Measurements; native source
recovery reuses the shared refusal contract; broader Services discovers exponential
emitters; Saved histogram captures follow a checked immutable export/sample
occurrence. Service widgets deliberately remain interval-only. A saved-style
renderer does not grant every possible relation.

The next useful sequence should move back to the general investigation loop,
rather than automatically enabling another native surface or executor. Preserve
**see telemetry → isolate a population → follow evidence → preserve/reuse intent
or observations**. An authored recipe, a fresh result, a held parent and an
immutable capture remain different objects even when their displays look alike.

## 381 boundary review

The candidate `internal/sourcejson.ReadObjects` is a small framing helper above
`internal/sourcehttp.ReadBounded`. Explicit byte/row limits, blank-line handling
and retained-whitespace policy belong to each source call. The helper does not own
SQL, typed parameters, status handling, request deadlines, numeric evaluation,
evidence clocks or lifecycle custody. Returning read causes unchanged and
classifying them at source owners retains cancellation/timeout identity.

`bytes.Lines`, `bytes.Clone`, `json.Valid` and an object-envelope check can preserve
raw tokens without decoding a map solely to discard it. Separate object/trailing
syntax, detached-row and independent domain tests remain necessary. Benchmark
results must name the bounded corpus; source inspection alone establishes no
performance improvement.

Root explicitly adopted budget-first overflow validation for both families as a
prerelease cleanup. This changes an already-over-budget malformed record row from
Malformed to Budget; ordinary invalid rows and native overflow precedence retain
their existing meanings. Do not keep a legacy priority knob. Record whitespace/
blank skipping and native physical-line retention/rejection remain explicit.

## Proposed next finite sequence

### Breadth — preserve and reuse a trace-search recipe

Concrete task: “Find traces that contain these service/operation conditions, then
let me keep that question for the next incident.”

`ui/trace-search.mjs` owns one to four checked span-condition drafts and a bounded
result table. `invalidate()` clears those drafts, and the panel has no Save,
portable-recipe or investigation-step continuation. Its request already has a
clean typed contract in `exploration/trace_search.go`: environment policy, fixed
window, qualified predicates and limit. `workbench/steps.go` currently admits only
record and measurement question steps. `workbench/view_composition.go` and
`ui/view-composition.mjs` are explicitly record-query compositions; a trace search
cannot honestly masquerade as one ordinary span predicate.

Add one explicit trace-search intent arm to the existing question/definition
workflow rather than another top-level library or an untyped query blob. Save and
reopen the whole checked request without reading source data; Search remains
explicit. The first cut can reopen the original fixed scope. Rebinding or rolling
windows should be a later explicit operation, not an implicit clock refresh.
Reuse the existing condition editor and trace-search owner. This adds no renderer,
query language, source engine or novel causal-join semantics.

Acceptance should include a nontrivial cross-service request, exact numeric
predicate above 2^53, empty/unreported environment policy, preserved qualified
conditions, source-free reopen and explicit rerun. The current semantics remain:
one span may satisfy several conditions; witnesses need not be parent/child;
reported trace IDs are not unique invocations or proof of complete traces.

### Depth — preserve a trace-search result and its qualified witnesses

Concrete task: “These were the traces and witnesses I found. Keep that evidence,
then let me follow one trace without losing the search that justified it.”

`exploration.TraceSearchResult` already records request, read time, basis, state,
matched-trace count, `hasMore`, bounded items, qualified witnesses, plan and
warnings. The existing capture question union has no trace-search result arm.
Use the existing immutable capture/custody path for a checked typed search result;
do not serialize its witness table as an ordinary record page or fabricate a
shared snapshot with a later expanded trace.

Source-free reopen should preserve the original observed search, including
bounded-list and witness limitations. Following one reported trace stages the
existing exact trace question and requires a fresh read. Back should restore the
same held result and citation without rerunning the search or rereading the
capture. Reuse the retained-parent pattern proved by 364/380 where it fits; do not
create a new global navigation owner. A fresh trace read can disagree with an old
witness or return no data after retention expiry; those are honest outcomes.

Acceptance needs independent raw trace/condition membership, immutable bytes,
qualified missing/empty values, bounded `hasMore`, source-unavailable reopen,
actual refusal/late-success behavior and a capture/search/trace citation that
keeps independent clocks and read bases separate. Preserve current trace-search
limits; this is evidence follow-through, not an unconstrained distributed join.

### Consolidate — one checked question-step dispatch, with current docs

Once trace search has a real preserved-intent consumer, reconcile its serialization
and dispatch with existing record/measurement steps. Use an ordinary explicit Go
sum-of-payloads and JavaScript tagged dispatch, with one admitted arm and matching
scope. Share the small validation/round-trip rules that actually repeat; retain
specialized acquisition and action owners. Do not build a universal executor,
subscription manager or reducer algebra merely to remove switch statements.

This is also the right finite point to curate the active capability map.
`docs/KNOWN_ISSUES.md` still repeats “379 current” and states that Saved exemplar
navigation is absent in older sections despite 380 being delivered. Public README
contains a capability map plus a long chronological milestone appendix. Keep one
current implemented/proposed/deferred summary and move history behind explicit
links. Historical notes must not become another planning queue. This is product
truth and maintenance work alongside dispatch cleanup, not a new documentation
platform or automatic generator.

## Keep independent gates and deferred work explicit

- Generic OTel fields and supported service/resource dashboards already exist;
  refine concrete missing contracts rather than add a Datadog/Grafana checklist.
- General exponential evaluation remains deferred by actual exporter fidelity and
  mathematical representation. Discovery and exemplar references do not restore
  lost thresholds, original scalar types or mixed bucket closure.
- Broad broker/database-server health, scheduled alert delivery, authenticated
  team concurrency, Databricks lowering, contributed kind execution and WYSIWYG
  authoring remain larger independent decisions. None is a prerequisite for the
  proposed trace-search preservation flow.
- Existing native gauge/counter display reuse is a legitimate later depth
  candidate, but it should not displace a return to general investigation after
  this predominantly native-focused sequence.
- Go 1.27.3 is already current. Use ordinary current stdlib APIs where they remove
  allocation or ambiguity, while keeping consumer APIs straightforward. Runtime
  upgrades and clever generic machinery are not product outcomes by themselves.

Root should combine this architecture proposal with the design/usability review,
finish 381 acceptance, and adopt only one finite next sequence. This document is
not an instruction to start all candidates concurrently.
