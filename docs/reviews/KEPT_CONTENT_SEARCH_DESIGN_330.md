# Find preserved observations by content — design review 330

Status: final source and all three actual frames accepted. Main native and actor
checks passed as reported and independently reviewed by root. No product/source
edits were made for this design review.

## Smallest useful task

Add one optional **Record text** input to the existing kept catalog form, beside
origin and signal. **Find kept observations** explicitly applies all three values.
An empty text value retains metadata discovery. Typing, clearing and changing a
filter invalidate the displayed task without reading; no per-keystroke search.
Reuse the existing table, paging, Open action and verified kept-record inspector.
Do not introduce a second search workspace, index, source query or navigation owner.

Adjacent help states: **Case-sensitive literal text in Body (logs), SpanName
(spans), ServiceName, TraceId and SpanId. Searches kept records only.** Match actual
reported string values, excluding attributes, serialized JSON, timestamps,
placeholder labels and inferred values. Do not normalize case, whitespace or
Unicode; canonically equivalent spellings may differ. Reject malformed Unicode
before transport. Preserve meaningful leading/trailing spaces. An empty string
means metadata search; a space is a literal query.

## Bounded work and counts

The approved backend envelope is 256 UTF-8 query bytes, at most 64 receipts and
32 MiB of kept bytes inspected per request, and at most 20 matching selection rows.
These are request work bounds, not a count of all matching records on the server.
The implementation contract must define whether a nonfitting next receipt is
left for the next request and how an individually over-budget receipt refuses;
it must always progress or return an explicit error, never loop on one cursor.

Each candidate receipt must pass its existing complete verifier before any of its
fields can contribute a match. A corrupt, missing or unread receipt cannot count
as a verified nonmatch. Any corrupt or unread candidate fails the whole request;
show a clear message that the search could not establish absence. Searching must neither repair metadata nor refresh
kept bytes nor read source telemetry.

Report page-local matching selections separately from matching records, plus
verified receipts/records checked and kept bytes checked. Total catalog inventory and metadata-eligible selections remain separate from
page-local checked/matched/byte counts. Metadata eligibility is not verified
content coverage or a content-match denominator. Suggested status:
**2 matching selections · 3 matching records · 12 selections checked on this page
· More selections remain.** If record-check totals are omitted from the protocol,
state selections checked only; never invent records inspected from matched rows.
A selection that matches three fields on one member contributes one matching
record. Repeated preserved members retain their original occurrence semantics.

A zero-match page with continuation says **No matches in the selections checked
on this page. More selections remain.** Keep More enabled. Exhaustion establishes
only no further matches in the fixed, admitted catalog basis and documented fields;
it does not establish absence in telemetry, other servers or newly kept receipts.
No receipt can be represented as inspected merely because its metadata was scanned.

## One useful witness per selection

Retain ordinary origin/signal/scope/custody columns and add compact content evidence
within the Selection cell: **3 matching records · first: record 2, Body**. The
wire witness includes the first matching original member index, canonical field
name and one actual, safe bounded snippet. Keep the entire literal match and at
most 32 Unicode code points on either side: the query is at most 256 UTF-8 bytes,
so the resulting snippet is at most 320 Unicode code points. Include explicit
ellipses when surrounding original content is omitted. The snippet must be an
actual portion of the verified field, preserving literal whitespace and Unicode;
do not use a substituted or normalized match. Never replace the preserved
original with the excerpt or expose unbounded values in a table cell/title.
Render all values as text, preserving witness whitespace visibly where needed.

Choose witnesses deterministically: original member order, then documented field
order (Body or SpanName, ServiceName, TraceId, SpanId). Count every matching member
in a verified receipt even when only one witness is shown. **Open kept
observations** uses the existing receipt and inspector; avoid promising direct
member focus unless the current inspector safely supports that handoff. A field
name/member ordinal is useful evidence without adding a navigation grammar.

## Cursor and lifecycle

