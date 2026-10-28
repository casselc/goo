# Portable aggregate compiler pilot

This repository now includes a small Go implementation alongside the unchanged
Python lab. Go was chosen for this slice because the toolchain is installed and
the sibling query service already uses Go. It does not rewrite the lab or settle
every future runtime choice. YAML+CEL and KCL remain authoring candidates; this
pilot implements a deliberately bounded YAML representation without either
expression runtime.

## Go library boundary

Import `github.com/casselc/goo/subjectfacts` for this prepared-fact language.
`Read` validates a definition, `Definition.Evaluate` calculates over prepared facts,
and `Definition.Compile` produces inspectable ClickHouse or Databricks artifacts.
The [external-package example](../subjectfacts/example_test.go) uses general request
outcomes and runs without a database, telemetry exporter or evidence store.

Use `github.com/casselc/goo/measurements` for bounded native OTel instrument
binding/evaluation. Those questions retain instrument identity, units, temporality
and reset/unknown semantics. Its [external evaluation example](../measurements/example_test.go)
uses supplied adapter rows, an explicit reported instance and a fixed clock without
a database or storage.
Preparation of lifecycle events remains in `lifecycle`.

The old Go `measurement` package is renamed directly to `subjectfacts` (227).
Callers update the import and qualifier; there is no compatibility alias. Existing
`goo/v1alpha1` grammar, CLI commands, JSON fields, generated SQL/metric-view YAML
and saved identifiers keep their existing contracts. Reflected OpenAPI components
move from `measurement.*` to `subjectfacts.*`; generated clients should regenerate
against the updated specification. The component bodies and HTTP payload shapes
are unchanged.

## Working commands

From the repository root, with Go 1.27+ on PATH (the validated toolchain is pinned in
[`.go-version`](../.go-version)):

```sh
go test ./...
go run ./cmd/goo validate -definition examples/portable/reservations.yaml
go run ./cmd/goo evaluate -definition examples/portable/reservations.yaml \
  -facts examples/portable/facts.json -evaluation demo-basis
go run ./cmd/goo compile -definition examples/portable/reservations.yaml \
  -target clickhouse -group-by service,release
go run ./cmd/goo compile -definition examples/portable/reservations.yaml \
  -target databricks -group-by service,release
```

In this cloud environment, initialize Go first:

```sh
export PATH=/workspace/.toolchains/go/bin:$PATH
export GOPATH=/workspace/.go GOCACHE=/workspace/.cache/go GOTOOLCHAIN=local
```

The root module now uses the same Go 1.27 language baseline as instrumentation,
examples and verification. CI selects the compiler from `.go-version`; the root
and two example Dockerfiles pin that same patch version. Nested modules remain
independent: run their checks from their own directories.

Raising the root `go` directive also removes Go 1.24 compatibility defaults from
newly built executables. In particular, Go now chooses container-aware
`GOMAXPROCS` and can update it periodically. A positive `GOMAXPROCS` environment
setting or an explicit runtime call overrides that automatic policy. The checked-in
Compose/Docker configuration supplies no such override. Existing running binaries
retain their compiled defaults until rebuilt. This is a policy change, with no
performance improvement claimed; the full set of compiled compatibility settings
can be inspected with `go list -f '{{.DefaultGODEBUG}}' ./cmd/goo`.

`compile` writes a JSON plan containing parameter types, the normalized definition,
source requirements, and query SQL. The Databricks plan additionally contains
metric-view YAML and non-replacing `CREATE VIEW` DDL. No command connects to a
warehouse or creates a database object. The query binds `evaluation_id` separately
(`{evaluation_id:String}` for ClickHouse, `:evaluation_id` for Databricks), plus
the boolean `evidence_impaired` from the complete evaluation receipt. This second
binding suppresses ratios even when an impaired scope has no subject rows.

The two example facts have independent outcomes: one good, one bad, total 2,
success 0.5. Grouping by release gives 1.0 for v1 and 0.0 for v2. Unknown or
impaired evidence suppresses the ratio rather than changing its denominator.

## Implemented semantics

