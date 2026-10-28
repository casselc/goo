# Independent whole-product product/usability review — 235

Reviewed 2026-10-28 against `5d12ce9`. Owner: independent product/usability review.
This report challenges the current roadmap rather than treating its order as an
accepted conclusion. It proposes changes; it does not modify runtime behavior.

## Scope and confidence

Read the current roadmap/backlog/review, product essence, generic startup,
Activity/filter/selection/capture/question/document/annotation code and recent
233/234 validation/storyboards. Independently exercised the running local Activity
in Chromium at 1280×900 with existing OTel traffic: built/applied a service filter,
switched signals, and separately selected an observed occurrence. Captured the
actual tail request bodies and selection state. No source workload, capture,
annotation, document, retention save or configuration change was published.
No runtime code or existing review files were edited.

Labels below distinguish **live verified**, **source verified**, **documented
boundary**, and **proposal/hypothesis**. This was not customer research, a complete
accessibility audit, a production scale test or fresh competitor research. Existing
whole-product screenshots and 233/234 journeys inform the wider surface review;
they are not represented as independent fresh executions here.

## Product judgment

The product's essence is sound: ordinary OTel should let an engineer scope an
incident, inspect a population, follow related observations, then retain and reuse
the understanding. Goo has enough primitives to support that loop. The main risk
is no longer missing chart types. It is **continuity of intent** across signals,
preservation methods and saved objects.

I would not keep mobile row density ahead of every other task. It is a visible
quality problem, but the live signal-switch behavior below can silently change
which system the user is investigating. That should be fixed first. Capture search
remains high-value; exact-observation preservation also needs an explicit product
contract before the main loop is considered complete without optional retention.

## Findings

### 1. High: changing signals discards a valid shared population and starts a broader read

**Live verified and source verified.** Applied the ordinary shared service filter
`column:service = lab-consumer` on Logs. The requested `/api/tail` body was:

```json
{"environment":"local","where":{"field":{"kind":"column","key":"service"},"op":"eq","value":"lab-consumer"}}
```

Selecting Spans immediately requested:

```json
{"signal":"spans","environment":"local"}
```

The filter chips were empty and the editor read “All records in scope.” No Apply
confirmation or page error occurred. This is not an unsupported log-body/severity
predicate: service is a valid shared column for both signals.

Evidence: `ui/stream-entry.mjs:92` calls `filters.reset()`, clears the editor and
calls `start()` on the active signal-change path. The generic inactive/staged paths
have different behavior. The source guards correctly avoid publishing stale rows,
but correct cancellation does not preserve the user's intended population.

**Task affected:** “These logs from service A look wrong; show me its spans.”
The scope changes from A to every service in the environment. At realistic scale,
this is a substantial comprehension and read-cost surprise even though the new
selector is internally valid.

**Proposal:** preserve validated shared resource/service predicates, or stage a
signal change with an explicit list of incompatible predicates and require Apply.
Never silently drop a common service/namespace selector. Do not blindly carry
signal-specific attributes or derived fields whose semantics differ. Acceptance
must record actual requested selectors, not only the visible chips; include nested
AND/OR predicates, missing/present-empty namespaces, saved-view drafts and a late
old-signal body. Keep the existing owner/cancellation guarantees.

### 2. High product decision: keeping an exact observed item depends on optional retention and an annotation

**Live verified:** the local retention status was `{"configured":true}`. Selecting
one observed Activity occurrence showed “1 selected · Add a tag or note to retain
this selection.” `Retain selected` was disabled. No save was attempted.

**Source verified/documented boundary:** `ui/tail.mjs:51–59` gates preservation on
configuration, nonexpired membership proof and a supplied tag or note. The proof
is deliberately short-lived. `ui/retention-status.mjs` says that with retention
unconfigured, local inspection remains available. `docs/TAIL_SELECTION_CUSTODY.md`
correctly distinguishes exact observed occurrences from physical event identity,
full traces or complete populations. The exact tail publisher uses the optional
retention path; general file-backed Capture is a separate operation.

The general Capture dialog in `ui/captures.mjs:57` honestly explains that it
rereads the fixed question and may observe late data. That is good semantics,
not a misleading implementation bug. It is nevertheless not a substitute for
keeping the particular anomalous row the user already saw.

