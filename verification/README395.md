# 395: reusable grouped count timelines

The pinned `grouped-count-v1` profile carries count-only group×interval data
through the existing timeline/records composition and immutable view store.
Typed missing, empty and Unicode categories remain distinct. Source intervals
are anchored to the exact query start, including partial final intervals.

Run the actual jobs serially against the existing collector/ClickHouse and
reviewed Go binary, with Node 24 and browser tools. From the repository root:

```sh
GOO_GENERIC_BINARY=/tmp/goo-395-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/395-grouped-timeline.spec.js --workers=1
GOO_GENERIC_BINARY=/tmp/goo-395-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/395-grouped-timeline-actor.spec.js --workers=1
```

The companion consumes the main run's fresh ordinary OTLP bank, creates a unique
campaign directory and writes a separate `actor.json`. Run with loopback access
in the cloud outer execution environment.

From the `verification` module, with Node 24 and Quint on `PATH`:

```sh
XDG_CACHE_HOME=/workspace/.cache go test . -run '^(TestHegelGroupedTimelineContract395|TestGroupedTimelineWireParity395)$' -count=1 -v
GOO_GROUPED_TIMELINE_395_REPORT=/workspace/goo/artifacts/grouped-timeline395/summary.json \
  GOO_GROUPED_TIMELINE_395_ACTOR=/workspace/goo/artifacts/grouped-timeline395/actor.json \
  go test . -run '^TestQuintGroupedTimelineObserved395$' -count=1 -v
```

Actual acceptance:

- Thirty-eight raw records independently admit six source/child/destination cell
  matrices: sixteen cells for nine source records and twenty for ten destination
  records, per signal. The query anchor is 137ms after a UTC second; final bins
  are 500ms and 700ms. Known zeros are derived only for observed categories.
- Main passes in 19.9s (runner 22.1s): 27 measured boundaries, six source requests,
  four configuration-object reads, two browser publications and 36 SQL statements.
  Two destination definition publications are explicit setup. Staging, local
  kinds/Find/order and held returns add no source reads. Saved bytes stay equal.
- A current known-zero missing-key cell stages its exact partial window and typed
  predicate in a nested child; Back restores the same accepted parent. A retired
  handler cannot alter it. Filled source heatmap text measures at least 5.11:1.
- The actor passes in 36.5s (runner 37.9s), observing 568 snapshots across nine
  phases, two actual HTTP pairs and twelve SQL statements with no prelude.
  Its source clock, basis, cells and Find text survive Back.
- Required independent admission and Quint replay pass in 4.840s, zero skips.
  Eleven main semantic mutants, one staging-read model control and four actor
  mutants are rejected. Generated shared Go/JavaScript projection, reuse/store
  and strict wire checks cover coherent and changed states and refusals.

Unknown counts remain unavailable; absent intervals become zero only under the
checked exhaustive grouped source contract. All-zero invented categories remain
refused. Eight groups, 1,000 observed cells and 240 intervals bound the profile
(up to 1,920 derived cells). Heatmap stops at 32 intervals without truncation.
Complete captured cells do not imply complete ingestion or transactional reads.
No resource-timing or actor pixel evidence is claimed.

Four genuine frames and the full flow are in the
[storyboard](../journeys/395-grouped-timeline-storyboard.md).
