# 455 scoped acceptance

455 adds a local Points display for already-held raw gauge observations. It uses
the common `goo.dataset/1` model and `goo-chart` kind routing, with a passive web
component and the existing application-owned raw inspector.

## Architecture and OTel semantics

Independent review accepted conservative eligibility: only finite, known-recorded
values plot. Missing/unknown flags, no-recorded-value and invalid/nonfinite values
remain explicit. Raw text, exact timestamps, attributes and occurrence keys remain
available; coordinates and coordinate collisions never establish unique producers
or equal raw values. Extreme values and subnormals retain finite geometry.

Review corrected an initial extent assumption: the raw API refuses 10,001 rows
atomically, so exactly 10,000 does not prove clipping. Every admitted bank retains
unknown extent; the read-budget fact is separate from the 1,000-location display
budget. Domains use all eligible held values, and remaining occurrences stay
counted and inspectable.

Generic Points/Table routing was tested rather than assuming registry admission
was sufficient. Raw imports retain the normal browser-held downgrade without
inventing a source basis or completeness. A profile-specific bounded import
budget supports 10,000 rows; ordinary limits remain unchanged. Pure histogram
facts were separated from rendering so schema admission stays display-independent.

## UX, accessibility and web standards

Follow-up reviews accepted desktop and 390/320px frames and actual interaction
probes after these fixes:

- Points clearly identifies the whole held bank when Table Find is active.
  Switching back preserves Find and explains an excluded selection.
- Native radio navigation retains radio focus. Roving arrows/Home/End navigate
  locations without activating a point.
- Repeated locations expose explicit occurrence choices. Every row also has a
  native local value/inspection alternative; nearby screen positions can overlap.
- Explicit selection reveals and focuses the shared inspector heading. Close
  restores a visible actual opener, equivalent row control or current display
  radio, including when the original opener belongs to the hidden display.

The final journey checks focus/visibility/hit targets before screenshot framing,
with zero source requests for local interactions. No scoped blockers remain. This
is not a broad WCAG or cross-browser audit, nearest-point selection, connected
series, fleet aggregation or source-completeness claim.

## Verification boundary and next step

The new journey uses real OTLP/ClickHouse membership and browser assertions plus
Bombadil. It does not force local plot gestures into a new state machine. Existing
model-backed gauge/sum/histogram context regression remains separate. Hegel tests
use synthetic banks admitted through the Go raw-source path; they are not live
collector evidence. Executed receipts are in [validation](../../VALIDATION.md).

456 should visualize one raw explicit histogram observation with exact bucket
counts and recording/bucket validity. Reuse the common grammar without constructing
an evaluated measurement, merging occurrences or deriving fleet percentiles.
