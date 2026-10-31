# Preserve first; add a finding when needed — 293

Executable: [293-preserve-attachment.spec.js](293-preserve-attachment.spec.js).

1. Open a saved service population and explicitly Run its fixed query against a
   real OTLP span in ClickHouse. Open Preserve: title and Save exploration view
   are visible; the optional finding form is closed.
   [Default Preserve](../docs/journeys/screenshots/293-preserve-default.png).
2. Open Add to investigation, write a finding, close and reopen it. Its draft
   survives. A retained button cannot act while its disclosure is closed.
3. Change the view title and save a fresh immutable definition. The old button
   cannot attach the new ID with its old textarea. The fresh button works without
   reloading, stages the exact new ID/text, and an explicit Save revision stores it.
4. Open another definition. Both retained old buttons refuse, preserving the new
   finding. An unapplied destination note also refuses attachment without losing
   either draft. [Optional finding](../docs/journeys/screenshots/293-attach-finding.png).
5. Clear the destination draft and explicitly attach. The investigation contains
   exactly the selected saved definition and text. Only Save revision publishes a
   revision; no step captures evidence or implicitly reads telemetry.
   [Stored finding](../docs/journeys/screenshots/293-exact-finding.png).

The nine-step observed Quint ledger checks one explicit source read and two explicit
revision writes across closed, changed-owner and pending-destination refusals.
Six corrupted ledgers and an unfinished prefix must fail. Independent Bombadil
must observe the exact source ID and text in a card and a stored revision before
completion; a click alone does not count. Existing 282/288 journeys exercise the
same native disclosure in their complete preservation flows.

Validation: real journey and observed Quint pass (66 seconds including campaign).
Independent Bombadil passes 30 seconds / 621 snapshots / 4 phases, observing exact
ID/text and a stored revision with zero source reads, no overflow or page errors.
All three frames above are generated from the running Go app. Actual retained
child staging via grouping and Back also preserves the open disclosure and note;
the hidden parent button refuses. Keyboard activation was not separately tested.
