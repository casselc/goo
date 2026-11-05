# Qualified surrounding-context planning — architecture 329

Status: final production source review accepted; generated execution recorded below.

Discovery.surroundingQuestion is the single pure planner for both surrounding-log
entry points. It admits a nonempty reported ServiceName and an own, plain
ResourceAttributes map, then distinguishes an absent namespace key from a present
empty or named string. Invalid reporting refuses before source acquisition. The
qualified predicate uses existing ServiceIdentity and predicate admission; no
service identity is inferred from a different field. KeptSaved now delegates
without overlaying or revalidating a second predicate. Its receipt/member and
log-only guards remain at their existing boundary.

The planner keeps the 1–300 second radius, default thirty seconds, original-window
clipping and anchor millisecond's exclusive upper bound (+1 millisecond). Typed
environment selection is copied by RecordEnvironment.questionFields; original
record predicates and aggregate declarations are omitted. Offset is admitted as a
safe integer in the existing Go query range 0–1000; limit remains 100. The older
RecordResults scope label now states exact service and namespace, and clears old
scope on refusal. Its namespace label uses the same own-key decision as admission.

Acquisition owners are unchanged: the older context action reads through its
existing readContext generation/abort owner, while kept continuation stages a
view for explicit Run and keeps the app-owned retained return. No shared source,
navigation, custody or cancellation manager is introduced. Trace-ID expansion
intentionally retains its broader cross-service scope. No endpoint, receipt
format, compatibility mode, cache or product module is added.

## Verification boundary

The existing 328 Hegel property now compares the kept query with the direct
Discovery query from each actual Go-valid/digested tail, page and capture member.
It independently derives clipping for 1s, 30s and 300s radii, checks paging,
environment identity, missing/empty/named namespace isolation and detached output,
and refuses invalid service/maps/namespace/clock/radius/offset inputs. Original
328 assertions and the unchanged 327 trace assertions remain. The shared fixture
is data-only and unchanged; there is no second test framework or expanded fixture
matrix.

Every returned variant is admitted by actual Go ExplorationView.Validate and
exploration.Compile. The compiler validates Query, binds values as parameters and
uses mapContains for resource-key presence; the property checks that compiled SQL
retains presence and service equality. Compilation does not execute SQL or prove
source membership. Independent JavaScript membership checks establish the planned
namespace distinction; native source/IO/focus/cancellation evidence belongs to the
maintained browser journey.

## Next general-purpose gap

The current kept flow can inspect one observation and open fresh trace or temporal
context, but loaded Find/Keep remains local to one Activity buffer. A useful next
question is an explicit search over already kept observations across receipts,
with honest local/catalog scope and source-free inspection. Define that task and
its evidence boundary before adding search indexing or another navigation owner.

Final generated command passes in 12.719 seconds, zero skips:
`go test -C verification -run '^(TestHegelKeptLogContext328|TestHegelKeptTraceContinuation327)$' -count=1 .`.
Thirty generated rounds per property admit 720 surrounding views and 180 trace
views (900 total). Surrounding variants also compile separately at offset 100;
view admission remains page zero. The first execution exposed that distinction
in the test fixture and was corrected before the passing run. Initial cache
population fetched pinned dependencies through the managed proxy after the
sandbox's socket denial; the final run needed no network escalation.
