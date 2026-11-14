# 439 first-use acquisition: architecture review

## Decision

Keep source capability checks scoped to each explicit acquisition. A discovery
request and a later raw-value request have different lifetimes and projections.
Their repeated gauge schema/access probes are not proof that a durable readiness
cache is safe: grants, schemas and availability can change between reads, and
metadata compatibility does not establish raw-value compatibility or freshness.

The concrete consolidation is the raw inspector's navigation lifecycle. Closing
or changing its question should discard its bank; temporarily leaving the same
mounted owner should retire pending acquisition authority while retaining an
already admitted result. Previously `measurements.suspend()` called destructive
`rawGauge.retire()`, forcing another explicit acquisition to revisit settled
observations. Separate `suspend()` from `retire()` without adding a new carrier,
readiness cache, evaluator path or global state owner.

## Acceptance boundaries

- Suspension increments the request generation before cancellation. An actual
  late successful response cannot replace a settled bank, receipt or focus.
- Settled rows, fixed question, source-read clock, local Find, page and detail
  remain supplied data owned by the same mounted inspector. Returning does not
  fetch, revalidate readiness or relabel the bank as fresh.
- A pending first read has no settled bank to retain. It cannot create one after
  departure. A pending reread may leave the earlier settled bank available.
- Scope edits, new metadata, a different instrument and explicit Close retain
  destructive retirement; no data crosses question identities.
- Local callbacks remain inactive while the owning surface is hidden. Neither
  suspension nor return moves focus. Source actions still require explicit Read.

## Source work

`sourcecap.Inspect` performs per-family fixed DESCRIBE and LIMIT 0 access checks,
followed by a bounded data query owned by the source adapter. Source scope
bootstrap and instrument discovery check their supported family sets; raw gauge
inspection checks only gauge metadata and its own raw projection. These checks
must remain independent of measurement evaluator preflight.

Pure namespace predicate extraction is possible in `measurements`, but is not a
measured source-work reduction. A new helper or runtime cache would add review
surface without addressing the observed destructive return. No such abstraction
is required for this milestone.

## Validation status

Source review accepts the bounded implementation: shared `stopRead()` advances
request generation before aborting, restores the opener and clears pending
controls. `suspend()` preserves only an existing settled bank; its pending-first
case destructively retires. Scope/new-discovery/Close still call `retire()`.
Find now has the same active/native-control guards as paging and inspection.
No source, evaluator, readiness or persistence code changed.

Independent `node tests/raw-gauge-observations-438.test.mjs` passes all **17 tests**
with zero skips (0.130s). New lifecycle checks cover same-bank local state/clock,
pending reread late success and focus, pending-first retirement, and explicit
subsequent acquisition. Actual first-use measurement and responsive acceptance
remain separately reported by the owning journey; this source check does not
claim browser or source-work results.
