# One captured-record review task — readiness gate365

Status: bounded semantic readiness gate accepted after resolver/publication proofs.
Only the public ResolveRecord capability is production implementation; full review
UI, HTTP/storage families and export remain proposed. No runtime/frame acceptance
is claimed for that future workflow.

## One user task, inside the current inspector

An investigator opens one exact stored record occurrence, chooses a declared
categorical review task, records a judgment, reviews/corrects it and exports a pinned
set of reviewed decisions. Use a contextual **Review example** disclosure beside
existing annotation controls in the captured-record inspector. Keep descriptive
tags/notes separate; neither is silently reinterpreted as training truth. No new
workspace, population propagation, model suggestions or source reread is required.

Preview the target capture short ID, absolute stored occurrence, signal, service,
event UTC and original custody before Save. The server resolves the occurrence
from the immutable capture and binds its raw identity/digest; browser-provided
SpanId/TraceId/visual row number is not authority. Duplicate IDs remain separate
occurrences. The record fragment does not become a label on its complete trace.

## Minimal immutable contract to close with architecture

One immutable task defines a title and bounded categorical vocabulary with stable
literal keys/display titles. Editing vocabulary creates another task version/content
ID; existing decisions never acquire new meaning. A negative class must be an
explicit vocabulary choice such as `not-timeout`, never the implicit complement
of a predicate or all unselected examples. No generic ontology/schema editor.

Candidate wire separation (names subject to architecture closure):

- Task: immutable content ID, title, bounded `labels:[{key,title}]`.
- Target input: `{captureID,absoluteIndex}`; publication resolves server-owned raw
  occurrence identity/digest and original receipt, not client-supplied evidence.
- Decision revision: task ID, resolved target, optional parent revision,
  `judgment:{kind:'label',value:key}|{kind:'uncertain'}|{kind:'abstain'}`,
  `review:'unreviewed'|'reviewed'|'disputed'`,
  `state:'active'|'withdrawn'`, note and server recording time/content ID.

No decision is **unlabeled**. An explicit declared negative label is a categorical
judgment. **Uncertain** records insufficient confidence; **abstain** records a choice
not to judge. **Disputed** marks an unresolved review disagreement. **Withdrawn**
retires the assertion without deleting history. **Reviewed** means a recorded local
review decision, not authenticated reviewer identity, objective ground truth,
telemetry completeness or causality. Keep these separate axes in persisted data;
UI can present concise appropriate actions rather than a wall of three selectors.

Existing meaning revisions permit branches; do not silently choose the latest
head or infer consensus. Either enforce bounded optimistic-parent conflict
publication or require explicit pinned revision selection and report unresolved
heads. The implementation must close this before export; no last-write-wins label.

## Natural actions and export boundary

Save judgment creates a revision; Mark reviewed, Dispute, Correct and Withdraw
are explicit follow-ups against the current revision. Preserve draft text on
refusal; publication failure is not a saved label. Local metadata requests do not
read telemetry. Reopening the exact decision shows its task version, target and
recorded history without relabeling adjacent records or future matches.

Export an explicitly selected, versioned manifest of pinned task/decision/target
revisions. Include only active reviewed categorical labels by default; preview
counts and exclusions for uncertain/abstained/unreviewed/disputed/withdrawn and
conflicting targets. Unlabeled never becomes a negative. Export must bind exact
member identity and evidence provenance; no automatic feature projection, train/test
split or leakage guarantee is promised. No new store migration/XTDB requirement
is assumed. Architecture must close bounded file publication and target resolver
before implementation, or defer this task and advance366.

## Acceptance if the readiness gate closes

A real retained duplicate-ID record is labelled with one explicit vocabulary value,
reviewed, corrected/withdrawn and reopened with source unavailable. An adjacent
occurrence remains unlabeled. Exercise uncertain, abstain, disputed and conflict
refusal without coercion into labels. Export pinned reviewed decisions, independently
verify target raw identity and excluded-state counts, and prove no telemetry reads
or capture mutation. Actual desktop/narrow frames show contextual task/target,
concise judgment/review meaning and explicit publication/export state. Metadata
identity/revision checks, not screenshots, establish dataset authority.


## Closed selected-branch semantics and gate acceptance

Accepted root/architecture's bounded explicit selected-revision branch contract.
A manifest pins an immutable task, selected reviewed decision, bounded known
ancestry/observed heads and server-resolved capture occurrence/hash. It does not
assert global/latest truth, consensus or an atomic knowledge cut. Unknown concurrent
or unselected branches are not claimed absent. Disputed or multiple pinned heads
are refused until an explicit resolution references the known heads; neither a
clock nor display ordering chooses a winning label.

An existing manifest remains immutable after a later withdrawal. A newly requested
historical reviewed pin must be visibly described as **historical/pinned**, never
Current label or present ground truth. Reopen must show the selected decision's
actual task version/value/review state and provenance, with later knowledge separate
if shown. Export selection and exclusion language in the future UI must preserve
this boundary rather than suggesting all current heads were exhaustively scanned.

The private publication proof reloads selected stored task and decision authority
rather than trusting draft label text, rejects foreign joins and binds the exact
raw capture member/hash. Main Go/generated proof acceptance is reported by root;
this is semantic readiness, not a claim that a review product or manifest export
endpoint exists. Full contextual UI, production review storage/HTTP and actual
journeys/screenshots remain proposed. The gate is sufficiently bounded to continue
planned work without a generic metadata framework or new store dependency.
