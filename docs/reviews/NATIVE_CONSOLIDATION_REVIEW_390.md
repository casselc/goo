# Native presentation consolidation review — 390

Architecture/usability review, November 9, 2026. Application and CSS source
has been reviewed, including the bounded Saved ordering follow-up, with no
remaining source blocker identified. The parent reports final actual native
entry/browser and held-return regressions pass. The parent also reports final
independent source/geometry admission, required model replay and image checks pass. This reviewer has not run builds, tests or
browser campaigns; reported outcomes below belong to the implementation,
component-test and parent-run acceptance owners.

## Outcome and scope

Keep native observations ahead of optional authoring, identity and provenance.
This is a presentation consolidation over the existing owners and finite native
families. Do not change counter arithmetic, gauge freshness, histogram coverage,
capture admission, HTTP contracts or source execution to obtain a compact layout.

The main source of excess height is identifiable in current source:

- Ordinary Measurements mounts library selection, definition, time-window inputs,
  actions, status and scope before the acquired result.
- The result adds title/state, warnings, freshness prose, pinned-stream text,
  inventory statistics and filtering before its first stream.
- An open stream adds full attribute inspection, expanded diagnostic explanation,
  Focus controls and display-save fields before its timeline/table.
- Service native widgets already present latest evaluated values in a bounded
  table; expanded retained evidence delegates to the same measurement renderer.
  Saved evidence similarly delegates with `summaryProvided`, after its receipt.

Reordering the shared renderer therefore benefits all three surfaces without
inventing separate page renderers or new data models.

## Minimal component and layout contract

Use existing `goo-chart-frame`, table, inspector and disclosure primitives unless
actual shared markup requires a small passive helper. No stateful summary element
or new layout framework is needed. Application-owned classes can delimit the
result, stream, context and display toolbar; component CSS must remain scoped to
its own boxes and direct children.

After a successful acquisition, ordinary authoring can collapse behind one
natural question disclosure. Preserve its actual controls and IDs so definition
and time-window edits still revoke old capabilities. Do not collapse active
refusals, retry controls or an in-progress operation merely because no result is
painted. The empty result is data with an explicit no-data/unknown-health meaning.

The first visible native content should carry a compact title, fixed UTC window,
state and count, then a short stream identity and primary visual or observation.
Unknown/stale state and lack of continuous or whole-window coverage remain visible.
Full resource/point identity, diagnostic explanation, focus/clear-stream controls,
read clock/basis and configuration can remain accessible behind native disclosures.
Avoid repeating the same explanation at measurement, stream and chart levels.

Display kind remains a local supplied-data choice. Saving configuration belongs
behind an explicit disclosure after the visual. A default title must not keep
describing a previous kind after Timeline/Table changes; an authored title must
not be overwritten by that automatic update. Keep uncertainty distinctions added
in 389: an unavailable calculation is not an unrecorded source value, and an
optional exact rate not supplied by the evaluator is not zero.

## Ownership boundaries to preserve

`Measurements.mount` owns source generation, abort/retry and the accepted result.
Its `canAct` acquisition lease is deliberately separate from the 388
`canStageObservation` capability over an unchanged held gauge result. Presentation
helpers must receive and honor those callbacks; do not merge them into one broad
"active" flag while deduplicating the gauge/counter display controls.

Three retained-return mechanisms have different lifetimes:

| Owner | Existing responsibility | Safe consolidation |
| --- | --- | --- |
| `native-record-navigation.mjs` | In-tab native parent, hash/history, opener/scroll restoration and record suspension | Passive return-bar layout/content and callback delegation |
| `resource-return.mjs` | One active resource origin per parent, revision/hold restoration and ancestry | Same passive layout, preserving its WeakMap/revision ownership |
| Saved/service native context adapters | Held evidence admission, exact stream/export membership, action leases and current-source checks | Shared renderer presentation; retain their admission and leases |

Do not move history, restoration, active origins or action guards into a new
navigation manager or custom-element lifecycle. Different return labels can stay
specific—Back to measurement/service/saved evidence—while their visual hierarchy
and scope disclosure agree. This preserves the useful domain name without
pretending all owners have the same retirement policy.

