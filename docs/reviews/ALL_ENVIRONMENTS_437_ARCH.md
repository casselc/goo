# 437 All-environments scope: architecture review

Status: contract, production source and independent focused checks accepted.
Final completed runtime artifacts were inspected; no architectural blocker
remains. The parent-owned browser/actor/native verdict is recorded separately
below and is not inferred from source inspection.

## Agreed boundary

All is a checked selector policy, not a fourth observed environment category.
Keep `EnvironmentScope.Validate` and `ObserveEnvironment` on named/empty/unreported.
`EnvironmentSelector` accepts effective `scope.mode: all` through a dedicated
semantic constructor/private policy. A zero value, blank legacy name, malformed
selector or named literal `all` must never broaden a read.

Parsing/identity and execution permission remain separate. Existing environment-
only Query and UI snapshots legitimately resolve selection without a signal.
The log Query/Tail owner admits All; unsupported spans, inventory, trace and native
owners refuse before SQL. Native lowering needs an explicit defensive refusal
because its old non-named fallback otherwise resembles empty-label matching.

All record SQL omits only the environment predicate. Signal, time, ordinary
predicates, row/group/byte budgets and parameterized values stay authoritative.
The configured workspace source remains the authorization boundary; this feature
does not introduce multi-user access control or permission-derived environment
lists.

Existing links, saved definitions, captures and source-free CEL planning carry
the full typed log question. Do not project All into an empty scalar environment,
drop it from an immutable identity, or introduce another artifact family.

## Row and handoff requirements

Reported environment facts retain current-key presence and legacy fallback.
Current empty masks legacy named. Neither key means unreported only when the
resource map is actually available and well typed. An absent/malformed map must
not authorize a guessed narrow query.

Wide row inspection is local. Trace/cross-signal context must refuse before API
execution until explicit typed narrowing. A narrowing action derives from actual
resource facts, never a title/pill string. The new scope cannot silently replace
the current question on a refused or retired operation.
All membership is independent of missing environment metadata. Oversized or
otherwise unselectable reported labels stay visible as facts; they cannot throw
the row renderer, disappear from the source population or authorize narrowing.

An All parent and a typed named child have different environment selections.
Keep existing comparison compatibility; do not remove a selector or widen pair
admission to disguise that difference.

## Acceptance to review

- Go and JS parity for explicit All, named literal `all`, current-only, empty,
  unreported, malformed/null/value-bearing selectors and current/legacy conflicts.
- Source plans retain every predicate/bound/budget except environment restriction;
  unsupported owners make zero source calls.
- Typed links, views, captures, CEL planning and held custody preserve complete
  questions and independent source clocks.
- Real OTLP rows spanning all observed categories, bounded high cardinality,
  explicit start/apply/narrowing, local Return, source/access refusal, keyboard
  and desktop/320/390 frames.
- Existing private invocation/ABA/retired callback guards remain intact; no new
  generic permission coordinator or durable readiness cache.

The [operator contract](../ALL_ENVIRONMENTS_LOG_ENTRY.md) records the supported and
refused owner matrix. Actual checks belong in [validation](../../VALIDATION.md).

## Independent source and focused checks

The implementation retains generic semantic resolution for environment-only
snapshots/identity comparisons. Query validation and compiler source lowering
enforce logs-only execution; inventory and trace requests validate span queries,
and native discovery/lowering explicitly reject All before source access.
Surrounding-record helpers and stored followups refuse wide row context.
Linked-population planning keeps its ordinary implicit companion semantics and
therefore refuses an unsupported wide spans companion rather than dropping it.

The selector schema admits All without widening the observed EnvironmentScope
schema. Native, inventory and trace schema references are narrowed separately;
log query, tail and package schemas express the All/logs condition.

Independent execution passed **13 new Go checks/subcases**, zero failures/skips,
across identity, exploration, measurements, workbench and record HTTP packages.
These include parameterized filters/bounds/budgets, mixed-label capture/reopen,
existing carrier round trips and unsupported routes making zero source calls.
An additional focused environment-regression run passed all five packages.

Independent JS execution passed **23 tests**, zero failures/skips, across All
semantics, existing environment protocol/flow, local history and picker behavior.
It covers named literal `all`, null/current-only/value-bearing refusal, All
membership independent of metadata, unavailable/oversized observation facts,
complete question links/history and unsupported-context helpers. No remaining
source-review blocker is identified.

## Final artifact inspection and parent-owned integration verdict

Independently inspected the completed `6785928db779` report: status passed,
cleanup true, no browser errors, 30 main steps and 32 configured-source SQL
requests. Its raw oracle contains 1,005 OTLP observations and 1,005 distinct work
IDs. The held Activity state before/after Return is identical. The fixed question
retains All and full count 1,005 while displaying 100 rows; its separate capture
retains 1,000 inspection records and reopens without a telemetry read.

The actual distinct-value request receives the existing budget refusal (502)
rather than a partial ranking. Otherwise valid wide spans, inventory, trace and
metric-discovery requests each receive 400 with zero SQL. Empty, Unreported and
literal named `all` handoffs each stage from actual resource facts, retain their
fixed predicates/window, and explicitly read the expected 201 records.

The final trail focus opt-in was also reviewed: it focuses Run only when the
returned API is still the current connected visit. It does not globally change
ordinary navigation or confer source authority. Fixed-row narrowing explicitly
resets inspection offset to zero without dropping its parent predicates/bounds.
The desktop entry and 320px fixed-result frames were inspected independently;
All remains an explicit scope, reported facts stay distinguishable, and the
1,005-total/100-displayed distinction is readable without horizontal clipping.

Parent-owned final integration: browser/actor/mandatory native passed in 1.4m
(1.5m runner), including 404 actor snapshots. The native log reports
`TestQuintAllEnvironments437Observed` pass in 15.74s, zero skips, with all eight
negative controls rejected. Source work is 31 main SQL requests plus one explicit
actor poll; an aged empty actor window is not represented as a complete source
census. The actual keyboard check/uncheck focus, Stage-to-Run focus and cleanup
results close the runtime gates. These results are separate from this review's
independent 13 Go/23 JS checks. No final architecture or source-gate blocker is
identified; physical AT and cross-engine acceptance are not claimed.

Parent completion: the final campaign additionally verifies the actual recalled
configuration through the existing UI YAML export. It passed 31 steps with 424
actor snapshots, no browser errors and complete cleanup; the final native file
passed in 18.25s with all eight negative controls rejected. This does not expand
the independent checks above. See [validation](../../VALIDATION.md).
