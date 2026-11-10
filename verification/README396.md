# 396: held presentation ownership

Run the actual journeys serially from the repository root with Node 24, loopback
access and the reviewed Go binary. The main consumes the existing 395 ordinary
OTLP bank; run 395 first if that bank is absent. It does not reset lab stores.

```sh
GOO_GENERIC_BINARY=/tmp/goo-396-reviewed-goo GOO_HELD_DRAFT_396_MODE=final NO_PROXY=127.0.0.1,localhost npx playwright test journeys/396-held-draft.spec.js --workers=1
GOO_GENERIC_BINARY=/tmp/goo-396-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/396-held-draft-actor.spec.js --workers=1
```

The default main mode is `final`. An explicit `baseline` run against the older
395 binary characterizes defects; it does not assert product acceptance.
Preserve its report separately from the final report. The actor creates a fresh
campaign directory and records actual fetch request/response pairs and physical
proxy SQL, rather than relying on browser resource timing.

From the verification module, with Node 24, Go and Quint on PATH:

```sh
XDG_CACHE_HOME=/workspace/.cache go test . -run '^TestHegelHeldStagingWrappers396$' -count=1 -v
GOO_HELD_DRAFT_396_BASELINE=/workspace/goo/artifacts/held-draft396-baseline/summary.json \
  GOO_HELD_DRAFT_396_REPORT=/workspace/goo/artifacts/held-draft396-final/summary.json \
  GOO_HELD_DRAFT_396_ACTOR=/workspace/goo/artifacts/held-draft396-final/actor.json \
  go test . -run '^TestQuintHeldDraftObserved396$' -count=1 -v
```

Main actual acceptance passes in 17.5s (19.7s runner): 20 measured boundaries,
four source requests, one stored-definition read, three configuration writes,
and 19 physical SQL attempts. The final SQL attempt is an actual refused Count
request (503), surfaced as an unavailable API response (502). Old callbacks
cannot publish or change the current draft while it is pending or failed.

The immutable baseline demonstrates the trail-resume visible/saved mismatch and
invalid Find publication. The final result saves exactly the visible local
intent, blocks a 258-byte invalid Find draft, and recovers after correction.
Nested Back was already correct in the baseline. No screenshots or model checks
are used as a substitute for actual source evidence.

See [the storyboard](../journeys/396-held-draft-storyboard.md) for the user flow.

The 35-second Bombadil companion passes in 35.8s (37.3s runner), recording 536
DOM snapshots across 13 phases. Two successful source HTTP requests account for
12 physical SQL reads; three successful configuration publications match the
visible definitions. Completion includes invalid-draft refusal and correction.
The campaign's screenshots are retained as driver artifacts; no camera freshness
or pixel-based assertion is claimed.
