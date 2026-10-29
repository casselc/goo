# Independent product/usability review — 243

Reviewed 2026-10-29 against `65ac193`. This is a whole-product task/source review,
not customer research, a new vendor survey, or an accessibility/performance
certification. It proposes work and changes no product behavior.

## Basis and limits

Read the product essence/direction, active roadmap/backlog, independent 236
reconciliation, generic startup, automatic-dashboard initiative, current capture
catalog contract, and 239–242 storyboards/validation. Inspected Activity,
preservation, capture, trace/finding and comparison owners plus shared presentation
contracts. Visually reviewed actual 241/242 frames and the parent's fresh 243 mobile
trace frame. No live data, definitions, annotations or configuration were published
by this reviewer; no runtime code or broad tests were changed/run.

The parent's new isolated real-OTLP243 task is distinct evidence, not my independent
reproduction: Logs → apply service scope → Spans → Inspect → cross-service trace
and linked logs → close trace → return to held Activity → fixed analysis → capture
→ leave → find/reopen passes. Its browser requests were three tail starts, two
supporting-context reads, one fixed analysis and one capture; return/find/reopen
add none. A separate parent-owned temporary catalog probe accepts 257 publications,
then refuses catalog GET with 400 while duplicate retry remains 201. The planning
reconciliation should use the parent's final artifacts/measurements for exact
results, not treat this report as another execution of them.

Surface coverage is deliberately unequal:

| Surface | Basis here | Judgment |
|---|---|---|
| Activity and inspection | Current owner/source,241 frames,239–241 validation, new combined task | A viable general stream entrance; preserve the delivered scope/focus work. |
| Trace/supporting context | Owner/CSS, fresh 243 frame and task | Normal close then parent-return works; overlay/parent interaction deserves a clear contract. |
| Saved captures | Current source/242 contract and real reopened frame | Search closes a major gap; the opened-evidence presentation is still unnecessarily heavy. |
| Services/resources/defaults | Initiative/source contracts and existing 234/187–216 acceptance | Reuse existing bounded defaults; no fresh execution of every profile/resource here. |
| Definitions/Operations | Scope/configuration/source documentation | Keep optional authoring/instrumentation separate from basic OTel adoption; no declared workflow was executed here. |
| Comparisons/meaning | Current breakdown/context/inspection code and prior audits | Support exists in several paths; consistency and task reuse matter more than another chart family. |

## Product judgment

The product now demonstrates the essential loop on its nominal path, with explicit
source ownership and better continuity than 236. The newly verified U0 availability
gap means the capture/search contract is not yet a closed correctness gate. Its next challenge is **making the same small set of actions
work for an ordinary installation, and making preservation outcomes easy to
understand**. There is no evidence here for a new Home screen, shared canvas,
query framework, arbitrary dashboard builder or another set of disconnected
semantic-convention panels.

Two decisions remain more fundamental than adding more profiles: how existing
unlabelled telemetry becomes observable, and whether keeping an observed item is
part of the baseline product. They should stop being incidental consequences of
an environment text field or an optional storage connection.

## Findings and recommended cuts

### U0 — High: an accepted capture can disable catalog search with no working recovery

**Actual-handler evidence from the independent Go and architecture reviewers; not
an experiment executed by this reviewer.** A valid 12,728-byte capture request
commits successfully with 201, but its 65,557-byte receipt cannot fit the 64 KiB
metadata entry envelope. The capture remains reopenable by exact ID with 200,
while search returns 503. Explicit rebuild indexes 0 and errors on the same limit.
The root/planning reconciliation should link those reviewers' precise reproducers.

This is not the deliberate metadata-versus-payload verification tradeoff. Nor is
it merely a request to support arbitrarily large stores. A capture admitted by the
existing public contract can make the ordinary find-again workflow unavailable,
including other captures, and the recovery instruction cannot repair that state.
The successful immutable publication is valuable and must not be relabeled failed
or deleted to make the catalog appear healthy.

**First bounded correction:** align accepted capture, metadata projection and
rebuild contracts before new features/polish. Prefer a bounded catalog summary
projection containing the identity/title/time and truthful declared scope needed
for search, rather than duplicating the complete question as catalog metadata.
Keep full questions in immutable evidence and retrieve them by exact ID. This is
an implementation/design candidate for the owners, not a prescribed new database.
Pre-admission before source execution is another candidate for truly unsupported
metadata, but rejecting future requests alone does not recover already accepted
captures. Raising a limit alone needs a demonstrated upper bound and recovery
contract; silent skips or page-local searching are not acceptable alternatives.

