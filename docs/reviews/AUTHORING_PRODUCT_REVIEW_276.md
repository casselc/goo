# Authoring product checkpoint — 276

Read-only whole-product review of the 276 working tree, 2026-10-31. This reviews
implemented structure; [VALIDATION.md](../../VALIDATION.md) records executed checks
and shipment status. No product code changed for this review.

Goo now has a credible shared display foundation. The next useful step is to make
an authored display reusable from the user's current context. Another long list
of presenter migrations, or another chart vendor, would delay that outcome.

## Implemented and still absent

| Surface | Implemented | Composition boundary |
| --- | --- | --- |
| Activity | Live logs/spans, local columns/filters, saved rolling selectors, and grouping into a fixed population. `ui/stream-entry.mjs` calls the existing fixed owner. | `ui/activity-views.mjs` and `workbench/activity_views.go` store rolling selector/columns, without canonical display declarations. Saved Activity views still reject typed environment selectors. Do not advertise universal view portability. |
| Fixed exploration | `ui/population-workspace.mjs` saves per-signal display intent in the existing immutable catalog. 276 exposes count, query-bound numeric aggregates and span-duration aggregates through one Measure control. Source and properties edit the same declaration. | `workbench/exploration_views.go` stores exact fixed questions; display scope references their signal. No parameterized template or current-scope application is implemented. `openRecent`/`plan` resets declarations; reopening a view restores its original fixed question. |
| Saved and investigation | `ui/saved-workspace.mjs` distinguishes documents from evidence. Captures reopen held data; documents retain questions, captures/comparisons, notes and review references. | `ui/investigation-draft.mjs` has no authored-view/template card. An immutable evidence reference is not a reusable display definition or a live subscription. |
| Services | `ui/service-dashboard.mjs` selects OTel-shaped HTTP/RPC/database/messaging profiles and displays declared application contracts, with separate native/record evidence. | Profiles and widget behavior remain application code (`recordProfiles`, `profilePicker`, `widgetTitle`). Application manifest declarations do not yet contribute reusable display packages. |
| Shared display model | `exploration/view_measures.go`, `ui/view-schema.mjs`, `ui/view-values.mjs` and `ui/scope-view.mjs` retain exact measures, units, support, unknown/no-valid states, and signed-bar table fallback. | Current authored profile is one ordinary grouping and bar/table. Lists, raw-record tables, trace waterfalls, timelines, native histograms, layout and scope-changing controls are not yet whole-screen declarations. |

The count/measure slice is deliberately small. Numeric means divide by valid
support; durations retain their empirical quantile method. A missing numeric value
can be an honest rendered result. A query without the requested binding cannot
be saved as though it supplied that measure. These are good defaults to retain.

## Priorities

1. **Next: apply a saved display to the current checked population.** Add one
   contextual action in the existing exploration view catalog/editor, alongside
   opening its original fixed question. Reuse presentation intent; keep the
   destination query, typed environment policy, predicates, numeric binding and
   fixed window owned by the destination. Show compatibility before applying:
   signal, grouping shape, selected measure and quantity/unit requirements.
   Incompatible items need a useful reason, not a wildcard scope or an invented
   measure. Numeric field/unit differences must be explicit rather than silently
   carrying an old quantity-specific title into a different measurement. Applying
   stages locally with zero reads; explicit Run remains the only acquisition step.
   A real journey should start in Activity, group a population, apply a saved
   display, and save/reopen that new checked configuration without touching the
   original revision. Do not create another view library or a general parameter
   expression language for this first reuse cut.

2. **Then: one useful declarative default package.** Make a bounded service
   record-analysis package—such as HTTP server requests by route—produce the same
   checked query/display declarations used by interactive authoring. Declare its
   OTel prerequisites, grouping, recorded-count/duration meaning and available
   kinds. Preview compatibility and destination scope before Run. Exercise the
   same package from a service context and a fixed exploration; keep native
   instrument observations separate. Start with one package and a simple import
   or manifest reference. Team namespaces, trust policy and a package marketplace
   do not need to precede proving this shared path. Existing service defaults
   remain useful while adoption proceeds.

3. **Next composition primitive: shared records plus aggregate display.** Promote
   the existing record columns/table and aggregate display into a small checked
   view tree with one query owner, rather than adding independent widget queries.
   This moves toward the requested entire-UI grammar and preserves the stream-first
   experience. Bound layout choices and existing table controls are sufficient
   initially; multiple displays must retain the same source snapshot and action
   ownership. An investigation can later reference this reusable view definition
   separately from the evidence it displays. Do not turn an evidence card into a
   controller that silently re-queries telemetry.

4. **Keep scope portability visible while expanding.** Before claiming one view
   works across Activity, fixed exploration and services, close each selected
   handoff's typed-selector and execution semantics. Current fixed-view support
   does not remove Activity's explicit persistence gate. Fix those concrete
   handoffs with the reuse feature; do not make universal scope migration a new
   prerequisite that prevents the bounded feature from shipping.

## Roadmap cleanup

- Mark V2-lite and numeric/duration authoring as delivered once 276 validation is
  complete. Replace the vague “broader compatible kinds” next slot with the
  current-scope reuse outcome above. That is a user task, not another inventory.
- Keep Plotly conditional on a concrete missing visualization. No vendor adapter,
  animation framework or broad chart portfolio is required for display reuse.
- Split V3: prove one shared declarative default package before contributed kinds,
  namespaces/trust policy and broad team extensibility. Those remain proposed.
- Defer responsive WYSIWYG, arbitrary layout, user scripts/CEL transforms and
  multiplayer editing until the small shared view tree works across actual
  contexts. Keep the full-UI grammar as the direction, not a claim of current
  capability or a gate requiring every screen to migrate first.
- Keep production authorization, source isolation, retention/GC, annotation
  workflows and durable alerting as separate product workstreams. A reusable
  chart definition does not implement those operational capabilities.

The next review should follow the first display-reuse/package journey across
Activity, fixed exploration and a service context. Judge whether a user can find,
apply, understand and preserve a useful representation without learning several
incompatible meanings of “view.”
