package verification

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"hegel.dev/go/hegel"
)

// The independent oracle checks literal row matching and counts after generated
// mutations. This exercises the actual registered filtering method; native tests
// cover browser reactions, focus, rendering and source IO separately.
func TestHegelLargeFacetFilter343(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		input := struct {
			Size   int `json:"size"`
			Rounds int `json:"rounds"`
			Seed   int `json:"seed"`
		}{hegel.Draw(tc, hegel.Integers(1, 250)), hegel.Draw(tc, hegel.Integers(1, 12)), hegel.Draw(tc, hegel.Integers(0, 1000000))}
		data, err := json.Marshal(input)
		if err != nil {
			tc.Fatal(err)
			return
		}
		command := exec.CommandContext(t.Context(), "node", "--no-warnings", "fixtures/large-facets-343.mjs")
		command.Stdin = bytes.NewReader(data)
		out, err := command.CombinedOutput()
		if err != nil {
			tc.Fatal(err, string(out))
			return
		}
		var result struct {
			History []struct {
				Query    string
				Rows     []string
				Hidden   []bool
				Scans    int
				Identity bool
				Result   struct{ Visible, Total int }
			}
		}
		if err = json.Unmarshal(out, &result); err != nil {
			tc.Fatal(err, string(out))
			return
		}
		if len(result.History) != 4*input.Rounds {
			tc.Fatal("unfinished filtering schedule", input, len(result.History))
		}
		for _, step := range result.History {
			if len(step.Rows) != len(step.Hidden) || !step.Identity || step.Scans != 1 || step.Result.Total != len(step.Rows) {
				tc.Fatal("filtering replaced rows, repeated work or returned stale totals", step)
				return
			}
			visible := 0
			for i, row := range step.Rows {
				matches := strings.Contains(strings.ToLower(row), strings.ToLower(step.Query))
				if matches {
					visible++
				}
				if step.Hidden[i] == matches {
					tc.Fatal("literal matching disagrees with independent oracle", row, step.Query)
					return
				}
			}
			if step.Result.Visible != visible {
				tc.Fatal("filtering returned stale match count", visible, step.Result.Visible)
			}
		}
	}, options(40)...)
}