**Acceptance:** reproduce the valid oversized-receipt case; retain its successful
commit and unchanged bytes/ID; find it and ordinary captures across the catalog;
rebuild recovers the accepted store; exact-ID reopen still fully verifies content.
Any intentionally refused capture is refused before source work/publication and
has an actionable explanation. Missing metadata/corruption/budget refusals remain
honest, but a recoverable state must not be advertised with an unusable rebuild
hint. Preserve filtered cursor/order and source-free catalog/reopen guarantees.

### U1 — High adoption gap: ordinary OTel still needs an exact environment label

**Source verified; not a newly discovered security defect.**
`exploration/model.go` requires a nonempty explicit environment, and record/native
source adapters bind the resource environment. Generic startup correctly avoids a
sample manifest but begins with an environment text field, not discovery of the
source's actual scope. Otherwise useful OTel without that optional attribute
cannot be exposed by typing an arbitrary name. A chosen environment with zero
matching records does not diagnose absent enrichment or prove an empty source.

**Recommended decision:** preserve explicit isolation and make missing-environment
adoption a named task. Prefer a checked, explicit unreported-environment scope
for already ingested data, alongside bounded scope discovery and clear collector
requirements. It must distinguish missing, present-empty and named values, with
an explicit policy for supported legacy/current environment keys. Collector
enrichment alone cannot retroactively expose older records already in the source.
Do not introduce a magic environment string, silent wildcard, or broad source read
when the old exact selector becomes invalid.

**Bounded acceptance:** a normal SDK fixture with service identity but no environment
can be diagnosed and deliberately selected through the chosen supported path;
a named environment remains isolated; missing and empty are separate; record,
service and native continuations either preserve the same checked scope or state
that the continuation is unsupported. A scope transition is explicit before a
read. This deserves design/contract work before another dashboard profile.

### U2 — High product decision: bookmarking and interpretation remain coupled

**Source verified/documented boundary.** `ui/tail.mjs` requires configured retention,
nonexpired server membership and a tag or note before enabling `Retain selected`.
`retention/tail.go` connects to the optional PG/XTDB transport and prepares an
annotation as part of publication. File-backed general Capture is a different
operation: it rereads the fixed question, as its dialog correctly explains.

That leaves “keep these unusual observations while I investigate” unavailable in
the ordinary optional-store-off configuration. On a configured store, it still
asks for authored interpretation before the user may know what the anomaly is.
A transient selection, a fresh question capture, and an exact observed bookmark
are not interchangeable outcomes.

**Recommended decision:** make exact observed bookmarking a baseline capability,
with optional authored interpretation afterward, while retaining the existing
server-attested custody and retry guarantees. A bounded local storage adapter for
the existing exact-observation contract is a better first proposal than making
XTDB mandatory or inventing a universal history/store. This is a product proposal,
not authorization to persist browser-supplied rows as authoritative evidence.
If exact keeping is intentionally optional instead, say so before selection and
name the available alternative as a new read.

**Acceptance:** one observed item can be kept/reopened without a guessed label;
expired/unavailable proof has an explicit outcome; duplicate retries preserve
identity; no fallback silently substitutes a reread. System bookmark metadata is
not a reviewed training label, and lack of annotation is not a negative example.
Keep tags/notes useful across product surfaces without starting a training export
project in this cut.

### U3 — Medium: trace exit and parent return need one coherent interaction model

**Parent live evidence plus independently inspected source/frame.** The initial 243
probe tried clicking the global `Return to paused activity` while a full-screen
mobile trace inspector covered it. Source confirms the deliberate fixed overlay
(`product.css`, mobile `inset:0`) and its own Close control. The parent then verified
that **Close trace inspector → Return to paused activity works**. This is not a
proven total pointer dead end, and merely lifting the background button's z-index
would be an unjustified fix.

There is still a coherence seam: the drawer is `role=complementary`, occludes the
parent affordance, and has its own close/focus fallback while Activity separately
owns the retained parent. A user following a trace from Activity should not need to
understand an intermediate fixed-question workspace to regain the stream.

**Recommended bounded cut:** define the parent-aware close policy, then test it.
An Activity-origin trace should close to its held Activity context, or expose one
clear visible return after close. A fixed-question-origin trace should return to
that fixed question. Choose honest modal/drawer semantics: a modal presentation
should not leave an occluded parent action in ordinary keyboard traversal; a
nonmodal presentation must provide an understandable reachable parent path.
Do not make every passive inspector modal or add a global navigation/history layer.

