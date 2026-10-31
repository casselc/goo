# Saved investigations

Implemented: local immutable server-captured operation investigations. Inspect an
operation package, choose an error interval/context, trace or scoped subject and
optionally a member-comparison explanation. **Capture and save investigation**
re-reads the selected absolute window and displays exactly the evidence stored.
This explicit recapture can differ from the preceding live result if data arrived
in between. The save is not a byte-for-byte freeze of the earlier browser result.

The receipt includes the package definition/revision and manifest, normalized
operation spans and conformance, source query/parameters, counter/ratio/fleet
results and histogram results when enabled, separate component bases, navigation
selection, server capture timestamp and optional predecessor. Reopen renders
these stored results without telemetry queries or loading the current source
package. This is reproducible inspection, not re-execution of the evaluator from
all raw input points; raw metric inputs and the renderer binary are not archived.

Refresh retains the original package revision, source question and half-open
absolute window. It creates a successor, preserving its original parent receipt.
The UI compares component bases, added/removed span identities and component
errors. Differences can reflect late arrival, deletion/retention, reset/gap repair
or changed source evidence; no automatic cause is asserted. Unchanged bases can
still produce a distinct receipt with a later capture timestamp. If the currently
installed package revision differs, refresh returns a conflict, while the old
capture remains reopenable. There is no automatic package migration.

Selection retains operation/outcome/time/duration filters, opened trace or scoped
subject, selected ratio explanation and bounded member-comparison explanation.
All fleet comparison rows remain in the captured results; comparison selections
match complete stream identity and actual window, never just a shared timestamp.
An explanation whose interval is absent after refresh is not fabricated. Delayed
capture/reopen responses do not overwrite a newer live package selection; the
receipt remains available by ID.

## Storage and API

The single Go server owns capture/storage, API routes and the embedded UI. `POST /investigations` accepts `{question,selection}`;
question is `{package,packageRevision,fromMs,toMs}`. `POST` with `{parent: id}`
refreshes the stored question/selection; changing its question is rejected.
`GET /investigations/{id}` reopens a verified receipt. Browser routes add `/api`.
A link `/#investigation=<id>` opens that immutable revision.

`goo serve-http -investigations-dir <directory>` defaults to
`artifacts/investigations`; Compose mounts the named `investigations` volume.
Files use canonical-content SHA-256 IDs, exact JSON numbers/string nanoseconds,
0600 file permissions, atomic no-overwrite publication, file/directory sync and
read-time integrity checks. The API accepts no file paths or client-supplied
source results. The question is bounded to 120 minutes, request to 16 KiB,
selection to 8 KiB and capture to 24 MiB. Existing source limits remain strict.
Source operations failure prevents a save; optional metric failures are captured
explicitly as unavailable, never zero/healthy. Component reads are sequential
and share a question window, not a transactional database snapshot.

The hash detects represented receipt corruption; it is not a signature or source
provenance proof. Capture time is local server receipt time, not event time or
immutable source ingestion custody. Completeness/sampling remain unverified.
Stored errors/bases preserve uncertainty. A local filesystem cannot protect
against an administrator rewriting both files and references. There is no
multi-user authorization, operation-capture listing, deletion/retention/total-disk quota,
encryption/key management, backup service, collaboration or portable export UI.
Use this shared loopback lab accordingly. The named volume survives service
restart/recreation; it is not a backup and is removed by explicit volume deletion.

## Executable evidence

[Journey 19](journeys/README.md#19-save-reopen-refresh-and-compare)
saves real application evidence and trace context, restarts the packages service,
ingests an explicitly labelled late test span via the real OTLP collector inside
the old event window, reopens unchanged data without operation/metric HTTP reads,
then refreshes and verifies the new span and changed basis in a successor while
the predecessor remains equal. The artificial span is a controlled telemetry
fixture, not a claim that an additional application request happened.

The immutable store has race-tested persistence/concurrency/corruption tests;
Hegel state/swarm explores refresh/reopen/rejected question changes; Quint validates
observed save/reopen/refresh/reject transitions, with negative controls for changed
original evidence or question. See VALIDATION.md for actual run results.

Operation waterfalls can now explicitly expand into [general live trace/log context](EXPLORATION_DESIGN.md). This creates independent live evidence; it is not part of the stored operation capture and does not alter its receipt.

## General span/log and inventory captures

The operation-specific format above remains supported. Manifest-free Explore and
inventory evidence now use a separate typed receipt, local catalog and bounded
stored record bank under the same volume. See [GENERAL_CAPTURES.md](GENERAL_CAPTURES.md)
for stable paging, fixed-plan refresh, link semantics and limits. Neither format
automatically archives independently expanded context or a browser baseline pair.


### Independent operation receipt presentation

Saved operation captures render in their own supplied-data presenter; they do not
replace or move the live Operations editor. **Open live Operations** returns to
its existing package/lookback/selection without querying or staging the receipt
as new live evidence. Fixed-question refresh still explicitly publishes a child.
See [ownership and process diagram](LIVE_SAVED_OWNERSHIP.md).