Bind continuation to normalized admitted filters, literal query bytes, fixed
catalog basis and deterministic newest-kept ordering with the existing tie-breaker.
Anchor the cursor after the last checked candidate, including nonmatches, rather
than after the last displayed result. An empty matching page can still progress.
A new keep invalidates the displayed task and requires explicit Find; never append
it to an in-progress search silently. Stale/reordered basis refuses clearly rather
than mixing pages. Previous must replay the same page cursor and bounds.

Preserve existing generation/AbortController/visibility guards. Changing filters,
leaving the workspace or hiding discovery retires old responses and paging/Open
controls. Editing during an outstanding request must allow explicit retry after
invalidation; retired responses cannot repopulate results or restore old counts.
Existing repair remains an explicit separate action and must not inherit the
content query as a promise that repair searched records.

## Acceptance required

Inspect real desktop and narrow frames for the form, positive results with one
witness, and zero/incomplete results. Long literal text and IDs must not widen the
table or obscure Open. Verify keyboard labels, status announcement and focus.
Native assertions cover genuine kept logs/spans, all documented fields and literal
punctuation/space/case behavior, source-free search, refusal of late actual
responses and exact original inspector opening. The narrow 390px frame
supplements the desktop integration frames. Multiple matching members,
no-match continuation, corrupt/unread receipts and stale cursors/new keeps are
Go unit/generated/store checks; they do not require a huge native fixture and
must not be reported as native coverage. Ledger search reads as kept-content IO;
assert zero telemetry reads, custody writes and implicit repair. Mocked frames
cannot substitute for the running integration task. The matchingRecords maximum
of 400 is a schema bound (20 returned selections × 20 members), not a claim that
that many matches were observed in the native fixture.

## Source wording handoff

The current compact status is accepted: **N selections · M matching records ·
Checked C / E eligible this page · More remain.** C is this request's checked
selection count; E is the metadata-eligible catalog count, not additional verified
content. With exhaustion, **No more pages.** closes this finite basis only.
The content caption correctly reads **Verified kept record matches; Open inspects
original records**, while metadata-only discovery retains its separate caption.
Bytes, catalog size, bounds and field/case rules are available in the closed
accessible **Search details** disclosure, with title text as an additional hint.
The disclosure preserves compact primary status while exposing the rules to
keyboard and touch users. The witness uses **Record N** and content-request
failure explicitly says **Absence is not established.**

The witness text is accepted as a bounded excerpt without completeness authority.
The Content match heading, original record ordinal/field and explicit Open action
provide sufficient context for this round; omission flags or additional API fields
are not required. Calling the witness an **excerpt** in supporting documentation
is required to avoid implying that the complete original field is displayed. This
is a documented bounded-excerpt contract, not a source blocker.

## Final actual frames and verification

All three maintained actual frames are accepted:

- `docs/journeys/screenshots/330-kept-content-witnesses.png`
- `docs/journeys/screenshots/330-kept-span-search.png`
- `docs/journeys/screenshots/330-kept-content-narrow.png`

The desktop views show the same final genuine receipt fixture `88e50c3a`, compact
form/status, verified-content caption, readable Body/SpanName excerpts and visible
Open actions. Markup in the log excerpt renders as literal text. The positive
390px narrow view stacks both real matches without horizontal overflow; scope,
record/field witness, literal excerpt and Open remain readable. Search details
stays separate from primary task status. No visual blocker was found.

The narrow capture is a targeted supplemental check using the two genuine final
main-run receipts copied byte-for-byte. Copying changed file mtimes, so the derived
index was repaired explicitly before discovery; no source reads were made. This
capture demonstrates positive narrow composition, not another real-ingestion or
source-acquisition run.

Root reports main native and actor PASS in 42.5 seconds (44.4 wrapper), with raw
review of 563 snapshots across ten phases, three verified actual DTO responses,
zero source reads/writes during the relevant search ledger, and exact original
Open. Existing Open performs two same-ID verified receipt reads; this duplicate IO
is recorded as a separate known issue, not a claim that Find reads telemetry or
performs custody writes. Native versus generated/store coverage remains separated
as described above. The schema maximum of 400 matching records is a protocol bound,
not evidence of a 400-match native fixture. No further design changes are requested.
