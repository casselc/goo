# Find within held Activity records — design 324

Status: final source and all three actual frames accepted; final native/actor exit is reported separately by verification.
Repository guidance/current roadmap were rechecked. Read with the existing
Activity stream, record-column and exact-selection contracts.

## User task and finite scope

A responder scans loaded logs/spans and wants to find an observation without
changing the source query, waiting for another poll or losing selected evidence.
Add local **Find loaded records** to the existing compact record/field controls,
not another verbose toolbar or query-language editor. This is a local presentation
projection over the bounded supplied buffer, not full-text source search.

Show **N of M loaded records match** adjacent to the input, separate from the
existing observed-in-connection/evicted counters. M is the buffer before Find;
N is the matching supplied occurrences. Neither is a full query-population total
or upstream completeness claim. A zero-match state says **No loaded records
match**, not No logs or No telemetry.

## Deliberate search meaning

Final source contract: search **record text, IDs and chosen values**. Core log
body/span name, service, actual severity/status, resource service namespace,
human UTC/duration and valid trace/span ancestry IDs remain searchable even when
clipped or omitted from selected core columns. Chosen qualified columns add their
present actual values. This is deliberately broader than visible pixels and
narrower than arbitrary hidden JSON; the compact hint names that boundary.
Do not search action labels, retention metadata, Missing/Null/Empty placeholders
or synthetic unset severity. A literal supplied string `Missing` remains data.

Use one deterministic case-insensitive literal substring, with no regex, quoting,
field-name grammar or implicit AND/OR terms. Metacharacters and quotes are data.
An empty query means all loaded records; Clear sets that empty query. Missing
fields are not the literal word Missing/null, and a present empty value is not a
synthetic quoted string. Exact/missing-value scoping remains the existing typed
filter task, not a new Find syntax. Human UTC value formatting may be reused for
a displayed timestamp field; raw nanoseconds are not the primary user language.

Changing chosen qualified columns reevaluates the same local query against the new
field set and updates the matched count. It does not fetch fields, edit source
predicates or silently expand to all attributes. Exclude action/control text from
the search projection so Find Inspect cannot match every row.

## Focus, clear and selection

Keep the input, result status and Clear in the existing controls. Native search
input behavior and explicit Clear should retain logical input focus; loading or
filtering rows must not rebuild that control on each keystroke. Use a polite live
status for matched counts, including zero, without announcing a new source read.
Respect keyboard/touch targets and visible focus using the shared control styles.

Hiding a nonmatching row does not deselect or alter its original occurrence,
inspection payload, read window, source basis, membership token or Keep intent.
Existing selected-evidence controls remain usable; report how many selected items
are hidden by Find in the current selection status if needed. Do not narrow a
pending Keep to visible rows or lose a selection merely because it stops matching.
Clear restores visibility, not evidence validity or expired capabilities.

Inspect, trace and service actions on a visible match use the original supplied
record/window and current owner lease. Hidden/stale row controls cannot execute.
A local Find must not mint a new source/result identity or turn row indexes into
durable observation identity. Return to a held paused parent restores its Find,
selection, scroll/focus and full original buffer without rereading.

## Streaming, eviction and reset boundaries

Find does not pause/resume the stream or create a request. New arrivals and
ordinary buffer evictions update N and M through the same bounded buffer; a
nonmatch is merely hidden, not evicted. Existing follow/scroll behavior must remain
predictable and typing focus stable. Keep the normal live/paused status visible
so changing matches are not mistaken for a static query result.

Pause and retained return preserve Find. Clear is local and immediate. A new
prepare, changed signal, applied source scope or explicit reset clears Find.
Same-scope manual Resume/reconnect preserves the query to continue the task,
although its existing connection lifecycle replaces the buffer and clears
selection. It does not claim to retain the old observations across reconnect.
Find never initiates reconnect or changes that existing lifecycle.

Find is not saved as an applied predicate, rolling selector, baseline reference
or durable evidence. This first cut needs no saved-search model, backend endpoint
or compatibility format.

## Acceptance

Use real loaded logs and spans with repeated values, case variants, punctuation,
chosen qualified fields and distinguishable hidden attributes. Verify positive,
zero and Clear matches with N/M counts and zero new source requests. Find only
matches the declared data fields, not control labels/placeholders/hidden JSON.
Change columns and verify local reevaluation. Empty/missing and literal syntax
must follow the stated semantics.

Select a record, hide it with Find, and prove the exact original selection/Keep
intent survives; inspect a visible match and follow its original trace/service
window. Pause/return preserves Find/focus/buffer. Stream new matching/nonmatching
rows, evict through the existing bound and verify counts without mislabelling
local hides as eviction. Exercise prepare/signal/scope reset and same-scope reconnect preservation, plus stale hidden controls. Update compact actual frames for
positive match, zero and hidden-selection continuity; source/lifecycle assertions
establish zero IO and unchanged query/observation ownership separately.

No product/browser edits are part of this design review. This is one ordinary
Activity task, not another facet, metric profile or broad component rewrite.

## Final source review

Accepted the existing record-controls integration: a labelled native search input,
local Clear, live matched/loaded status and concise search-boundary hint. No new
source controls, request owner or query language were introduced. Search is
case-insensitive per-value literal matching, bounded to 256 well-formed UTF-16
code units and 1024 UTF-8 bytes; invalid input has native validity and aria-invalid
feedback rather than changing the applied question.

The row projection retains original nodes and payloads. Hidden selection is
reported separately and dismissal falls back to a visible selected row or Find;
a focused row newly hidden by reevaluation returns focus to Find. Existing
visibility/lease guards remain authoritative for actions. Find/Clear and qualified
column changes use the same local reevaluation seam; append/eviction update the
loaded denominator. The control is 32px minimum with 44px coarse-pointer sizing.

This is source acceptance, not a claim that browser focus, screenshots, streaming
or source-budget assertions have passed. Final actual-frame review follows the
sole verifier's completed run; no product/browser edits were made for this review.

## Actual-frame review

Accepted the three actual 324 frames independently of the final runner/actor
exit. `324-literal-held-log-find.png` shows literal `[.*]` matching two of four
loaded logs, including an appended observation, rather than interpreting regex.
The compact control sits with Record fields; paused state and separate loaded,
evicted and connection-observed facts remain visible. These are supplied-buffer
counts, not full-population counts.

`324-hidden-original-selection-kept.png` shows zero of four loaded matches and
an empty records body while the independent selection panel says one selected
not shown, selection unchanged, and Kept 1 observations. The visible state does
not imply that zero matches removed the original selection. Exact original-ref
custody and the single Keep write are native assertions, not facts inferred from
the screenshot alone. Zero is expressed numerically; no source-absence claim is
made.

`324-filtered-span-held-ancestry.png` shows the original child observation's UTC,
3.000 ms duration and reported trace ancestry after inspection. The crop is the
inspector, so it does not visually prove the Find input or hidden parent row;
ID/UTC/duration matching, ancestry continuity, hidden-action refusal and retained
Find/rows with zero-read paused return are separately exercised by the native
journey. No visual accessibility or semantic blocker was found. Final source
budgets and actor outcome remain the verifier's report; no product/browser edits
were made and no additional layout changes are requested.
