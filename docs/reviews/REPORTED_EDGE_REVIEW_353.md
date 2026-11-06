# Exact typed reported-edge review — 353

Status: final source and actual-frame review accepted. No production edits or browser execution by this reviewer.

## Grain and natural action

Investigate reported edge selects captured **child span records** whose reported
parent ID resolves to one selected parent service owner in the discovery scope.
It is narrower than Open child service spans (wider population). Neither action
means unique requests, deduplicated calls, end-to-end latency, async causality or
a complete trace. After 353 supports typed exact edges, replace the local 352
unavailable explanation with the existing exact action only where admission succeeds.
Stage first and require the existing explicit Run; Back retains discovery evidence.

## Ownership must precede child filtering

The current scalar SQL groups parents by TraceId/SpanId and admits a parent only
when the namespace-presence/value/service owner set is exactly the selected owner.
Typed parity must apply the same complete environment policy and fixed half-open
window to both child membership and parent ownership. DiscoveryNamespace, when
explicit, constrains that same scope and retains missing/empty/literal semantics.

Do not apply child service, operation, status or arbitrary `where` filters before
resolving the parent set. Doing so can hide a conflicting parent owner and invent
an edge. A valid parent can be outside the chosen child predicate and still own the
relationship. Parent ownership is a source-scoped reported identity, not a join on
span ID alone: TraceId is required too.

Repeated records with the same trace/span ID and the same qualified owner do not
create owner ambiguity; they also do not disappear from the child-record count.
Conflicting owners, including identical service text in different namespace
presence/value identities, are ambiguous and excluded from exact edges. Missing
parents remain unresolved; invalid/all-zero IDs remain invalid; same qualified
owner remains internal. Do not select an arbitrary first parent or infer an external
node. Child and parent current/legacy conflicts must each obey the chosen policy.

## Admission and custody footguns

Before this cut, selected-row admission checks typed environment, signal and time,
not child service/namespace or parent relationship. Extending query admission alone
would therefore leave malformed source/capture rows falsely looking like exact
edges. Add checked raw child membership (qualified service/namespace, canonical
valid TraceId/ParentSpanId and fixed time/policy) at the relevant source/snapshot/
capture boundaries, including overflow rows before presentation trimming.

A returned question echo is not proof that a row has the requested relationship.
The relationship authority is the checked source query execution and its admitted
result; raw child fields alone cannot prove that the parent had no other owner.
If a parent-ownership witness is added, close and validate its exact scope/identity
rather than treating an arbitrary echoed owner as proof. If no witness is stored,
state that limit honestly: reopen validates stored child membership and receipt
integrity, but cannot independently rerun parent uniqueness from child data alone.
Do not add a hidden parent lookup on reopening immutable evidence.

Count, groups, rows and field discovery remain separate reads. Late telemetry can
change ownership between reads; no transactional-snapshot claim follows. Preserve
existing source bounds and fail explicitly on exceeded budgets; no truncation,
cap increase or count derived from the visible page is justified. Parent lookup
must remain a bounded source query under the existing cancellation contract.

## Exact versus wider continuation and reuse

Exact child continuation retains the ReportedEdge definition, original policy,
qualified identities and fixed window through rows, capture and supported package
planning. Removing ReportedEdge is a wider population, not a same-population transform.
A trace/log context follow-up uses its established exact IDs and original bounds
but is not falsely labelled as the same edge population. Captures reopen their
stored observation and original read clock; an explicit reread may legitimately
resolve different parents later.

## Acceptance controls

Generated/independent cases should cover all four typed forms, current/legacy
conflicts, missing/empty namespace contrasts, duplicate same-owner records,
conflicting owners, same SpanId in different traces, parent outside child filter,
parents outside the window/policy and wrong child membership. Include exact-empty
service identity and canonical/invalid ID contrasts. Mutated query echoes, raw
wrong-service/namespace rows and missing/null proof fields must not pass capture
admission simply because their reported query matches.

Native acceptance follows typed discovery → exact edge → staged question → Run →
known edge-only rows → capture/reopen → Back, with a wider-child contrast that
contains additional records. Check original identities/policy/window, independent
receipts, zero-read staging/Back/reopen and old-handler/late-response refusal.
Actual desktop/narrow frames should put exact versus wider actions beside their
qualified identities, and describe child-span grain without more verbose controls.
Final source and actual-frame reviews are recorded below.

## Independent final-source acceptance

Reviewed the candidate compiler, model, reported-edge source, selected-row admission
and three UI changes. The shared checked environment expression now scopes both
child membership and parent resolution. Parent scope contains only time, environment
and optional discovery namespace before grouping; selected parent owner and child
predicates do not hide conflicting owners before uniqueness is established.

Selected-row validation now applies to scalar as well as typed edge questions.
It checks well-formed raw JSON before detachment, explicit child service including
empty identity, exact namespace presence/value, canonical nonzero trace/parent IDs,
and the existing environment/signal/half-open time proof. Existing source/snapshot/
capture calls share that admission, including raw overflow rows. The code explicitly
states parent uniqueness comes from the checked source plan; it is not reconstructed
from a bounded child page or refreshed on capture reopen.

UI exact links now retain typed fields after backend support, and counted-value
queries keep ReportedEdge rather than widening it. The wider child action remains
distinct. Independent source review accepts the bounded implementation. Native
exact/wider membership, custody and actual-frame results are recorded separately;
this does not establish transactional consistency or independent offline
parent-uniqueness proof from stored child rows.

The final local ReportedEdge decoder also checks raw identity JSON before Unicode
repair and rejects unknown nested fields. Valid U+FFFD and paired supplementary
Unicode remain supported. Full Go/UI/API and generated real-SQL oracle checks are
recorded separately in validation; this reviewer does not claim browser execution.

The final request identity boundary also uses a local ReportedEdge JSON decoder
that rejects malformed raw Unicode before strict nested decoding. Unknown identity
fields are refused, while legitimate replacement/supplementary characters remain
valid. This closes request-side repair into a different owner without introducing
a new global decoder policy.

## Actual desktop/narrow acceptance

Reviewed all four current frames at original detail:
`353-exact-edge-1280.png`, `353-exact-edge-390.png`,
`353-exact-records.png` and `353-saved-exact-390.png`. Desktop and narrow keep
qualified parent/child identities next to separate Investigate reported edge and
wider-child actions. Empty namespace is explicit; narrow cards retain labels and
counts without clipped identities. Unresolved/ambiguous diagnostics remain visible
rather than being represented as selected edges.

The applied effective-policy frame shows three captured child spans, original
UTC bounds, the qualified reported-parent relationship and a clearly wider-child
alternative. It does not present unique calls or a complete trace. The saved
narrow frame identifies frozen evidence, the same effective policy/fixed window
and three source records; independent capture meaning and unknown coverage are
explicit. All four frames pass visual acceptance.

Native exact/wider contrasts across four policy cases, scalar equivalence,
YAML/plan preservation, source budgets and capture/reopen establish membership and
custody independently. The pictures alone do not prove parent uniqueness, no new
reads or immutable storage integrity. Independent source and visual acceptance
are complete; operational actor/model results retain their separate validation.
