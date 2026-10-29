# Environment selector wire review — 249

Architecture approved explicit version/policy/scope with response-only metadata
outside the echoed question, derived from the actual checked source selection.
The browser must compare exact requested policy/scope, not merely accept a valid
DTO. Shared lowering must derive bindings and refuse collisions. Public main-flow
rollout remains gated on saved-codec adoption or server-side refusal.

Frontend found no source/read/control regression and approved the pre-render
metadata check. Go/frontend identified case-insensitive Go field aliases and
lossy lone-surrogate replacement; Go additionally found duplicate JSON object
merging. Corrections use exact raw-map keys, pre-lossy Unicode checks and final
raw-map field decoding to match browser last-value semantics. Valid pairs,
replacement characters and escaped literal backslashes remain supported.
Shared raw fixture controls and final focused Go race/browser tests pass.

Reviews were read-only. Frontend ran the focused protocol test; reviewers did not
run the root suite. Actual broader checks are recorded in VALIDATION. This does
not approve effective Query/Tail input acceptance or complete main-flow A0.
