#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
output="${1:-$PWD/artifacts/report-app}"
mkdir -p "$(dirname "$output")"
go run ./cmd/goo manifest -dir examples/report-app > /dev/null
cd examples/report-app
CGO_ENABLED=0 go run github.com/DataDog/orchestrion go build -o "$output" .
