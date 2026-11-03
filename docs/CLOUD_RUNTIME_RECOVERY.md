# Managed workspace validation prerequisites

Source code surviving an environment restore does not imply temporary tools and
lab processes survived. Before a browser journey, check the actual compiler with
`go version`, Playwright with `node_modules/.bin/playwright --version`, Chromium,
and the configured artifact directory. An unrelated executable named `go` is not
the Go compiler.

Keep compiler installations and npm dependencies in persistent workspace storage.
For this workspace, `/workspace/.toolchains/go` now points to
`/workspace/.toolchains/installed-go1.27.2`, not a `/tmp` directory. Verify official
Go archives against the release-index JSON SHA-256 before using them; an invented
adjacent `.sha256` URL may return an HTML documentation redirect instead.

Restore JavaScript dependencies using the unchanged lockfile, for example
`npm ci --cache /workspace/.cache/npm`, with the supplied proxy/CA trust and strict
TLS verification. `node_modules` should be a real workspace directory, not a link
to an expired temporary directory. Do not update dependency versions as recovery.

The configured `artifacts` link currently targets `/tmp/goo-session-artifacts`.
Its target may need recreating after a restore; keep the existing retention policy
and do not invent prior evidence if its temporary contents were lost. Durable
storyboards and screenshots remain under `docs/journeys`.

Use `scripts/run_journeys.sh` for live tests: it pauses and restores only the demo
traffic generator and leaves existing stream entries and volumes intact. Confirm
Redis/ClickHouse/collector prerequisites are running before its drain step. A
missing tool, artifact target or stopped dependency is a setup failure, not proof
that a product journey passed or failed. Keep actor requests and main-journey
requests separately attributed when running optional Bombadil campaigns.

Finish Docker builds and container recreation before Chromium journeys. Docker
network changes can produce `net::ERR_NETWORK_CHANGED` during ES-module loading
even when the journey uses a separate loopback Go process. Keep these phases
sequential; an interrupted import is not evidence of a product startup exception.
The generic journey runtime uses a fixed loopback port; assign one wrapper owner
at a time rather than launching parallel journeys against the same runtime.

API schema checks use the pinned `requirements-api.txt` offline tooling. In this
workspace it is installed in `/workspace/.venvs/goo-api`; avoid a temporary venv
link that disappears on restore. These validators are not product runtime dependencies.
