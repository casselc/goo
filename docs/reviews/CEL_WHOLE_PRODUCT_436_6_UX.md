# 436.6 independent UX, accessibility and web standards review

Accepted for the bounded 436.6 investigation/authoring/preservation loop after
the shared severity-label correction. This review covers existing 436.4/436.5 real
frames, current source, independent focused checks and the new actual Services /
native record-context frames. It is not a physical assistive-technology or cross-engine audit.

## Product coherence

The product has a defensible primary loop: watch records, scope a population,
inspect, refine, then preserve or reuse. CEL is an optional additional-condition
representation inside existing Filters, not a competing query application. The
current Activity frame keeps the authoring controls folded until requested.
Conditions remains the default. Fixed questions and recalled saved source use the
same native mode choice, textarea, validation and existing Run owner.

This retains the useful product patterns established in earlier Datadog,
Honeycomb, Grafana and HyperDX comparisons: the data stays central, scopes remain
visible, advanced authoring is contextual, and reusable definitions do not pretend
to be captured evidence. This round did not conduct fresh vendor research.

- Persistent textarea and native select/fieldset/details controls are preferable
  to a custom tab/editor system here. Invalid text survives mode switching; visual
  edits require an explicit discard while source owns an additional condition.
- The short inline diagnostic, associated description, `aria-invalid`, and closed
  full diagnostic disclosure are coherent. Explicit checks announce a result;
  typing does not announce the full compiler output repeatedly. The textarea is
  compact and resizable, and review of captured 320/390/1440 frames found no new
  horizontal overflow.
- Parent scope is supplied independently of editable source. Additional conditions
  cannot accidentally replace unsupported attribute or presence predicates.
  Fixed bounds and Activity's fresh rolling Apply remain different user actions.
- Save is still the existing ExplorationView action. The conditional “CEL source
  included” details and explicit source discard communicate preservation without
  another library or publication workflow. Display Apply retains the destination;
  adopting another question is explicitly Open/Edit original.
- Valid-unsaved and invalid draft recall both receive the same owned keep/discard
  choice. Restored bytes, stored provenance and fresh private invocation ownership
  are distinct. Native textarea line-ending normalization must not silently alter
  an unchanged imported source.
- Native disabled states and independent direct-callback guards match. Held rows
  remain inspectable while question-dependent Run/Next/Tail actions withdraw.
  These source/DOM checks support the interaction contract, not screen-reader
  speech ordering or comprehensive keyboard conformance.

## Material acceptance finding

The existing 436.4/436.5 frames reveal a semantic presentation contradiction:
records with numeric OTel severity 9 (Info) display producer text FATAL in red.
`goo-record-summary` selected its tone with a regex over `SeverityText` in
`record-presentation.mjs`, even though classification and CEL filtering correctly
use `SeverityNumber`. A human scanning the stream could reasonably infer Fatal.
This conflicts with the whole-product numeric-severity acceptance criterion.

The bounded resolution is shared record presentation: show the admitted numeric
band, preserve distinctly attributed literal producer text, and derive tone from
the numeric value. Missing/invalid numeric values must not acquire urgency from a
producer string. Accessible text/tooltip should name both facts. Loaded Find must
match newly visible canonical band labels as well as literal producer text. Span
status should use its own admitted exact enum rather than inheriting log rules.
No change to source records, queries, saved meaning or aggregate counts is needed.

The shared correction is accepted after independent source and test review.
`recordLabel` uses the existing pure OTel classifier, shows **Info · producer:
FATAL** for this conflict and keeps Info neutral. The tooltip retains the exact
numeric and producer facts. Recognized numeric Error/Fatal and Warn own urgency;
missing/invalid values stay neutral. Span urgency uses exact reported `Error`,
independently from log severity. Loaded Find recognizes canonical numeric bands
and the literal producer text. No raw records, query meaning or counts change.

The label uses plain attributed visible text, not `aria-label` on a generic span,
a new ARIA role or an extra tab stop per row. An intermediate generic-span naming
attempt was removed during review. This is a standards-grounded source decision,
not a claim about every screen reader. The final real context frames are accepted below.

## Bounded follow-on priorities

These are not reasons to expand CEL's vocabulary during this consolidation:

1. Unify compact fixed-scope presentation. The fixed editor still repeats service,
   message and severity shortcuts plus a large selector chip above Filters. The
   Activity entry is materially more compact. Consolidate the common scope/editor
   once, preserving the original parent details and existing execution policy.
2. Reduce shared preservation prose. The menu correctly separates question,
   evidence and definition but repeats “no telemetry read,” configuration warnings
   and capture explanations. A short stable label per action plus one precise
   evidence explanation would retain the distinction with less operator reading.
3. Prioritize reliable incident recall and return over another authoring platform.
   Operators need retained scope, useful population comparison, inspection and
   predictable recovery across Services, native measurements and saved questions.
   Reconcile the older exact-recall/paired-capture backlog against the now-delivered
   source-preserving owner rather than adding duplicate save surfaces.

## Independently executed checks

```sh
node --experimental-test-module-mocks --test \
  tests/cel-expression-editor.test.mjs \
  tests/cel-preservation-integration.test.mjs \
  tests/cel-provenance.test.mjs \
  tests/fixed-question-editor-358.test.mjs \
  tests/fixed-authoring-read-guards.test.mjs
```

All five file runners pass, zero failures/skips, **0.376 seconds**. These are
focused native-control, persistent-state and callback checks, not a new browser,
source, physical AT or cross-engine campaign. Earlier real frames were inspected
at their captured widths; actual new browser/model results belong in VALIDATION.


The corrected shared presentation and passive graph were independently checked:

```sh
node --experimental-test-module-mocks --test \
  tests/record-severity-tone.test.mjs \
  tests/record-presentation.test.js \
  tests/display-modules.test.mjs
```

All three file runners pass, zero failures/skips, **0.532 seconds**. The cases
cover numeric and string inputs over all 256 UInt8 values, missing/malformed data,
producer conflicts, exact span status, visible-band Find and actual custom-element
label rendering. The public graph remains pure; no transport or application owner
is imported by the shared record label.

## Native record-context frames

The actual `4366-native-records-390.png` and `4366-native-returned-390.png` frames
were inspected after the semantic-label fix. Info rows now visibly attribute the
conflicting FATAL text to the producer and remain neutral. The shared table fits
the narrow viewport. The returned counter view retains its selected interval,
producer facts and UTC bounds; associated records are expressly described as
context, not inferred cause. No duplicate CEL workflow is added to measurements.

A nonblocking density follow-on is native-rate number formatting: the current
chart ceiling and selected-interval caption can show sixteen decimal digits.
Compact human-readable scale/summary values with exact details would improve
scanning without changing rates, units, reset semantics or source facts. This
belongs with the planned shared-presentation consolidation rather than another
CEL feature. The final Services record/returned frames are accepted as well. They use the
same record table and native selected-interval detail rather than introducing
separate authoring chrome. Histogram bucket ranges remain explicit, with no
interpolated percentile or inferred causal membership claim.

The real journey reports a passing 26.2-second run with exact parent snapshots,
source totals independent of the retained 100-row page, no extra reads on Return
and restored opener focus. Those execution assertions are the journey owner's
evidence; this independent review inspected the four actual 390px frames and
source. No additional browser session, physical AT or cross-engine check is
claimed by this reviewer. No open UX/accessibility/web-standards correctness
finding remains within this scoped acceptance.
