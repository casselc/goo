# Roadmap after the 411 checkpoint

The [408 cadence](ROADMAP_2026_11_11_CHECKPOINT_408.md) delivered service captured-record comparisons (409) and incoming HTTP response context (410). [Quantitative composition consolidation](QUANTITATIVE_VIEW_COMPOSITION.md) (411) is delivered, with actual desktop/mobile comparison, held ownership and saved highlighting checks. The [review reconciliation](reviews/ROADMAP_RECONCILIATION_411.md) records the architectural and product decisions behind this next finite queue; [validation](../VALIDATION.md) records checks actually run.

Keep the loop: **current ordinary telemetry → narrow a population → inspect or compare → preserve/reuse**. Review and annotation begin from a selected immutable captured occurrence; they do not replace Activity with a labeling workspace. Manifests remain optional. This prerelease requires no legacy formats or compatibility layer.

## 412 — breadth: contextual review-task discovery

User task: “Use an existing labeling vocabulary for this captured example without finding and copying its artifact ID.” Add a bounded, source-free task catalogue and chooser inside the existing **Review example** disclosure. Keep exact-ID loading available. Reuse the existing workbench file-store scan bounds and validated immutable task loader; catalogue presentation is not a new task store or a source query.

Show useful task titles and vocabulary context while keeping immutable task identity available. Require explicit selection: neither a catalogue refresh nor a similarly named task may silently change the selected vocabulary. Empty, unavailable, invalid and scan-limit states must leave the current task and local judgment intact. Task declaration and exact-ID recovery remain existing paths. Do not infer categories from tags, absence or prior records.

Acceptance: real captured-record entry, bounded catalogue admission, duplicate task names with distinct immutable IDs, explicit selection, and no telemetry reads. Slow, failed or departed responses cannot adopt into a closed/replaced inspector or overwrite a newer choice. Preserve uncertainty, abstention, explicit Save, pending-publication recovery and unverified local attribution. Update the HTTP/OpenAPI contract and relevant store/property/actor checks for the new metadata operation.

## 413 — depth: adjacent occurrences in one immutable capture

User task: “Review the next captured example without repeatedly closing and reopening the record inspector.” Add Previous/Next within the current immutable capture through the existing Saved record/inspection owner. Keep the selected task only through an explicit, visible continuation policy; judgments and publication targets belong to individual occurrences.

Use validated absolute retained indices, not visual row order or reported trace/span IDs. Preserve the original capture, raw record identity, source-read basis and custody receipt. Beginning/end boundaries are explicit. Navigation cannot silently discard an unsaved judgment or detach a pending publication from its target; design the bounded draft/pending disposition before implementation. No implicit Save or automatic label propagation.

Acceptance: actual adjacent retained logs/spans, repeated reported IDs remaining distinct, first/last boundaries, original absolute indices including nonzero captured offsets, explicit task continuity and draft/pending handling. Navigation performs only existing stored-evidence/metadata reads, never telemetry acquisition. Retired callbacks cannot publish or adopt into the replacement occurrence. A reviewed subset is not an exhaustive training set, and absence of a decision remains unlabeled.

## 414 — consolidation: compact shared record operations

User task: “Inspect and act on a record without a stack of competing buttons and repeated explanations.” Consolidate record-action hierarchy across Activity, record detail, captured evidence and review entry using the existing shared record interactions and passive disclosure components. Keep the primary inspection flow prominent; group secondary investigation and preservation commands where operators naturally look for them.

Use modern semantic HTML, accessible names, predictable keyboard/focus behavior and compact shared styles. Preserve meaningful differences between question Save, independent evidence Capture and local review publication. Do not hide action availability or move source/navigation/publication authority into display components. Avoid a global action registry or lifecycle manager.

Acceptance: real desktop and 390px frames across the participating surfaces, keyboard/disclosure/focus paths, unchanged source/read/write budgets, and held-return behavior. Local controls remain usable where their existing contracts permit; retired source controls remain retired. Reduce repeated prose and stacked controls without removing clocks, uncertainty, exact target identity or pending-save recovery needed to make a decision.

## Review gates and deferred candidates

Finish and review each milestone before advancing. Keep heavy checks serial and preserve telemetry, stores and streams. Review the whole product again after 414 rather than extending this queue automatically.

The large historical verification package remains a concrete maintenance concern. Extract one semantic family only if private helper coupling materially blocks this queue; otherwise defer it to a separately bounded change with measured cold/warm compile time and memory, independent oracles and negative controls preserved. New features should continue the isolated semantic-package pattern used by 409/410 rather than adding private numbered-helper dependencies.

Fixed-summary captured-rate comparisons remain a useful alternative, especially for unequal windows, but do not displace this cadence: 409 already supports preceding-window captured-count comparisons. Exponential histogram execution, grouped P95, arbitrary executable kinds, authenticated collaboration, broader platform health and external warehouse deployment retain separate design/source gates. No current measurements justify a new query engine, global source manager or runtime performance rewrite. Go 1.27 facilities should be used for concrete reliability, performance or usability benefits, not as a reason for public type machinery or wholesale migration.