## Acceptance that can establish the claim

Capture the old reviewed binary and new binary at natural initial scroll on
desktop and 390-pixel screens. Record the actual scroll container, visual bounds,
first observation, relevant disclosure state and page overflow. Forced scrolling
can inspect a plot, but cannot establish entry visibility.

Check ordinary acquired, no-data and refused states; service latest-value and
expanded retained content; and saved native evidence with its receipt intact.
Also check standalone/server-rendered and nested SPA component composition under
real application styles. Numeric and unavailable rails must share physical time
coordinates, not merely equal percentage attributes. Long labels, keyboard focus,
coarse-pointer targets and return restoration need concrete checks.

Local disclosure/kind/selection changes must add no source or custody reads.
Edits, Apply, pending/replaced results, hidden owners and retired callbacks must
retain their prior revocation behavior. Run remains explicit; held Back must keep
evidence, clocks, selection and focus without rereading. Generated properties and
measured Quint replay can verify those bounded contracts, alongside source/raw
checks; none alone proves arbitrary UI schedules or complete source collection.

## Initial review decision

The presentation-only approach is appropriate. No Go runtime or library upgrade
is required for this slice; the project already targets Go 1.27.3. Shared styles
and passive display helpers should be introduced only where the same mechanics
and authority boundaries have been demonstrated. Final implementation source,
before/after frames and actual runtime/model outcomes remain to be reviewed.

Inputs: [active roadmap](../ROADMAP_2026_11_09_CHECKPOINT_387.md),
[gauge observation context](../GAUGE_OBSERVATION_CONTEXT.md),
[counter display reuse](../NATIVE_COUNTER_VIEW_REUSE.md), and
[validation](../../VALIDATION.md).

## Source follow-up during implementation

The implementation is using owned native-result/stream/context/toolbar classes
and existing disclosure/chart/table primitives. The two return adapters only gain
shared presentation classes; their history, WeakMap, hold and action-lease logic
remain in their existing modules.

Saved native evidence retains a visible receipt prefix and explicitly labelled
custody UTC time. Its measurement summary carries environment, fixed window,
state and count. Full receipt/read/basis information moves into an explicit
provenance disclosure; it is not rewritten or conflated with evaluation time.

Two bounded review findings were sent to the owners:

- Once a successful result hides duplicate `measureScope`, later result-clearing
  paths need to reveal that compact scope again. The UI owner confirms Apply,
  new Run and Cancel now do so and is adding mounted regression coverage. The
  visible result header remains the scope presentation while evidence is painted.
- The newly important context/setup/result/capture disclosure summaries need
  adequate coarse-pointer targets in their own scoped styles. Do not solve this
  with a broad global `summary` rule that changes unrelated screens.

The all-unknown, non-brush numeric plot may collapse its unused vertical space
provided it retains time-axis width, the explicit no-numeric-data notice and the
selectable unavailable rail. Its CSS condition preserves recorded-zero marks
because those have a `data-value` attribute; it excludes the actual `brush`
attribute used by the component. Physical width/alignment, server-rendered
composition and keyboard outcomes still require runtime checks.

One runtime risk remains to inspect rather than assert as a bug: opening the full
question setup automatically after a source refusal can push the refusal/retry
below the narrow viewport. Active error visibility takes precedence over exposing
all editing controls; the geometry acceptance should check the actual failure
state and keep its compact scope visible.

The UI owner has removed the forced setup opening on read failure: the disclosure
retains its prior state, while the refusal/retry and fixed scope remain separate.
Initial failed reads can still leave the initially open authoring panel intact;
the actual first/repeated failure geometry must establish visibility rather than
assuming it from DOM order. The design owner has added scoped coarse-pointer
minimum sizes and focus styles for the newly important disclosures.

One component-boundary improvement was proposed to the parent: compact chart-frame
header/title/meta/note styling can belong to a passive `density="compact"` CSS
profile in `chart-frame.css`, rather than an application stylesheet reaching into
the frame's chrome. It needs no new state, subscription or navigation API and can
work equally in server HTML and an SPA. The native owner then selects density and
retains only its own margin/height layout. This is a recommendation pending the
parent's scope decision, not an asserted runtime failure.

