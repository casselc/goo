# Developing Goo

Read the [README](README.md) for the product and [repository guidance](AGENTS.md)
for implementation conventions. The [documentation index](docs/README.md) links
current behavior, architecture and the active roadmap.

## Build and test

Use Go 1.27 or newer; [.go-version](.go-version) pins the validated compiler.
Node is used for frontend tests and browser tooling, not as a product server.

```sh
go build -o /tmp/goo ./cmd/goo
go test ./...
npm ci
npm run test:visualizations
```

The repository has multiple Go modules. Root `go test ./...` does not test nested
modules. The [verification guide](verification/README.md) covers Hegel properties,
Quint models, native prerequisites and the pinned sibling tooling. Missing optional
tools must not be reported as passing checks.

For real integration journeys, start the [local lab](docs/LOCAL_LAB.md), then run:

```sh
npm run test:journeys
```

The wrapper pauses and restores the lab traffic generator while retaining stored
data. Journey tests may require additional native tools; consult their guides.
[Bombadil testing](docs/BOMBADIL_TESTING.md) covers exploratory UI actions and invariants.
See [VALIDATION.md](VALIDATION.md) for executed checks, rather than assuming every
workflow is covered by one command.

## Documentation conventions

The README is a product introduction and quick start, not a milestone log.
Keep it focused on what Goo does, who it serves and how to try it.

- Add user-facing guides to the [documentation index](docs/README.md).
- Update the capability guide when supported behavior changes.
- Keep setup commands aligned with the CLI and Compose configuration.
- Update executable journeys and generated screenshots when workflows change.
- Put test results in VALIDATION.md and detailed contracts in their owning docs.
- Label proposals and historical material explicitly; keep one active roadmap pointer.

Do not prepend each new feature or review report to the README. Link a detailed
guide from the relevant task or capability instead.
