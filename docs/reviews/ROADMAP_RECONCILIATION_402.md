# Review after the 402 consolidation

Architecture/Go and design/usability agents independently reviewed the current
capability index, source and roadmap. The finite next cadence is adopted below;
402 acceptance and executed results remain in VALIDATION.md.

| Candidate | Decision and reason |
| --- | --- |
| Trace-search continuation | Adopt breadth: the existing `limit+1` search detects more IDs, but operators reach a dead end at the limit. Preserve explicit fresh-read semantics. |
| Ungrouped record/summary templates | Adopt depth: 400 supports authoring compact summaries, but template admission still requires grouping and rejects summary children. Extend the existing grammar and owners. |
| Shared ClickHouse parameter encoding | Adopt consolidation: exploration escapes all ASCII controls, while measurements has a different smaller replacement list. Share the pure wire encoder with cross-family round-trip proofs. This is an inconsistency, not a confirmed corruption bug. |
| Host sample context | Defer: matching resource logs/spans already exist over the full fixed host window. A sample-centered arm needs a concrete operator need and independent host.id/sample gates; stocks do not inherit service-gauge admission. |
| Grouped P95 reuse | Defer: raw quantiles exist, but reuse still needs independent raw-cell/population quantile support. Mean support is insufficient. |
| Passive grammar entry/gallery | Defer until the finite cadence closes; useful embedding discoverability, no need for a registry, bundler migration or executable plugin system. |

Recent lessons: leaf validation cannot establish enclosing carrier admission;
physical field-discovery SQL counts separately from a conceptual record request;
scalar environments retain their existing current-only meaning; draft edits
permanently retire staging; Back renews only unchanged retained context. Layout
geometry also needs visual review: a sticky cell can meet position assertions
while becoming an unreadable one-character column.

No broad rewrite, generic cloning abstraction, speculative source concurrency,
framework migration or language-feature quota is justified. Go 1.27.3 remains the
current toolchain. Use modern facilities for demonstrated benefits; keep public
APIs ordinary and typed.
