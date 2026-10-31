# Authored captured-count comparison — proposed290

**Status: proposed, unimplemented.** This is a bounded next-feature contract for
architecture review, not a delivered grammar or API. It follows the completed
focus/timeline/Preserve sequence and the [289 product review](reviews/FLOW_PRODUCT_289.md).
The separate289 paging work does not implement this proposal.

## User outcome and first slice

A responder can add **Captured record count vs baseline** to a single-query
records/aggregate composition, choose one existing immutable capture, and see the
current held population count beside that baseline. Saving keeps the baseline
reference and presentation. Reopening stages the view; it does not query telemetry
or refresh the baseline. An explicit Run obtains the current side.

Support exactly one optional comparison child, one measure (`count`), and one
aligned summary presentation using the existing `goo-measure-comparison` with
`display.fields: ['count']`. Count means captured span/log records, not unique
requests, health, or a traffic rate. Unequal window durations remain permitted
only with a visible duration difference warning: raw counts are not normalized by
time. Do not add automatic red/green regression thresholds or significance claims.

This first slice establishes operand ownership and reuse. Grouped comparisons,
latency measures, shares/rates, previous-window queries, arbitrary baseline
expressions, live comparisons and native metrics remain separate extensions.

## Existing implementation to reuse, and its limits

- `workbench/view_composition.go` currently admits only aggregate, records and
  optional timeline roles. It explicitly forbids children owning queries or
  captured data. A comparison reference is therefore a **planned bounded schema
  extension**, not an existing supported source editor option.
- `exploration.CompareCohorts` validates supplied cohort accounting and derives
  `goo.comparison/1` summary snapshots through `view_comparison.go`. Supplied held
  provenance does not grant source completeness. `ui/view-comparison.mjs` checks
  canonical differences; `ui/measure-summary.mjs` renders the passive aligned
  summary/exact values. Reuse these semantics rather than calculate a JavaScript
  delta or invent a second comparison Dataset.
- `GET /api/captures/{id}` loads immutable stored evidence without telemetry reads.
  The existing cohort comparison operation can compare supplied held aggregates;
  its caller-held provenance must remain conservative even when the browser
  obtained one side from storage. A storage GET alone does not mint runtime source
  attestations for a serialized snapshot.
- Stored comparison receipts already exist in `workbench/comparisons.go`. They
  require **two exact capture IDs** and a server-verified expected result; their
  reopen path does not rerun a newer comparison engine. The new live/held parent
  side is not automatically such a capture and must not be passed off as one.
- `CompareEvidence` rejects ordinary question-definition differences, including
  different filters, while the cohort path can admit independently filtered
  populations. The first authored slice should use the stricter compatible-query
  boundary below so its eventual evidence-preservation path does not promise a
  comparison the existing stored receipt path will reject.

These are separate acquisition/persistence paths over reusable math and passive
presentation. Do not introduce a universal saved-object union or new evidence store.

## Configuration and ownership

The parent remains the sole fixed-query owner. The proposed child stores a checked
baseline capture ID and closed presentation options; it owns no executable query,
subscription, credentials, result rows, scope rewrite or completeness claim. The
baseline reference is a dependency of this comparison role, not a second parent
query or a general source binding language.

Illustrative role shape, **not a wire contract accepted today**:

```json
{"comparison":{"version":"goo.view/1",
  "kind":{"name":"captured-count-comparison","version":1},
  "baselineCapture":"<exact immutable capture ID>",
  "display":{"value":"count"}}}
```

Architecture review must settle the final role/declaration shape and schema
version. It should retain strict unknown-field validation, unique roles, bounded
child count, local ordering/visibility and at least one visible child. No arbitrary
nested layouts or per-child queries are required. Saved definition integrity does
not imply the referenced capture still exists or is readable.

Use the parent's held **whole-population count aggregate**, not its visible rows,
retained row count or newest inspection-page count. Parent paging must leave the
comparison's selection snapshot and identity held, as it does the aggregate/time
children. Explicit Run replaces the selection side and invalidates old actions;
changing the baseline invalidates the pair without altering the parent query.
Config-only hide/order/title changes are local and read-free. Opening/closing an
editor or comparison disclosure performs no telemetry reads.

## Compatibility and quality gates

