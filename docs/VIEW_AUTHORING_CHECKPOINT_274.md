# Finite adoption checkpoint and next authoring step — 274

The goal is a reusable observability interface, not a migration of every presenter
before users can compose anything. This checkpoint closes the current native
histogram adoption sequence and sets a concrete next product outcome.

## What the shared foundation supports

Canonical supplied data now covers ordinary count tables/breakdowns/facets,
intervals, summaries, numeric and duration measures, recorded status fractions,
numeric histograms, native export/time-bucket histograms and aligned comparisons.
Exact scope, units, counts, unknown states and independent clocks travel with the
data. H0 separates immutable data from runtime completeness; H1 separates captured
record-action ownership from data identity. Passive components do not fetch.

274 consolidates the existing native capture-comparison screen: one aligned
histogram with count/observation-share choices and exact values replaces duplicate
chart/table/heatmap presenters. Coverage diagnostics and percentile bucket ranges
retain their existing semantics. It adds no comparison lab or source reads.

## A bounded gate, not an endless conversion list

The following owners remain explicit and do not block initial reusable authoring:

| Existing specialization | Boundary for authored views |
| --- | --- |
| Native gauge/stock/counter timelines and categorical state | Keep current presenters; advertise only kinds supported by the checked profile. |
| Native heatmaps, compatible fleet merges and window coverage evaluation | Keep the evaluator and source owner; do not infer generic reducers or automatic rebucketing. |
| Error triage and linked multi-signal population workspaces | Preserve their application actions and scope owners; a single display declaration does not imply whole-screen composition. |
| Live tails and baseline/live comparisons | Retain one subscription owner and existing read/retention semantics. Initial saved displays need not recreate this controller. |

A missing kind should produce a clear compatibility result. It should not require
converting unrelated surfaces, weakening scope validation or adding another
standalone demo before an authoring feature can ship.

## Next product outcome: extend the existing saved-view workflow

The next implementation step is a reusable named display configuration that users
can save from an existing exploration, reopen and explicitly run. Build on
`workbench.ExplorationView`, its current configuration catalog and the existing
**Save this exploration view** flow. Do not add a parallel view library or a new
workbench solely for the pilot.

The declaration needs a checked typed scope, pinned public kind/version and bounded
display/field options. It stores intent, not captured data, source witnesses or DOM
state. The property editor and source representation must describe the same AST.
Changing a compatible kind is local; explicit Run owns source acquisition. Save
and reopen must retain the chosen scope and presentation without reading data.

The first useful authoring slice is one canonical component in that existing
workflow. It should preserve current saved-workspace capabilities while making
supported presentation choices reusable. Native scopes must use their checked
measurement contract; exact bounds or deterministic bucket policy replace
transient timeline row indices. Unsupported variants remain explicit. Current
`environmentSelector` save rejection must be reconciled with the typed scope
contract, never bypassed with a wildcard or replacement environment string.

Acceptance should demonstrate a real scope → presentation edit → save → reopen →
explicit Run journey, plus source/property round-trip and malformed-scope controls.
It should add a reusable capability a user can find in the existing product.

## Renderer and composition priorities

A Plotly adapter is optional and should serve a concrete missing visualization.
It is not a prerequisite for saving declarations over existing native components.
Normal authors should use named observability kinds rather than vendor traces.
Team-contributed kinds/templates follow a checked declaration contract. Responsive
whole-screen WYSIWYG follows reusable single-component declarations; it is not part
of this first authoring slice.

See the [execution roadmap](ROADMAP_2026_10_30.md) for implemented milestones and
[the grammar proposal](VIEW_GRAMMAR_PROPOSAL_260.md) for the wider product direction.
