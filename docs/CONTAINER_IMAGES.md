# Container images

Goo's product runtime is one Go executable with embedded UI, API documentation and
UI example catalogs. There are two image paths. Neither requires Python in the
Goo container; the independent Python sample traffic applications still use the
Dockerfile's `sample` target.

## Go-only product image with ko

Install the pinned [ko release](https://github.com/ko-build/ko/releases/tag/v0.19.1)
using the Go toolchain in [`.go-version`](../.go-version):

```sh
go install github.com/google/ko@v0.19.1
make image
```

If source-install dependencies are unavailable, use the official release archive
for the pinned version and platform and verify its published checksum before
extracting `ko`. The Linux x86_64 archive used for this validation has SHA-256
`635ac6ea3fd376c935fee597fbb29ab2c2449f49ef1655085fe3aa9c25fed7a5`.
This checksum check is not a provenance-attestation verification.

`make image` uses [.ko.yaml](../.ko.yaml), builds `./cmd/goo` with `CGO_ENABLED=0`
and readonly modules, and loads the image into the local Docker daemon. It does
not push to a registry. `KO=/path/to/ko` selects an installed executable;
`IMAGE_PLATFORM=linux/arm64` selects a different local platform. The default is
`linux/amd64`. The actual image reference is also written to `/tmp/goo-ko-image.refs`
(`IMAGE_REFS=/chosen/path` overrides it). Use that reference rather than assuming
image naming or a digest survives loading into a local daemon.

The base is pinned to the inspected distroless Debian 12 static **nonroot**
multi-platform index:

```text
gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
```

This avoids ko's rolling default base. Reinspect and deliberately update the
pinned digest for base maintenance. The Go executable is the image entry point
(`/ko-app/goo`), so pass `serve-http` as the first argument, without another `goo`:

```sh
GOO_IMAGE="$(cat /tmp/goo-ko-image.refs)"
docker run --rm -p 127.0.0.1:8088:8089 \
  --tmpfs /data:rw,uid=65532,gid=65532,mode=0700 \
  "$GOO_IMAGE" \
  serve-http -profile generic -listen :8089 -investigations-dir /data
```

Supply the normal ClickHouse connection environment for source-backed exploration.
`/health` is process liveness; it does not query ClickHouse. Static resources,
capabilities and profile metadata work without a data-source connection. The
image contains no shell or Python. A fresh tmpfs is suitable only for disposable
checks; mount persistent storage for retained captures. Existing root-owned lab
volumes are not automatically writable by UID 65532. Choose an explicit ownership
migration before switching those volumes; the image build never changes them.

### Filesystem declarations are still separate inputs

This ko image intentionally does not copy the checkout into the runtime. Generic
startup and the embedded UI are supported without sample files, but the built-in
measurement catalog and workflow definitions still read `config/`; the lab profile
also reads the configured operation-package YAML and its `sourceDirectory` Go
annotation inputs. Workflows validate referenced `examples/telemetry` files. A
bare ko image therefore has an empty measurement-package catalog and cannot load
the lab's workflow catalog or operation-package configuration. It is **not yet a
replacement for the configured Docker lab image**.

For a configured deployment, mount the chosen declarations read-only at stable
paths and pass the existing package/binding/inventory arguments; the current lab
layout is `/app/config`, `/app/examples/operation-app`,
`/app/examples/report-app` and `/app/examples/telemetry`, with working directory
`/app`. Keep the writable investigations directory separate. Future portable
bundling must cover all those referenced inputs and test the resulting image.

Ko's [static assets convention](https://ko.build/features/static-assets/) bundles
`cmd/goo/kodata` at the path announced by `KO_DATA_PATH` (v0.19.1 uses `/var/run/ko`).
Goo does not currently interpret that variable or change its configuration roots.
Do not add a `kodata` directory and assume existing relative paths will work.
Despite the documentation's symlink tip, pinned ko v0.19.1 rejects symlinks escaping
the kodata root; any later bundle should stage actual files and define its path
contract explicitly. The UI already uses Go embedding and needs no kodata copy.

## Configured local lab image

The main [Dockerfile](../Dockerfile)'s final `runtime` target builds the same Go
executable and includes the current lab configuration, operation annotation
sources and telemetry-package declarations. [Compose](../compose.yaml) selects
that target for `packages` and maps both loopback ports 8088 and 8089 to its one Go
listener. The `investigations` volume retains its existing name and mount path.

```sh
docker compose build packages
docker compose up -d --no-deps packages
```

When upgrading an older lab, first stop its specific obsolete `query` container
so it releases port 8088; do not delete data volumes or unrelated containers.
There is no Python proxy service in the active Compose definition. The remaining
Dockerfile/Compose path is kept until a configured ko bundle has equivalent
explicit inputs and passes the real lab journeys.

## Managed environment trust

Ko builds use the host Go toolchain and build cache, not a Docker build stage.
Preserve the managed proxy, credentials and CA trust. The selected `gcr.io` base
must be allowed by network policy. `KO_DOCKER_REPO=ko.local` selects loading into the local
Docker daemon. Do not change registry credentials or publish an image merely to
exercise the local path.

Docker builds can use [compose.build-ca.yaml](../compose.build-ca.yaml) to mount a
trusted combined CA bundle during dependency downloads. Build trust is separate
from runtime HTTPS trust; never copy a session-specific CA into a release image
or disable TLS verification. Ko's [configuration](https://ko.build/configuration/),
[build command](https://ko.build/reference/ko_build/) and
[Dockerfile migration guide](https://ko.build/advanced/migrating-from-dockerfile/)
explain its supported boundaries.

## Validation — 2026-10-28 UTC

Pinned ko v0.19.1 built and loaded the `linux/amd64` generic product image locally;
no registry push was performed. The official release archive's checksum matched.
The source-install route was attempted first but two transitive Go module archive
requests were unavailable in this managed environment; project module files were
unchanged. The release-binary fallback used the existing proxy and trust policy.

An isolated container ran as UID 65532 with a read-only root filesystem, disposable
UID-owned `/data` tmpfs and an ephemeral loopback port. Liveness, embedded HTML/ESM,
OpenAPI, UI configuration and API capabilities responded successfully with no
available ClickHouse connection. Its measurement catalog was empty, matching the
bare-image limitation above. Exported rootfs inspection found `/ko-app/goo` and no
shell or Python executable. The temporary container was removed; the local image
was retained for review. Existing lab containers and retained volumes were not
used or modified by this check. This does not validate the configured lab through
ko, persistent-volume migration, ARM64 or remote image publication.
