# Records around a retained gauge observation

Milestone 401 brings the existing gauge observation gesture to Services and
immutable Saved measurements. [Validation](../VALIDATION.md) records actual
checks; ordinary Measurements retains its [388 contract](GAUGE_OBSERVATION_CONTEXT.md).

Select a reported point in the existing gauge timeline inspector, then choose
**Logs around observation** or **Spans around observation**. The action stages a
fixed record question. Only **Run** reads fresh records. Return restores the held
measurement, selected point, focus and scroll without a native reread or another
capture GET. No new toolbar, source evaluator or HTTP route is introduced.

The selected timestamp must identify one actual sample in one retained stream.
An observed zero remains zero; an unavailable magnitude can still supply context
when its reported timestamp is retained. Stocks, resource gauges, duplicate
stream/sample identities and timestamps outside the original window are refused.

The existing planner forms ±60 seconds around that timestamp, clipped to the
original half-open nanosecond bounds. Millisecond execution bounds round outward;
exact timestamp predicates preserve the clipped interval. The question keeps the
original environment policy and the selected service, instance and exact missing,
empty or named namespace. Metric attributes and scope/schema metadata are not
record predicates. This is time association, not causality or a health policy.

## Two retained-evidence boundaries

Services first matches the successful result to the actual requested widget
question and current visit. The owner freezes that result, then recomputes the
selected observation plan. Draft picker changes retire staging immediately,
without reads. Apply, replacement, departure and stale callbacks retain their
existing lifetime rules. Reverting a draft does not revive retired capabilities.

Saved first validates the capture response against the requested visit/receipt,
then admits the full immutable receipt/result envelope at its native owner. It
recomputes stream/sample membership and the plan from that held result. Browser
metadata checks do not replace the server's immutable-content/replay checks, and
a well-shaped receipt ID alone cannot establish the expected visit identity.

The citation keeps three different clocks: reported sample time, original native
read time/window, and capture custody time when present. The record result has its
own fresh read time. A current clock or caller-supplied timestamp cannot substitute
for membership. Saving a record question alone does not preserve this native citation.

```mermaid
sequenceDiagram
  participant User
  participant Owner as Services or Saved retained owner
  participant Plan as Existing observation planner
  participant Fixed as Fixed record question
  participant Source as Record source
  User->>Owner: Select retained sample
  Owner->>Plan: Admit request/receipt and unique stream/sample
  Plan-->>Owner: Exact clipped question and original citation
  Owner->>Fixed: Stage question; retain native DOM and selection
  Note over Owner,Fixed: No source read or capture GET
  User->>Fixed: Explicit Run
  Fixed->>Source: Read contextual logs or spans
  Source-->>Fixed: Independent records and fresh read clock
  User->>Fixed: Return
  Fixed-->>Owner: Restore unchanged evidence, focus and scroll
  Note over Owner: Renew only retained context, not capture/save/read authority
```
