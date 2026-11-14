# CEL in the whole product: architecture review

Status: independent source review of the delivered predicate pilot (`870e98a0`),
2026-11-14. Recommendations below are proposed; no product behavior changes or
runtime validation are claimed. Read against the current 430 roadmap, not the
historical queue inside IMPLEMENTATION_PLAN.

## Decision

Make CEL another authoring representation of typed Goo meaning, available in the
same population, definition and view workflows. Keep the existing semantic
families and execution owners. The architectural investment now is a common
checked-authoring protocol and explicit revision binding, not a universal query
object, editor workspace, evaluator, backend or plugin registry.

There is no legacy compatibility floor to preserve. Rename/restructure internal
interfaces when that removes ambiguous ownership; do not add parallel CEL fields
to every request and teach each consumer to prefer one. Historical captured
meaning still matters: prerelease freedom is not permission to reinterpret saved
evidence under a new compiler.

Three highest-priority changes before an app bridge:

1. Separate **refine a parent** from **replace an authored condition**. The current
   `querycel.Program.Apply` always ANDs with the complete parent. It is correct for
   contextual narrowing, but applying every edit to the last result accumulates
   stale conditions and consumes depth/node budgets.
2. Bind CEL source, library/compiler version and lowering digest to an exact
   authoring revision and predicate slot. The ordinary typed AST remains executable
   meaning; source is retained authoring provenance, not an alternate execution
   payload or an instruction to recompile historical evidence on open.
3. Add typed compiler diagnostics and target capabilities at the pure boundary,
   then reuse existing staging/Run and publication owners. Do not let compilation
   or a successful local CEL evaluation confer source-read, completeness or custody
   authority.

## What the actual product already provides

| Existing owner | Concrete contract | Consequence for CEL |
| --- | --- | --- |
| `exploration/model.go`: `Query`, `Predicate`, `Field.Type`, `Query.Validate` | Typed record population, fixed window, grouping, numeric/summary profiles, explicit missing/empty operations and bounded predicates | Compile record CEL into this predicate; CEL does not own grouping, time, environment or paging |
| `exploration/compiler.go`: `Compile`, `compiler.field`, `compiler.predicate` | Logical-to-physical source adaptation and parameterized ClickHouse planning | No SQL interpolation, browser SQL or additional CEL executor |
| `querycel/compile.go`: `Compile`; `querycel/program.go`: `Predicate`, `Apply`, `Eval` | Checked overload admission; detached lowering; whole-parent AND; bounded reference evaluation requiring referenced fields even in short-circuited branches | Good foundation; current fixed-column subset cannot claim arbitrary visual-predicate round trips |
| `ui/question-session.mjs`: `Session.stage/start/publish`, `identity`, `combineWhere` | Draft/applied generations, exact requested-question matching; composition of base/filters/shortcuts | App CEL authoring must stage through the same session and use its identity/generation gates |
| `ui/fixed-question-editor.mjs`: `mount` | Editor visibility/interaction control; no source execution | Add an optional expression representation inside these controls, not a separate execution lifecycle |
| `ui/population-workspace.mjs`: `stageRootQueries`, comparison preparation, selection drafts | Existing root/child scopes, staged refinements, explicit reads, held parent and independent comparison operands | CEL edits invalidate source-dependent actions through existing ownership; baseline/other-signal changes require explicit scope choices |
| `exploration/population.go`: `PlanPopulation` | Independent signal reads; rejects incompatible companion predicates rather than dropping them | Log CEL cannot silently become span CEL; maintain independently checked operands |
| `measurements/model.go`: `Definition`, `Question`; `measurements/revision.go`: `Definition.Revision`; `measurements/source.go`: `Evaluate`; `measurements/lifecycle.go`: `ObserveLifecycle` | Native source identity, units/temporality/gaps, fixed evaluation and separate lifecycle custody | Retain these family semantics; do not lower native measurement definitions into record predicates |
| `subjectfacts/model.go`: `Definition`, `Measure`, `UnknownWhen`; `subjectfacts/evaluate.go`: `Evaluate` | Prepared subject grain, population and measure conditions, unknown propagation; validates rows before filtering | Future CEL here needs a declared typed fact environment and its own lowering, not reuse of OTel record bindings |
| `subjectfacts/compile.go`: `Definition.Compile` | ClickHouse/Databricks prepared-fact plans with explicit preparation requirements | Existing Databricks support is not support for record CEL or raw native preparation |
| `workbench/view_declarations.go`, `view_templates.go`: `ExpandViewTemplate`; `native_views.go`: `ReuseNativeView` | Display declarations and checked reuse preserve destination question | Applying a view never installs donor CEL scope or provenance over the destination |
| `workbench/exploration_views.go`: `ViewContentID`; `view_artifacts.go`: `EncodeViewArtifact` | Saved configuration identity, strict portable YAML, explicit import separate from source reads | Extend one artifact/definition owner intentionally; do not invent CEL package storage |
| `exploration/capture_envelope.go` | Closed family question/result/receipt admission | Store admitted meaning in existing evidence families; CEL does not create a seventh evidence arm |

