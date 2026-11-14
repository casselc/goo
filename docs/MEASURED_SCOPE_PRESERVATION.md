# Scope and preservation controls

The fixed-question editor groups service, message, literal severity and supported
span shortcuts under **Quick filters**. Closing the disclosure preserves their
values. Run still uses the complete checked scope; literal producer severity text
remains separate from numeric OTel severity classification.

The capture form shows destination controls when they apply. **Capture only**
keeps evidence separate. Choosing the current investigation can expose its
question links; choosing a saved revision uses the existing revision attachment
workflow. Changing destinations clears a stale question-link selection rather
than applying an invisible target.

Recent / Pinned keeps row actions near their scope titles. These layout changes
reuse native controls and the existing journal, query, capture and navigation
owners. Opening or closing a disclosure, staging configuration and returning to
held data do not authorize telemetry reads or publication.

Counter rate presentation keeps exact values in the supplied data and detail
views. Compact display must not turn a nonzero rate into zero. The ≈ marker
identifies a rounded display; exact and recorded rate details remain distinct.
Record-time
constraints and interval bounds use the existing UTC presentation with exact
detail; stored values and fixed query bounds are unchanged.

```mermaid
flowchart LR
  R[Recent / Pinned] -->|Stage complete definition| F[Fixed question]
  F -->|Explicit Run| O[Observed population]
  O --> P[Preserve]
  P --> D[Relevant destination controls]
  D -->|Explicit confirmation| C[Existing capture and attachment]
  C -->|Return| H[Retained investigation]
```

Layout measurements compare the same applied/read state at 320, 390 and 1440
pixels. Edit and confirmation states are measured separately. Smaller controls
or shorter dialogs are not evidence of faster queries. Request counts, retained
row work and before/after positions are recorded in the
[storyboard](journeys/measured-preservation.md) and [validation](../VALIDATION.md).
See the [architecture review](reviews/MEASURED_CONSOLIDATION_429_ARCH.md) and
[UX review](reviews/MEASURED_CONSOLIDATION_429_UX.md).
