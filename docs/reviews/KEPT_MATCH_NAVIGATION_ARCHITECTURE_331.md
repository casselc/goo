# Kept matching-member navigation architecture review — 331

Status: ownership contract reviewed before implementation; final source review and
twenty generated actual-owner scenarios pass. Running-app evidence remains
separate in VALIDATION.md.

## Authority and minimum change

The content witness is a receipt-scoped navigation hint. Its memberIndex addresses
an element of the immutable receipt's members array, not the original page index,
record timestamp, trace/span identity, excerpt or a matching field value. Opening
must first use the existing exact-ID transport and strict complete-receipt
validation. Only the verified original member can supply inspector content. An
invalid hint must not select a different record or invalidate an otherwise valid
receipt; ordinary receipt opening remains available.

Catalog owns the explicit **Open matching record** interaction and supplies the
exact receipt ID and member position. It retains ordinary **Open kept observations**.
The hint is ephemeral in-tab intent; copied exact-ID links remain receipt links.
Captures owns route acquisition and hint lifetime. KeptSaved owns rendering and
selection through its existing member inspector. Display components remain
passive. No content re-search, source acquisition, custody write, global cache,
compatibility adapter or second inspector is justified.

## One route visit, one verified read

A kept route visit owns ID, acquisition promise, cancellation and completion
admission. Hashchange, popstate and an explicit matching-member activation converge
on the same visit. Repeated events for its ID join the pending read or reuse the
already rendered receipt. Passive duplicate events must neither issue a GET nor
rebuild the inspector. An explicit matching-member intent may select its requested
occurrence in that existing receipt.

A different kept ID, route departure or another evidence owner retires the visit
before aborting its request. Both success and failure handlers check ownership
before adopting state; cancellation alone does not exclude already-completed or
abort-insensitive responses. Hints retire with their visit. A revisit after route
departure acquires a fresh receipt, even for an earlier ID. A failed acquisition remains terminal for passive notifications of that visit,
while explicit activation may retry; it must not become a permanent rejected-promise
cache. Late completion of an old same-ID visit cannot acquire a newer same-ID
visit's hint or overwrite its receipt.

## Retained continuation

The existing kept → staged trace/surrounding-log → return path retains the original
receipt and inspector DOM. Returning to this retained evidence restores route,
focus and scroll without a receipt GET, inspector rebuild or source read. This
retention is scoped to the existing continuation owner; it does not authorize
cross-visit caching. Existing active/visible checks and continuation leases remain
in force.

## Verification requirements

Meaningful acceptance must observe GET counts and inspector identity, rather than
only checking the final URL: duplicate popstate/hashchange; pending and loaded
joins; distinct-ID supersession; late success and late failure; failure then retry;
route departure followed by a fresh same-ID visit; invalid hint and exact occurrence
selection where duplicate record values exist. Content navigation must expose the
full verified original record and its provenance, with no source reads. The
retained continuation return must preserve the same inspector node and focus.
Generated state traces should compare actual acquisition-owner behavior with an
independent visit oracle and include deliberate stale-adoption/read-duplication
negative controls where applicable. Running-app journey evidence is separate
from a pure state property and cannot be inferred from it.

## Observed verification

The Hegel property `TestHegelKeptMatchNavigation331` invokes the actual captures
module against deferred abort-insensitive custody transport and passive display
adapters. Twenty generated scenarios vary receipt IDs, one to eight repeated
visit cycles and one to twelve duplicate notifications. Each cycle independently
expects five acquisitions and two receipt adoptions. It checks pending/loaded
joins without rebuild, superseded hint retirement, distinct-ID late success,
same-ID fresh visit, failed-visit passive deduplication, an ordinary exact-link
click retry, exact member selection, false-inspect refusal and late failure after
a same-workspace route departure. The final focused run passes in 3.708 seconds.

Two temporary source mutations establish nonvacuity without modifying product
files: bypassing same-ID joining fails with `pending duplicate GET` (4 versus 1);
removing success admission fails with `late success adopted` (1 versus 0).
This property does not validate strict receipt transport, real DOM focus, original
record rendering or retained continuation identity; their focused frontend and
running-browser checks supply those separate claims. The initial property command
omitted XDG_CACHE_HOME and failed on a read-only Hegel cache; the corrected run
used `/workspace/.cache` and passed.
