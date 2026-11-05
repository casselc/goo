# Fixed-population scope chrome — 339

Status: source, four final actual viewport frames, native IO/lifecycle checks
and bounded actor completion accepted.

## Bounded design

The existing applied/configured population disclosure already owns selector
inspection and local filter editing. Reuse it rather than add another panel.
Its native summary now presents Applied/Configured, signals, a compact environment
name and the selected fixed UTC span. A Previous window or Fixed baseline window
badge identifies the second definition, with its exact half-open bounds in the
title. Full labels distinguish named empty-looking strings, empty environment
labels and unreported environments.

Move the former always-expanded selected/baseline window paragraphs into this
same disclosure. Its expanded facts retain exact ISO UTC half-open bounds, each
environment's explicit current-only/effective policy and the baseline selector.
The existing selector editing, restore, expansion and removal controls remain
inside their original owner. A new passive presentation helper derives strings
and a full definition key; it performs no fetch, persistence or observation
projection.

Disclosure open state and summary focus are captured from the owner's prior DOM
before local rendering and restored only when all selected/companion queries and
the baseline definition match. No toggle listener, source owner, navigation
manager or saved state is added. Retired native disclosure elements cannot mutate
new controls. Different query definitions collapse the disclosure.

The population status stays outside the disclosure. In particular, a rejected
wrong-field gesture must keep its actionable guidance visible. Existing Run,
Use fresh, staged selection, Cancel and Back handlers keep their current guards.
The count comparison and facet components retain their own observation receipts:
page-level definition facts must not merge older held pair receipts with newer
independent category reads.

## Source checks

Three focused presentation cases exercise exact half-open UTC facts, same versus
previous-window labels, independent definition keys, empty/unreported versus
literal names, current-only/effective policies, cross-day formatting, source
immutability and zero IO. Existing population workspace tests pass in the same
focused run. An initially invalid test mutated a previous baseline into an
unsupported nonadjacent window; the fixture was corrected to the supported
same-window fixed baseline rather than weakening window validation.

Native validation must separately prove disclosure keyboard operation, unchanged
queries/IO, same-definition restoration, changed-definition reset and current
controls after lifecycle replacement. Fresh entry and naturally scrolled task
viewports are separate measurements. Desktop/narrow screenshots must demonstrate
actual content placement, not a cropped improvement. Source styling is owned by
the scoped-control review; final acceptance follows actual frames.

See [roadmap](../ROADMAP_2026_11_05.md) and the preceding
[category pair design](CATEGORY_PAIR_DESIGN_338.md).

## Actual viewport review

The refreshed desktop entry frame includes the application header and genuine
post-Run state. Scope facts occupy one compact summary; the unloaded facet is
honestly unloaded, and the held 107/13 comparison retains its own windows and
read clocks. The embedded Summary selector is compact rather than stretched
across the card. The exact-scope frame shows native keyboard focus on the opened
summary and full environment policy, selected/baseline half-open ISO UTC bounds
and baseline selector inside the existing disclosure.

The 390×844 entry keeps Applied/logs, the environment, Selected UTC span and
Previous window label readable without horizontal overflow. Its facet fits in
the viewport, while the comparison facts still require scrolling. Do not claim
that all results or all counts are initially visible. The partial-result frame
keeps known selected POST count 27 distinct from the unavailable baseline dash;
comparative bars are withheld, and separate held comparison facts remain 27/4.
These are actual current frames, not cropped mocks or actor-image evidence.

The native journey passes in 1.1 minutes. Fresh entry is measured at scroll zero
before and after: the desktop facet moves from 575.58px to 546.80px and comparison
from 810.11px to 781.33px (28.78px earlier). Narrow entry moves the facet from
669.44px to 637.23px and comparison from 941.16px to 908.95px (32.20px earlier).
That comparison still begins below the 844px viewport; this is a measured
improvement, not complete initial visibility.

The supplied component selector measures 99×37px with 12px text both standalone
and nested, with its notice still visible; the previous broad application style
stretched it to 948px/310px widths. Native keyboard scope disclosure checks issue
zero IO and retain exact policy/title bounds. Partial selected count 27 remains
known data while its track is withheld. Retry reads only the failed side; Back
reads nothing and returns focus. The bounded actor reports 383 snapshots through
all eight phases, four verified transitions and scopeVerified=true. These actor
DOM/effect checks remain separate from the four Playwright visual frames.

## Selected next breadth task — 340 (proposed)

Make the existing **Find traces containing…** workflow work from the typed
ordinary OTel scopes already used by Activity and service exploration. Source
review finds a concrete gap: `ui/trace-search.mjs` requires a scalar nonempty
`parent.environment` and builds an environment-only request; the Go
`TraceSearchRequest`/compiler likewise uses current-name selection, and Open trace
constructs a scalar continuation. Effective-policy, empty and unreported contexts
therefore cannot naturally use the existing trace-wide search.

Reuse the current typed environment selector and validation, trace-wide
conditions, checked witnesses and exact-trace continuation. Carry the original
policy, mode, fixed window and predicates through request, compilation, reference
validation, witnesses and the newly staged trace question. Do not add another
search engine or silently reinterpret an effective scope as current-only.

A concrete journey starts from an ordinary effective OTel span population, chooses
an observed field's Find traces containing action, adds one condition from a
different service, explicitly searches, inspects the independent matching
witnesses, and stages the exact reported TraceID for explicit continuation. Repeat
for empty and unreported labels, including current/legacy precedence conflicts.
The same reported TraceID is a telemetry relationship, not proof of causal order,
complete trace ancestry or a business identity join across hosts. Preserve source
retirement, bounded condition/result limits and retained parent return.

Native acceptance should prove the effective/empty/unreported policy end to end,
unknown/missing behavior, a decoy outside the original policy/window, independent
witness checks and no automatic read on continuation staging. Any ensuing saved trace view remains checked query configuration; fresh search
witnesses do not become frozen evidence. No new saved-search resource is required.
This is the selected next proposal, not implemented by 339.

The earlier [332 design](WHOLE_PRODUCT_DESIGN_332.md) also identified ordinary
previous-window mean-duration comparison. That remains a later depth candidate
after a focused measure-aware contract; it is not a parallel queue or a reason
to broaden the existing count kind silently. Reuse existing duration aggregates
and contributing counts, never average scalar percentiles.
