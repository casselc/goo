# syntax=docker/dockerfile:1
FROM golang:1.27.2-alpine AS go-build
# Bind the source once rather than copying the full toolchain snapshot for every
# package on VFS builders. Both Go caches are disposable build-only mounts.
RUN --mount=type=bind,source=.,target=/src,readonly \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=secret,id=build_ca \
    cd /src && \
    if [ -f /run/secrets/build_ca ]; then export SSL_CERT_FILE=/run/secrets/build_ca; fi && \
    CGO_ENABLED=0 go build -o /out/goo ./cmd/goo

# Python belongs only to the independent legacy traffic/sample applications.
# It is not present in the Goo server image.
FROM python:3.12-slim AS sample
WORKDIR /app
COPY requirements.txt .
RUN --mount=type=secret,id=build_ca \
    if [ -f /run/secrets/build_ca ]; then PIP_CERT=/run/secrets/build_ca pip install --no-cache-dir -r requirements.txt; else pip install --no-cache-dir -r requirements.txt; fi
COPY . .
ENV PYTHONUNBUFFERED=1
CMD ["python", "sample.py", "web"]

FROM alpine:3.22 AS runtime
WORKDIR /app
COPY --from=go-build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=go-build /out/goo /usr/local/bin/goo
COPY config ./config
# Lab operation packages validate their declared source annotations at startup.
# These are input data for the scanner; no Go compiler is needed at runtime.
COPY examples/operation-app ./examples/operation-app
COPY examples/report-app ./examples/report-app
# Workflow declarations also validate the referenced telemetry package files.
COPY examples/telemetry ./examples/telemetry
EXPOSE 8089
CMD ["goo", "serve-http", "-profile", "generic", "-listen", ":8089"]
