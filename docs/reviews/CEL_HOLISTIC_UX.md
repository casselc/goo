# CEL in the whole investigation journey

Status: design recommendation, 2026-11-14. All storyboards and component changes below are **planned**, not implemented acceptance. Review basis: source at the delivered CEL library/CLI pilot, repository guidance, the top of VALIDATION/KNOWN_ISSUES, implementation-plan/430 roadmap, CEL guide and acceptance, whole-product design 399, unknown-data entry 430, and the recent accessibility polish review. No browser execution, assistive-technology session, contrast measurement or new source query was performed for this review.

Personally inspected existing generated frames: Activity severity at 320, 390 and 1440; staged source message search at 1440; operator Services at 390; operator Saved at 1440; trace continuation 403 at 390. These represent their recorded journeys, not a fresh run against the current tree. Source inspection includes stream-entry, exploration, filter-editor, scope-actions, measurements and existing package/view/record-context ownership. The shipped CEL pilot admits fixed record fields and one severity helper; app editing, attribute presence and general extension lowering remain absent.

## Recommendation

Make CEL an optional way to edit the same contextual filter draft. Preserve **watch → scope → inspect → ask → preserve/reuse**. A novice should complete this loop without seeing CEL. An expert should reach text in one disclosure from Filters or Edit question, using the exact same typed question and explicit read gate. No CEL navigation item, separate explorer, blank language landing page, second execution engine or permanent editor above the stream.

Three priorities, in order:

1. Deliver a source-free bridge from the existing Filters/Edit question surface to the current compiler, with honest visual/text coverage and immutable whole-parent refinement.
2. Preserve draft, pause, source-read and return semantics across every authoring mode. A successful compile stages a question; only the existing Apply/Run action reads telemetry.
3. Consolidate compact scope/status presentation before adding more chrome. The 320/390 severity frames already put the first record near/below the viewport bottom. CEL cannot consume another permanent toolbar row.

These language slices should be reconciled explicitly with 436, whose aggregate-selection/held-return work is still the next numbered operator milestone. The older 427–429 preservation work remains pending; this review does not mark it delivered.

## What the actual interface tells us

The desktop severity frame has a useful stable record table and a restrained supplied-data histogram. At narrow widths, environment, signal, Investigate, Views, Filters, Resume, Hide volume, scope chips, refresh mode, Explain, band choice, Filter, chart, text-severity disclosure, pause/read counters and Find consume almost the entire first screen. Wrapping prevents horizontal overflow in these captures, but fitting is not the same as scanning efficiently. Keep records as the visual destination; fold optional analysis controls with their opened chart, and keep only the active scope, stream state and primary action outside it.

The staged search frame correctly focuses Run analysis and preserves Return to paused activity. It also repeats the fixed-read warning in a heading, action context and separate status paragraph. Replace that repetition with one concise state sentence and one scope disclosure, retaining the exact event-time bounds and complete selector. The empty body is legitimate before execution; it does not need placeholder charts or a second query builder.

Services discovery has a clear task and native environment/search controls. Its event window, namespace and coverage facts are useful but verbose at 390. Use the same compact scope summary proposed for Activity, with complete details nearby. The Saved frame clearly differentiates stored evidence and fresh trace reads. Preserve this distinction when text editing is introduced: editing a question creates a new draft, never rewrites saved evidence. The narrow trace-continuation frame correctly explains representative witnesses, pagination and independent reads, but long basis identifiers dominate space. Keep the limitation visible and put machine receipts in existing scope/evidence details.

## One draft, two presentations

The common model is an ordinary typed Goo question plus authoring-local state: original parent, editable refinement, text buffer, validation revision and opener. CEL is authoring syntax, not a second executable query payload. Continue using current question admission, source adapters, predicate budgets and question identity. Do not infer source support from CEL type checking alone.

Inside Filters, offer a native labelled choice **Conditions / CEL** only when the user opens authoring. Prefer a labelled native select or radios over a custom tab keyboard system. Conditions remains the default. Text starts from the selected editable filter, with a readable preview beneath it; environment, signal, clocks, measure, grouping and limits stay in the enclosing question surface.

The bridge must distinguish three cases:

| Existing predicate | Conditions | CEL presentation |
| --- | --- | --- |
| Fully representable by both | Fully editable | Generated equivalent text; subsequent edits round-trip semantically |
| Visual predicate outside pilot CEL vocabulary, such as attributes/presence | Fully editable | Keep it as a readable parent; edit a supported additional filter separately |
| Text not admitted by the source compiler | Retain last valid conditions | Retain exact draft text with a specific error; Apply/Run unavailable for that draft |