## Final source review

The compact density recommendation is implemented in `chart-frame.css` as an
opt-in `density="compact"` profile. The application selects that attribute; its
native stylesheet no longer reaches into the frame's private header/title/meta/
note chrome. The profile changes spacing and font sizing only, preserving the
same supplied state, unit, window, note and owned chart/action children. The
component-test owner reports the standalone/server and nested SPA checks pass
at 1280 and 390 pixels, preserving focus, children and open details with no API
calls. This reviewer inspected the implementation and test source, not that run.

The outer stream summary has also been added to the scoped coarse-pointer
minimum target rule. Long measurement titles and stream identity can be visually
clamped, while their full DOM text/title remains available; unit and condition
are outside the clamped identity and are never hidden by that rule. The table
header override preserves literal unit casing within the native owner.

The final application changes preserve source/evidence responsibilities:

- Setup closes only after successful render. Apply, pending Run, Cancel and
  refusal reveal compact fixed scope when no result header is painted. Source
  refusal preserves the prior authoring disclosure state. The parent reports
  mounted coverage of those transitions in the 39 focused checks and the full
  1,404-test frontend pass, with zero skips.
- NativeSave now uses a kind-independent default title, so local kind changes
  cannot leave a stale Timeline/Table default. Existing authored input stays in
  the same control. Save still checks its acquisition, visibility and inert
  guards; gauge held staging retains its separate capability.
- Focus/Explore-all controls moved into disclosures and now explicitly honor
  acquisition/visibility/inert guards before staging a new question. No read or
  publication occurs from disclosure or custom-element lifecycle.
- Saved native context moves existing scope/receipt nodes into a disclosure after
  full capture admission. The visible receipt/custody line uses the original
  immutable values. Non-native and kept replacements reset the presentation
  marker only as their accepted content becomes current; a refused replacement
  retains the accepted evidence and styling.
- Service changes replace only the inventory statistics chrome with a compact
  count sentence. Latest-value meaning, held result admission and read leases
  are unchanged. Both return adapters only acquire the shared CSS class.

No remaining source blocker was identified in this bounded review. Physical
natural-entry, failure/empty/unknown geometry and actual held-return/read-budget
acceptance are still required; unit and component tests do not establish those
claims on their own.

## Saved ordering follow-up from actual geometry

The parent reports ordinary wide/narrow entry checks pass, but Saved narrow entry
still leaves a later numeric mark and its unknown rail below the viewport. The
proposed remedy is real DOM ordering, not a forced scroll, CSS-only order or a
smaller plot: place accepted native evidence before the capture catalog in the
same Saved workspace parent, with full provenance after the primary result.

The accepted-only boundary matters. A pending or refused replacement must not
restore/reposition the currently held native evidence as if new content had been
accepted. Existing `loadKept` called `placeReceipt()` before awaiting the Keep;
that should move to its accepted-success branch before old summary nodes are
cleared. Accepted non-native/Kept/operation content can restore the original DOM
anchor order. Existing nodes should move rather than be cloned, preserving
listeners, current data, focus and logical tab order.

The compact Frozen receipt/custody line also needs to survive a failed replacement.
Loading/unavailable notices currently replace that status text while the old
native evidence remains painted. Preserve the accepted native ID prefix and
original custody UTC time, adding the operation notice separately or alongside
them. Raw receipt/read/basis data must remain available in its disclosure. These
are bounded presentation/fidelity fixes, not another request owner or store.

The fresh implementation addresses those findings. The private
`native-capture-layout.mjs` helper reorders existing nodes within the same Saved
workspace, retains hidden home anchors, and restores accepted non-native layouts.
It appends the existing scope wrapper after native data and moves the existing
recovery notice above primary content even when currently hidden, so a later
owner-driven visibility change needs no new observer. It owns no listener,
request, history, evidence or action lease.

Call sites select native ordering only after accepted page validation, and restore
other layouts in their accepted content branches. `loadKept` now places the
receipt after its successful read/current-visit guard. Loading and failed
replacement notices use a small status presenter: the same immutable receipt
prefix/custody-time nodes remain and only the notice text changes. Full receipt,
read clock and evidence basis remain in their explicit disclosure. Ordinary
non-native statuses retain their prior text meaning.

