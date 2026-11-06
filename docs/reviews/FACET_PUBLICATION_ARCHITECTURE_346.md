# Measure and simplify facet publication — 346

Status: baseline, implementation and matched native result accepted. The initial
contract below was grounded in 345 (`537c879`); production changes followed the
genuine baseline rather than setter-count speculation.

The existing facet already admits at most 1,000 values per operand and preserves
a disjoint 2,000-value union. 343 reduced local Find to one DOM pass and exposed
long-label suffixes. It did not reduce the roughly 32,000-element paired tree.
346 targets demonstrated repeated publication/build work, not the query engine,
source cardinality, renderer catalogue or a new universal virtualizer.

## Source-grounded mechanism

`CategoricalFacet` separately renders on connected `.data` and `.interaction`
assignment. Its connection callback renders again, and its data getter detaches
supplied state. The population `renderFacet` owner installs interaction then
publishes data; retained element reinsertion can also reconnect. Return renewal
calls updateSelection/syncCompositionControls, and closeChild subsequently syncs
again; those paths call refreshFacet and reconstruct the same full ready tree.

These are operation-level duplication candidates, not proof that every source
publication builds twice. `setFacetValues` normally calls only refreshFacet/data;
the independently completing selected and baseline reads legitimately publish
separate partial/ready observations. A setter duration contains its nested render
duration; summing them double-counts work. A slow individual live render cannot
be causally explained solely by the setter pattern.

## Baseline before implementation

Use the genuine admitted 1,000-value ordinary facet and disjoint 1,000/1,000 pair,
with the same explicitly sized 384-byte labels, source fixture, styled desktop and
narrow viewports as the maintained large-facet journey. Record executable/source
provenance and exact comparison-body size; the independent 1 MiB compare-body cap
is not changed. Supplied 2,048-byte profiles remain a separate passive diagnostic,
not evidence that they traversed that API.

Measure each logical operation separately: owner initialization/connection,
selected-ready and baseline-ready publications, a local staged choice/clear and
retained Back. Capture connected component identity, reason/operation, number and
state of actual render calls, setter calls, synchronous operation duration and
next-paint opportunity. Keep network/source/compare counters separate. Do not
collapse all renders from a long scenario into one unexplained mount number.
Report samples/medians honestly; source acquisition is not component mount cost.

Preserve the semantic oracle throughout measurement: full literal union, exact
counts/order, missing versus empty, baseline-only values, partial unknown versus
ready absence, independent receipts/clocks, loaded Find query and disclosure,
focused current control and caret. Snapshot/trace instrumentation must not cause
extra product reads or secretly batch the events being measured.

## Bounded candidate if demonstrated

A passive synchronous atomic publication may accept the supplied data and current
interaction together, validate/detach the model, install both and advance the old
render epoch before one connected build. The application can then publish once
per logical state transition rather than let independent refresh callers each
rebuild. Keep source acquisition and query/session policy in the existing owner.
This is a local component update, not a global scheduler, microtask queue or
subscription. No compatibility shim is needed in this prerelease application,
but SSR data attributes, detached property assignment and the documented passive
component contract still need deliberate supported semantics.

The atomic method must not expose partially installed data/interaction to current
handlers. Failed model admission must fail closed and retire the old authority;
invalid interaction must not leave a new data tree governed by an old owner's
callbacks. A retained old handler must remain inert after publication or same-scope
ABA return. Interaction functions themselves remain application-owned and are not
serialized into source data or saved grammar.

Skipping repeated application publication requires real owner authority. A caller
can supply different values with the same claimed basis/count; those strings are
not proof of semantic equality. Use the actual owned immutable adopted state or
an explicit local publication revision, including declaration, relevant lifecycle
and disabled/staged state. In the current source, retirement mutates loading sides
to unloaded inside facetRead: naive root-reference caching would miss that change.
Every real mutation must invalidate any local no-op decision. Prefer eliminating
duplicate calls over adding a large data cache or deep-equality engine.

## Acceptance

- Demonstrate fewer full builds for the identified logical operations against the
  pinned baseline; report synchronous and paint results separately. Do not promise
  the 32,000-node initial cost is solved, or that fewer setter calls prove all
  latency improves. Independent source-side publications remain independently
  visible; do not wait for both and hide usable partial data.
