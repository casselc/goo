# Development runtime recovery before 209

The cloud restart cleared `/tmp`. The repository persisted at `9b0993c`, but
its temporary Go toolchain, dependency, cache and ignored-artifact targets were
missing. Prior ignored browser/model journals are unavailable in this session;
committed documentation/frames and named-volume captures remain. New 209 checks
use fresh artifacts, not claimed replays of those missing journals.

Restored under `/tmp`: official Go 1.27.2 (release archive SHA256
`ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5`),
lockfile-based Playwright 1.56.1, Quint 0.32.0 and Bombadil 0.7.9, dev/API Python
venvs, module/build caches and the artifact root. The physical Node dependency
path ends in `node_modules`, so ordinary sibling resolution works without the
208 isolated launcher. No dependency version or lockfile changed. Receipt:
`artifacts/runtime-restore-209/restoration.json`.

ClickHouse, Redis and XTDB had stopped during the restart. There was no unused
Docker image/build cache to reclaim and the overlay had only about 63 MiB free.
While ClickHouse remained stopped, two diagnostic files were independently
read-copied and verified by complete byte count and SHA256:

| File in /var/log/clickhouse-server | Bytes | SHA256 |
| --- | ---: | --- |
| clickhouse-server.log | 742,807,963 | `fc2e7ae0b07f15fd115ccadbc50f943b261287d4c009f5c8a7e4df4dfa6616cc` |
| clickhouse-server.err.log | 186,653,364 | `6c9f857affde346c1b3bc0cdde3ea98f3f3ee169f5a2ff807abce48e2ed796c9` |

Only those two stopped diagnostic files were truncated after verification,
preserving UID/GID 101/101 and mode 0640; zero lengths were verified before
starting the existing containers. Archives and the manifest remain under
`/tmp/goo-clickhouse-debug-archive-209`. No application/system tables, volumes,
captures, other logs or container configuration were deleted or recreated.
The observed overlay free space subsequently reached 2.2 GiB; that is an
observed filesystem result, not an allocated-byte guarantee from truncation.

Existing named-volume capture storage survived. The 208 individual CPU capture
remains 6,407 bytes, SHA256
`17c2d718aff5053f9ad9a5c801447d87ca203db2545e678b6afcb23c902a2df4`.
ClickHouse /ping, Redis PONG and collector readiness pass after starting the
existing dependency containers. Ordinary development traffic resumes without a
fixture/data reset.

## XTDB health probe correction

XTDB's actual GET `/healthz/alive` and `/healthz/ready` return HTTP 200, with
zero unhealthy databases. Its inherited Docker healthcheck uses HEAD via
`wget --spider`, and Jetty returns HTTP 405 for that method. The source override
in compose.xtdb.yaml now uses GET against the same alive endpoint. The exact
command succeeds in the existing container, and combined Compose configuration
validation passes. No running-container recreation or second restart was done;
its old Docker healthcheck definition remains until normal container recreation.
This is a probe-method mismatch, not evidence of unhealthy storage or a reason
to change application data.
