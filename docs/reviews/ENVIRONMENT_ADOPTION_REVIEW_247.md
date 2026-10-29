# Environment adoption architecture, Go and frontend review — 247

Architecture recommended a checked scope/diagnostics and separate read-only
adoption cut rather than adding a flag to general Query. Tail, inventory,
packages, context/followups, dashboards and capture comparisons copy or interpret
only the old environment string; enabling all continuations piecemeal risks loss
of selection. The implementation follows that boundary and shares row projection,
not a dummy named Query or rewritten SQL.

Go reviewed presence precedence, bound SQL parameters, source scope/time guards
and budgets and found no blocker. Frontend reviewed explicit reads, passive
inspection and ownership. Both noted that a label-combination count is not an
effective-preview total; the UI/contract now explain that distinction.

Frontend also requested focus reset when replacing previews and a defined owner
transition policy. The owner resets focus and invalidates pending reads when
Activity becomes inactive, in addition to disclosure/signal/window/visibility
changes. The browser task covers late responses after window edits and record/
measurement mode round trips. Inspection reuses shared focus/return primitives;
preview records reuse shared compact row construction instead of parallel cells.
Diagnostic table layout is scoped to its owner and does not style nested tables.

Reviews were read-only; reviewers did not edit files or run the root checks.
Executed tests and known remaining A0 work are recorded in VALIDATION and the
[contract](../ENVIRONMENT_ADOPTION.md). This is not a full accessibility audit,
production load claim or completed main-flow environment migration.
