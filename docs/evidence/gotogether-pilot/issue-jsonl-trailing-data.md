Tested as an outside consumer pinned to `50157b805ae455de165f37bf39da78690958763c` (`v0.0.0-20261026025747-50157b805ae4`), using Go 1.27.2. This reproduction needs no Quint installation.

`qtrace.ReadJSONL` documents one flat JSON object per line, but a newline-terminated line containing a second object or trailing garbage succeeds with one step and no error. The trailing bytes silently disappear.

Minimal failing regression:

```go
package repro_test

import (
    "strings"
    "testing"
    "github.com/chucklehead-dev/gotogether/qtrace"
)

func TestRejectTrailingData(t *testing.T) {
    first := `{"quint.action":"save","quint.seq":1,"quint.recorder":"r","quint.actor":"captures","quint.outcome":"ok"}`
    second := `{"quint.action":"reopen","quint.seq":2,"quint.recorder":"r","quint.actor":"captures","quint.outcome":"ok"}`
    for _, suffix := range []string{second, "garbage"} {
        steps, err := qtrace.ReadJSONL(strings.NewReader(first + suffix + "\n"))
        if err == nil {
            t.Errorf("accepted malformed line: %d steps, error=%v", len(steps), err)
        }
    }
}
```

Actual for both cases: `1 step, error=<nil>`; the regression fails. Expected: an error identifying the malformed line. These inputs end in a newline, so they are not a cut-off final line requiring TornTail recovery.

The cause appears to be `parseLine` decoding once with `json.Decoder.Decode` and not checking for remaining input:
https://github.com/chucklehead-dev/gotogether/blob/50157b805ae455de165f37bf39da78690958763c/qtrace/jsonl.go#L119

Source comparison shows the same behavior remains in `2992868`; only the pinned commit was executed.

This is an evidence-integrity issue: a record or corruption can be discarded without an incomplete/error indication. The reproduction proves parser data loss; it does not claim a separately demonstrated downstream false-green model verdict.

Suggested fix: require EOF after the decoded object (a second Decode must return io.EOF). Retain trailing-whitespace acceptance and existing TornTail behavior; add tests for a second object and garbage on newline-terminated middle/final lines.

Other pilot results were positive: fresh pinned consumer/CLI installation and installed skills worked; hello's faithful, mutant, exhaustion and shrinking tests passed on Quint 0.32.0/TypeScript, and an archived Goo capture journey plus existing negative controls agreed with Goo's old checker. No gotogether source patch was made in this pilot.
