# Population and embedded-control styling — 339

Status: source and native computed-style/viewport acceptance pass.

The 338 frames exposed a full-width Comparison view selector. Source inspection
finds two distinct causes: the application's broad population-facet select rule
can stretch embedded controls, while the composed comparison also inherits a
page-wide `label { display: grid }` rule even outside that facet container.

## Ownership correction

The passive count-pair presentation now owns its direct label and select layout
in `components.css`: inline, wrapping controls with intrinsic width, a 32px select
target and the existing 44px coarse-pointer target. This applies to standalone
and nested use rather than patching only one application screenshot.

The recent-population application keeps full-width controls only on its own
known distribution and comparison editor paths. Its label, toolbar, result-tab
and status sizing rules follow those same owners. Direct application notes keep
their compact/mobile treatment; descendant passive notes no longer inherit a
mobile `display: none` rule. Chart rows retain their intentional application
layout. No general CSS reset or new widget family is introduced.

The existing scope disclosure keeps its native summary marker. Named summary
spans compact the environment and fixed window, expose the full environment/policy
through the title, wrap UTC text on narrow screens and preserve a 32/44px summary
target. Exact policy/window rows and actions remain application-owned.

## Validation contract

Native computed-style checks must compare an equivalent supplied count-pair
presentation standalone and inside the real recent/facet ancestor classes at
1280px and 390px. Compare select width and font, actual notice visibility, keyboard
focus, coarse-pointer targets and document overflow. Check that application-owned
editor selects still fit their intended parent. A known count whose geometry is
withheld must keep its numeric state while its decorative track stays hidden.

Measure a fresh saved-view entry separately from task-scrolled interaction
frames; the old interaction screenshot does not measure pristine entry height.
Disclosure, local Find and view changes must cause no source reads or writes.
The architecture/native review owns those effect and retained-state claims; CSS
source inspection alone does not establish browser behavior.

## Observed acceptance

The final fresh-document fixture measures the same genuine 107/13 saved pair at
scroll zero. Data begins 28.78px earlier desktop and 32.20px earlier narrow.
Equivalent standalone/nested count selectors both measure 99×37px with 12px font,
inline-flex labels and visible notices at both widths, replacing full-panel
948px/310px controls. Keyboard scope inspection preserves exact facts and issues
no source reads; local staging/Back and the independent known/unknown observations
also retain their existing effects. Known 27 with withheld geometry has a computed
hidden track, actual data-value 27 and normal known styling. No overflow is observed.

Coarse-pointer 44px rules are retained by source inspection; a separate touch/coarse
browser emulation is not claimed. The narrow comparison still starts below the
844px initial viewport. See [validation](../../VALIDATION.md) for the full native,
actor, observed-model and changed-window regression results.
