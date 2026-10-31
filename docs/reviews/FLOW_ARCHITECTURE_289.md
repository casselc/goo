# Architecture checkpoint after contextual preservation — 289

2026-10-31. Read-only architecture review of the source at288, the current
roadmap and the274/285/288 contracts. This review ran no browser, source benchmark
or latency measurement. The parallel paging report records its own measurements;
executed delivery checks belong in [VALIDATION](../../VALIDATION.md).

## Judgment and next bounded cut

The focus → authored timeline → Preserve sequence is coherent. Keep its existing
query owner, typed supplied inputs and distinct preserved objects. The next useful
architecture improvement is a **bounded record-page read for composed views**, if
measurement confirms the redundant work. This can remove discarded aggregate,
field-inventory and timeline work without introducing another evaluator, component
coordinator or storage family.

There is an important existing behavioral distinction in
`ui/population-workspace.mjs:pageSignal`:

| Context | Current paging behavior | Safe initial change |
| --- | --- | --- |
| Composed record/aggregate/timeline view | Full Explore rereads all aspects, but only `entry.recordResult` changes; `entry.result` stays held. | Acquire the new record page and its real fresh count without computing discarded sibling aspects. |
| Ordinary population | Replaces the whole result, including new aggregates and totals. | Leave unchanged in this cut. A page-only replacement would silently change current refresh semantics. |
| Stored capture | Reopens retained bytes and original custody. | Leave unchanged; no source read optimization should turn this into a fresh read. |

Source-call reduction is not proof of lower latency. Report actual statements and
bytes separately from benchmark timings; do not claim a production speedup from a
small synthetic fixture.

## Proposed acquisition contract

A count-plus-records profile in the existing exploration source is a proportionate
first step. `Source.fetch` currently acquires count, rows, optional groups, fields,
optional time series and optional grouped series. It then constructs aggregate
Datasets that a composition page request does not use.

A narrow profile may reuse the checked query/compiler, row projection, source
limits and result snapshot bank. It should obtain a **real fresh count** and a
bounded row page. This retains the current count/page disagreement diagnostics and
selection/Keep token workflow. It avoids a new token family and a storage union
solely to save one additional count read. A pure row-only profile can be evaluated
later if measured count cost justifies those additional semantics.

The profile must not fill a fresh page with the old count or turn its row count
into a denominator. Do not manufacture empty aggregate Datasets, empty field
inventories with completeness claims, or `no_data` for the whole population merely
because a high-offset page is empty. With count-plus-rows, `no_data` still follows
the independent count read and `changed_between_reads` follows the existing
count/page comparison. It is not proof that the retained chart and new page agree.

Keep the complete checked question, including typed environment policy, predicates,
fixed bounds, grouping and configured measure intent. Acquisition profile and
query intent are separate concepts. A supported summary in the question does not
mean an unrequested summary was materialized. The route/profile contract should
explicitly say which result aspects it acquires; no client-controlled JSON claim
may promote it to a complete aggregate.

## Safety requirements

1. **Separate identities and clocks.** The held aggregate/time inputs retain their
   original basis, capture clock and runtime witness. The new page has its own
   basis/read clock, delivery offset and continuation. Fresh page count must be
   labelled separately from held chart totals when they differ. Never splice new
   rows into a result under the old basis.
2. **Keep actual selected bytes.** `ResultSnapshots.Freeze` and selected-record
   retention currently preserve a full result token. The minimal page path must
   freeze its own rows and token through that same trusted source boundary;
   reusing the root result token would retain the wrong offset/index. Prove Keep,
   retry and displayed-record selection use the new page observations.
3. **Retain execution guards.** Check owner, requested question, page request serial,
   active generation and result membership before publishing. Abort on new root
   Run/navigation. Late page responses cannot alter a newer root or child. During
   page reads, prevent old row actions; on failure retain valid held chart data and
   report page unavailability without installing a partial result.