Do not claim whole-query round trip when the pilot covers only a fixed-field predicate subset. Never drop unrepresentable clauses, generate plausible-but-invalid CEL, silently convert missing to empty, or simulate unsupported syntax by filtering loaded records. Label a partial projection “Existing filters retained” and show the complete parent in the scope disclosure. An attribute-heavy question can still gain an admitted fixed-field CEL refinement without becoming text-editable in full.

Switching modes with invalid text keeps that text. The Conditions surface may show the last valid draft with a clear “Text changes need review” message; editing those conditions requires an explicit “Use conditions instead” action that explains it discards the invalid text. Do not invent two silently diverging active drafts. Whitespace and parentheses alone should not make a semantically unchanged question dirty, but preserve the user's original text until they request regenerated formatting. Semantic equivalence is required; byte-for-byte regenerated text is not.

## Refinement and replacement must be different operations

A gesture from a record, facet, severity band or selected interval means **refine this parent**, even if the parent contains OR. Preview the relationship in readable form: “Within current scope, also: Error severity and message contains ‘timeout’.” Internally it remains `(whole parent) AND (new predicate)`. Removing the additional predicate restores the parent exactly. Never edit a parent branch accidentally or flatten away OR grouping.

“Edit question” permits deliberate editing of the selected question's existing conditions. Present inherited context as such, and give any broader change a clear consequence. No default Refine/Replace toggle is needed in every editor. A separately invoked “Edit existing filters” action can enter replacement editing, where scope preview makes the resulting whole selector visible before Apply/Run. CEL's Program.Apply refinement contract must not be reused as though it were a whole-question replacement API.

## Planned storyboards

**Novice, live logs:** open Activity with the existing signal/scope discovery; watch records. Inspect a field, choose a value or numeric Error band, and see a staged scope chip while the held stream pauses. Apply scope starts a fresh rolling read; cancel/discard restores the existing applied scope while remaining paused. Inspect a matching record, open reported trace context, then return to the exact held records and selection. Save a reusable view only when its scope is applied. No CEL appears.

**Novice, a question from evidence:** enter “timeout” in Find loaded records. The local count remains explicitly about loaded records. Invoke the existing source-message search gesture, see one fixed-window draft and Run analysis. Open Edit question only to change conditions, grouping or presentation. Run, inspect a trace, then Return to paused activity without a read. Find's case-insensitive all-loaded-fields matching and the case-sensitive Body source filter retain distinct labels/help; do not relabel them as equivalent search.

**Expert, contextual CEL:** open Filters → CEL on an applied logs scope. Add `otel.severityBand(log.severity_number) == "error" && log.body.contains("timeout")`. Validate/compile without a telemetry read, show the readable refinement and unchanged parent/window policy, then use the enclosing Apply scope or Run analysis. Return to Conditions to inspect the equivalent severity range and literal Body match. A supported span condition is offered only on spans; changing signal retains incompatible text and explains the field error rather than deleting it.

**Expert, Services/Measurements to records:** start from an observed service or supported native instrument. Choose the existing record-context gesture and retain the resource/environment policy, measurement interval and qualification that nearby records are time-associated context. In that staged record question, add CEL in the same editor. Run explicitly, inspect, then return to the original service/measurement state without rereading it. CEL must not turn a metric label into a trace join or introduce a new metric-query language.

**Preserve and reuse:** save the valid ordinary question and compatible view with the existing preservation surface. Reopen stages the saved intent; it does not auto-run. Inspect/Apply a reusable presentation against another compatible question with the destination selector intact. Export code from the existing package/export disclosure: canonical typed definition first; optional equivalent CEL only where supported, with library version in technical details. A stored capture stays immutable; “Open fixed question” makes a new editable draft. A fresh capture remains an explicitly labelled new source read.

## State, error and focus contract

| Event | Required user-visible behavior |
| --- | --- |
| Open authoring or change Conditions/CEL mode | No telemetry read; no stream resume; focus stays on the chosen control |
| Type a change | Draft becomes unapplied, current result actions become unavailable as appropriate; Activity uses its existing pause/held contract |
| Compile successfully | Show readable draft; keep Apply/Run explicit; do not publish old results as matching new text |
| Compile fails | Preserve exact text, parent, held evidence and cursor; associate concise error with the textarea and block execution of this draft |
| Change signal/parent while compile is pending | Discard stale compile reply by revision/identity; explain current incompatibility without destructive reset |
| Apply scope | Existing fresh rolling Activity semantics; not the held chart's fixed event window |
| Run analysis | Existing fixed staged bounds and complete selector; separate loading/cancel/error state |
| Return/discard | Restore held parent state without source reads or implicit resume; restore actual opener or a stable labelled fallback |
| Reopen saved configuration | Stage its intent, preserving the existing dirty-draft conflict choice; never overwrite a dirty draft silently |

