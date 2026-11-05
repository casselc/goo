# Follow a kept-content witness — design review 331

Status: final source and actual frames accepted after the content-order correction.
Root independently reviewed final native and nonvacuous actor evidence; both pass.

## Smallest interaction

Keep **Open kept observations** as ordinary receipt navigation. On content-search
rows only, add **Open matching record** beside the existing witness in its Content
match cell. Use one compact action and normal wrapping/stacking at 390px. Do not add a
column, menu, picker, repeated excerpt or persistent explanation bar. The existing
Record N / field witness and bounded excerpt already explain the direct action.
Use a button for an app-owned ephemeral member handoff; do not present a direct
member link whose modified activation silently becomes ordinary receipt opening.
Metadata results retain only ordinary Open. The direct action promises the first
matching member identified by that row, not every matching record in the receipt.

The ordinary action opens the receipt table without requesting a member. If that
same receipt is already retained, existing local inspection may remain. The
second action opens that same verified
receipt and immediately reveals its existing record inspector with keyboard focus
on the inspector heading. Closing returns focus to that member's **Inspect kept
record** button in the receipt table; it must not unexpectedly navigate back to
search. Existing trace/context continuation and retained return stay in that same
inspector and preserve the chosen member. A concise Record N identification in the
inspector is useful if it fits existing provenance; no new navigation surface is
needed.

## Exact identity and refusal

Use the verified search witness's zero-based `memberIndex` as a receipt-scoped hint.
Record N in the UI is that index plus one. Validate the hint against the immutable
loaded receipt before revealing any member. Receipt `members[i]` and result-page
member `index` are different: the latter is the original result occurrence/page
position, potentially sparse and nonzero. Do not convert one into the other,
search by excerpt/field value, or choose the first superficially similar record.

Invalid, malformed, missing-target or out-of-range hints fail explicitly. Show
**Requested kept record unavailable** with the reason; do not silently land on
record 1 or show an old inspector as though it satisfied the request. A valid
receipt can remain available as clearly identified previously opened evidence,
but ordinary opening is a separate explicit action. Missing/corrupt receipt errors
remain receipt errors. No fallback source read, custody mutation, repair or new
content query is implied by direct opening.

## Lifecycle and compactness

One route/navigation intent performs one exact-ID verified receipt GET even when
both hashchange and popstate report it. The admission guard may reuse the current retained same-ID receipt; it must
retire on leaving its route and must not become a general receipt cache. Explicit
retry after failure creates a new intent. The same receipt with a different
member hint must focus the new member. Superseding opens, other routes,
close and hidden/inactive owners retire pending hints and late responses. Retired
search-row links remain inert under the existing generation/visibility guards.

The catalog caption should continue explaining that original records are
inspected. The direct action supplies the new precision without bloating the
primary status or Search details. Keep the original excerpt as text and preserve
long literal/ID wrapping. Focus and activation must work by keyboard and touch.

## Acceptance to review after implementation

Review source ownership and actual desktop + 390px frames showing the two compact
actions, a direct nonfirst log member, and a direct span member. Confirm original
record payload/provenance, heading focus, Close focus on the exact row button,
ordinary Open without automatic inspection, and existing continuation return.
Native assertions should count exactly one receipt GET per open intent and zero
telemetry/custody work; exercise actual held late responses and superseding routes.
Generated/unit controls should distinguish sparse original-page indexes from
member indexes, duplicate-value occurrences, invalid hints, close retirement and
same-receipt/different-hint transitions. Record native and generated scope
separately. Screenshots alone cannot establish exact identity or IO counts.

## Following consolidation checkpoint

After 331, review one complete entry → scope → compare → preserve task across
Activity and the population/saved workspaces before selecting the next milestone.
The recent delivered work provides useful concrete seams: typed scope policy now
crosses record/native paths (318–320); staged previous-window comparison and saved
baseline continuity retain different read authority (321–322); passive count-pair
presentation was consolidated (323); loaded Find and preservation share exact
references (324–326); kept continuation and search added route/focus ownership
(327–331). Review how the user recognizes the active scope, staged versus read
state, frozen versus reread evidence, and return destination across those tasks.

