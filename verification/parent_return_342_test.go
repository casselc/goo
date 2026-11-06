package verification

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"

	"hegel.dev/go/hegel"
)

// Generated schedules exercise the production local-handler and remount-state
// helpers. Actual population navigation, transport and receipts are checked by
// the live journey; this gate does not substitute a DOM or transport model.
func TestHegelRetainedParent342(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		input := struct {
			Rounds     int `json:"rounds"`
			Duplicates int `json:"duplicates"`
			Seed       int `json:"seed"`
		}{hegel.Draw(tc, hegel.Integers(1, 16)), hegel.Draw(tc, hegel.Integers(1, 8)), hegel.Draw(tc, hegel.Integers(0, 1000000))}
		data, err := json.Marshal(input)
		if err != nil {
			tc.Fatal(err)
			return
		}
		command := exec.CommandContext(t.Context(), "node", "--no-warnings", "fixtures/parent-return-342.mjs")
		command.Stdin = bytes.NewReader(data)
		out, err := command.CombinedOutput()
		if err != nil {
			tc.Fatal(err, string(out))
			return
		}
		var result struct {
			Rounds, Duplicates, Seed, LegitimateEffects, RetiredEffects, FailedRenewals, OldCallbacks int
			DraftPreserved                                                                            bool
		}
		if err = json.Unmarshal(out, &result); err != nil {
			tc.Fatal(err, string(out))
			return
		}
		if result.Rounds != input.Rounds || result.Duplicates != input.Duplicates || result.Seed != input.Seed || result.LegitimateEffects != 1+input.Rounds*input.Duplicates || result.RetiredEffects != 0 || result.FailedRenewals != 4*input.Rounds || result.OldCallbacks != input.Rounds+1 || !result.DraftPreserved {
			tc.Fatal("retained local handler or draft ownership failed", input, result)
		}
	}, options(40)...)
}
