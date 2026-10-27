package verification

import (
	"bytes"
	"encoding/json"
	"hegel.dev/go/hegel"
	"os/exec"
	"testing"
)

// Generated withdrawal schedules exercise the actual JavaScript owner, not a
// second execution state machine used as an expected implementation.
func TestHegelResourceDefaultWithdrawal(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		phase := hegel.Draw(tc, hegel.Integers(0, 1))
		withdraw := hegel.Draw(tc, hegel.Integers(0, 2))
		family := "memory"
		if hegel.Draw(tc, hegel.Integers(0, 1)) == 1 {
			family = "cpu"
		}
		input, _ := json.Marshal(map[string]any{"phase": phase, "withdraw": withdraw, "family": family})
		cmd := exec.Command("node", "../scripts/check_resource_default_schedule.js")
		cmd.Stdin = bytes.NewReader(input)
		output, err := cmd.CombinedOutput()
		if err != nil {
			tc.Errorf("actual owner schedule %s: %v %s", input, err, output)
		}
	}, options(8)...)
}