## The common layer to introduce now

Use a small pure authoring result, explicitly family-tagged, containing the admitted
semantic value, diagnostics, representation metadata and binding information.
The record arm contains an ordinary predicate/question; future fact/native/view
arms remain closed and separately validated. This is a transport/protocol over
family types, not a union blob executable by one generic evaluator.

A record compile request should bind: operation (`refine` or `replace`), signal,
exact base-question identity, draft generation, target predicate slot identity,
expression, CEL library version and selected source target/profile. The server
must independently admit the supplied base. A successful response returns detached
lowering, resulting ordinary question, exact binding, compiler/library versions
and structured diagnostics. SQL plan inspection may accompany it through the
existing compiler; compilation itself reads no source and publishes nothing.

For refinement, retain the complete parent inside AND, including OR and hidden
constraints. For editing, hold the immutable parent and replace exactly the owned
slot, then rebuild/validate the complete question. No structural subtree guesswork
or text search to discover ownership. A user switching to whole-predicate editing
must explicitly adopt that scope. Clearing a refine slot restores its parent;
clearing a replacement has explicitly defined empty-predicate behavior.

Bind plan requests and responses to complete semantic question identity, profile,
target and adapter/compiler revision. A plan is inspectable intent, not evidence
that source tables are ready or that a query ran. Existing capabilities/readiness
remain source-owner checks. Live rolling intent and a fixed run window are distinct:
compile a selector without pretending that a prior fixed planning receipt authorizes
the next Activity poll.

## Canonical meaning, source and revisions

The admitted typed AST is the executable authority. Do not use raw CEL text hashes
as population equality: whitespace, aliases and equivalent helper spellings need
not create different measurement meaning. Conversely, do not claim theorem-level
boolean equivalence or sort predicate children speculatively: existing
`QuestionSession.identity` retains child order. Initially canonical means existing
validated family normalization/identity, not a new global algebraic normal form.

Retain exact authored source, representation kind, declared library version,
lowering/compiler version, slot/base binding and lowered-value digest in a checked
revision-owned authoring record. This can be a closed optional envelope beside the
existing semantic definition in its owning versioned artifact. Validate its binding
on save/import; source plus AST supplied by a client cannot be trusted independently.
Changing only formatting changes authoring revision/provenance when necessary,
without manufacturing a new executed population. Define these identities separately
instead of silently changing `ViewContentID` or measurement revision recipes.

On reopen, execute/display retained admitted meaning without compiling with today's
library. An unavailable old authoring library can disable CEL editing and offer an
explicit upgrade preview; it must not prevent inspection of retained meaning or
rewrite an immutable capture. Upgrade is a new revision with visible semantic diff.
A capture may cite the authoring revision; its ordinary question, historical plan,
result basis and custody clocks remain authoritative in the existing store.

Visual edits replace the affected owned meaning and regenerate canonical CEL only
when the complete edited subtree is representable. Otherwise retain the valid
visual AST and explicitly report that this representation cannot edit it. Never
leave old source looking current. Preserve the exact original authored text as
history; canonical generated text is a separate presentation. The initial CEL
subset excludes attribute maps, missing/empty predicates and numeric bindings
already supported by record visual controls. Therefore “visual/text round trip”
must state its admitted subset and include a lossless refusal path.

## Diagnostics and capability design

`querycel` currently returns formatted errors. Add stable diagnostic codes/stages
(parse, type, unsupported lowering, binding conflict, budget, source capability),
source span where available, target/profile and an actionable message. Retain
ordinary validation errors as causes; do not make UI behavior depend on English
substrings. Refusal produces no partially adopted draft or misleading runnable
receipt. A stale response reports/ignores a binding conflict and cannot restore
retired controls.

