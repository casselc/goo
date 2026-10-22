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

FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN --mount=type=secret,id=build_ca \
    if [ -f /run/secrets/build_ca ]; then PIP_CERT=/run/secrets/build_ca pip install --no-cache-dir -r requirements.txt; else pip install --no-cache-dir -r requirements.txt; fi
COPY . .
COPY --from=go-build /out/goo /usr/local/bin/goo
ENV PYTHONUNBUFFERED=1
CMD ["python", "server.py"]
