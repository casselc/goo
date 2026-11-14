# 437 All-environments logs: UX and accessibility review

Status: final production source, focused checks and all nine responsive runtime
frames accepted. The parent owns complete browser/native acceptance. This review does not
claim physical assistive-technology or cross-engine testing.

## Entry contract

The existing Activity toolbar has a named environment input, a typed-selector
chip, Find data, signal choice and explicit Start/Apply. An operator should be
able to choose **All environments** beside these controls without first knowing
labels or reading a long explanation. Use a native scope select or checkbox;
keep one explicit primary source-read action.

Choosing All stages a typed scope. It does not start a query. Blank input,
unavailable discovery and malformed scopes must not choose All automatically.
Typing the literal name `all` remains a named environment. Empty labels,
unreported labels and All are separate concepts.

While All is selected, disable or hide the named input so it cannot contradict
the chosen scope. Preserve an honest path back to a selected environment; do not
quietly replace an unsupported All selector when switching to spans or another
surface. Refuse before a read with a concise visible reason and a clear action.
No new authentication or cross-source scope is implied by this entry.

## Observation context

All is a query scope, not a reported environment observation. Show neutral,
compact reported-environment facts on log rows under All. Named values should
remain literal; a producer's name `all` must not resemble the query-scope label.
Keep empty and unreported visibly distinct.

Use the existing current-key/legacy-key precedence, including an explicitly
empty current key winning over a named legacy key. Conflicts and full keys can
remain in existing detail/tooltip presentation. No environment name implies
health, ownership or authorization.

Unreported is justified only when the canonical resource-attributes map is
present and valid and neither environment key exists. Missing or malformed
resource metadata is unavailable context, not proof of Unreported. A row action
cannot construct a narrower query from a display label.

Pills should retain literal accessible DOM text. Avoid unsupported ARIA naming
on generic spans or a new tab stop on every row. Existing record inspection can
expose complete reported keys and precedence. CSS ellipsis must not erase the
underlying full text.

## Scope changes and context actions

Apply/Start/Run remain the source-read owners. Opening controls, choosing a scope,
staging a fixed question, changing a local display and returning to held data
must not acquire telemetry or publish artifacts.

Narrowing should retain the original filters and fixed bounds, then explicitly
add or select the intended typed environment. An unsupported trace, service,
native or span continuation must visibly refuse before its source call. It must
not interpret All as a blank scalar environment or silently choose one row's
display name.

The exact selector must survive links, Recent/Pinned, saved definitions and
supported captures. Pending/departed callbacks and scope A → B → A remain owned
by existing invocation guards. Failed or incompatible adoption retains the
current draft and held observations.

## Acceptance still required

- Actual logs from named, empty, unreported, conflicting and literal-`all`
  producers; clear row context and exact typed scope.
- Source-inert selection/staging, explicit reads, refusal and retained return.
- Keyboard choice, focus after apply/refusal/return, native disabled/hidden state
  and current versus retired callbacks.
- Readable 1440/390/320 layouts with no horizontal overflow or verbose controls
  preceding observations.
- Exact links/recall/captures where supported, and before-read refusal where not.

The parent records executed integration/model results in
[validation](../../VALIDATION.md). Final source and frame findings will be added
after implementation is ready.

## Implementation review

Independent focused execution currently passes **37 tests**, zero failures/skips,
across All-environments semantics, record intent, shared record presentation,
Activity entry and passive record lists. This is separate from the parent-owned
real browser/native acceptance.

The first source pass correctly preserves selector-only All, current/legacy
precedence, unavailable resource metadata and exact typed narrowing. All is never
inferred from a name or failed discovery. Unsupported span, trace, surrounding,
service and native paths refuse before execution. Choosing Spans retains the
unsupported All draft and visible reason; the checked All control remains
clearable so the operator can choose an explicit narrower scope.

Review findings returned to implementation:

1. Guard the retained hidden application Apply callback after All staging hides
   it in favor of Tail's Start control. The final callback now checks native
   readiness and current context, with an actual production-handler regression.