Separate three questions: can CEL type-check, can this family lower it, and can the
selected source execute that admitted plan? Advertise supported operations from
reviewed compiler/adapter declarations. Do not equate CEL extension registration
with SQL capability or assume a global per-function boolean: signal, input type,
source target and profile matter. A failed source readiness probe is not a type
error. Keep capability inventory finite for this slice; a third-party registry is
unnecessary.

## Whole-product journeys and acceptance

1. **Activity to fixed population.** From the applied log selector, add a severity
   or Body condition through the shared filter controls. CEL compiles without a
   telemetry read; the fixed child shows the full parent and new condition. Editing
   the text twice replaces that condition once. Run reads the resulting population;
   Back restores the held paused parent with no read. Test parent OR preservation,
   changed signal/time/environment while compilation is pending, hidden controls,
   out-of-order responses and invalid/dead-branch expressions.
2. **Population to saved definition/package.** Build a count population, refine it
   visually or with CEL, select grouping/measures through their existing controls,
   save/export with checked authoring provenance, import and reopen the identical
   admitted question. Current `ui/telemetry-package-intent.mjs` `packagePlan` already
   verifies definition revision and exact fixed-question derivation: extend that
   contract rather than bypass it. A saved relative lookback rebinds time explicitly;
   fixed evidence does not. Test source-only changes, semantic changes, tampered
   source/AST pairing and old-library reopen.
3. **Apply display versus open original question.** Apply a saved bar/table/native
   display to the current destination. Preserve its complete question and compatible
   authoring binding, clear stale result as required, and wait for existing Run.
   Open original stages the donor question intentionally. Test hidden comparisons,
   separate baselines and unsupported profiles; no donor filter leaks through reuse.
4. **Measures and quality.** A filter over raw records never becomes a native stream
   selector, derived measure or health policy implicitly. Expose population,
   numerator, denominator and quality policy as distinct typed slots when introduced.
   Validate source rows and preserve missing/unknown/invalid states before applying
   conditions. CEL boolean failure must not erase invalid excluded data or convert
   unknown upstream coverage to healthy. Test zero denominators, missing attributes,
   reset/gap semantics and impaired scopes with zero rows.
5. **Code and portable target.** YAML and UI produce the same admitted family model;
   CLI validation/planning and app draft planning share compiler contracts. Record
   CEL to Databricks refuses until that adapter is implemented and parity tested.
   Prepared-fact Databricks continues with its existing explicit requirements. Test
   parameters/literals, target-specific refusal, unit/type preservation and immutable
   plan/source provenance without deploying a new serving layer.

## Sequence and alignment with older proposals

**Now, before or with the source-free app bridge:** settle refine/replace ownership,
source/semantic identity, diagnostic codes and target bindings; provide one pure
record authoring contract reused by CLI and app; mount it in the existing fixed
question/filter controls. Retain 436 selection/held-return consolidation as the
numbered priority: its action leases and staging contracts are exactly where CEL
must integrate. Implement one full Activity → fixed → Save/export → reopen journey,
including invalid/stale/refusal paths, before expanding vocabulary.

**Next bounded language slice:** define presence and empty-value semantics against
`exploration.compiler.field` and `Match`, then add guarded attributes with reference/
source parity. This addresses normal OTel filtering more directly than introducing
whole-query CEL builders. Promote reusable checked predicate/definition references
through existing catalogs only after source bindings and revision behavior are clear.

**Later:** native/fact CEL conditions, typed derived measures, explicit quality-rule
slots, portable record lowering and declarative view-expression bindings. Each
requires its own type/unit/grain/missingness/evaluation-time and target contract.
Do not embed arbitrary CEL inside passive components or let expressions perform
navigation, fetches or publication. Commands remain closed typed intents executed
by app owners.

The pilot guide correctly describes a predicate authoring compiler. The current
430 roadmap's “source-free app draft bridge and visual/text round trip” is endorsed
with the representability and revision gates above. View grammar260's shared typed
context, named scopes and host-owned commands remain the direction; its broad
conditional/repeat/expression vocabulary is not authorization to build a universal
CEL screen runtime now. Older YAML/CEL/KCL and Databricks proposals are aspirations
for common admitted semantics, not competing canonical runtimes. Deprecate any
follow-on proposal that stores a CEL-only executable definition, recompiles captures
on read, adds an independent text-editor Run owner or claims every existing visual
predicate is representable by this pilot.

Review validation: source inspection only; no product tests run because this change
adds this review document alone.
