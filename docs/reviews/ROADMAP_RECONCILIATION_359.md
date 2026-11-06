# Roadmap reconciliation — 359

Proposal after accepted 356 breadth → 357 depth → 358 consolidation. Read with the
[architecture review](WHOLE_PRODUCT_ARCHITECTURE_359.md),
[design review](WHOLE_PRODUCT_DESIGN_359.md) and
[active roadmap](../ROADMAP_2026_11_05.md). This document changes no production
code, source budget or active execution order. Root adoption is separate.

The product loop remains **see current telemetry → isolate a population → inspect
supporting signals → preserve or reuse**. The next sequence makes preservation
useful after telemetry changes, deepens comparison of that retained evidence, then
removes one demonstrated convention inconsistency. It is not another callback,
styling or ownership maintenance queue.

## Close the previous sequence and correct the inventory

| Delivered | Keep | Remove from the active queue |
| --- | --- | --- |
| 356 service-context templates | Existing catalogue, checked destination binding and explicit Run | Another template catalogue or specimen-scope inheritance |
| 357 current RPC | Pinned v1.45.0 profile, optional method groups, independent native seconds and explicit older interpretation | Alias unions, mandatory-method assumptions and one merged RPC denominator |
| 358 fixed-question read/Edit | Actual Session authority, complete checked questions, raw drafts, fixed clocks, explicit reads and widget-side mutation admission | Another editor owner or a continuing fixed-question CSS/lease cleanup sequence |
| Existing infrastructure | Host CPU/memory/network/disk/filesystem and pod/container resource/state/history work | A proposal to add basic Kubernetes or USE dashboards as if those foundations were absent |
| Existing trace/view grammar | Supplied-page waterfall, conservative timing projection, pure trace-page comparison and record kinds | A new trace engine, capture store or blanket presenter migration |

358's measured plots move up approximately 431 px desktop and 890 px narrow;
the selected charts still remain below the fold. Preserve that limit honestly
without automatically selecting another chrome reduction. The roughly
32,000-element paired facet tree also remains a measured structural limit, not
proof that a specific virtualizer or latency improvement is required.

## Proposed finite sequence

### 360 — breadth: inspect captured spans without rereading telemetry

**User task:** “Show the timing in the span evidence I actually saved, even if the
source has changed or is unavailable.”

`ui/saved-records.mjs::showSaved` currently renders a table. Its trace examples
and expansion actions call `records.openTrace`, whose trace/linked-log acquisition
is fresh. In contrast, `record-span-waterfall` already presents supplied span
pages and preserves original row positions. Connect those existing capabilities.

Start with the current checked saved span page and a contextual Table/Span
waterfall choice. The page is the authority, including mixed trace IDs, duplicates
and malformed observations. If a trace-specific subset is offered, it must be
an explicit local selection of those supplied rows with preserved original
positions; it is not permission to fetch all pages or a complete trace. A generic
capture's full population count is not the selected fragment's count.

Preserve the immutable receipt, original question, source-read/capture clocks,
page cursor/offset, retained limits and raw fields. Existing supplied-page timing
rules retain missing parents, ambiguous identities, unavailable timing and known
zero durations. Inspect uses the exact stored observation. Fresh Expand trace and
linked logs remain separately named, explicit source operations. Never route a
local frozen-row inspection through the fresh inspector merely to reuse its UI.

**Acceptance:** capture ordinary multi-service spans, reopen while telemetry reads
are refused, switch representation and inspect exact original occurrences. Include
scalar and typed policies, mixed traces, duplicate IDs, missing parents, partial
pages/limited captures, zero versus invalid duration and large exact timestamps.
Switching kind, local selection and Back read no telemetry or mutate captured
bytes. Existing receipt-store GETs/paging are permitted and accounted separately;
“source-free” does not mean no HTTP. Actual desktop/narrow frames distinguish
captured page evidence from the separately tested explicit fresh expansion.
Use meaningful generated ownership/data properties and positive native/actor/model
checks, retaining the existing passive component and current capture-visit owner.

**Stop:** a retained-page representation and its local inspection. No complete-trace
promise, automatic parent/log read, new store, capture format migration, fleet
summary or newly inferred relationship.

### 361 — depth: compare two deliberately retained trace pages

**User task:** “How does this preserved slow/error fragment differ from my chosen
reference?”

Reuse `exploration.CompareTracePages`, its JS parity model and
`ui/trace-comparison-view.mjs`. They already express signed exact durations,
nullable service/operation groups, conservative outside-observed-children values,
independent origins and withheld row alignment. The missing part is a clear
stored-page operand entry, not another comparison calculation.