- Preserve all exact data/receipt/query identities and readiness distinctions.
  Local Find still searches the full admitted union and emits the original value,
  including distinguishing suffixes beyond the preview. No truncation or phantom
  zero/complete state is introduced.
- Same-read Find/disclosure/focus/caret survive relevant update/return. A different
  admitted read retains its existing explicit reset semantics. Interaction-owner
  retirement and late source refusal/cancellation continue to work.
- Owner setup, stage/clear and Back perform no new source, compare, capture, save
  or metadata effects. Only an explicit existing Read/Retry/Run acquires data.
  No local component lifecycle starts work.
- Mounted tests cover atomic update, invalid input, detached/SSR connection,
  independent properties and stale handlers. Generated schedules exercise repeated
  publication, owner replacement and ABA; reuse existing lifecycle models for
  unchanged source/depth effects rather than invent a model for setter counts.
- Actual OTel/native observations and current frames verify source/adoption/focus
  behavior. Pure/supplied microbenchmarks complement those facts, not replace them.

Architecture will review the measured baseline before candidate selection, then
review the actual frozen source and final before/after results. Cross-layer
request-byte admission and broader specialized recovery remain separate tasks.

## Genuine baseline and candidate selection

The styled genuine baseline now establishes duplicate full builds on Back in all
four viewport/profile cases. The pinned executable hash is
`e8452c90e29c26a4fdcacd7d343748b14d5e40ad5c8bd01aa78a369a2419c61f`;
operation data is recorded under the 346 baseline summary. Each Back rebuilds the
same ready signature five times: connection, interaction, data, data, data.

| Back case | Ready builds | Rows constructed cumulatively | Summed render sync ms | Operation through paint opportunity ms |
| --- | ---: | ---: | ---: | ---: |
| Desktop ordinary | 5 | 5,000 | 477.7 | 2,154.7 |
| Desktop paired | 5 | 10,000 | 1,590.8 | 5,723.0 |
| Narrow ordinary | 5 | 5,000 | 490.8 | 1,905.4 |
| Narrow paired | 5 | 10,000 | 2,059.5 | 6,921.8 |

These are individual controlled samples, not population percentiles or a claim
that all wall time is render time. All four Back operations cause zero source
and zero comparison requests. Ordinary source publication builds the 1,000-row
ready state once; paired publication independently builds the 1,000-row partial
and 2,000-row ready states once each. Staging builds once. Reducing those legitimate
independent observations is not the selected optimization.

This supports the local atomic component publication plus per-owner redundant
refresh guard. The guard compares actual owned declaration/adopted-state references
and local disabled/staged flags; both nested loading-to-unloaded mutation sites
must first replace the facetRead root. New ownership always republishes. This
avoids removing broad refresh fanout that also maintains unrelated controls,
without making a caller-supplied basis/count string a cache authority. Final
source review and matched native measurements are still required.

## Candidate source acceptance

The reviewed implementation adds synchronous `publish({data, interaction})` and
adopts pre-upgrade properties without intermediate renders. Invalid atomic input
withdraws both values; one connected render advances the old epoch and leaves no
old row action authority. Independent setters remain immediate.

The population owner now publishes before first attachment, avoids reinserting a
retained element already under its expected parent, and uses one local last-state
tuple. It compares declaration/owned read-root references plus disabled/staged
flags, never supplied basis or count strings. Both previous loading-state mutation
sites use fresh roots; ready roots remain unchanged. New owners start without a
last publication and therefore install fresh interaction authority. Source reads
still require the connected current owner, independently of preconnection passive
publication. Source review finds no blocking issue. Matched final measurements
and semantic/native acceptance remain separate gates.


## Matched final acceptance

The final real-source run retains one ordinary ready build and one partial plus
one complete paired build per acquisition. Retained Back builds once in all four
ordinary/paired desktop/narrow cases, compared with five identical ready builds
in each baseline. Source requests, comparisons, exact values/read clocks, staged
intent, Find and current focus remain checked. Generated publication schedules
and full UI checks pass; final frames are accepted.

This supports less duplicate local work, not a uniform end-to-end latency claim.
Desktop paired wall-through-paint time is higher in the final individual sample
despite reduced render CPU; other phases vary too. The paired tree remains about
32,000 elements. See [validation](../../VALIDATION.md) for actual sample values,
receipt admission and the final paired-facet actor/model regression.