4. **Keep authority in application adapters.** `record-page.mjs` is a bounded
   supplied-data profile. `source-snapshots.mjs` privately binds admitted page to
   its wrapper. Extend trusted admission if required; do not teach a passive
   table, disclosure or view declaration to fetch or mint witnesses.
5. **Preserve independent capture.** Preserve → Capture confirmation remains the
   existing fresh bounded read. A held chart plus a later page is not an atomic
   capture, and a minimal transient result must not accidentally become eligible
   for whole-exploration capture promotion.
6. **Reuse source safety.** Parameterized scope, read-only ClickHouse options,
   cancellation, execution/memory/byte limits, overflow-row validation and exact
   timestamp/duration handling apply unchanged. Validate rejected queries before
   any source request.

## Acceptance for that cut

- Compare ordinary full Run and composition page operations using a recording
  source or real source instrumentation. Show precisely which SQL statements are
  eliminated; a representative count+rows page should not read groups, fields or
  intervals. Keep ordinary population behavior unchanged.
- A real multi-page composition must preserve the exact aggregate/timeline basis,
  clock and values while admitting a different page. Add late telemetry so a
  fresh count can visibly differ without changing the held chart.
- Select/Keep a newly paged record and independently reopen the stored selection.
  Verify exact fields and page-relative indexes, not only successful clicks.
- Exercise failed and delayed page responses, root rerun/navigation while pending,
  hidden/retired row handles, typed policies and present-empty/missing attributes.
- Validate the observed lifecycle and negative traces; add focused semantic
  properties for unchanged nonpaging scope and held-input identity. Update the
  existing composition storyboard/diagram instead of creating a separate demo.

## Priorities after paging

1. Review the complete Activity → focused authoring → Preserve → investigation
   flow against actual compact desktop/mobile frames. A contextual attachment
   disclosure could reduce Preserve height, but that is a small product polish
   task rather than another architecture milestone.
2. Choose the next reusable profile by an actual user task. Query-supported numeric
   timeline measures or an authored baseline comparison may reuse existing typed
   input roles; either needs its own support/ownership acceptance. Do not deliver
   both through an unbounded universal renderer migration.
3. Team-contributed templates and visual editing remain the larger grammar goal.
   The current single-query stack is an intentional bounded composition, not yet
   a whole-screen authored runtime. Advance it through one reusable task/profile,
   retaining host acquisition and action authority. Another generic registry,
   Plotly dependency or new top-level builder is not required by this checkpoint.

The current libraries already separate query semantics, source admission,
immutable data, active execution authority and passive presentation. Recent288
bugs show why moving controls needs real Apply/Back/invocation tests; they do not
show that those layers should be collapsed. Keep extraction proportional to the
next feature and avoid replacing explicit ownership with a general coordinator.

## Implementation review follow-up

The in-progress289 source uses a narrow public `RecordPageResult` while retaining
an internal `Result` for existing selection custody. This is compatible with the
current token consumers. Inspected all production `ResultSnapshots.Load` call
sites: Keep selects exact rows; XTDB retention selects exact rows and the real
count; capture promotion explicitly requires an explanation token and rejects
ordinary exploration tokens. No new snapshot kind is required to preserve that
boundary. A page-token capture-promotion refusal test should retain it.

Initial review identified one concrete projection issue: clearing grouping before
compilation can remove derived record columns referenced only by that grouping
(`exception_type`, `body_pattern_v1`, `severity_number`). The narrow read must keep
the original row projection even while removing aggregate acquisition. Full/page
projection parity is an acceptance boundary, not merely identical predicates.

The projection correction now compiles original-question row SQL and parameters
separately from the plain count's parameters. Reviewed that separation: group-only
derived columns remain available without reading groups or summaries. The public
page DTO omits the internal plan, and page basis still hashes its own original
question and acquired content. No remaining source-level architecture blocker was
identified; final test/browser outcomes are owned by the delivery agents.
