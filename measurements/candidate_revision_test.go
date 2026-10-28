package measurements

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestGeneratedCandidateIdentityCompatibility(t *testing.T) {
	type reference struct {
		Name        string `json:"name"`
		Revision    string `json:"revision"`
		QuestionKey string `json:"questionKey,omitempty"`
	}
	runtime := map[string][]reference{}
	instruments := []Instrument{}
	for _, metric := range []string{"go.memory.used", "go.memory.gc.goal", "go.goroutine.count"} {
		for _, i := range GoRuntimeCandidates(runtimeInstrument(t, metric, nil)) {
			instruments = append(instruments, i)
			runtime[metric] = append(runtime[metric], reference{Name: i.Definition.Name, Revision: Hash(*i.Definition)})
		}
	}
	pod, err := PlanPodDashboard(PodDashboardRequest{Discovery: podDiscoveryFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	refs := []reference{}
	for _, w := range pod.Candidates {
		refs = append(refs, reference{Name: w.Question.Definition.Name, Revision: Hash(w.Question.Definition), QuestionKey: w.Key})
	}
	got, err := json.MarshalIndent(struct {
		Runtime  map[string][]reference `json:"runtime"`
		Defaults map[string]bool        `json:"runtimeDefaults"`
		Pod      []reference            `json:"pod"`
	}{runtime, GoRuntimeDefaults(instruments), refs}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("../testdata/storage/generated-candidates.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(append(got, '\n'), want) {
		t.Fatalf("generated candidate identity/order/defaults changed:\n%s", got)
	}
}

func TestDefinitionSeedExcludesAuthoringLabelsWithoutIssuingARevision(t *testing.T) {
	d := definition(t, "worker-utilization")
	before, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	seed, err := definitionSeed(d)
	if err != nil {
		t.Fatal(err)
	}
	legacy := d
	legacy.Name, legacy.Title = "", ""
	if seed != Hash(legacy) {
		t.Fatal("authoring seed identity changed")
	}
	after, err := json.Marshal(d)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("seed mutated definition", err)
	}
	renamed := d
	renamed.Name, renamed.Title = "another-name", "Another title"
	other, err := definitionSeed(renamed)
	if err != nil || seed != other {
		t.Fatal("labels multiplied seed", err)
	}
	final, err := d.Revision()
	if err != nil {
		t.Fatal(err)
	}
	renamedFinal, err := renamed.Revision()
	if err != nil || final == renamedFinal {
		t.Fatal("authored labels missing from final revision", err)
	}
	unnamed := d
	unnamed.Name = ""
	if _, err := definitionSeed(unnamed); err != nil {
		t.Fatal("pre-name seed unavailable", err)
	}
	if revision, err := unnamed.Revision(); err == nil || revision != "" {
		t.Fatal("seed admitted as authored revision", revision, err)
	}
	changed := d
	changed.MaxGapMS++
	other, err = definitionSeed(changed)
	if err != nil || other == seed {
		t.Fatal("evaluation policy missing from seed", err)
	}
}
