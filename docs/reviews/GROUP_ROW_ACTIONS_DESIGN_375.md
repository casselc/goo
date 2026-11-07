# Group identity and actions on narrow screens — design 375

Status: source-only readiness proposal; candidate and actual matched frames pending.
No new responsive framework, data model or acquisition behavior is required.

## Read and act on the same row

370's real narrow frames needed horizontal panning: one crop showed service/type,
another counts/actions. The user should instead read an exact group identity and
confidently Inspect or Compare that same group at 390 px. Reuse the existing
`goo-table[mobile-layout=stacked]` opt-in and data-label mechanism for shared
grouped-record rows; desktop retains normal table headings/columns.

Each row keeps every grouping label, full captured count and primary continuation
together. Long and Unicode labels wrap; Missing key, Present empty string and
literal lookalike strings remain visibly distinct. Do not truncate the only
identifying portion or replace the full literal used by links. Grouping headers
become visible per-cell labels when native table headings are visually hidden;
preserve native table semantics and exact header associations.

Use existing action text, Inspect group then Compare within service. A compact
wrapping action area is preferable to tiny stacked word-by-word links. The primary
action should have a usable target/focus outline; do not solve width by reducing
text or pointer targets until their meaning is lost. Count is a count of matching
records, not unique issues/requests or the number of mounted rows.

## Scope and interaction constraints

This is presentation only. Preserve row DOM/action references, current-group identity,
question generation and existing renewed action leases. No synthetic rows, source
reads, sort changes, selection reinterpretation or automatic child opening. Held
Back restores the same current row and focus; old detached/hidden/replaced callbacks
remain inert. Responsive layout alone must not remount the question owner.

Ordinary and typed triage should use the same grouped-row presenter. Optional
numeric/duration cells must retain labels and exact unknown/not-recorded meanings.
If a secondary wide diagnostic table cannot fit, keep its deliberate bounded
horizontal inspection rather than silently hiding a measure. Do not sweep unrelated
record/raw/native tables or redesign the whole app.

## Genuine matched acceptance

At desktop and 390 px, show ordinary grouped records and typed service/exception
groups with long/literal/missing/empty labels. Measure a representative complete
row's identity, count and primary action bounds in one viewport; no documentation
pan should be needed to connect them. Verify keyboard Tab to actual current link,
focus center visibility, explicit staged child then Run and held Back with source0.
Exact membership/receipts and old-callback negatives remain independent native
proof, not inferred from screenshots. A row may be taller to wrap a genuinely long
identity; no requirement says all groups fit one screen.

374's exponential-histogram gate exposes a material mathematical/source boundary.
Any future representation should receive its own explicit design review before
new kinds or conversion promises;375 does not invent or implement that model.
