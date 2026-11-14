# Unknown-data entry — bounded acceptance

430 implements the first usability tranche of the
[ad-hoc telemetry review](ADHOC_OTLP_REVIEW_2026_11_14.md).
The [contract](../UNKNOWN_DATA_ENTRY.md) and
[executable storyboard](../journeys/unknown-data-entry.md) describe its scope.

## Independent UX follow-up

Accepted without blockers after source review and inspection of generated entry
frames at 1440/320 and paused frames at 1440/390/320. The same native chooser is
adjacent to Environment, with an unscoped accent and no extra introductory prose.
Controls wrap, records remain visible, and the Resume focus ring is retained.
Named accessible labels and row headings preserve exact scope meaning.

## Independent accessibility follow-up

Accepted for the bounded fixes. The reviewer checked source, ran the focused
actual-handler tests successfully, and viewed entry at 320px and paused Activity
at 390px. Explicit Pause/Resume transfers focus only when the disappearing control
was the actual focused activator. Other/programmatic/noncompact/refused transitions
do not steal focus. Unchanged status writes are deduplicated; idle Find updates
use `aria-live=off`. Named preview buttons expose full values and scope cells are
native row headers. Discovery read/preview/stage/inspection controls and reply
adoption honor native availability.

This is not a full accessibility certification or speech-output test. Changing
Find totals can still announce every batch; they are not rate limited.

## Integration ownership

EnvironmentAdoption remains the existing application-owned controller. Placement
and disclosure do not query telemetry; explicit Read and Preview do. Stage checks
the preview's typed policy and leaves Start explicit. No backend contract, scope
mode, query engine, public data format or capture family was introduced.

The real generic journey starts without a saved-view bootstrap and covers
unreported, empty, conflicting current/legacy and legacy-only scopes. It separately
accounts for the existing source read on switching an already applied signal.
Its actor must observe completion rather than infer success from clicks.

See [VALIDATION.md](../../VALIDATION.md) for executed suite results. Source-family
availability, metrics-only bootstrap, source-error categorization and sparse-data
recovery remain separate roadmap work.
