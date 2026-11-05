package verification

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"

	"hegel.dev/go/hegel"
)

// Exercise the actual captures owner with deferred, abort-insensitive custody
// transport. The oracle counts reads/adoptions independently of owner internals;
// display adapters are passive doubles, so this does not prove browser focus.
func TestHegelCaptureRoute332(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		input := map[string]int{"rounds": hegel.Draw(tc, hegel.Integers(1, 8)), "duplicates": hegel.Draw(tc, hegel.Integers(1, 12)), "seed": hegel.Draw(tc, hegel.Integers(1, 10000))}
		data, err := json.Marshal(input)
		if err != nil {
			tc.Fatal(err)
			return
		}
		command := exec.Command("node", "--no-warnings", "--experimental-test-module-mocks", "fixtures/capture-route-332.mjs")
		command.Stdin = bytes.NewReader(data)
		out, err := command.CombinedOutput()
		if err != nil {
			tc.Fatal(err, string(out))
			return
		}
		var result struct {
			Rounds      int `json:"rounds"`
			Reads       int `json:"reads"`
			Adoptions   int `json:"adoptions"`
			Writes      int `json:"writes"`
			Attachments int `json:"attachments"`
		}
		if err = json.Unmarshal(out, &result); err != nil {
			tc.Fatal(err, string(out))
			return
		}
		if result.Rounds != input["rounds"] || result.Reads != 9*input["rounds"] || result.Adoptions != 6*input["rounds"] || result.Writes != 6*input["rounds"] || result.Attachments != 4*input["rounds"] {
			tc.Fatal("lost acquisition trace", result, input)
		}
	}, options(20)...)
}
