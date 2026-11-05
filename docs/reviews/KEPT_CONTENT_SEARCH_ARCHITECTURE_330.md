# Bounded kept content search architecture review — 330

Status: source review complete; focused generated verification passes. Running-app
acceptance and full-suite evidence are recorded separately in VALIDATION.md.

## Authority and work bounds

`KeptObservationStore.Search` owns acquisition. Optional `text` is an exact,
case-sensitive literal with valid UTF-8 and at most 256 bytes; empty text retains
metadata discovery. Content reads verified immutable receipts, never telemetry
adapters, captures or source queries. It neither repairs the index nor rewrites
custody. Matching admits only string Body for logs / SpanName for spans, then
ServiceName, TraceId and SpanId. Other fields and the other signal's primary field
cannot establish membership. Each original member counts once even if multiple
fields match; repeated members retain occurrence multiplicity.

The catalog remains bounded to 1,000 admitted physical entries with bounded
metadata. A content request verifies at most 64 receipts and reads at most 32 MiB
of receipt bytes, returning at most 20 matching selections. Metadata traversal is
additional bounded work, not represented as content checked. Metadata eligibility
is computed before content work. Total and eligible describe the catalog basis;
checked, bytes, matched and matchingRecords describe the current content page.
The metadata-only branch preserves its existing whole-filter matched count and
reports eligible metadata as checked, with no content bytes or match witnesses.

The next receipt is deferred if it cannot fit the remaining byte budget. Every
admitted receipt is at most 3 MiB, below the fresh request budget, so a nonempty
remaining candidate set always advances before publishing continuation. This
relationship prevents a zero-progress cursor and makes the last-checked anchor
safe. Keep these constants related if future receipt limits change.

## Integrity and continuation

Every candidate is completely loaded through the existing strict receipt verifier
before matching. Its recomputed summary must equal catalog metadata before a
match is attached. Missing, unreadable or corrupt candidates fail the page; they
cannot become verified nonmatches. The catalog explicitly rejects a persisted
`summary.match`: transient content evidence cannot become derived-index authority.
Same-size, same-mtime corruption still fails receipt digest verification when the
candidate is reached. Corruption beyond the current bounded page is necessarily
unexamined until continuation; the UI must not infer global absence from one page.

The deterministic order is descending immutable keptAt with ID tie-breaker. The
cursor binds exact admitted query bytes and metadata filters, the catalog checksum
basis, and the last checked ID. It advances across nonmatching receipts and may
accompany zero results. New publication or metadata/basis changes refuse old
continuations. No content-match total is inferred for unvisited candidates.
Directory locking coordinates cooperating store writers; this is not protection
against an administrator changing filesystem entries outside the store protocol.

## Witness and application lifecycle

A witness selects the first matching original member and first admitted field in
the documented order. Its excerpt preserves the complete literal plus at most 32
Unicode code points on each side. A 256-byte UTF-8 literal contains at most 256
code points, proving the 320-code-point bound. Rune boundaries keep excerpts valid
UTF-8. The bounded excerpt is transient; Open uses the existing original receipt
inspector. The accepted contract uses a bounded plain excerpt without omission markers.
The excerpt establishes a literal match, not field completeness; the original
record inspector remains the authority for the entire preserved value.

The app adapter retains generation and AbortController ownership. Input changes
retire old responses, clear results and paging, and enable explicit retry. Hiding
discovery, workspace navigation and keep-change events also retire old controls.
Previous replays the stored page cursor; More uses the server's last-checked
cursor even for empty matching pages. Text is rendered through text nodes, not
HTML. Search details expose field and work bounds without another navigation or
source-read owner. Explicit repair remains a distinct metadata operation.

## Verification

`verification/kept_content_search_property_test.go` uses Hegel to generate 20
catalogs with 21–75 published receipts, one to three members, both signals and
signal filters. Its independent source-record oracle covers literal punctuation,
case, meaningful spaces, Unicode, excluded-field decoys, member multiplicity,
first-member/field witness choice, bounded Unicode snippets, page-local accounting,
and disjoint/complete continuation. A changed-query negative control proves
continuation binding rejects reuse. The focused run passed in 5.839 seconds with
`XDG_CACHE_HOME=/workspace/.cache` (the default home cache is read-only here).
Backend native tests separately exercise no-hit continuation, byte/receipt/result
bounds, corrupt receipt rejection and derived-match index rejection. Full-suite
and browser outcomes must be taken from their actual recorded runs.