**Blocked task when retention is off:** “Keep these three unusual observations as
a bookmark while I inspect the trace, without changing which observations I kept.”
The user can inspect locally or rerun/capture a query, but cannot make that exact
selection durable through the baseline path. On a configured instance, requiring
a tag/note also couples bookmarking to authored interpretation before the user
knows what the anomaly means. The optional-off case was verified from code, not
simulated by changing the live environment.

**Proposal/decision before code:** choose whether exact occurrence preservation is
an optional advanced capability or part of the core general-observability promise.
Expose that choice before selection, make bookmarking distinct from labeling, and
consider optional annotations with an honest default title. If a baseline fallback
is designed, it must preserve server-attested membership and immutable bytes; do
not silently replace it with a fresh query or client-asserted telemetry. An expired
selection needs an explicit next step, not an implication that rerunning recovers
the same occurrence. This does not imply making XTDB mandatory or inventing another
universal store. Existing custody/retry/citation guarantees should be retained.

### 3. High adoption boundary: “ordinary OTel” still requires a populated exact environment attribute

**Source verified/documented boundary:** `exploration/model.go:99–100` requires a
nonempty explicit environment. `exploration/compiler.go:101` requires the presence
and exact value of `ResourceAttributes['deployment.environment.name']`. Native
measurement definitions also require a nonempty Environment. Generic startup
correctly avoids loading sample manifests, but starting the generic stream asks
the user to supply an environment; it is not source connectivity or schema discovery.

An SDK can emit otherwise useful OTel without that optional resource attribute.
Entering an arbitrary environment cannot expose such records. This is an adoption
precondition, not a newly demonstrated source-security defect, and explicit scope
is valuable. The current API's deliberate isolation must not be removed by an
unreviewed wildcard or alias fallback.

**Proposal:** make the resource/source requirements prominent in onboarding and
empty-state diagnostics. Decide a bounded path: documented collector enrichment,
checked missing-environment selection, or explicit environment discovery under the
existing budget. Preserve missing versus empty versus named values. Do not treat
“no records in the chosen environment” as proof the source is empty. A concrete
non-demo SDK fixture lacking the attribute should be part of adoption acceptance.
This deserves an earlier decision than another convention-specific dashboard.

### 4. High: capture discovery is a real core-loop gap; keep P2

**Source verified and supported by current review frames.** `CaptureStore.List`
uses a 20-item ID-ordered page with a 10,000-entry scan bound; the HTTP list contract
accepts only a cursor. The catalog has no whole-catalog title/scope/recency query.
A user who saves and navigates away cannot naturally find “what I just saved.”

The roadmap correctly separates this from already-delivered optional retained-
evidence search. Preserve that distinction. Start with useful receipt metadata,
explicit captured-time ordering and stable filtered cursors. Do not silently sort
one page and call it recent, use file mtime as capture time, or merge same-title
artifacts. If a derived metadata index is needed, separately specify publication,
rebuild and integrity behavior. Do not make catalog browsing reread telemetry.

**Additional usability acceptance:** the successful save outcome should provide
an obvious durable link and a way to return to the current investigation. Finding
an object is not enough if the user cannot tell whether it is a definition, a
fixed evidence capture, a retained selection or a document revision.

### 5. Medium: P1 should optimize common scoping tasks, not only row/button geometry

**Source verified; prioritization is a proposal.** Activity exposes environment and
signal immediately, but the generic predicate editor is behind Filters. Building
one basic service predicate uses condition/namespace/field/type/operator/value
controls. The core Service cell is a passive qualified label in
`ui/record-presentation.mjs:16`; custom field columns and the record inspector have
refinement actions, but the common “only this service” gesture is indirect.

Current mobile frames also repeat Open trace/Select/Inspect rows. Compacting them
is worthwhile, but hiding actions can make scoping harder if those are the only
obvious discoverable affordances.

**Proposal:** evaluate “scope to this service/namespace” as the first contextual
refinement, along with one clear inspect affordance, compact selection and trace
continuation. Reuse qualified field/value actions and explicit Apply, not a second
predicate language or broad new toolbar. Test the complete task and its requested
selector before comparing row heights. Keep keyboard/touch alternatives and do not
make row click, selection and trace navigation accidentally synonymous.

