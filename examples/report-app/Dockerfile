# syntax=docker/dockerfile:1
FROM golang:1.27.2 AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY annotations ./annotations
COPY metrics ./metrics
COPY measurement ./measurement
COPY lifecycle ./lifecycle
COPY prebuilt ./prebuilt
COPY exploration ./exploration
COPY investigations ./investigations
COPY cmd ./cmd
COPY instrumentation ./instrumentation
COPY examples/report-app ./examples/report-app
# A managed proxy may require a combined CA bundle supplied as proxy_ca.
# The mount is ephemeral and never becomes an image layer.
RUN --mount=type=secret,id=proxy_ca \
    if [ -f /run/secrets/proxy_ca ]; then export SSL_CERT_FILE=/run/secrets/proxy_ca; fi; \
    go run ./cmd/goo manifest -dir examples/report-app > /tmp/manifest.json && \
    cd examples/report-app && \
    CGO_ENABLED=0 go run github.com/DataDog/orchestrion go build -o /out/report-app .
FROM debian:bookworm-slim
COPY --from=build /out/report-app /usr/local/bin/report-app
CMD ["report-app"]
