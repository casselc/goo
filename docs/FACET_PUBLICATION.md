# Publishing counted facets

369 keeps this atomic publication contract and bounds mounted categorical rows
to 100 matches. Full-model search/counts/scales and exact read receipts remain
authoritative; see [current measurements](LARGE_FACET_INTERACTIONS.md).

A counted facet displays supplied observations; its application owner decides
when values may be read and which current interaction may stage a selection.
Publishing data and interaction together avoids rebuilding the same large list
for each separate assignment.

`goo-categorical-facet.publish({data, interaction})` validates and detaches the
supplied data, installs both properties synchronously and builds once when
connected. Detached publication is adopted on connection. Invalid publication
withdraws both data and interaction; old row handlers do not remain actionable.
There is no timer, network acquisition or deferred global update queue.

```js
facet.publish({
  data: suppliedFacetModel,
  interaction: {current: () => ownerIsCurrent()}
});
```

The existing `data` JSON attribute, `.data` and `.interaction` properties remain
supported. Independent connected property assignments each publish immediately;
use the atomic method when changing both. SSR markup and detached/pre-upgrade
properties adopt on connection without intermediate full builds.

The population adapter also avoids republishing the identical owned state during
one owner visit. It checks its actual adopted-state/declaration references and
local disabled/staged flags. Every new owner publishes again. Loading-to-unloaded
transitions replace the owned root, so an actual change is not skipped. A caller's
basis or count string is never proof that different data can be ignored.

Independent source observations remain independent: a paired facet can display
the available side before the other finishes. Missing/empty literal values,
baseline-only categories, exact counts, read clocks and unknown-versus-zero
semantics do not change. Local Find searches all admitted values. Publication
preserves the established same-read Find/disclosure/focus contract and retires
obsolete handlers; it does not restart a retired Read or Retry.

## What was measured

The genuine baseline measured an ordinary 1,000-value facet and a disjoint
2,000-value pair at desktop and narrow sizes. Back rebuilt the identical ready
state five times in each case: connection, interaction and three data updates.
That constructed 5,000 ordinary or 10,000 paired rows cumulatively, despite zero
source or comparison requests. Ordinary ready-source publication built once;
paired partial and ready publications each built once and remain necessary.

The selected change targets duplicate publication, not all rendering cost. The
paired tree still contains roughly 32,000 elements. The source limits, comparison
body-byte budget and value admission are unchanged; no virtualizer, truncated
result set or cache authority is introduced.

Matched final native checks reduce Back from five ready builds to one in all four
cases. Semantic gates and actual desktop/narrow frame acceptance pass. Render CPU
samples fall, but wall time varies; this is not a uniform end-to-end speed claim.
The paired-facet actor/model regression also passes; stale actor images are
excluded from visual acceptance. See
[validation](../VALIDATION.md) and the
[architecture and baseline](reviews/FACET_PUBLICATION_ARCHITECTURE_346.md),
[component contract](UI_COMPONENTS.md) and [large facets](LARGE_FACETS.md).
