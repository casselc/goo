# Continue a trace search

The [main journey](403-trace-continuation.spec.js) ingests ordinary OTel spans and
independently reads the persisted rows. Seven matching cross-service trace IDs,
a foreign environment, a one-condition trace and an end-boundary trace establish
scope and membership. A two-item limit makes continuation visible without a large
fixture. The [actor](403-trace-continuation-actor.spec.js) explores the same source
bank independently. Executed results belong in VALIDATION.md.

1. Reopen a fixed trace-search question. Conditions, environment policy, bounds and
   list limit are staged without reading spans. Run explicitly reads the first
   lexical page; its full matching-trace count is independent of the list limit.
2. After a successful read, the existing search conditions fold into a native
   disclosure so results and page controls lead. Reopen it to edit.
   **Next page (fresh read)** replaces the accepted page. It sends the same search
   plus its bound position. Witnesses remain qualified by condition; the count
   before position and remaining count after position are different facts.
3. Open a listed trace through the existing fresh-read drawer. Its span/log reads
   preserve the original environment and window. Back restores the same accepted
   search page, row DOM and witness disclosures without rereading the search.
4. Capture the current positioned page. Capture explicitly rereads and stores the
   page question, position, witnesses and fresh read clock in immutable custody.
   Reopening with the source unavailable reads only the stored capture.
5. Save the search question. The reusable authored intent contains conditions,
   fixed scope and list limit, but no continuation position or captured evidence.
   Reopening starts an empty first-page draft; execution is explicit.
6. Continue through the last page. **First page (fresh read)** restarts the same
   search, rather than treating previously visited pages as a stable accumulated
   result. A source query after the final ID can have an empty page while the full
   matching count remains nonzero.
7. Hidden, inert and retired Next callbacks cannot read. Reopen **Search conditions**
   and edit the limit while a
   continuation response is held aborts/invalidates that visit; releasing the late
   response cannot restore its old rows or position.

The actor takes first → next → first using actual controls at 390px and observes
exact request/result IDs, position binding and three explicit single-SQL reads.
It publishes no capture or document. Neither journey claims a stable snapshot,
complete union, causal span ordering or complete traces. Each page is a fresh read;
late ingestion or corrections can change membership between visits.

Actual frames are generated from the running app after acceptance:

- [Frame](../docs/journeys/screenshots/403-trace-continuation.png)
- [Frame](../docs/journeys/screenshots/403-trace-continuation-390.png)
- [Frame](../docs/journeys/screenshots/403-positioned-capture-390.png)