Source syntax checks pass. No further ownership or clock blocker was identified;
the owner's dedicated DOM-order/status-identity tests and refreshed actual Saved
narrow geometry remain required before acceptance. In particular, source-level
use of state-preserving DOM moves is not itself a claim that runtime focus,
custom-element children or every viewport outcome have been verified.


## Parent-reported actual acceptance follow-up

The parent reports the final main journey passed in 1.3 minutes with 14 natural
entry frames and 27 measured boundaries. Six fresh main Playwright frames and
two standalone component frames were inspected by the parent. Ordinary counter
first data is now approximately 649 pixels down the desktop viewport and 581
pixels down the narrow viewport, versus 1,775 and 2,609 before; saved counter
narrow is approximately 614 pixels, versus 2,205. This closes the geometry
follow-up at the browser/report level without a forced test scroll to the plot.

Final frontend checks passed 1,406 tests with zero skips (29.319 seconds), embedded
UI/browser API checks passed in 0.631 seconds, and API documentation covered 112
operations. Existing actual saved context, service context, gauge reuse and
counter reuse journeys passed in 19.0, 19.3, 13.1 and 15.9 seconds respectively.
Their existing scope/read/owner/held-return assertions were retained. Actual wire
checking admitted 137 responses and 43 request bodies.

The parent reports Bombadil completed 662 snapshots through five phases with
DOM/source outcomes, one source read, zero store reads and zero UI writes.
Camera frames were stale and are excluded from pixel evidence. A DOM field named
`paint` is not itself camera-freshness evidence; the separately captured fresh
Playwright frames support the visual review.

These are parent-reported results. This reviewer performed source inspection and
documentation, not the runs or image inspection. At that checkpoint, the
independent baseline/final source/geometry ledger, required Quint replay and
Docker/local image verification were still outstanding; the root acceptance
closure below records their later outcomes. The browser and actor results alone
do not imply those outcomes. See [validation](../../VALIDATION.md).


## Parent-reported marker-alignment follow-up

The separate independent ledger caught a real service composition issue after
the initial visibility pass: generic service button margins shifted the chart's
numeric and unavailable marks by approximately 3.5 pixels. The component now owns
zero marker margins, with no change to source data, scales or model tolerance.
The parent reports a rebuilt-binary main pass (1.3 minutes, 14 entries and 27
boundaries), a fresh service-context pass (13.1 seconds test, 15.5 seconds runner),
and all three component cases passing in 7.6 seconds. The strengthened nested
real-product-CSS case takes 2.2 seconds and checks zero/known/unknown geometry
without source reads.

The latest actor has 719 snapshots and five phases. Its camera contains four
unique images, unlike the earlier all-identical 662-snapshot camera run, but no
completed-actor pixel verdict is claimed. DOM/source outcomes and fresh
Playwright pixel captures remain separate evidence. At that checkpoint the
independent baseline/final ledger, required Quint replay and final image gates
were still pending; their later outcomes are recorded in the root closure below.
These outcomes are parent-reported, not additional runs by this reviewer.


## Root acceptance closure

The parent reports independent baseline admission passed in 0.03 seconds and the
final source/geometry admission plus required Quint replay passed in 5.31 seconds
(package 5.355 seconds), with zero skips. All 14 entries and 27 measured facts,
raw source membership, exact marker X/Y, whole-item clipping, capture clocks/
bytes and physical SQL were admitted. Six semantic mutants and an unintended
local-read model control were rejected.

The final trusted-CA Docker build and Compose configuration passed. Updating only
the packages container preserved dependencies, stores, volumes and the XTDB
orphan. Local health and seven changed embedded UI/CSS source-byte checks passed.
No bounded 390 acceptance item remains pending. These are parent-run results;
the reviewer performed source/document review, not the executions or fresh-image
inspection. Earlier pending statements above describe their historical checkpoint.
Actor completion remains DOM/source evidence; it is not a completed camera-pixel
verdict. See [validation](../../VALIDATION.md) for the final evidence boundaries.
