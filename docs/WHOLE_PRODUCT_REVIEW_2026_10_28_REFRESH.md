# Whole-product review after 234

Reviewed 2026-10-28 against `db03e6f`. This refresh supersedes the execution
recommendations in the earlier same-day review; its findings and external-source
research remain historical. This is an expert task/source review, not customer
research or an accessibility/performance certification.

Goo's core remains: **observe now → isolate a population → follow supporting
signals → save or reuse the understanding**. Ordinary OTel works before a manifest.
The next work should improve that loop rather than extend identity cleanup or
add another independent dashboard/editor.

## What was inspected

Read-only Chromium walkthrough of Activity, service discovery, Saved, Definitions
and Operations at 1440×1000, then retained Activity at 390×844. Existing local
OTel traffic supplied observations. No workloads, captures, annotations or
investigations were published, and no stored data was deleted. All six checked
frames had no page-level horizontal overflow; no uncaught page errors were seen.
Only Activity and service discovery performed source reads in this walkthrough.
Operations was inspected at its unrun entrance. It was not executed here.

Source reconciliation included the active roadmap, older backlog/product/library
reviews, capture list storage/HTTP behavior, Activity/source owners, trace search,
comparison limits and validation through 234. Recent 233/234 journeys supply
separate evidence for host/native/service/Kubernetes return behavior; they were
not rerun for this documentation refresh. No new competitor documentation was
fetched: the earlier review's dated references and verification limits still apply.

## Current observations and decisions

| Area | Current evidence | Decision |
| --- | --- | --- |
| Activity | Desktop stream is substantially denser than the older review frame. At mobile width each short record still repeats Open trace/Select/Inspect as stacked buttons. Toolbar disclosures also leave distracting separator fragments. | First improve mobile scanning and contextual actions using the existing record/display primitives. Preserve inspect/select/trace access and keyboard/touch discoverability. |
| Saved | Useful capture rows appear early, but the first page shows 20 items in stored-ID order. Dates are visibly nonchronological and there is no capture search. `CaptureStore.List` sorts IDs; HTTP accepts only a cursor. | Make “find what I just saved” and “find evidence for this service/environment” first-class tasks. Search/order must cover the bounded catalog, not only one visible page. |
| Definitions | Separate families, search and paging are delivered. The local test volume has 187 immutable record revisions, often sharing a title. Search explicitly covers the loaded catalog. | Preserve exact revisions and drafts. Do not merge by title or rebuild another definition browser. Test-volume repetition is not itself evidence of a production feature gap. |
| Services/resources | Discovery distinguishes observed and declared-only services and retains namespace identity, units and fixed windows. Rows can be tall, but the recent return/scope work resolves the primary ownership gap. | Keep delivered dashboards. Consider a measured inventory-density follow-up after Activity; do not automatically add more OTel profiles. |
| Comparison | Existing rankings/partitions/interval views are useful. The documented absence of minimum-support controls means a visually large relative change can come from few records. | Add descriptive support/denominator context and an optional support threshold. Do not claim statistical significance, anomaly probability or causal explanation. |
| Trace search/reuse | Bounded multi-condition trace search is delivered. Its documented/UI contract does not save a reusable trace-search definition. The old C5 proposal remains relevant. | Select one recurring cross-service question and make that definition reusable through existing saved-question surfaces. Do not archive a witness page as a complete trace. |
| Operations | Explicit package selection and unrun state remain clear. Ordinary OTel entry does not depend on visiting this surface. | Retain as optional application-contract workflow; no separate redesign is justified by this entrance review. |
| Libraries/testing | Prepared/native package names, checked definition/question references, runtime ownership and pinned opt-in verification are delivered through 232. | Close the finite cleanup queue. Source/dedup/custody hashes stay audited follow-ups; no universal store/query/history abstraction. |

### Startup observation

One fresh local load made seven API calls: configuration, capabilities, measurement
packages, telemetry packages, two retained-evidence status reads and one tail POST.
It loaded 137 ESM resources and had 3,060 DOM nodes at the sampled Activity point.
The final navigation sample had 3,378 nodes. These are one local session's counts,
not throughput, memory, scan-cost or latency benchmarks. The tail is the only
initial telemetry request; the other six calls are configuration/catalog/status.
The duplicate retention-status read remains a narrow enabling follow-up. It does
not justify a new global cache, bundler or an eager optional XTDB dependency.

## Reconciliation of earlier recommendations

- Delivered: source cancellation through JSON and hidden follow-ons; compact shared
  record rows; first startup deduplication; Saved/Definitions entrances; contextual
  lookback and exact UTC scope; retained resource/native returns and visible focus.
- Delivered bounded library work: `subjectfacts`, storage-free native evaluation,
  explicit runtime/default binding, checked definition revisions, private generated
  seeds and checked fixed-question keys. Preserve identifiers and dependency seams.
- Delivered bounded verification: root fast CI, pinned offline/woven entry points,
  and a read-only running-lab API check. Fresh provisioning/restart acceptance and
  hosted execution of every optional suite are not established by those checks.
- Existing dashboards include bounded HTTP/RPC/database/messaging/runtime/host/pod/
  container profiles. Older statements that Kubernetes or pool work is wholly
  planned are superseded; newer convention versions need their own source contract.
- Still relevant: capture discovery, comparison support context and recurring trace
  question reuse. Broader numeric complements remain a separate semantic proposal.
- Preserve but defer: annotation/review and optional bitemporal retention work.
  Classifier exports require a selected label/version/consent/use-case contract;
  this review does not make XTDB mandatory or discard the user's training goals.
- Shared production needs authentication/query isolation/retention/restore choices.
  A collaboration canvas, alerts, RUM/profiling, arbitrary expression engine and
  live Databricks execution remain separate product decisions.

## Implementation direction

The [roadmap](ROADMAP_2026_10_28.md) owns order; the
[backlog](BACKLOG_2026_10_28.md) owns bounded acceptance. Start with Activity mobile
scanning, then capture discovery, comparison support, and recurring trace-question
reuse. Keep ordinary OTel and exact namespace/attribute/unit/time contracts in
every cut. Use focused ownership/property/model/browser checks where they address
an actual failure class. Keep screenshots and storyboards current without retaining
large disposable diagnostic bundles.

## Inspected frames

[Desktop Activity](journeys/screenshots/235-review-activity-desktop.png) ·
[Mobile Activity](journeys/screenshots/235-review-activity-mobile.png) ·
[Services](journeys/screenshots/235-review-services.png) ·
[Saved](journeys/screenshots/235-review-investigations.png) ·
[Definitions](journeys/screenshots/235-review-definitions.png) ·
[Operations entrance](journeys/screenshots/235-review-operations.png)
