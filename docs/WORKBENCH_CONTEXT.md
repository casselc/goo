# Deployment context and investigation notes

Goo now keeps human interpretation and deployment receipts alongside immutable
telemetry evidence. These are local workspace capabilities in the Go runtime.
They reuse Saved investigations; no extra permanent navigation section or new
measurement language is introduced.

## Three separate objects and clocks

| Object | Meaning | Time and identity |
| --- | --- | --- |
| General capture | Server-read span/log/inventory evidence, or a two-population comparison | Fixed telemetry question bounds, source basis, capture receipt ID and custody time |
| Deployment receipt | An authored claim about a service or environment deployment | Explicit `occurredAtMs` event time and independent server `recordedAt`; immutable receipt ID |
| Investigation document revision | Ordered hypotheses, findings, conclusions, next steps and exact receipt references | Its own server recording time and revision ID; explicit parent/root lineage |

A late-recorded deployment does not change an old telemetry capture. A narrative
revision does not change its referenced evidence. Deployment proximity alone is
context, with no causal attribution or automatic failure classification.
An entered version/revision is an authored label, not inferred `service.version`
coverage or verification that the sample/production service actually deployed.
There is no verified author identity or automatic CI integration in this slice.

## Deployment context

Explore places **Deployment context** beside the existing fixed-time chart.
Applying a question prepares its explicit environment and half-open time bounds;
**Load deployment context** separately reads local receipts known now. This is
also available when inspecting a saved Explore capture, with an explicit fresh
context label. It is not automatically loaded or embedded into the capture.
Environment-wide and service-labelled receipts appear together; their service
labels do not silently filter the current query or prove service membership.

Record a title, optional service, version/revision, UTC event time and optional
HTTP(S) release URL. Custody time comes from the server and cannot be supplied by
the caller. Empty/unrecognized services are not inferred from telemetry.
The optional release link is displayed, never fetched by the server; URLs with
embedded credentials or executable schemes are rejected.

A 20-interval lane provides local marker navigation, with an ordered list of exact
event/custody times below it. Clicking a lane interval filters only this marker
list; **Show all markers** restores it. Counts, duration/error summaries and
captured question/basis remain unchanged. Receipt lists use event time in
`[fromMs,toMs)`, never their later recording time. A read reports `readAt`; receipts
recorded after that cutoff are excluded from that result. This does not promise
a transaction across concurrent filesystem publications or a verified clock.

Marker creation is append-only. Event times outside the current window are
allowed and identified as outside the current view. Duplicate submissions can
create separate receipts; no deployment-deduplication or retry-idempotency guarantee
is provided.
Corrections/retractions and external pipeline authentication remain future work.

## Ordered investigation notes

Use **Add to notes** from any general saved capture, including comparisons and
inventory, or **Add deployment to notes** from a marker. These append exact IDs to
the current draft while preserving its title/cards. Notes can also start from
**New investigation**, with a manual general-capture ID citation. Legacy operation
capture IDs are not yet supported by this document reference kind.

The editor supports:

- Plain-text hypothesis, finding, conclusion and next-step cards; HTML remains
  literal text. No executable/Markdown rich-text renderer is added.
- Frozen general-capture references and deployment-receipt references with optional
  captions. Reference existence/integrity is checked before publication.
- Editing notes and keyboard-accessible move-up/move-down/remove actions over the
  ordered list. Reordering does not change capture or marker IDs; keyboard focus
  follows the moved/remaining card. Opening a citation focuses its evidence heading.
- Saving new immutable revisions, opening an exact parent, browsing ID-ordered
  revision pages, and copying `#document=<revision ID>` links.
- Retaining dirty drafts when another revision is requested. An unapplied note
  must be added/applied before saving; Revert explicitly restores the stored base.
  Late revision reads do not pull the user back after navigation. Drafts live in
  the current browser page until saved, not durable browser storage.

Each save names an explicit parent; **there is no mutable latest head**. Two saves
from one parent create preserved branches, not silent last-write-wins updates.
The catalog lists every published revision and labels parentage. Automatic merge,
conflict resolution, verified authorship, presence and shared access control are
not implemented. This is a local revision forest, not a collaborative editor.
Retries after an unknown save response may create another immutable branch.

Opening a document reads its stored narrative and IDs without executing telemetry.
**Open frozen capture** opens the exact stored receipt, retaining normal stable
record paging; explicit trace/log expansion is still a separate fresh read.
**Show deployment receipt** reads the exact local marker. References do not follow
successor captures or refreshed marker lists automatically. Documents store IDs,
not additional copies of all referenced records; future retention must account
for those references. UI layout and live inspector state are not archived.

## API and storage limits

The Python lab proxies the same bounded Go APIs used by the UI:

| Route | Contract |
| --- | --- |
| `POST /api/deployments` | `{environment, service?, title, revision, occurredAtMs, url?}`; creates an immutable local deployment receipt |
| `GET /api/deployments?environment=...&fromMs=...&toMs=...` | Explicit environment, half-open window at most 120 minutes; complete local context or error above 200 matched receipts |
| `GET /api/deployments/{id}` | Exact verified receipt, without telemetry access |
| `POST /api/documents` | `{parent?, document:{title,cards}}`; creates an immutable narrative revision after resolving references |
| `GET /api/documents/{id}` | Exact verified narrative; no telemetry access or automatic source execution |
| `GET /api/documents?cursor=...` | At most 20 revisions per ID-ordered seek page; search in the UI is local to that page |

