package generations_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/chucklehead-dev/gotogether/qhegel"
	"github.com/chucklehead-dev/gotogether/qobs"
	"github.com/chucklehead-dev/gotogether/qtest"
	"hegel.dev/go/hegel"
)

func checkBaselines(t testing.TB) {
	t.Helper()
	b, err := os.ReadFile("testdata/baselines.json")
	if err != nil {
		t.Fatal(err)
	}
	var hashes map[string]string
	if err = json.Unmarshal(b, &hashes); err != nil {
		t.Fatal(err)
	}
	for file, want := range hashes {
		b, err := os.ReadFile(filepath.Join("../../lifecycle", file))
		if err != nil {
			t.Fatal(err)
		}
		got := sha256.Sum256(b)
		if hex.EncodeToString(got[:]) != want {
			t.Fatalf("%s changed: review the contract and regenerate/recheck overlays against the new source; do not update the baseline blindly", file)
		}
	}
}
func TestOverlayBaselines(t *testing.T) { checkBaselines(t) }
func mutantWorkload(name string, duplicates, generation int) []operation {
	ops := []operation{{Action: "accept"}}
	for i := 0; i < duplicates; i++ {
		ops = append(ops, operation{Action: "duplicate"})
	}
	switch name {
	case "dropTenant":
		return append(ops, operation{Action: "accept", Tenant: 1})
	case "allowMixed":
		return append(ops, operation{Action: "freeze"}, operation{Action: "terminal", Successful: true}, operation{Action: "correct"}, operation{Action: "selectGeneration", Generation: generation}, operation{Action: "mixed"})
	case "changedScope":
		return append(ops, operation{Action: "freeze"}, operation{Action: "changedScope"})
	default:
		panic("unknown mutant " + name)
	}
}
func TestFaithfulAtMutantWorkloads(t *testing.T) {
	for _, name := range []string{"dropTenant", "allowMixed", "changedScope"} {
		t.Run(name, func(t *testing.T) {
			qtest.Validate(t, bindingPath, directed(t, mutantWorkload(name, 2, 1), true)).Expect(t, qtest.Conforms(), qtest.Settles())
		})
	}
}
func TestMutant(t *testing.T) {
	name := qtest.Mutant()
	if name == "" {
		t.Skip("TestMutants builds the three source overlays")
	}
	action := ""
	switch name {
	case "dropTenant":
		action = "accept"
	case "allowMixed":
		action = "mixed"
	case "changedScope":
		action = "changedScope"
	default:
		t.Fatalf("unknown mutant %q", name)
	}
	v := qhegel.FindViolation(t, qhegel.ViolationConfig{Binding: bindingPath, Opts: []qtest.Option{qtest.RecorderOptions(qobs.WithActor("generations"), qobs.WithProcess("kernel"))}}, func(tc hegel.TestCase, c *qhegel.Case) {
		w := newWorld()
		duplicates := hegel.Draw(tc, hegel.Integers(0, 6))
		generation := hegel.Draw(tc, hegel.Integers(0, 1))
		for _, op := range mutantWorkload(name, duplicates, generation) {
			if err := w.apply(c.Context(t.Context()), c.Recorder, op); err != nil {
				tc.Errorf("actual kernel call: %v", err)
				return
			}
		}
	}, hegel.WithTestCases(30), hegel.WithSeed(42))
	v.Expect(t, qtest.Mismatch(action))
	if v.Last().Action != action {
		t.Fatalf("wrong failing operation: %s", v.Last().Action)
	}
	save(t, name+"-smallest.jsonl", v.Steps)
	save(t, name+"-summary.json", map[string]any{"steps": len(v.Steps), "failingAction": action, "quintRuns": v.QuintRuns, "cacheHits": v.CacheHits, "summary": v.Summary})
}
func TestMutants(t *testing.T) { checkBaselines(t); qtest.Mutants{Package: "../../lifecycle"}.Run(t) }