Validate the ID syntax during configuration. On explicit preparation/Run, load
and integrity-check the stored capture through the existing owner. Report a
missing/unreadable baseline locally and keep the parent usable; never substitute
another capture. A paginated capture catalog may filter candidates, but the final
owner must revalidate the chosen exact ID and actual data.

First-profile eligibility:

1. One ordinary grouped record query and an existing `kind: explore` capture of
   the same signal. Exclude native measurements, pattern/state profiles, numeric
   bindings/histograms and reported-edge-specific profiles in this first slice.
2. Exact compatible checked environment policy and ordinary root definition,
   including predicates, grouping and requested measures. Windows may differ.
   Inspection page bounds must not become semantic differences, but normalization
   must use existing checked root-query rules rather than strip arbitrary fields.
   A comparison of a burst subset with a differently filtered whole population is
   outside this first slice. Do not drop a filter to manufacture compatibility.
3. The baseline contains an admitted whole-population count/group result, with
   valid existing accounting; bounded retained rows cannot supply missing totals.
   The current side has an applied clean held aggregate and no pending scope change.
4. The canonical comparison validates. Source-read disagreement remains a
   diagnostic/withheld difference. A known empty captured count is not unknown:
   retain the canonical engine's zero/no-data arithmetic and display semantics,
   with no normalized rate/share inferred. Test this explicitly rather than copy
   the stored receipt UI's more conservative wording into another contract.

Preserve each side's fixed half-open UTC window, source read clock, content
identity/source basis, and captured extent/sampling states. A stored baseline also
has a distinct receipt custody time. The new comparison has no shared source clock
and does not establish an atomic snapshot. Equal windows, compatible queries or a
complete returned aggregate do not establish complete upstream telemetry.

## Interaction sketch and storyboard

Planned UI sketch:

```text
View settings
  [ ] Captured count comparison
      Baseline capture [Choose captured evidence…]
      [Prepare baseline]  Loads stored evidence; no telemetry read

Captured records vs baseline
  Baseline: chosen capture · UTC window · stored read time
  Current: held aggregate · UTC window · applied read time
                    Baseline     Current     Change
  Captured records     120          180        +60
  [Exact scope and quality]
```

The example numbers are illustrative, not a screenshot or benchmark. “Current”
means the applied fixed result, not necessarily now. For a staged definition show
**Run population view to compare**; do not display a saved current count. Baseline
preparation is an explicit storage read and never an automatic telemetry refresh.

1. Start from Activity, narrow a service, and Customize its signal into the
   existing one-query composition. Run explicitly.
2. Enable the comparison in View settings, choose an existing compatible capture,
   and Prepare baseline. Load stored evidence and derive the comparison; no
   telemetry read. Show counts, aligned bars, exact signed change and both windows.
3. Choose another baseline. Withdraw the old pair immediately; revalidate and
   prepare the new exact reference. A delayed old preparation cannot replace it.
4. Page records. Their page and read clock advance while the compared current
   aggregate remains held. A compact status identifies that held result. An
   explicit new Run refreshes the selection side and comparison together.
5. Preserve → Save exploration view stores the child and capture reference.
   Reopen stages configuration without telemetry reads or an implicitly prepared
   comparison. Run explicitly reads current data and prepares the baseline through
   its existing storage path.
6. To preserve evidence, use existing explicit capture confirmation for the
   current root query. That is a **new independent read** and may differ from the
   held count. Compare the resulting two exact capture IDs through the existing
   capture comparison flow, review its actual result, then save its existing
   comparison receipt/card. Never silently attach the held display as that receipt.

Opening the baseline's stored evidence remains a storage-only action. Defer new
side/group scope gestures until the child has an explicit typed action profile;
first implementation needs no inferred navigation from a summary bar.

## Verification and delivery gates

Use real OTLP fixtures across two fixed windows with known different counts and
more than one current record page. Verify exact root query/selector preservation,
explicit baseline identity, zero extra telemetry reads for Prepare/save/reopen,
parent held aggregate identity during paging and a new selection after Run.
Include unequal-duration warning, known zero, invalid/missing baseline,
incompatible predicates/environment, source-read disagreement, late arrivals,
canceled preparation and stale/hidden/retired handlers. Pending source edits and
investigation notes must remain intact.

