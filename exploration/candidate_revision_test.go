package exploration

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/casselc/goo/measurements"
)

func TestGeneratedServiceCandidateIdentityCompatibility(t *testing.T) {
	in := dashboardHTTPInput()
	in.Signals = []string{"metrics"}
	in.Profile = "auto"
	in.Instruments = []measurements.Instrument{dashboardRuntimeInstrument("go.goroutine.count", "one", ""), dashboardRuntimeInstrument("go.memory.gc.goal", "one", ""), dashboardRuntimeInstrument("go.memory.used", "one", "stack"), dashboardRuntimeInstrument("go.memory.used", "one", "other")}
	plan, err := PlanServiceDashboard(in, nil)
	if err != nil {
		t.Fatal(err)
	}
	type reference struct {
		Name     string `json:"name"`
		Key      string `json:"key"`
		Default  bool   `json:"default"`
		Selected bool   `json:"selected"`
	}
	refs := []reference{}
	for _, option := range plan.NativeOptions {
		refs = append(refs, reference{option.Instrument.Definition.Name, option.Key, option.Default, option.Selected})
	}
	got, err := json.MarshalIndent(refs, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("../testdata/storage/generated-service-candidates.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(append(got, '\n'), want) {
		t.Fatalf("generated candidate identity/order/defaults changed:\n%s", got)
	}
}
