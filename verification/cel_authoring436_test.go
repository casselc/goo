package verification

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/querycel"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type celAuthoring436Step struct {
	Action    string `json:"action"`
	Ready     bool   `json:"ready"`
	Pending   bool   `json:"pending"`
	Active    bool   `json:"active"`
	Alive     bool   `json:"alive"`
	LastValid bool   `json:"lastValid"`
	Accepted  bool   `json:"accepted"`
	Reads     int    `json:"reads"`
}

// Exercise the real JavaScript owner with real source-free Go planner receipts.
// The projection verifies local authority and effects, not browser or source truth.
func TestQuintCELAuthoring436ObservedAuthority(t *testing.T) {
	base := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 20, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "missing"}}
	id, err := querycel.BaseIdentity(base)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "node", "scripts/cel-authoring436.mjs")
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	output, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()
	encoder := json.NewEncoder(input)
	if err = encoder.Encode(map[string]any{"base": base, "baseIdentity": id}); err != nil {
		t.Fatal(err)
	}
	var steps []celAuthoring436Step
	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		var message struct {
			Request *querycel.PlanRequest `json:"request"`
			Ledger  []celAuthoring436Step `json:"ledger"`
		}
		if err = json.Unmarshal(scanner.Bytes(), &message); err != nil {
			t.Fatal(err)
		}
		if message.Request != nil {
			receipt, err := querycel.PlanRecord(*message.Request)
			var diagnostic *querycel.Diagnostic
			if err != nil && !errors.As(err, &diagnostic) {
				t.Fatal(err)
			}
			if err = encoder.Encode(map[string]any{"receipt": receipt, "diagnostic": diagnostic}); err != nil {
				t.Fatal(err)
			}
		} else {
			steps = message.Ledger
		}
	}
	input.Close()
	if err = scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if err = cmd.Wait(); err != nil {
		t.Fatal(err, stderr.String())
	}
	if len(steps) != 26 {
		t.Fatalf("incomplete observed journey: %d steps", len(steps))
	}
	spec, err := filepath.Abs("model/cel_authoring436.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := "spec: " + fmt.Sprintf("%q", spec) + "\nmodule: cel_authoring436\nconstants: {SOURCE_READS: '0'}\ninvariants: [safety]\nactions:\n"
	expect := "'ready == {{.obs.ready}} and pending == {{.obs.pending}} and active == {{.obs.active}} and alive == {{.obs.alive}} and lastValid == {{.obs.lastValid}} and accepted == {{.obs.accepted}} and reads == {{.obs.reads}}'"
	for _, action := range []string{"edit", "check", "accept", "failed", "stale", "hide", "show", "clear", "reparent", "dispose"} {
		binding += "  " + action + ":\n    expect: " + expect + "\n"
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "binding.yaml")
	if err = os.WriteFile(file, []byte(binding), 0600); err != nil {
		t.Fatal(err)
	}
	checker, err := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	check := func(events []celAuthoring436Step) bool {
		sink := &qobs.MemorySink{}
		recorder := qobs.NewRecorder(sink)
		for _, s := range events {
			recorder.Record(context.Background(), s.Action, nil, "obs.ready", s.Ready, "obs.pending", s.Pending, "obs.active", s.Active, "obs.alive", s.Alive, "obs.lastValid", s.LastValid, "obs.accepted", s.Accepted, "obs.reads", s.Reads)
		}
		result, err := checker.Check(context.Background(), sink.Steps())
		if err != nil {
			t.Fatal(err)
		}
		return result.OK()
	}
	if !check(steps) {
		t.Fatalf("actual authoring authority rejected: %+v", steps)
	}
	for mutant := 0; mutant < 6; mutant++ {
		altered := append([]celAuthoring436Step(nil), steps...)
		switch mutant {
		case 0:
			altered[4].Accepted = true // deferred A→B→A reply
		case 1:
			altered[7].Accepted = true // direct retired stage callback
		case 2:
			altered[8].Ready = true // invalid text cannot stage old valid receipt
		case 3:
			altered[15].Accepted = true // hidden owner reply
		case 4:
			altered[len(altered)-1].Accepted = true // disposed owner error
		case 5:
			altered[len(altered)-1].Reads++ // local authoring must not read
		}
		if check(altered) {
			t.Fatalf("incorrect authority trace accepted: mutant %d", mutant)
		}
	}
	t.Logf("validated %d real owner steps against Quint; six authority/effect mutants rejected", len(steps))
}
