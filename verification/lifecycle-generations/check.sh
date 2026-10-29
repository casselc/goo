#!/usr/bin/env bash
# Run from any directory. Requires Go >=1.27, Quint 0.33.0 and Java 21.
set -euo pipefail
cd "$(dirname "$0")"
export GOTOGETHER_STRICT=1
go tool gotogether simulate -binding model/generations.binding.yaml
go tool gotogether simulate -binding model/generations.binding.yaml -verify 9
go test ./... -count=1
(cd ../.. && go test ./lifecycle ./subjectfacts -count=1)
