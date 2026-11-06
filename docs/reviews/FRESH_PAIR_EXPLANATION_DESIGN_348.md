# Explain fresh comparison populations — design review 348

Status: implemented source and all five actual-frame reviews accepted. Native
execution/custody results are recorded separately in the journey/validation.

After seeing a change, a user asks which recorded populations differ. Fresh count
and mean-duration comparisons should offer the existing explanation workflow in
place, without a detour into a different comparison surface. This is descriptive
count/share analysis, never causal or mean-latency attribution.

## Reuse and entry

`ui/explanation.mjs` already owns the explicit field picker, bounded multi-field
reads, ranked share differences, side-filter planning and observed-result custody.
`ui/distributions.mjs` and Activity use it. Mount this existing capability beside
an applied fresh pair in `ui/population-workspace.mjs`; add no sidebar, ranking
engine, automatic scanner or renderer catalogue.

Use one collapsed contextual **Explain population differences** entry. Keep the
existing one-to-six field controls and Run explanation inside. For a mean pair,
say briefly that this compares record counts/population shares and does not
explain latency changes. Do not make the primary comparison more verbose or imply
that an unopened disclosure already computed an explanation.

## Questions, labels and read boundaries

Use both whole original fixed questions, including distinct predicates, typed
policy and UTC windows. Any admitted projection stripping belongs to the existing
planner and must not weaken the underlying population scope. Count/share fields
are selected explicitly; suggestions remain observed hints, not an exhaustive
schema or OTel-conformance guarantee.

Opening, selecting fields and inspecting supplied rows perform no telemetry reads.
Run explanation independently reads both populations for each requested field under
the existing budgets. Its new read clocks belong to the explanation and do not
replace the held pair's clocks or imply a common atomic snapshot.

Retain the pair's natural roles through the result and side actions: Previous
window/Selected window, or Fixed baseline window/Selected window as appropriate.
Generic Baseline/Selection must not suggest that a frozen capture was used or
reverse the operand. Existing record counts, rational percentage-point change and
shared percentage scale remain descriptive; missing keys and empty strings stay
distinct. Zero share change is not an anomalous lead; unavailable share is not zero.

## Follow a lead and return

A ranked value's side action stages that exact whole operand conjoined with the
existing include/exclude planner delta. The child still requires explicit Run.
Preserve independent operand identity: a baseline lead must not accidentally use
the selected window or rebuild its query from visible facet rows. Existing value
byte bounds and disabled unsupported actions remain clear.

Back restores the held explanation and original pair, chosen fields, local result
presentation and current usable side controls. Return focus to the current semantic
opener, visible and unobstructed; disclosure or Back performs no replacement reads.
Old detached actions and late explanations cannot publish after departure or
retarget an identical-definition return. Do not dispose held observed output merely
because a staged child temporarily owns the visible workspace.

## Preserve the observation

Save observed explanation uses the currently shown result token and existing title
control. It does not rerun the source. Reopening proves the preserved result and
its independent clocks; Reuse explanation request stages configuration only.
Original mean/count pair observations are not silently converted into a new capture.
An uncertain save acknowledgment remains uncertain and cannot cause automatic
re-execution or duplicate publication.

## Acceptance

Use genuine ordinary Activity → fresh fixed comparison → explicit multi-field
explanation → baseline or selected lead → staged child → Run → Back → Save/reopen.
Include a mean pair with explicit count/share language, unequal populations, missing
and present-empty values, and a partial field without inventing a ranked result.
Check whole questions/policies/bounds, independent explanation receipts, exact side
predicate, retained parent and advertised source counts. Mounted/generated negatives
cover mean-attribution copy, unsupported fields/values, pending/late/departed owners
and save-without-reread.

Actual desktop/narrow frames should show compact contextual entry, readable ranked
values and operand labels, side continuation and retained return. Use observed
traces for IO/custody rather than infer those facts from screenshots. No fresh-entry
above-fold claim follows from a task-scrolled explanation frame.

## Frozen implementation source acceptance

The frozen two-file implementation mounts the existing explanation owner behind
Explain population differences only for a prepared fresh pair. Count/share copy is
explicit; mean comparisons add the no-latency-attribution sentence. Contextual
role labels now carry through headers, per-side actions and staged-plan text while
uncontextual existing callers retain their existing labels.

The returned filter plan is checked against the expected complete retained side
question before staging a records child. The child remains explicit Run. Parent
explanation snapshots retain selected fields, result, staged plan, pagination and
disclosure state; returning creates fresh handlers and locates a semantic current
opener. Save continues using the observed result token, with no source rerun.
Source review accepts the bounded integration. Real desktop/narrow visual and
execution acceptance remain pending; restored state is not a durable branch store.

## First actual partial-field frame

Reviewed `348-partial-field.png`. Previous window 13 versus Selected window 107
are explicit, and the completed service field retains ranked counts/share changes
with correctly labelled side actions. The unavailable cohort field has no invented
comparable rows; its known selected read and unavailable previous side remain
independently inspectable. The view says partial and descriptive, not causal or
statistically significant. Run and Save observed explanation are distinct.
First partial-field appearance passes. This frame does not establish child/Back,
mean wording, narrow layout or preserved-result reopening; those remain pending.

## Final actual frame acceptance

Reviewed `348-restored-explanation.png`, `348-narrow-composition.png`,
`348-preserved-observation.png` and `348-mean-population-composition.png` after
the partial-field frame. The restored result retains the 13/107 roles, selected
fields, ranked exact count/share differences, missing-versus-empty rows and open
gamma Previous-window actions. Its staged side question is explicitly Previous
window; Run and Save observed explanation remain separate controls.

Narrow rows keep their own Previous window and Selected window labels and shared
percentage bars, with wrapped field/action content inside the viewport. They are
a task-scrolled existing responsive layout, not a fresh-entry density claim. The
preserved observation shows the same ranked result using the existing stored
Baseline/Selection labels and independent scope disclosure; those generic stored
roles are not represented as newly saved contextual presentation metadata.

The mean-pair frame explicitly says count/share comparison does not explain latency
changes, then presents service count/share changes rather than duration drivers.
All five actual frames pass visual acceptance. Native exact child/Back focus,
old-link refusal, source counts and save/reopen-without-reread establish execution
and custody separately; these are not facts inferred from screenshots.
