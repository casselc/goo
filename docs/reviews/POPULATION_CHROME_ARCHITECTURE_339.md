# Population chrome consolidation — 339

Status: final source and native 339 acceptance pass; affected regression gates
are recorded separately in the validation log.

The [active roadmap](../ROADMAP_2026_11_05.md) selects a consolidation after
[paired category populations](../CATEGORY_PAIRS.md). This cut changes presentation
and shared display policy, not query authority or observation ownership.

## Compact definition facts, retain independent observations

Page-level scope chrome describes the checked question definitions. Reuse the
existing UTC/window formatter and a native details disclosure rather than adding
a scope manager, source owner or public grammar kind. The compact summary may
abbreviate environment and window text; its disclosed rows must retain the full
environment policy and exact selected and previous/fixed-baseline half-open UTC
bounds. Cross-day intervals and unequal fixed baselines remain distinct. Literal
empty/unreported/named environments must not collapse into the same identity.

A disclosure toggle has no telemetry, metadata, storage, navigation or mutation
effect. Existing Run, Filter both windows, Save and Back owners keep their guards.
Wrong-field refusal and other actionable task status stay outside the collapsed
facts. Disclosure state belongs only to the same checked selected/baseline
question definitions, including their environment policies; an observation clock
changing is not a new definition, while changing either question is.

The held count comparison and independently newer facet observations retain their
own counts, scope windows, read clocks and receipts. Do not copy their observations
into page-level definition chrome or infer a common transactional snapshot. A failed
facet read can coexist with an older successful count comparison. Both must remain
truthfully identified, including partial counts and unknown coverage.

## Control ownership

The broad application rule `.recent-canvas .population-facets select` currently
changes nested passive selectors. Narrow that rule to explicit application-owned
controls. Standalone and nested comparison controls should retain the component's
own compact sizing. Avoid an inverse override catalogue or a new universal table/
control styling layer. Verify the actual computed size and overflow at desktop
and narrow widths; a selector text assertion alone cannot prove visual behavior.

## Numeric knowledge and geometry are separate

The shared count-cell helper needs an explicit policy for withholding comparative
geometry. `geometry: false` should retain a supplied exact numeric value, its
`data-value`, display text and known state while omitting its bar. A true null/
unavailable value remains unavailable even if an alternate display string is
provided. Known zero and large uint64 strings keep their exact value. The default
must preserve existing callers' shared zero-based scale and rendered bars.

The paired facet should pass the actual known count and this policy rather than
encode a known partial count as null plus replacement text. Local Find must not
recompute the full union scale. This does not certify upstream completeness, add
rates/deltas or change the existing ready/unknown/absent data model. Renaming the
projection's internal readiness flag is outside this bounded cleanup.

## Acceptance and validation

- Focused shared-cell checks cover known positive, zero, large exact value,
  unavailable value, default geometry, explicit geometry suppression and accessible
  numeric naming. Mounted facet checks must retain the 338 partial/zero semantics.
- A fresh native saved-view entry and task-scrolled frames are separate views.
  Measure the initial viewport without cropping away page chrome. Open/close the
  definition disclosure and prove zero additional requests or writes and unchanged
  checked definitions and observation receipts.
- Exercise the ordinary common-filter/Run/Back path and failed-side Retry with
  the existing real OTel fixture. Keep focus behavior and retired-control refusal.
  Source/read effects should still conform to the existing 338 model; creating a
  second effect model for presentation-only changes would duplicate its contract.
- Inspect standalone and nested selectors and count cells, narrow overflow,
  keyboard operation, exact disclosed UTC/policies, healthy-zero/failed-side
  counts and the separation of old comparison versus newer facet receipts.
- Regenerate affected storyboards/frames and retain compact actual results.
  Driver DOM/API completion remains separate from image-capture validity.

No backend DTO, source route, dependency update, query engine, compatibility alias
or saved-data migration is implied. Go remains on the current verified toolchain;
this presentation task does not justify gratuitous runtime or module changes.

## Final source review

The implementation reuses the existing `.population-scope-context` disclosure;
it does not add another panel or a toggle handler. Before rerender, it reads the
old disclosure's DOM open state and summary focus, restoring them only when the
full selected/companion/baseline definition key is unchanged. Changing predicates,
grouping, limits, windows or environment policy changes that key. The helper is
presentation-only and leaves source objects intact.

The disclosed facts use normalized environment selection, so scalar names retain
explicit current-only policy and empty/unreported modes remain distinguishable
from identically worded literal names. Both selected and baseline definitions keep
exact half-open UTC bounds and explicit policy text. Existing filter editors,
Run/Save/Back guards and independent component receipts are not moved into a new
owner. Task status remains outside the collapsed definition facts.

The scoped CSS excludes nested passive labels/selectors/notices, including the
mobile hiding rule that previously could affect component notices. The shared
count-pair label/select owns compact intrinsic sizing independently. Actual
computed sizing and visibility remain native acceptance requirements.

The shared cell accepts a strict boolean geometry policy before mutating the host.
A withheld known number retains its exact value and known class, while its track
is hidden and no fill is created; a null value remains unavailable regardless of
alternate text. The paired facet supplies the real count plus the readiness-based
geometry policy. The projection's existing data contract and source/read owners
remain unchanged. Focused checks cover zero, values above JavaScript integer
precision, uint64 maximum, default scaling, malformed policy and partial facets.