2. Disabled trace/surrounding controls need a concise visible explanation for
   keyboard users, not only hover titles. More actions and record inspection now
   explain narrowing beside the supported Stage action.
3. Keep the disabled name field's description truthful under All, preserve a
   separator between service and environment text, and shorten the row pill's
   named prefix so ordinary environment names remain distinguishable. These
   bounded changes are implemented; refreshed frames remain pending.
4. A checkbox change must retain checkbox focus. Reusing the explicit Stage-button
   helper currently moves focus to Start, allowing a second Space intended to
   toggle the checkbox to trigger a source read. A checkbox-specific no-focus
   staging path was requested. The source now keeps checkbox focus on both
   check and uncheck; real Space/Start proof remains the browser acceptance gate.

The first actual fixed All frames at 1440/390/320 were inspected. Scope wording,
1,005 total versus 100 displayed records and distinct empty/unreported facts are
readable without horizontal overflow. Named pill text was unnecessarily clipped
by its long prefix; the shortened prefix needs final frame confirmation. Initial
Activity choice/keyboard frames and the complete driver/model verdict remain
pending.

A real fixed-row narrowing run also found focus left on an inactive opener/body
after creating the new visit. The opt-in Stage action now focuses the returned
Run control only while that mounted API is still current and connected. A focused
callback test covers positive adoption, unrequested focus and reentrant replacement.
This remains a no-read staging action; the final browser run must confirm focus.

The refreshed fixed All frame at 320 pixels now shows complete ordinary named
environments with the compact Env prefix; empty and unreported labels remain
readable. The shared presentation still retains full literal DOM text and detail.
An independent All/trail rerun passes 11 tests, zero skips, including current-only
focus adoption and reentrant replacement refusal.

Disabled linked-population and recent-window shortcut controls also need concise
visible refusal text near their existing controls. This is a contextual hint, not
new prose before the log stream; implementation is addressing it before final
acceptance.

## Final acceptance

All nine final entry, paused Activity and fixed-record frames from the 18:59
binary were independently inspected at 1440, 390 and 320 pixels. No horizontal
clipping or unreadable controls remain. The native checkbox is 14×14 pixels,
with a clickable label height of 32 pixels on desktop and 44 pixels on narrow
screens. Its inline alignment, native focus outline and separate Start control
fit the shared toolbar. Named pills retain readable literal values; Empty and
Unreported remain distinct, including the named value `all`.

The actual completed report `6785928db779` was read independently: status passed,
cleanup true, errors empty, and 30 recorded steps. Checking/unchecking by Space
retains checkbox focus and changes neither telemetry reads, writes, metadata nor
SQL. Spans refuses the inherited All draft without a source call and leaves an
explicit clear/narrow path. Stage from Empty, Unreported and named-`all` rows adds
no telemetry read or store write, then focuses the connected new Run control.
Legitimate staging metadata GETs are retained in the ledger rather than described
as zero API work.

The fixed view clearly separates 1,005 full matching records from its 100 displayed
rows. The existing capture bank remains bounded at 1,000; neither page nor bank
size becomes a completeness claim. Paused Activity shows its loaded/evicted counts
and continuity qualification. Disabled wide-context actions now have visible
local refusal text, and the supported typed Stage action remains available.

The parent reports the final browser/Bombadil/native campaign passing with 404
actor snapshots and a 15.74-second native gate, zero skips and eight negative
controls rejected. These campaign results are not inferred from screenshots;
executed details are maintained in [validation](../../VALIDATION.md). All source,
keyboard-focus and presentation findings in this review are closed. Physical AT
and cross-engine testing remain outside this bounded review.

Parent completion: the final campaign additionally verifies the actual recalled
configuration through the existing UI YAML export. It passed 31 steps with 424
actor snapshots, no browser errors and complete cleanup; the final native file
passed in 18.25s with all eight negative controls rejected. This does not expand
the independent checks above. See [validation](../../VALIDATION.md).