Hegel should generate compatible/incompatible query pairs and assert exact count
differences and detached snapshots through the real canonical validators. Quint
should distinguish staged, baseline-preparing, applied, paged and retired owners;
negative traces must reject implicit telemetry reads and stale pair publication.
Bombadil should independently observe a prepared pair, saved definition and
staged reopen; clicking Save or reaching a conforming prefix is insufficient.
Update the executable storyboard, actual screenshots, architecture diagram and
validation record once implemented. This proposal itself claims none of those
checks passed.

## Tradeoffs and next decision

Count is deliberately narrow and already supported by the canonical summary
comparison. It proves baseline custody and definition reuse without widening
quantile, temporal alignment or group-difference semantics. A latency-regression
workflow may ultimately need p95 or aligned histogram evidence, but extending
measures before this ownership boundary works would conceal the harder issue.

An immutable capture baseline makes the reference stable and acquisition explicit;
it is less convenient than automatic previous-window baselines. A separately
captured current side preserves trustworthy existing receipt publication, at the
cost of an honest extra confirmation and possibly changed result. Keep that
difference visible instead of fusing source reads.

**Next step:** architecture review of the bounded role/reference shape, exact
query compatibility normalization and runtime cohort admission. Then implement
one count child and its complete journey under the ongoing development roadmap.

## Architectural disposition — proposed290 review

**Approved as a proposed bounded first cut, with the following implementation
requirements. No production comparison child is implemented by this review.**
Reviewed the current composition validators, `CompareCohorts`,
`cohortViewComparison`, `CompareEvidence`, capture normalization and passive
measure comparison presenter. The proposed capture-ID dependency fits the existing
owner; no general data-binding framework or new evidence store is needed.

- Require **exactly one eligible grouping field** and explicitly exclude
  `spanStatus: true` in this first profile. The existing canonical
  `cohortViewComparison` intentionally returns no summary comparison for status
  questions. Either status support gets its own later contract or such parents
  receive a compatibility explanation; do not show an empty child after a
  successful-looking preparation. Duration-summary parents can remain eligible
  only when both validated operands supply their configured measures; rendering
  still exposes count alone. Never manufacture missing duration support merely
  to obtain a count comparison.
- Validate both root questions first. Use the existing checked environment
  equality (`SameEnvironment` semantics), including equivalent legacy named and
  explicit current-only representations, rather than raw JSON equality or a
  string-only environment check. Normalize only documented delivery paging using
  `CaptureQuestion.Normalized`, then disregard the two fixed windows for
  definition compatibility. Preserve predicates, grouping, bucket width,
  summary/measure directives and every remaining definition field. Do not
  generalize the numeric comparator's broader filtering allowance into this
  count profile.
- The canonical `ViewComparison` state/changes own the count display, not the
  outer cohort envelope's older normalized-share wording. A known-zero operand
  yields a known exact count difference even when state is `no_data`; source
  disagreement remains diagnostic with unavailable difference. Use the real
  comparator output and validator in the zero and disagreement tests.
- Keep the declaration closed and small: one typed comparison child with one
  exact capture ID and count-only display options, plus the existing sibling
  visibility/order mechanism. Settle one display vocabulary when implementing
  (the illustrative `display.value` is not automatically the presenter's
  `display.fields` input). Update Go/browser/source/property validators together;
  this prerelease needs no legacy-role fallback or second declaration authority.
- Preparation authority binds the active parent generation, baseline ID and
  admitted held selection identity. Baseline changes withdraw the pair before
  loading. Root Run, reparent/reuse or navigation retires delayed preparation;
  paging does not replace the held selection. Reapplying a saved composition to
  another query must recheck the capture against that destination root, and
  refuse or leave the comparison explicitly unavailable if incompatible. Never
  rewrite the capture's selector to make reuse succeed.
- Keep capture custody separate from baseline source read time and canonical
  held-content revision. The child stores only the capture reference, not a
  comparison receipt. Existing receipt publication still requires two actual
  capture IDs and a reviewed server-verified comparison after the new independent
  current capture. A successful authored comparison must not enable saving its
  transient selection token as a full capture or promise identical later evidence.

The implementable first journey is therefore one ordinary grouped fixed parent,
one existing compatible capture, one explicit baseline storage preparation, one
count-only aligned child, held behavior during paging, and source-free definition
save/staged reopen. Keep acquisition, comparison arithmetic and rendering in their
existing layers. Extend measures or baseline query expressions only after that
complete journey passes its positive and refusal cases.
