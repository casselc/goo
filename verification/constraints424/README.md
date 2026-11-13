# Contextual field constraints (424)

The native Quint model separates local inspection/edit/removal from explicit
counted reads and Activity Apply. It checks counter conservation, original-scope
retirement, ordered Apply and a forbidden removal under an OR ancestor. Negative
controls inject acquisitions into local/stale actions and attempt unsafe ordering.

The actual completed browser report is replayed with every required continuous
fact. Its first source read establishes a relative baseline; all later effects
must be accounted for. An extra stale-action acquisition is rejected. This
source-budget replay complements the browser's actual DOM, ancestry, predicate,
raw-source and owner-lease assertions; it does not infer those facts from clicks.

Hegel drives the production JavaScript planner through generated namespace,
signal, environment-policy, Unicode/empty/literal values and duplicate-clause
cases. An independent finite-domain evaluator checks that removing one exact
conjunct preserves all original matches and every unchosen condition. Tests also
cover OR refusal, numeric binding preservation, changed question metadata,
forged path/content and sequential stale-intent rejection. Planning performs no
source read.

From `verification`, with Go/Node/Quint on PATH:

```sh
GOO_CONSTRAINTS424_REPORT=../artifacts/field-constraints424/report.json \
  go test -p 1 ./constraints424 -count=1 -v
```

Quint is mandatory and fails if missing. Without the report environment variable,
only the observed-ledger test skips; do supply it for acceptance.
