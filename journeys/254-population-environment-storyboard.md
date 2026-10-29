# 254 — internal population and comparison scope

[Executable journey](254-population-environment.spec.js) mounts the production
population owner in an isolated Go runtime. It does not activate public selector
entry or Saved mounting.

1. Stage an unreported span population and prior-window baseline. Rendering sends
   no source request. Saving an exploration view is disabled.
2. Run explicitly. Read the current population and baseline independently; each
   retains the exact selector. The baseline becomes available to the comparison
   adapter. Supplied-cohort `/api/compare` calculation is separate from telemetry
   reads. The uniquely filtered populations are empty, without a health claim.
3. Attempt the disabled view handler directly. Local validation refuses before a
   publication request, including after controls have been updated.
4. Corrupt a real planner reply from unreported to empty environment. The owner
   refuses it before any automatic record reads.
5. Accept the unmodified planner reply. Both independent signals read with the
   checked scope. Dispose the owner and delete its isolated runtime/store.

Unit properties also cover every shared selector fixture through narrowing,
widening, categorical selection/rest, raw-duration regions, log patterns and
explanations. Numeric side continuations must retain the parent's full scope,
measures and predicates. Wrong-policy controls are rejected.

Harness corrections: inspect the disabled control inside its closed disclosure,
and lay out the fixture host outside the existing sidebar. These are internal
owner checks, not a new product entry design or an accessibility acceptance claim.
