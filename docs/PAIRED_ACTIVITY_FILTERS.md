# Separate filters for Logs + Spans Activity

Use paired Activity to watch related logs and spans under one shared environment/service selection and recent-window policy. Each lane can add its own conditions: for example, error logs beside slow server spans.

Open **Log filters** or **Span filters** in the lane header. Use the common visual condition editor or its CEL authoring mode. The compact applied condition remains visible when the editor is closed.

- Editing a condition does not start or restart a tail. An existing live connection may continue polling. CEL checking may call compiler/context endpoints; it does not acquire telemetry.
- **Apply log filter** starts a new recent log window. **Apply span filter** starts a new recent span window. The other lane keeps its existing read and held rows.
- **Discard log draft** or **Discard span draft** restores that lane's applied condition without reading telemetry.
- Resume refuses a lane with a pending draft. Apply or discard the draft first. Saving a view is also disabled while either lane has a pending draft.

## Shared scope stays protected

Each physical query combines the shared condition and its lane condition with ordinary predicate AND:

```text
Logs query  = shared environment + shared condition AND log condition
Spans query = shared environment + shared condition AND span condition
```

A lane condition cannot replace the shared scope. Shared predicates must work for both signals; lane predicates are checked for their own signal. The combined query must fit the ordinary limits of 32 conditions/groups and four nested levels. There is no silent truncation or partial acceptance.

The lanes retain independent read windows, pause/resume state, and bounds. Applying one filter starts that lane's current recent window; it does not synchronize the other lane's clock. Paired Activity is not a merged chronological stream or a completeness claim.

## Save and reopen

A saved paired view preserves the shared selector, each lane's applied predicate, and its column layout. The library preview and detail show lane conditions so they are visible before recall. Reopening stages the view paused; explicit **Start** acquires telemetry.

Saved rolling views contain compiled ordinary predicate ASTs. CEL expression text and authoring provenance are not saved in this format. No held rows, read windows, cursors, or execution state are saved.

See the [executable storyboard](journeys/paired-activity-filters.md) and [paired Activity contract](PAIRED_ACTIVITY.md).
