# 452 scoped acceptance

452 adds an explicit lookup for one reported OTel link target outside the held
trace page. The next roadmap slice remains inverse reference discovery (453),
then measured reference consolidation (454).

## Independent source review

- **Architecture:** accepted strict position-correlated link arrays, exact
  nonzero canonical target tuple, conjunctive target query, preserved environment
  policy/window and detached intent. Source acquisition stays in the app owner;
  generation and cancellation checks retire responses and cached actions.
- **UX:** initial review rejected indistinguishable multi-link actions. The
  shared waterfall now shows a link ordinal and short span ID, with the full
  tuple in its accessible name and tooltip. Source follow-up accepted the change.
- **Accessibility and web standards:** initial review identified lost keyboard
  focus when Run disables itself and missing scroll reveal on stage. Stage now
  uses shared revealFocus; keyboard Run moves focus to Back before disabling,
  and completion does not steal focus. Source follow-up accepted these fixes.
- Architecture also requested bounded reference rendering. Only the first 32
  links offer actions, with an explicit omitted count. A unique held tuple may
  be selected locally; duplicate tuples never silently choose a first row.

Rendered review additionally found a raw provenance hash overflowing narrow
results, cramped desktop columns at 320px, and Return focus recentering the held
trace. The hash now lives in the existing provenance disclosure, target rows use
the shared mobile record layout, and an already-visible held opener regains
focus without scrolling. Regression checks retain strict containment and scroll
assertions. UX and accessibility/web-standards follow-up accepted the actual
1440/390/320 frames and verified keyboard/held-return behavior. No scoped
blockers remain; final actor/model completion is a separate validation gate.

Browser and real-source results are recorded in the
[validation record](../../VALIDATION.md) and
[executed storyboard](../journeys/span-link-target.md). Source review alone is
not a browser, screen-reader or production authorization audit.

## Scope

The target panel reuses the existing trace dialog, shared record table,
attributes, focus helper and source-refusal presentation. There is no new query
language, product proxy or passive-component source client. Target lookup is a
separate read; Back restores held trace DOM and provenance without rereading.

Inverse reference search, recursive traversal, complete-trace reconstruction and
causality/delivery inference are outside 452. Lookup limits remain visible; the
configured source boundary is unchanged.
