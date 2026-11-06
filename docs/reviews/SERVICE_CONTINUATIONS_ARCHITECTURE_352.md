# Typed service continuations — architecture review 352

Source and native closure accepted; generated and execution checks are recorded separately
in [validation](../../VALIDATION.md).

The change uses the existing record environment codec and record-question
validator. `widerQuestion` preserves the admitted discovery policy and fixed
bounds, while selecting the reported child's service and qualified namespace.
It intentionally omits `reportedEdge`: the population is wider than the selected
relationship. Returned predicates and scope fields are detached from their input.

Every typed selector, including current-only named, remains unsupported by the
existing exact-edge compiler. The renderer checks that capability before building
an exact link, and the direct exact-question helper also refuses it. Scalar exact
links retain their existing duration/status/operation profile. No scalar
projection or backend change is hidden in this UI milestone.

The existing inventory owner admits returned discovery environment, namespace
and fixed bounds before mounting distributed services. This change adds no new
response authority or source executor. Link construction, view switching and
configuration disclosure do not acquire telemetry.

The service browser passes only checked environment fields to the existing
Activity library. That library already checks every returned item's policy and
retires pending metadata on disposal. It does not receive the selected service,
namespace or fixed timestamps. A saved selector remains an independent rolling
configuration; its normal open route stages it and explicit Start reads it.
The scoped heading/note states that saved filters may target other services.

Native metric, deployment and annotation capabilities keep their existing
current-only checks. Their restrictions no longer suppress ordinary record
configuration discovery. No compatibility layer, new generic owner framework,
dependency inference or additional storage format is introduced.

Native layout inspection found the unbroken receipt basis causing document
overflow at 390px. The receipt now reuses the existing wrapping identifier style.
The relationship table uses the shared compact stacked-table component with
caption, native table sections, column scopes and cell labels. This also keeps
parent/child identities beside actions on narrow screens; no query semantics
changed. Final desktop/narrow frames and the measured 390px layout audit pass;
model/negative and focused actor outcomes are recorded in validation.