Prefer validation after a short idle or explicit validation action, with quiet pending feedback. Do not announce syntax errors on every keystroke. Provide line/column and a plain-language repair, e.g. “This filter supports literal contains; regular expressions are not available.” Source capability errors belong to the existing read/refusal surface, not a syntax underline. Unsupported attribute presence should say what is unavailable and retain the visual route. Parser bounds and complete-parent predicate limits need usable messages before Run.

Use a native textarea, persistent visible label, linked help/error via aria-describedby and aria-invalid only after validation. Keep browser selection, undo, IME, spellcheck-off code input, multiline editing and Tab navigation native. No mandatory syntax editor, focus trap or Tab insertion. A keyboard shortcut may supplement a visible validation action but must not replace it. Do not move focus on a successful background compile. On explicit refused execution, focus the first relevant invalid field or linked error summary. Preserve selection/caret through errors; expose copyable complete text and diagnostics.

Stream status should announce meaningful state transitions once, not every poll or keystroke. Existing Find totals can still announce each batch per known issues; adding CEL should not compound that. Native availability guards must reject hidden, inert, disabled, detached and retired actions as well as stale asynchronous callbacks. Cover route changes and held-return cycles, not just the first happy path.

## Shared components and removals

Reuse `filter-editor.mjs` for Conditions and existing fixed-question/stream adapters for mutation and reads. Introduce a small filter-authoring presentation adapter that owns local mode/text/diagnostics, receives the parent and returns an admitted ordinary predicate/question; it must not fetch telemetry or navigate. The compiler bridge needs a bounded source-free transport contract with revision-safe adoption. Do not put app compiler requests into passive display-element lifecycle.

Extract/reuse a compact scope summary for Activity, fixed analysis, Services and Saved: signal/environment plus window policy and draft/applied/stored state outside; complete predicates, exact UTC bounds, library provenance and receipts inside native details. Share formatting and supplied facts, not the controllers whose discovery/custody contracts differ. Reuse existing result-state, field-presentation, inspection-focus, question-session, preservation and view-composition mechanisms.

Remove duplicate staged-read warnings, duplicate draft selector displays in the same open editor, and new CEL-specific Save/Run/Export buttons. Keep one primary owner action. Fold code reference/help/version information locally. Technical field identifiers belong beside friendly labels or in details; checked overload IDs and uint suffixes are compiler/reference details, not the primary filter vocabulary. Conditions should show “Duration ≤ 1 ms” and “Error · 17–20”; exact ns and generated `u` literals remain available in CEL. Preserve large uint64 values as exact text, never through JavaScript Number. Generated code must not round values to make them look simpler.

Do not add a general plugin marketplace, arbitrary host-function registration UI, universal context manager, new capture format or chart framework. Prerelease status permits replacing awkward labels and duplicated composition; it does not justify losing existing saved/custody semantics. There is no compatibility floor requiring a new CEL-only parallel editor.

## Acceptance priorities

**P0 — first integrated bridge:** demonstrate source-free compile, honest supported/partial mode switching, unchanged whole-parent OR under AND, exact strings/uint64 values, stale-response rejection, and zero reads on mode switch/validation/stage/save/held return. Refused syntax must emit no executable partial question. Run actual keyboard/mobile journeys from Activity and staged fixed analysis with one shared editor. Verify one Apply/Run read gate using observed requests, not click-count assumptions.

**P0 — trust and recovery:** invalid text, unsupported signal/helper, unavailable source, over-budget parent composition, cancel, dirty saved-view conflict and repeated return must retain text and held evidence. Verify Apply uses rolling time and Run uses fixed time. Include missing/empty visual predicates retained under a CEL refinement, nested OR parents, long/Unicode/NUL-literal refusals, and integers above JavaScript's exact-number boundary. An unsupported expression must never become local-page evaluation.

**P1 — whole-product continuity:** use the bridge in one real Services-to-records and one Measurements-to-records journey; prove original metric/parent state returns without a read. Save/reopen/export an ordinary typed question and compatible reusable view. Keep captured evidence immutable and fresh execution explicit. Assert destination query identity when applying a presentation. Do not expand into metric CEL execution merely to make this demonstration broad.

**P1 — compactness/accessibility:** compare actual before/after 320, 390 and 1440 frames with records visible sooner when analysis is folded; no page overflow at 320/zoom, unclipped focused controls, readable exact bounds, long values contained, textarea resizing contained, visible focus and native keyboard completion. Screen-reader checks should cover control names, error association, dirty/pause/read announcements and opener restoration; state any unavailable AT/cross-engine checks explicitly. Screenshots alone establish neither contrast conformance nor speech behavior.

Keep the acceptance story finite: watch, refine visually, inspect, ask a fixed question, optionally edit CEL, run, return, preserve and reuse. A text editor passing parser tests alone is not whole-product integration.
