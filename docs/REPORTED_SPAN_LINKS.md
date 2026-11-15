# Follow a reported span link

In a trace inspector, select a span and choose **Look up linked span** to inspect
an OTel link whose target is outside the returned spans. Each action shows its
link position and a short span ID; the accessible name and tooltip contain the
full trace/span pair.

The lookup stages first. Check the displayed environment and UTC window, then
choose **Run target lookup**. The query matches both reported target IDs, retains
the original environment policy and fixed half-open time window, and omits
service, namespace and population conditions. A linked target may belong to
another service. The lookup never widens the environment or time window.

**Back to trace**, or Escape, returns to the held trace and restores focus without
reading it again. Returning during a request aborts it; a late response cannot
replace the held trace. The trace inspector's close control still closes the
whole inspector. After returning, Escape follows the existing inner span-detail
then outer trace-inspector hierarchy.

If exactly one target tuple is already returned, **Follow linked span** selects
that row locally. Duplicate tuples remain separate observations and never select
an arbitrary first match. In the expanded trace inspector, the existing row
selection may separately read associated logs; target lookup itself does not
expand another trace or read its logs.

## Meaning and limits

- A link reports a trace/span reference, not causality, delivery, processing or a
  unique invocation. No matches means none were returned in this environment and
  window; it is not proof of failed delivery.
- Both IDs must be canonical nonzero lowercase hexadecimal values. Missing,
  invalid or misaligned link arrays produce an unavailable reference; Goo does
  not guess a trace-only target.
- The inspector offers the first 32 reported links and shows how many further
  links are not displayed. A lookup returns at most 100 span observations;
  additional matches are disclosed rather than treated as a complete result.
- Target acquisition is available in expanded record trace inspection. Other
  waterfall owners can display links and follow a unique held target, but do not
  acquire targets unless their application adapter supplies that action.
- Inverse discovery (“which spans link here?”), recursive traversal and widening
  the lookup window are separate roadmap work.

## Implementation

`ui/span-link-intent.mjs` owns pure tuple admission and ordinary-question planning.
`ui/visualizations.mjs` presents bounded reference actions. The passive
`goo-trace-waterfall` remains a supplied-data component with no query client.
`ui/span-link-context.mjs` owns the staged panel, request lifetime, response
admission and held return; `ui/record-results.mjs` supplies its source client and
revokes it when the parent changes.

Returned questions must match the requested question. Every displayed target row
must match both IDs, the exact environment policy and the fixed time bounds.
The target read retains its own basis/read timestamp; the original trace basis
and read timestamp remain separate in **Question and reported link**.

```mermaid
sequenceDiagram
  actor Operator
  participant Trace as Held trace inspector
  participant Lookup as Target lookup owner
  participant Source as Go API / ClickHouse
  Operator->>Trace: Look up reported link
  Trace->>Lookup: Admit tuple and stage exact question
  Note over Trace,Lookup: Held origin retained; no source read
  Operator->>Lookup: Run target lookup
  Lookup->>Source: Ordinary spans question, exact IDs + environment/window
  Source-->>Lookup: Target observations and independent read basis
  Operator->>Lookup: Back to trace
  Lookup->>Trace: Restore held DOM and focus
  Note over Trace,Source: No return read
```

See the [executable storyboard](journeys/span-link-target.md) and
[validation record](../VALIDATION.md) for executed checks.
