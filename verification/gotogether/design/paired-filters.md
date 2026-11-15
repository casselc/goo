# Paired Activity: one scope, independent conditions

An operator wants error logs beside slow server spans for the same service and
window. Editing the log condition should leave the span lane alone. Applying a
condition explicitly starts that lane's read; a shared scope replacement retires
old source authority for both lanes. A late result must not become current.

This is a **design comparison for roadmap item 450**, not an implementation or
an observed trace of Goo. The existing native source/browser gates remain
independent. Boolean conditions stand for two distinct checked predicates, and
scope revisions stand for complete environment/service/time identities. The
model does not validate CEL, query lowering, telemetry completeness or fairness.

## Alternatives and state

Only one switch changes in each alternative. `shared filter` tests coupled
condition ownership; `automatic read` tests acquisition on draft edit; `split
scope` tests independently drifting lane scope; `unfenced result` is a correctness control that adopts
an old response. None is an implementation mutant of the existing application.

```quint
module pairedFilters {
  const SHARED_FILTER: bool
  const AUTO_READ: bool
  const SHARED_SCOPE: bool
  const FENCED: bool

  var scope: int -> int
  var draft: int -> bool
  var pendingScope: int -> int
  var crossedEdit: bool
  var implicitRead: bool
  var staleAdoption: bool
  var turns: int

  action init = all {
    scope' = Map(0 -> 0, 1 -> 0),
    draft' = Map(0 -> false, 1 -> false),
    pendingScope' = Map(0 -> -1, 1 -> -1),
    crossedEdit' = false, implicitRead' = false, staleAdoption' = false,
    turns' = 0,
  }
```

## Operator and source actions

`editFilter` changes a draft without acquiring data in the chosen design.
`apply` explicitly issues a lane read and captures its scope revision.
`changeScope` replaces scope authority; held results may remain visible under
their old labels, but `deliver` cannot adopt an obsolete response as current.
The ghost flags record violations when they happen rather than lose them after
another edit. Four transitions suffice to expose the three-step late response.

```quint
  action editFilter(lane: int): bool = all {
    turns < 4,
    draft' = if (SHARED_FILTER)
      draft.put(lane, not(draft.get(lane))).put(1 - lane, not(draft.get(lane)))
      else draft.put(lane, not(draft.get(lane))),
    crossedEdit' = (crossedEdit or
      (SHARED_FILTER and draft.get(1 - lane) != not(draft.get(lane)))),
    pendingScope' = if (AUTO_READ) pendingScope.put(lane, scope.get(lane))
      else pendingScope,
    implicitRead' = (implicitRead or AUTO_READ),
    scope' = scope, staleAdoption' = staleAdoption, turns' = turns + 1,
  }

  action apply(lane: int): bool = all {
    turns < 4,
    pendingScope.get(lane) == -1,
    pendingScope' = pendingScope.put(lane, scope.get(lane)),
    scope' = scope, draft' = draft, crossedEdit' = crossedEdit,
    implicitRead' = implicitRead, staleAdoption' = staleAdoption,
    turns' = turns + 1,
  }

  action changeScope(lane: int): bool = all {
    turns < 4,
    scope' = if (SHARED_SCOPE)
      Map(0 -> scope.get(lane) + 1, 1 -> scope.get(lane) + 1)
      else scope.put(lane, scope.get(lane) + 1),
    draft' = draft, pendingScope' = pendingScope,
    crossedEdit' = crossedEdit, implicitRead' = implicitRead,
    staleAdoption' = staleAdoption, turns' = turns + 1,
  }

  action deliver(lane: int): bool = all {
    turns < 4,
    pendingScope.get(lane) >= 0,
    staleAdoption' = (staleAdoption or
      (not(FENCED) and pendingScope.get(lane) != scope.get(lane))),
    pendingScope' = pendingScope.put(lane, -1),
    scope' = scope, draft' = draft, crossedEdit' = crossedEdit,
    implicitRead' = implicitRead, turns' = turns + 1,
  }

  action step = {
    nondet lane = Set(0, 1).oneOf()
    any { editFilter(lane), apply(lane), changeScope(lane), deliver(lane) }
  }

  val alignedScope = scope.get(0) == scope.get(1)
  val independentDrafts = not(crossedEdit)
  val explicitAcquisition = not(implicitRead)
  val currentResponseOnly = not(staleAdoption)
}
```

## Checked comparison

<!-- gotogether:matrix -->
| | independent | shared filter | automatic read | split scope | unfenced result |
|---|---|---|---|---|---|
| alignedScope | holds | holds | holds | breaks in 1 step | holds |
| independentDrafts | holds | breaks in 1 step | holds | holds | holds |
| explicitAcquisition | holds | holds | breaks in 1 step | holds | holds |
| currentResponseOnly | holds | holds | holds | holds | breaks in 3 steps |
<!-- /gotogether:matrix -->

The table is written by the pinned upstream tool and checked by
`TestPairedFilterDesignMatrix`. It exhausts the model's four-transition state
space, not arbitrary-length product executions. Breaking alternatives must
actually produce counterexamples; claiming every property holds is a failing
negative control. A second control makes the chosen design unable to read,
which the required example must reject.

## Implementation consequence

Proceed with one shared scope and independently authored lane predicates. Keep
local draft editing source-free and apply only the selected lane. Reuse checked
record predicates and existing query owners; bind each response to that lane's
applied query and generation. Preserve explicit gaps, refusals and old held
labels. This matrix does not settle concurrent read ordering, source bounds,
overlapping reads or same-scope query generations, filter editor accessibility,
Pause/Resume or saved-definition format. Those
remain implementation and real OTLP journey acceptance for 450, followed by the
451 consolidation pass. No new product query format is introduced by this trial.

## Reproduce and maintain

From `verification/gotogether`, with the repository-pinned Go tool and Quint
0.33.1 on PATH:

```sh
go tool gotogether simulate -binding design/paired-filters.binding.yaml -matrix -graph -states 10000 -timeout 2m -write-matrix design/paired-filters.md
go tool gotogether tangle design/paired-filters.md
```

Retangle **after** writing the table: the table changes the Markdown digest,
so `-write-matrix` leaves its previously generated sibling stale in this pin.
Commit the Markdown and generated sibling together. CI checks `tangle -check`
first, then runs the matrix comparison and both negative controls. No private
matrix renderer or model extractor is used. `currentResponseOnly` here checks
scope revision retirement; same-scope query replacement still needs its own
implementation/generation tests.
