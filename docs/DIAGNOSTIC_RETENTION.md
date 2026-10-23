# Development diagnostic retention

Browser recordings are disposable debugging material. They are separate from
ClickHouse telemetry, saved captures/receipts, investigations, annotations and
XTDB records. This policy does not connect to or delete from those stores.
Committed documentation screenshots under `docs/journeys/screenshots` are outside
its scope too.

Run `npm run diagnostics:plan` for a dry run or `npm run diagnostics:prune` to
apply. The journey wrapper applies retention on exit, after restoring its traffic
generator; a cleanup failure does not replace the test exit status. Direct
`npx playwright` invocations require the explicit prune command afterward.

| Raw browser recording | Retention |
| --- | --- |
| Explicit successful outcome | 24 hours |
| Explicit failure/cancellation | 7 days |
| Legacy or unrecognized outcome | 72 hours |
| `.keep` or `.active` in the run directory or an ancestor | Protected |

The target is **512 MiB of unique allocated raw browser files**. To relieve
pressure, successful recordings can expire early after two quiet hours. Pins,
active runs and recent failed/unknown outcomes take precedence; the report says
`budget_exceeded` if they prevent reaching the target. This is a soft budget,
not a guarantee against disk exhaustion. Age uses the newest file in a bundle,
including logs/ledgers, so ongoing diagnostics protect their earlier frames.

Only these files within the repository's `artifacts` directory (or its resolved
symlink target) are eligible:

- PNG/JPEG screenshots and WebM browser videos.
- `trace.zip` browser attachments.
- `trace.jsonl` files whose first record matches Bombadil's browser-record schema.

Compact JSON ledgers, other model/action traces, validation results, logs and
unknown file types remain. Logs are intentionally excluded from this bounded
policy; it is not a total artifact-store budget. Deletion removes raw trace/frame
files together by bundle; historical references to expired raw files no longer
open. Compact evidence is not a substitute for an expired replay recording.

The tool defaults to dry run and has no arbitrary root option. It ignores
symlink files/directories, serializes its own invocations, rechecks bundle pins
and freshness, and skips files changed since planning. It estimates physical
reclamation using inode/link counts, rather than summing hardlinked copies.
Apply writes a gzip JSONL inventory of deleted paths, sizes, timestamps and
reasons under `artifacts/retention`. Empty directories and compact records remain.

Bombadil's shared test adapter creates `.active` before launching and atomically
publishes `run-result.json` on success, failure or cancellation. A crashed owner
may leave `.active`; inspect the run and remove that stale marker manually once
no producer is using it. Legacy results are not guessed from log text or folder
names. Playwright attachments without an outcome file use the legacy policy.
Pin an important failure before its retention deadline with `touch <run>/.keep`.

There is no background scheduler. Retention runs through the journey wrapper or
an explicit command; producer growth can exceed the target between executions.
