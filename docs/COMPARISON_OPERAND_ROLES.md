# Comparison operand names

Comparison roles come from the application owner that acquired the operands.
They remain consistent in facts, Summary, Groups, narrow labels, accessible group
actions, absence explanations and exact-change tooltips. Group-to-record children
retain the same operand name in their heading, alongside the original UTC window;
Back restores the supplied parent pair ([journey](journeys/334-comparison-record-drill.md)).

- Fresh adjacent queries: **Previous window** and **Selected window**.
- Other fixed query baselines: their existing fixed-baseline and selected names.
- Preserved baselines compared with a freshly read fixed population: **Frozen
  baseline** and **Selected window**.
- Preserved baselines compared with held Activity: **Frozen baseline** and
  **Current held**, only where those are the actual owners.
- Cohort comparisons: the actual **Rest of parent**, **Held parent (overlapping)**
  and **Selected** roles.

A role name does not establish evidence authority. Existing read/custody clocks,
fixed scopes, coverage warnings and snapshot validation still establish what was
supplied. Relabeling a component cannot turn a fresh query into kept evidence.

The passive Summary/Groups `display.labels` property accepts two literal strings
in baseline/selection order, each nonblank and at most 120 UTF-8 bytes, with valid
Unicode and no control characters. Both are copied and frozen. Omitting labels
uses neutral **Operand A** / **Operand B**; the exact change remains B minus A.
Product adapters supply the natural task names above. The shared count-pair
presenter derives labels from its already checked operand facts, rather than
asking a display component to infer acquisition mode.

Changing labels is local presentation. It does not reread source data, alter
values, update receipt bytes or change scope/compatibility. Components continue
to support server-supplied JSON and detached SPA property updates.

See [count-pair presentation](COUNT_PAIR_PRESENTATION.md),
[Activity window comparison](ACTIVITY_WINDOW_COMPARISON.md) and
[component library](UI_COMPONENTS.md).
