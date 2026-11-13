# Broaden a constrained Activity field

Executable: [field-constraints.spec.js](../../journeys/field-constraints.spec.js).
Status: implemented with actual desktop/narrow frames accepted in the bounded
[UX review](../reviews/FIELD_CONSTRAINTS_REVIEW_424.md). Final executed browser,
actor and model outcomes are recorded in [validation](../../VALIDATION.md).
The implementation is described in
[Contextual field constraints](../CONTEXTUAL_FIELD_CONSTRAINTS.md).

An operator sees only one cohort because an earlier filter constrains that field.
They want to change that condition without rebuilding the environment, service,
lookback policy and unrelated predicates.

The real-source fixture has an AND condition for `attribute:cohort = "common"`
and a separate mixed OR condition: that same attribute value **or**
`resource:cohort = "resource-blue"`. Removing the AND condition broadens 258
matching records to 260 while retaining the OR condition. A missing attribute
with the other resource value remains excluded. These expected totals are checked
against raw source records rather than added from the on-screen values.

| Step | Operator action | Required observation |
| --- | --- | --- |
| Inspect | Inspect a held Activity log and choose Facet by on `attribute:cohort`. | Existing field controls open without an acquisition. |
| Locate | Open Applied constraints. | Exact same-field conditions appear separately, with operators, values and Boolean ancestry. A `resource:cohort` clause remains distinct. |
| Edit | Choose Edit condition on the second matching clause. | The exact existing editor condition receives focus; neither the draft nor source changes merely from entry. |
| Remove | Choose Remove condition on a clause with only AND ancestors. | One complete validated draft is staged; every unchosen condition and typed environment policy remains intact. Held records do not change. |
| Retire counts | Inspect the previously loaded field values. | Old counted results and selection controls withdraw; they are not relabeled as counts of the broader draft. |
| Refuse unsafe rewrite | Inspect a same-field clause nested in a mixed-field OR group. | Edit remains available; Remove is unavailable with an OR explanation. The unrelated branch is preserved. |
| Refuse stale actions | Replay a withdrawn or changed-question action. | No predicate change, source acquisition or obsolete-control revival. |
| Apply | Explicitly Apply scope. | One new Activity acquisition uses the staged predicates and unchanged rolling lookback in a new current window; the old fixed window is not silently replayed. |

Loading values again after Apply remains an existing explicit follow-on action,
not a step claimed by this particular journey. Its new counted result would
describe the new Activity window, not the retired original scope.

Local entry, native disclosure, Edit, Remove and stale-action checks must preserve
source/SQL counters. Explicit Apply owns its Activity read; explicit counted Load
owns its separate count/group reads. The real-source fixture must independently
check membership rather than treating a click or displayed count as success.

Actual app captures show both exact clauses, the OR edit-only limitation, original
258-record counts, and staged Apply with its explicit new-window cue:

- [Clause list at 1440px](screenshots/field-constraints-1440.png)
- [Clause list at 390px](screenshots/field-constraints-390.png)
- [Clause list at 320px](screenshots/field-constraints-320.png)
- [Staged removal and Apply focus at 390px](screenshots/field-constraints-staged-390.png)

The main report is `artifacts/field-constraints424/report.json`, version
`goo.fieldConstraints424/1`. Its continuous ledger distinguishes local edits,
the explicit counted read and explicit Apply. The optional 35-second Bombadil
actor runs after that main ledger. Existing native verification checks the
observed ledger and generated removal semantics; no successful actor or model
outcome is implied by the screenshots alone.

The shared observed-field list retains its bounded native scroll area. A long
clause list can place later field actions below that area's visible edge; native
keyboard focus scrolls the target into view. This slice does not redesign the
entire Boolean editor or automatically broaden OR groups.
