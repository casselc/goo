# Go review of the C0 correction — 244

The independent Go reviewer inspected the working implementation read-only, then
reviewed the enumeration follow-up. No blocking correctness finding remained.
Reviewer checks were source review; executed checks are recorded in VALIDATION.

The bounded summary fixes escaped accepted-input → index/rebuild mismatch without
changing authoritative receipts. The measurement enums match the current validator,
and canonical scope preserves lifecycle → logs. The reviewer requested coverage
for all six measurement kinds and inventory/explanation; a compact projection/
read-back table now exercises them, distinct from the public writer regressions.

The reviewer caught an upgrade edge: retaining v1 while creating v2 adds a physical
directory entry. The final shared reader reserves only the actual `.catalog-v1`
and `.catalog-v2` directories separately; all other physical entries count toward
10,000. Reading at most 10,003 entries detects overflow even with both reserved
directories present. Search, rebuild and legacy list use the same admission rule.
An at-capacity migration regression passes; adding arbitrary `.catalog-v3` refuses
all three readers. This is documented rather than silently changing one reader.

The whole-store 256 MiB rebuild limit and metadata-versus-payload integrity boundary
remain explicit. Targeted repair is a separate roadmap item. The review did not
claim production load, exhaustive filesystem race safety or full accessibility.