### 6. Medium: preservation vocabulary still exposes mechanisms more than intent

**Source verified; comprehension impact is a hypothesis to test.** The UI offers
Save Activity view, Capture evidence, Retain selected, Save package revision, Save
investigation and investigation-document revisions. Some distinctions are required:
a configuration is not evidence, and an observed selection is not a reread result.
However, Operations' `Save investigation` (`ui/index.html:12`,
`ui/investigations.mjs:17`) saves its legacy fixed-window evidence path, while
`ui/documents.mjs` uses Investigation for an authored document of questions,
evidence and notes. The same word refers to different preservation outcomes.

**Proposal:** standardize action outcome summaries and destinations: save question,
keep observed items, capture fresh fixed evidence, add to investigation, save
investigation revision. Keep technical provenance/proof limits in progressive
disclosure. Consolidate terminology/copy and reuse existing dialogs before adding
another durable object family. Do not collapse genuinely different read/custody
semantics into one generic Save operation.

### 7. Medium: support context is more valuable than automatic low-support suppression

**Documented boundary/proposal.** P3 is justified: a large relative change can have
little support. Show exact side counts, denominators and unavailable coverage near
the ranked change before adding a control. Keep rare but important failures visible
by default; a new deployment's first few errors can be the incident, not noise.
Thresholding should be explicit and descriptive, with meaningful zero/no-data/
unknown states, and must not alter totals or imply significance. A threshold UI
is not an anomaly detector, classifier confidence or causality engine.

### 8. Lower priority: trace-question persistence is useful, but does not repair the earlier continuity gaps

**Source/document verified.** Bounded multi-condition trace search exists; its
contract explicitly does not save a reusable search definition. The proposed
slow-database-plus-error task is credible. Preserve existential witness semantics,
condition-1 parent scope, fixed windows, missing/empty fields and large numeric
values. Reopening should stage without executing. Use the existing question/
investigation surfaces rather than a parallel “trace workspace” store. This should
follow shared-signal scope and preservation/discovery work, as currently planned.

## Whole-product strengths and boundaries to preserve

- Activity is now the real entry point, with explicit pause/resume, coverage/eviction
  labels and source-owned continuation. Do not regress to an upfront package wizard.
- Host/native/service/Kubernetes returns now have a coherent bounded immediate
  parent and retained focus. General history/canvas work is not justified by a
  wish for every interaction to look identical.
- Existing OTel service/resource profiles are useful and honest about schema, units,
  populations, unknown coverage and declared versus observed bindings. Keep them;
  do not infer application health from span status or fleet percentiles from scalars.
- Ad hoc record/native definition creation already supports manifest-free workflows.
  No new “build a grammar” entry point is needed to satisfy that direction.
- Exact scope, immutable captures, read/custody clocks, independent annotations and
  versioned definitions are differentiators. Their details should remain inspectable
  without becoming the main visual content of every screen.
- Annotation/training goals span entities, observations, questions and documents.
  A reviewed export vocabulary can stay deferred; ordinary bookmarking should not
  be treated as a classifier assertion or forced to collect a premature label.
- Framework-neutral passive components and app-owned source/navigation state remain
  appropriate. There is no product justification here for a frontend framework,
  universal result hierarchy, new query optimizer or mandatory bitemporal database.

## Recommended order and concrete journeys

1. **Continuity gate:** scope Logs to one service, switch to Spans and back, preserve
   supported predicates or stage incompatibilities explicitly; verify actual calls.
2. **Adoption/preservation decisions:** document the exact environment and retention
   prerequisites, then choose the baseline path for missing-env telemetry and keeping
   exact observed items. Do not let optional backends silently define the core promise.
3. **Find saved work:** implement bounded general capture discovery and recency;
   save → navigate away → find → reopen must remain source-free on reopen.
4. **Common scoping plus density:** make service/namespace narrowing easy, then
   compact Activity controls/rows with same-data desktop/mobile acceptance.
5. **Support-aware comparison**, followed by **recurring trace-question reuse**.

Mobile density can proceed in parallel as a bounded display task. I would not let
it delay the signal-scope fix, nor spend another main milestone on duplicate status
reads, residual hashes or broad architectural cleanup. Those are enabling work,
not the highest-value user outcomes established by this review.
