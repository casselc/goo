package verification

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

type heldSchedule396 struct {
	Data     *exploration.Result `json:"data"`
	Grouped  bool                `json:"grouped"`
	Schedule []int               `json:"schedule"`
}

// Real production closures and Session judge bounded generated owner schedules;
// no synthetic renderer or universal replacement lifecycle is introduced.
func TestHegelHeldStagingWrappers396(t *testing.T) {
	cases := []heldSchedule396{}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		schedule := make([]int, 32)
		for i := range schedule {
			schedule[i] = hegel.Draw(tc, hegel.Integers(0, 255))
		}
		data := groupedFixture395(t, seed, []string{"logs", "spans"}[seed%2], false)
		for _, grouped := range []bool{false, true} {
			cases = append(cases, heldSchedule396{data, grouped, append([]int(nil), schedule...)})
		}
	}, options(16)...)
	wire, _ := json.Marshal(cases)
	cmd := exec.CommandContext(t.Context(), "node", "held_staging_396_fixture.mjs")
	cmd.Stdin = bytes.NewReader(wire)
	output, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("actual producer/root wrapper schedules: %v\n%s", e, output)
	}
	var result struct {
		Attempts, Accepted, Rejected, Stale int
		Complete                            bool
	}
	if json.Unmarshal(output, &result) != nil || !result.Complete || result.Accepted < len(cases)*3 || result.Rejected == 0 || result.Stale == 0 {
		t.Fatalf("generated effects or positive/refused controls incomplete: %s", output)
	}
	t.Logf("%d wrapper schedules, %d actual closure attempts, %d accepted stages, %d refusals, %d stale replies", len(cases), result.Attempts, result.Accepted, result.Rejected, result.Stale)
}
