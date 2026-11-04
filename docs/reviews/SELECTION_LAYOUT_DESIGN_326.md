# Selection layout regression review — 326

Status: frozen source and all three regenerated actual-frame comparisons accepted.

The source-only consolidation reads one detached column layout per synchronous
Find/filter pass instead of cloning the same layout for each row, and stops
retaining an unused preflight plan on batch presentation state. The checked union
is still recomputed at activation; it does not become cached mutation authority.

No labels, markup, styles, focus handoff, pause/Find behavior, selection limits,
command leases or Keep ownership change. The 325 contract therefore remains the
acceptance basis. Measured clone reductions are work-count evidence, not a user
latency promise; noisy wall-time measurements do not establish responsiveness.
Compare the same three actual 325 journey frames against the previous committed
frames after the sole verifier regenerates them. No new product requirements,
product/browser edits or layout changes are part of this review.

## Regenerated actual-frame comparison

Accepted all three regenerated 325 frames against the committed `4cfffc7`
versions, viewed as before/after pairs. Control placement, visible focus, row
geometry, counts, refusal copy and hidden-selection/kept state remain equivalent.
The overflow still shows 21/24 matches plus one hidden selection and refuses a
22-item union; cluster union remains three selected with two visible; span Keep
remains three original observations with one hidden. Only expected run-specific
UTC values and fixture identifiers changed. No visible regression or new layout
requirement was found. Native focus/custody/source assertions are reported by the
verifier separately; final actor exit is not inferred from these images.
