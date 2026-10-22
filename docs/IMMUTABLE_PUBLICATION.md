# Immutable file publication — C1 package/workbench/capture adoption

Reusable telemetry packages (`exploration.PackageStore`) and workbench artifacts
(`workbench.Store`) share `internal/immutablefile.Publish`. Exploration captures
now use that path too; investigation captures use its exclusive `PublishNew` entry.
This is filesystem mechanics only. Callers provide trusted directory/name/prefix and already encoded
bytes; they retain validation, size limits, identity, canonicalization, lineage,
catalog rules and authoritative reloads. No stored schema or HTTP response changes.

Before extraction, `testdata/storage` froze one reusable OTel span package and one
deployment marker. Package revision hashes cover the definition bytes; deployment
identity hashes the typed envelope with an empty ID, including its independent
recorded and event times. The fixture tests passed against the old implementations
before the helper was introduced. The same tests run afterward and check exact
stored bytes/IDs, reopen, duplicate publication, refusal of truncated/foreign-field/
digest-mismatched objects, preservation of a corrupt winner and scratch cleanup.

```mermaid
sequenceDiagram
  participant Caller as Family store
  participant File as Immutable publisher
  participant Disk as Local filesystem
  Caller->>Caller: Validate, hash, encode, enforce size
  Caller->>File: Publish trusted path + bytes
  File->>Disk: Create private temporary sibling
  File->>Disk: Write, sync and close
  File->>Disk: Link to final name without replacement
  Disk-->>File: New winner or existing name
  alt Exclusive collision
    File->>Disk: Remove temporary sibling
    File-->>Caller: Original link error
  else New winner or accepted existing name
    File->>Disk: Sync containing directory
    File->>Disk: Remove temporary sibling
    File-->>Caller: created flag + error
  end
  Caller->>Caller: Apply family duplicate policy and reload/check identity
```

`created=false, err=nil` means a name already existed, not that it contained valid
or matching bytes. The idempotent consumers retain their previous duplicate policy:
keep the winner and let the family loader check it. Workbench's private publisher
alone does not validate the winner; its public create methods subsequently load it.
Investigation captures instead retain their original exclusive link error through
`PublishNew`; a collision returns the actual `*os.LinkError` without a directory
sync, as before. No repair or overwrite is implied. An error after a successful
link can coexist with a published file. Existing directory/file modes remain 0700/0600 for newly
created objects; existing permissions are not changed. Deferred temporary cleanup
and its ignored removal errors retain the previous behavior. Sync calls do not
establish power-loss durability; no crash/power-loss experiment is claimed.

## Store inventory and remaining C1 work

| Owner | Current contract | Adoption |
| --- | --- | --- |
| Workbench artifacts | Typed JSON, family hash/validation, 128 KiB limit, subdirectory catalogs, reload after create | Shared publication mechanics; deployment golden fixture |
| Reusable telemetry packages | Typed definition hash, 32 KiB limit, idempotent duplicate then verified reload | Shared publication mechanics; package golden fixture |
| Exploration captures | Embedded exact evidence, family digest/plan/parent checks, 24 MiB limit, duplicate then reload | Shared idempotent publication; exact capture/page fixtures |
| Investigation captures | Number-preserving embedded JSON digest, 24 MiB limit, parent question compatibility, duplicate link is an error | Shared exclusive publication; number-preserving capture fixture |
| Lifecycle first-observation custody | Identity excludes received time; first winner supplies custody time; no directory sync today | Existing writer retained; pin clock and race behavior before adoption |
| Native measurement YAML route | Validated typed definition identity, YAML bytes; existing winner accepted; no directory sync/reload today | Existing writer retained; pin route response and corruption behavior before adoption |

Existing store tests cover branching/parent mismatch and exact large embedded
integers; the initial two golden fixtures are not exhaustive identity coverage for all
workbench families or evidence stores. C1 remains active for lifecycle custody
and native YAML boundaries; their current absence of directory sync requires an explicit compatibility decision before
adoption. Broader per-family fixture coverage remains bounded, not exhaustive.

## Verification

The real filesystem race test launches 24 competing distinct complete payloads and
requires one complete winner, no replacement, private file mode and no temporary
leaks. Failure tests cover a non-directory storage path and a link failure after a
successful temporary write. A separate Hegel campaign generates 100 sequences of
2–12 writes with zero-to-4096-byte initial payloads, preserving the first bytes
through every duplicate and fresh read. These are filesystem checks rather than
new UI/model protocols; existing browser storyboards and Quint transition models
are unchanged. There is no new UI screenshot for this backend-only extraction.

## Capture continuation fixtures

The exploration fixture pins typed-envelope identity, exact raw evidence integers,
independent source-read/receipt timestamps and the complete first-page JSON
response. Duplicate publication reloads and rejects a corrupt existing winner
without replacing it. Its strict envelope decoder still refuses unknown fields,
truncation, digest mismatch and trailing JSON. Foreign receipt cursors and changed
refresh questions remain refused. Existing tests retain native, inventory,
comparison, source-plan and concurrent capture behavior.

The investigation fixture pins its distinct Number-preserving embedded JSON
normalization: changing key order preserves identity; changing a large integer
does not. Reopen preserves original evidence bytes, a valid refresh retains its
parent, and changed parent questions remain refused. This loader historically
ignores unknown fields; the fixture makes that policy explicit rather than changing
it during a filesystem refactor. Truncation, digest mismatch and trailing JSON
remain refused. The exclusive publisher test pins link-error type/existence
classification and preserves original bytes on collision.

Both fixture tests passed on the old writers before migration. They now accompany
the same affected race suite and generated duplicate-write sequences; Hegel also
requires exclusive duplicate attempts to return an existence error on every step.
No new UI or execution-state protocol is introduced.
