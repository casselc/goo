# Saved trace-search intent review — 382

Read-only architecture/usability review of the implemented source candidate.
This review ran no heavy tests, browser campaign or source query; actual checks
are recorded separately in [validation](../../VALIDATION.md). The reviewed
candidate has no remaining practical source-review blocker.

## One typed question, not another recipe registry

`workbench.Step` now admits an explicit `trace-search` arm carrying the existing
`exploration.TraceSearchRequest`, with `Traces` presentation. The complete ordered
condition set, qualified fields, exact numeric strings, environment policy,
fixed bounds and list limit remain request intent. The payload must match the
step scope and cannot mix record/measurement arms or native nanosecond scope.
The existing investigation revision/card store remains the persistence owner.

The new wire guard rejects explicit nulls, missing required scope/request fields,
unknown or case-aliased typed object keys and malformed Unicode before Go scalar
zero defaults can acquire different meaning. Normalized last-value JSON behavior
avoids merging earlier duplicate object occurrences into a different request.
User-selected attribute/resource names remain string values and are not restricted
to the canonical typed-object key set. JavaScript uses the existing pure predicate
and numeric validators, with closed request/group/field admission.

## Save and reopen without an implicit read

The existing contextual search panel gains one save action and explicit list-limit
control. Saving stages an investigation question card; publication still belongs
to the existing document revision flow. Reopening uses the existing fixed visit
and trace-search owner with a staged standalone mode. Ordinary record controls
are hidden/inert, so a saved cross-span condition set does not become a runnable
all-span predicate or silently depend on a previously applied record result.
Search remains explicit and uses the complete retained request.

`InvestigationTrail` participates in search-step dispatch and fixed-entry ownership;
changing only the protocol union would have left record-only hash handling and
Activity entry behavior incomplete. Existing record and measurement dispatch
remain separate. Standalone Close restores a visible draft control's focus only
for an explicit active user close; navigation/invalidation does not steal focus.

## Bound every existing fallback

Existing question-card labels, population descriptions, destination choices and
capture conversion previously treated every non-record question as a measurement.
Those fallbacks now branch explicitly. Search question cards do not offer Run and
capture, cannot accept linked capture evidence, and cannot pass through the
record/native capture-question converter. Search evidence and qualified witness
custody belong to the separate 383 milestone; no completeness or causal relation
is inferred by saving intent.

The old 48,000-character step-link limit could admit a legal saved condition set
that could not reopen. New search intent is bounded to the existing 64 KiB request
budget in both Go and JavaScript, including Go JSON escaping, and search links
have a bounded 256 KiB allowance. Large valid round trips and oversized requests
have explicit checks. Ordinary question/link budgets retain their existing rules.

Reopen callbacks bind their rendered document/context lease and require a current,
contained, connected, visible, non-inert control. Search Save/Run callbacks are
rebound by revision; editing, closing, navigation and later scopes retire earlier
actions and pending publication. The existing owners retain cancellation and
source acquisition; no new executor or universal navigation manager was added.