No new Hegel or Quint model is added for this presentation-only change. Existing
338 projection properties and observed effects provide the appropriate regression;
the new native disclosure and computed-style checks own the new interaction claims.

## Selected next general-purpose task — 340

The next breadth cut is **typed environment scope for trace-wide search and its
continuations**, rather than more facet polish. This is proposed work, not a
capability delivered by 339. Architecture and design agree that it unlocks an
existing natural investigation from ordinary OTel entry contexts.

Source confirms a real capability gap:

- `ui/trace-search.mjs` `buildRequest` requires a nonempty scalar
  `parent.environment` and writes only that name. A direct no-IO probe of the
  actual function accepts scalar `prod`, but rejects effective named `prod`,
  empty-label and unreported typed selectors with its fixed-environment error.
- `exploration.TraceSearchRequest` contains only the scalar environment.
  `CompileTraceSearch` calls `identity.CurrentEnvironment`; each condition's
  ordinary query, witness validation and reference record membership also use
  the scalar assumption.
- The UI's fixed-trace continuation reconstructs `environment: request.environment`
  rather than retaining the admitted typed policy. Adding a frontend label alone
  cannot repair request admission, source membership or continuation authority.

Reuse the existing record environment selection contract and physical source
adapter. Carry the original checked selector through request validation, SQL
parameters, echoed response, witness membership, independent reference execution
and exact-trace continuation. Preserve current-only versus effective precedence,
including current-key override of a conflicting legacy value, and distinguish
reported empty from wholly unreported environment. The existing 1–4 independent
span predicates remain existential tests; condition 1 retains the originating
population predicate while added conditions may match other services. No parent
ordering, topology, distinct-span, invocation or causal relationship is inferred.

Bounded acceptance:

1. Start from a genuine Activity or Services span population using an effective
   typed environment, then use **Find traces containing…** or the operation seed.
   Adding conditions stages the search; explicit Search acquires its one bounded
   source aggregate, with exact full-population trace total and lexical list limit.
2. Use real standard OTel resource keys to prove current-only, effective named,
   conflicting current/legacy, current empty, legacy-only empty and unreported
   membership. Include cross-service matching conditions and excluded controls.
   Missing/empty labels must not be fabricated into scalar names.
3. Validate each returned witness against the same admitted environment policy,
   exact window and condition, retaining qualified missing/empty field evidence.
   Reject forged/mismatched echoed policy, false proof and malformed witnesses.
   Extend the independent Go reference oracle and compiled source parity checks.
4. Continue from a result/witness into the exact reported trace or staged record
   view while preserving its typed selection and original half-open window. Keep
   the current explicit source-read boundary and retained parent return; refuse
   departed and ABA completions without automatic rereads or writes.
5. Keep condition/node/list/source budgets, reported-edge refusal, no-data,
   cancellation and source uncertainty explicit. Update wire/OpenAPI fixtures,
   semantic/model tests and ordinary no-manifest native stories where required.

This is not a new trace engine, search language, scope manager, telemetry join or
renderer catalogue. Previous-window duration summaries remain a later depth
candidate requiring their own measured contract; they are not a parallel queue.

## Native 339 acceptance

Architecture independently inspected the four current frames: fresh desktop,
fresh 390px entry, expanded exact scope and partial known-count presentation. The
fresh frames use document scroll zero and the same genuinely saved 338 definition,
with original selected/baseline counts 107/13. The facet and comparison start
28.78px earlier at desktop width and 32.20px earlier at narrow width. This is a
modest measured improvement, not a claim that all narrow-view comparison data is
above the fold. Exact scope is an interaction frame, separate from entry height.

The expanded disclosure shows normalized effective policy and full half-open UTC
selected and previous bounds. The compact summary retains the environment and
window relation; task status remains separately visible. Opening/closing the
native summary by keyboard preserves focus and makes no source request. The
source task's local records presentation change preserves the open same-question
disclosure. Original checked questions and supplied comparison facts are unchanged.

The computed passive selector probes independently compare standalone and nested
hosts at both widths: each selector measures 99×37px with 12px font and inline-flex
label; its independent-population notice remains visible. These probe results
cover the reviewed configurations, not every possible application ancestor or
coarse-pointer environment.

The partial frame visibly preserves known selected count 27 while baseline is an
unavailable dash; the numeric cell retains `data-value=27`, known state and withheld
geometry with a hidden track. An older held 27/4 comparison remains separately
identified below. Baseline Retry makes one lookup and leaves the healthy witness
unchanged. The native summary records zero source reads for disclosure, staging
and Back, no browser errors and unchanged source questions.

The 35-second actor records 383 snapshots across phases 0–7, four verified
transitions and positive scope verification, six source requests including two
value lookups, two comparisons, zero writes, immutable parent facts and no
horizontal overflow. This is DOM/API completion; actual image capture is audited
separately and is not used to infer visual acceptance. The independent Playwright
frames above provide the reviewed visual evidence. The full 338 regression remains
a separate effect/semantic gate, not something this architecture review relabels.
