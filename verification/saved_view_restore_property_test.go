package verification

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"

	"hegel.dev/go/hegel"
)

// The mounted population owner receives generated passive route schedules and
// deferred abort-insensitive definition reads. The fixture's independent oracle
// counts requests and caller adoption; native 335 separately checks real rendered
// controls, focus, completed records and browser-dispatched history events.
func TestHegelSavedViewRestore335(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		input := map[string]int{
			"rounds":     hegel.Draw(tc, hegel.Integers(1, 4)),
			"duplicates": hegel.Draw(tc, hegel.Integers(1, 10)),
			"seed":       hegel.Draw(tc, hegel.Integers(1, 100000)),
		}
		data, err := json.Marshal(input)
		if err != nil {
			tc.Fatal(err)
			return
		}
		command := exec.Command("node", "--no-warnings", "--experimental-test-module-mocks", "fixtures/saved-view-restore-335.mjs")
		command.Stdin = bytes.NewReader(data)
		out, err := command.CombinedOutput()
		if err != nil {
			tc.Fatal(err, string(out))
			return
		}
		var result struct {
			Rounds    int `json:"rounds"`
			Reads     int `json:"reads"`
			Adoptions int `json:"adoptions"`
			Refused   int `json:"refused"`
			Telemetry int `json:"telemetry"`
			Writes    int `json:"writes"`
		}
		if err = json.Unmarshal(out, &result); err != nil {
			tc.Fatal(err, string(out))
			return
		}
		if result.Rounds != input["rounds"] || result.Reads != 8*result.Rounds || result.Adoptions != 3*result.Rounds || result.Refused != 4*result.Rounds || result.Telemetry != 0 || result.Writes != 0 {
			tc.Fatal("lost saved-route schedule or independent admission checks", result, input)
		}
	}, options(20)...)
}
