# syntax=docker/dockerfile:1
FROM golang:1.27.2 AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY annotations ./annotations
COPY metrics ./metrics
COPY measurements ./measurements
COPY measurement ./measurement
COPY lifecycle ./lifecycle
COPY prebuilt ./prebuilt
COPY exploration ./exploration
COPY workbench ./workbench
COPY investigations ./investigations
COPY cmd ./cmd
COPY instrumentation ./instrumentation
COPY examples/operation-app ./examples/operation-app
# A managed proxy may require a combined CA bundle supplied as proxy_ca.
# The mount is ephemeral and never becomes an image layer.
RUN --mount=type=secret,id=proxy_ca \
    if [ -f /run/secrets/proxy_ca ]; then export SSL_CERT_FILE=/run/secrets/proxy_ca; fi; \
    go run ./cmd/goo manifest -dir examples/operation-app > /tmp/manifest.json && \
    cd examples/operation-app && \
    CGO_ENABLED=0 go run github.com/DataDog/orchestrion go build -o /out/operation-app .
FROM debian:bookworm-slim
COPY --from=build /out/operation-app /usr/local/bin/operation-app
CMD ["operation-app"]
