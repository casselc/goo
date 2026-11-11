# Literal source-parameter transport

[Shared parameter encoding](CLICKHOUSE_PARAMETER_ENCODING.md) (405) now consolidates
the validated source writers. This audit retains its original source/byte contract;
405 validation records fresh round trips and the old/shared decoder comparison.

The 393 audit found matching literal selection through the existing record and
metric source encoders. They remain unchanged. A separate validation defect was
fixed: native `Source.Run` and `Evaluate` now reject malformed typed input before
JSON cloning can replace invalid UTF-8 with a valid replacement character.

The actual ClickHouse audit passed 150 public-source arms, twelve changed-literal/decode-intent
selections, and 147 distinct server-decoded `String` echoes in 4.10s
(package 4.123s), with zero skips. This establishes the tested round trips; it is
not a performance comparison or proof that arbitrary strings are valid inputs.
The combined post-fix run also passed generated intent and typed-refusal checks
with zero skips.

## What was tested

Ordinary OTLP logs and native gauges carry the same `literal` attribute values.
The finite bank contains 50 targets, 50 distinct near misses and one missing-key
control in each source: 202 physical rows altogether. Eight bounded, read-only
queries settled the bank and independently checked original UTF-8 against
ClickHouse `hex(...)`, key presence, IDs, timestamps, resource identity and metric
scope/schema/value fields before any public source audit.

Values include every ASCII control from `01` through `1F`, DEL, backslashes,
trailing backslashes, escaped-looking `\n`, `\x41` and `\u0041`, empty text,
spaces, punctuation, NFC/NFD text, BOM, emoji and Unicode line separators. A true
newline and the two-character spelling `\n` remain different values. Target and
near-miss identifiers are observed in returned evidence, not added as direct
literal-filter predicates.

Each target is read through record `Source.Fetch`, metric `Source.Run` with
direct `Binding.Attributes`, and the native exact-stream path. Exact-stream JSON
is tested separately because its own JSON escapes can hide differences at the
outer Escaped-text boundary. Actual typed HTTP parameters, server-decoded values
and returned membership are checked rather than comparing encoded spellings.
The controls change valid query values, including backslash-n to LF and
double-backslash-n to backslash-n. They return distinct known IDs, which the
original target-ID oracle rejects.

## Public validation still applies

The shared bank uses valid UTF-8 without NUL within the 512-byte intersection.
The public value limits differ:

| Value field | Maximum UTF-8 bytes |
| --- | ---: |
| Record `Predicate.Value` | 1024 |
| Metric `Binding.Attributes` value | 512 |
| Exact-stream attribute/resource value | 2048 |

Existing map, key and serialized exact-stream limits also apply. In particular,
513 bytes is refused by direct metric attributes; that size alone is legal for
a record predicate or exact-stream value. NUL and invalid UTF-8 are refusal cases,
not legal OTLP bank entries. Typed Go tests construct invalid UTF-8 directly;
JSON transport is not used to pretend those bytes survived unchanged.

Readonly settings, source limits and family-specific failures remain unchanged.
No application UI changed, so the existing journeys and screenshots remain
current; no additional screenshot, actor or model claim is attached to this
transport audit.

See [reproduction and verification commands](../verification/README393.md) and
the [ordinary OTLP bank tool](../journeys/support/literal-transport393-bank.js).
