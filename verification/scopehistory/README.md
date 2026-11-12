# Scope-history state verification

`go test ./scopehistory -count=1 -v` from `verification/` runs 64 Hegel cases
against the actual `ui/scope-history.mjs` Journal through a Node bridge. Each
case forces pin capacity and expiry boundaries, then generates 100 commands.
An independent Go state oracle checks ordered identity, exact zero continuation
offset, pin admission, recent expiry, clearing, explicit persistence, reload and
reset. Poisoned intent fields and lost-pin/invented-persistence controls reject.

The bridge denies source acquisition and checks caller inputs remain unchanged.
This is local configuration-state verification, not UI lifetime, browser storage
security, authentication, or full product model verification. Browser journeys
cover actual staging and source-read authority separately.
