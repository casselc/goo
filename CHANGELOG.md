# Change log

## 2026-10-16 Codex handoff

Bundled the local lab with repository instructions, a Codex continuation task, architecture, proposed language and measurement RFCs, staged work, known issues, and primary-source references. Exported the earlier design as a portable Markdown snapshot. Rechecked unit/API/embedded ClickHouse tests, Compose configuration, and documentation links.

## 2026-10-16 Initial prototype

Added the Compose environment, instrumented HTTP/gRPC reservation flow, Redis Streams consumer, OTel ClickHouse exporter, query/explain API, JSON query controls, and initial UI. Added SQL checks in embedded ClickHouse and an application/OTLP integration check using a Redis test server. Included an optional HyperDX profile and a full-stack smoke script.

The prototype remains a fixed-query demonstration. General expression evaluation, complete SLO accounting, package publication, and persisted alarm evaluation are proposed follow-on work.
