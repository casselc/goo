# Shared record interactions — architecture review 298

Review of checkpoint `b4bd2c8`, before implementation. The recommendations below
are the proposed acceptance contract, not a claim of executed validation.

## Decision

Extract the existing application-owned record interaction controller from
`ui/record-results.mjs` and use it for Table and List. Keep one source-context
owner for trace expansion, linked logs and surrounding logs. Keep selection,
Keep submission, inspector content and field actions in one page controller;
presentations supply row/item action and inspector mounting points. A list adapter
must not render a hidden table or duplicate the selection state machine.

`goo-record-list` remains a passive supplied-data component. A standalone JSON
attribute or `.page` value grants display only, even when it resembles an admitted
response. The application gets its page through `SourceSnapshots.getRecordPage`
and holds result token, question and basis outside the public display model.
Source reads, Keep/retention requests, navigation and query changes stay outside
the public component and its dependency graph.

## Controller boundary

The controller receives an immutable page-local record sequence, signal, delivery
question, basis and optional result/capture authority. It receives explicit
callbacks for source context and scope actions, plus one application currentness
predicate. It owns selected positions, selection revision, submission lock/busy
state, shared-field disclosure, Keep construction and inspector content. It returns
selection state and presentation attachment/disposal operations. Presentation
changes and acquisition are separate operations.

Native Table/List surfaces attach the same controller-owned actions to their
current row/item mounts. Preserve the existing maximum of 20 selected observations,
explicit Clear, submitted Keep retry semantics, and app pending-state notifications.
No selection is persisted in view configuration or inferred from trace/span IDs.
Hide/show of the passive context column must not unexpectedly remove the sole
selection or Inspect controls; action placement is application presentation.

Do not require the controller to query `tbody`, `td`, `.record-detail-row`,
`goo-table`, or a List-specific class to establish authority. A presentation port
can provide a row/item anchor, action host and inspector host. Shared content uses
the existing structured record/attribute presentation. Layout-specific columns
remain in their adapter rather than the state machine.

## Observation identity

Use validated integer positions within this exact page for selection membership.
Two equal records, including the same supplied object occurring twice, are two
page observations. `Map<record, input>` and `rows.indexOf(record)` merge those
occurrences and are unsuitable for the extracted owner. A position is neither an
event ID nor authority that may survive a reread.

Live Keep indexes are page-local, bound to its result token and expected delivery
question/basis. Capture Keep indexes retain the existing capture offset mapping.
Never use the configured page-zero root, held aggregate basis or current mutable
saved capture to construct the intent for a newer inspection page. A later page's
count can disagree with held charts; selection does not reconcile those reads.

## Currentness and retirement

Use the exact private record owner identity together with generation, page serial,
admitted data identity, role visibility, connected live mount, application activity,
pending read/count state and absence of an active child. A DOM node's connection
alone is insufficient: removing and reattaching old nodes must not revive actions.
Initialize the private owner before mounting adapters that can call back.

Apply the same guard to selection changes, Clear, shared-field actions, Inspect,
trace/surrounding navigation, numeric plotting, field updates and asynchronous
completion. Shared-field closures additionally retain the selection revision;
Keep retains its exact intent and revision through retry. Existing inline facet
and selection-start callbacks should delegate to the strict common predicate
rather than reproduce a subset that omits private owner equality.

A component `.display` replacement currently rebuilds all List children. If an app
adapter attaches controls afterward, it must explicitly reattach from its current
controller and retire detached handles. Avoid source actions in component lifecycle
or implicit DOM observers. Back should renew only current-page application mounts
with fresh owners; it must not reauthorize previously retained event closures.
If a selected/submitted workflow cannot be preserved safely, refuse kind/layout
changes with a clear native pending state rather than discard it silently.

The optional XTDB `Retention.selection` currently checks only `host.isConnected`
and accepts no currentness callback. Sharing it without change would reproduce a
stale reattachment path. Extend it with the controller's current/revision guard
before exposing it from both presentations, including response/error handling,
or explicitly leave it unavailable in this cut. `Keeps.mount` already accepts an
active callback and expected question/basis and should remain the shared path.

## Acceptance and scope

Prove equivalent Table/List selection, Inspect, shared-field staging, Keep payload,
trace expansion and log context on actual telemetry. Verify one admitted page and
one selection owner, no reads on kind/layout/disclosure changes, and unchanged
held chart/read bases during paging. Exercise duplicate observations, row limit,
missing/invalid IDs, stale reattached DOM, late Keep responses, pending submission,
child navigation/Back and List display rerender. Test public serialized/detached
List use separately to prove no source capability was added.

Extend the actual journey, Bombadil actions/invariants and existing model/property
checks around the shared owner. Use negative controls for stale and mismatched
page/token/basis behavior; a test that only observes a click is insufficient.
No new endpoint, backend store, query shape, event identity, completeness guarantee
or generic interaction framework is needed for this milestone.

## Implementation review

The extracted `record-interactions.mjs` now supplies one selection/inspection/action
owner with two small surface adapters. RecordResults retains the context reader
and delegates both presentations through the existing mount function. The List
offers inert row-item action/inspector hosts; public component lifecycle still
does no acquisition, navigation or publication. The shared controller uses integer
positions, including duplicate occurrences, with one 20-observation selection
machine. Live and capture Keep payloads retain their respective position mapping.

Source review found two retirement gaps during implementation. First, a detached
shared-fields disclosure captured its revision when opened, allowing an old
control to adopt a later selection; an old Clear could also clear later selection.
The revised code captures revision when constructing the selection UI and checks
revision and membership before acting. Second, a List display/page replacement
could replace slots while leaving the outer component connected. Currentness now
checks the exact page, display and immutable item-host array captured by its
surface. The shared target WeakMap owner additionally retires all previous mounts,
including direct Explore/capture tables without a population owner.

All shared action wrappers now require the mounted owner, connected target and
viewport, target containment, current surface generation, application activity and
the host's private page predicate. Field actions and plotting reuse that guard;
asynchronous Keep and optional retention require the original selection revision.
Retention checks its active callback at request initiation and result/error
presentation. The application renews record mounts after child Back, keeping old
closures retired rather than reactivating their DOM. Source review of this
extraction has no remaining authority blocker. Executed browser/model results
and any follow-up lifecycle adjustments belong in the milestone validation record.