**Acceptance:** normal pointer and keyboard X/Escape/return paths at 320/390/700px
and desktop restore the intended visible opener, held rows and original scope,
with zero resume/read side effects. Test the visible close path separately from
a programmatic click behind an overlay. Keep explicit Resume and source guards.

### U4 — Medium: reopened evidence spends too much space repeating its own meaning

**Screenshot observation, source corroborated; not a new measured geometry test.**
The actual 242 desktop frame shows a single saved log after a title/action masthead,
a stored-evidence note, frozen evidence summary, an additional saved-evidence note,
paging/reuse controls, a large `1` matching-count readout, deployment disclosure and
result tabs. The same count/scope/coverage message appears repeatedly. The record
body begins around the lower part of a 1000px viewport even though this is a one-row
capture. The improved Activity density should not be lost as soon as evidence is
opened again.

**Recommended cut:** consolidate the existing evidence masthead using the shared
passive evidence-summary/record primitives. Show type, environment, fixed window,
capture time, state and count once, with one clear source-effect statement. Keep
raw receipt/basis and detailed limitations in their existing disclosures. Do not
remove unknown coverage, confuse event time with capture time, or hide that rerun
is a new read.

**Acceptance:** compare the same one-row and longer captures on desktop/mobile;
reach their records promptly without traversing equivalent summaries; retain exact
UTC/tooltips, keyboard actions and all stored/inspection/full-population distinctions.
This is a finite presentation cut, not a global CSS rewrite or another custom shell.

### U5 — Medium: comparison support needs consolidation, not a new support primitive

**Source verified; no fresh statistical or comparative usability study.** Existing
comparison breakdown code already has side counts and status `error/count`, exact
shared scales, duration availability and explicit one-sided/unavailable behavior.
Context and stored-side inspection also exist. The roadmap's P3 should not assume
support is wholly absent or start by hiding rare groups with a threshold.

**Recommended first P3 slice:** audit the concrete compare paths and make exact
side support/denominators and unknown coverage consistently visible beside the
ranked change. Only then consider an optional clearly defined support filter.
Keep low-support high-change examples visible by default. Source completeness,
sampling and separate reads must remain distinguishable from sample count.
No confidence/significance label follows from a descriptive difference.

**Acceptance:** the same rare-failure example and a one-sided/zero/unknown example
across record, explanation and saved comparison paths give the user a comparable
support account; exact large counts survive; presentation choices do not change
population totals, captures or interpretation. Reuse current shared comparison
components before introducing another widget.

## Reconciliation and priority

- **Preserve delivered behavior:** 239 shared signal scope; 240 explicit Activity
  inspect/return; 241 compact row/attribute actions; 242 title/scope discovery,
  capture-time order and truthful metadata-versus-payload verification on the
  admitted nominal path. Do not rebuild these because older reports called them
  missing; 242 is not a closed correctness gate until U0 is resolved.
- **Finite correctness gate:** correct U0 first, then keep I1 malformed prepared text, and promote I2's
  evidence from source-confirmed to endpoint-reproduced using the parent's isolated
  257-publication probe. Fix these before expanding revision-producing features;
  do not broaden them into a generic hash/store cleanup.
- **Product order after that gate:** establish U1's supported adoption contract and U2's preservation
  decision early; take the small U3/U4 continuity/presentation cuts alongside them;
  then consolidated P3 support and the existing P4 recurring trace-question task.
  The complete normal 243 loop now passes, so an occluded-background click alone
  is not grounds for holding every other task.
- **Keep the right defaults:** existing HTTP/RPC/database/messaging/service/resource
  slices are a foundation. Add another profile only for a named diagnosis with
  available OTel identity/unit/version/coverage—not a competitor checklist. Manifest
  authoring, weaving and Operations remain optional aids, not prerequisites for
  ordinary data exploration.
- **Defer:** additional top-level surfaces, dashboard/canvas frameworks, scheduling,
  RUM/profiling, mandatory temporal storage, classifier exports and generalized
  query engines. Preserve existing offline compiler/annotation capability without
  claiming these larger systems are delivered.

The review recommends fewer repeated representations of the same scope and clearer
ownership of preserve/return actions. It does not recommend flattening distinct
signals, clocks, definitions, exact occurrences and fixed captures into one object.
