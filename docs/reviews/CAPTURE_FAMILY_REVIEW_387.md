# Capture family boundary review — milestone 387

Read-only architecture/usability source review, November 9, 2026. No heavy tests,
runtime acquisition or production edits were performed. Implementation acceptance
is pending.

## Concrete gaps

`ui/saved-capture-protocol.mjs` returns early for captures without typed environment
selectors. That bypasses general family admission: missing/unknown kinds, missing
active results and mixed non-trace-search payloads may pass. Its measurement branch
checks question identity but returns without excluding sibling result arms. Basic
receipt fields and family-specific page metadata also lack one uniform boundary.

Go `validateCapture` already has stronger typed exclusivity and family validators.
However, ordinary typed JSON decoding loses the distinction between omitted and
inactive null arms and accepts case-insensitive field aliases. A finite wire guard
must run before those distinctions disappear. Do not replace the existing family
semantic validators with shallow envelope checks.

## Smallest coherent contract

Use an exact six-family classifier: explore, inventory, comparison, explanation,
measurement and trace-search. Shared admission validates receipt metadata, exactly
one matching nonnull question arm and exactly one matching nonnull result arm;
unknown keys and any inactive arm presence, including null, are refused. Dispatch
to the existing family semantic validator after admission. Apply the check
regardless of environment selector representation. A finite table/switch is enough;
no extensible plugin registry or generic schema framework is needed.

Keep three boundaries explicit: metadata receipt/reference, full stored artifact,
and API page envelope. A document reference or catalog receipt is metadata and
must not need embedded evidence. A full artifact has no paging wrapper; a page
must carry its receipt, one result and explicit family-specific retained/paging
fields. Known optional null values *inside* an admitted family retain their base
semantics; this work must not blanket-ban null maps or selectors.

## Preserved semantics

Only Explore uses local stored-record pagination. Its immutable receipt question
is normalized, while the displayed result question changes offset/limit in 100-row
pages. Count/groups/basis describe the captured result, not the page. The cursor
binds local offset to immutable receipt ID, not authorization. Retained records
are not total source count; `recordsLimited` denotes original source truncation.

Measurement reports retained rows without cursors. Trace search reports retained
items and original `hasMore` as `recordsLimited`, also without stored record
cursors. Inventory, comparison and explanation currently expose zero retained
record count and false `recordsLimited`; do not invent different counts while
consolidating admission.

Custody `capturedAt`, source `readAt` and frozen evaluation time remain separate.
Validate their own representations without inventing cross-clock ordering or
aging stored evidence against wall clock. Refresh keeps the original normalized
question and parent relationship. Explanation saves a held observed result and
refuses automatic source rerun; historical plans and other refresh paths retain
their existing family-specific semantics.

## Acceptance focus

Exercise all six positive families and cross-family, unknown, missing, null and
case-alias mutations at both Go/JavaScript boundaries. Include ordinary historical
(non-typed-environment) captures, receipt-only document references, exact Explore
page identity and no-cursor families. Reopening/refused refresh must not silently
rerun, weaken scope or mutate immutable stored bytes. Actual UI refusal should
clear incompatible evidence and leave a concise useful error rather than rendering
the wrong family.


## Implemented source review

The candidate now has the finite boundary in `exploration/capture_envelope.go`
and `ui/saved-capture-protocol.mjs`. Root-only RawMessage envelope processing
preserves nested family payload representation and avoids constructing a second
recursive graph of large evidence. Exact-case keys, unsupported kinds, missing,
null and mixed active arms are rejected before dispatch. Go retains its existing
family semantic validators and historical-plan treatment rather than adding a
universal current-compiler admission gate.

The browser no longer conditions general admission on typed environment selectors.
Receipt-only document references have their own validator and cannot masquerade
as full evidence pages. Explore page checks bind canonical cursors, normalized
receipt question, displayed offset/limit, retained count and HasMore together.
Non-page families refuse cursor presence; null stored empty arrays remain accepted
where the base family permits them. Capture custody and source clocks are not
collapsed or ordered against one another.

One final source gap remains: JavaScript's comparison page branch compares both
record questions but omits equality of the receipt's `distribution` intent and
the result request's `distribution`. Go checks it. Align the effective boolean
intent and add a small regression so distribution and ordinary comparisons cannot
be exchanged under an otherwise identical receipt. This requires no new semantic
comparison engine.

The parent reports full root Go and 1,382 frontend checks plus 28 focused
null/empty-projection checks passed. This reviewer inspected source only and did
not rerun those checks. Generated API checking and actual journey acceptance are
pending; the comparison intent gap above must close before source acceptance.


## Final coherence follow-up

The JavaScript comparison intent gap is closed: omitted/null distribution intent
retains Go's false default, true must match, and wrong nonnull types are refused.
The parent reports 23 focused checks passed, generated API checking and
new 387 properties passed; an older 383 fixture identity correction was pending.
Those results are attributed, not rerun by this reviewer.

One concrete delivery-bound gap remains in Go `validateCapturePage`: trace-search
retained records and item length can agree above the receipt request's limit,
since only the generic 1,000-record ceiling is applied. The admitted request
already restricts its limit to 1..100. Require trace-search item/retained count not
to exceed that exact request limit, matching the schema, JavaScript and existing
trace-search semantic validator. This is page-envelope coherence and needs no
expansion of individual witness semantics. Actual runtime acceptance remains
pending.


## Runtime-discovered page ownership follow-up

The trace-search delivery bound is now enforced against the receipt request limit,
closing the final Go page coherence gap. Subsequent actual browser work found
another useful distinction: a failed requested page visit must not overwrite the
identity of the successfully retained page still painted below it.

The final `ui/captures.mjs` now stores successful page identity in `heldStoredPage`,
separately from the current request/visit state. That identity is bound to the
exact retained page object and its immutable ID/offset; it is not itself an active
URL or action capability. Existing active-route/visibility/current-owner guards
remain on actions. Showing a non-Explore capture clears the held record-page
location, as does opening kept evidence. `SavedProtocol.validateVisit` validates
the requested capture ID and cursor against the returned page, preventing a valid
but wrong requested page from being painted. Null/empty passive projections retain
their permitted no-data meaning rather than requiring fabricated rows.

No authority regression or remaining source blocker was identified in this bounded
review. The parent reports the actual main journey passed all six families, empty
cases, 30 immutable files, 18 malformed cases, wrong requested-page refusal/retry,
document references and retained child return. Full UI passed 1,386 checks, with
additional focused checks. This reviewer inspected source but did not rerun the
journey. The actor was finishing and final ledger replay remained pending at this
checkpoint; neither is asserted passed here.


## Final acceptance report

The parent subsequently reports the completed browser journey passed in 1.3
minutes with 39 observed facts. The independently driven actor produced 663
snapshots across six phases; the parent visually accepted its retained parent,
null witness and focus outcome. Its measured boundary was one source read, zero
store reads and zero writes. Wire checking covered 92 genuine responses and 11
requests; 18 malformed responses were schema-invalid, while the deliberately
wrong requested page was schema-valid but correctly refused by contextual visit
validation.

The final required-model replay passed in 7.91 seconds with zero skips, covering
39 facts/steps. Thirteen semantic mutants and the reference-read negative model
control were rejected. These runtime, image and model outcomes are parent-reported;
this reviewer performed source inspection and documentation updates, not another
run. No acceptance item remains pending in this review's bounded capture-family
scope.