Explicitly pin/choose the baseline and selection from checked immutable capture
pages. Each side retains capture ID, receipt, original question/policy/window,
basis, capture/read time, cursor/offset and displayed-row/retention limits. Use
the existing verified capture/page owner for acquisition; do not accept arbitrary
browser-authored rows as verified evidence or introduce a global reference manager.
Local projection and operand actions perform no telemetry read. Stored-page GETs
remain visible in a separate ledger. Do not silently combine all retained pages.

Groups present on one page are not zero work on the other trace. Unique-label
matches are descriptive candidates, not the same invocation, incident or causal
path. Repeated labels are not zipped by ordering. Mixed/invalid IDs, missing
timing, overlapping intervals and partial pages retain the existing model's
unknown/refusal semantics; summed inclusive spans are not wall time or CPU time.

**Acceptance:** select two different captured pages/windows and compare with
readable baseline/selection roles and one common elapsed scale. Include different
environment policies, nullable/empty/literal labels, repeated groups/IDs, one-sided
groups, unknown timing and unequal/partial extents. Inspect each stored side and
Back without telemetry, preserving exact page identities, immutable bytes and
independent receipts. Reuse Go/JS parity fixtures plus independent raw assertions;
exercise late page replies, replaced operands and retired actions in real entry
journeys. No automatic typical-trace selection or statistical/causal claim.

**Stop:** the existing supplied-page comparison over two explicit retained
operands. Do not add persistent comparison artifacts, a capture graph, a new
numeric engine or a cross-query occurrence identity API unless a separately
reviewed requirement demonstrates it is necessary.

### 362 — consolidate: current RPC capability parity

**User task:** “The service dashboard recognizes this duration instrument; its
measurement entry should explain the same supported shape and prerequisites.”

The source discrepancy is concrete: `exploration/dashboard.go` recognizes
`rpc.server.call.duration` Histogram seconds, while curated paths in
`measurements/instrument_capabilities.go` still cover HTTP/DB/messaging and leave
this instrument generic. Add the missing current RPC capability through the
existing `InstrumentCapabilityPlan` and presentation paths.

Keep the pinned 357 convention facts, exact name/kind/unit, temporality, full
stream identity and bounded attribute coverage. Metadata shape, capability,
observed data and verified convention compliance remain distinct. Conditional
`rpc.method` must not become a mandatory admission gate. Older millisecond
instruments and custom metrics remain separate useful profiles.

Share small internal checked facts only where actual duplication is demonstrated;
Go planning remains authoritative and the UI's independent same-row positive
witness remains independent. No public generic profile registry, descriptor
generator, alias system or advanced type machinery is required. Do not generate
both implementation and its purported semantic oracle from one table.

**Acceptance:** genuine current RPC instrument discovery → service/native option
→ capability inspection → explicit measurement execution and frozen reopen.
Compare whole emitted questions and descriptions under current/older/mixed,
wrong-name/unit/kind, missing/conditional dimensions and omitted-stream cases.
Preserve source budgets, exact stream identity, independent clocks and custom
metric support. Generated/API/native tests must reject drift without converting
metadata eligibility into producer conformance. Use Go 1.27.3/stdlib facilities
for simpler internals, not an unrelated runtime migration.

## Deferred work remains relevant, not a parallel queue

- **Large facets:** first obtain a fresh styled mount/search/action/Back baseline
  separating source, synchronous build, layout/paint and wall time. Only then
  select bounded presentation if material cost is demonstrated. Keep full-model
  Find, exact labels/counts, uncertainty, global scales, selection and zero-read
  local paging. Do not truncate measurement authority or raise source limits.
- **Broker backlog:** `lab.queue.depth` is a demo, not a generic broker default.
  Pin a real exporter/receiver, metric shape and queue/destination/partition
  identity before implementation. Processing spans do not establish backlog,
  lag, capacity or broker health. This uncertainty makes it a worse immediate
  depth cut than existing retained-page comparison.
- **Reviewed training labels:** tags/notes and retained evidence exist; task
  vocabularies, abstention/review and pinned dataset export remain proposed.
  Start with a concrete evidence-review task, not automatic ground truth, a
  classifier runtime or an XTDB replacement.
- **Exemplars, verified workload topology, multiplayer, production authorization,
  retention guarantees, Databricks lowering and whole-view WYSIWYG:** retain their
  own source/authority and user contracts. None is a hidden prerequisite for the
  selected retained-evidence loop.

Keep the blocked vulnerability assessment and historical broad-harness failures
explicit; dedicated accepted journeys do not make them disappear. Reconcile stale
“Kubernetes next” prose with actual delivered modules. After these three outcomes,
review the whole product again instead of extending another maintenance queue.