Ground the checkpoint in current adapters and short running-app passes, recording
specific duplicate ownership or ambiguous controls and their observed cost. Choose
one broader structural/usability improvement that benefits these flows if evidence
supports it. Keep acquisition and retained-evidence owners distinct; shared passive
presentation and strict handoff admission are useful candidates only where they
remove actual duplication. A generic navigation framework, receipt cache or another
kept action is not established by 331. If its route/hint lifecycle is fully resolved,
close that gap rather than prolonging the kept-only series.

Goo remains prerelease: remove obsolete routes/contracts instead of adding
compatibility adapters. Do not turn this checkpoint into a parity catalogue or
another layout-only iteration. Existing deferred deployment, retention and alert
contracts remain separate product decisions. This recommendation expands no 331
implementation scope.

## Final source review

The implementation uses a button in Content match beside the witness, preserving
ordinary receipt links and avoiding misleading modified-link activation. Both use
the captures adapter's exact-ID owner. Pending/loaded/failed visits deduplicate the
paired browser route events; explicit retry after failure is distinct. A requested
member indexes the verified receipt array and invokes the existing row inspection
path. `inspect` reports actual success, `clear` retires its button collection, and
failure produces a visible matching-record refusal without selecting record 1.
Inspector heading focus and Close-to-row behavior reuse InspectionFocus. No new
inspector, persisted hint, source query or custody write was introduced.

No source design blocker remains. Same-ID retained reuse is accepted as a current
visit behavior, not a general cache or a fresh verification claim. Native single-GET,
late-response, exact-member and focus assertions plus actual desktop/narrow frames
are confirmed in the final review below. Architecture reports twenty generated
owner scenarios and two mutation negative controls; these supplement native checks.

## Actual frame review and corrected content order

Final `331-matching-result.png` and `331-matching-member.png` are accepted.
The result shows three selected span members but one match, **Record 2 · SpanName**,
a bounded excerpt, the compact direct button and separate ordinary Open. The
initial inspector viewport now visibly presents **Literal331 span 8508a83b** and
**50.000 ms** before kept provenance. Close, trace continuation and the complete
basis panel retain their existing composition. The source inserts provenance
after Records.detail's original summary; no screenshot-only scrolling is needed.

Root identified the initial content-order defect after my first visual review:
provenance displaced SpanName below the internal viewport. That earlier inspector
acceptance was withdrawn. The minimal kept-only reorder resolves it without
removing any basis facts or adding a widget. Final native assertions explicitly
check summary geometry at scroll zero, exact member 1 and raw/DTO immutability.

All three refreshed 330 frames are accepted. The positive 390px frame shows one
span match; scope, witness, wrapping button and ordinary link remain readable
without horizontal overflow. Desktop log markup remains literal text. This is
the current narrow evidence, distinct from 330's historical two-result capture.
All six shared 327/328 frames were refreshed; renewed original inspector frames
visibly show the named/empty-namespace Body values before full provenance, with
compact actions and readable no-trace refusal. No visual blocker remains.

Root reports final main native checks passed for nonfirst log/span members,
one-GET opening, Close focus, ordinary Open, held late-response rejection,
summary geometry at scroll zero and receipt/DTO immutability. The first actor
attempt produced an empty trace and was correctly rejected as evidence; final
nonvacuous actor acceptance passes after root raw review: 552 actual state
snapshots across five phases, complete `visibleOriginal=true`, one verified
receipt/search/read, zero source reads/writes, and no bad states or violations.
The actual query is `Literal331 span 8508a83b`. Final wrapper runs for 328
(25.2 seconds) and 331 (45.0 seconds) pass together: two tests in 1.2 minutes.
Source, actual frames and final design acceptance are complete.