- Strict, single-document `goo/v1alpha1` YAML. Unknown keys, duplicate YAML keys,
  unsupported operators, invalid identifiers, and missing references fail.
- Prepared subject-fact sources with string/bool fields, optional string enums,
  explicitly scoped subject keys, and per-backend physical table bindings.
- Direct-field dimensions, equality predicates, one population filter, filtered
  and unfiltered counts, and ratios of those counts.
- The ratio denominator must be an unfiltered count of the same population.
  This proves numerator containment for this subset. More elaborate compatible
  populations are intentionally not inferred yet.
- Optional `unknownWhen` count references gate a ratio. A positive blocker
  produces `unknown`, zero denominator produces `no_data`, and otherwise the
  ratio is computed from full-population counts. Count results remain exact
  uint64 values; ratios use Float64 and comparisons use documented tolerances.
- Grouped empty input yields no invented groups. Ungrouped empty input yields
  zero counts and a `no_data` ratio. Empty-scope evidence lives in a separate
  evaluation receipt; this pilot cannot infer it from absent facts.
- The reference evaluator requires one evaluation ID, validates selected facts
  before population filtering, rejects duplicate scoped subjects, and keeps
  identical business IDs in different tenants distinct.
- Deterministic ClickHouse and Databricks output. Atomic measures precede
  composed measures. Only validated identifiers and typed literals enter SQL;
  arbitrary SQL expressions are not accepted. The dollar-quoted Databricks DDL
  rejects literals containing its `$$` delimiter.

The Databricks backend targets the bounded YAML 1.1 profile in the pinned official
reference cited by [the research note](DATABRICKS_METRIC_VIEWS.md). It does not use
window measures, joins, materialization, UDFs, or advanced display metadata.

## Local conformance checks

With `requirements-dev.txt` installed:

```sh
python scripts/check_portable.py
```

This builds the real CLI, expands independently authored fixture recipes, runs
the Go reference evaluator, executes generated ClickHouse SQL in embedded chDB,
and compares both with explicit expected counts, ratios, and states. It covers:

- 1001 mature subjects plus excluded pending and other-evaluation facts;
- weighted rollup (1000/1001, not the average of per-release ratios);
- unknown and impaired evidence;
- cross-tenant subject identity;
- zero traffic, both grouped and ungrouped;
- a new declarative count and a SQL-looking string containing quote/backslash
  characters, without changing any compiler branch.

Go tests additionally exercise input rejection, duplicate subjects (including
filtered-out duplicates), enum/type errors, invalid groupings and references,
denominator compatibility, stable output, and DDL delimiter/name protection.
The parity runner's fixture-only parameter substitution is not a production
execution adapter. Production values must use backend parameter binding.

Databricks artifacts are YAML-parsed and structurally checked, not executed.
Embedded ClickHouse is not a Databricks emulator, and local tests do not establish
Spark SQL acceptance, Unity Catalog privileges, or actual lakehouse ingestion.

## Required preparation and current limits

Plans report `requires-prepared-source`: this compiler aggregates existing
subject facts; it does not generate them from OTel events. SQL execution assumes
the declared fields, keys, enums and evidence contract already hold. Unlike the
reference evaluator, the emitted aggregate query is not an input validator.
A serving adapter must validate the source or enforce preparation invariants
before running it. Never point the generated count query at duplicate raw logs.

The caller must pin one evaluation basis. Generated queries do that; a user
querying the published metric view directly could still combine evaluation IDs.
No row-level security, grants, deployment controller, or basis-enforcement API
has been implemented. Scope-wide quality must propagate into prepared facts and
receipts; malformed unkeyed records must not silently vanish during preparation.

The separate [lifecycle preparer](LIFECYCLE_PREPARATION.md) now implements
normalized-event deduplication, mature cohorts, exact deadline boundaries,
late-evidence snapshots and correction-link validation. It feeds these aggregate
evaluators. Source adapters for real OTel records, SQL preparation backends and
durable publication remain pending. The Python API still uses its original
query modes and retains its listed measurement defects. Databricks DDL/query
parity and exporter-Parquet ingestion remain eventual instance tests.