Card kinds are `note` (`role`, `text`), `capture` and `deployment` (`reference`,
optional `caption`), and `question` (`step`, optional `caption`). Documents require
a bounded UTF-8 title and a question or at least one card, with at most 32 cards.
Note text is at most 4,000 UTF-8 bytes; each artifact/request is at most 128 KiB.
Lists scan at most 10,000 directory entries and fail rather than imply a complete
partial list. Local files use content digests, atomic immutable publication and
integrity verification. A digest is not authorization or proof of authorship.
Storage lives under the existing investigations volume's `workbench` directory.
No new service, port, database or dependency is added. Production authentication,
quotas/indexing, authorization on every reference, retention and backups remain
separate requirements for shared deployment.

## Executable storyboard and verification

[Journey 36](journeys/README.md#36-deployment-context-and-ordered-notes) generates
work through the actual Go report sample and observes its exported OTel trace in
ClickHouse. The user records a deployment claim, pins it and a frozen capture,
edits/reorders notes, saves a successor and a separate branch, preserves an
unapplied draft, and follows stored evidence without rerunning the source.
A narrow-screen frame documents the same ordered workflow. No actual sample
rollout is claimed by this manual-marker journey.

Native tests cover scope boundaries, integrity, unsafe URLs, caps/paging, unavailable
references and concurrent branches. Hegel state/swarm exercises note edits/order,
save/fork/read history, marker arrivals and scope/window changes. Quint checks the
actual browser observe/draft/revise/branch/reopen ledger with six rejecting
controls. The opt-in Bombadil actor types, applies, reorders and reverts local
unsaved drafts; it never publishes notes, writes deployments or queries telemetry.
These are bounded verification results, not universal scheduling or collaboration
proofs. See [validation](../VALIDATION.md).

## Question-first investigation workspace

An investigation can now start with a title and question before evidence exists.
Its brief keeps **Question**, **Current finding** and **Open questions** together;
these are authored interpretations, not computed health claims. The active
investigation's title/question remain visible when navigating into Explore.
Return to investigation restores the current in-page draft without executing a
query. Drafts still require explicit saving to survive a reload.

The capture dialog offers **Capture only**, **Current investigation draft**, and
paged saved investigation revisions. Choosing the current draft appends an exact
capture reference; the investigation must then be saved explicitly. Choosing a
saved revision appends the reference in a new immutable child revision, preserving
its question, finding, open questions, cards and parent. The picker identifies
exact revisions; it does not infer a mutable latest head. An unrelated dirty draft
is retained. If attachment fails, the capture remains saved and can be added again.
Capture and document publication are separate operations, not a transaction.

Capture cards display fixed environment/time scope and capture time. Native
measurements show their source services, selectors and exact definition revision;
record captures show their predicate population. An ad hoc question does not gain
an invented package identity. Metadata uses `GET /api/captures/{id}/receipt`, which
validates the stored capture and returns its receipt/definition revision without
returning its record bank or querying telemetry. Full stored evidence opens through
the existing capture tools. Refresh and fresh-definition execution remain explicit
reads and cannot change an earlier citation.

`question`, `finding` and `openQuestions` are optional fields on document inputs,
limited to 4,000 UTF-8 bytes each. A revision requires a question or at least one
card, with at most 32 cards; the shared 128 KiB document bound remains. Empty fields
are omitted from canonical JSON, preserving old version-1 document hashes. New
question-only writes normalize omitted cards to an empty array. Earlier revisions
remain readable and branchable.

Storyboard 41 follows real checkout/report spans, report logs and an independent
OTel counter through capture, investigation handoff and fresh reuse. The same
workflow supports arbitrary bounded record and measurement questions.

## Contextual next questions

Evidence cards now offer **Investigate further**. Stored-only planning, explicit
scope previews and user-triggered execution connect duration/status summaries,
reported outcome attributes, exact trace/log IDs and fixed-window resource
instrument discovery. **Record missing evidence or instrumentation** appends an
authored question with the origin capture ID to the existing open-question brief;
it remains an unsaved draft until explicitly saved. See INVESTIGATION_FOLLOWUPS.md.

## Typed question steps

Documents also accept `{kind:"question", step:… , caption?}` cards. The step
contains version, kind, fixed scope, typed record/measurement definition and view.
It contains no evidence. [Investigation steps](INVESTIGATION_STEPS.md) document the
validated union, source-free reopen flow, matched/different windows, native sharing
and executable mixed-question journey. Question cards use the same ordering,
32-card limit, 128 KiB bound and immutable revision history as existing cards.

## Linked question evidence

Question cards accept an optional unique 64-hex `key` and an ordered `evidence`
array of at most 16 unique capture IDs. The store resolves linked captures before
publishing. [Linked evidence](LINKED_QUESTION_EVIDENCE.md) describes explicit
capture/child-revision publication and retained-capture recovery. Receipt metadata
exposes stored `readAt` and `summary` fields without source evaluation.
