# Revision and hash consumers — 228

Implemented: `measurements.Definition.Revision() (string, error)` validates an
authored definition before returning its JSON/SHA-256 revision. Invalid definitions
return an empty identifier and an error. Compile, stored-result replay, native
capture scope and package catalog/plan/publication/winner verification use it.
YAML encoding errors are checked before immutable publication.

Valid definitions hash the same bytes as before. There is no normalization,
default insertion, store migration, source read or new receipt. Declared names,
namespace selectors and nil/empty map representations keep their existing JSON
encoding. Frozen stored IDs and built-in definition IDs remain unchanged.

## Distinct remaining families

| Family | Consumers | Boundary and follow-up |
| --- | --- | --- |
| Candidate definitions and authoring seeds | `measurements/go_runtime.go`, `pod_dashboard.go`, `exploration/dashboard.go` | Some hashes intentionally precede required names/titles. Do not replace seeds with validated revisions. Separate seed construction from final candidate references before migrating. |
| Fixed questions | Host, pool and pod dashboard helpers | Exact window, selectors and definition form the key. Define a checked question identity separately; preserve nanosecond bounds. |
| Resources, instruments and capabilities | Host discovery, discovery, pod metrics, instrument capabilities | These identify source metadata, including unsupported instruments. Definition validation would reject useful discovery records. |
| Replay result basis | `measurements/source.go`, `replay.go` | Hashes question, fixed evaluation time and rows. Replay intentionally accepts historical raw-row and canonical-row bases. No algorithm change here. |
| Lifecycle custody | `measurements/lifecycle.go` | Business payload identity omits receipt time; immutable first-observed custody remains separate. |
| Test snapshots | Nested verification helpers | Arbitrary structures used to detect mutation are not authored definitions. |

`measurements.Hash(any)` remains a documented legacy unchecked JSON digest for
these families. It still ignores marshal errors; this delivery does not claim a
universal checked-hash API. Do not remove or rename it until actual callers and
stored identities have an explicit migration or preservation contract. Other
packages' digests have separate canonical contracts as well.

## Validation scope

The revision tests compare all built-in definitions with the prior digest and a
frozen stored revision, check nonmutation and selector/map representations, and
reject invalid definitions without issuing an identifier. Existing HTTP storage,
replay and scope tests exercise migrated consumers. No UI behavior or payload
shape changes; screenshots and frontend campaigns are unaffected.
