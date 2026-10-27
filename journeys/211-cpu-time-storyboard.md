# Journey 211 — CPU activity from cumulative CPU seconds

Status: source/core/API/property gates pass; live acceptance pending.
[Contract and diagram](../docs/CPU_TIME_RATES.md).

1. Find hosts from ordinary OTLP without service labels. Open a host reporting
   memory plus cumulative CPU time, with multiple CPU modes/cores.
2. Read memory and one deterministic CPU activity default within two native
   reads. Preserve unknown intervals through decreases, epoch changes, missing
   values and gaps; show CPU-seconds/s rather than a saturation percentage.
3. Inspect the original native question and explicitly Run. Select an interval,
   read UTC/exact delta/rational rate and stage logs/spans without another read.
4. Run each record question. Verify exact lower/upper boundary membership and
   host/environment scope across different SDK maps. Retained Back preserves
   the measurement, interval and opener; a trace expansion remains broader.
5. Capture native evidence using the existing fresh-read owner and reopen with
   source access blocked. Verify the complete stored bytes remain unchanged.
6. Hold real replies and withdraw the owner; exercise the new path with a bounded
   Bombadil actor. Reuse independent raw arithmetic and observed Quint replay;
   Hegel state schedules/parity cover arithmetic and interval mapping.
7. Capture and review actual desktop/mobile frames. Prepared OTLP proves the
   collector/ClickHouse/app path, not a running host scraper or source authenticity.
