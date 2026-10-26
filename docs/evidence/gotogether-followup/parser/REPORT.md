# Pinned parser regression pilot

Pin: `65093ec1c98abd095f6d399fae038e466555c231`.
Remote module: `v0.0.0-20261026141556-65093ec1c98a`; no local replacements.
Go 1.27.2, GOWORK=off, GOTOOLCHAIN=local, GOMAXPROCS=2.

`go test -v -count=1 -timeout=30s .`: PASS, 0.023s; no skipped tests.
Eight rejection variants cover second JSON object or garbage on the same line, line1/line2, with/without trailing newline. All return an ordinary line-specific trailing-data error, never TornTail. Valid trailing whitespace and blank lines pass. A truly truncated final object retains the valid prefix and returns *qtrace.TornTail on line2 (plus explicit loss record at this pin).

Reproduce using parser_test.go and pinned go.mod/go.sum; results.txt contains raw output. CLI/Quint are not needed for this parser-only check. Initial network-denied go-get was retried with approved network access; install.txt records successful exact-pin download. No Goo changes, no latest-main execution, no browser/heavy tests. Remote CI status unavailable (GitHub API Forbidden), not assumed green.
