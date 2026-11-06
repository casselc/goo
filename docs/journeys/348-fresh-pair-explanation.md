# Explain the composition of a fresh fixed-window pair

Status: implemented and verified against the reviewed Go server. The executable
[journey](../../journeys/348-fresh-pair-explanation.spec.js) uses genuine OTLP and
ClickHouse reads, with explicitly planted failed acquisitions.

The user starts in Activity, scopes ordinary telemetry to a named environment and
an exact service namespace, pauses the loaded stream, and compares the previous
window. The selected and baseline questions retain the effective typed environment
policy, namespace predicate and original UTC bounds. Explicit Run acquires the two
whole populations; the small displayed record page does not define their counts.

Explain fields opens without reading. The user chooses service and a cohort
attribute, then explicitly runs the explanation. This is a fresh count-based
composition comparison of the original two questions. It does not reuse the pair's
old clocks as new observations. A mean-duration comparison in the bounded pilot remains a separate latency observation: field proportions do not
explain or attribute a latency difference.
The bounded mean pilot uses a fresh spans batch at its own clock. In Activity,
choosing service in the grouping picker stages the completed fixed question
without reading; Run, linked-population customization and the existing mean
controls then build the mean pair. It does not assume that Analyze automatically
adds a grouping. This differs from the primary count entry's direct previous-window
comparison action.

The user selects a value on the baseline side, stages its exact supporting-record
question without reading, explicitly runs it, and returns with Back. The original
pair, explanation result, selected fields and current controls must remain usable.
Back performs no replacement reads. Save observed explanation preserves the
acquired result token; reopening the preserved result performs no source read.

A scoped local upstream proxy forwards healthy SQL unchanged to the real
OTLP-backed ClickHouse source. One deliberately failed field read establishes
partial-state behavior: healthy counts and fields remain independently known,
while the failed field publishes no invented comparable rows. An explicitly
retired pending explanation cannot later replace the current owner or save state.
Successful telemetry and UI results are not fabricated.

## Verification

- Exact original side queries, typed policy, namespace, fixed bounds and counts.
- Independent count/field clocks and a count-only physical plan, including a
  bounded mean comparison if the final contract admits it.
- No source IO on opening, field edits, side staging, Back, saving an observed
  token or reopening its preserved result.
- Explicitly advertised acquisition/retry boundaries, partial fields and late
  owner retirement; old controls cannot adopt into a returned owner.
- Actual desktop and narrow frames, preserved observed result, and an appropriate
  existing explanation lifecycle/model/actor regression.

The main journey passed in **39.9 seconds** (42.2-second runner), with
**21 source API requests**, 47 total API requests, 80 upstream SQL requests
including readiness and ordinary multi-query reads, and no browser errors.

The count fixture has 107 selected records and 13 previous-window records, while
Activity and the selected inspection page show only 100 rows. Selected service
counts are 70/37; previous counts are 10/3. Cohorts include missing and present
empty separately, each observed once in each window. The selected-only beta and
previous-only gamma remain distinct, including gamma's exact 0/3 counts.

| Action boundary | Source API requests |
| --- | ---: |
| Count Activity discovery and namespace-scoped streams | 4 |
| Explicit fresh count-pair Run | 2 |
| Partial explanation, then explicit complete rerun | 2 |
| Explicit previous-side gamma record Run | 1 |
| Pending explanation cancelled | 1 |
| Fresh spans Activity discovery and scoped streams | 4 |
| Grouped span root, linked population and customization reads | 4 |
| Explicit fresh mean-pair Run | 2 |
| Explicit mean composition explanation | 1 |
| Open/edit/stage/Back/save observed/reopen preserved | 0 |

Each two-field explanation performs six physical reads: two whole counts and
four independent field reads. The planted baseline-cohort 503 produces a partial
result; the service field and whole 107/13 counts remain known, the selected
cohort partition remains available, and no comparable cohort rows are invented.
The explicit rerun reacquires the whole explanation rather than disguising it as
a failed-side-only retry.

The previous-side gamma question retains the complete original namespace/policy,
window and profile plus exactly its gamma predicate. Its one explicit record read
returns three worker records. Back restores the observed explanation, selected
fields, open disclosures and staged plan without rereading. Focus returns to the
connected, visible, unobstructed canonical staged opener. Reattaching and invoking
the old actual continuation callback changes neither local state nor source IO.

