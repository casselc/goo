# Development diagnostic retention

The Compose lab mounts `config/clickhouse/development-logging.xml`. This is a
local development policy, not a production telemetry retention policy:

- ClickHouse file logging is warning-level, with 10 MB rotation and three files.
- High-volume internal text/trace, processor, query-view/query-metric, latency
  and periodic metric diagnostic writers are disabled.
- Internal query, part and error logs retain one day (TTL cleanup is asynchronous).
- Application OTLP logs, traces and metrics, and Goo captures are unaffected.

On 2026-10-27, cleanup cleared only eleven explicitly named `system` diagnostic
log tables. Available overlay space rose from 39 MB to 3.9 GB. Six application
table row counts were preserved or increased across the cleanup/restart.
ClickHouse ping, Goo capabilities and UI configuration returned HTTP 200;
Compose validation passed. The three retained log tables report one-day TTLs.
Before/after metadata receipts are in `/tmp/goo-diagnostic-cleanup-210`; removed
diagnostic histories are not retained as another large archive.

The existing ClickHouse container received the same config file and was
restarted without recreation. Its first startup rejected unreadable config;
permissions were corrected to 0644 and successful startup was verified. The
Compose mount makes the policy survive ordinary container recreation. Old
buffered diagnostic rows flushed during shutdown can remain; disabling their
writers does not remove historical tables. No application tables or Docker
volumes were pruned.
