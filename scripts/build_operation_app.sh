#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
output="${1:-$PWD/artifacts/operation-app}"
mkdir -p "$(dirname "$output")"
go run ./cmd/goo manifest -dir examples/operation-app > /dev/null
cd examples/operation-app
CGO_ENABLED=0 go run github.com/DataDog/orchestrion go build -o "$output" .