Save sends only title and the observed result token. The saved result equals the
observed complete explanation with the ephemeral token removed; its file remains
byte-identical through a later cancelled acquisition. Reopening the preserved
result performs no source read. The cancelled upstream connection closes before
its late refusal can replace the current state.

The fresh mean pilot reports exact means **403000000/107 ns** and
**19000000/13 ns**. Its held mean receipt remains unchanged after explanation.
The explanation reports record counts/shares of 107/13, preserving the original
summary/grouping request, while its four physical SQL reads contain no duration
or quantile computation. The visible note explicitly says this does not explain
latency changes.

The first integration run exposed a real generic routing defect: the global
question-link delegate ignored an already consumed click and replaced the local
child with a new global question. The delegate now respects defaultPrevented;
focused and full UI checks cover consumed and ordinary links. The successful
rerun uses that reviewed fix. Separate outer tool-response delays were not the
cause of this assertion failure. Two unreached test assumptions were corrected
before rerun: the existing capture subdirectory and an explicit ordinary grouping
gesture for the mean entry.

- [Partial field with healthy independent observations](screenshots/348-partial-field.png)
- [Restored explanation after supporting records](screenshots/348-restored-explanation.png)
- [Narrow value/share comparison](screenshots/348-narrow-composition.png)
- [Preserved observed explanation](screenshots/348-preserved-observation.png)
- [Mean pair's separate population composition](screenshots/348-mean-population-composition.png)

The shared explanation regression was validated **in parts**, rather than
claiming that its earlier failed combined run passed:

- The existing six-field numeric-selection native flow passed its independent
  source/count/ranking, late-arrival custody, side-filter and high-cardinality
  oracles. Its observed Quint ledger passed in **64.07 seconds**, with all 21
  negative controls and no skips. A separate replay of an earlier actual ledger
  passed in 78.93 seconds. The original 50-second model timeout was a test budget
  limit; only this observed gate now uses 120 seconds.
- The existing 60-second Bombadil actor was run separately against the actual
  UI-created saved capture. It passed in **60.175 seconds**, with 1,004 named
  snapshots through complete, exactly two source operations, one promotion and
  one pure plan. The final state has six ready fields, a valid promotion ID,
  preserved frozen evidence and preserved side scope. The wrapper now requires
  actual positive completion, not just a successful process exit.
- The ownership native flow passed separately in **55.0 seconds** (57.6-second
  runner). Its existing model passed in **30.79 seconds** under the unchanged
  50-second budget, with no skips and all 15 negative controls. Twelve observed
  events end at five explicit source reads and four pure plans.

The historical regression needed two honest control adaptations: numeric bins
now use admitted `bin:68` identity rather than obsolete ordinal `68`, and the
actual reuse button lives in the saved-evidence toolbar rather than inside the
explanation body. The unchanged keyboard gesture selects exact bins [8,16) and
[16,32), and the pure plan must retain indexes 68–70 and the 160/160 oracle.
No product compatibility shim or fabricated UI event was added.

Workspace departure permanently retires the root numeric brush lease. The
ownership journey now explicitly reruns the same parent question before making a
new brush selection; its 320-record parent has an independently later clock.
The model requires that real refreshed response and rejects missing/wrong/stale
refreshes or an omitted source read. The later pending pure filter is retired by
closing and reopening the explanation disclosure, preserving the current
observed result without a source read. This does not claim that route return
silently preserves a usable brush. The misleading tray copy was corrected to
require Run explicitly after returning.

Driver image evidence is separate. The positive actor's 1,004 JPEGs have four
unique hashes, with 975 repeating the terminal image. That image correctly shows
the original frozen 160/160 saved evidence; the fresh reused editor is below the
viewport. It neither contradicts the fixed result nor independently proves fresh
six-field completion. Named DOM states and response oracles establish that
completion; the five new native frames establish the fresh-pair appearance.
A labelled terminal sample, audit, named checkpoints and compact campaign summary
remain after reviewed raw-driver cleanup.

Compact fixture clocks, exact questions, complete observed/ownership model inputs,
pass logs, result receipts, custody/replay inputs and labelled failure notes remain
under `artifacts/fresh-pair-explanation348/` and
`artifacts/population-explanations/`. Genuine lab telemetry, saved views, captures,
volumes and all 51 pre-existing pending stream entries are preserved.
