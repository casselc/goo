# Follow a typed error population

Milestone 370 is implemented and validated with Go/UI checks, generated
membership/owner tests, genuine browser journeys, observed model replay and
positive actor completion. See the [storyboard](journeys/370-typed-error-triage.md)
and [validation](../VALIDATION.md).

From Activity, open **Errors & exceptions**, choose the reported-evidence rule
and type field, then **Open fixed triage**. Run that staged fixed window to see
service/type populations. **Inspect group** stages the exact selected service
and type. **Compare within service** stages the same evidence rule within that
service, grouped by type. It is not an error-rate comparison against healthy
traffic. Each child waits for explicit **Run question**.

```mermaid
sequenceDiagram
  participant User
  participant Activity as Paused Activity
  participant Parent as Held triage question
  participant Child as Staged child question
  participant Source as Record source
  User->>Activity: Errors & exceptions
  Activity->>Parent: Stage typed scope and completed window
  User->>Parent: Run question
  Parent->>Source: Read reported-evidence population
  Source-->>Parent: Service/type groups and independent read basis
  User->>Parent: Inspect group or Compare within service
  Parent->>Child: Stage complete checked child question
  Note over Parent,Child: No source read; parent result stays held
  User->>Child: Run question
  Child->>Source: Read exact child population
  User->>Child: Back to question
  Child->>Child: Suspend child and retire pending publication
  Child->>Parent: Restore held result and current focused link
```

The original typed environment policy, fixed bounds and nested predicates remain
intact. Namespace constraints stay predicates; a raw service label does not become
a globally unique namespaced identity. Missing keys and present empty strings
remain distinct for service/type selection. Reported error status, exception
events and structured error data are evidence rules, not proof that every
matching operation failed or that an incident exists.

Links belong to an actual admitted parent group and applied result. A creation-
time lease, current owner, connection, visibility and inert checks govern clicks.
Departure and replacement retire old callbacks, even if a node is later
reattached. Back rebinds the same held group anchors without a source read or
record repaint. It restores only this group-action capability, not all revoked
chart or query actions. A late child response cannot publish into the returned
parent.

Fresh modified clicks and copied links preserve the complete ordinary fixed
question; opening it still stages before Run. They do not preserve this tab's
held parent, read basis or investigation trail. No new endpoint, query grammar,
router, issue-deduplication engine or error ontology is introduced.

See [architecture review](reviews/TYPED_ERROR_TRIAGE_ARCHITECTURE_370.md),
[design review](reviews/TYPED_ERROR_TRIAGE_DESIGN_370.md) and
[validation](../VALIDATION.md).
